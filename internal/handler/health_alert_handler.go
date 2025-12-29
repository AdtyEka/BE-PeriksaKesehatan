package handler

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
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
func (h *HealthAlertHandler) CheckHealthAlerts(c *gin.Context) {
	var req request.HealthAlertRequest

	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	// Ambil user ID dari context (sudah divalidasi oleh middleware)
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Panggil service untuk memeriksa alerts
	resp, err := h.healthAlertService.CheckHealthAlerts(userID, &req)
	if err != nil {
		// Cek apakah error adalah validasi
		if err.Error() == "systolic harus valid (1-300 mmHg)" ||
			err.Error() == "diastolic harus valid (1-200 mmHg)" ||
			err.Error() == "blood_sugar harus valid (1-1000 mg/dL)" ||
			err.Error() == "weight harus valid (1-500 kg)" {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
		utils.InternalServerError(c, "Gagal memeriksa health alerts", err.Error())
		return
	}

	// Response sukses
	utils.SuccessResponse(c, http.StatusOK, "Health alerts berhasil diperiksa", resp)
}

