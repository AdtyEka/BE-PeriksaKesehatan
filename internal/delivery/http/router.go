package http

import (
	"BE-PeriksaKesehatan/internal/repository"

	"github.com/gin-gonic/gin"
)

// SetupRouter mengkonfigurasi semua route aplikasi
func SetupRouter(userRepo *repository.UserRepository) *gin.Engine {
	// Setup Gin dengan default middleware
	router := gin.Default()

	// Setup handler
	authHandler := NewAuthHandler(userRepo)

	// API Routes
	api := router.Group("/api/v1")
	{
		// Hanya route autentikasi (register & login)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}
	}

	return router
}

