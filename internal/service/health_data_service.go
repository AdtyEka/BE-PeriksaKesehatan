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

	"github.com/jung-kurt/gofpdf"
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

// ValidateHealthData melakukan validasi range nilai data kesehatan dengan nullable-aware.
// Minimal satu metrik kesehatan harus diisi. Jika systolic dikirim, diastolic juga harus dikirim.
func (s *HealthDataService) ValidateHealthData(req *request.HealthDataRequest) error {
	if err := utils.RequireAtLeastOneHealthMetric(
		req.Systolic, req.Diastolic, req.BloodSugar, nil, req.HeartRate, req.Weight, req.Height,
	); err != nil {
		return err
	}

	if (req.Systolic != nil && req.Diastolic == nil) || 
	   (req.Systolic == nil && req.Diastolic != nil) {
		return errors.New("systolic dan diastolic harus dikirim bersamaan")
	}

	if err := utils.ValidateNullableInt(req.Systolic, "systolic", 90, 180); err != nil {
		return err
	}

	if err := utils.ValidateNullableInt(req.Diastolic, "diastolic", 60, 120); err != nil {
		return err
	}

	if err := utils.ValidateNullableInt(req.BloodSugar, "blood_sugar", 60, 300); err != nil {
		return err
	}

	if err := utils.ValidateNullableFloat64(req.Weight, "weight", 20.0, 200.0); err != nil {
		return err
	}

	if err := utils.ValidateNullableInt(req.Height, "height", 50, 250); err != nil {
		return err
	}

	if err := utils.ValidateNullableInt(req.HeartRate, "heart_rate", 40, 180); err != nil {
		return err
	}

	return nil
}

