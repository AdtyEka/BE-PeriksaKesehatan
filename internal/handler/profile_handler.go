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

type ProfileHandler struct {
	profileService *service.ProfileService
	authRepo       *repository.AuthRepository
	jwtSecret      string
}

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

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

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

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

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

func (h *ProfileHandler) GetPersonalInfo(c *gin.Context) {
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

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

func (h *ProfileHandler) UpdatePersonalInfo(c *gin.Context) {
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	var req request.UpdatePersonalInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

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

func (h *ProfileHandler) GetHealthTargets(c *gin.Context) {
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	resp, err := h.profileService.GetHealthTargets(userID)
	if err != nil {
		utils.InternalServerError(c, "Gagal mengambil target kesehatan", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Target kesehatan berhasil diambil", resp)
}

func (h *ProfileHandler) UpdateHealthTargets(c *gin.Context) {
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	var req request.UpdateHealthTargetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

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

func (h *ProfileHandler) GetSettings(c *gin.Context) {
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

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

func (h *ProfileHandler) UpdateSettings(c *gin.Context) {
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	var req request.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

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

func (h *ProfileHandler) getUserIDFromToken(c *gin.Context) (uint, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, jwt.ErrSignatureInvalid
	}

	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	isBlacklisted, err := h.authRepo.IsTokenBlacklisted(tokenString)
	if err != nil {
		return 0, err
	}
	if isBlacklisted {
		return 0, jwt.ErrTokenExpired
	}

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

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, jwt.ErrInvalidKey
	}

	sub, ok := claims["sub"]
	if !ok {
		return 0, jwt.ErrInvalidKey
	}

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

