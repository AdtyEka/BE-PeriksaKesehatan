package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"
	"time"
)

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

