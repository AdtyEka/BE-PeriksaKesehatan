package request

import "time"

// HealthHistoryRequest untuk filter riwayat kesehatan
type HealthHistoryRequest struct {
	// Filter rentang waktu
	// Opsi: "7days", "30days", "3months", "custom"
	// Jika "custom", wajib isi StartDate dan EndDate
	TimeRange string `json:"time_range" form:"time_range"` // default: "7days"

	// Untuk custom range
	StartDate *time.Time `json:"start_date" form:"start_date" time_format:"2006-01-02"`
	EndDate   *time.Time `json:"end_date" form:"end_date" time_format:"2006-01-02"`

	// Filter jenis metrik (bisa multiple)
	// Opsi: "tekanan_darah", "gula_darah", "berat_badan", "aktivitas"
	// Jika kosong, akan mengambil semua metrik
	Metrics []string `json:"metrics" form:"metrics"`
}

