package main

import (
	"GAMERS-BE/internal/monitor/handler"
	"GAMERS-BE/internal/monitor/metrics"
	"GAMERS-BE/internal/monitor/middleware"
	"GAMERS-BE/internal/user/application"
	"GAMERS-BE/internal/user/infra/persistence"
	"GAMERS-BE/internal/user/presentation"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 모니터링 초기화
	metricsCollector := metrics.NewMetricsCollector(100) // 최대 100개의 메트릭 히스토리 저장
	requestMonitor := middleware.NewRequestMonitor(
		100, // 최대 100개의 요청 기록 저장
		// Body capture 비활성화로 성능 향상 (필요시 활성화)
		// middleware.WithBodyCapture(true),
		// 특정 경로만 모니터링하려면 사용
		// middleware.WithPathFilter("/api/users"),
	)

	// 5초마다 자동으로 메트릭 수집 (2초에서 5초로 변경하여 부하 감소)
	metricsCollector.StartAutoCollect(5 * time.Second)

	// 기존 컴포넌트 초기화
	userRepository := persistence.NewInMemoryUserRepository()
	userService := application.NewUserService(userRepository)
	userController := presentation.NewUserController(userService)

	router := gin.Default()

	// 모니터링 미들웨어 등록 (모든 요청을 추적)
	router.Use(requestMonitor.Middleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// User routes
	userController.RegisterRoutes(router)

	// 모니터링 대시보드 등록
	monitorHandler := handler.NewMonitorHandler(metricsCollector, requestMonitor)
	monitorHandler.RegisterRoutes(router)

	log.Println("===========================================")
	log.Println("🎮 GAMERS API Server Starting")
	log.Println("===========================================")
	log.Println("Server:          http://localhost:8080")
	log.Println("Monitor:         http://localhost:8080/monitor")
	log.Println("Health Check:    http://localhost:8080/health")
	log.Println("===========================================")

	if err := router.Run(os.Getenv("port")); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
