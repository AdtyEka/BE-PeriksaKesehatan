package handler

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/repository"
	"BE-PeriksaKesehatan/internal/service"
	"BE-PeriksaKesehatan/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// HealthAlertHandler menangani semua request terkait health alerts
type HealthAlertHandler struct {
	healthAlertService *service.HealthAlertService
	authRepo           *repository.AuthRepository
	jwtSecret          string
}

// NewHealthAlertHandler membuat instance baru dari HealthAlertHandler
func NewHealthAlertHandler(healthAlertService *service.HealthAlertService, authRepo *repository.AuthRepository, jwtSecret string) *HealthAlertHandler {
	return &HealthAlertHandler{
		healthAlertService: healthAlertService,
		authRepo:           authRepo,
		jwtSecret:          jwtSecret,
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

	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
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

// getUserIDFromToken mengambil user ID dari JWT token di header Authorization
func (h *HealthAlertHandler) getUserIDFromToken(c *gin.Context) (uint, error) {
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
		return 0, jwt.ErrTokenExpired
	}

	// Parse dan validasi token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
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

	// Validasi expiry time
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

	// Convert ke uint
	userIDFloat, ok := sub.(float64)
	if !ok {
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

