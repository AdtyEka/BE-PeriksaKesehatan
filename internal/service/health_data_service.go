package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
	"BE-PeriksaKesehatan/pkg/utils"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type HealthDataService struct {
	healthDataRepo  *repository.HealthDataRepository
	healthAlertRepo *repository.HealthAlertRepository
}

func NewHealthDataService(
	healthDataRepo *repository.HealthDataRepository,
	healthAlertRepo *repository.HealthAlertRepository,
) *HealthDataService {
	return &HealthDataService{
		healthDataRepo:  healthDataRepo,
		healthAlertRepo: healthAlertRepo,
	}
}


// ValidateHealthData melakukan validasi range nilai data kesehatan
// Validasi ini eksplisit membedakan antara nil, zero value, dan nilai valid
func (s *HealthDataService) ValidateHealthData(req *request.HealthDataRequest) error {
	if req == nil {
		return errors.New("request tidak boleh nil")
	}

	// Validasi systolic: 90 – 180 (hanya jika tidak null)
	if req.Systolic != nil {
		systolic := *req.Systolic
		if systolic < 90 || systolic > 180 {
			return fmt.Errorf("systolic harus berada dalam range 90-180 mmHg, got: %d", systolic)
		}
	}

	// Validasi diastolic: 60 – 120 (hanya jika tidak null)
	if req.Diastolic != nil {
		diastolic := *req.Diastolic
		if diastolic < 60 || diastolic > 120 {
			return fmt.Errorf("diastolic harus berada dalam range 60-120 mmHg, got: %d", diastolic)
		}
	}

	// Validasi blood_sugar: 60 – 300 (hanya jika tidak null)
	if req.BloodSugar != nil {
		bloodSugar := *req.BloodSugar
		if bloodSugar < 60 || bloodSugar > 300 {
			return fmt.Errorf("blood_sugar harus berada dalam range 60-300 mg/dL, got: %d", bloodSugar)
		}
	}

	// Validasi weight: 20 – 200 (hanya jika tidak null)
	// Catatan: weight 0.0 dianggap invalid, harus explicit nil atau dalam range
	if req.Weight != nil {
		weight := *req.Weight
		if weight <= 0 {
			return errors.New("weight harus lebih besar dari 0 kg")
		}
		if weight < 20 || weight > 200 {
			return fmt.Errorf("weight harus berada dalam range 20-200 kg, got: %.2f", weight)
		}
	}

	// Validasi heart_rate: 40 – 180 (hanya jika tidak null)
	if req.HeartRate != nil {
		heartRate := *req.HeartRate
		if heartRate < 40 || heartRate > 180 {
			return fmt.Errorf("heart_rate harus berada dalam range 40-180 bpm, got: %d", heartRate)
		}
	}

	// Validasi Activity: jika tidak nil, harus tidak empty string
	// Ini membedakan antara null (tidak ada data) dan empty string (data kosong)
	if req.Activity != nil && *req.Activity == "" {
		return errors.New("activity tidak boleh empty string, gunakan null jika tidak ada data")
	}

	return nil
}

func (s *HealthDataService) CreateHealthData(userID uint, req *request.HealthDataRequest) (*response.HealthDataResponse, error) {
	// Guard clause: validasi input
	if req == nil {
		return nil, errors.New("request tidak boleh nil")
	}
	if userID == 0 {
		return nil, errors.New("user_id tidak valid")
	}

	// 1) Validasi data
	if err := s.ValidateHealthData(req); err != nil {
		return nil, fmt.Errorf("validasi gagal: %w", err)
	}

	// 2) Bentuk entity dari request dengan normalisasi Activity
	// Normalisasi: empty string -> nil untuk konsistensi
	var activity *string
	if req.Activity != nil && *req.Activity != "" {
		activity = req.Activity
	} else {
		activity = nil
	}

	healthData := &entity.HealthData{
		UserID:     userID,
		Systolic:   req.Systolic,
		Diastolic:  req.Diastolic,
		BloodSugar: req.BloodSugar,
		Weight:     req.Weight,
		HeartRate:  req.HeartRate,
		Activity:   activity,
	}

	// 3) Simpan ke database dulu (biar dapat ID & CreatedAt)
	if err := s.healthDataRepo.CreateHealthData(healthData); err != nil {
		return nil, fmt.Errorf("gagal menyimpan data kesehatan: %w", err)
	}

	// Guard clause: pastikan healthData sudah punya ID setelah create
	if healthData.ID == 0 {
		return nil, errors.New("gagal mendapatkan ID setelah menyimpan data kesehatan")
	}

	// 4) Generate alert dari data yang sudah tersimpan
	alerts := s.generateAlertsFromHealthData(*healthData)

	// 5) Simpan alert ke DB (kalau gagal, data utama tetap aman)
	// Catatan: error saat save alert tidak menghentikan proses, hanya log
	if len(alerts) > 0 {
		if err := s.healthAlertRepo.SaveMany(alerts); err != nil {
			// Log error tapi tidak return error (data utama sudah tersimpan)
			// Di production, bisa gunakan logger di sini
			_ = err
		}
	}

	// 6) Buat response dengan mapping yang eksplisit
	resp := &response.HealthDataResponse{
		ID:         healthData.ID,
		UserID:     healthData.UserID,
		Systolic:   healthData.Systolic,
		Diastolic:  healthData.Diastolic,
		BloodSugar: healthData.BloodSugar,
		Weight:     healthData.Weight,
		HeartRate:  healthData.HeartRate,
		Activity:   healthData.Activity,
		CreatedAt:  healthData.CreatedAt,
	}

	return resp, nil
}

