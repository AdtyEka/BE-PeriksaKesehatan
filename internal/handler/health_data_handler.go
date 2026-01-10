package handler

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/repository"
	"BE-PeriksaKesehatan/internal/service"
	"BE-PeriksaKesehatan/pkg/middleware"
	"BE-PeriksaKesehatan/pkg/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	timezoneUtils "BE-PeriksaKesehatan/pkg/utils"
)

// HealthDataHandler menangani semua request terkait data kesehatan
type HealthDataHandler struct {
	healthDataService *service.HealthDataService
	authRepo          *repository.AuthRepository
}

// NewHealthDataHandler membuat instance baru dari HealthDataHandler
func NewHealthDataHandler(healthDataService *service.HealthDataService, authRepo *repository.AuthRepository) *HealthDataHandler {
	return &HealthDataHandler{
		healthDataService: healthDataService,
		authRepo:          authRepo,
	}
}

// CreateHealthData menangani request input data kesehatan
func (h *HealthDataHandler) CreateHealthData(c *gin.Context) {
	var req request.HealthDataRequest

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

	// Panggil service untuk membuat data kesehatan
	resp, err := h.healthDataService.CreateHealthData(userID, &req)
	if err != nil {
		// Cek apakah error adalah validasi (termasuk validasi nullable-aware)
		// Error validasi biasanya dimulai dengan nama field atau "minimal"
		errMsg := err.Error()
		if strings.Contains(errMsg, "harus berada dalam range") ||
			strings.Contains(errMsg, "wajib diisi") ||
			strings.Contains(errMsg, "minimal satu") ||
			strings.Contains(errMsg, "bersamaan") {
			utils.BadRequest(c, "Validasi gagal", errMsg)
			return
		}
		utils.InternalServerError(c, "Gagal menyimpan data kesehatan", errMsg)
		return
	}

	// Response sukses
	utils.SuccessResponse(c, http.StatusCreated, "Data kesehatan berhasil disimpan", resp)
}

// GetHealthDataByUserID menangani request untuk mendapatkan data kesehatan terbaru user
// Mengembalikan 1 record terbaru (inkremental) yang berisi semua data kesehatan user
func (h *HealthDataHandler) GetHealthDataByUserID(c *gin.Context) {
	// Ambil user ID dari context (sudah divalidasi oleh middleware)
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Panggil service untuk mendapatkan data kesehatan terbaru
	healthData, err := h.healthDataService.GetHealthDataByUserID(userID)
	if err != nil {
		utils.InternalServerError(c, "Gagal mengambil data kesehatan", err.Error())
		return
	}

	// Jika tidak ada data, kembalikan null dengan pesan yang jelas
	if healthData == nil {
		utils.SuccessResponse(c, http.StatusOK, "Belum ada data kesehatan", nil)
		return
	}

	// Build response dari entity ke response DTO
	resp := &response.HealthDataResponse{
		ID:         healthData.ID,
		UserID:     healthData.UserID,
		Systolic:   healthData.Systolic,
		Diastolic:  healthData.Diastolic,
		BloodSugar: healthData.BloodSugar,
		Weight:     healthData.Weight,
		Height:     healthData.HeightCM,
		HeartRate:  healthData.HeartRate,
		Activity:   healthData.Activity,
		CreatedAt:  timezoneUtils.ToJakarta(healthData.CreatedAt),
	}

	// Response sukses
	utils.SuccessResponse(c, http.StatusOK, "Data kesehatan berhasil diambil", resp)
}

// GetHealthHistory menangani request untuk mendapatkan riwayat kesehatan dengan filter
func (h *HealthDataHandler) GetHealthHistory(c *gin.Context) {
	// Ambil user ID dari context (sudah divalidasi oleh middleware)
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Bind query parameters atau JSON body ke struct
	var req request.HealthHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// Jika query binding gagal, coba bind JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			// Jika keduanya gagal, gunakan default
			req.TimeRange = "7days"
		}
	}

	// Set default jika kosong
	if req.TimeRange == "" {
		req.TimeRange = "7days"
	}

	// Panggil service untuk mendapatkan riwayat kesehatan (struktur internal)
	// Ini digunakan untuk mendapatkan TrendCharts yang sudah dihitung
	resp, err := h.healthDataService.GetHealthHistory(userID, &req)
	if err != nil {
		if err.Error() == "start_date dan end_date wajib diisi untuk custom range" {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
		utils.InternalServerError(c, "Gagal mengambil riwayat kesehatan", err.Error())
		return
	}

	// Mapping ke struktur response API baru menggunakan mapping layer
	// Mapping layer akan mengambil data untuk semua periode (7Days, 1Month, 3Months)
	// dan melakukan transform tanpa mengubah logic perhitungan
	apiResp, err := h.healthDataService.MapHealthHistoryToAPIResponse(userID, &req, resp)
	if err != nil {
		utils.InternalServerError(c, "Gagal memproses data riwayat kesehatan", err.Error())
		return
	}

	// Response sukses dengan bentuk JSON baru:
	// {
	//   "summary": {
	//     "start_date": "2025-01-01",
	//     "end_date": "2025-01-31",
	//     "7Days": { ... },
	//     "1Month": { ..., "weeks": [ ... ] },
	//     "3Months": { ..., "weeks": [ ... ] }
	//   },
	//   "trend_charts": { ... (tetap sama) ... },
	//   "reading_history": {
	//     "7Days": [ ... ],
	//     "1Month": { "start_date": "...", "end_date": "...", "records": [ ... ] },
	//     "3Months": { "start_date": "...", "end_date": "...", "records": [ ... ] }
	//   }
	// }
	utils.SuccessResponse(c, http.StatusOK, "Riwayat kesehatan berhasil diambil", apiResp)
}

// DownloadHealthReport menangani request untuk mengunduh laporan riwayat kesehatan dalam format PDF
func (h *HealthDataHandler) DownloadHealthReport(c *gin.Context) {
	// Ambil user ID dari context (sudah divalidasi oleh middleware)
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Bind query parameters untuk filter
	var req request.HealthHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// Jika query binding gagal, gunakan default
		req.TimeRange = "7days"
	}

	// Set default jika kosong
	if req.TimeRange == "" {
		req.TimeRange = "7days"
	}

	// Generate laporan PDF
	fileBuffer, filename, err := h.healthDataService.GenerateReportPDF(userID, &req)
	if err != nil {
		if err.Error() == "start_date dan end_date wajib diisi untuk custom range" {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
		utils.InternalServerError(c, "Gagal membuat laporan PDF", err.Error())
		return
	}

	// Set header untuk download PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Kirim file
	c.Data(http.StatusOK, "application/pdf", fileBuffer.Bytes())
}

