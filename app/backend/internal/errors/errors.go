package errors

import (
	"fmt"
	"net/http"
)

// Danh sách mã lỗi chuẩn hóa toàn hệ thống (SCREAMING_SNAKE_CASE)
const (
	CodeBadRequest             = "BAD_REQUEST"
	CodeValidationFailed       = "VALIDATION_FAILED"
	CodeUnauthorized           = "UNAUTHORIZED"
	CodeForbidden              = "FORBIDDEN"
	CodeNotFound               = "NOT_FOUND"
	CodeMethodNotAllowed       = "METHOD_NOT_ALLOWED"
	CodeConflict               = "CONFLICT"
	CodeIdempotencyConflict    = "IDEMPOTENCY_CONFLICT"
	CodeOutOfStock             = "OUT_OF_STOCK"
	CodeFlashSaleLimitExceeded = "FLASH_SALE_LIMIT_EXCEEDED"
	CodeFlashSaleNotActive     = "FLASH_SALE_NOT_ACTIVE"
	CodeReservationExpired     = "RESERVATION_EXPIRED"
	CodePaymentFailed          = "PAYMENT_FAILED"
	CodeRateLimited            = "RATE_LIMITED"
	CodeInternalError          = "INTERNAL_SERVER_ERROR"
)

// AppError biểu diễn lỗi có cấu trúc trong hệ thống backend
type AppError struct {
	StatusCode int         `json:"-"`
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	Err        error       `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAppError tạo một AppError mới
func NewAppError(statusCode int, code, message string, details ...interface{}) *AppError {
	var d interface{}
	if len(details) > 0 {
		d = details[0]
	}
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Details:    d,
	}
}

// ErrBadRequest trả về lỗi 400 Bad Request
func ErrBadRequest(message string, details ...interface{}) *AppError {
	return NewAppError(http.StatusBadRequest, CodeBadRequest, message, details...)
}

// ErrValidation trả về lỗi 422 Validation Failed kèm chi tiết vi phạm
func ErrValidation(message string, details interface{}) *AppError {
	return &AppError{
		StatusCode: http.StatusUnprocessableEntity,
		Code:       CodeValidationFailed,
		Message:    message,
		Details:    details,
	}
}

// ErrUnauthorized trả về lỗi 401 Unauthorized
func ErrUnauthorized(message string) *AppError {
	if message == "" {
		message = "Yêu cầu đăng nhập hoặc phiên làm việc đã hết hạn"
	}
	return NewAppError(http.StatusUnauthorized, CodeUnauthorized, message)
}

// ErrForbidden trả về lỗi 403 Forbidden
func ErrForbidden(message string) *AppError {
	if message == "" {
		message = "Bạn không có quyền thực hiện thao tác này"
	}
	return NewAppError(http.StatusForbidden, CodeForbidden, message)
}

// ErrNotFound trả về lỗi 404 Not Found
func ErrNotFound(resource string) *AppError {
	message := fmt.Sprintf("Không tìm thấy tài nguyên %s yêu cầu", resource)
	return NewAppError(http.StatusNotFound, CodeNotFound, message)
}

// ErrMethodNotAllowed trả về lỗi 405 Method Not Allowed
func ErrMethodNotAllowed(method string) *AppError {
	message := fmt.Sprintf("Phương thức HTTP %s không được hỗ trợ", method)
	return NewAppError(http.StatusMethodNotAllowed, CodeMethodNotAllowed, message)
}

// ErrConflict trả về lỗi 409 Conflict
func ErrConflict(message string) *AppError {
	return NewAppError(http.StatusConflict, CodeConflict, message)
}

// ErrInternal trả về lỗi 500 Internal Server Error và gói kèm lỗi gốc
func ErrInternal(err error) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       CodeInternalError,
		Message:    "Đã xảy ra lỗi nội bộ hệ thống. Vui lòng thử lại sau.",
		Err:        err,
	}
}
