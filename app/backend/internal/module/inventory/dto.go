package inventory

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/model"
	"github.com/gin-gonic/gin"
)

const adjustmentBodyLimit = 4 << 10

type AdjustmentRequest struct {
	Delta  int64  `json:"delta"`
	Reason string `json:"reason"`
}

type InventoryResponse struct {
	SKUID     string `json:"sku_id"`
	SKUCode   string `json:"sku_code,omitempty"`
	OnHand    int64  `json:"on_hand"`
	Reserved  int64  `json:"reserved"`
	Available int64  `json:"available"`
	Version   int64  `json:"version"`
}

type InventoryPageResponse struct {
	Items      []InventoryResponse `json:"items"`
	NextCursor *string             `json:"next_cursor"`
}

type ListInventoriesQuery struct {
	Cursor string
	Limit  int
}

type InventoryAdjustmentResponse struct {
	ID           string    `json:"id"`
	SKUID        string    `json:"sku_id"`
	ActorUserID  string    `json:"actor_user_id"`
	Delta        int64     `json:"delta"`
	OnHandBefore int64     `json:"on_hand_before"`
	OnHandAfter  int64     `json:"on_hand_after"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"created_at"`
}

type InventoryAdjustmentPageResponse struct {
	Items      []InventoryAdjustmentResponse `json:"items"`
	NextCursor *string                       `json:"next_cursor"`
}

type ListAdjustmentHistoryQuery struct {
	Cursor string
	Limit  int
}

func decodeAdjustment(
	context *gin.Context,
) (AdjustmentRequest, []problem.Violation) {
	var request AdjustmentRequest

	mediaType, _, err := mime.ParseMediaType(
		context.GetHeader("Content-Type"),
	)
	if err != nil || mediaType != "application/json" {
		return request, []problem.Violation{{
			Field:  "Content-Type",
			Reason: "must be application/json",
		}}
	}

	decoder := json.NewDecoder(http.MaxBytesReader(
		context.Writer,
		context.Request.Body,
		adjustmentBodyLimit,
	))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		return request, []problem.Violation{{
			Field:  "body",
			Reason: "must contain only delta and reason",
		}}
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return request, []problem.Violation{{
			Field:  "body",
			Reason: "must contain one JSON object",
		}}
	}

	return request, request.validate()
}

func (request AdjustmentRequest) validate() []problem.Violation {
	var violations []problem.Violation

	if request.Delta == 0 {
		violations = append(violations, problem.Violation{
			Field:  "delta",
			Reason: "must not be zero",
		})
	}

	reason := strings.TrimSpace(request.Reason)
	switch {
	case reason == "":
		violations = append(violations, problem.Violation{
			Field:  "reason",
			Reason: "must contain a visible character",
		})
	case utf8.RuneCountInString(reason) > 500:
		violations = append(violations, problem.Violation{
			Field:  "reason",
			Reason: "must be at most 500 characters",
		})
	}

	return violations
}

func newInventoryResponse(inventory model.Inventory) InventoryResponse {
	return InventoryResponse{
		SKUID:     inventory.SKUID,
		OnHand:    inventory.OnHand,
		Reserved:  inventory.Reserved,
		Available: inventory.OnHand - inventory.Reserved,
		Version:   inventory.Version,
	}
}

func newInventoryAdjustmentResponse(adjustment model.InventoryAdjustment) InventoryAdjustmentResponse {
	return InventoryAdjustmentResponse{
		ID:           adjustment.ID,
		SKUID:        adjustment.SKUID,
		ActorUserID:  adjustment.ActorUserID,
		Delta:        adjustment.Delta,
		OnHandBefore: adjustment.OnHandBefore,
		OnHandAfter:  adjustment.OnHandAfter,
		Reason:       adjustment.Reason,
		CreatedAt:    adjustment.CreatedAt,
	}
}
