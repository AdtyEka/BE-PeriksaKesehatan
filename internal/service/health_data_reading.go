package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"fmt"
	"sort"
)

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

