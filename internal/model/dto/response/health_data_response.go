package response

import "time"

// HealthDataResponse untuk mengirim balik data setelah input data kesehatan
//
// MIGRASI NULLABLE-READY:
// Field-field berikut sekarang nullable untuk mencerminkan field yang tidak dikirim.
// Field yang nil akan di-omit dari JSON response (menggunakan omitempty).
// Ini memastikan API contract eksplisit dan tidak ada field dengan zero value yang membingungkan.
type HealthDataResponse struct {
	ID         uint       `json:"id"`
	UserID     uint       `json:"user_id"`
	
	// Field nullable - hanya muncul di response jika dikirim di request
	Systolic   *int       `json:"systolic,omitempty"`
	Diastolic  *int       `json:"diastolic,omitempty"`
	BloodSugar *int       `json:"blood_sugar,omitempty"`
	Weight     *float64   `json:"weight,omitempty"`
	Height     *int       `json:"height,omitempty"`
	HeartRate  *int       `json:"heart_rate,omitempty"`
	Activity   *string    `json:"activity,omitempty"`
	
	CreatedAt  time.Time  `json:"created_at"`
}

