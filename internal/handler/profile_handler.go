package handler

import (
	"BE-PeriksaKesehatan/config"
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/repository"
	"BE-PeriksaKesehatan/internal/service"
	"BE-PeriksaKesehatan/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ProfileHandler menangani semua request terkait profile
type ProfileHandler struct {
	profileService *service.ProfileService
	authRepo       *repository.AuthRepository
	jwtSecret      string
}

// NewProfileHandler membuat instance baru dari ProfileHandler
func NewProfileHandler(
	profileService *service.ProfileService,
	authRepo *repository.AuthRepository,
) *ProfileHandler {
	cfg := config.LoadConfig()
	secret := cfg.JWTSecret

	return &ProfileHandler{
		profileService: profileService,
		authRepo:       authRepo,
		jwtSecret:      secret,
	}
}

// ==================== GET PROFILE ====================

// GetProfile menangani GET /profile
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Panggil service
	resp, err := h.profileService.GetProfile(userID)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		utils.InternalServerError(c, "Gagal mengambil profil", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profil berhasil diambil", resp)
}

// ==================== UPDATE PROFILE ====================

// UpdateProfile menangani PUT /profile
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Bind JSON request
	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	// Panggil service
	err = h.profileService.UpdateProfile(userID, &req)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		if err.Error() == "tidak ada data untuk diupdate" {
			utils.BadRequest(c, "Tidak ada data untuk diupdate", nil)
			return
		}
		utils.InternalServerError(c, "Gagal mengupdate profil", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profil berhasil diupdate", nil)
}

// ==================== PERSONAL INFO ====================

// GetPersonalInfo menangani GET /profile/personal-info
func (h *ProfileHandler) GetPersonalInfo(c *gin.Context) {
	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Panggil service
	resp, err := h.profileService.GetPersonalInfo(userID)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		utils.InternalServerError(c, "Gagal mengambil informasi pribadi", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Informasi pribadi berhasil diambil", resp)
}

// UpdatePersonalInfo menangani PUT /profile/personal-info
func (h *ProfileHandler) UpdatePersonalInfo(c *gin.Context) {
	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Bind JSON request
	var req request.UpdatePersonalInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	// Panggil service
	err = h.profileService.UpdatePersonalInfo(userID, &req)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		if err.Error() == "format tanggal lahir tidak valid, gunakan format YYYY-MM-DD" ||
			err.Error() == "tanggal lahir tidak boleh di masa depan" {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
		utils.InternalServerError(c, "Gagal mengupdate informasi pribadi", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Informasi pribadi berhasil diupdate", nil)
}

// ==================== HEALTH TARGETS ====================

// GetHealthTargets menangani GET /profile/health-targets
func (h *ProfileHandler) GetHealthTargets(c *gin.Context) {
	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Panggil service
	resp, err := h.profileService.GetHealthTargets(userID)
	if err != nil {
		utils.InternalServerError(c, "Gagal mengambil target kesehatan", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Target kesehatan berhasil diambil", resp)
}

// UpdateHealthTargets menangani PUT /profile/health-targets
func (h *ProfileHandler) UpdateHealthTargets(c *gin.Context) {
	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Bind JSON request
	var req request.UpdateHealthTargetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	// Panggil service
	err = h.profileService.UpdateHealthTargets(userID, &req)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		utils.InternalServerError(c, "Gagal mengupdate target kesehatan", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Target kesehatan berhasil diupdate", nil)
}

// ==================== SETTINGS ====================

// GetSettings menangani GET /profile/settings
func (h *ProfileHandler) GetSettings(c *gin.Context) {
	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Panggil service
	resp, err := h.profileService.GetSettings(userID)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		utils.InternalServerError(c, "Gagal mengambil pengaturan", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pengaturan berhasil diambil", resp)
}

// UpdateSettings menangani PUT /profile/settings
func (h *ProfileHandler) UpdateSettings(c *gin.Context) {
	// Ambil user ID dari JWT token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Bind JSON request
	var req request.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	// Panggil service
	err = h.profileService.UpdateSettings(userID, &req)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		if err.Error() == "tidak ada data untuk diupdate" {
			utils.BadRequest(c, "Tidak ada data untuk diupdate", nil)
			return
		}
		utils.InternalServerError(c, "Gagal mengupdate pengaturan", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pengaturan berhasil diupdate", nil)
}

// ==================== HELPER FUNCTIONS ====================

// getUserIDFromToken mengambil user ID dari JWT token di header Authorization
func (h *ProfileHandler) getUserIDFromToken(c *gin.Context) (uint, error) {
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

