package service

import (
	"strings"
	"math"
)

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

