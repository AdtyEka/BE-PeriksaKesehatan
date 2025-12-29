package request

// HealthDataRequest untuk menangkap input JSON saat input data kesehatan
type HealthDataRequest struct {
	Systolic   *int     `json:"systolic"`   // Tekanan darah sistolik (mmHg)
	Diastolic  *int     `json:"diastolic"`  // Tekanan darah diastolik (mmHg)
	BloodSugar *int     `json:"blood_sugar"` // Gula darah (mg/dL)
	Weight     *float64 `json:"weight"`     // Berat badan (kg)
	HeartRate  *int     `json:"heart_rate"` // Detak jantung (bpm)
	Activity   *string  `json:"activity"`   // Aktivitas terbaru (opsional)
}

