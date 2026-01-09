package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"math"
	"sort"
	"strconv"
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
	// endDate adalah hari ini (untuk filter berdasarkan record_date, kita hanya perlu tanggal)
	// Gunakan UTC untuk konsistensi timezone
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	// days-1 karena hari ini sudah termasuk, jadi kita perlu mundur (days-1) hari
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(days - 1))

	var filtered []entity.HealthData
	for _, d := range data {
		// Normalisasi record_date ke awal hari untuk perbandingan
		recordDate := time.Date(d.RecordDate.Year(), d.RecordDate.Month(), d.RecordDate.Day(), 0, 0, 0, 0, d.RecordDate.Location())

		// Filter: recordDate harus >= startDate dan <= endDate
		// Gunakan perbandingan yang lebih inklusif untuk memastikan data hari ini masuk
		if !recordDate.Before(startDate) && !recordDate.After(endDate) {
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

	// Filter dan build untuk 7Days (per hari)
	data7Days := s.filterDataByTimeRange(data, 7)
	charts.Days7 = s.buildBloodPressureTrendPoints(data7Days)

	// Filter dan build untuk 1Month (30 hari - per minggu)
	data1Month := s.filterDataByTimeRange(data, 30)
	charts.Month1 = s.buildBloodPressureTrendPointsWeek(data1Month)

	// Filter dan build untuk 3Months (90 hari - per bulan)
	data3Months := s.filterDataByTimeRange(data, 90)
	charts.Months3 = s.buildBloodPressureTrendPointsMonth(data3Months)

	return charts
}

// buildBloodPressureTrendPoints membangun array titik data tren tekanan darah per hari (7Days)
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

// buildBloodPressureTrendPointsWeek membangun array titik data tren tekanan darah per minggu (1Month)
func (s *HealthDataService) buildBloodPressureTrendPointsWeek(data []entity.HealthData) []response.BloodPressureTrendPointWeek {
	// Filter hanya data yang memiliki tekanan darah
	var bpData []entity.HealthData
	for _, d := range data {
		if d.Systolic != nil && d.Diastolic != nil {
			bpData = append(bpData, d)
		}
	}

	if len(bpData) == 0 {
		return []response.BloodPressureTrendPointWeek{}
	}

	// Cari tanggal terawal untuk menghitung week number
	// Week 1 adalah minggu pertama dalam range 30 hari terakhir
	now := time.Now()
	rangeStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -29)

	// Group by week (minggu dimulai dari Senin)
	weekMap := make(map[string][]entity.HealthData)
	for _, d := range bpData {
		weekKey, _, _ := s.getWeekRange(d.RecordDate, rangeStartDate)
		weekMap[weekKey] = append(weekMap[weekKey], d)
	}

	// Hitung rata-rata per minggu
	var points []response.BloodPressureTrendPointWeek
	for weekKey, weekData := range weekMap {
		if len(weekData) == 0 {
			continue
		}

		// Ambil start_date dan end_date dari data pertama di minggu tersebut
		_, startDate, endDate := s.getWeekRange(weekData[0].RecordDate, rangeStartDate)

		// Hitung rata-rata systolic dan diastolic
		var totalSystolic, totalDiastolic float64
		count := 0
		for _, d := range weekData {
			if d.Systolic != nil && d.Diastolic != nil {
				totalSystolic += float64(*d.Systolic)
				totalDiastolic += float64(*d.Diastolic)
				count++
			}
		}

		if count > 0 {
			points = append(points, response.BloodPressureTrendPointWeek{
				Week:      weekKey,
				StartDate: startDate,
				EndDate:   endDate,
				Systolic:  math.Round(totalSystolic / float64(count)),
				Diastolic: math.Round(totalDiastolic / float64(count)),
			})
		}
	}

	// Sort by start_date (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].StartDate < points[j].StartDate
	})

	return points
}

