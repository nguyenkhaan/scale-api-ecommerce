package middleware

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"cloudian/cloudian-scale/internal/errors"
	"cloudian/cloudian-scale/internal/response"
)

// RecoveryMiddleware bắt mọi panic và chuyển thành response JSON lỗi 500 chuẩn hóa
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				reqID := response.GetRequestID(c)
				stack := string(debug.Stack())

				log.Printf("[PANIC RECOVERED] req_id=%s panic=%v\n%s", reqID, r, stack)

				appErr := errors.ErrInternal(fmt.Errorf("%v", r))
				response.Error(c, appErr)
				c.Abort()
			}
		}()

		c.Next()
	}
}
