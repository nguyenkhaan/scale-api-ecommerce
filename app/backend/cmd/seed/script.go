package main

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/database"
	"cloudian/cloudian-scale/internal/module/auth"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"uuid"
)

const fixtureRecordCount = 4

const seedCommandTimeout = 60 * time.Second

//go:embed data.json
var fixtureJSON []byte

type fixture struct {
	Users                  []userFixture                   `json:"users"`
	Sessions               []sessionFixture                `json:"sessions"`
	LoginThrottles         []loginThrottleFixture          `json:"login_throttles"`
	Categories             []categoryFixture               `json:"categories"`
	Products               []productFixture                `json:"products"`
	SKUs                   []skuFixture                    `json:"skus"`
	Inventories            []inventoryFixture              `json:"inventories"`
	InventoryAdjustments   []inventoryAdjustmentFixture    `json:"inventory_adjustments"`
	FlashSaleCampaigns     []flashSaleCampaignFixture      `json:"flash_sale_campaigns"`
	FlashSaleItems         []flashSaleItemFixture          `json:"flash_sale_items"`
	FlashSaleCustomerUsage []flashSaleCustomerUsageFixture `json:"flash_sale_customer_usage"`
	Carts                  []cartFixture                   `json:"carts"`
	CartItems              []cartItemFixture               `json:"cart_items"`
}

type userFixture struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

type sessionFixture struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	TokenSeed string `json:"token_seed"`
	CSRFSeed  string `json:"csrf_seed"`
}

type loginThrottleFixture struct {
	EmailNormalized string `json:"email_normalized"`
	FailureCount    int    `json:"failure_count"`
}

type categoryFixture struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	IsVisible bool   `json:"is_visible"`
}

