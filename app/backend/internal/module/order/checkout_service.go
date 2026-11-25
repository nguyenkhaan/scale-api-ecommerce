package order

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/cart"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	errIdempotencyExists     = errors.New("idempotency record already exists")
	ErrIdempotencyConflict   = errors.New("idempotency key was used with a different request")
	ErrIdempotencyInProgress = errors.New("idempotency request is still in progress")
)

type PlaceOrderInput struct {
	UserID, IdempotencyKey, RequestID string
	Request                           PlaceOrderRequest
}

type PlaceOrderResult struct {
	Status   int
	Body     json.RawMessage
	Replayed bool
	Outcome  CheckoutOutcome
}

type CheckoutService struct {
	database *gorm.DB
	recorder CheckoutRecorder
}

func NewCheckoutService(database *gorm.DB, recorder CheckoutRecorder) (*CheckoutService, error) {
	if database == nil {
		return nil, errors.New("checkout service requires a database")
	}
	if recorder == nil {
		recorder = noOpCheckoutRecorder{}
	}
	return &CheckoutService{database: database, recorder: recorder}, nil
}

func (s *CheckoutService) PlaceOrder(ctx context.Context, input PlaceOrderInput) (PlaceOrderResult, error) {
	var result PlaceOrderResult
	var recordedCode string
	requestHash := CanonicalRequestHash(input.Request)
	checkoutDatabase := s.database.Session(&gorm.Session{Logger: logger.Discard}).WithContext(ctx)
	err := checkoutDatabase.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`SET LOCAL lock_timeout = '2s'`).Error; err != nil {
			return fmt.Errorf("set checkout lock timeout: %w", err)
		}
		var id string
		insert := tx.Raw(`INSERT INTO idempotency_records (user_id, operation, idempotency_key, request_hash)
			VALUES (?, 'place_order', ?, ?)
			ON CONFLICT (user_id, operation, idempotency_key) DO NOTHING
			RETURNING id`, input.UserID, input.IdempotencyKey, requestHash[:]).Scan(&id)
		if insert.Error != nil {
			return fmt.Errorf("insert checkout idempotency record: %w", insert.Error)
		}
		if insert.RowsAffected == 0 {
			return errIdempotencyExists
		}
		if err := tx.Exec(`SET LOCAL lock_timeout = DEFAULT`).Error; err != nil {
			return fmt.Errorf("restore checkout lock timeout: %w", err)
		}
		if err := tx.SavePoint("checkout_quote").Error; err != nil {
			return fmt.Errorf("create checkout quote savepoint: %w", err)
		}

		quote, err := cart.LockAndQuote(tx, input.UserID, selectionFromRequest(input.Request))
		if err != nil {
			var lineErrors *cart.LineErrors
			if errors.As(err, &lineErrors) {
				body := problem.LineErrorsProblem("One or more items can no longer be purchased.", lineErrors.Lines)
				body.RequestID = input.RequestID
				return s.persistRejection(tx, input, &result, &recordedCode, body)
			}
			if errors.Is(err, cart.ErrCartItemChanged) {
				body := checkoutProblem(http.StatusConflict, problem.CodeCartItemChanged,
					"The cart changed after it was previewed.", input.RequestID)
				return s.persistRejection(tx, input, &result, &recordedCode, body)
			}
			var validationError *cart.ValidationError
			if errors.As(err, &validationError) {
				return fmt.Errorf("checkout selection rejected: %w", err)
			}
			return fmt.Errorf("lock and quote checkout cart: %w", err)
		}

		lineProblems, cartChanged := compareQuote(input.Request, quote)
		if cartChanged {
			body := checkoutProblem(http.StatusConflict, problem.CodeCartItemChanged,
				"The cart changed after it was previewed.", input.RequestID)
			return s.persistRejection(tx, input, &result, &recordedCode, body)
		}
		if len(lineProblems) > 0 {
			body := problem.LineErrorsProblem("The checkout prices changed. Review the current prices and try again.", lineProblems)
			body.RequestID = input.RequestID
			return s.persistRejection(tx, input, &result, &recordedCode, body)
		}
		if input.Request.ExpectedTotal != quote.TotalAmount {
			problems := priceChangedProblems(input.Request, quote)
			body := problem.LineErrorsProblem("The checkout total changed. Review the current prices and try again.", problems)
			body.RequestID = input.RequestID
			return s.persistRejection(tx, input, &result, &recordedCode, body)
		}

		orderNumber, err := newOrderNumber(quote.Now)
		if err != nil {
			return err
		}
		var orderID string
		if err := tx.Raw(`INSERT INTO orders (user_id, order_number, status, currency, total_amount,
			recipient_name, recipient_phone, recipient_address, expires_at)
			VALUES (?, ?, 'pending_payment', ?, ?, ?, ?, ?, now() + interval '5 minutes') RETURNING id`,
			input.UserID, orderNumber, input.Request.Currency, quote.TotalAmount,
			input.Request.Recipient.Name, input.Request.Recipient.Phone, input.Request.Recipient.Address,
		).Scan(&orderID).Error; err != nil {
			return fmt.Errorf("insert checkout order: %w", err)
		}
		if err := createOrderSnapshots(tx, orderID, quote); err != nil {
			return err
		}
		if err := reserveQuote(tx, input.UserID, quote); err != nil {
			return err
		}
		created, err := getOwnedOrder(ctx, tx, input.UserID, orderID)
		if err != nil {
			return fmt.Errorf("load created checkout order: %w", err)
		}
		body, err := json.Marshal(created)
		if err != nil {
			return fmt.Errorf("encode created checkout order: %w", err)
		}
		storedBody, err := saveIdempotencyResponse(tx, input, http.StatusCreated, body)
		if err != nil {
			return err
		}
		result = PlaceOrderResult{Status: http.StatusCreated, Body: storedBody, Outcome: OutcomeCreated}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		if errors.Is(err, errIdempotencyExists) {
			replayed, code, resolveErr := s.resolveExistingAttempt(ctx, input, requestHash)
			if resolveErr == nil {
				s.recorder.RecordCheckout(OutcomeReplayed, code)
				return replayed, nil
			}
			err = resolveErr
		}
		if isLockNotAvailable(err) {
			err = ErrIdempotencyInProgress
		}
		if errors.Is(err, ErrIdempotencyConflict) {
			s.recorder.RecordCheckout(OutcomeConflict, problem.CodeIdempotencyConflict)
			return PlaceOrderResult{}, err
		}
		if errors.Is(err, ErrIdempotencyInProgress) {
			s.recorder.RecordCheckout(OutcomeInProgress, problem.CodeIdempotencyInProgress)
			return PlaceOrderResult{}, err
		}
		s.recorder.RecordCheckout(OutcomeError, "")
		return PlaceOrderResult{}, err
	}
	s.recorder.RecordCheckout(result.Outcome, recordedCode)
	return result, nil
}

