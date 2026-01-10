package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"sort"
	"strconv"
	"time"
)

// MapHealthHistoryToAPIResponse mengubah HealthHistoryResponse internal ke struktur API response baru
// Mengambil data untuk semua periode (7Days, 1Month, 3Months) dan melakukan mapping
// Menggunakan logic yang sudah ada tanpa mengubah perhitungan
func (s *HealthDataService) MapHealthHistoryToAPIResponse(
	userID uint,
	req *request.HealthHistoryRequest,
	internalResp *response.HealthHistoryResponse,
) (*response.HealthHistoryAPIResponse, error) {
	// Tentukan rentang waktu untuk mendapatkan start_date dan end_date global
	startDate, endDate, err := s.parseTimeRange(req)
	if err != nil {
		return nil, err
	}

	// Ambil data untuk semua periode menggunakan query yang sudah ada
	now := time.Now()
	endDateGlobal := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	
	// Data untuk 7Days (7 hari terakhir)
	startDate7Days := endDateGlobal.AddDate(0, 0, -6)
	data7Days, err := s.healthDataRepo.GetHealthDataByUserIDWithFilter(userID, startDate7Days, endDateGlobal)
	if err != nil {
		return nil, err
	}
	filteredData7Days := s.filterByMetrics(data7Days, req.Metrics)
	
	// Data untuk 1Month (30 hari terakhir)
	startDate1Month := endDateGlobal.AddDate(0, 0, -29)
	data1Month, err := s.healthDataRepo.GetHealthDataByUserIDWithFilter(userID, startDate1Month, endDateGlobal)
	if err != nil {
		return nil, err
	}
	filteredData1Month := s.filterByMetrics(data1Month, req.Metrics)
	
	// Data untuk 3Months (90 hari terakhir)
	startDate3Months := endDateGlobal.AddDate(0, 0, -89)
	data3Months, err := s.healthDataRepo.GetHealthDataByUserIDWithFilter(userID, startDate3Months, endDateGlobal)
	if err != nil {
		return nil, err
	}
	filteredData3Months := s.filterByMetrics(data3Months, req.Metrics)

	// Build API response
	apiResp := &response.HealthHistoryAPIResponse{
		Summary: response.HealthHistorySummaryByRange{
			StartDate: startDate.Format("2006-01-02"),
			EndDate:   endDate.Format("2006-01-02"),
		},
		TrendCharts: internalResp.TrendCharts, // Tetap sama seperti sebelumnya
	}

	// Summary untuk 7Days (tanpa weeks)
	summary7Days := s.calculateSummary(filteredData7Days, []entity.HealthData{}, req.Metrics)
	apiResp.Summary.Days7 = &summary7Days

	// Summary untuk 1Month (dengan weeks)
	apiResp.Summary.Month1 = s.buildSummaryWithWeeks(filteredData1Month, req.Metrics, startDate1Month)

	// Summary untuk 3Months (dengan weeks)
	apiResp.Summary.Months3 = s.buildSummaryWithWeeks(filteredData3Months, req.Metrics, startDate3Months)

	// Reading history untuk 7Days
	apiResp.ReadingHistory.Days7 = s.buildReadingHistory(filteredData7Days)

	// Reading history untuk 1Month (dengan grouping)
	readingHistory1Month := s.buildReadingHistory(filteredData1Month)
	apiResp.ReadingHistory.Month1 = &response.ReadingHistoryGrouped{
		StartDate: startDate1Month.Format("2006-01-02"),
		EndDate:   endDateGlobal.Format("2006-01-02"),
		Records:   readingHistory1Month,
	}

	// Reading history untuk 3Months (dengan grouping)
	readingHistory3Months := s.buildReadingHistory(filteredData3Months)
	apiResp.ReadingHistory.Months3 = &response.ReadingHistoryGrouped{
		StartDate: startDate3Months.Format("2006-01-02"),
		EndDate:   endDateGlobal.Format("2006-01-02"),
		Records:   readingHistory3Months,
	}

	return apiResp, nil
}

// buildSummaryWithWeeks membangun summary dengan agregasi per minggu
// Menggunakan data yang sudah ada dan logic calculateSummary yang sudah ada
func (s *HealthDataService) buildSummaryWithWeeks(
	data []entity.HealthData,
	metrics []string,
	rangeStartDate time.Time,
) *response.HealthSummaryWithWeeks {
	// Hitung summary keseluruhan menggunakan logic yang sudah ada
	overallSummary := s.calculateSummary(data, []entity.HealthData{}, metrics)

	// Kelompokkan data per minggu menggunakan logic yang sama dengan getWeekRange
	// Tapi menghitung week number dengan benar untuk range berapa pun (tidak dibatasi Week 4)
	weekMap := make(map[string][]entity.HealthData)
	weekDateMap := make(map[string]struct {
		startDate string
		endDate   string
	})

	for _, d := range data {
		// Gunakan logic yang sama dengan getWeekRange untuk menghitung start_date dan end_date
		weekKey, startDateStr, endDateStr := s.getWeekRangeForPeriod(d.RecordDate, rangeStartDate)
		weekMap[weekKey] = append(weekMap[weekKey], d)
		
		// Simpan start_date dan end_date untuk week ini (gunakan yang pertama ditemukan)
		if _, exists := weekDateMap[weekKey]; !exists {
			weekDateMap[weekKey] = struct {
				startDate string
				endDate   string
			}{startDate: startDateStr, endDate: endDateStr}
		}
	}

	// Hitung summary per minggu menggunakan logic yang sudah ada
	var weeks []response.HealthSummaryWeek
	for weekKey, weekData := range weekMap {
		if len(weekData) == 0 {
			continue
		}

		// Ambil start_date dan end_date dari map
		weekDates := weekDateMap[weekKey]

		// Hitung summary untuk minggu ini menggunakan logic yang sudah ada
		weekSummary := s.calculateSummary(weekData, []entity.HealthData{}, metrics)

		weeks = append(weeks, response.HealthSummaryWeek{
			Week:      weekKey,
			StartDate: weekDates.startDate,
			EndDate:   weekDates.endDate,
			Summary:   weekSummary,
		})
	}

	// Sort weeks by start_date (terlama ke terbaru)
	sort.Slice(weeks, func(i, j int) bool {
		return weeks[i].StartDate < weeks[j].StartDate
	})

	return &response.HealthSummaryWithWeeks{
		HealthSummaryResponse: overallSummary,
		Weeks:                  weeks,
	}
}

// getWeekRangeForPeriod menghitung range minggu untuk tanggal tertentu
// Menggunakan logic yang sama dengan getWeekRange tapi tanpa batasan Week 4
// Ini memungkinkan perhitungan weeks yang benar untuk periode panjang (3Months)
func (s *HealthDataService) getWeekRangeForPeriod(date time.Time, rangeStartDate time.Time) (weekKey, startDate, endDate string) {
	// Cari hari Senin dari minggu tersebut (logic yang sama dengan getWeekRange)
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // Minggu = 7
	}
	daysFromMonday := weekday - 1 // 0 = Senin, 6 = Minggu

	monday := date.AddDate(0, 0, -daysFromMonday)
	sunday := monday.AddDate(0, 0, 6)

	startDate = monday.Format("2006-01-02")
	endDate = sunday.Format("2006-01-02")

	// Hitung week number berdasarkan posisi dalam range (logic yang sama dengan getWeekRange)
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
	// TIDAK ada batasan maksimal Week 4 - ini memungkinkan perhitungan yang benar untuk periode panjang

	weekKey = "Week " + strconv.Itoa(weekNum)
	return
}