// GetHealthDataByUserID mengambil semua data kesehatan berdasarkan UserID
func (s *HealthDataService) GetHealthDataByUserID(userID uint) ([]entity.HealthData, error) {
	return s.healthDataRepo.GetHealthDataByUserID(userID)
}

// ==================== OPERASI UNTUK RIWAYAT KESEHATAN ====================

// GetHealthHistory mengambil riwayat kesehatan dengan filter dan menghasilkan ringkasan, grafik tren, dan catatan
func (s *HealthDataService) GetHealthHistory(userID uint, req *request.HealthHistoryRequest) (*response.HealthHistoryResponse, error) {
	// Tentukan rentang waktu
	startDate, endDate, err := s.parseTimeRange(req)
	if err != nil {
		return nil, err
	}

	// Ambil data dengan filter
	healthDataList, err := s.healthDataRepo.GetHealthDataByUserIDWithFilter(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Ambil data periode sebelumnya untuk perbandingan
	periodLength := endDate.Sub(startDate)
	prevDataList, _ := s.healthDataRepo.GetHealthDataForComparison(userID, startDate, endDate, periodLength)

	// Filter berdasarkan metrik jika ada
	filteredData := s.filterByMetrics(healthDataList, req.Metrics)

	// Buat response
	result := &response.HealthHistoryResponse{}

	// Ringkasan statistik
	result.Summary = s.calculateSummary(filteredData, prevDataList, req.Metrics)

	// Grafik tren
	result.TrendCharts = s.calculateTrendCharts(filteredData, req.Metrics)

	// Catatan pembacaan
	result.ReadingHistory = s.buildReadingHistory(filteredData)

	return result, nil
}

// parseTimeRange mengkonversi time_range ke startDate dan endDate
func (s *HealthDataService) parseTimeRange(req *request.HealthHistoryRequest) (time.Time, time.Time, error) {
	now := time.Now()
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	var startDate time.Time

	// Jika custom range, gunakan StartDate dan EndDate
	if req.TimeRange == "custom" {
		if req.StartDate == nil || req.EndDate == nil {
			return time.Time{}, time.Time{}, errors.New("start_date dan end_date wajib diisi untuk custom range")
		}
		startDate = time.Date(req.StartDate.Year(), req.StartDate.Month(), req.StartDate.Day(), 0, 0, 0, 0, req.StartDate.Location())
		endDate = time.Date(req.EndDate.Year(), req.EndDate.Month(), req.EndDate.Day(), 23, 59, 59, 999999999, req.EndDate.Location())
	} else {
		// Default atau time range yang sudah ditentukan
		switch req.TimeRange {
		case "30days":
			startDate = endDate.AddDate(0, 0, -30)
		case "3months":
			startDate = endDate.AddDate(0, -3, 0)
		default: // "7days" atau kosong
			startDate = endDate.AddDate(0, 0, -7)
		}
	}

	return startDate, endDate, nil
}

// filterByMetrics memfilter data berdasarkan metrik yang diminta
func (s *HealthDataService) filterByMetrics(data []entity.HealthData, metrics []string) []entity.HealthData {
	if len(metrics) == 0 {
		return data // Jika tidak ada filter, kembalikan semua
	}

	// Untuk saat ini, kita tetap kembalikan semua data karena semua metrik ada dalam satu record
	// Filter akan dilakukan saat menghitung statistik dan grafik
	return data
}

// calculateSummary menghitung ringkasan statistik
func (s *HealthDataService) calculateSummary(data, prevData []entity.HealthData, metrics []string) response.HealthSummaryResponse {
	summary := response.HealthSummaryResponse{}

	// Cek apakah metrik diminta atau tidak ada filter
	includeAll := len(metrics) == 0
	includeBP := includeAll || s.containsMetric(metrics, "tekanan_darah")
	includeBS := includeAll || s.containsMetric(metrics, "gula_darah")
	includeWeight := includeAll || s.containsMetric(metrics, "berat_badan")
	includeActivity := includeAll || s.containsMetric(metrics, "aktivitas")

	if includeBP && len(data) > 0 {
		summary.BloodPressure = s.calculateBloodPressureSummary(data, prevData)
	}

	if includeBS && len(data) > 0 {
		summary.BloodSugar = s.calculateBloodSugarSummary(data, prevData)
	}

	if includeWeight && len(data) > 0 {
		summary.Weight = s.calculateWeightSummary(data, prevData)
	}

	if includeActivity && len(data) > 0 {
		summary.Activity = s.calculateActivitySummary(data, prevData)
	}

	return summary
}

// calculateBloodPressureSummary menghitung ringkasan tekanan darah
// Menggunakan safe dereference untuk menghindari panic
func (s *HealthDataService) calculateBloodPressureSummary(data, prevData []entity.HealthData) *response.BloodPressureSummary {
	if len(data) == 0 {
		return nil
	}

	var sumSystolic, sumDiastolic float64
	var countSystolic, countDiastolic int
	for _, d := range data {
		// Safe dereference dengan nil check eksplisit
		if d.Systolic != nil {
			systolic := *d.Systolic
			sumSystolic += float64(systolic)
			countSystolic++
		}
		if d.Diastolic != nil {
			diastolic := *d.Diastolic
			sumDiastolic += float64(diastolic)
			countDiastolic++
		}
	}

	// Guard clause: jika tidak ada data yang valid, return nil
	if countSystolic == 0 && countDiastolic == 0 {
		return nil
	}

	var avgSystolic, avgDiastolic float64
	if countSystolic > 0 {
		avgSystolic = sumSystolic / float64(countSystolic)
	}
	if countDiastolic > 0 {
		avgDiastolic = sumDiastolic / float64(countDiastolic)
	}

	// Hitung rata-rata periode sebelumnya dengan safe dereference
	var prevAvgSystolic, prevAvgDiastolic float64
	var prevCountSystolic, prevCountDiastolic int
	if len(prevData) > 0 {
		for _, d := range prevData {
			if d.Systolic != nil {
				systolic := *d.Systolic
				prevAvgSystolic += float64(systolic)
				prevCountSystolic++
			}
			if d.Diastolic != nil {
				diastolic := *d.Diastolic
				prevAvgDiastolic += float64(diastolic)
				prevCountDiastolic++
			}
		}
		if prevCountSystolic > 0 {
			prevAvgSystolic /= float64(prevCountSystolic)
		}
		if prevCountDiastolic > 0 {
			prevAvgDiastolic /= float64(prevCountDiastolic)
		}
	}

	// Hitung persentase perubahan dengan guard clause untuk avoid division by zero
	changePercent := 0.0
	if prevAvgSystolic > 0 {
		changePercent = ((avgSystolic - prevAvgSystolic) / prevAvgSystolic) * 100
	}

	return &response.BloodPressureSummary{
		AvgSystolic:     roundTo2Decimals(avgSystolic),
		AvgDiastolic:    roundTo2Decimals(avgDiastolic),
		ChangePercent:   roundTo2Decimals(changePercent),
		SystolicStatus:  s.getBloodPressureStatus(avgSystolic, true),
		DiastolicStatus: s.getBloodPressureStatus(avgDiastolic, false),
		NormalRange:     "90-120 / 60-80 mmHg",
	}
}

// calculateBloodSugarSummary menghitung ringkasan gula darah
// Menggunakan safe dereference untuk menghindari panic
func (s *HealthDataService) calculateBloodSugarSummary(data, prevData []entity.HealthData) *response.BloodSugarSummary {
	if len(data) == 0 {
		return nil
	}

	var sum float64
	var count int
	for _, d := range data {
		// Safe dereference dengan nil check eksplisit
		if d.BloodSugar != nil {
			bloodSugar := *d.BloodSugar
			sum += float64(bloodSugar)
			count++
		}
	}
	
	// Guard clause: jika tidak ada data yang valid, return nil
	if count == 0 {
		return nil
	}
	avgValue := sum / float64(count)

	// Hitung rata-rata periode sebelumnya dengan safe dereference
	var prevAvg float64
	var prevCount int
	if len(prevData) > 0 {
		for _, d := range prevData {
			if d.BloodSugar != nil {
				bloodSugar := *d.BloodSugar
				prevAvg += float64(bloodSugar)
				prevCount++
			}
		}
		if prevCount > 0 {
			prevAvg /= float64(prevCount)
		}
	}

	// Hitung persentase perubahan dengan guard clause untuk avoid division by zero
	changePercent := 0.0
	if prevAvg > 0 {
		changePercent = ((avgValue - prevAvg) / prevAvg) * 100
	}

	return &response.BloodSugarSummary{
		AvgValue:     roundTo2Decimals(avgValue),
		ChangePercent: roundTo2Decimals(changePercent),
		Status:       s.getBloodSugarStatus(avgValue),
		NormalRange:  "70-100 mg/dL",
	}
}

// calculateWeightSummary menghitung ringkasan berat badan
// Menggunakan safe dereference untuk menghindari panic
func (s *HealthDataService) calculateWeightSummary(data, prevData []entity.HealthData) *response.WeightSummary {
	if len(data) == 0 {
		return nil
	}

	var sum float64
	var count int
	for _, d := range data {
		// Safe dereference dengan nil check eksplisit
		if d.Weight != nil {
			weight := *d.Weight
			// Guard clause: skip jika weight <= 0 (invalid data)
			if weight > 0 {
				sum += weight
				count++
			}
		}
	}
	
	// Guard clause: jika tidak ada data yang valid, return nil
	if count == 0 {
		return nil
	}
	avgWeight := sum / float64(count)

	// Hitung rata-rata periode sebelumnya dengan safe dereference
	var prevAvg float64
	var prevCount int
	if len(prevData) > 0 {
		for _, d := range prevData {
			if d.Weight != nil {
				weight := *d.Weight
				// Guard clause: skip jika weight <= 0 (invalid data)
				if weight > 0 {
					prevAvg += weight
					prevCount++
				}
			}
		}
		if prevCount > 0 {
			prevAvg /= float64(prevCount)
		}
	}

	// Hitung persentase perubahan dengan guard clause untuk avoid division by zero
	changePercent := 0.0
	if prevAvg > 0 {
		changePercent = ((avgWeight - prevAvg) / prevAvg) * 100
	}

	// Tentukan tren dengan threshold yang jelas
	trend := "Stabil"
	if changePercent > 1 {
		trend = "Naik"
	} else if changePercent < -1 {
		trend = "Turun"
	}

	// BMI tidak dihitung karena tidak ada tinggi badan di entity
	return &response.WeightSummary{
		AvgWeight:    roundTo2Decimals(avgWeight),
		BMI:          nil,
		Trend:        trend,
		ChangePercent: roundTo2Decimals(changePercent),
	}
}

// calculateActivitySummary menghitung ringkasan aktivitas
func (s *HealthDataService) calculateActivitySummary(data, prevData []entity.HealthData) *response.ActivitySummary {
	if len(data) == 0 {
		return nil
	}

	// Untuk aktivitas, kita hitung total langkah dan kalori dari activity field
	// Karena struktur activity adalah string, kita asumsikan format tertentu atau hitung berdasarkan data yang ada
	// Untuk sementara, kita hitung jumlah record sebagai proxy untuk aktivitas
	totalSteps := len(data) * 1000 // Estimasi, karena tidak ada field steps terpisah
	totalCalories := float64(len(data)) * 200.0 // Estimasi

	// Hitung periode sebelumnya
	prevTotalSteps := len(prevData) * 1000

	// Hitung persentase perubahan
	changePercent := 0.0
	if prevTotalSteps > 0 {
		changePercent = ((float64(totalSteps) - float64(prevTotalSteps)) / float64(prevTotalSteps)) * 100
	}

	return &response.ActivitySummary{
		TotalSteps:    totalSteps,
		TotalCalories: roundTo2Decimals(totalCalories),
		ChangePercent: roundTo2Decimals(changePercent),
	}
}

// calculateTrendCharts menghitung data untuk grafik tren
func (s *HealthDataService) calculateTrendCharts(data []entity.HealthData, metrics []string) response.TrendChartsResponse {
	charts := response.TrendChartsResponse{}

	// Cek apakah metrik diminta atau tidak ada filter
	includeAll := len(metrics) == 0
	includeBP := includeAll || s.containsMetric(metrics, "tekanan_darah")
	includeBS := includeAll || s.containsMetric(metrics, "gula_darah")
	includeWeight := includeAll || s.containsMetric(metrics, "berat_badan")
	includeActivity := includeAll || s.containsMetric(metrics, "aktivitas")

	if includeBP {
		charts.BloodPressure = s.buildBloodPressureTrend(data)
	}

	if includeBS {
		charts.BloodSugar = s.buildBloodSugarTrend(data)
	}

	if includeWeight {
		charts.Weight = s.buildWeightTrend(data)
	}

	if includeActivity {
		charts.Activity = s.buildActivityTrend(data)
	}

	return charts
}

// buildBloodPressureTrend membangun data tren tekanan darah per hari
func (s *HealthDataService) buildBloodPressureTrend(data []entity.HealthData) []response.BloodPressureTrendPoint {
	// Group by date
	dateMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		dateStr := d.CreatedAt.Format("2006-01-02")
		dateMap[dateStr] = append(dateMap[dateStr], d)
	}

	// Hitung rata-rata per hari dengan safe dereference
	var points []response.BloodPressureTrendPoint
	for dateStr, dayData := range dateMap {
		var sumSystolic, sumDiastolic float64
		var countSystolic, countDiastolic int
		for _, d := range dayData {
			// Safe dereference dengan nil check eksplisit
			if d.Systolic != nil {
				systolic := *d.Systolic
				sumSystolic += float64(systolic)
				countSystolic++
			}
			if d.Diastolic != nil {
				diastolic := *d.Diastolic
				sumDiastolic += float64(diastolic)
				countDiastolic++
			}
		}
		var avgSystolic, avgDiastolic float64
		if countSystolic > 0 {
			avgSystolic = sumSystolic / float64(countSystolic)
		}
		if countDiastolic > 0 {
			avgDiastolic = sumDiastolic / float64(countDiastolic)
		}
		points = append(points, response.BloodPressureTrendPoint{
			Date:      dateStr,
			Systolic:  roundTo2Decimals(avgSystolic),
			Diastolic: roundTo2Decimals(avgDiastolic),
		})
	}

	// Sort by date
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points
}

