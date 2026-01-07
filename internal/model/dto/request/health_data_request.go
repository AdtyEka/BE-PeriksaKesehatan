package request

// HealthDataRequest untuk menangkap input JSON saat input data kesehatan
// 
// MIGRASI NULLABLE-READY:
// Field-field berikut sekarang nullable-ready untuk mendukung partial update
// dan fleksibilitas input. Validasi dilakukan di service layer untuk memastikan:
// 1. Minimal satu metrik kesehatan harus diisi
// 2. Field yang dikirim harus valid (jika tidak nil)
// 3. Tidak ada silent fallback ke zero value
//
// CATATAN BACKWARD COMPATIBILITY:
// - Request lama dengan semua field masih valid
// - Field yang tidak dikirim akan nil (bukan zero value)
// - Validasi di service layer akan memastikan minimal satu field diisi
type HealthDataRequest struct {
	// Tekanan darah sistolik (mmHg) - nullable, validasi: 0-180 jika dikirim
	Systolic *int `json:"systolic"`
	
	// Tekanan darah diastolik (mmHg) - nullable, validasi: 0-120 jika dikirim
	// NOTE: Jika systolic dikirim, diastolic juga harus dikirim (business rule)
	Diastolic *int `json:"diastolic"`
	
	// Gula darah (mg/dL) - nullable, validasi: 0-300 jika dikirim
	BloodSugar *int `json:"blood_sugar"`
	
	// Berat badan (kg) - nullable, validasi: 20-200 jika dikirim
	Weight *float64 `json:"weight"`
	
	// Tinggi badan (cm) - nullable, validasi: 50-250 jika dikirim
	Height *int `json:"height"`
	
	// Detak jantung (bpm) - nullable, validasi: 0-180 jika dikirim
	HeartRate *int `json:"heart_rate"`
	
	// Aktivitas terbaru (opsional) - sudah nullable dari awal
	Activity *string `json:"activity"`
}

