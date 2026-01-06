package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
)

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

	// Hitung status berdasarkan rata-rata (menggunakan kombinasi sistolik dan diastolik)
	avgSystolicInt := int(avgSystolic)
	avgDiastolicInt := int(avgDiastolic)
	status := s.getBloodPressureStatus(avgSystolicInt, avgDiastolicInt)

	return &response.BloodPressureSummary{
		AvgSystolic:     roundTo2Decimals(avgSystolic),
		AvgDiastolic:    roundTo2Decimals(avgDiastolic),
		ChangePercent:   roundTo2Decimals(changePercent),
		SystolicStatus:  status, // Status berdasarkan kombinasi sistolik dan diastolik
		DiastolicStatus: status, // Status sama karena menggunakan kombinasi
		NormalRange:     "90-139 / 60-89 mmHg (WHO)",
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

	avgValueInt := int(avgValue)
	return &response.BloodSugarSummary{
		AvgValue:     roundTo2Decimals(avgValue),
		ChangePercent: roundTo2Decimals(changePercent),
		Status:       s.getBloodSugarStatus(avgValueInt),
		NormalRange:  "70-140 mg/dL (WHO - Gula Darah Sewaktu)",
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

	// Hitung BMI jika ada tinggi badan
	var bmi *float64
	var bmiValue float64
	hasHeight := false
	for _, d := range validData {
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
		for _, d := range validData {
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
func (s *HealthDataService) calculateActivitySummary(data, prevData []entity.HealthData) *response.ActivitySummary {
	if len(data) == 0 {
		return nil
	}

	// Untuk aktivitas, default adalah 0 jika tidak ada input aktivitas yang valid
	// Hanya menghitung jika ada activity field yang tidak kosong
	totalSteps := 0
	totalCalories := 0.0
	activityCount := 0
	for _, d := range data {
		if d.Activity != nil && *d.Activity != "" {
			activityCount++
			// Estimasi: setiap aktivitas = 1000 langkah dan 200 kalori
			// Ini bisa disesuaikan jika ada format spesifik di activity field
			totalSteps += 1000
			totalCalories += 200.0
		}
	}

	// Hitung periode sebelumnya
	prevTotalSteps := 0
	prevActivityCount := 0
	for _, d := range prevData {
		if d.Activity != nil && *d.Activity != "" {
			prevActivityCount++
			prevTotalSteps += 1000
		}
	}

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

