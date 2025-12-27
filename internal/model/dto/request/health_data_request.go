package request

// HealthDataRequest untuk menangkap input JSON saat input data kesehatan
type HealthDataRequest struct {
	Systolic   int     `json:"systolic" binding:"required"`   // Tekanan darah sistolik (mmHg)
	Diastolic  int     `json:"diastolic" binding:"required"`  // Tekanan darah diastolik (mmHg)
	BloodSugar int     `json:"blood_sugar" binding:"required"` // Gula darah (mg/dL)
	Weight     float64 `json:"weight" binding:"required"`     // Berat badan (kg)
	HeartRate  int     `json:"heart_rate" binding:"required"` // Detak jantung (bpm)
	Activity   *string `json:"activity"`                       // Aktivitas terbaru (opsional)
}

