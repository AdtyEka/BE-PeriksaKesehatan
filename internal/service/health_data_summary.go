package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"sort"
	"time"

	timezoneUtils "BE-PeriksaKesehatan/pkg/utils"
)

// calculateSummary menghitung ringkasan statistik berbasis perubahan harian
// Tidak lagi membandingkan dengan periode sebelumnya, tetapi:
// - Mengagregasi nilai harian (1 nilai representatif per hari)
// - Menghitung change_percent periode sebagai:
//   ((nilai_terakhir - nilai_pertama) / nilai_pertama) * 100
//   dengan aturan edge-case yang sudah ditentukan.
func (s *HealthDataService) calculateSummary(data, _ []entity.HealthData, metrics []string) response.HealthSummaryResponse {
	summary := response.HealthSummaryResponse{}

	// Cek apakah metrik diminta atau tidak ada filter
	includeAll := len(metrics) == 0
	includeBP := includeAll || s.containsMetric(metrics, "tekanan_darah")
	includeBS := includeAll || s.containsMetric(metrics, "gula_darah")
	includeWeight := includeAll || s.containsMetric(metrics, "berat_badan")
	includeActivity := includeAll || s.containsMetric(metrics, "aktivitas")

	if includeBP && len(data) > 0 {
		summary.BloodPressure = s.calculateBloodPressureSummary(data)
	}

	if includeBS && len(data) > 0 {
		summary.BloodSugar = s.calculateBloodSugarSummary(data)
	}

	if includeWeight && len(data) > 0 {
		summary.Weight = s.calculateWeightSummary(data)
	}

	if includeActivity && len(data) > 0 {
		summary.Activity = s.calculateActivitySummary(data)
	}

	return summary
}

// calculatePeriodChangePercent menghitung change_percent berbasis nilai harian:
// ((nilai_terakhir - nilai_pertama) / nilai_pertama) * 100
// Aturan:
// - Jika kurang dari 2 hari data  -> 0
// - Jika nilai_pertama == 0      -> 0
func calculatePeriodChangePercent(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	first := values[0]
	last := values[len(values)-1]

	if first == 0 {
		return 0
	}

	return ((last - first) / first) * 100
}

// calculateBloodPressureSummary menghitung ringkasan tekanan darah dengan nullable-aware,
// berbasis agregasi harian (1 nilai per hari).
func (s *HealthDataService) calculateBloodPressureSummary(data []entity.HealthData) *response.BloodPressureSummary {
	if len(data) == 0 {
		return nil
	}

	// Kelompokkan per hari berdasarkan RecordDate, hanya data dengan tekanan darah lengkap
	type agg struct {
		sumSys float64
		sumDia float64
		count  int
	}

	dailyMap := make(map[time.Time]*agg)
	var dates []time.Time

	for _, d := range data {
		if d.Systolic != nil && d.Diastolic != nil {
			recordDateJakarta := timezoneUtils.ToJakarta(d.RecordDate)
			day := timezoneUtils.DateInJakarta(recordDateJakarta.Year(), recordDateJakarta.Month(), recordDateJakarta.Day(), 0, 0, 0, 0)
			if _, ok := dailyMap[day]; !ok {
				dailyMap[day] = &agg{}
				dates = append(dates, day)
			}
			a := dailyMap[day]
			a.sumSys += float64(*d.Systolic)
			a.sumDia += float64(*d.Diastolic)
			a.count++
		}
	}

	if len(dailyMap) == 0 {
		return nil // Tidak ada data tekanan darah yang valid
	}

	// Urutkan tanggal ASC
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	var dailySys []float64
	var dailyDia []float64
	var totalSys, totalDia float64
	var dayCount int

	for _, day := range dates {
		a := dailyMap[day]
		if a.count == 0 {
			continue
		}
		avgSys := a.sumSys / float64(a.count)
		avgDia := a.sumDia / float64(a.count)

		dailySys = append(dailySys, avgSys)
		dailyDia = append(dailyDia, avgDia)

		totalSys += avgSys
		totalDia += avgDia
		dayCount++
	}

	if dayCount == 0 {
		return nil
	}

	// Rata-rata keseluruhan selama periode (berbasis nilai harian)
	avgSystolic := totalSys / float64(dayCount)
	avgDiastolic := totalDia / float64(dayCount)

	// Hitung persentase perubahan periode dari nilai harian
	// Di sini kita gunakan perubahan berdasarkan systolic sebagai representatif
	changePercent := calculatePeriodChangePercent(dailySys)

	// Hitung status berdasarkan rata-rata (menggunakan kombinasi sistolik dan diastolik)
	avgSystolicInt := int(avgSystolic)
	avgDiastolicInt := int(avgDiastolic)
	status := s.getBloodPressureStatus(avgSystolicInt, avgDiastolicInt)

	return &response.BloodPressureSummary{
		AvgSystolic:    roundTo2Decimals(avgSystolic),
		AvgDiastolic:   roundTo2Decimals(avgDiastolic),
		ChangePercent:  roundTo2Decimals(changePercent),
		SystolicStatus: status,  // Status berdasarkan kombinasi sistolik dan diastolik
		DiastolicStatus: status, // Status sama karena menggunakan kombinasi
		NormalRange:    "90-139 / 60-89 mmHg (WHO)",
	}
}

