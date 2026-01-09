package service

import (
	"math"
	"strings"
)

// Konstanta status kesehatan sesuai standar WHO
const (
	StatusRendah = "RENDAH"
	StatusNormal = "NORMAL"
	StatusTinggi = "TINGGI"

	bmiLowerThresholdValue = 18.5
	bmiUpperThresholdValue = 25.0
)

// getBloodPressureStatus menentukan status tekanan darah berdasarkan kombinasi sistolik dan diastolik (WHO)
// Menggunakan helper function untuk konsistensi
func (s *HealthDataService) getBloodPressureStatus(systolic, diastolic int) string {
	return getBloodPressureStatusValue(systolic, diastolic)
}

// getBloodSugarStatus menentukan status gula darah sewaktu (WHO)
// Menggunakan helper function untuk konsistensi
func (s *HealthDataService) getBloodSugarStatus(value int) string {
	return getBloodSugarStatusValue(value)
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
	return getBMIStatusValue(bmi)
}

// getHeartRateStatus menentukan status detak jantung
// Menggunakan helper function untuk konsistensi
func (s *HealthDataService) getHeartRateStatus(value int) string {
	return getHeartRateStatusValue(value)
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

// getBMIStatusValue digunakan lintas service untuk memastikan konsistensi klasifikasi BMI
func getBMIStatusValue(bmi float64) string {
	if bmi < bmiLowerThresholdValue {
		return StatusRendah
	}
	if bmi >= bmiUpperThresholdValue {
		return StatusTinggi
	}
	return StatusNormal
}

// getBloodPressureStatusValue menentukan status tekanan darah berdasarkan kombinasi sistolik dan diastolik (WHO)
// Digunakan lintas service untuk konsistensi
// RENDAH jika sistolik < 90 atau diastolik < 60
// NORMAL jika sistolik 90–139 dan diastolik 60–89
// TINGGI jika sistolik ≥ 140 atau diastolik ≥ 90
func getBloodPressureStatusValue(systolic, diastolic int) string {
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

// getBloodSugarStatusValue menentukan status gula darah sewaktu (WHO)
// Digunakan lintas service untuk konsistensi
// RENDAH jika < 70 mg/dL
// NORMAL jika 70–140 mg/dL
// TINGGI jika > 140 mg/dL
func getBloodSugarStatusValue(bloodSugar int) string {
	if bloodSugar < 70 {
		return StatusRendah
	}
	if bloodSugar > 140 {
		return StatusTinggi
	}
	return StatusNormal
}

// getHeartRateStatusValue menentukan status detak jantung
// Digunakan lintas service untuk konsistensi
// NORMAL jika 60-100 bpm
// RENDAH jika < 60 bpm
// TINGGI jika > 100 bpm
func getHeartRateStatusValue(heartRate int) string {
	if heartRate >= 60 && heartRate <= 100 {
		return StatusNormal
	}
	if heartRate < 60 {
		return StatusRendah
	}
	return StatusTinggi
}