func (s *CheckoutService) persistRejection(
	tx *gorm.DB,
	input PlaceOrderInput,
	result *PlaceOrderResult,
	recordedCode *string,
	body problem.Problem,
) error {
	if err := tx.RollbackTo("checkout_quote").Error; err != nil {
		return fmt.Errorf("rollback rejected checkout quote: %w", err)
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode checkout rejection: %w", err)
	}
	storedBody, err := saveIdempotencyResponse(tx, input, http.StatusConflict, encoded)
	if err != nil {
		return err
	}
	*result = PlaceOrderResult{Status: http.StatusConflict, Body: storedBody, Outcome: OutcomeRejected}
	*recordedCode = body.Code
	return nil
}

func saveIdempotencyResponse(tx *gorm.DB, input PlaceOrderInput, status int, body []byte) ([]byte, error) {
	var stored string
	update := tx.Raw(`UPDATE idempotency_records SET response_status = ?, response_body = ?::jsonb, completed_at = now()
		WHERE user_id = ? AND operation = 'place_order' AND idempotency_key = ?
		RETURNING response_body::text`,
		status, string(body), input.UserID, input.IdempotencyKey).Scan(&stored)
	if update.Error != nil {
		return nil, fmt.Errorf("save checkout idempotency response: %w", update.Error)
	}
	if update.RowsAffected != 1 {
		return nil, fmt.Errorf("save checkout idempotency response: updated %d rows", update.RowsAffected)
	}
	return []byte(stored), nil
}