// calculateBloodSugarSummary menghitung ringkasan gula darah dengan nullable-aware
// Berbasis agregasi harian (1 nilai rata-rata per hari).
func (s *HealthDataService) calculateBloodSugarSummary(data []entity.HealthData) *response.BloodSugarSummary {
	if len(data) == 0 {
		return nil
	}

	// Kelompokkan per hari berdasarkan RecordDate, hanya data dengan gula darah
	type agg struct {
		sum   float64
		count int
	}

	dailyMap := make(map[time.Time]*agg)
	var dates []time.Time

	for _, d := range data {
		if d.BloodSugar != nil {
			recordDateJakarta := timezoneUtils.ToJakarta(d.RecordDate)
			day := timezoneUtils.DateInJakarta(recordDateJakarta.Year(), recordDateJakarta.Month(), recordDateJakarta.Day(), 0, 0, 0, 0)
			if _, ok := dailyMap[day]; !ok {
				dailyMap[day] = &agg{}
				dates = append(dates, day)
			}
			a := dailyMap[day]
			a.sum += float64(*d.BloodSugar)
			a.count++
		}
	}

	if len(dailyMap) == 0 {
		return nil // Tidak ada data gula darah yang valid
	}

	// Urutkan tanggal ASC
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	var dailyValues []float64
	var total float64
	var dayCount int

	for _, day := range dates {
		a := dailyMap[day]
		if a.count == 0 {
			continue
		}
		avg := a.sum / float64(a.count)
		dailyValues = append(dailyValues, avg)
		total += avg
		dayCount++
	}

	if dayCount == 0 {
		return nil
	}

	avgValue := total / float64(dayCount)

	// Hitung persentase perubahan periode dari nilai harian
	changePercent := calculatePeriodChangePercent(dailyValues)

	avgValueInt := int(avgValue)
	return &response.BloodSugarSummary{
		AvgValue:      roundTo2Decimals(avgValue),
		ChangePercent: roundTo2Decimals(changePercent),
		Status:        s.getBloodSugarStatus(avgValueInt),
		NormalRange:   "70-140 mg/dL (WHO - Gula Darah Sewaktu)",
	}
}

