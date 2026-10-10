package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/router"
)

func main() {
	// Nạp .env khi chạy local; môi trường deploy truyền biến trực tiếp nên thiếu file không phải lỗi
	_ = godotenv.Load()

	cfg := config.LoadConfig()

	r := router.SetupRouter(cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Chạy HTTP server trong goroutine độc lập
	go func() {
		log.Printf("==========================================")
		log.Printf(" SaleScale Backend API is running!")
		log.Printf(" Port:       %s", cfg.Port)
		log.Printf(" Env:        %s", cfg.Env)
		log.Printf(" Health:     http://localhost:%s%s/health", cfg.Port, router.APIPrefix)
		log.Printf(" Ping:       http://localhost:%s%s/ping", cfg.Port, router.APIPrefix)
		log.Printf("==========================================")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Khởi động server thất bại: %v", err)
		}
	}()

	// Lắng nghe tín hiệu dừng (Graceful Shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Đang tiến hành dừng server an toàn (Graceful shutdown)...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server buộc phải dừng với lỗi: %v", err)
	}

	log.Println("Backend server đã dừng hoàn tất.")
}
