package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"sort"
	"time"
)

// calculateTrendCharts menghitung data untuk grafik tren dengan filter waktu
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

// filterDataByTimeRange memfilter data berdasarkan range waktu (7Days, 1Month, 3Months)
// Menggunakan record_date sebagai acuan
// days: jumlah hari total (contoh: 7 untuk 7Days = hari ini + 6 hari sebelumnya)
func (s *HealthDataService) filterDataByTimeRange(data []entity.HealthData, days int) []entity.HealthData {
	now := time.Now()
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	// days-1 karena hari ini sudah termasuk, jadi kita perlu mundur (days-1) hari
	startDate := endDate.AddDate(0, 0, -(days - 1))

	var filtered []entity.HealthData
	for _, d := range data {
		// Normalisasi record_date ke awal hari untuk perbandingan
		recordDate := time.Date(d.RecordDate.Year(), d.RecordDate.Month(), d.RecordDate.Day(), 0, 0, 0, 0, d.RecordDate.Location())
		startDateNormalized := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		endDateNormalized := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location())

		if (recordDate.Equal(startDateNormalized) || recordDate.After(startDateNormalized)) &&
			(recordDate.Equal(endDateNormalized) || recordDate.Before(endDateNormalized)) {
			filtered = append(filtered, d)
		}
	}

	return filtered
}

// getLatestDataPerDay mengambil 1 data terakhir per hari dari data yang sudah di-group
// Jika dalam 1 hari ada banyak data, gunakan yang terakhir (berdasarkan created_at)
func (s *HealthDataService) getLatestDataPerDay(dayData []entity.HealthData) *entity.HealthData {
	if len(dayData) == 0 {
		return nil
	}

	// Sort by created_at DESC untuk mendapatkan yang terakhir
	sorted := make([]entity.HealthData, len(dayData))
	copy(sorted, dayData)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	return &sorted[0]
}

// buildBloodPressureTrend membangun data tren tekanan darah dengan filter waktu
func (s *HealthDataService) buildBloodPressureTrend(data []entity.HealthData) response.BloodPressureTrendCharts {
	charts := response.BloodPressureTrendCharts{}

	// Filter dan build untuk 7Days
	data7Days := s.filterDataByTimeRange(data, 7)
	charts.Days7 = s.buildBloodPressureTrendPoints(data7Days)

	// Filter dan build untuk 1Month (30 hari)
	data1Month := s.filterDataByTimeRange(data, 30)
	charts.Month1 = s.buildBloodPressureTrendPoints(data1Month)

	// Filter dan build untuk 3Months (90 hari)
	data3Months := s.filterDataByTimeRange(data, 90)
	charts.Months3 = s.buildBloodPressureTrendPoints(data3Months)

	return charts
}

// buildBloodPressureTrendPoints membangun array titik data tren tekanan darah per hari
func (s *HealthDataService) buildBloodPressureTrendPoints(data []entity.HealthData) []response.BloodPressureTrendPoint {
	// Group by record_date (hanya data yang memiliki tekanan darah)
	dateMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		if d.Systolic != nil && d.Diastolic != nil {
			dateStr := d.RecordDate.Format("2006-01-02")
			dateMap[dateStr] = append(dateMap[dateStr], d)
		}
	}

	// Ambil 1 data terakhir per hari dan buat titik data
	var points []response.BloodPressureTrendPoint
	for dateStr, dayData := range dateMap {
		latestData := s.getLatestDataPerDay(dayData)
		if latestData != nil && latestData.Systolic != nil && latestData.Diastolic != nil {
			points = append(points, response.BloodPressureTrendPoint{
				Date:      dateStr,
				Systolic:  float64(*latestData.Systolic),
				Diastolic: float64(*latestData.Diastolic),
			})
		}
	}

	// Sort by date (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points
}

// buildBloodSugarTrend membangun data tren gula darah dengan filter waktu
func (s *HealthDataService) buildBloodSugarTrend(data []entity.HealthData) response.BloodSugarTrendCharts {
	charts := response.BloodSugarTrendCharts{}

	// Filter dan build untuk 7Days
	data7Days := s.filterDataByTimeRange(data, 7)
	charts.Days7 = s.buildBloodSugarTrendPoints(data7Days)

	// Filter dan build untuk 1Month (30 hari)
	data1Month := s.filterDataByTimeRange(data, 30)
	charts.Month1 = s.buildBloodSugarTrendPoints(data1Month)

	// Filter dan build untuk 3Months (90 hari)
	data3Months := s.filterDataByTimeRange(data, 90)
	charts.Months3 = s.buildBloodSugarTrendPoints(data3Months)

	return charts
}

