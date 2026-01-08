package response

import "time"

// EducationVideoItem adalah item video edukasi dalam alert
type EducationVideoItem struct {
	ID         uint   `json:"id"`
	VideoTitle string `json:"video_title"`
	VideoURL   string `json:"video_url"`
	CategoryID uint   `json:"category_id"`
}

// HealthAlertResponse adalah response untuk health alert
type HealthAlertResponse struct {
	AlertType        string              `json:"alert_type"`
	Category         string              `json:"category"`
	Value            string              `json:"value"`
	Label            string              `json:"label"`
	Status           string              `json:"status"`
	RecordedAt       time.Time           `json:"recorded_at"`
	Explanation      string              `json:"explanation"`
	ImmediateActions []string            `json:"immediate_actions"`
	MedicalAttention []string            `json:"medical_attention"`
	ManagementTips   []string            `json:"management_tips"`
	EducationVideos  []EducationVideoItem `json:"education_videos"`
}

// CheckHealthAlertsResponse adalah response untuk endpoint check-health-alerts
type CheckHealthAlertsResponse struct {
	Alerts []HealthAlertResponse `json:"alerts"`
}

