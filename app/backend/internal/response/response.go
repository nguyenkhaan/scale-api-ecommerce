package response

import (
	"time"

	"github.com/gin-gonic/gin"

	"cloudian/cloudian-scale/internal/errors"
)

// Response cấu trúc JSON phản hồi chuẩn mực toàn hệ thống
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    MetaInfo    `json:"meta"`
}

// ErrorInfo cấu trúc chi tiết lỗi trả về cho client
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// MetaInfo chứa thông tin bổ sung cho mỗi phản hồi
type MetaInfo struct {
	RequestID  string          `json:"request_id"`
	Timestamp  string          `json:"timestamp"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
}

// PaginationInfo chứa thông số phân trang danh sách
type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// Success gửi JSON phản hồi thành công (HTTP 2xx)
func Success(c *gin.Context, statusCode int, data interface{}) {
	reqID := GetRequestID(c)
	res := Response{
		Success: true,
		Data:    data,
		Meta: MetaInfo{
			RequestID: reqID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	c.JSON(statusCode, res)
}

// SuccessWithPagination gửi JSON phản hồi danh sách kèm thông tin phân trang
func SuccessWithPagination(c *gin.Context, statusCode int, data interface{}, pagination PaginationInfo) {
	reqID := GetRequestID(c)
	res := Response{
		Success: true,
		Data:    data,
		Meta: MetaInfo{
			RequestID:  reqID,
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			Pagination: &pagination,
		},
	}
	c.JSON(statusCode, res)
}

// Error gửi JSON phản hồi lỗi chuẩn hóa dựa trên AppError
func Error(c *gin.Context, appErr *errors.AppError) {
	if appErr == nil {
		appErr = errors.ErrInternal(nil)
	}
	reqID := GetRequestID(c)
	res := Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
		Meta: MetaInfo{
			RequestID: reqID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	c.JSON(appErr.StatusCode, res)
}

// AbortWithError ghi nhận AppError và dừng pipeline của Gin
func AbortWithError(c *gin.Context, appErr *errors.AppError) {
	Error(c, appErr)
	c.Abort()
}

// GetRequestID lấy request_id từ context hoặc header
func GetRequestID(c *gin.Context) string {
	if val, exists := c.Get("request_id"); exists {
		if id, ok := val.(string); ok && id != "" {
			return id
		}
	}
	headerID := c.GetHeader("X-Request-ID")
	if headerID != "" {
		return headerID
	}
	return "unknown"
}
