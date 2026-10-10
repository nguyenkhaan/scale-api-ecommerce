package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/response"
)

var startTime = time.Now()

// HealthHandler quản lý endpoint kiểm tra trạng thái sức khỏe hệ thống
type HealthHandler struct {
	cfg *config.Config
}

// NewHealthHandler khởi tạo HealthHandler
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

// HealthCheck trả về thông tin trạng thái hoạt động của Backend
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	uptime := time.Since(startTime).Truncate(time.Second).String()

	data := gin.H{
		"service":     "salescale-backend",
		"status":      "UP",
		"environment": h.cfg.Env,
		"uptime":      uptime,
	}

	response.Success(c, http.StatusOK, data)
}

// Ping trả về phản hồi đơn giản để đo liveness
func (h *HealthHandler) Ping(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"message": "pong",
	})
}
