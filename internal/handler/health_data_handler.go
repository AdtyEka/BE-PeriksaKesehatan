package handler

import (
	"BE-PeriksaKesehatan/config"
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/repository"
	"BE-PeriksaKesehatan/internal/service"
	"BE-PeriksaKesehatan/pkg/utils"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// HealthDataHandler menangani semua request terkait data kesehatan
type HealthDataHandler struct {
	healthDataService *service.HealthDataService
	authRepo          *repository.AuthRepository
	jwtSecret         string
}

// NewHealthDataHandler membuat instance baru dari HealthDataHandler
func NewHealthDataHandler(healthDataService *service.HealthDataService, authRepo *repository.AuthRepository) *HealthDataHandler {
	// Ambil secret dari environment (sama seperti AuthHandler)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// fallback: coba ambil dari config untuk memastikan kalau LoadConfig sudah dijalankan
		cfg := config.LoadConfig()
		secret = cfg.JWTSecret
	}

	return &HealthDataHandler{
		healthDataService: healthDataService,
		authRepo:          authRepo,
		jwtSecret:         secret,
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

	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Panggil service untuk membuat data kesehatan
	resp, err := h.healthDataService.CreateHealthData(userID, &req)
	if err != nil {
		// Cek apakah error adalah validasi
		if err.Error() == "systolic harus berada dalam range 90-180 mmHg" ||
			err.Error() == "diastolic harus berada dalam range 60-120 mmHg" ||
			err.Error() == "blood_sugar harus berada dalam range 60-300 mg/dL" ||
			err.Error() == "weight harus berada dalam range 20-200 kg" ||
			err.Error() == "heart_rate harus berada dalam range 40-180 bpm" {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
		utils.InternalServerError(c, "Gagal menyimpan data kesehatan", err.Error())
		return
	}

	// Response sukses
	utils.SuccessResponse(c, http.StatusCreated, "Data kesehatan berhasil disimpan", resp)
}

// GetHealthDataByUserID menangani request untuk mendapatkan riwayat data kesehatan user
func (h *HealthDataHandler) GetHealthDataByUserID(c *gin.Context) {
	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Panggil service untuk mendapatkan data kesehatan
	healthDataList, err := h.healthDataService.GetHealthDataByUserID(userID)
	if err != nil {
		utils.InternalServerError(c, "Gagal mengambil data kesehatan", err.Error())
		return
	}

	// Response sukses
	utils.SuccessResponse(c, http.StatusOK, "Data kesehatan berhasil diambil", healthDataList)
}

// GetHealthHistory menangani request untuk mendapatkan riwayat kesehatan dengan filter
func (h *HealthDataHandler) GetHealthHistory(c *gin.Context) {
	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
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

	// Panggil service untuk mendapatkan riwayat kesehatan
	resp, err := h.healthDataService.GetHealthHistory(userID, &req)
	if err != nil {
		if err.Error() == "start_date dan end_date wajib diisi untuk custom range" {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
		utils.InternalServerError(c, "Gagal mengambil riwayat kesehatan", err.Error())
		return
	}

	// Response sukses
	utils.SuccessResponse(c, http.StatusOK, "Riwayat kesehatan berhasil diambil", resp)
}

// DownloadHealthReport menangani request untuk mengunduh laporan riwayat kesehatan
func (h *HealthDataHandler) DownloadHealthReport(c *gin.Context) {
	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Ambil format dari query parameter (default: csv)
	format := c.DefaultQuery("format", "csv")
	if format != "csv" && format != "json" {
		utils.BadRequest(c, "Format tidak valid", "Format harus 'csv' atau 'json'")
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

	var fileBuffer *bytes.Buffer
	var filename string

	// Generate laporan berdasarkan format
	if format == "csv" {
		fileBuffer, filename, err = h.healthDataService.GenerateReportCSV(userID, &req)
		if err != nil {
			if err.Error() == "start_date dan end_date wajib diisi untuk custom range" {
				utils.BadRequest(c, "Validasi gagal", err.Error())
				return
			}
			utils.InternalServerError(c, "Gagal membuat laporan CSV", err.Error())
			return
		}
		// Set header untuk download CSV
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	} else {
		fileBuffer, filename, err = h.healthDataService.GenerateReportJSON(userID, &req)
		if err != nil {
			if err.Error() == "start_date dan end_date wajib diisi untuk custom range" {
				utils.BadRequest(c, "Validasi gagal", err.Error())
				return
			}
			utils.InternalServerError(c, "Gagal membuat laporan JSON", err.Error())
			return
		}
		// Set header untuk download JSON
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	}

	// Kirim file
	c.Data(http.StatusOK, c.GetHeader("Content-Type"), fileBuffer.Bytes())
}

// getUserIDFromToken mengambil user ID dari JWT token di header Authorization
func (h *HealthDataHandler) getUserIDFromToken(c *gin.Context) (uint, error) {
	// Ambil token dari header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, jwt.ErrSignatureInvalid
	}

	// Parse token (format: "Bearer <token>")
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// Cek apakah token sudah di-blacklist
	isBlacklisted, err := h.authRepo.IsTokenBlacklisted(tokenString)
	if err != nil {
		return 0, err
	}
	if isBlacklisted {
		return 0, jwt.ErrTokenExpired // Menggunakan ErrTokenExpired untuk menunjukkan token tidak valid
	}

	// Parse dan validasi token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validasi bahwa signing method adalah HS256 (sama seperti saat login)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, jwt.ErrSignatureInvalid
	}

	// Ambil claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, jwt.ErrInvalidKey
	}

	// Validasi expiry time (jwt library sudah validasi, tapi kita pastikan)
	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			return 0, jwt.ErrTokenExpired
		}
	}

	// Ambil user ID dari claims
	sub, ok := claims["sub"]
	if !ok {
		return 0, jwt.ErrInvalidKey
	}

	// Convert ke uint (JWT library mengkonversi angka menjadi float64 saat parsing JSON)
	userIDFloat, ok := sub.(float64)
	if !ok {
		// Coba sebagai string (untuk kompatibilitas)
		userIDStr, ok := sub.(string)
		if !ok {
			return 0, jwt.ErrInvalidKey
		}
		userIDUint, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint(userIDUint), nil
	}

	return uint(userIDFloat), nil
}