// buildBloodPressureTrendPointsMonth membangun array titik data tren tekanan darah per bulan (3Months)
func (s *HealthDataService) buildBloodPressureTrendPointsMonth(data []entity.HealthData) []response.BloodPressureTrendPointMonth {
	// Filter hanya data yang memiliki tekanan darah
	var bpData []entity.HealthData
	for _, d := range data {
		if d.Systolic != nil && d.Diastolic != nil {
			bpData = append(bpData, d)
		}
	}

	if len(bpData) == 0 {
		return []response.BloodPressureTrendPointMonth{}
	}

	// Group by month
	monthMap := make(map[string][]entity.HealthData)
	for _, d := range bpData {
		monthKey := s.getMonthKey(d.RecordDate)
		monthMap[monthKey] = append(monthMap[monthKey], d)
	}

	// Hitung rata-rata per bulan
	var points []response.BloodPressureTrendPointMonth
	for monthKey, monthData := range monthMap {
		if len(monthData) == 0 {
			continue
		}

		// Hitung rata-rata systolic dan diastolic
		var totalSystolic, totalDiastolic float64
		count := 0
		for _, d := range monthData {
			if d.Systolic != nil && d.Diastolic != nil {
				totalSystolic += float64(*d.Systolic)
				totalDiastolic += float64(*d.Diastolic)
				count++
			}
		}

		if count > 0 {
			points = append(points, response.BloodPressureTrendPointMonth{
				Month:     monthKey,
				Systolic:  math.Round(totalSystolic / float64(count)),
				Diastolic: math.Round(totalDiastolic / float64(count)),
			})
		}
	}

	// Sort by month (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Month < points[j].Month
	})

	return points
}

// buildBloodSugarTrend membangun data tren gula darah dengan filter waktu
func (s *HealthDataService) buildBloodSugarTrend(data []entity.HealthData) response.BloodSugarTrendCharts {
	charts := response.BloodSugarTrendCharts{}

	// Filter dan build untuk 7Days (per hari)
	data7Days := s.filterDataByTimeRange(data, 7)
	charts.Days7 = s.buildBloodSugarTrendPoints(data7Days)

	// Filter dan build untuk 1Month (30 hari - per minggu)
	data1Month := s.filterDataByTimeRange(data, 30)
	charts.Month1 = s.buildBloodSugarTrendPointsWeek(data1Month)

	// Filter dan build untuk 3Months (90 hari - per bulan)
	data3Months := s.filterDataByTimeRange(data, 90)
	charts.Months3 = s.buildBloodSugarTrendPointsMonth(data3Months)

	return charts
}

// buildBloodSugarTrendPoints membangun array titik data tren gula darah per hari (7Days)
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

// buildBloodSugarTrendPointsWeek membangun array titik data tren gula darah per minggu (1Month)
func (s *HealthDataService) buildBloodSugarTrendPointsWeek(data []entity.HealthData) []response.BloodSugarTrendPointWeek {
	// Filter hanya data yang memiliki gula darah
	var bsData []entity.HealthData
	for _, d := range data {
		if d.BloodSugar != nil {
			bsData = append(bsData, d)
		}
	}

	if len(bsData) == 0 {
		return []response.BloodSugarTrendPointWeek{}
	}

	// Cari tanggal terawal untuk menghitung week number
	now := time.Now()
	rangeStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -29)

	// Group by week
	weekMap := make(map[string][]entity.HealthData)
	for _, d := range bsData {
		weekKey, _, _ := s.getWeekRange(d.RecordDate, rangeStartDate)
		weekMap[weekKey] = append(weekMap[weekKey], d)
	}

	// Hitung rata-rata per minggu
	var points []response.BloodSugarTrendPointWeek
	for weekKey, weekData := range weekMap {
		if len(weekData) == 0 {
			continue
		}

		// Ambil start_date dan end_date dari data pertama di minggu tersebut
		_, startDate, endDate := s.getWeekRange(weekData[0].RecordDate, rangeStartDate)

		// Hitung rata-rata
		var totalValue float64
		count := 0
		for _, d := range weekData {
			if d.BloodSugar != nil {
				totalValue += float64(*d.BloodSugar)
				count++
			}
		}

		if count > 0 {
			points = append(points, response.BloodSugarTrendPointWeek{
				Week:      weekKey,
				StartDate: startDate,
				EndDate:   endDate,
				AvgValue:  math.Round(totalValue / float64(count)),
			})
		}
	}

	// Sort by start_date (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].StartDate < points[j].StartDate
	})

	return points
}

