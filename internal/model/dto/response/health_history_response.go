package response

import "time"

// HealthHistoryResponse adalah response utama untuk riwayat kesehatan
type HealthHistoryResponse struct {
	Summary      HealthSummaryResponse      `json:"summary"`       // Ringkasan statistik
	TrendCharts  TrendChartsResponse        `json:"trend_charts"`  // Data grafik tren
	ReadingHistory []ReadingHistoryResponse `json:"reading_history"` // Catatan pembacaan kronologis
}

// HealthSummaryResponse berisi ringkasan statistik untuk semua metrik
type HealthSummaryResponse struct {
	BloodPressure *BloodPressureSummary `json:"blood_pressure,omitempty"` // Ringkasan tekanan darah
	BloodSugar    *BloodSugarSummary    `json:"blood_sugar,omitempty"`    // Ringkasan gula darah
	Weight        *WeightSummary        `json:"weight,omitempty"`         // Ringkasan berat badan
	Activity      *ActivitySummary     `json:"activity,omitempty"`       // Ringkasan aktivitas
}

// BloodPressureSummary ringkasan statistik tekanan darah
type BloodPressureSummary struct {
	AvgSystolic      float64 `json:"avg_systolic"`       // Rata-rata systolic
	AvgDiastolic     float64 `json:"avg_diastolic"`     // Rata-rata diastolic
	ChangePercent    float64 `json:"change_percent"`     // Persentase perubahan dari periode sebelumnya
	SystolicStatus   string  `json:"systolic_status"`    // Status: RENDAH / NORMAL / TINGGI (WHO)
	DiastolicStatus  string  `json:"diastolic_status"`  // Status: RENDAH / NORMAL / TINGGI (WHO)
	NormalRange      string  `json:"normal_range"`      // Rentang normal: "90-139 / 60-89 mmHg (WHO)"
}

// BloodSugarSummary ringkasan statistik gula darah
type BloodSugarSummary struct {
	AvgValue      float64 `json:"avg_value"`       // Rata-rata gula darah
	ChangePercent float64 `json:"change_percent"`  // Persentase perubahan
	Status        string  `json:"status"`          // Status: RENDAH / NORMAL / TINGGI (WHO)
	NormalRange   string  `json:"normal_range"`    // Rentang normal: "70-140 mg/dL (WHO - Gula Darah Sewaktu)"
}

// WeightSummary ringkasan statistik berat badan
type WeightSummary struct {
	AvgWeight    float64 `json:"avg_weight"`     // Rata-rata berat badan
	BMI          *float64 `json:"bmi,omitempty"` // BMI (jika tinggi badan tersedia)
	Trend        string  `json:"trend"`          // Tren: Stabil / Naik / Turun
	ChangePercent float64 `json:"change_percent"` // Persentase perubahan
}

// ActivitySummary ringkasan statistik aktivitas
type ActivitySummary struct {
	TotalSteps     int     `json:"total_steps"`      // Total langkah
	TotalCalories  float64 `json:"total_calories"`   // Total kalori
	ChangePercent  float64 `json:"change_percent"`   // Persentase perubahan aktivitas
}

// TrendChartsResponse berisi data time-series untuk grafik tren dengan filter waktu
type TrendChartsResponse struct {
	BloodPressure BloodPressureTrendCharts `json:"blood_pressure,omitempty"` // Data tren tekanan darah
	BloodSugar    BloodSugarTrendCharts    `json:"blood_sugar,omitempty"`    // Data tren gula darah
	Weight        WeightTrendCharts        `json:"weight,omitempty"`         // Data tren berat badan
	Activity      ActivityTrendCharts      `json:"activity,omitempty"`       // Data tren aktivitas
}

// BloodPressureTrendCharts berisi data tren tekanan darah dengan filter waktu
type BloodPressureTrendCharts struct {
	Days7    []BloodPressureTrendPoint `json:"7Days"`    // Data 7 hari terakhir
	Month1   []BloodPressureTrendPoint `json:"1Month"`   // Data 30 hari terakhir
	Months3  []BloodPressureTrendPoint `json:"3Months"`  // Data 90 hari terakhir
}

// BloodSugarTrendCharts berisi data tren gula darah dengan filter waktu
type BloodSugarTrendCharts struct {
	Days7    []BloodSugarTrendPoint `json:"7Days"`    // Data 7 hari terakhir
	Month1   []BloodSugarTrendPoint `json:"1Month"`   // Data 30 hari terakhir
	Months3  []BloodSugarTrendPoint `json:"3Months"`  // Data 90 hari terakhir
}

// WeightTrendCharts berisi data tren berat badan dengan filter waktu
type WeightTrendCharts struct {
	Days7    []WeightTrendPoint `json:"7Days"`    // Data 7 hari terakhir
	Month1   []WeightTrendPoint `json:"1Month"`   // Data 30 hari terakhir
	Months3  []WeightTrendPoint `json:"3Months"`  // Data 90 hari terakhir
}

// ActivityTrendCharts berisi data tren aktivitas dengan filter waktu
type ActivityTrendCharts struct {
	Days7    []ActivityTrendPoint `json:"7Days"`    // Data 7 hari terakhir
	Month1   []ActivityTrendPoint `json:"1Month"`   // Data 30 hari terakhir
	Months3  []ActivityTrendPoint `json:"3Months"`  // Data 90 hari terakhir
}

// BloodPressureTrendPoint satu titik data untuk grafik tekanan darah
type BloodPressureTrendPoint struct {
	Date      string  `json:"date"`       // Tanggal (format: YYYY-MM-DD)
	Systolic  float64 `json:"systolic"`   // Rata-rata systolic hari itu
	Diastolic float64 `json:"diastolic"` // Rata-rata diastolic hari itu
}

// BloodSugarTrendPoint satu titik data untuk grafik gula darah
type BloodSugarTrendPoint struct {
	Date      string  `json:"date"`      // Tanggal (format: YYYY-MM-DD)
	AvgValue  float64 `json:"avg_value"` // Rata-rata gula darah hari itu
}

// WeightTrendPoint satu titik data untuk grafik berat badan
type WeightTrendPoint struct {
	Date   string  `json:"date"`   // Tanggal (format: YYYY-MM-DD)
	Weight float64 `json:"weight"` // Rata-rata berat badan hari itu
}

// ActivityTrendPoint satu titik data untuk grafik aktivitas
type ActivityTrendPoint struct {
	Date     string  `json:"date"`      // Tanggal (format: YYYY-MM-DD)
	Steps    int     `json:"steps"`     // Total langkah hari itu
	Calories float64 `json:"calories"`  // Total kalori hari itu
}

// ReadingHistoryResponse satu catatan pembacaan dalam riwayat
type ReadingHistoryResponse struct {
	ID          uint      `json:"id"`           // ID record
	DateTime    time.Time `json:"date_time"`    // Tanggal & waktu pengukuran
	MetricType  string    `json:"metric_type"`  // Jenis metrik: "tekanan_darah", "gula_darah", "berat_badan", "aktivitas"
	Value       string    `json:"value"`        // Nilai pengukuran (format string untuk fleksibilitas)
	Context     *string   `json:"context"`      // Konteks (puasa, setelah makan, dll)
	Status      string    `json:"status"`       // Status: RENDAH / NORMAL / TINGGI (WHO)
	Notes       *string   `json:"notes"`        // Catatan tambahan
}

