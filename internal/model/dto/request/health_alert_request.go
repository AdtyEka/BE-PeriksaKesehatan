package request

import "time"

// HealthAlertRequest untuk menangkap input JSON saat check health alerts
type HealthAlertRequest struct {
	Systolic   int       `json:"systolic" binding:"required"`   // Tekanan darah sistolik (mmHg)
	Diastolic  int       `json:"diastolic" binding:"required"`  // Tekanan darah diastolik (mmHg)
	BloodSugar int       `json:"blood_sugar" binding:"required"` // Gula darah (mg/dL)
	Weight     float64   `json:"weight" binding:"required"`     // Berat badan (kg)
	RecordedAt time.Time `json:"recorded_at" binding:"required"` // Timestamp pengukuran
}

