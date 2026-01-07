package request

// EducationalVideoRequest untuk menangkap input JSON saat menambah video edukasi
type EducationalVideoRequest struct {
	VideoTitle  string `json:"video_title" binding:"required"`        // Judul video
	VideoURL    string `json:"video_url" binding:"required"`          // URL video
	CategoryIDs []uint `json:"category_ids" binding:"required,min=1"` // Array ID kategori (minimal 1)
}

