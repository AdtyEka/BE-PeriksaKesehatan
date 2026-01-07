package handler

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/service"
	"BE-PeriksaKesehatan/pkg/middleware"
	"BE-PeriksaKesehatan/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileService *service.ProfileService
}

func NewProfileHandler(
	profileService *service.ProfileService,
) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Ambil personal info sebagai sumber utama data profil
	personalInfo, err := h.profileService.GetPersonalInfo(userID)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		utils.InternalServerError(c, "Gagal mengambil profil", err.Error())
		return
	}

	// Ambil ringkasan profil (weight, height, age) dari health data & tanggal lahir
	profileSummary, err := h.profileService.GetProfile(userID)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		utils.InternalServerError(c, "Gagal mengambil profil", err.Error())
		return
	}

	// Satukan informasi personal info + ringkasan profil sesuai kontrak response
	// Name diambil dari user (register/auth), bukan dari personal_info
	resp := struct {
		Name      string   `json:"name"`
		BirthDate *string  `json:"birth_date"`
		Phone     *string  `json:"phone"`
		Address   *string  `json:"address"`
		PhotoURL  *string  `json:"photo_url"`
		Weight    *float64 `json:"weight"`
		Height    *int     `json:"height"`
		Age       *int     `json:"age"`
	}{
		Name:      profileSummary.Name, // Name diambil dari user.Nama (register/auth), bukan dari personal_info
		BirthDate:  personalInfo.BirthDate,
		Phone:      personalInfo.Phone,
		Address:    personalInfo.Address,
		PhotoURL:   personalInfo.PhotoURL,
		Weight:     profileSummary.Weight,
		Height:     profileSummary.Height,
		Age:        profileSummary.Age,
	}

	utils.SuccessResponse(c, http.StatusOK, "Profil berhasil diambil", resp)
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	var err error
	var oldPhotoURL *string

	// Ambil photoURL lama jika akan diupdate (untuk hapus file lama nanti)
	personalInfo, _ := h.profileService.GetPersonalInfo(userID)
	if personalInfo != nil && personalInfo.PhotoURL != nil {
		oldPhotoURL = personalInfo.PhotoURL
	}

	// Handle multipart/form-data
	var req request.UpdateProfileMultipartRequest
	if err = c.ShouldBind(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	// Validasi format birth_date jika dikirim
	if req.BirthDate != nil && *req.BirthDate != "" {
		_, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			utils.BadRequest(c, "Validasi gagal", "format tanggal lahir tidak valid, gunakan format YYYY-MM-DD")
			return
		}
	}

	// Handle file upload (optional)
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

	// Panggil service dengan photo
	err = h.profileService.UpdateProfileWithMultipart(userID, &req, photoURL)
	if err != nil {
		// Rollback: hapus file baru jika sudah diupload
		if photoURL != nil {
			_ = utils.DeleteProfileImage(*photoURL)
		}

		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		if err.Error() == "format tanggal lahir tidak valid, gunakan format YYYY-MM-DD" ||
			err.Error() == "tanggal lahir tidak boleh di masa depan" {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
		if err.Error() == "tidak ada data untuk diupdate" {
			utils.BadRequest(c, "Tidak ada data untuk diupdate", nil)
			return
		}
		utils.InternalServerError(c, "Gagal mengupdate profil", err.Error())
		return
	}

	// Ambil snapshot personal info terbaru untuk response
	updatedResp, err := h.profileService.GetPersonalInfo(userID)
	if err != nil {
		utils.InternalServerError(c, "Gagal mengambil informasi pribadi terbaru", err.Error())
		return
	}

	// Hapus foto lama jika ada foto baru yang diupload
	if photoURL != nil && oldPhotoURL != nil && *oldPhotoURL != "" {
		_ = utils.DeleteProfileImage(*oldPhotoURL)
	}

	// Response konsisten: kembalikan snapshot personal info terbaru
	utils.SuccessResponse(c, http.StatusOK, "Profil berhasil diupdate", updatedResp)
}

func (h *ProfileHandler) CreatePersonalInfo(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Parse multipart form
	var req request.CreatePersonalInfoRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	// Validasi format birth_date jika dikirim
	if req.BirthDate != nil && *req.BirthDate != "" {
		_, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			utils.BadRequest(c, "Validasi gagal", "format tanggal lahir tidak valid, gunakan format YYYY-MM-DD")
			return
		}
	}

	// Validasi phone jika dikirim: panjang 10-15 digit dan numeric
	if req.Phone != nil && *req.Phone != "" {
		phone := *req.Phone
		if len(phone) < 10 || len(phone) > 15 {
			utils.BadRequest(c, "Validasi gagal", "phone harus 10-15 digit")
			return
		}
		for _, char := range phone {
			if char < '0' || char > '9' {
				utils.BadRequest(c, "Validasi gagal", "phone harus numeric")
				return
			}
		}
	}

	// Handle file upload (optional)
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
	resp, err := h.profileService.CreatePersonalInfo(userID, &req, photoURL)
	if err != nil {
		// Rollback: hapus file jika sudah diupload
		if photoURL != nil {
			_ = utils.DeleteProfileImage(*photoURL)
		}

		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		if err.Error() == "personal info sudah ada" {
			utils.ErrorResponse(c, http.StatusConflict, "Data personal info sudah ada", nil)
			return
		}
		if err.Error() == "format tanggal lahir tidak valid, gunakan format YYYY-MM-DD" ||
			err.Error() == "tanggal lahir tidak boleh di masa depan" ||
			err.Error() == "phone harus 10-15 digit" ||
			err.Error() == "phone harus numeric" {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
		utils.InternalServerError(c, "Gagal membuat informasi pribadi", err.Error())
		return
	}

	// Response konsisten dengan GET/PUT: kembalikan snapshot personal info terbaru
	utils.SuccessResponse(c, http.StatusOK, "Profil berhasil dibuat", resp)
}

