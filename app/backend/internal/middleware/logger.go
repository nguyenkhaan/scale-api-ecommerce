package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"cloudian/cloudian-scale/internal/response"
)

// LoggerMiddleware ghi log có cấu trúc cho mỗi request đến backend
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		reqID := response.GetRequestID(c)

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		log.Printf("[API] | %3d | %13v | %15s | %-7s %s | req_id=%s\n",
			statusCode,
			latency,
			clientIP,
			method,
			path,
			reqID,
		)
	}
}
