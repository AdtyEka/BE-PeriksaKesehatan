package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
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

// HealthDataService menangani business logic untuk data kesehatan
type HealthDataService struct {
	healthDataRepo *repository.HealthDataRepository
}

// NewHealthDataService membuat instance baru dari HealthDataService
func NewHealthDataService(healthDataRepo *repository.HealthDataRepository) *HealthDataService {
	return &HealthDataService{
		healthDataRepo: healthDataRepo,
	}
}

// ValidateHealthData melakukan validasi range nilai data kesehatan
func (s *HealthDataService) ValidateHealthData(req *request.HealthDataRequest) error {
	// Validasi systolic: 90 – 180
	if req.Systolic < 90 || req.Systolic > 180 {
		return errors.New("systolic harus berada dalam range 90-180 mmHg")
	}

	// Validasi diastolic: 60 – 120
	if req.Diastolic < 60 || req.Diastolic > 120 {
		return errors.New("diastolic harus berada dalam range 60-120 mmHg")
	}

	// Validasi blood_sugar: 60 – 300
	if req.BloodSugar < 60 || req.BloodSugar > 300 {
		return errors.New("blood_sugar harus berada dalam range 60-300 mg/dL")
	}

	// Validasi weight: 20 – 200
	if req.Weight < 20 || req.Weight > 200 {
		return errors.New("weight harus berada dalam range 20-200 kg")
	}

	// Validasi heart_rate: 40 – 180
	if req.HeartRate < 40 || req.HeartRate > 180 {
		return errors.New("heart_rate harus berada dalam range 40-180 bpm")
	}

	return nil
}

