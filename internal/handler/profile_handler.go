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

func (h *ProfileHandler) CreateProfile(c *gin.Context) {
	// Ambil user_id dari JWT context
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Parse multipart form
	var req request.CreateProfileRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	// Validasi field wajib
	if req.Name == "" {
		utils.BadRequest(c, "Validasi gagal", "name wajib diisi")
		return
	}
	if req.Email == "" {
		utils.BadRequest(c, "Validasi gagal", "email wajib diisi")
		return
	}

	// Validasi email format sudah dilakukan oleh binding tag "email"
	// Double check tidak diperlukan karena gin sudah memvalidasi

	// Validasi optional fields
	if req.Weight != nil && *req.Weight <= 0 {
		utils.BadRequest(c, "Validasi gagal", "weight harus lebih besar dari 0")
		return
	}
	if req.Height != nil && *req.Height <= 0 {
		utils.BadRequest(c, "Validasi gagal", "height harus lebih besar dari 0")
		return
	}
	if req.Age != nil && *req.Age <= 0 {
		utils.BadRequest(c, "Validasi gagal", "age harus lebih besar dari 0")
		return
	}

	// Handle file upload
	var photoURL *string
	fileHeader, err := c.FormFile("photo")
	if err != nil {
		// Error bisa berarti file tidak ada (opsional) atau error lain
		// Cek apakah error adalah "no such file" (file tidak dikirim, ini OK karena optional)
		if err.Error() != "http: no such file" && err.Error() != "multipart: NextPart: EOF" {
			// Error selain "file tidak ada" berarti error lain
			utils.BadRequest(c, "Gagal membaca file", err.Error())
			return
		}
		// File tidak dikirim, ini OK karena optional
		fileHeader = nil
	}

	// Validasi dan upload file jika ada
	if fileHeader != nil {
		// Validasi file
		if err := utils.ValidateImageFile(fileHeader); err != nil {
			// Cek jenis error untuk status code yang tepat
			if err.Error() == "ukuran file terlalu besar, maksimal 2 MB" {
				utils.ErrorResponse(c, http.StatusRequestEntityTooLarge, "File terlalu besar", err.Error())
				return
			}
			if err.Error() == "tipe file tidak didukung, hanya jpg, jpeg, png, dan webp yang diizinkan" ||
				err.Error() == "ekstensi file tidak didukung, hanya .jpg, .jpeg, .png, dan .webp yang diizinkan" {
				utils.ErrorResponse(c, http.StatusUnsupportedMediaType, "Tipe file tidak didukung", err.Error())
				return
			}
			utils.BadRequest(c, "Validasi file gagal", err.Error())
			return
		}

		// Upload file
		uploadedPath, err := utils.UploadProfileImage(fileHeader, userID)
		if err != nil {
			utils.InternalServerError(c, "Gagal mengupload foto", err.Error())
			return
		}

		// Simpan path untuk disimpan di database
		photoURL = &uploadedPath
	}

	// Panggil service
	resp, err := h.profileService.CreateProfile(userID, &req, photoURL)
	if err != nil {
		// Rollback: hapus file jika sudah diupload
		if photoURL != nil {
			_ = utils.DeleteProfileImage(*photoURL)
		}

		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		if err.Error() == "profile sudah ada" {
			utils.ErrorResponse(c, http.StatusConflict, "Profile sudah ada", nil)
			return
		}
		utils.InternalServerError(c, "Gagal membuat profil", err.Error())
		return
	}

	// Response sukses
	utils.SuccessResponse(c, http.StatusCreated, "Profil berhasil dibuat", resp)
}

