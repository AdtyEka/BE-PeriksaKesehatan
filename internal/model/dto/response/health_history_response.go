package response

import "time"

// HealthHistoryResponse adalah response utama untuk riwayat kesehatan
type HealthHistoryResponse struct {
	Summary        HealthSummaryResponse      `json:"summary"`         // Ringkasan statistik (untuk internal service)
	TrendCharts    TrendChartsResponse        `json:"trend_charts"`    // Data grafik tren
	ReadingHistory []ReadingHistoryResponse   `json:"reading_history"` // Catatan pembacaan kronologis (flat, untuk internal service)
}

// HealthHistorySummaryByRange membungkus ringkasan statistik per rentang waktu
// Struktur ini digunakan khusus untuk response API agar sesuai kebutuhan frontend:
// {
//   "summary": {
//     "start_date": "2025-01-01",
//     "end_date": "2025-01-31",
//     "7Days": { ... },
//     "1Month": { ... },
//     "3Months": { ... }
//   },
//   ...
// }
type HealthHistorySummaryByRange struct {
	StartDate string                    `json:"start_date"` // Tanggal mulai global (format: YYYY-MM-DD)
	EndDate   string                    `json:"end_date"`   // Tanggal akhir global (format: YYYY-MM-DD)
	Days7     *HealthSummaryResponse    `json:"7Days,omitempty"`   // Ringkasan untuk 7 hari terakhir
	Month1    *HealthSummaryWithWeeks   `json:"1Month,omitempty"`  // Ringkasan untuk 30 hari terakhir (dengan weeks)
	Months3   *HealthSummaryWithWeeks   `json:"3Months,omitempty"` // Ringkasan untuk 3 bulan terakhir (dengan weeks)
}

// HealthSummaryWithWeeks berisi ringkasan statistik dengan agregasi per minggu
type HealthSummaryWithWeeks struct {
	HealthSummaryResponse
	Weeks []HealthSummaryWeek `json:"weeks"` // Agregasi per minggu
}

// HealthSummaryWeek berisi ringkasan statistik untuk satu minggu
type HealthSummaryWeek struct {
	Week      string                `json:"week"`       // Label minggu: "Week 1", "Week 2", dll
	StartDate string                `json:"start_date"` // Tanggal mulai minggu (format: YYYY-MM-DD)
	EndDate   string                `json:"end_date"`   // Tanggal akhir minggu (format: YYYY-MM-DD)
	Summary   HealthSummaryResponse `json:"summary"`    // Ringkasan statistik untuk minggu tersebut
}

// HealthReadingHistoryByRange membungkus riwayat pembacaan per rentang waktu
// {
//   "reading_history": {
//     "7Days": [ ... ],
//     "1Month": { "start_date": "...", "end_date": "...", "records": [ ... ] },
//     "3Months": { "start_date": "...", "end_date": "...", "records": [ ... ] }
//   }
// }
type HealthReadingHistoryByRange struct {
	Days7   []ReadingHistoryResponse        `json:"7Days,omitempty"`   // Riwayat untuk 7 hari terakhir
	Month1  *ReadingHistoryGrouped          `json:"1Month,omitempty"`  // Riwayat untuk 30 hari terakhir (dengan grouping)
	Months3 *ReadingHistoryGrouped          `json:"3Months,omitempty"` // Riwayat untuk 3 bulan terakhir (dengan grouping)
}

// ReadingHistoryGrouped berisi riwayat pembacaan dengan informasi tanggal
type ReadingHistoryGrouped struct {
	StartDate string                  `json:"start_date"` // Tanggal mulai (format: YYYY-MM-DD)
	EndDate   string                  `json:"end_date"`   // Tanggal akhir (format: YYYY-MM-DD)
	Records   []ReadingHistoryResponse `json:"records"`    // Catatan pembacaan
}

// HealthHistoryAPIResponse adalah bentuk response yang dikirimkan ke client
// dengan struktur summary & reading_history yang sudah dipisah per rentang waktu.
// TrendCharts tetap sama seperti sebelumnya.
type HealthHistoryAPIResponse struct {
	Summary        HealthHistorySummaryByRange  `json:"summary"`
	TrendCharts    TrendChartsResponse          `json:"trend_charts"`
	ReadingHistory HealthReadingHistoryByRange  `json:"reading_history"`
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
	Days7    []BloodPressureTrendPoint      `json:"7Days"`    // Data 7 hari terakhir (per hari)
	Month1   []BloodPressureTrendPointWeek  `json:"1Month"`   // Data 30 hari terakhir (per minggu)
	Months3  []BloodPressureTrendPointMonth `json:"3Months"`  // Data 90 hari terakhir (per bulan)
}

// BloodSugarTrendCharts berisi data tren gula darah dengan filter waktu
type BloodSugarTrendCharts struct {
	Days7    []BloodSugarTrendPoint      `json:"7Days"`    // Data 7 hari terakhir (per hari)
	Month1   []BloodSugarTrendPointWeek  `json:"1Month"`   // Data 30 hari terakhir (per minggu)
	Months3  []BloodSugarTrendPointMonth `json:"3Months"`  // Data 90 hari terakhir (per bulan)
}

// WeightTrendCharts berisi data tren berat badan dengan filter waktu
type WeightTrendCharts struct {
	Days7    []WeightTrendPoint      `json:"7Days"`    // Data 7 hari terakhir (per hari)
	Month1   []WeightTrendPointWeek  `json:"1Month"`   // Data 30 hari terakhir (per minggu)
	Months3  []WeightTrendPointMonth `json:"3Months"`  // Data 90 hari terakhir (per bulan)
}