// CreateHealthData membuat data kesehatan baru dengan validasi
func (s *HealthDataService) CreateHealthData(userID uint, req *request.HealthDataRequest) (*response.HealthDataResponse, error) {
	// Validasi data
	if err := s.ValidateHealthData(req); err != nil {
		return nil, err
	}

	// Buat entity dari request
	healthData := &entity.HealthData{
		UserID:     userID,
		Systolic:   req.Systolic,
		Diastolic:  req.Diastolic,
		BloodSugar: req.BloodSugar,
		Weight:     req.Weight,
		HeartRate:  req.HeartRate,
		Activity:   req.Activity,
	}

	// Simpan ke database
	if err := s.healthDataRepo.CreateHealthData(healthData); err != nil {
		return nil, err
	}

	// Buat response
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
func (s *HealthDataService) calculateBloodPressureSummary(data, prevData []entity.HealthData) *response.BloodPressureSummary {
	if len(data) == 0 {
		return nil
	}

	var sumSystolic, sumDiastolic float64
	for _, d := range data {
		sumSystolic += float64(d.Systolic)
		sumDiastolic += float64(d.Diastolic)
	}

	avgSystolic := sumSystolic / float64(len(data))
	avgDiastolic := sumDiastolic / float64(len(data))

	// Hitung rata-rata periode sebelumnya
	var prevAvgSystolic, prevAvgDiastolic float64
	if len(prevData) > 0 {
		for _, d := range prevData {
			prevAvgSystolic += float64(d.Systolic)
			prevAvgDiastolic += float64(d.Diastolic)
		}
		prevAvgSystolic /= float64(len(prevData))
		prevAvgDiastolic /= float64(len(prevData))
	}

	// Hitung persentase perubahan
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
func (s *HealthDataService) calculateBloodSugarSummary(data, prevData []entity.HealthData) *response.BloodSugarSummary {
	if len(data) == 0 {
		return nil
	}

	var sum float64
	for _, d := range data {
		sum += float64(d.BloodSugar)
	}
	avgValue := sum / float64(len(data))

	// Hitung rata-rata periode sebelumnya
	var prevAvg float64
	if len(prevData) > 0 {
		for _, d := range prevData {
			prevAvg += float64(d.BloodSugar)
		}
		prevAvg /= float64(len(prevData))
	}

	// Hitung persentase perubahan
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
func (s *HealthDataService) calculateWeightSummary(data, prevData []entity.HealthData) *response.WeightSummary {
	if len(data) == 0 {
		return nil
	}

	var sum float64
	for _, d := range data {
		sum += d.Weight
	}
	avgWeight := sum / float64(len(data))

	// Hitung rata-rata periode sebelumnya
	var prevAvg float64
	if len(prevData) > 0 {
		for _, d := range prevData {
			prevAvg += d.Weight
		}
		prevAvg /= float64(len(prevData))
	}

	// Hitung persentase perubahan
	changePercent := 0.0
	if prevAvg > 0 {
		changePercent = ((avgWeight - prevAvg) / prevAvg) * 100
	}

	// Tentukan tren
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

	// Hitung rata-rata per hari
	var points []response.BloodPressureTrendPoint
	for dateStr, dayData := range dateMap {
		var sumSystolic, sumDiastolic float64
		for _, d := range dayData {
			sumSystolic += float64(d.Systolic)
			sumDiastolic += float64(d.Diastolic)
		}
		points = append(points, response.BloodPressureTrendPoint{
			Date:      dateStr,
			Systolic:  roundTo2Decimals(sumSystolic / float64(len(dayData))),
			Diastolic: roundTo2Decimals(sumDiastolic / float64(len(dayData))),
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

	// Hitung rata-rata per hari
	var points []response.BloodSugarTrendPoint
	for dateStr, dayData := range dateMap {
		var sum float64
		for _, d := range dayData {
			sum += float64(d.BloodSugar)
		}
		points = append(points, response.BloodSugarTrendPoint{
			Date:     dateStr,
			AvgValue: roundTo2Decimals(sum / float64(len(dayData))),
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

	// Hitung rata-rata per hari
	var points []response.WeightTrendPoint
	for dateStr, dayData := range dateMap {
		var sum float64
		for _, d := range dayData {
			sum += d.Weight
		}
		points = append(points, response.WeightTrendPoint{
			Date:   dateStr,
			Weight: roundTo2Decimals(sum / float64(len(dayData))),
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
		// Tekanan darah
		history = append(history, response.ReadingHistoryResponse{
			ID:         d.ID,
			DateTime:   d.CreatedAt,
			MetricType: "tekanan_darah",
			Value:      fmt.Sprintf("%d/%d mmHg", d.Systolic, d.Diastolic),
			Context:    nil,
			Status:     s.getBloodPressureStatus(float64(d.Systolic), true),
			Notes:      nil,
		})

		// Gula darah
		history = append(history, response.ReadingHistoryResponse{
			ID:         d.ID,
			DateTime:   d.CreatedAt,
			MetricType: "gula_darah",
			Value:      fmt.Sprintf("%d mg/dL", d.BloodSugar),
			Context:    nil,
			Status:     s.getBloodSugarStatus(float64(d.BloodSugar)),
			Notes:      nil,
		})

		// Berat badan
		history = append(history, response.ReadingHistoryResponse{
			ID:         d.ID,
			DateTime:   d.CreatedAt,
			MetricType: "berat_badan",
			Value:      fmt.Sprintf("%.2f kg", d.Weight),
			Context:    nil,
			Status:     "Normal", // Berat badan tidak punya status abnormal
			Notes:      nil,
		})

		// Detak jantung
		history = append(history, response.ReadingHistoryResponse{
			ID:         d.ID,
			DateTime:   d.CreatedAt,
			MetricType: "detak_jantung",
			Value:      fmt.Sprintf("%d bpm", d.HeartRate),
			Context:    nil,
			Status:     s.getHeartRateStatus(float64(d.HeartRate)),
			Notes:      nil,
		})

		// Aktivitas (jika ada)
		if d.Activity != nil && *d.Activity != "" {
			history = append(history, response.ReadingHistoryResponse{
				ID:         d.ID,
				DateTime:   d.CreatedAt,
				MetricType: "aktivitas",
				Value:      *d.Activity,
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
		writer.Write([]string{""})
		writer.Write([]string{"BERAT BADAN"})
		writer.Write([]string{"Rata-rata", fmt.Sprintf("%.2f kg", historyResp.Summary.Weight.AvgWeight)})
		writer.Write([]string{"Tren", historyResp.Summary.Weight.Trend})
		writer.Write([]string{"Persentase Perubahan", fmt.Sprintf("%.2f%%", historyResp.Summary.Weight.ChangePercent)})
		if historyResp.Summary.Weight.BMI != nil {
			writer.Write([]string{"BMI", fmt.Sprintf("%.2f", *historyResp.Summary.Weight.BMI)})
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

