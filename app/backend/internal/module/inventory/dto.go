package inventory

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
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
	OnHand    int64  `json:"on_hand"`
	Reserved  int64  `json:"reserved"`
	Available int64  `json:"available"`
	Version   int64  `json:"version"`
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