// ActivityTrendCharts berisi data tren aktivitas dengan filter waktu
type ActivityTrendCharts struct {
	Days7    []ActivityTrendPoint      `json:"7Days"`    // Data 7 hari terakhir (per hari)
	Month1   []ActivityTrendPointWeek  `json:"1Month"`   // Data 30 hari terakhir (per minggu)
	Months3  []ActivityTrendPointMonth `json:"3Months"`  // Data 90 hari terakhir (per bulan)
}

// BloodPressureTrendPoint satu titik data untuk grafik tekanan darah (7Days - per hari)
type BloodPressureTrendPoint struct {
	Date      string  `json:"date"`       // Tanggal (format: YYYY-MM-DD)
	Systolic  float64 `json:"systolic"`   // Rata-rata systolic hari itu
	Diastolic float64 `json:"diastolic"` // Rata-rata diastolic hari itu
}

// BloodPressureTrendPointWeek satu titik data untuk grafik tekanan darah (1Month - per minggu)
type BloodPressureTrendPointWeek struct {
	Week      string  `json:"week"`       // Label minggu: "Week 1", "Week 2", dll
	StartDate string  `json:"start_date"` // Tanggal mulai (format: YYYY-MM-DD)
	EndDate   string  `json:"end_date"`   // Tanggal akhir (format: YYYY-MM-DD)
	Systolic  float64 `json:"systolic"`   // Rata-rata systolic minggu itu
	Diastolic float64 `json:"diastolic"` // Rata-rata diastolic minggu itu
}

// BloodPressureTrendPointMonth satu titik data untuk grafik tekanan darah (3Months - per bulan)
type BloodPressureTrendPointMonth struct {
	Month     string  `json:"month"`     // Label bulan: "Dec 2025", "Jan 2026", dll
	Systolic  float64 `json:"systolic"`   // Rata-rata systolic bulan itu
	Diastolic float64 `json:"diastolic"` // Rata-rata diastolic bulan itu
}

// BloodSugarTrendPoint satu titik data untuk grafik gula darah (7Days - per hari)
type BloodSugarTrendPoint struct {
	Date     string  `json:"date"`      // Tanggal (format: YYYY-MM-DD)
	AvgValue float64 `json:"avg_value"` // Rata-rata gula darah hari itu
}

// BloodSugarTrendPointWeek satu titik data untuk grafik gula darah (1Month - per minggu)
type BloodSugarTrendPointWeek struct {
	Week      string  `json:"week"`       // Label minggu: "Week 1", "Week 2", dll
	StartDate string  `json:"start_date"` // Tanggal mulai (format: YYYY-MM-DD)
	EndDate   string  `json:"end_date"`   // Tanggal akhir (format: YYYY-MM-DD)
	AvgValue  float64 `json:"avg_value"`  // Rata-rata gula darah minggu itu
}

// BloodSugarTrendPointMonth satu titik data untuk grafik gula darah (3Months - per bulan)
type BloodSugarTrendPointMonth struct {
	Month    string  `json:"month"`     // Label bulan: "Dec 2025", "Jan 2026", dll
	AvgValue float64 `json:"avg_value"` // Rata-rata gula darah bulan itu
}

// WeightTrendPoint satu titik data untuk grafik berat badan (7Days - per hari)
type WeightTrendPoint struct {
	Date   string  `json:"date"`   // Tanggal (format: YYYY-MM-DD)
	Weight float64 `json:"weight"` // Rata-rata berat badan hari itu
}

// WeightTrendPointWeek satu titik data untuk grafik berat badan (1Month - per minggu)
type WeightTrendPointWeek struct {
	Week      string  `json:"week"`       // Label minggu: "Week 1", "Week 2", dll
	StartDate string  `json:"start_date"` // Tanggal mulai (format: YYYY-MM-DD)
	EndDate   string  `json:"end_date"`   // Tanggal akhir (format: YYYY-MM-DD)
	Weight    float64 `json:"weight"`     // Rata-rata berat badan minggu itu
}

// WeightTrendPointMonth satu titik data untuk grafik berat badan (3Months - per bulan)
type WeightTrendPointMonth struct {
	Month  string  `json:"month"`  // Label bulan: "Dec 2025", "Jan 2026", dll
	Weight float64 `json:"weight"` // Rata-rata berat badan bulan itu
}

// ActivityTrendPoint satu titik data untuk grafik aktivitas (7Days - per hari)
type ActivityTrendPoint struct {
	Date     string  `json:"date"`      // Tanggal (format: YYYY-MM-DD)
	Steps    int     `json:"steps"`     // Total langkah hari itu
	Calories float64 `json:"calories"`  // Total kalori hari itu
}

// ActivityTrendPointWeek satu titik data untuk grafik aktivitas (1Month - per minggu)
type ActivityTrendPointWeek struct {
	Week      string  `json:"week"`       // Label minggu: "Week 1", "Week 2", dll
	StartDate string  `json:"start_date"` // Tanggal mulai (format: YYYY-MM-DD)
	EndDate   string  `json:"end_date"`   // Tanggal akhir (format: YYYY-MM-DD)
	Steps     int     `json:"steps"`      // Total langkah minggu itu
	Calories  float64 `json:"calories"`   // Total kalori minggu itu
}

// ActivityTrendPointMonth satu titik data untuk grafik aktivitas (3Months - per bulan)
type ActivityTrendPointMonth struct {
	Month    string  `json:"month"`    // Label bulan: "Dec 2025", "Jan 2026", dll
	Steps    int     `json:"steps"`     // Total langkah bulan itu
	Calories float64 `json:"calories"`  // Total kalori bulan itu
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