// buildBloodSugarTrend membangun data tren gula darah per hari
func (s *HealthDataService) buildBloodSugarTrend(data []entity.HealthData) []response.BloodSugarTrendPoint {
	// Group by date
	dateMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		dateStr := d.CreatedAt.Format("2006-01-02")
		dateMap[dateStr] = append(dateMap[dateStr], d)
	}

	// Hitung rata-rata per hari dengan safe dereference
	var points []response.BloodSugarTrendPoint
	for dateStr, dayData := range dateMap {
		var sum float64
		var count int
		for _, d := range dayData {
			// Safe dereference dengan nil check eksplisit
			if d.BloodSugar != nil {
				bloodSugar := *d.BloodSugar
				sum += float64(bloodSugar)
				count++
			}
		}
		var avgValue float64
		if count > 0 {
			avgValue = sum / float64(count)
		}
		points = append(points, response.BloodSugarTrendPoint{
			Date:     dateStr,
			AvgValue: roundTo2Decimals(avgValue),
		})
	}

	// Sort by date
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points
}

// buildWeightTrend membangun data tren berat badan per hari
func (s *HealthDataService) buildWeightTrend(data []entity.HealthData) []response.WeightTrendPoint {
	// Group by date
	dateMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		dateStr := d.CreatedAt.Format("2006-01-02")
		dateMap[dateStr] = append(dateMap[dateStr], d)
	}

	// Hitung rata-rata per hari dengan safe dereference
	var points []response.WeightTrendPoint
	for dateStr, dayData := range dateMap {
		var sum float64
		var count int
		for _, d := range dayData {
			// Safe dereference dengan nil check eksplisit dan validasi > 0
			if d.Weight != nil {
				weight := *d.Weight
				// Guard clause: skip jika weight <= 0 (invalid data)
				if weight > 0 {
					sum += weight
					count++
				}
			}
		}
		var avgWeight float64
		if count > 0 {
			avgWeight = sum / float64(count)
		}
		points = append(points, response.WeightTrendPoint{
			Date:   dateStr,
			Weight: roundTo2Decimals(avgWeight),
		})
	}

	// Sort by date
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points
}

