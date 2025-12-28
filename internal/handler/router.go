package handler

import (
	"BE-PeriksaKesehatan/internal/repository"
	"BE-PeriksaKesehatan/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRouter mengkonfigurasi semua route aplikasi
func SetupRouter(userRepo *repository.UserRepository) *gin.Engine {
	// Setup Gin dengan default middleware
	router := gin.Default()

	// Setup repository
	healthDataRepo := repository.NewHealthDataRepository(userRepo.GetDB())

	// Setup service
	healthDataService := service.NewHealthDataService(healthDataRepo)

	// Setup handler
	authHandler := NewAuthHandler(userRepo)
	healthDataHandler := NewHealthDataHandler(healthDataService)

	// API Routes
	api := router.Group("/api")
	{
		// Route autentikasi (register & login)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Route data kesehatan (memerlukan autentikasi)
		health := api.Group("/health")
		{
			health.POST("/data", healthDataHandler.CreateHealthData)
			health.GET("/data", healthDataHandler.GetHealthDataByUserID)
			health.GET("/history", healthDataHandler.GetHealthHistory)                    // Endpoint riwayat kesehatan
			health.GET("/history/download", healthDataHandler.DownloadHealthReport)       // Endpoint download laporan
		}
	}

	return router
}

