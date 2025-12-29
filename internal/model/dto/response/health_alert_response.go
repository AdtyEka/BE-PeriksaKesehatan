package response

import "time"

// HealthAlertResponse adalah response untuk health alert
type HealthAlertResponse struct {
	AlertType      string    `json:"alert_type"`
	Value          string    `json:"value"`
	Label          string    `json:"label"`
	Message        string    `json:"message"`
	Status         string    `json:"status"`
	Recommendations []string `json:"recommendations"`
	RecordedAt     time.Time `json:"recorded_at"`
}

// CheckHealthAlertsResponse adalah response untuk endpoint check-health-alerts
type CheckHealthAlertsResponse struct {
	Alerts []HealthAlertResponse `json:"alerts"`
}