// buildActivityTrend membangun data tren aktivitas per hari
func (s *HealthDataService) buildActivityTrend(data []entity.HealthData) []response.ActivityTrendPoint {
	// Group by date
	dateMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		dateStr := d.CreatedAt.Format("2006-01-02")
		dateMap[dateStr] = append(dateMap[dateStr], d)
	}

	// Hitung total per hari
	var points []response.ActivityTrendPoint
	for dateStr, dayData := range dateMap {
		// Estimasi langkah dan kalori
		steps := len(dayData) * 1000
		calories := float64(len(dayData)) * 200.0
		points = append(points, response.ActivityTrendPoint{
			Date:     dateStr,
			Steps:    steps,
			Calories: roundTo2Decimals(calories),
		})
	}

	// Sort by date
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points
}

// buildReadingHistory membangun catatan pembacaan kronologis
func (s *HealthDataService) buildReadingHistory(data []entity.HealthData) []response.ReadingHistoryResponse {
	var history []response.ReadingHistoryResponse

	// Sort by created_at DESC (terbaru ke terlama)
	sortedData := make([]entity.HealthData, len(data))
	copy(sortedData, data)
	sort.Slice(sortedData, func(i, j int) bool {
		return sortedData[i].CreatedAt.After(sortedData[j].CreatedAt)
	})

	for _, d := range sortedData {
		// Tekanan darah - safe dereference dengan nil check eksplisit
		if d.Systolic != nil && d.Diastolic != nil {
			systolic := *d.Systolic
			diastolic := *d.Diastolic
			history = append(history, response.ReadingHistoryResponse{
				ID:         d.ID,
				DateTime:   d.CreatedAt,
				MetricType: "tekanan_darah",
				Value:      fmt.Sprintf("%d/%d mmHg", systolic, diastolic),
				Context:    nil,
				Status:     s.getBloodPressureStatus(float64(systolic), true),
				Notes:      nil,
			})
		}

		// Gula darah - safe dereference
		if d.BloodSugar != nil {
			bloodSugar := *d.BloodSugar
			history = append(history, response.ReadingHistoryResponse{
				ID:         d.ID,
				DateTime:   d.CreatedAt,
				MetricType: "gula_darah",
				Value:      fmt.Sprintf("%d mg/dL", bloodSugar),
				Context:    nil,
				Status:     s.getBloodSugarStatus(float64(bloodSugar)),
				Notes:      nil,
			})
		}

		// Berat badan - safe dereference dengan validasi > 0
		if d.Weight != nil {
			weight := *d.Weight
			// Guard clause: skip jika weight <= 0 (invalid data)
			if weight > 0 {
				history = append(history, response.ReadingHistoryResponse{
					ID:         d.ID,
					DateTime:   d.CreatedAt,
					MetricType: "berat_badan",
					Value:      fmt.Sprintf("%.2f kg", weight),
					Context:    nil,
					Status:     "Normal", // Berat badan tidak punya status abnormal
					Notes:      nil,
				})
			}
		}

		// Detak jantung - safe dereference
		if d.HeartRate != nil {
			heartRate := *d.HeartRate
			history = append(history, response.ReadingHistoryResponse{
				ID:         d.ID,
				DateTime:   d.CreatedAt,
				MetricType: "detak_jantung",
				Value:      fmt.Sprintf("%d bpm", heartRate),
				Context:    nil,
				Status:     s.getHeartRateStatus(float64(heartRate)),
				Notes:      nil,
			})
		}

		// Aktivitas - safe dereference dengan check empty string
		if utils.IsNotEmptyString(d.Activity) {
			activity := *d.Activity
			history = append(history, response.ReadingHistoryResponse{
				ID:         d.ID,
				DateTime:   d.CreatedAt,
				MetricType: "aktivitas",
				Value:      activity,
				Context:    nil,
				Status:     "Normal",
				Notes:      nil,
			})
		}
	}

	return history
}