func (s *HealthDataService) CreateHealthData(userID uint, req *request.HealthDataRequest) (*response.HealthDataResponse, error) {
	// Cek apakah ada record terakhir milik user
	existingData, err := s.healthDataRepo.GetLatestHealthDataByUserID(userID)
	if err != nil {
		return nil, err
	}

	var healthData *entity.HealthData
	var isUpdate bool

	if existingData != nil {
		// UPDATE: Merge data lama + data baru
		// Field yang tidak dikirim (nil) akan tetap menggunakan nilai lama
		healthData = existingData
		isUpdate = true

		// Untuk update, validasi lebih fleksibel: Activity saja sudah cukup
		// Tapi tetap validasi field yang dikirim jika ada
		if req.Systolic != nil || req.Diastolic != nil {
			if (req.Systolic != nil && req.Diastolic == nil) || 
			   (req.Systolic == nil && req.Diastolic != nil) {
				return nil, errors.New("systolic dan diastolic harus dikirim bersamaan")
			}
			// Setelah cek di atas, kita tahu keduanya tidak nil atau keduanya nil
			// Validasi hanya jika keduanya tidak nil
			if req.Systolic != nil && req.Diastolic != nil {
				if err := utils.ValidateNullableInt(req.Systolic, "systolic", 90, 180); err != nil {
					return nil, err
				}
				if err := utils.ValidateNullableInt(req.Diastolic, "diastolic", 60, 120); err != nil {
					return nil, err
				}
			}
		}
		if req.BloodSugar != nil {
			if err := utils.ValidateNullableInt(req.BloodSugar, "blood_sugar", 60, 300); err != nil {
				return nil, err
			}
		}
		if req.Weight != nil {
			if err := utils.ValidateNullableFloat64(req.Weight, "weight", 20.0, 200.0); err != nil {
				return nil, err
			}
		}
		if req.Height != nil {
			if err := utils.ValidateNullableInt(req.Height, "height", 50, 250); err != nil {
				return nil, err
			}
		}
		if req.HeartRate != nil {
			if err := utils.ValidateNullableInt(req.HeartRate, "heart_rate", 40, 180); err != nil {
				return nil, err
			}
		}

		// Update hanya field yang dikirim (tidak nil)
		if req.Systolic != nil {
			healthData.Systolic = req.Systolic
		}
		if req.Diastolic != nil {
			healthData.Diastolic = req.Diastolic
		}
		if req.BloodSugar != nil {
			healthData.BloodSugar = req.BloodSugar
		}
		if req.Weight != nil {
			healthData.Weight = req.Weight
		}
		if req.Height != nil {
			healthData.HeightCM = req.Height
		}
		if req.HeartRate != nil {
			healthData.HeartRate = req.HeartRate
		}
		if req.Activity != nil {
			healthData.Activity = req.Activity
		}

		// Validasi: minimal ada 1 field yang akan di-update
		hasUpdate := req.Systolic != nil || req.Diastolic != nil || req.BloodSugar != nil || 
			req.Weight != nil || req.Height != nil || req.HeartRate != nil || req.Activity != nil
		if !hasUpdate {
			return nil, errors.New("minimal satu field harus diisi untuk update")
		}

		// Lakukan update
		if err := s.healthDataRepo.UpdateHealthData(healthData); err != nil {
			return nil, err
		}

		// Reload data untuk mendapatkan updated_at terbaru
		healthData, err = s.healthDataRepo.GetHealthDataByID(healthData.ID)
		if err != nil {
			return nil, err
		}
	} else {
		// INSERT: Buat record baru
		// Untuk insert, validasi lebih ketat: minimal satu metrik kesehatan harus diisi
		if err := s.ValidateHealthData(req); err != nil {
			return nil, err
		}

		healthData = &entity.HealthData{
			UserID: userID,
		}

		if req.Systolic != nil {
			healthData.Systolic = req.Systolic
		}
		if req.Diastolic != nil {
			healthData.Diastolic = req.Diastolic
		}
		if req.BloodSugar != nil {
			healthData.BloodSugar = req.BloodSugar
		}
		if req.Weight != nil {
			healthData.Weight = req.Weight
		}
		if req.Height != nil {
			healthData.HeightCM = req.Height
		}
		if req.HeartRate != nil {
			healthData.HeartRate = req.HeartRate
		}
		if req.Activity != nil {
			healthData.Activity = req.Activity
		}

		if err := s.healthDataRepo.CreateHealthData(healthData); err != nil {
			return nil, err
		}
	}

	// Build response
	resp := &response.HealthDataResponse{
		ID:        healthData.ID,
		UserID:    healthData.UserID,
		CreatedAt: healthData.CreatedAt,
	}

	// Set semua field (termasuk yang nil) untuk response
	resp.Systolic = healthData.Systolic
	resp.Diastolic = healthData.Diastolic
	resp.BloodSugar = healthData.BloodSugar
	resp.Weight = healthData.Weight
	resp.Height = healthData.HeightCM
	resp.HeartRate = healthData.HeartRate
	resp.Activity = healthData.Activity

	_ = isUpdate // Mark as used (untuk debugging jika diperlukan)

	return resp, nil
}

// GetHealthDataByUserID mengembalikan 1 record health data terbaru milik user
// Menggunakan logic inkremental: selalu mengembalikan record aktif terbaru
func (s *HealthDataService) GetHealthDataByUserID(userID uint) (*entity.HealthData, error) {
	return s.healthDataRepo.GetLatestHealthDataByUserID(userID)
}

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

