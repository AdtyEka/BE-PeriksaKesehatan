package service

import (
	"strings"
	"math"
)

// Konstanta status kesehatan sesuai standar WHO
const (
	StatusRendah = "RENDAH"
	StatusNormal = "NORMAL"
	StatusTinggi = "TINGGI"
)

// getBloodPressureStatus menentukan status tekanan darah berdasarkan kombinasi sistolik dan diastolik (WHO)
// RENDAH jika sistolik < 90 atau diastolik < 60
// NORMAL jika sistolik 90–139 dan diastolik 60–89
// TINGGI jika sistolik ≥ 140 atau diastolik ≥ 90
func (s *HealthDataService) getBloodPressureStatus(systolic, diastolic int) string {
	if systolic < 90 || diastolic < 60 {
		return StatusRendah
	}
	if systolic >= 140 || diastolic >= 90 {
		return StatusTinggi
	}
	if systolic >= 90 && systolic <= 139 && diastolic >= 60 && diastolic <= 89 {
		return StatusNormal
	}
	// Fallback untuk kasus edge case
	return StatusNormal
}

// getBloodSugarStatus menentukan status gula darah sewaktu (WHO)
// RENDAH jika < 70 mg/dL
// NORMAL jika 70–140 mg/dL
// TINGGI jika > 140 mg/dL
func (s *HealthDataService) getBloodSugarStatus(value int) string {
	if value < 70 {
		return StatusRendah
	}
	if value > 140 {
		return StatusTinggi
	}
	return StatusNormal
}

// calculateBMI menghitung BMI berdasarkan berat badan (kg) dan tinggi badan (cm)
func calculateBMI(weightKg float64, heightCm int) float64 {
	if heightCm <= 0 {
		return 0
	}
	heightM := float64(heightCm) / 100.0
	return weightKg / (heightM * heightM)
}

// getBMIStatus menentukan status berdasarkan BMI (WHO)
// RENDAH jika BMI < 18.5
// NORMAL jika BMI 18.5–24.9
// TINGGI jika BMI ≥ 25
func (s *HealthDataService) getBMIStatus(bmi float64) string {
	if bmi < 18.5 {
		return StatusRendah
	}
	if bmi >= 25 {
		return StatusTinggi
	}
	return StatusNormal
}

// getHeartRateStatus menentukan status detak jantung
// Tetap menggunakan logika yang sama untuk konsistensi
func (s *HealthDataService) getHeartRateStatus(value int) string {
	if value >= 60 && value <= 100 {
		return StatusNormal
	}
	if value < 60 {
		return StatusRendah
	}
	return StatusTinggi
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