// Helper functions untuk status

// getBloodPressureStatus menentukan status tekanan darah
func (s *HealthDataService) getBloodPressureStatus(value float64, isSystolic bool) string {
	if isSystolic {
		if value >= 90 && value <= 120 {
			return "Normal"
		} else if value >= 121 && value <= 139 {
			return "Perhatian"
		} else {
			return "Abnormal"
		}
	} else {
		if value >= 60 && value <= 80 {
			return "Normal"
		} else if value >= 81 && value <= 89 {
			return "Perhatian"
		} else {
			return "Abnormal"
		}
	}
}

// getBloodSugarStatus menentukan status gula darah
func (s *HealthDataService) getBloodSugarStatus(value float64) string {
	if value >= 70 && value <= 100 {
		return "Normal"
	} else if (value >= 101 && value <= 125) || (value >= 60 && value <= 69) {
		return "Perhatian"
	} else {
		return "Abnormal"
	}
}

// getHeartRateStatus menentukan status detak jantung
func (s *HealthDataService) getHeartRateStatus(value float64) string {
	if value >= 60 && value <= 100 {
		return "Normal"
	} else if (value >= 101 && value <= 120) || (value >= 50 && value <= 59) {
		return "Perhatian"
	} else {
		return "Abnormal"
	}
}