// calculateBloodPressureSummary menghitung ringkasan tekanan darah dengan nullable-aware
// Hanya menghitung dari data yang memiliki nilai tekanan darah (tidak nil)
func (s *HealthDataService) calculateBloodPressureSummary(data, prevData []entity.HealthData) *response.BloodPressureSummary {
	if len(data) == 0 {
		return nil
	}

	// Filter data yang memiliki tekanan darah (keduanya tidak nil)
	var validData []entity.HealthData
	for _, d := range data {
		if d.Systolic != nil && d.Diastolic != nil {
			validData = append(validData, d)
		}
	}

	if len(validData) == 0 {
		return nil // Tidak ada data tekanan darah yang valid
	}

	var sumSystolic, sumDiastolic float64
	for _, d := range validData {
		if d.Systolic != nil && d.Diastolic != nil {
			sumSystolic += float64(*d.Systolic)
			sumDiastolic += float64(*d.Diastolic)
		}
	}

	avgSystolic := sumSystolic / float64(len(validData))
	avgDiastolic := sumDiastolic / float64(len(validData))

	// Hitung rata-rata periode sebelumnya (hanya data yang valid)
	var prevAvgSystolic, prevAvgDiastolic float64
	var prevValidCount int
	if len(prevData) > 0 {
		for _, d := range prevData {
			if d.Systolic != nil && d.Diastolic != nil {
				prevAvgSystolic += float64(*d.Systolic)
				prevAvgDiastolic += float64(*d.Diastolic)
				prevValidCount++
			}
		}
		if prevValidCount > 0 {
			prevAvgSystolic /= float64(prevValidCount)
			prevAvgDiastolic /= float64(prevValidCount)
		}
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

// calculateBloodSugarSummary menghitung ringkasan gula darah dengan nullable-aware
// Hanya menghitung dari data yang memiliki nilai gula darah (tidak nil)
func (s *HealthDataService) calculateBloodSugarSummary(data, prevData []entity.HealthData) *response.BloodSugarSummary {
	if len(data) == 0 {
		return nil
	}

	// Filter data yang memiliki gula darah (tidak nil)
	var validData []entity.HealthData
	for _, d := range data {
		if d.BloodSugar != nil {
			validData = append(validData, d)
		}
	}

	if len(validData) == 0 {
		return nil // Tidak ada data gula darah yang valid
	}

	var sum float64
	for _, d := range validData {
		if d.BloodSugar != nil {
			sum += float64(*d.BloodSugar)
		}
	}
	avgValue := sum / float64(len(validData))

	// Hitung rata-rata periode sebelumnya (hanya data yang valid)
	var prevAvg float64
	var prevValidCount int
	if len(prevData) > 0 {
		for _, d := range prevData {
			if d.BloodSugar != nil {
				prevAvg += float64(*d.BloodSugar)
				prevValidCount++
			}
		}
		if prevValidCount > 0 {
			prevAvg /= float64(prevValidCount)
		}
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

// calculateWeightSummary menghitung ringkasan berat badan dengan nullable-aware
// Hanya menghitung dari data yang memiliki nilai berat badan (tidak nil)
func (s *HealthDataService) calculateWeightSummary(data, prevData []entity.HealthData) *response.WeightSummary {
	if len(data) == 0 {
		return nil
	}

	// Filter data yang memiliki berat badan (tidak nil)
	var validData []entity.HealthData
	for _, d := range data {
		if d.Weight != nil {
			validData = append(validData, d)
		}
	}

	if len(validData) == 0 {
		return nil // Tidak ada data berat badan yang valid
	}

	var sum float64
	for _, d := range validData {
		if d.Weight != nil {
			sum += *d.Weight
		}
	}
	avgWeight := sum / float64(len(validData))

	// Hitung rata-rata periode sebelumnya (hanya data yang valid)
	var prevAvg float64
	var prevValidCount int
	if len(prevData) > 0 {
		for _, d := range prevData {
			if d.Weight != nil {
				prevAvg += *d.Weight
				prevValidCount++
			}
		}
		if prevValidCount > 0 {
			prevAvg /= float64(prevValidCount)
		}
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

// buildBloodPressureTrend membangun data tren tekanan darah per hari dengan nullable-aware
func (s *HealthDataService) buildBloodPressureTrend(data []entity.HealthData) []response.BloodPressureTrendPoint {
	// Group by date (hanya data yang memiliki tekanan darah)
	dateMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		if d.Systolic != nil && d.Diastolic != nil {
			dateStr := d.CreatedAt.Format("2006-01-02")
			dateMap[dateStr] = append(dateMap[dateStr], d)
		}
	}

	// Hitung rata-rata per hari (hanya dari data yang valid)
	var points []response.BloodPressureTrendPoint
	for dateStr, dayData := range dateMap {
		var sumSystolic, sumDiastolic float64
		var validCount int
		for _, d := range dayData {
			if d.Systolic != nil && d.Diastolic != nil {
				sumSystolic += float64(*d.Systolic)
				sumDiastolic += float64(*d.Diastolic)
				validCount++
			}
		}
		if validCount > 0 {
			points = append(points, response.BloodPressureTrendPoint{
				Date:      dateStr,
				Systolic:  roundTo2Decimals(sumSystolic / float64(validCount)),
				Diastolic: roundTo2Decimals(sumDiastolic / float64(validCount)),
			})
		}
	}

	// Sort by date
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points
}

// buildBloodSugarTrend membangun data tren gula darah per hari dengan nullable-aware
func (s *HealthDataService) buildBloodSugarTrend(data []entity.HealthData) []response.BloodSugarTrendPoint {
	// Group by date (hanya data yang memiliki gula darah)
	dateMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		if d.BloodSugar != nil {
			dateStr := d.CreatedAt.Format("2006-01-02")
			dateMap[dateStr] = append(dateMap[dateStr], d)
		}
	}

	// Hitung rata-rata per hari (hanya dari data yang valid)
	var points []response.BloodSugarTrendPoint
	for dateStr, dayData := range dateMap {
		var sum float64
		var validCount int
		for _, d := range dayData {
			if d.BloodSugar != nil {
				sum += float64(*d.BloodSugar)
				validCount++
			}
		}
		if validCount > 0 {
			points = append(points, response.BloodSugarTrendPoint{
				Date:     dateStr,
				AvgValue: roundTo2Decimals(sum / float64(validCount)),
			})
		}
	}

	// Sort by date
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points
}

// buildWeightTrend membangun data tren berat badan per hari dengan nullable-aware
func (s *HealthDataService) buildWeightTrend(data []entity.HealthData) []response.WeightTrendPoint {
	// Group by date (hanya data yang memiliki berat badan)
	dateMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		if d.Weight != nil {
			dateStr := d.CreatedAt.Format("2006-01-02")
			dateMap[dateStr] = append(dateMap[dateStr], d)
		}
	}

	// Hitung rata-rata per hari (hanya dari data yang valid)
	var points []response.WeightTrendPoint
	for dateStr, dayData := range dateMap {
		var sum float64
		var validCount int
		for _, d := range dayData {
			if d.Weight != nil {
				sum += *d.Weight
				validCount++
			}
		}
		if validCount > 0 {
			points = append(points, response.WeightTrendPoint{
				Date:   dateStr,
				Weight: roundTo2Decimals(sum / float64(validCount)),
			})
		}
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

// buildReadingHistory membangun catatan pembacaan kronologis dengan nullable-aware
// Hanya menambahkan history untuk metrik yang benar-benar ada (tidak nil)
func (s *HealthDataService) buildReadingHistory(data []entity.HealthData) []response.ReadingHistoryResponse {
	var history []response.ReadingHistoryResponse

	// Sort by created_at DESC (terbaru ke terlama)
	sortedData := make([]entity.HealthData, len(data))
	copy(sortedData, data)
	sort.Slice(sortedData, func(i, j int) bool {
		return sortedData[i].CreatedAt.After(sortedData[j].CreatedAt)
	})

	for _, d := range sortedData {
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

		// Gula darah (hanya jika ada)
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

		// Berat badan (hanya jika ada)
		if d.Weight != nil {
			weight := *d.Weight
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

		// Detak jantung (hanya jika ada)
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

		// Aktivitas (jika ada dan tidak kosong)
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

// GenerateReportPDF menghasilkan laporan dalam format PDF dengan desain yang lebih baik
func (s *HealthDataService) GenerateReportPDF(userID uint, req *request.HealthHistoryRequest) (*bytes.Buffer, string, error) {
	// Ambil data riwayat kesehatan
	historyResp, err := s.GetHealthHistory(userID, req)
	if err != nil {
		return nil, "", err
	}

	// Tentukan rentang waktu untuk nama file
	startDate, endDate, _ := s.parseTimeRange(req)
	timeRangeStr := fmt.Sprintf("%s_to_%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	filename := fmt.Sprintf("riwayat_kesehatan_%s.pdf", timeRangeStr)

	// Buat PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 25, 20)
	
	// Helper function untuk truncate text agar tidak melebihi lebar
	truncateText := func(text string, maxWidth float64, fontSize float64) string {
		pdf.SetFont("Arial", "", fontSize)
		textWidth := pdf.GetStringWidth(text)
		if textWidth <= maxWidth {
			return text
		}
		// Truncate dengan ellipsis
		for len(text) > 0 {
			text = text[:len(text)-1]
			newWidth := pdf.GetStringWidth(text + "...")
			if newWidth <= maxWidth {
				return text + "..."
			}
		}
		return "..."
	}

	// Helper function untuk draw tabel formal dengan border
	drawFormalTable := func(headers []string, rows [][]string, colWidths []float64) {
		startX := pdf.GetX()
		startY := pdf.GetY()
		headerHeight := 10.0
		cellPadding := 3.0
		totalWidth := 0.0
		for _, w := range colWidths {
			totalWidth += w
		}
		
		// Header tabel
		pdf.SetFillColor(240, 240, 240) // Abu-abu sangat muda
		pdf.SetDrawColor(0, 0, 0)       // Hitam
		pdf.SetLineWidth(0.5)
		pdf.SetFont("Arial", "B", 10)
		pdf.SetTextColor(0, 0, 0)
		
		// Draw header background dan border
		pdf.Rect(startX, startY, totalWidth, headerHeight, "FD")
		
		// Draw vertical lines untuk header
		xPos := startX
		for i := 0; i <= len(colWidths); i++ {
			if i > 0 {
				pdf.Line(xPos, startY, xPos, startY+headerHeight)
			}
			if i < len(colWidths) {
				xPos += colWidths[i]
			}
		}
		
		// Header text
		xPos = startX + cellPadding
		for i, header := range headers {
			if i < len(colWidths) {
				cellWidth := colWidths[i] - (cellPadding * 2)
				truncatedHeader := truncateText(header, cellWidth, 10)
				pdf.SetXY(xPos, startY+3)
				pdf.Cell(cellWidth, 7, truncatedHeader)
				xPos += colWidths[i]
			}
		}
		
		// Data rows
		pdf.SetY(startY + headerHeight)
		pdf.SetFont("Arial", "", 9)
		pdf.SetFillColor(255, 255, 255)
		
		for i, row := range rows {
			rowY := pdf.GetY()
			
			// Hitung tinggi row yang dibutuhkan (untuk multi-line)
			maxLines := 1
			cellLines := make([][]string, len(row))
			for j, cell := range row {
				if j < len(colWidths) {
					cellWidth := colWidths[j] - (cellPadding * 2)
					lines := pdf.SplitText(cell, cellWidth)
					if len(lines) == 0 {
						lines = []string{""}
					}
					cellLines[j] = lines
					if len(lines) > maxLines {
						maxLines = len(lines)
					}
				}
			}
			lineHeight := 4.0
			actualRowHeight := (lineHeight * float64(maxLines)) + 4.0 // +4 untuk padding atas bawah
			
			// Zebra striping (baris genap abu-abu sangat muda)
			if i%2 == 0 {
				pdf.SetFillColor(250, 250, 250)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}
			pdf.Rect(startX, rowY, totalWidth, actualRowHeight, "FD")
			
			// Draw vertical lines untuk row
			xPos = startX
			for j := 0; j <= len(colWidths); j++ {
				if j > 0 {
					pdf.Line(xPos, rowY, xPos, rowY+actualRowHeight)
				}
				if j < len(colWidths) {
					xPos += colWidths[j]
				}
			}
			
			// Row data dengan MultiCell untuk wrap text
			xPos = startX + cellPadding
			for j := range row {
				if j < len(colWidths) && j < len(cellLines) {
					cellWidth := colWidths[j] - (cellPadding * 2)
					lines := cellLines[j]
					
					// Tulis setiap baris teks
					currentY := rowY + 2
					for lineIdx, line := range lines {
						if lineIdx > 0 {
							currentY += lineHeight
						}
						pdf.SetXY(xPos, currentY)
						pdf.Cell(cellWidth, lineHeight, line)
					}
					
					xPos += colWidths[j]
				}
			}
			
			// Set Y ke posisi terendah dari semua kolom
			pdf.SetY(rowY + actualRowHeight)
			
			// Cek jika perlu halaman baru
			if pdf.GetY() > 270 {
				pdf.AddPage()
				startY = pdf.GetY()
				startX = pdf.GetX()
				// Redraw header
				pdf.SetFillColor(240, 240, 240)
				pdf.Rect(startX, startY, totalWidth, headerHeight, "FD")
				
				// Draw vertical lines untuk header
				xPos = startX
				for i := 0; i <= len(colWidths); i++ {
					if i > 0 {
						pdf.Line(xPos, startY, xPos, startY+headerHeight)
					}
					if i < len(colWidths) {
						xPos += colWidths[i]
					}
				}
				
				xPos = startX + cellPadding
				pdf.SetFont("Arial", "B", 10)
				for i, header := range headers {
					if i < len(colWidths) {
						cellWidth := colWidths[i] - (cellPadding * 2)
						truncatedHeader := truncateText(header, cellWidth, 10)
						pdf.SetXY(xPos, startY+3)
						pdf.Cell(cellWidth, 7, truncatedHeader)
						xPos += colWidths[i]
					}
				}
				pdf.SetY(startY + headerHeight)
				pdf.SetFont("Arial", "", 9)
			}
		}
		
		pdf.SetFillColor(255, 255, 255) // Reset
	}

	// ========== HALAMAN COVER ==========
	pdf.AddPage()
	
	// Header formal dengan garis bawah
	pdf.SetXY(20, 30)
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(170, 10, "LAPORAN RIWAYAT KESEHATAN")
	
	// Garis bawah header
	pdf.SetLineWidth(1.0)
	pdf.SetDrawColor(0, 0, 0)
	pdf.Line(20, 42, 190, 42)
	pdf.Ln(15)
	
	// Info identitas laporan
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)
	
	// Periode laporan
	pdf.Cell(50, 7, "Periode Laporan:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(120, 7, fmt.Sprintf("%s s/d %s", startDate.Format("02 Januari 2006"), endDate.Format("02 Januari 2006")))
	pdf.Ln(8)
	
	// Tanggal dibuat
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(50, 7, "Tanggal Dibuat:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(120, 7, time.Now().Format("02 Januari 2006, 15:04:05 WIB"))
	pdf.Ln(20)

	// ========== RINGKASAN STATISTIK ==========
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(170, 10, "RINGKASAN STATISTIK")
	pdf.Ln(12)

	// Buat tabel ringkasan statistik
	var summaryRows [][]string
	
	// Tekanan Darah
	if historyResp.Summary.BloodPressure != nil {
		summaryRows = append(summaryRows, []string{
			"Tekanan Darah",
			fmt.Sprintf("Systolic: %.1f mmHg", historyResp.Summary.BloodPressure.AvgSystolic),
			fmt.Sprintf("Diastolic: %.1f mmHg", historyResp.Summary.BloodPressure.AvgDiastolic),
			fmt.Sprintf("%s / %s", historyResp.Summary.BloodPressure.SystolicStatus, historyResp.Summary.BloodPressure.DiastolicStatus),
		})
		summaryRows = append(summaryRows, []string{
			"",
			fmt.Sprintf("Rentang Normal: %s", historyResp.Summary.BloodPressure.NormalRange),
			fmt.Sprintf("Perubahan: %.1f%%", historyResp.Summary.BloodPressure.ChangePercent),
			"",
		})
		summaryRows = append(summaryRows, []string{"", "", "", ""}) // Spacer
	}

	// Gula Darah
	if historyResp.Summary.BloodSugar != nil {
		summaryRows = append(summaryRows, []string{
			"Gula Darah",
			fmt.Sprintf("Rata-rata: %.1f mg/dL", historyResp.Summary.BloodSugar.AvgValue),
			fmt.Sprintf("Status: %s", historyResp.Summary.BloodSugar.Status),
			fmt.Sprintf("Rentang Normal: %s", historyResp.Summary.BloodSugar.NormalRange),
		})
		summaryRows = append(summaryRows, []string{
			"",
			fmt.Sprintf("Perubahan: %.1f%%", historyResp.Summary.BloodSugar.ChangePercent),
			"",
			"",
		})
		summaryRows = append(summaryRows, []string{"", "", "", ""}) // Spacer
	}

	// Berat Badan
	if historyResp.Summary.Weight != nil {
		bmiText := ""
		if historyResp.Summary.Weight.BMI != nil {
			bmiText = fmt.Sprintf("BMI: %.1f", *historyResp.Summary.Weight.BMI)
		}
		summaryRows = append(summaryRows, []string{
			"Berat Badan",
			fmt.Sprintf("Rata-rata: %.1f kg", historyResp.Summary.Weight.AvgWeight),
			fmt.Sprintf("Tren: %s", historyResp.Summary.Weight.Trend),
			bmiText,
		})
		summaryRows = append(summaryRows, []string{
			"",
			fmt.Sprintf("Perubahan: %.1f%%", historyResp.Summary.Weight.ChangePercent),
			"",
			"",
		})
		summaryRows = append(summaryRows, []string{"", "", "", ""}) // Spacer
	}

	// Aktivitas
	if historyResp.Summary.Activity != nil {
		summaryRows = append(summaryRows, []string{
			"Aktivitas",
			fmt.Sprintf("Total Langkah: %s", formatNumber(historyResp.Summary.Activity.TotalSteps)),
			fmt.Sprintf("Total Kalori: %.0f kkal", historyResp.Summary.Activity.TotalCalories),
			fmt.Sprintf("Perubahan: %.1f%%", historyResp.Summary.Activity.ChangePercent),
		})
	}

	// Draw tabel ringkasan
	if len(summaryRows) > 0 {
		headers := []string{"Parameter", "Nilai", "Status/Tren", "Keterangan"}
		colWidths := []float64{45, 55, 50, 20} // Lebar kolom disesuaikan agar teks tidak bocor
		drawFormalTable(headers, summaryRows, colWidths)
		pdf.Ln(10)
	}

	// ========== CATATAN PEMBACAAN ==========
	if len(historyResp.ReadingHistory) > 0 {
		// Cek jika perlu halaman baru
		if pdf.GetY() > 250 {
			pdf.AddPage()
		} else {
			pdf.Ln(15)
		}

		pdf.SetFont("Arial", "B", 14)
		pdf.SetTextColor(0, 0, 0)
		pdf.Cell(170, 10, "CATATAN PEMBACAAN")
		pdf.Ln(12)

		// Siapkan data untuk tabel
		var readingRows [][]string
		for _, record := range historyResp.ReadingHistory {
			dateTime := record.DateTime.Format("02/01/2006 15:04")
			metricType := record.MetricType
			value := record.Value
			status := record.Status
			context := ""
			if record.Context != nil {
				context = *record.Context
			}
			
			notes := ""
			if record.Notes != nil && *record.Notes != "" {
				notes = *record.Notes
			}
			
			readingRows = append(readingRows, []string{
				dateTime,
				metricType,
				value,
				status,
				context,
				notes,
			})
		}

		// Draw tabel formal
		headers := []string{"Tanggal & Waktu", "Jenis Metrik", "Nilai", "Status", "Konteks", "Catatan"}
		colWidths := []float64{38, 32, 25, 28, 22, 25} // Lebar kolom disesuaikan agar teks tidak bocor
		drawFormalTable(headers, readingRows, colWidths)
	}

	// Footer di setiap halaman
	pdf.AliasNbPages("")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(0, 10, fmt.Sprintf("Halaman %d dari {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	// Output PDF ke buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, "", fmt.Errorf("gagal membuat PDF: %w", err)
	}

	return &buf, filename, nil
}

// formatNumber memformat angka dengan separator ribuan
func formatNumber(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}
	
	var result strings.Builder
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(".")
		}
		result.WriteRune(char)
	}
	return result.String()
}