func (h *ProfileHandler) UpdatePersonalInfo(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	// Cek content type untuk menentukan apakah JSON atau multipart
	contentType := c.GetHeader("Content-Type")
	isMultipart := false
	if contentType != "" {
		isMultipart = len(contentType) >= 19 && contentType[:19] == "multipart/form-data"
	}

	var err error
	var oldPhotoURL *string

	// Ambil photoURL lama jika akan diupdate (untuk hapus file lama nanti)
	personalInfo, _ := h.profileService.GetPersonalInfo(userID)
	if personalInfo != nil && personalInfo.PhotoURL != nil {
		oldPhotoURL = personalInfo.PhotoURL
	}

	if isMultipart {
		// Handle multipart/form-data (dengan support file upload)
		var req request.UpdateProfileMultipartRequest
		if err = c.ShouldBind(&req); err != nil {
			utils.BadRequest(c, "Data tidak valid", err.Error())
			return
		}

		// Validasi format birth_date jika dikirim
		if req.BirthDate != nil && *req.BirthDate != "" {
			_, err := time.Parse("2006-01-02", *req.BirthDate)
			if err != nil {
				utils.BadRequest(c, "Validasi gagal", "format tanggal lahir tidak valid, gunakan format YYYY-MM-DD")
				return
			}
		}

		// Handle file upload (optional)
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

		// Panggil service dengan photo
		err = h.profileService.UpdatePersonalInfoWithPhoto(userID, &req, photoURL)
		if err != nil {
			// Rollback: hapus file baru jika sudah diupload
			if photoURL != nil {
				_ = utils.DeleteProfileImage(*photoURL)
			}

			if err.Error() == "user tidak ditemukan" {
				utils.NotFound(c, "User tidak ditemukan")
				return
			}
			if err.Error() == "personal info tidak ditemukan, silakan buat terlebih dahulu" {
				utils.NotFound(c, "Personal info tidak ditemukan, silakan buat terlebih dahulu")
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

		// Ambil snapshot personal info terbaru untuk response
		updatedResp, err := h.profileService.GetPersonalInfo(userID)
		if err != nil {
			utils.InternalServerError(c, "Gagal mengambil informasi pribadi terbaru", err.Error())
			return
		}

		// Hapus foto lama jika ada foto baru yang diupload
		if photoURL != nil && oldPhotoURL != nil && *oldPhotoURL != "" {
			_ = utils.DeleteProfileImage(*oldPhotoURL)
		}

		// Response konsisten: kembalikan snapshot personal info terbaru
		utils.SuccessResponse(c, http.StatusOK, "Profil berhasil diupdate", updatedResp)
	} else {
		// Handle JSON (tanpa file upload)
		var req request.UpdatePersonalInfoRequest
		if err = c.ShouldBindJSON(&req); err != nil {
			utils.BadRequest(c, "Data tidak valid", err.Error())
			return
		}

		// Validasi format birth_date jika dikirim
		if req.BirthDate != nil && *req.BirthDate != "" {
			_, err := time.Parse("2006-01-02", *req.BirthDate)
			if err != nil {
				utils.BadRequest(c, "Validasi gagal", "format tanggal lahir tidak valid, gunakan format YYYY-MM-DD")
				return
			}
		}

		err = h.profileService.UpdatePersonalInfo(userID, &req)
		if err != nil {
			if err.Error() == "user tidak ditemukan" {
				utils.NotFound(c, "User tidak ditemukan")
				return
			}
			if err.Error() == "personal info tidak ditemukan, silakan buat terlebih dahulu" {
				utils.NotFound(c, "Personal info tidak ditemukan, silakan buat terlebih dahulu")
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

		// Ambil snapshot personal info terbaru untuk response
		updatedResp, err := h.profileService.GetPersonalInfo(userID)
		if err != nil {
			utils.InternalServerError(c, "Gagal mengambil informasi pribadi terbaru", err.Error())
			return
		}

		// Response konsisten: kembalikan snapshot personal info terbaru
		utils.SuccessResponse(c, http.StatusOK, "Profil berhasil diupdate", updatedResp)
	}
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

func (h *ProfileHandler) CreateHealthTargets(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
		return
	}

	var req request.CreateHealthTargetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	err := h.profileService.CreateHealthTargets(userID, &req)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			utils.NotFound(c, "User tidak ditemukan")
			return
		}
		if err.Error() == "health targets sudah ada, gunakan PUT untuk update" {
			utils.ErrorResponse(c, http.StatusConflict, "Health targets sudah ada, gunakan PUT untuk update", nil)
			return
		}
		utils.InternalServerError(c, "Gagal membuat target kesehatan", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Health targets created successfully", nil)
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