// containsMetric mengecek apakah metrik ada dalam slice
func (s *HealthDataService) containsMetric(metrics []string, metric string) bool {
	for _, m := range metrics {
		if strings.ToLower(m) == strings.ToLower(metric) {
			return true
		}
	}
	return false
}

// roundTo2Decimals membulatkan float ke 2 desimal
func roundTo2Decimals(num float64) float64 {
	return math.Round(num*100) / 100
}

// ==================== OPERASI UNTUK DOWNLOAD LAPORAN ====================

// GenerateReportCSV menghasilkan laporan dalam format CSV
func (s *HealthDataService) GenerateReportCSV(userID uint, req *request.HealthHistoryRequest) (*bytes.Buffer, string, error) {
	// Ambil data riwayat kesehatan
	historyResp, err := s.GetHealthHistory(userID, req)
	if err != nil {
		return nil, "", err
	}

	// Buat buffer untuk CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Tentukan rentang waktu untuk nama file
	startDate, endDate, _ := s.parseTimeRange(req)
	timeRangeStr := fmt.Sprintf("%s_to_%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	filename := fmt.Sprintf("riwayat_kesehatan_%s.csv", timeRangeStr)

	// Header CSV
	headers := []string{
		"Tanggal & Waktu",
		"Jenis Metrik",
		"Nilai",
		"Status",
		"Konteks",
		"Catatan",
	}
	if err := writer.Write(headers); err != nil {
		return nil, "", err
	}

	// Tulis data reading history
	for _, record := range historyResp.ReadingHistory {
		context := ""
		if record.Context != nil {
			context = *record.Context
		}
		notes := ""
		if record.Notes != nil {
			notes = *record.Notes
		}

		row := []string{
			record.DateTime.Format("2006-01-02 15:04:05"),
			record.MetricType,
			record.Value,
			record.Status,
			context,
			notes,
		}
		if err := writer.Write(row); err != nil {
			return nil, "", err
		}
	}

	// Tambahkan ringkasan statistik
	writer.Write([]string{""})
	writer.Write([]string{"=== RINGKASAN STATISTIK ==="})

	// Tekanan Darah
	if historyResp.Summary.BloodPressure != nil {
		writer.Write([]string{""})
		writer.Write([]string{"TEKANAN DARAH"})
		writer.Write([]string{"Rata-rata Systolic", fmt.Sprintf("%.2f mmHg", historyResp.Summary.BloodPressure.AvgSystolic)})
		writer.Write([]string{"Rata-rata Diastolic", fmt.Sprintf("%.2f mmHg", historyResp.Summary.BloodPressure.AvgDiastolic)})
		writer.Write([]string{"Persentase Perubahan", fmt.Sprintf("%.2f%%", historyResp.Summary.BloodPressure.ChangePercent)})
		writer.Write([]string{"Status Systolic", historyResp.Summary.BloodPressure.SystolicStatus})
		writer.Write([]string{"Status Diastolic", historyResp.Summary.BloodPressure.DiastolicStatus})
		writer.Write([]string{"Rentang Normal", historyResp.Summary.BloodPressure.NormalRange})
	}

	// Gula Darah
	if historyResp.Summary.BloodSugar != nil {
		writer.Write([]string{""})
		writer.Write([]string{"GULA DARAH"})
		writer.Write([]string{"Rata-rata", fmt.Sprintf("%.2f mg/dL", historyResp.Summary.BloodSugar.AvgValue)})
		writer.Write([]string{"Persentase Perubahan", fmt.Sprintf("%.2f%%", historyResp.Summary.BloodSugar.ChangePercent)})
		writer.Write([]string{"Status", historyResp.Summary.BloodSugar.Status})
		writer.Write([]string{"Rentang Normal", historyResp.Summary.BloodSugar.NormalRange})
	}

	// Berat Badan
	if historyResp.Summary.Weight != nil {
		weightSummary := historyResp.Summary.Weight
		writer.Write([]string{""})
		writer.Write([]string{"BERAT BADAN"})
		writer.Write([]string{"Rata-rata", fmt.Sprintf("%.2f kg", weightSummary.AvgWeight)})
		writer.Write([]string{"Tren", weightSummary.Trend})
		writer.Write([]string{"Persentase Perubahan", fmt.Sprintf("%.2f%%", weightSummary.ChangePercent)})
		// Safe dereference untuk BMI
		if weightSummary.BMI != nil {
			bmi := *weightSummary.BMI
			writer.Write([]string{"BMI", fmt.Sprintf("%.2f", bmi)})
		}
	}

	// Aktivitas
	if historyResp.Summary.Activity != nil {
		writer.Write([]string{""})
		writer.Write([]string{"AKTIVITAS"})
		writer.Write([]string{"Total Langkah", fmt.Sprintf("%d", historyResp.Summary.Activity.TotalSteps)})
		writer.Write([]string{"Total Kalori", fmt.Sprintf("%.2f", historyResp.Summary.Activity.TotalCalories)})
		writer.Write([]string{"Persentase Perubahan", fmt.Sprintf("%.2f%%", historyResp.Summary.Activity.ChangePercent)})
	}

	writer.Flush()
	return &buf, filename, nil
}

// GenerateReportJSON menghasilkan laporan dalam format JSON
func (s *HealthDataService) GenerateReportJSON(userID uint, req *request.HealthHistoryRequest) (*bytes.Buffer, string, error) {
	// Ambil data riwayat kesehatan
	historyResp, err := s.GetHealthHistory(userID, req)
	if err != nil {
		return nil, "", err
	}

	// Tentukan rentang waktu untuk nama file
	startDate, endDate, _ := s.parseTimeRange(req)
	timeRangeStr := fmt.Sprintf("%s_to_%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	filename := fmt.Sprintf("riwayat_kesehatan_%s.json", timeRangeStr)

	// Buat struktur laporan lengkap
	report := map[string]interface{}{
		"periode": map[string]interface{}{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
			"time_range": req.TimeRange,
		},
		"ringkasan_statistik": historyResp.Summary,
		"grafik_tren":         historyResp.TrendCharts,
		"catatan_pembacaan":   historyResp.ReadingHistory,
		"generated_at":        time.Now().Format("2006-01-02 15:04:05"),
	}

	// Marshal ke JSON dengan indent
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	buf.Write(jsonData)

	return &buf, filename, nil
}

// GetHealthAlerts membuat daftar alert kesehatan untuk ditampilkan di UI.
func (s *HealthDataService) GetHealthAlerts(userID uint, req *request.HealthHistoryRequest) ([]response.HealthAlertResponse, error) {
	// Tentukan rentang waktu (pakai logika yang sama seperti riwayat)
	startDate, endDate, err := s.parseTimeRange(req)
	if err != nil {
		return nil, err
	}

	// Ambil data kesehatan user dalam rentang waktu
	healthDataList, err := s.healthDataRepo.GetHealthDataByUserIDWithFilter(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	if len(healthDataList) == 0 {
		// Tidak ada data, tidak ada alert
		return []response.HealthAlertResponse{}, nil
	}

	// Urutkan data berdasarkan waktu (paling lama → paling baru)
	sorted := make([]entity.HealthData, len(healthDataList))
	copy(sorted, healthDataList)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})

	first := sorted[0]                       // data paling awal di periode ini
	latest := sorted[len(sorted)-1]          // data terbaru
	alerts := []response.HealthAlertResponse{} // hasilnya nanti

	// ========== 1. ALERT TEKANAN DARAH ==========
	// Safe dereference dengan nil check eksplisit
	if latest.Systolic != nil && latest.Diastolic != nil {
		systolic := *latest.Systolic
		diastolic := *latest.Diastolic
		systolicStatus := s.getBloodPressureStatus(float64(systolic), true)
		diastolicStatus := s.getBloodPressureStatus(float64(diastolic), false)

		if systolicStatus != "Normal" || diastolicStatus != "Normal" {
			level := "Perhatian"
			subtitle := "Perhatian!"
			statusText := "Tekanan Darah Perlu Dipantau"

			if systolicStatus == "Abnormal" || diastolicStatus == "Abnormal" {
				level = "Tinggi"
				statusText = "Tekanan Darah Tinggi (Hipertensi)"
			}

			value := fmt.Sprintf("%d / %d", systolic, diastolic)
			unit := "mmHg"

			alerts = append(alerts, response.HealthAlertResponse{
				Type:       "blood_pressure",
				Level:      level,
				Title:      "Tekanan Darah Anda",
				Subtitle:   subtitle,
				Value:      value,
				Unit:       unit,
				StatusText: statusText,
				RecordedAt: latest.CreatedAt,
			})
		}
	}

	// ========== 2. ALERT GULA DARAH ==========
	// Safe dereference dengan nil check eksplisit
	if latest.BloodSugar != nil {
		bloodSugar := *latest.BloodSugar
		bsStatus := s.getBloodSugarStatus(float64(bloodSugar))
		if bsStatus != "Normal" {
			level := "Perhatian"
			subtitle := "Gula Darah Tidak Normal"
			statusText := "Gula Darah Perlu Dipantau"

			if bsStatus == "Abnormal" {
				level = "Kritis"
				if bloodSugar < 70 {
					subtitle = "Gula Darah Rendah"
					statusText = "Hipoglikemia (Gula Darah Rendah)"
				} else {
					subtitle = "Gula Darah Tinggi"
					statusText = "Hiperglikemia (Gula Darah Tinggi)"
				}
			}

			value := fmt.Sprintf("%d", bloodSugar)
			unit := "mg/dL"

			alerts = append(alerts, response.HealthAlertResponse{
				Type:       "blood_sugar",
				Level:      level,
				Title:      "Gula Darah Anda",
				Subtitle:   subtitle,
				Value:      value,
				Unit:       unit,
				StatusText: statusText,
				RecordedAt: latest.CreatedAt,
			})
		}
	}

	// ========== 3. ALERT PENURUNAN BERAT BADAN ==========
	// Bandingkan berat pertama & terbaru dalam periode dengan safe dereference
	if first.Weight != nil && latest.Weight != nil {
		firstWeight := *first.Weight
		latestWeight := *latest.Weight
		
		// Guard clause: pastikan kedua weight valid (> 0)
		if firstWeight > 0 && latestWeight > 0 {
			weightDiff := firstWeight - latestWeight // kalau turun, hasilnya positif
			if weightDiff >= 5 { // turun >= 5 kg dalam periode
				level := "Sedang"
				statusText := "Penurunan Signifikan"

				days := int(latest.CreatedAt.Sub(first.CreatedAt).Hours() / 24)
				if days <= 0 {
					days = 1
				}
				extraText := fmt.Sprintf("↓ %.1f kg dalam %d hari", roundTo2Decimals(weightDiff), days)

				alerts = append(alerts, response.HealthAlertResponse{
					Type:       "weight_loss",
					Level:      level,
					Title:      "Berat Badan Anda",
					Subtitle:   "Penurunan Berat Badan",
					Value:      fmt.Sprintf("%.2f", latestWeight),
					Unit:       "kg",
					Extra:      &extraText,
					StatusText: statusText,
					RecordedAt: latest.CreatedAt,
				})
			}
		}
	}

	return alerts, nil
}

func (s *HealthDataService) generateAlertsFromHealthData(d entity.HealthData) []entity.HealthAlert {
	var alerts []entity.HealthAlert
	recordedAt := d.CreatedAt
	if recordedAt.IsZero() {
		recordedAt = time.Now()
	}

	// 1) Blood Pressure Alert - safe dereference dengan nil check eksplisit
	if d.Systolic != nil && d.Diastolic != nil {
		systolic := *d.Systolic
		diastolic := *d.Diastolic
		if systolic >= 140 || diastolic >= 90 {
			level := "Tinggi"
			statusText := "Tekanan Darah Tinggi (Hipertensi)"
			if systolic >= 160 || diastolic >= 100 {
				level = "Kritis"
				statusText = "Tekanan Darah Sangat Tinggi"
			}

			alerts = append(alerts, entity.HealthAlert{
				UserID:        d.UserID,
				HealthDataID:  d.ID,
				Type:          "blood_pressure",
				Level:         level,
				Title:         "Tekanan Darah Anda",
				Subtitle:      "Perhatian!",
				Value:         fmt.Sprintf("%d / %d", systolic, diastolic),
				Unit:          "mmHg",
				StatusText:    statusText,
				RecordedAt:    recordedAt,
			})
		}
	}

	// 2) Blood Sugar Alert - safe dereference dengan nil check eksplisit
	if d.BloodSugar != nil {
		bloodSugar := *d.BloodSugar
		if bloodSugar >= 140 {
			level := "Tinggi"
			statusText := "Gula Darah Tinggi"
			if bloodSugar >= 180 {
				level = "Kritis"
				statusText = "Hiperglikemia (Gula Darah Tinggi)"
			}
			alerts = append(alerts, entity.HealthAlert{
				UserID:        d.UserID,
				HealthDataID:  d.ID,
				Type:          "blood_sugar",
				Level:         level,
				Title:         "Gula Darah Anda",
				Subtitle:      "Gula Darah Tinggi",
				Value:         fmt.Sprintf("%d", bloodSugar),
				Unit:          "mg/dL",
				StatusText:    statusText,
				RecordedAt:    recordedAt,
			})
		}
	}

	// 3) Heart Rate Alert - safe dereference dengan nil check eksplisit
	if d.HeartRate != nil {
		heartRate := *d.HeartRate
		if heartRate >= 101 {
			level := "Tinggi"
			statusText := "Detak Jantung Tinggi"
			if heartRate >= 120 {
				level = "Kritis"
				statusText = "Detak Jantung Sangat Tinggi"
			}
			alerts = append(alerts, entity.HealthAlert{
				UserID:        d.UserID,
				HealthDataID:  d.ID,
				Type:          "heart_rate",
				Level:         level,
				Title:         "Detak Jantung Anda",
				Subtitle:      "Perhatian!",
				Value:         fmt.Sprintf("%d", heartRate),
				Unit:          "bpm",
				StatusText:    statusText,
				RecordedAt:    recordedAt,
			})
		}
	}

	return alerts
}


