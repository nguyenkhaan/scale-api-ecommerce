package middleware

import (
	"fmt"
	"regexp"
	"uuid"

	"github.com/gin-gonic/gin"
)

const (
	maxLength       = 128
	RequestIdHeader = "X-Request-ID"
	RequestIdKey    = "request_id"
)

var requestIdPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func isValidRequestId(requestID string) bool {
	if requestID == "" {
		return false
	}
	if len(requestID) > maxLength {
		return false
	}
	return requestIdPattern.MatchString(requestID)
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		requestId := context.GetHeader(RequestIdHeader)
		if !isValidRequestId(requestId) {
			requestId = uuid.NewV4().String()
		}
		context.Set(RequestIdKey, requestId)
		context.Header(RequestIdHeader, requestId)
		context.Next()
	}
}

func RequestLogFormatter(parameters gin.LogFormatterParams) string {
	requestID, _ := parameters.Keys[RequestIdKey].(string)
	return fmt.Sprintf(
		"%s status=%d latency=%s method=%s path=%q request_id=%s\n",
		parameters.TimeStamp.UTC().Format("2006-01-02T15:04:05Z07:00"),
		parameters.StatusCode,
		parameters.Latency,
		parameters.Method,
		parameters.Path,
		requestID,
	)
}