// buildBloodSugarTrendPointsMonth membangun array titik data tren gula darah per bulan (3Months)
func (s *HealthDataService) buildBloodSugarTrendPointsMonth(data []entity.HealthData) []response.BloodSugarTrendPointMonth {
	// Filter hanya data yang memiliki gula darah
	var bsData []entity.HealthData
	for _, d := range data {
		if d.BloodSugar != nil {
			bsData = append(bsData, d)
		}
	}

	if len(bsData) == 0 {
		return []response.BloodSugarTrendPointMonth{}
	}

	// Group by month
	monthMap := make(map[string][]entity.HealthData)
	for _, d := range bsData {
		monthKey := s.getMonthKey(d.RecordDate)
		monthMap[monthKey] = append(monthMap[monthKey], d)
	}

	// Hitung rata-rata per bulan
	var points []response.BloodSugarTrendPointMonth
	for monthKey, monthData := range monthMap {
		if len(monthData) == 0 {
			continue
		}

		// Hitung rata-rata
		var totalValue float64
		count := 0
		for _, d := range monthData {
			if d.BloodSugar != nil {
				totalValue += float64(*d.BloodSugar)
				count++
			}
		}

		if count > 0 {
			points = append(points, response.BloodSugarTrendPointMonth{
				Month:    monthKey,
				AvgValue: math.Round(totalValue / float64(count)),
			})
		}
	}

	// Sort by month (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Month < points[j].Month
	})

	return points
}

// buildWeightTrend membangun data tren berat badan dengan filter waktu
func (s *HealthDataService) buildWeightTrend(data []entity.HealthData) response.WeightTrendCharts {
	charts := response.WeightTrendCharts{}

	// Filter dan build untuk 7Days (per hari)
	data7Days := s.filterDataByTimeRange(data, 7)
	charts.Days7 = s.buildWeightTrendPoints(data7Days)

	// Filter dan build untuk 1Month (30 hari - per minggu)
	data1Month := s.filterDataByTimeRange(data, 30)
	charts.Month1 = s.buildWeightTrendPointsWeek(data1Month)

	// Filter dan build untuk 3Months (90 hari - per bulan)
	data3Months := s.filterDataByTimeRange(data, 90)
	charts.Months3 = s.buildWeightTrendPointsMonth(data3Months)

	return charts
}

// buildWeightTrendPoints membangun array titik data tren berat badan per hari (7Days)
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
				Weight: math.Round(*latestData.Weight),
			})
		}
	}

	// Sort by date (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points
}

// buildWeightTrendPointsWeek membangun array titik data tren berat badan per minggu (1Month)
func (s *HealthDataService) buildWeightTrendPointsWeek(data []entity.HealthData) []response.WeightTrendPointWeek {
	// Filter hanya data yang memiliki berat badan
	var weightData []entity.HealthData
	for _, d := range data {
		if d.Weight != nil {
			weightData = append(weightData, d)
		}
	}

	if len(weightData) == 0 {
		return []response.WeightTrendPointWeek{}
	}

	// Cari tanggal terawal untuk menghitung week number
	now := time.Now()
	rangeStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -29)

	// Group by week
	weekMap := make(map[string][]entity.HealthData)
	for _, d := range weightData {
		weekKey, _, _ := s.getWeekRange(d.RecordDate, rangeStartDate)
		weekMap[weekKey] = append(weekMap[weekKey], d)
	}

	// Hitung rata-rata per minggu
	var points []response.WeightTrendPointWeek
	for weekKey, weekData := range weekMap {
		if len(weekData) == 0 {
			continue
		}

		// Ambil start_date dan end_date dari data pertama di minggu tersebut
		_, startDate, endDate := s.getWeekRange(weekData[0].RecordDate, rangeStartDate)

		// Hitung rata-rata
		var totalWeight float64
		count := 0
		for _, d := range weekData {
			if d.Weight != nil {
				totalWeight += *d.Weight
				count++
			}
		}

		if count > 0 {
			points = append(points, response.WeightTrendPointWeek{
				Week:      weekKey,
				StartDate: startDate,
				EndDate:   endDate,
				Weight:    math.Round(totalWeight / float64(count)),
			})
		}
	}

	// Sort by start_date (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].StartDate < points[j].StartDate
	})

	return points
}

