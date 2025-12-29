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

type ProfileHandler struct {
	profileService *service.ProfileService
	authRepo       *repository.AuthRepository
}

func NewProfileHandler(
	profileService *service.ProfileService,
	authRepo *repository.AuthRepository,
) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		authRepo:       authRepo,
	}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
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
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	var req request.UpdateProfileRequest
	var err error
	if err = c.ShouldBindJSON(&req); err != nil {
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
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
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
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	var req request.UpdatePersonalInfoRequest
	var err error
	if err = c.ShouldBindJSON(&req); err != nil {
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
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
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
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	var req request.UpdateHealthTargetsRequest
	var err error
	if err = c.ShouldBindJSON(&req); err != nil {
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
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
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
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	var req request.UpdateSettingsRequest
	var err error
	if err = c.ShouldBindJSON(&req); err != nil {
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