// calculateWeightSummary menghitung ringkasan berat badan dengan nullable-aware
// Berbasis agregasi harian (1 nilai rata-rata berat per hari).
func (s *HealthDataService) calculateWeightSummary(data []entity.HealthData) *response.WeightSummary {
	if len(data) == 0 {
		return nil
	}

	// Kelompokkan per hari berdasarkan RecordDate, hanya data dengan berat badan
	type agg struct {
		sum   float64
		count int
	}

	dailyMap := make(map[time.Time]*agg)
	var dates []time.Time

	for _, d := range data {
		if d.Weight != nil {
			recordDateJakarta := timezoneUtils.ToJakarta(d.RecordDate)
			day := timezoneUtils.DateInJakarta(recordDateJakarta.Year(), recordDateJakarta.Month(), recordDateJakarta.Day(), 0, 0, 0, 0)
			if _, ok := dailyMap[day]; !ok {
				dailyMap[day] = &agg{}
				dates = append(dates, day)
			}
			a := dailyMap[day]
			a.sum += *d.Weight
			a.count++
		}
	}

	if len(dailyMap) == 0 {
		return nil // Tidak ada data berat badan yang valid
	}

	// Urutkan tanggal ASC
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	var dailyWeights []float64
	var total float64
	var dayCount int

	for _, day := range dates {
		a := dailyMap[day]
		if a.count == 0 {
			continue
		}
		avg := a.sum / float64(a.count)
		dailyWeights = append(dailyWeights, avg)
		total += avg
		dayCount++
	}

	if dayCount == 0 {
		return nil
	}

	avgWeight := total / float64(dayCount)

	// Hitung persentase perubahan periode dari nilai harian
	changePercent := calculatePeriodChangePercent(dailyWeights)

	// Tentukan tren berbasis changePercent
	trend := "Stabil"
	if changePercent > 1 {
		trend = "Naik"
	} else if changePercent < -1 {
		trend = "Turun"
	}

	// Hitung BMI jika ada tinggi badan
	var bmi *float64
	var bmiValue float64
	hasHeight := false
	for _, d := range data {
		if d.HeightCM != nil {
			hasHeight = true
			heightCM := *d.HeightCM
			bmiValue = calculateBMI(avgWeight, heightCM)
			bmi = &bmiValue
			break // Ambil tinggi badan pertama yang ditemukan
		}
	}

	// Jika ada beberapa data dengan tinggi badan berbeda, hitung rata-rata BMI
	if hasHeight {
		var totalBMI float64
		var bmiCount int
		for _, d := range data {
			if d.Weight != nil && d.HeightCM != nil {
				bmiVal := calculateBMI(*d.Weight, *d.HeightCM)
				totalBMI += bmiVal
				bmiCount++
			}
		}
		if bmiCount > 0 {
			avgBMI := totalBMI / float64(bmiCount)
			bmiValue = roundTo2Decimals(avgBMI)
			bmi = &bmiValue
		}
	}

	return &response.WeightSummary{
		AvgWeight:     roundTo2Decimals(avgWeight),
		BMI:           bmi,
		Trend:         trend,
		ChangePercent: roundTo2Decimals(changePercent),
	}
}

// calculateActivitySummary menghitung ringkasan aktivitas
// Berbasis agregasi harian (1 estimasi aktivitas per hari).
func (s *HealthDataService) calculateActivitySummary(data []entity.HealthData) *response.ActivitySummary {
	if len(data) == 0 {
		return nil
	}

	// Kelompokkan per hari berdasarkan RecordDate, hanya data dengan activity tidak kosong
	type agg struct {
		count int
	}

	dailyMap := make(map[time.Time]*agg)
	var dates []time.Time

	for _, d := range data {
		if d.Activity != nil && *d.Activity != "" {
			recordDateJakarta := timezoneUtils.ToJakarta(d.RecordDate)
			day := timezoneUtils.DateInJakarta(recordDateJakarta.Year(), recordDateJakarta.Month(), recordDateJakarta.Day(), 0, 0, 0, 0)
			if _, ok := dailyMap[day]; !ok {
				dailyMap[day] = &agg{}
				dates = append(dates, day)
			}
			a := dailyMap[day]
			a.count++
		}
	}

	if len(dailyMap) == 0 {
		// Tidak ada aktivitas valid, tetap kembalikan 0 dengan changePercent 0
		return &response.ActivitySummary{
			TotalSteps:    0,
			TotalCalories: 0,
			ChangePercent: 0,
		}
	}

	// Urutkan tanggal ASC
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	// Asumsi: setiap entri aktivitas dalam satu hari merepresentasikan satu "aktivitas".
	// Estimasi asli: 1 aktivitas = 1000 langkah & 200 kalori.
	var totalSteps int
	var totalCalories float64
	var dailySteps []float64

	for _, day := range dates {
		a := dailyMap[day]
		if a.count == 0 {
			continue
		}
		stepsForDay := a.count * 1000
		caloriesForDay := float64(a.count) * 200.0

		totalSteps += stepsForDay
		totalCalories += caloriesForDay
		dailySteps = append(dailySteps, float64(stepsForDay))
	}

	// Hitung persentase perubahan periode dari nilai harian (berdasarkan langkah)
	changePercent := calculatePeriodChangePercent(dailySteps)

	return &response.ActivitySummary{
		TotalSteps:    totalSteps,
		TotalCalories: roundTo2Decimals(totalCalories),
		ChangePercent: roundTo2Decimals(changePercent),
	}
}

