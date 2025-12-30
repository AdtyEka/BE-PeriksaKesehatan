package handler

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/service"
	"BE-PeriksaKesehatan/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// EducationalVideoHandler menangani semua request terkait educational videos
type EducationalVideoHandler struct {
	educationalVideoService *service.EducationalVideoService
}

// NewEducationalVideoHandler membuat instance baru dari EducationalVideoHandler
func NewEducationalVideoHandler(educationalVideoService *service.EducationalVideoService) *EducationalVideoHandler {
	return &EducationalVideoHandler{
		educationalVideoService: educationalVideoService,
	}
}

// AddEducationalVideo menangani request untuk menambah video edukasi
func (h *EducationalVideoHandler) AddEducationalVideo(c *gin.Context) {
	var req request.EducationalVideoRequest

	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	// Panggil service untuk menambah video
	resp, err := h.educationalVideoService.AddEducationalVideo(&req)
	if err != nil {
		// Cek apakah error adalah validasi
		if err.Error() == "video_title tidak boleh kosong" ||
			err.Error() == "video_url tidak boleh kosong" ||
			err.Error() == "video_url harus berupa URL yang valid" ||
			err.Error() == "category_id tidak boleh kosong" ||
			err.Error() == "kategori tidak ditemukan" {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
		utils.InternalServerError(c, "Gagal menambah video edukasi", err.Error())
		return
	}

	// Response sukses
	utils.SuccessResponse(c, http.StatusCreated, "Video edukasi berhasil ditambahkan", resp)
}

// GetAllEducationalVideos menangani request untuk mengambil semua kategori beserta videonya
func (h *EducationalVideoHandler) GetAllEducationalVideos(c *gin.Context) {
	// Panggil service untuk mengambil semua kategori dan video
	resp, err := h.educationalVideoService.GetAllEducationalVideos()
	if err != nil {
		utils.InternalServerError(c, "Gagal mengambil video edukasi", err.Error())
		return
	}

	// Response sukses dengan format langsung sesuai requirement
	c.JSON(http.StatusOK, resp)
}

// GetEducationalVideosByID menangani request untuk mengambil video edukasi berdasarkan kategori ID
func (h *EducationalVideoHandler) GetEducationalVideosByID(c *gin.Context) {
	// Ambil ID dari path parameter
	idStr := c.Param("id")
	if idStr == "" {
		utils.BadRequest(c, "ID kategori tidak boleh kosong", nil)
		return
	}

	// Validasi ID adalah integer
	_, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID kategori tidak valid", nil)
		return
	}

	// Panggil service untuk mengambil video
	resp, err := h.educationalVideoService.GetEducationalVideosByCategoryID(idStr)
	if err != nil {
		if err.Error() == "ID kategori tidak valid" {
			utils.BadRequest(c, "ID kategori tidak valid", nil)
			return
		}
		if err.Error() == "kategori tidak ditemukan" {
			utils.NotFound(c, "Kategori tidak ditemukan")
			return
		}
		utils.InternalServerError(c, "Gagal mengambil video edukasi", err.Error())
		return
	}

	// Response sukses dengan format langsung sesuai requirement
	c.JSON(http.StatusOK, resp)
}

