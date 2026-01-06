package handler

import (
	"BE-PeriksaKesehatan/internal/repository"
	"BE-PeriksaKesehatan/internal/service"
	"BE-PeriksaKesehatan/pkg/middleware"
	"BE-PeriksaKesehatan/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthAlertHandler menangani semua request terkait health alerts
type HealthAlertHandler struct {
	healthAlertService *service.HealthAlertService
	authRepo           *repository.AuthRepository
}

// NewHealthAlertHandler membuat instance baru dari HealthAlertHandler
func NewHealthAlertHandler(healthAlertService *service.HealthAlertService, authRepo *repository.AuthRepository) *HealthAlertHandler {
	return &HealthAlertHandler{
		healthAlertService: healthAlertService,
		authRepo:           authRepo,
	}
}

// CheckHealthAlerts menangani request untuk memeriksa health alerts
// Endpoint GET yang mengambil data kesehatan terbaru dari database
func (h *HealthAlertHandler) CheckHealthAlerts(c *gin.Context) {
	// Ambil user ID dari context (sudah divalidasi oleh middleware)
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Panggil service untuk memeriksa alerts (tanpa request body)
	resp, err := h.healthAlertService.CheckHealthAlerts(userID)
	if err != nil {
		utils.InternalServerError(c, "Gagal memeriksa health alerts", err.Error())
		return
	}

	// Response sukses
	utils.SuccessResponse(c, http.StatusOK, "Health alerts berhasil diperiksa", resp)
}

