package flashsale

import (
	"strconv"
	"strings"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"github.com/gin-gonic/gin"
)

var (
	publicStatuses = map[string]bool{"scheduled": true, "active": true, "ended": true}
	adminStatuses  = map[string]bool{"draft": true, "scheduled": true, "active": true, "stopped": true, "ended": true}
)

func decodeListQuery(ctx *gin.Context, isPublic bool) (ListFlashSalesQuery, []problem.Violation) {
	query := ListFlashSalesQuery{Cursor: ctx.Query("cursor"), Limit: defaultPageSize}
	var violations []problem.Violation

	if rawLimit, exists := ctx.GetQuery("limit"); exists {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit < 1 || limit > 100 {
			violations = append(violations, violation("limit", "must be an integer between 1 and 100"))
		} else {
			query.Limit = limit
		}
	}

	if status, exists := ctx.GetQuery("status"); exists {
		status = strings.TrimSpace(status)
		allowed := adminStatuses
		if isPublic {
			allowed = publicStatuses
		}
		if allowed[status] {
			query.Status = &status
		} else {
			violations = append(violations, violation("status", "is not an allowed status value"))
		}
	}
	return query, violations
}