// buildBloodSugarTrendPoints membangun array titik data tren gula darah per hari
func (s *HealthDataService) buildBloodSugarTrendPoints(data []entity.HealthData) []response.BloodSugarTrendPoint {
	// Group by record_date (hanya data yang memiliki gula darah)
	dateMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		if d.BloodSugar != nil {
			dateStr := d.RecordDate.Format("2006-01-02")
			dateMap[dateStr] = append(dateMap[dateStr], d)
		}
	}

	// Ambil 1 data terakhir per hari dan buat titik data
	var points []response.BloodSugarTrendPoint
	for dateStr, dayData := range dateMap {
		latestData := s.getLatestDataPerDay(dayData)
		if latestData != nil && latestData.BloodSugar != nil {
			points = append(points, response.BloodSugarTrendPoint{
				Date:     dateStr,
				AvgValue: float64(*latestData.BloodSugar),
			})
		}
	}

	// Sort by date (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points
}

// buildWeightTrend membangun data tren berat badan dengan filter waktu
func (s *HealthDataService) buildWeightTrend(data []entity.HealthData) response.WeightTrendCharts {
	charts := response.WeightTrendCharts{}

	// Filter dan build untuk 7Days
	data7Days := s.filterDataByTimeRange(data, 7)
	charts.Days7 = s.buildWeightTrendPoints(data7Days)

	// Filter dan build untuk 1Month (30 hari)
	data1Month := s.filterDataByTimeRange(data, 30)
	charts.Month1 = s.buildWeightTrendPoints(data1Month)

	// Filter dan build untuk 3Months (90 hari)
	data3Months := s.filterDataByTimeRange(data, 90)
	charts.Months3 = s.buildWeightTrendPoints(data3Months)

	return charts
}

// buildWeightTrendPoints membangun array titik data tren berat badan per hari
func (s *HealthDataService) buildWeightTrendPoints(data []entity.HealthData) []response.WeightTrendPoint {
	// Group by record_date (hanya data yang memiliki berat badan)
	dateMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		if d.Weight != nil {
			dateStr := d.RecordDate.Format("2006-01-02")
			dateMap[dateStr] = append(dateMap[dateStr], d)
		}
	}

	// Ambil 1 data terakhir per hari dan buat titik data
	var points []response.WeightTrendPoint
	for dateStr, dayData := range dateMap {
		latestData := s.getLatestDataPerDay(dayData)
		if latestData != nil && latestData.Weight != nil {
			points = append(points, response.WeightTrendPoint{
				Date:   dateStr,
				Weight: *latestData.Weight,
			})
		}
	}

	// Sort by date (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points
}

// buildActivityTrend membangun data tren aktivitas dengan filter waktu
func (s *HealthDataService) buildActivityTrend(data []entity.HealthData) response.ActivityTrendCharts {
	charts := response.ActivityTrendCharts{}

	// Filter dan build untuk 7Days
	data7Days := s.filterDataByTimeRange(data, 7)
	charts.Days7 = s.buildActivityTrendPoints(data7Days)

	// Filter dan build untuk 1Month (30 hari)
	data1Month := s.filterDataByTimeRange(data, 30)
	charts.Month1 = s.buildActivityTrendPoints(data1Month)

	// Filter dan build untuk 3Months (90 hari)
	data3Months := s.filterDataByTimeRange(data, 90)
	charts.Months3 = s.buildActivityTrendPoints(data3Months)

	return charts
}

// buildActivityTrendPoints membangun array titik data tren aktivitas per hari
func (s *HealthDataService) buildActivityTrendPoints(data []entity.HealthData) []response.ActivityTrendPoint {
	// Group by record_date
	dateMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		dateStr := d.RecordDate.Format("2006-01-02")
		dateMap[dateStr] = append(dateMap[dateStr], d)
	}

	// Ambil 1 data terakhir per hari dan buat titik data
	var points []response.ActivityTrendPoint
	for dateStr, dayData := range dateMap {
		latestData := s.getLatestDataPerDay(dayData)
		if latestData != nil {
			steps := 0
			calories := 0.0
			// Hanya hitung jika ada activity field yang tidak kosong
			if latestData.Activity != nil && *latestData.Activity != "" {
				steps = 1000
				calories = 200.0
			}
			points = append(points, response.ActivityTrendPoint{
				Date:     dateStr,
				Steps:    steps,
				Calories: calories,
			})
		}
	}

	// Sort by date (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points
}