// buildWeightTrendPointsMonth membangun array titik data tren berat badan per bulan (3Months)
func (s *HealthDataService) buildWeightTrendPointsMonth(data []entity.HealthData) []response.WeightTrendPointMonth {
	// Filter hanya data yang memiliki berat badan
	var weightData []entity.HealthData
	for _, d := range data {
		if d.Weight != nil {
			weightData = append(weightData, d)
		}
	}

	if len(weightData) == 0 {
		return []response.WeightTrendPointMonth{}
	}

	// Group by month
	monthMap := make(map[string][]entity.HealthData)
	for _, d := range weightData {
		monthKey := s.getMonthKey(d.RecordDate)
		monthMap[monthKey] = append(monthMap[monthKey], d)
	}

	// Hitung rata-rata per bulan
	var points []response.WeightTrendPointMonth
	for monthKey, monthData := range monthMap {
		if len(monthData) == 0 {
			continue
		}

		// Hitung rata-rata
		var totalWeight float64
		count := 0
		for _, d := range monthData {
			if d.Weight != nil {
				totalWeight += *d.Weight
				count++
			}
		}

		if count > 0 {
			points = append(points, response.WeightTrendPointMonth{
				Month:  monthKey,
				Weight: math.Round(totalWeight / float64(count)),
			})
		}
	}

	// Sort by month (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Month < points[j].Month
	})

	return points
}

// buildActivityTrend membangun data tren aktivitas dengan filter waktu
func (s *HealthDataService) buildActivityTrend(data []entity.HealthData) response.ActivityTrendCharts {
	charts := response.ActivityTrendCharts{}

	// Filter dan build untuk 7Days (per hari)
	data7Days := s.filterDataByTimeRange(data, 7)
	charts.Days7 = s.buildActivityTrendPoints(data7Days)

	// Filter dan build untuk 1Month (30 hari - per minggu)
	data1Month := s.filterDataByTimeRange(data, 30)
	charts.Month1 = s.buildActivityTrendPointsWeek(data1Month)

	// Filter dan build untuk 3Months (90 hari - per bulan)
	data3Months := s.filterDataByTimeRange(data, 90)
	charts.Months3 = s.buildActivityTrendPointsMonth(data3Months)

	return charts
}

// buildActivityTrendPoints membangun array titik data tren aktivitas per hari (7Days)
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
				Calories: math.Round(calories),
			})
		}
	}

	// Sort by date (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points
}

// buildActivityTrendPointsWeek membangun array titik data tren aktivitas per minggu (1Month)
func (s *HealthDataService) buildActivityTrendPointsWeek(data []entity.HealthData) []response.ActivityTrendPointWeek {
	if len(data) == 0 {
		return []response.ActivityTrendPointWeek{}
	}

	// Cari tanggal terawal untuk menghitung week number
	now := time.Now()
	rangeStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -29)

	// Group by week
	weekMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		weekKey, _, _ := s.getWeekRange(d.RecordDate, rangeStartDate)
		weekMap[weekKey] = append(weekMap[weekKey], d)
	}

	// Hitung total per minggu
	var points []response.ActivityTrendPointWeek
	for weekKey, weekData := range weekMap {
		if len(weekData) == 0 {
			continue
		}

		// Ambil start_date dan end_date dari data pertama di minggu tersebut
		_, startDate, endDate := s.getWeekRange(weekData[0].RecordDate, rangeStartDate)

		// Hitung total steps dan calories
		totalSteps := 0
		totalCalories := 0.0
		for _, d := range weekData {
			if d.Activity != nil && *d.Activity != "" {
				totalSteps += 1000
				totalCalories += 200.0
			}
		}

		points = append(points, response.ActivityTrendPointWeek{
			Week:      weekKey,
			StartDate: startDate,
			EndDate:   endDate,
			Steps:     totalSteps,
			Calories:  math.Round(totalCalories),
		})
	}

	// Sort by start_date (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].StartDate < points[j].StartDate
	})

	return points
}

