package response

import "time"

// HealthDataResponse untuk mengirim balik data setelah input data kesehatan
type HealthDataResponse struct {
	ID         uint       `json:"id"`
	UserID     uint       `json:"user_id"`
	Systolic   int        `json:"systolic"`
	Diastolic  int        `json:"diastolic"`
	BloodSugar int        `json:"blood_sugar"`
	Weight     float64    `json:"weight"`
	HeartRate  int        `json:"heart_rate"`
	Activity   *string    `json:"activity"`
	CreatedAt  time.Time  `json:"created_at"`
}

