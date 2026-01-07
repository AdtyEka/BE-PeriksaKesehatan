package response

// EducationalVideoItem adalah item video dalam response
type EducationalVideoItem struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// CategoryWithVideosResponse adalah response untuk kategori beserta videonya
type CategoryWithVideosResponse struct {
	ID       uint                 `json:"id"`
	Kategori string              `json:"kategori"`
	Videos   []EducationalVideoItem `json:"videos"`
}

// GetEducationalVideosByIDResponse adalah response untuk endpoint get-educational-videos/{id}
type GetEducationalVideosByIDResponse struct {
	ID       uint                 `json:"id"`
	Kategori string              `json:"kategori"`
	Videos   []EducationalVideoItem `json:"videos"`
}

// GetAllEducationalVideosResponse adalah response untuk endpoint get-educational-videos
type GetAllEducationalVideosResponse struct {
	Data []CategoryWithVideosResponse `json:"data"`
}

// AddEducationalVideoResponse adalah response untuk endpoint add-educational-video
type AddEducationalVideoResponse struct {
	ID          uint   `json:"id"`
	VideoTitle  string `json:"video_title"`
	VideoURL    string `json:"video_url"`
	CategoryIDs []uint `json:"category_ids"`
}

