package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"sort"
)

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

