package handler

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/service"
	"BE-PeriksaKesehatan/pkg/utils"
	"net/http"

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
			err.Error() == "health_condition tidak boleh kosong" {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
		utils.InternalServerError(c, "Gagal menambah video edukasi", err.Error())
		return
	}

	// Response sukses
	utils.SuccessResponse(c, http.StatusCreated, "Video edukasi berhasil ditambahkan", resp)
}

// GetEducationalVideos menangani request untuk mengambil video edukasi berdasarkan kondisi kesehatan
func (h *EducationalVideoHandler) GetEducationalVideos(c *gin.Context) {
	// Ambil health_condition dari path parameter
	healthCondition := c.Param("health_condition")
	if healthCondition == "" {
		utils.BadRequest(c, "health_condition tidak boleh kosong", nil)
		return
	}

	// Panggil service untuk mengambil video
	resp, err := h.educationalVideoService.GetEducationalVideosByHealthCondition(healthCondition)
	if err != nil {
		if err.Error() == "health_condition tidak boleh kosong" {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
		utils.InternalServerError(c, "Gagal mengambil video edukasi", err.Error())
		return
	}

	// Response sukses
	utils.SuccessResponse(c, http.StatusOK, "Video edukasi berhasil diambil", resp)
}

