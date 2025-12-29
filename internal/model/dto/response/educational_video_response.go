package response

// EducationalVideoItem adalah item video dalam response
type EducationalVideoItem struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// GetEducationalVideosResponse adalah response untuk endpoint get-educational-videos
type GetEducationalVideosResponse struct {
	HealthCondition string                 `json:"health_condition"`
	Videos          []EducationalVideoItem `json:"videos"`
}

// AddEducationalVideoResponse adalah response untuk endpoint add-educational-video
type AddEducationalVideoResponse struct {
	ID              uint   `json:"id"`
	VideoTitle      string `json:"video_title"`
	VideoURL        string `json:"video_url"`
	HealthCondition string `json:"health_condition"`
}