type productFixture struct {
	ID          string `json:"id"`
	CategoryID  string `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Status      string `json:"status"`
}

type skuFixture struct {
	ID                  string          `json:"id"`
	ProductID           string          `json:"product_id"`
	Code                string          `json:"code"`
	Attributes          json.RawMessage `json:"attributes"`
	RegularPrice        int64           `json:"regular_price"`
	MaxQuantityPerOrder int             `json:"max_quantity_per_order"`
	Status              string          `json:"status"`
}

type inventoryFixture struct {
	SKUID    string `json:"sku_id"`
	OnHand   int64  `json:"on_hand"`
	Reserved int64  `json:"reserved"`
	Version  int64  `json:"version"`
}

type inventoryAdjustmentFixture struct {
	ID           string `json:"id"`
	SKUID        string `json:"sku_id"`
	ActorUserID  string `json:"actor_user_id"`
	Delta        int64  `json:"delta"`
	OnHandBefore int64  `json:"on_hand_before"`
	OnHandAfter  int64  `json:"on_hand_after"`
	Reason       string `json:"reason"`
}

type flashSaleCampaignFixture struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type flashSaleItemFixture struct {
	ID            string `json:"id"`
	CampaignID    string `json:"campaign_id"`
	SKUID         string `json:"sku_id"`
	SalePrice     int64  `json:"sale_price"`
	QuotaTotal    int64  `json:"quota_total"`
	QuotaReserved int64  `json:"quota_reserved"`
	QuotaSold     int64  `json:"quota_sold"`
	PurchaseLimit int    `json:"purchase_limit"`
}

type flashSaleCustomerUsageFixture struct {
	FlashSaleItemID  string `json:"flash_sale_item_id"`
	UserID           string `json:"user_id"`
	ReservedQuantity int64  `json:"reserved_quantity"`
	SoldQuantity     int64  `json:"sold_quantity"`
}

type cartFixture struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Status string `json:"status"`
}

type cartItemFixture struct {
	ID                  string  `json:"id"`
	CartID              string  `json:"cart_id"`
	SKUID               string  `json:"sku_id"`
	Quantity            int     `json:"quantity"`
	IntendedPriceSource string  `json:"intended_price_source"`
	FlashSaleItemID     *string `json:"flash_sale_item_id"`
	Version             int64   `json:"version"`
}

func loadFixture() (fixture, error) {
	var data fixture
	if err := json.Unmarshal(fixtureJSON, &data); err != nil {
		return fixture{}, fmt.Errorf("decode embedded fixture: %w", err)
	}
	return data, nil
}

func validateFixture(data fixture) error {
	if err := validateFixtureCount("users", len(data.Users), 2); err != nil {
		return err
	}
	for name, count := range map[string]int{
		"sessions":                  len(data.Sessions),
		"login_throttles":           len(data.LoginThrottles),
		"categories":                len(data.Categories),
		"products":                  len(data.Products),
		"skus":                      len(data.SKUs),
		"inventories":               len(data.Inventories),
		"inventory_adjustments":     len(data.InventoryAdjustments),
		"flash_sale_campaigns":      len(data.FlashSaleCampaigns),
		"flash_sale_items":          len(data.FlashSaleItems),
		"flash_sale_customer_usage": len(data.FlashSaleCustomerUsage),
		"carts":                     len(data.Carts),
		"cart_items":                len(data.CartItems),
	} {
		if err := validateFixtureCount(name, count, fixtureRecordCount); err != nil {
			return err
		}
	}

	users := make(map[string]userFixture, len(data.Users))
	for _, user := range data.Users {
		if err := requireUUID("user", user.ID); err != nil {
			return err
		}
		if _, exists := users[user.ID]; exists {
			return fmt.Errorf("duplicate user id %q", user.ID)
		}
		if user.Email == "" || user.Password == "" || user.Status != "active" {
			return fmt.Errorf("invalid user %q", user.ID)
		}
		if user.Role != "admin" && user.Role != "customer" {
			return fmt.Errorf("invalid role for user %q", user.ID)
		}
		users[user.ID] = user
	}
	if user, ok := usersByEmail(users, "admin@gmail.com"); !ok || user.Role != "admin" || user.Password != "cloudian" {
		return fmt.Errorf("missing required admin fixture")
	}
	if user, ok := usersByEmail(users, "customer@gmail.com"); !ok || user.Role != "customer" || user.Password != "cloudian" {
		return fmt.Errorf("missing required customer fixture")
	}

	categories := make(map[string]struct{}, len(data.Categories))
	for _, category := range data.Categories {
		if err := addUUID(categories, "category", category.ID); err != nil {
			return err
		}
		if strings.TrimSpace(category.Name) == "" || strings.TrimSpace(category.Slug) != category.Slug {
			return fmt.Errorf("invalid category %q", category.ID)
		}
	}
	products := make(map[string]struct{}, len(data.Products))
	for _, product := range data.Products {
		if err := addUUID(products, "product", product.ID); err != nil {
			return err
		}
		if _, ok := categories[product.CategoryID]; !ok {
			return fmt.Errorf("product %q references missing category %q", product.ID, product.CategoryID)
		}
		if product.Status != "active" && product.Status != "inactive" {
			return fmt.Errorf("invalid product status for %q", product.ID)
		}
	}
	skus := make(map[string]skuFixture, len(data.SKUs))
	for _, sku := range data.SKUs {
		if err := requireUUID("sku", sku.ID); err != nil {
			return err
		}
		if _, exists := skus[sku.ID]; exists {
			return fmt.Errorf("duplicate sku id %q", sku.ID)
		}
		if _, ok := products[sku.ProductID]; !ok {
			return fmt.Errorf("sku %q references missing product %q", sku.ID, sku.ProductID)
		}
		if sku.RegularPrice <= 0 || sku.MaxQuantityPerOrder <= 0 || !jsonObject(sku.Attributes) {
			return fmt.Errorf("invalid sku %q", sku.ID)
		}
		skus[sku.ID] = sku
	}

	inventories := make(map[string]inventoryFixture, len(data.Inventories))
	for _, inventory := range data.Inventories {
		if _, ok := skus[inventory.SKUID]; !ok {
			return fmt.Errorf("inventory references missing sku %q", inventory.SKUID)
		}
		if _, exists := inventories[inventory.SKUID]; exists || inventory.Reserved < 0 || inventory.Reserved > inventory.OnHand || inventory.Version < 0 {
			return fmt.Errorf("invalid inventory for sku %q", inventory.SKUID)
		}
		inventories[inventory.SKUID] = inventory
	}
	for _, adjustment := range data.InventoryAdjustments {
		if _, ok := inventories[adjustment.SKUID]; !ok {
			return fmt.Errorf("adjustment %q references missing inventory", adjustment.ID)
		}
		if _, ok := users[adjustment.ActorUserID]; !ok {
			return fmt.Errorf("adjustment %q references missing actor", adjustment.ID)
		}
		if adjustment.Delta == 0 || adjustment.OnHandAfter != adjustment.OnHandBefore+adjustment.Delta || strings.TrimSpace(adjustment.Reason) == "" {
			return fmt.Errorf("invalid adjustment %q", adjustment.ID)
		}
		if inventory := inventories[adjustment.SKUID]; inventory.OnHand != adjustment.OnHandAfter {
			return fmt.Errorf("adjustment %q does not match inventory", adjustment.ID)
		}
	}

	campaigns := make(map[string]struct{}, len(data.FlashSaleCampaigns))
	for _, campaign := range data.FlashSaleCampaigns {
		if err := addUUID(campaigns, "flash sale campaign", campaign.ID); err != nil {
			return err
		}
		if strings.TrimSpace(campaign.Name) == "" || campaign.Status != "draft" {
			return fmt.Errorf("invalid flash sale campaign %q", campaign.ID)
		}
	}
	items := make(map[string]flashSaleItemFixture, len(data.FlashSaleItems))
	for _, item := range data.FlashSaleItems {
		if err := requireUUID("flash sale item", item.ID); err != nil {
			return err
		}
		if _, ok := campaigns[item.CampaignID]; !ok {
			return fmt.Errorf("flash sale item %q references missing campaign", item.ID)
		}
		sku, ok := skus[item.SKUID]
		if !ok || item.SalePrice <= 0 || item.SalePrice >= sku.RegularPrice || item.QuotaTotal <= 0 || item.QuotaReserved < 0 || item.QuotaSold < 0 || item.QuotaReserved+item.QuotaSold > item.QuotaTotal || item.PurchaseLimit <= 0 {
			return fmt.Errorf("invalid flash sale item %q", item.ID)
		}
		items[item.ID] = item
	}
	for _, usage := range data.FlashSaleCustomerUsage {
		item, ok := items[usage.FlashSaleItemID]
		if !ok {
			return fmt.Errorf("usage references missing flash sale item %q", usage.FlashSaleItemID)
		}
		if _, ok := users[usage.UserID]; !ok || usage.ReservedQuantity < 0 || usage.SoldQuantity < 0 || usage.ReservedQuantity+usage.SoldQuantity > int64(item.PurchaseLimit) {
			return fmt.Errorf("invalid flash sale usage for %q", usage.FlashSaleItemID)
		}
	}

	carts := make(map[string]cartFixture, len(data.Carts))
	activeCarts := make(map[string]int)
	for _, cart := range data.Carts {
		if _, ok := users[cart.UserID]; !ok || (cart.Status != "active" && cart.Status != "archived") {
			return fmt.Errorf("invalid cart %q", cart.ID)
		}
		if _, exists := carts[cart.ID]; exists {
			return fmt.Errorf("duplicate cart id %q", cart.ID)
		}
		if cart.Status == "active" {
			activeCarts[cart.UserID]++
		}
		carts[cart.ID] = cart
	}
	customer, _ := usersByEmail(users, "customer@gmail.com")
	if activeCarts[customer.ID] != 1 {
		return fmt.Errorf("fixture must have one active customer cart")
	}
	for _, item := range data.CartItems {
		if _, ok := carts[item.CartID]; !ok {
			return fmt.Errorf("cart item %q references missing cart", item.ID)
		}
		if _, ok := skus[item.SKUID]; !ok || item.Quantity <= 0 || item.Version < 1 {
			return fmt.Errorf("invalid cart item %q", item.ID)
		}
		if item.IntendedPriceSource == "regular" && item.FlashSaleItemID != nil {
			return fmt.Errorf("regular cart item %q has flash sale item", item.ID)
		}
		if item.IntendedPriceSource == "flash_sale" {
			if item.FlashSaleItemID == nil || items[*item.FlashSaleItemID].SKUID != item.SKUID {
				return fmt.Errorf("invalid flash sale cart item %q", item.ID)
			}
		}
	}

	return nil
}

func validateFixtureCount(name string, actual, expected int) error {
	if actual != expected {
		return fmt.Errorf("%s count = %d, want %d", name, actual, expected)
	}
	return nil
}

func usersByEmail(users map[string]userFixture, email string) (userFixture, bool) {
	for _, user := range users {
		if user.Email == email {
			return user, true
		}
	}
	return userFixture{}, false
}

func addUUID(values map[string]struct{}, label, value string) error {
	if err := requireUUID(label, value); err != nil {
		return err
	}
	if _, exists := values[value]; exists {
		return fmt.Errorf("duplicate %s id %q", label, value)
	}
	values[value] = struct{}{}
	return nil
}

func requireUUID(label, value string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("invalid %s id %q", label, value)
	}
	return nil
}

func jsonObject(value json.RawMessage) bool {
	var decoded map[string]any
	return json.Unmarshal(value, &decoded) == nil && decoded != nil
}

func canSeed(environment, confirmation string) bool {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "dev", "development", "test":
		return strings.EqualFold(strings.TrimSpace(confirmation), "true")
	default:
		return false
	}
}

func hashFixturePassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), auth.DefaultPasswordCost)
	if err != nil {
		return "", fmt.Errorf("hash fixture password: %w", err)
	}
	return string(hash), nil
}

func seed(ctx context.Context, database *gorm.DB, data fixture) error {
	if database == nil {
		return errors.New("seed database is required")
	}
	if err := validateFixture(data); err != nil {
		return fmt.Errorf("validate fixture: %w", err)
	}

	passwordHashes := make(map[string]string, len(data.Users))
	for _, user := range data.Users {
		hash, err := hashFixturePassword(user.Password)
		if err != nil {
			return err
		}
		passwordHashes[user.ID] = hash
	}

	return database.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		for _, user := range data.Users {
			if err := execute(transaction, `
				INSERT INTO users (id, email, email_normalized, password_hash, role, status)
				VALUES (?, ?, lower(?), ?, ?, ?)
				ON CONFLICT (id) DO UPDATE SET
					email = EXCLUDED.email,
					email_normalized = EXCLUDED.email_normalized,
					password_hash = EXCLUDED.password_hash,
					role = EXCLUDED.role,
					status = EXCLUDED.status,
					updated_at = CURRENT_TIMESTAMP`, user.ID, user.Email, user.Email, passwordHashes[user.ID], user.Role, user.Status); err != nil {
				return fmt.Errorf("upsert user %q: %w", user.Email, err)
			}
		}
		for _, category := range data.Categories {
			if err := execute(transaction, `
				INSERT INTO categories (id, name, slug, is_visible)
				VALUES (?, ?, ?, ?)
				ON CONFLICT (id) DO UPDATE SET
					name = EXCLUDED.name, slug = EXCLUDED.slug, is_visible = EXCLUDED.is_visible,
					updated_at = CURRENT_TIMESTAMP`, category.ID, category.Name, category.Slug, category.IsVisible); err != nil {
				return fmt.Errorf("upsert category %q: %w", category.ID, err)
			}
		}
		for _, product := range data.Products {
			if err := execute(transaction, `
				INSERT INTO products (id, category_id, name, description, image_url, status)
				VALUES (?, ?, ?, ?, ?, ?)
				ON CONFLICT (id) DO UPDATE SET
					category_id = EXCLUDED.category_id, name = EXCLUDED.name,
					description = EXCLUDED.description, image_url = EXCLUDED.image_url,
					status = EXCLUDED.status, updated_at = CURRENT_TIMESTAMP`, product.ID, product.CategoryID, product.Name, product.Description, product.ImageURL, product.Status); err != nil {
				return fmt.Errorf("upsert product %q: %w", product.ID, err)
			}
		}
		for _, sku := range data.SKUs {
			if err := execute(transaction, `
				INSERT INTO skus (id, product_id, code, attributes, regular_price, max_quantity_per_order, status)
				VALUES (?, ?, ?, ?::jsonb, ?, ?, ?)
				ON CONFLICT (id) DO UPDATE SET
					product_id = EXCLUDED.product_id, code = EXCLUDED.code, attributes = EXCLUDED.attributes,
					regular_price = EXCLUDED.regular_price, max_quantity_per_order = EXCLUDED.max_quantity_per_order,
					status = EXCLUDED.status, updated_at = CURRENT_TIMESTAMP`, sku.ID, sku.ProductID, sku.Code, string(sku.Attributes), sku.RegularPrice, sku.MaxQuantityPerOrder, sku.Status); err != nil {
				return fmt.Errorf("upsert sku %q: %w", sku.ID, err)
			}
		}
		for _, inventory := range data.Inventories {
			if err := execute(transaction, `
				INSERT INTO inventories (sku_id, on_hand, reserved, version)
				VALUES (?, ?, ?, ?)
				ON CONFLICT (sku_id) DO UPDATE SET
					on_hand = EXCLUDED.on_hand, reserved = EXCLUDED.reserved,
					version = EXCLUDED.version, updated_at = CURRENT_TIMESTAMP`, inventory.SKUID, inventory.OnHand, inventory.Reserved, inventory.Version); err != nil {
				return fmt.Errorf("upsert inventory %q: %w", inventory.SKUID, err)
			}
		}
		for _, adjustment := range data.InventoryAdjustments {
			if err := execute(transaction, `
				INSERT INTO inventory_adjustments
					(id, sku_id, actor_user_id, delta, on_hand_before, on_hand_after, reason)
				VALUES (?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT (id) DO UPDATE SET
					sku_id = EXCLUDED.sku_id, actor_user_id = EXCLUDED.actor_user_id,
					delta = EXCLUDED.delta, on_hand_before = EXCLUDED.on_hand_before,
					on_hand_after = EXCLUDED.on_hand_after, reason = EXCLUDED.reason`, adjustment.ID, adjustment.SKUID, adjustment.ActorUserID, adjustment.Delta, adjustment.OnHandBefore, adjustment.OnHandAfter, adjustment.Reason); err != nil {
				return fmt.Errorf("upsert inventory adjustment %q: %w", adjustment.ID, err)
			}
		}
		for _, campaign := range data.FlashSaleCampaigns {
			if err := execute(transaction, `
				INSERT INTO flash_sale_campaigns (id, name, status, starts_at, ends_at)
				VALUES (?, ?, ?, CURRENT_TIMESTAMP + INTERVAL '1 day', CURRENT_TIMESTAMP + INTERVAL '2 days')
				ON CONFLICT (id) DO UPDATE SET
					name = EXCLUDED.name, status = EXCLUDED.status, starts_at = EXCLUDED.starts_at,
					ends_at = EXCLUDED.ends_at, updated_at = CURRENT_TIMESTAMP`, campaign.ID, campaign.Name, campaign.Status); err != nil {
				return fmt.Errorf("upsert flash sale campaign %q: %w", campaign.ID, err)
			}
		}
		for _, item := range data.FlashSaleItems {
			if err := execute(transaction, `
				INSERT INTO flash_sale_items
					(id, campaign_id, sku_id, sale_price, quota_total, quota_reserved, quota_sold, purchase_limit)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT (id) DO UPDATE SET
					campaign_id = EXCLUDED.campaign_id, sku_id = EXCLUDED.sku_id,
					sale_price = EXCLUDED.sale_price, quota_total = EXCLUDED.quota_total,
					quota_reserved = EXCLUDED.quota_reserved, quota_sold = EXCLUDED.quota_sold,
					purchase_limit = EXCLUDED.purchase_limit, updated_at = CURRENT_TIMESTAMP`, item.ID, item.CampaignID, item.SKUID, item.SalePrice, item.QuotaTotal, item.QuotaReserved, item.QuotaSold, item.PurchaseLimit); err != nil {
				return fmt.Errorf("upsert flash sale item %q: %w", item.ID, err)
			}
		}
		for _, usage := range data.FlashSaleCustomerUsage {
			if err := execute(transaction, `
				INSERT INTO flash_sale_customer_usage
					(flash_sale_item_id, user_id, reserved_quantity, sold_quantity)
				VALUES (?, ?, ?, ?)
				ON CONFLICT (flash_sale_item_id, user_id) DO UPDATE SET
					reserved_quantity = EXCLUDED.reserved_quantity, sold_quantity = EXCLUDED.sold_quantity,
					updated_at = CURRENT_TIMESTAMP`, usage.FlashSaleItemID, usage.UserID, usage.ReservedQuantity, usage.SoldQuantity); err != nil {
				return fmt.Errorf("upsert flash sale customer usage %q: %w", usage.FlashSaleItemID, err)
			}
		}
		for _, cart := range data.Carts {
			if err := execute(transaction, `
				INSERT INTO carts (id, user_id, status)
				VALUES (?, ?, ?)
				ON CONFLICT (id) DO UPDATE SET
					user_id = EXCLUDED.user_id, status = EXCLUDED.status, updated_at = CURRENT_TIMESTAMP`, cart.ID, cart.UserID, cart.Status); err != nil {
				return fmt.Errorf("upsert cart %q: %w", cart.ID, err)
			}
		}
		for _, item := range data.CartItems {
			if err := execute(transaction, `
				INSERT INTO cart_items
					(id, cart_id, sku_id, quantity, intended_price_source, flash_sale_item_id, version)
				VALUES (?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT (id) DO UPDATE SET
					cart_id = EXCLUDED.cart_id, sku_id = EXCLUDED.sku_id, quantity = EXCLUDED.quantity,
					intended_price_source = EXCLUDED.intended_price_source,
					flash_sale_item_id = EXCLUDED.flash_sale_item_id, version = EXCLUDED.version,
					updated_at = CURRENT_TIMESTAMP`, item.ID, item.CartID, item.SKUID, item.Quantity, item.IntendedPriceSource, item.FlashSaleItemID, item.Version); err != nil {
				return fmt.Errorf("upsert cart item %q: %w", item.ID, err)
			}
		}
		for _, session := range data.Sessions {
			tokenHash := sha256.Sum256([]byte(session.TokenSeed))
			csrfHash := sha256.Sum256([]byte(session.CSRFSeed))
			if err := execute(transaction, `
				INSERT INTO sessions
					(id, user_id, token_hash, csrf_secret_hash, expires_at, revoked_at, last_seen_at)
				VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP + INTERVAL '1 hour', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				ON CONFLICT (id) DO UPDATE SET
					user_id = EXCLUDED.user_id, token_hash = EXCLUDED.token_hash,
					csrf_secret_hash = EXCLUDED.csrf_secret_hash, expires_at = EXCLUDED.expires_at,
					revoked_at = EXCLUDED.revoked_at, last_seen_at = EXCLUDED.last_seen_at`, session.ID, session.UserID, tokenHash[:], csrfHash[:]); err != nil {
				return fmt.Errorf("upsert revoked session %q: %w", session.ID, err)
			}
		}
		for _, throttle := range data.LoginThrottles {
			if err := execute(transaction, `
				INSERT INTO login_throttles (email_normalized, failure_count, window_started_at)
				VALUES (?, ?, CURRENT_TIMESTAMP)
				ON CONFLICT (email_normalized) DO UPDATE SET
					failure_count = EXCLUDED.failure_count, window_started_at = EXCLUDED.window_started_at`, throttle.EmailNormalized, throttle.FailureCount); err != nil {
				return fmt.Errorf("upsert login throttle %q: %w", throttle.EmailNormalized, err)
			}
		}
		return nil
	})
}

func execute(transaction *gorm.DB, query string, arguments ...any) error {
	return transaction.Exec(query, arguments...).Error
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	log.Print("development fixture seed completed")
}

func run() error {
	applicationConfig, err := config.Load()
	if err != nil {
		return err
	}
	if !canSeed(applicationConfig.Env, os.Getenv("ALLOW_SEED")) {
		return errors.New("seed requires ENV=dev, development, or test and ALLOW_SEED=true")
	}
	fixture, err := loadFixture()
	if err != nil {
		return err
	}
	if err := validateFixture(fixture); err != nil {
		return fmt.Errorf("validate seed fixture: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), seedCommandTimeout)
	defer cancel()
	postgresDatabase, err := database.Open(ctx, applicationConfig.DatabaseUrl)
	if err != nil {
		return err
	}
	defer postgresDatabase.Close()
	return seed(ctx, postgresDatabase.ORM(), fixture)
}
