package request

// EducationalVideoRequest untuk menangkap input JSON saat menambah video edukasi
type EducationalVideoRequest struct {
	VideoTitle string `json:"video_title" binding:"required"` // Judul video
	VideoURL   string `json:"video_url" binding:"required"`   // URL video
	CategoryID uint   `json:"category_id" binding:"required"` // ID kategori
}