func (s *CheckoutService) resolveExistingAttempt(ctx context.Context, input PlaceOrderInput, requestHash [32]byte) (PlaceOrderResult, string, error) {
	var record struct {
		RequestHash    []byte       `gorm:"column:request_hash"`
		ResponseStatus *int         `gorm:"column:response_status"`
		ResponseBody   *string      `gorm:"column:response_body"`
		CompletedAt    sql.NullTime `gorm:"column:completed_at"`
	}
	query := s.database.Session(&gorm.Session{Logger: logger.Discard}).WithContext(ctx).Raw(`SELECT request_hash, response_status, response_body::text AS response_body, completed_at
		FROM idempotency_records WHERE user_id = ? AND operation = 'place_order' AND idempotency_key = ?`,
		input.UserID, input.IdempotencyKey).Scan(&record)
	if query.Error != nil {
		return PlaceOrderResult{}, "", fmt.Errorf("read existing checkout attempt: %w", query.Error)
	}
	if query.RowsAffected != 1 {
		return PlaceOrderResult{}, "", errors.New("idempotency conflict row disappeared")
	}
	if !bytes.Equal(record.RequestHash, requestHash[:]) {
		return PlaceOrderResult{}, problem.CodeIdempotencyConflict, ErrIdempotencyConflict
	}
	if !record.CompletedAt.Valid || record.ResponseStatus == nil || record.ResponseBody == nil {
		return PlaceOrderResult{}, problem.CodeIdempotencyInProgress, ErrIdempotencyInProgress
	}
	code := checkoutStoredProblemCode([]byte(*record.ResponseBody))
	return PlaceOrderResult{
		Status: *record.ResponseStatus, Body: json.RawMessage(*record.ResponseBody),
		Replayed: true, Outcome: OutcomeReplayed,
	}, code, nil
}

func checkoutStoredProblemCode(body []byte) string {
	var stored struct {
		Code string `json:"code"`
	}
	if json.Unmarshal(body, &stored) == nil {
		return stored.Code
	}
	return ""
}

func isLockNotAvailable(err error) bool {
	var databaseError *pgconn.PgError
	return errors.As(err, &databaseError) && databaseError.Code == "55P03"
}

func selectionFromRequest(request PlaceOrderRequest) cart.CheckoutSelection {
	selection := cart.CheckoutSelection{Items: make([]cart.SelectedCartItem, len(request.Lines))}
	for i, line := range request.Lines {
		selection.Items[i] = cart.SelectedCartItem{CartItemID: line.CartItemID, Version: line.CartItemVersion}
	}
	return selection
}

func compareQuote(request PlaceOrderRequest, quote cart.LockedQuote) ([]problem.LineProblem, bool) {
	quoted := make(map[string]cart.QuotedLine, len(quote.Lines))
	for _, line := range quote.Lines {
		quoted[line.CartItemID] = line
	}
	var problems []problem.LineProblem
	cartChanged := false
	for _, requested := range request.Lines {
		current, ok := quoted[requested.CartItemID]
		if !ok || current.SKUID != requested.SKUID || current.Quantity != requested.Quantity {
			cartChanged = true
			continue
		}
		if current.UnitPrice != requested.UnitPrice || current.PriceSource != requested.PriceSource ||
			!sameOptionalString(current.FlashSaleItemID, requested.FlashSaleItemID) {
			price := current.UnitPrice
			problems = append(problems, problem.LineProblem{
				SKUID: current.SKUID, RequestedQuantity: current.Quantity,
				Code: problem.CodePriceChanged, CurrentUnitPrice: &price,
			})
		}
	}
	return problems, cartChanged
}