// buildActivityTrendPointsMonth membangun array titik data tren aktivitas per bulan (3Months)
func (s *HealthDataService) buildActivityTrendPointsMonth(data []entity.HealthData) []response.ActivityTrendPointMonth {
	if len(data) == 0 {
		return []response.ActivityTrendPointMonth{}
	}

	// Group by month
	monthMap := make(map[string][]entity.HealthData)
	for _, d := range data {
		monthKey := s.getMonthKey(d.RecordDate)
		monthMap[monthKey] = append(monthMap[monthKey], d)
	}

	// Hitung total per bulan
	var points []response.ActivityTrendPointMonth
	for monthKey, monthData := range monthMap {
		if len(monthData) == 0 {
			continue
		}

		// Hitung total steps dan calories
		totalSteps := 0
		totalCalories := 0.0
		for _, d := range monthData {
			if d.Activity != nil && *d.Activity != "" {
				totalSteps += 1000
				totalCalories += 200.0
			}
		}

		points = append(points, response.ActivityTrendPointMonth{
			Month:    monthKey,
			Steps:    totalSteps,
			Calories: math.Round(totalCalories),
		})
	}

	// Sort by month (terlama ke terbaru)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Month < points[j].Month
	})

	return points
}

// getWeekRange menghitung range minggu untuk tanggal tertentu
// Minggu dimulai dari Senin (ISO 8601 week)
// Mengembalikan: weekKey (format: "Week 1"), startDate, endDate
// rangeStartDate adalah tanggal awal range (30 hari yang lalu)
// Week 1 adalah minggu pertama dalam range (paling lama), Week 2 adalah minggu kedua, dst
// Maksimal Week 4 karena range hanya 30 hari
func (s *HealthDataService) getWeekRange(date time.Time, rangeStartDate time.Time) (weekKey, startDate, endDate string) {
	// Cari hari Senin dari minggu tersebut
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // Minggu = 7
	}
	daysFromMonday := weekday - 1 // 0 = Senin, 6 = Minggu

	monday := date.AddDate(0, 0, -daysFromMonday)
	sunday := monday.AddDate(0, 0, 6)

	startDate = monday.Format("2006-01-02")
	endDate = sunday.Format("2006-01-02")

	// Hitung week number berdasarkan posisi dalam range 30 hari
	// Cari hari Senin dari rangeStartDate untuk konsistensi
	rangeWeekday := int(rangeStartDate.Weekday())
	if rangeWeekday == 0 {
		rangeWeekday = 7
	}
	rangeDaysFromMonday := rangeWeekday - 1
	rangeMonday := rangeStartDate.AddDate(0, 0, -rangeDaysFromMonday)

	// Hitung selisih hari antara Senin dari tanggal ini dengan Senin dari rangeStartDate
	daysDiff := int(monday.Sub(rangeMonday).Hours() / 24)
	// Week 1 adalah minggu pertama (rangeMonday), Week 2 adalah minggu kedua, dst
	weekNum := (daysDiff / 7) + 1
	if weekNum < 1 {
		weekNum = 1
	}
	// Batasi maksimal Week 4 karena range hanya 30 hari (sekitar 4 minggu)
	// Ini mencegah munculnya Week 5, Week 6, dst yang tidak masuk akal untuk range 30 hari
	if weekNum > 4 {
		weekNum = 4
	}

	weekKey = "Week " + strconv.Itoa(weekNum)
	return
}

// getMonthKey menghasilkan key bulan dalam format "Dec 2025", "Jan 2026", dll
func (s *HealthDataService) getMonthKey(date time.Time) string {
	monthNames := []string{
		"Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
	}
	month := monthNames[date.Month()-1]
	year := date.Year()
	return month + " " + strconv.Itoa(year)
}
