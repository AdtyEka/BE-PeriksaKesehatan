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

	// Ambil data dengan filter untuk summary dan reading history
	healthDataList, err := s.healthDataRepo.GetHealthDataByUserIDWithFilter(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Ambil data 90 hari untuk trend charts (agar bisa menampilkan 7Days, 1Month, 3Months)
	now := s.getIndonesiaTime()
	trendEndDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	trendStartDate := trendEndDate.AddDate(0, 0, -89) // 90 hari termasuk hari ini
	trendDataList, err := s.healthDataRepo.GetHealthDataByUserIDWithFilter(userID, trendStartDate, trendEndDate)
	if err != nil {
		return nil, err
	}

	// Ambil data periode sebelumnya untuk perbandingan
	periodLength := endDate.Sub(startDate)
	prevDataList, _ := s.healthDataRepo.GetHealthDataForComparison(userID, startDate, endDate, periodLength)

	// Filter berdasarkan metrik jika ada
	filteredData := s.filterByMetrics(healthDataList, req.Metrics)
	filteredTrendData := s.filterByMetrics(trendDataList, req.Metrics)

	// Buat response
	result := &response.HealthHistoryResponse{}

	// Ringkasan statistik (gunakan data sesuai time range request)
	result.Summary = s.calculateSummary(filteredData, prevDataList, req.Metrics)

	// Grafik tren (gunakan data 90 hari)
	result.TrendCharts = s.calculateTrendCharts(filteredTrendData, req.Metrics)

	// Catatan pembacaan (gunakan data sesuai time range request)
	result.ReadingHistory = s.buildReadingHistory(filteredData)

	return result, nil
}

// parseTimeRange mengkonversi time_range ke startDate dan endDate
func (s *HealthDataService) parseTimeRange(req *request.HealthHistoryRequest) (time.Time, time.Time, error) {
	now := s.getIndonesiaTime()
	// endDate adalah hari ini (akhir hari untuk memastikan semua data hari ini termasuk)
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	var startDate time.Time

	// Jika custom range, gunakan StartDate dan EndDate
	if req.TimeRange == "custom" {
		if req.StartDate == nil || req.EndDate == nil {
			return time.Time{}, time.Time{}, errors.New("start_date dan end_date wajib diisi untuk custom range")
		}
		startDate = time.Date(req.StartDate.Year(), req.StartDate.Month(), req.StartDate.Day(), 0, 0, 0, 0, req.StartDate.Location())
		endDate = time.Date(req.EndDate.Year(), req.EndDate.Month(), req.EndDate.Day(), 23, 59, 59, 0, req.EndDate.Location())
	} else {
		// Default atau time range yang sudah ditentukan
		// Untuk 7days: ambil 7 hari terakhir termasuk hari ini (hari ini + 6 hari sebelumnya)
		// Untuk 30days: ambil 30 hari terakhir termasuk hari ini (hari ini + 29 hari sebelumnya)
		switch req.TimeRange {
		case "30days":
			startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -29) // 30 hari termasuk hari ini
		case "3months":
			startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, -3, 0)
		default: // "7days" atau kosong
			startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -6) // 7 hari termasuk hari ini
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