func priceChangedProblems(request PlaceOrderRequest, quote cart.LockedQuote) []problem.LineProblem {
	quoted := make(map[string]cart.QuotedLine, len(quote.Lines))
	for _, line := range quote.Lines {
		quoted[line.CartItemID] = line
	}
	problems := make([]problem.LineProblem, 0, len(request.Lines))
	for _, line := range request.Lines {
		if current, ok := quoted[line.CartItemID]; ok {
			price := current.UnitPrice
			problems = append(problems, problem.LineProblem{
				SKUID: current.SKUID, RequestedQuantity: current.Quantity,
				Code: problem.CodePriceChanged, CurrentUnitPrice: &price,
			})
		}
	}
	return problems
}

func sameOptionalString(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func checkoutProblem(status int, code, detail, requestID string) problem.Problem {
	body := problem.CodeProblem(status, code, detail)
	body.RequestID = requestID
	return body
}

func createOrderSnapshots(tx *gorm.DB, orderID string, quote cart.LockedQuote) error {
	for _, line := range quote.Lines {
		var orderItemID string
		if err := tx.Raw(`INSERT INTO order_items (order_id, sku_id, sku_code, product_name, sku_attributes,
			quantity, unit_price, subtotal, price_source, flash_sale_item_id, cart_item_id, cart_item_version)
			VALUES (?, ?, ?, ?, ?::jsonb, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
			orderID, line.SKUID, line.SKUCode, line.ProductName, string(line.SKUAttributes),
			line.Quantity, line.UnitPrice, line.Subtotal, line.PriceSource, line.FlashSaleItemID,
			line.CartItemID, line.CartItemVersion,
		).Scan(&orderItemID).Error; err != nil {
			return fmt.Errorf("insert checkout item snapshot: %w", err)
		}
		if err := tx.Exec(`INSERT INTO inventory_reservations (order_item_id, sku_id, quantity) VALUES (?, ?, ?)`,
			orderItemID, line.SKUID, line.Quantity).Error; err != nil {
			return fmt.Errorf("create inventory reservation: %w", err)
		}
	}
	return nil
}

func reserveQuote(tx *gorm.DB, userID string, quote cart.LockedQuote) error {
	inventoryQuantities := make(map[string]int64, len(quote.Lines))
	flashQuantities := make(map[string]int64)
	for _, line := range quote.Lines {
		inventoryQuantities[line.SKUID] += int64(line.Quantity)
		if line.PriceSource == "flash_sale" && line.FlashSaleItemID != nil {
			flashQuantities[*line.FlashSaleItemID] += int64(line.Quantity)
		}
	}
	skuIDs := sortedKeys(inventoryQuantities)
	for _, skuID := range skuIDs {
		quantity := inventoryQuantities[skuID]
		update := tx.Exec(`UPDATE inventories SET reserved = reserved + ?, version = version + 1, updated_at = now()
			WHERE sku_id = ? AND on_hand - reserved >= ?`, quantity, skuID, quantity)
		if update.Error != nil {
			return fmt.Errorf("reserve inventory: %w", update.Error)
		}
		if update.RowsAffected != 1 {
			return fmt.Errorf("reserve inventory for SKU %s: row unavailable after quote", skuID)
		}
	}
	flashIDs := sortedKeys(flashQuantities)
	for _, flashID := range flashIDs {
		quantity := flashQuantities[flashID]
		update := tx.Exec(`UPDATE flash_sale_items SET quota_reserved = quota_reserved + ?, updated_at = now()
			WHERE id = ? AND quota_reserved + quota_sold + ? <= quota_total`, quantity, flashID, quantity)
		if update.Error != nil {
			return fmt.Errorf("reserve flash-sale quota: %w", update.Error)
		}
		if update.RowsAffected != 1 {
			return fmt.Errorf("reserve flash-sale quota %s: row unavailable after quote", flashID)
		}
		usage := tx.Exec(`UPDATE flash_sale_customer_usage SET reserved_quantity = reserved_quantity + ?, updated_at = now()
			WHERE flash_sale_item_id = ? AND user_id = ?`, quantity, flashID, userID)
		if usage.Error != nil {
			return fmt.Errorf("reserve flash-sale customer usage: %w", usage.Error)
		}
		if usage.RowsAffected != 1 {
			return fmt.Errorf("reserve flash-sale customer usage %s: row missing after quote", flashID)
		}
	}
	return nil
}

func sortedKeys(values map[string]int64) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
