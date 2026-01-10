package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	timezoneUtils "BE-PeriksaKesehatan/pkg/utils"
)

// UserProfileInfo berisi informasi profil user untuk laporan
type UserProfileInfo struct {
	Name   string
	Age    *int
	Height *int
}

// getUserProfileInfo mengambil informasi profil user untuk laporan
func (s *HealthDataService) getUserProfileInfo(userID uint) (*UserProfileInfo, error) {
	profile := &UserProfileInfo{
		Name: "User", // Default name
	}

	// Ambil personal info untuk nama dan umur
	personalInfo, err := s.personalInfoRepo.GetPersonalInfoByUserID(userID)
	if err == nil && personalInfo != nil {
		profile.Name = personalInfo.Name
		
		// Hitung umur dari birth_date
		if personalInfo.BirthDate != nil {
			age := s.calculateAge(*personalInfo.BirthDate)
			profile.Age = &age
		}
	}

	// Ambil tinggi badan dari health data terbaru
	latestHealthData, err := s.healthDataRepo.GetLatestHealthDataByUserID(userID)
	if err == nil && latestHealthData != nil && latestHealthData.HeightCM != nil {
		profile.Height = latestHealthData.HeightCM
	}

	return profile, nil
}

// calculateAge menghitung umur dari tanggal lahir
func (s *HealthDataService) calculateAge(birthDate time.Time) int {
	now := timezoneUtils.NowInJakarta()
	birthDateJakarta := timezoneUtils.ToJakarta(birthDate)
	age := now.Year() - birthDateJakarta.Year()

	birthMonthDay := timezoneUtils.DateInJakarta(now.Year(), birthDateJakarta.Month(), birthDateJakarta.Day(), 0, 0, 0, 0)
	if now.Before(birthMonthDay) {
		age--
	}

	return age
}

// GenerateReportCSV menghasilkan laporan dalam format CSV
func (s *HealthDataService) GenerateReportCSV(userID uint, req *request.HealthHistoryRequest) (*bytes.Buffer, string, error) {
	// Ambil data riwayat kesehatan
	historyResp, err := s.GetHealthHistory(userID, req)
	if err != nil {
		return nil, "", err
	}

	// Ambil data profil user
	profileInfo, _ := s.getUserProfileInfo(userID)

	// Buat buffer untuk CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Tentukan rentang waktu untuk nama file
	startDate, endDate, _ := s.parseTimeRange(req)
	timeRangeStr := fmt.Sprintf("%s_to_%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	filename := fmt.Sprintf("riwayat_kesehatan_%s.csv", timeRangeStr)

	// Informasi Profil User
	writer.Write([]string{"=== INFORMASI PASIEN ==="})
	writer.Write([]string{"Nama", profileInfo.Name})
	if profileInfo.Age != nil {
		writer.Write([]string{"Umur", fmt.Sprintf("%d tahun", *profileInfo.Age)})
	}
	if profileInfo.Height != nil {
		writer.Write([]string{"Tinggi Badan", fmt.Sprintf("%d cm", *profileInfo.Height)})
	}
	writer.Write([]string{""})
	writer.Write([]string{""})

	// Header CSV
	headers := []string{
		"Tanggal & Waktu",
		"Jenis Metrik",
		"Nilai",
		"Status",
		"Konteks",
		"Catatan",
	}
	if err := writer.Write(headers); err != nil {
		return nil, "", err
	}

	// Tulis data reading history
	for _, record := range historyResp.ReadingHistory {
		context := ""
		if record.Context != nil {
			context = *record.Context
		}
		notes := ""
		if record.Notes != nil {
			notes = *record.Notes
		}

		row := []string{
			record.DateTime.Format("2006-01-02 15:04:05"),
			record.MetricType,
			record.Value,
			record.Status,
			context,
			notes,
		}
		if err := writer.Write(row); err != nil {
			return nil, "", err
		}
	}

	// Tambahkan ringkasan statistik
	writer.Write([]string{""})
	writer.Write([]string{"=== RINGKASAN STATISTIK ==="})

	// Tekanan Darah
	if historyResp.Summary.BloodPressure != nil {
		writer.Write([]string{""})
		writer.Write([]string{"TEKANAN DARAH"})
		writer.Write([]string{"Rata-rata Systolic", fmt.Sprintf("%.2f mmHg", historyResp.Summary.BloodPressure.AvgSystolic)})
		writer.Write([]string{"Rata-rata Diastolic", fmt.Sprintf("%.2f mmHg", historyResp.Summary.BloodPressure.AvgDiastolic)})
		writer.Write([]string{"Persentase Perubahan", fmt.Sprintf("%.2f%%", historyResp.Summary.BloodPressure.ChangePercent)})
		writer.Write([]string{"Status Systolic", historyResp.Summary.BloodPressure.SystolicStatus})
		writer.Write([]string{"Status Diastolic", historyResp.Summary.BloodPressure.DiastolicStatus})
		writer.Write([]string{"Rentang Normal", historyResp.Summary.BloodPressure.NormalRange})
	}

	// Gula Darah
	if historyResp.Summary.BloodSugar != nil {
		writer.Write([]string{""})
		writer.Write([]string{"GULA DARAH"})
		writer.Write([]string{"Rata-rata", fmt.Sprintf("%.2f mg/dL", historyResp.Summary.BloodSugar.AvgValue)})
		writer.Write([]string{"Persentase Perubahan", fmt.Sprintf("%.2f%%", historyResp.Summary.BloodSugar.ChangePercent)})
		writer.Write([]string{"Status", historyResp.Summary.BloodSugar.Status})
		writer.Write([]string{"Rentang Normal", historyResp.Summary.BloodSugar.NormalRange})
	}

	// Berat Badan
	if historyResp.Summary.Weight != nil {
		writer.Write([]string{""})
		writer.Write([]string{"BERAT BADAN"})
		writer.Write([]string{"Rata-rata", fmt.Sprintf("%.2f kg", historyResp.Summary.Weight.AvgWeight)})
		writer.Write([]string{"Tren", historyResp.Summary.Weight.Trend})
		writer.Write([]string{"Persentase Perubahan", fmt.Sprintf("%.2f%%", historyResp.Summary.Weight.ChangePercent)})
		if historyResp.Summary.Weight.BMI != nil {
			writer.Write([]string{"BMI", fmt.Sprintf("%.2f", *historyResp.Summary.Weight.BMI)})
		}
	}

	// Aktivitas
	if historyResp.Summary.Activity != nil {
		writer.Write([]string{""})
		writer.Write([]string{"AKTIVITAS"})
		writer.Write([]string{"Total Langkah", fmt.Sprintf("%d", historyResp.Summary.Activity.TotalSteps)})
		writer.Write([]string{"Total Kalori", fmt.Sprintf("%.2f", historyResp.Summary.Activity.TotalCalories)})
		writer.Write([]string{"Persentase Perubahan", fmt.Sprintf("%.2f%%", historyResp.Summary.Activity.ChangePercent)})
	}

	writer.Flush()
	return &buf, filename, nil
}

// GenerateReportJSON menghasilkan laporan dalam format JSON
func (s *HealthDataService) GenerateReportJSON(userID uint, req *request.HealthHistoryRequest) (*bytes.Buffer, string, error) {
	// Ambil data riwayat kesehatan
	historyResp, err := s.GetHealthHistory(userID, req)
	if err != nil {
		return nil, "", err
	}

	// Ambil data profil user
	profileInfo, _ := s.getUserProfileInfo(userID)

	// Tentukan rentang waktu untuk nama file
	startDate, endDate, _ := s.parseTimeRange(req)
	timeRangeStr := fmt.Sprintf("%s_to_%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	filename := fmt.Sprintf("riwayat_kesehatan_%s.json", timeRangeStr)

	// Buat struktur profil user untuk JSON
	profileData := map[string]interface{}{
		"nama": profileInfo.Name,
	}
	if profileInfo.Age != nil {
		profileData["umur"] = fmt.Sprintf("%d tahun", *profileInfo.Age)
	}
	if profileInfo.Height != nil {
		profileData["tinggi_badan"] = fmt.Sprintf("%d cm", *profileInfo.Height)
	}

	// Buat struktur laporan lengkap
	report := map[string]interface{}{
		"informasi_pasien": profileData,
		"periode": map[string]interface{}{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
			"time_range": req.TimeRange,
		},
		"ringkasan_statistik": historyResp.Summary,
		"grafik_tren":         historyResp.TrendCharts,
		"catatan_pembacaan":   historyResp.ReadingHistory,
		"generated_at":        timezoneUtils.NowInJakarta().Format("2006-01-02 15:04:05"),
	}

	// Marshal ke JSON dengan indent
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	buf.Write(jsonData)

	return &buf, filename, nil
}

// GenerateReportPDF menghasilkan laporan dalam format PDF dengan desain yang lebih baik
func (s *HealthDataService) GenerateReportPDF(userID uint, req *request.HealthHistoryRequest) (*bytes.Buffer, string, error) {
	// Ambil data riwayat kesehatan
	historyResp, err := s.GetHealthHistory(userID, req)
	if err != nil {
		return nil, "", err
	}

	// Ambil data profil user
	profileInfo, _ := s.getUserProfileInfo(userID)

	// Tentukan rentang waktu untuk nama file
	startDate, endDate, _ := s.parseTimeRange(req)
	timeRangeStr := fmt.Sprintf("%s_to_%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	filename := fmt.Sprintf("riwayat_kesehatan_%s.pdf", timeRangeStr)

	// Buat PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 25, 20)

	// Helper function untuk truncate text agar tidak melebihi lebar
	truncateText := func(text string, maxWidth float64, fontSize float64) string {
		pdf.SetFont("Arial", "", fontSize)
		textWidth := pdf.GetStringWidth(text)
		if textWidth <= maxWidth {
			return text
		}
		// Truncate dengan ellipsis
		for len(text) > 0 {
			text = text[:len(text)-1]
			newWidth := pdf.GetStringWidth(text + "...")
			if newWidth <= maxWidth {
				return text + "..."
			}
		}
		return "..."
	}

	// Helper function untuk draw tabel formal dengan border
	drawFormalTable := func(headers []string, rows [][]string, colWidths []float64) {
		startX := pdf.GetX()
		startY := pdf.GetY()
		headerHeight := 10.0
		cellPadding := 3.0
		totalWidth := 0.0
		for _, w := range colWidths {
			totalWidth += w
		}

		// Header tabel
		pdf.SetFillColor(240, 240, 240) // Abu-abu sangat muda
		pdf.SetDrawColor(0, 0, 0)       // Hitam
		pdf.SetLineWidth(0.5)
		pdf.SetFont("Arial", "B", 10)
		pdf.SetTextColor(0, 0, 0)

		// Draw header background dan border
		pdf.Rect(startX, startY, totalWidth, headerHeight, "FD")

		// Draw vertical lines untuk header
		xPos := startX
		for i := 0; i <= len(colWidths); i++ {
			if i > 0 {
				pdf.Line(xPos, startY, xPos, startY+headerHeight)
			}
			if i < len(colWidths) {
				xPos += colWidths[i]
			}
		}

		// Header text
		xPos = startX + cellPadding
		for i, header := range headers {
			if i < len(colWidths) {
				cellWidth := colWidths[i] - (cellPadding * 2)
				truncatedHeader := truncateText(header, cellWidth, 10)
				pdf.SetXY(xPos, startY+3)
				pdf.Cell(cellWidth, 7, truncatedHeader)
				xPos += colWidths[i]
			}
		}

		// Data rows
		pdf.SetY(startY + headerHeight)
		pdf.SetFont("Arial", "", 9)
		pdf.SetFillColor(255, 255, 255)

		for i, row := range rows {
			rowY := pdf.GetY()

			// Hitung tinggi row yang dibutuhkan (untuk multi-line)
			maxLines := 1
			cellLines := make([][]string, len(row))
			for j, cell := range row {
				if j < len(colWidths) {
					cellWidth := colWidths[j] - (cellPadding * 2)
					lines := pdf.SplitText(cell, cellWidth)
					if len(lines) == 0 {
						lines = []string{""}
					}
					cellLines[j] = lines
					if len(lines) > maxLines {
						maxLines = len(lines)
					}
				}
			}
			lineHeight := 4.0
			actualRowHeight := (lineHeight * float64(maxLines)) + 4.0 // +4 untuk padding atas bawah

			// Zebra striping (baris genap abu-abu sangat muda)
			if i%2 == 0 {
				pdf.SetFillColor(250, 250, 250)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}
			pdf.Rect(startX, rowY, totalWidth, actualRowHeight, "FD")

			// Draw vertical lines untuk row
			xPos = startX
			for j := 0; j <= len(colWidths); j++ {
				if j > 0 {
					pdf.Line(xPos, rowY, xPos, rowY+actualRowHeight)
				}
				if j < len(colWidths) {
					xPos += colWidths[j]
				}
			}

			// Row data dengan MultiCell untuk wrap text
			xPos = startX + cellPadding
			for j := range row {
				if j < len(colWidths) && j < len(cellLines) {
					cellWidth := colWidths[j] - (cellPadding * 2)
					lines := cellLines[j]

					// Tulis setiap baris teks
					currentY := rowY + 2
					for lineIdx, line := range lines {
						if lineIdx > 0 {
							currentY += lineHeight
						}
						pdf.SetXY(xPos, currentY)
						pdf.Cell(cellWidth, lineHeight, line)
					}

					xPos += colWidths[j]
				}
			}

			// Set Y ke posisi terendah dari semua kolom
			pdf.SetY(rowY + actualRowHeight)

			// Cek jika perlu halaman baru
			if pdf.GetY() > 270 {
				pdf.AddPage()
				startY = pdf.GetY()
				startX = pdf.GetX()
				// Redraw header
				pdf.SetFillColor(240, 240, 240)
				pdf.Rect(startX, startY, totalWidth, headerHeight, "FD")

				// Draw vertical lines untuk header
				xPos = startX
				for i := 0; i <= len(colWidths); i++ {
					if i > 0 {
						pdf.Line(xPos, startY, xPos, startY+headerHeight)
					}
					if i < len(colWidths) {
						xPos += colWidths[i]
					}
				}

				xPos = startX + cellPadding
				pdf.SetFont("Arial", "B", 10)
				for i, header := range headers {
					if i < len(colWidths) {
						cellWidth := colWidths[i] - (cellPadding * 2)
						truncatedHeader := truncateText(header, cellWidth, 10)
						pdf.SetXY(xPos, startY+3)
						pdf.Cell(cellWidth, 7, truncatedHeader)
						xPos += colWidths[i]
					}
				}
				pdf.SetY(startY + headerHeight)
				pdf.SetFont("Arial", "", 9)
			}
		}

		pdf.SetFillColor(255, 255, 255) // Reset
	}

	// ========== HALAMAN COVER ==========
	pdf.AddPage()

	// Header formal dengan garis bawah
	pdf.SetXY(20, 30)
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(170, 10, "LAPORAN RIWAYAT KESEHATAN")

	// Garis bawah header
	pdf.SetLineWidth(1.0)
	pdf.SetDrawColor(0, 0, 0)
	pdf.Line(20, 42, 190, 42)
	pdf.Ln(15)

	// ========== INFORMASI PASIEN ==========
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(170, 7, "Informasi Pasien")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	// Nama
	pdf.Cell(50, 7, "Nama:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(120, 7, profileInfo.Name)
	pdf.Ln(7)

	// Umur
	if profileInfo.Age != nil {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(50, 7, "Umur:")
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(120, 7, fmt.Sprintf("%d tahun", *profileInfo.Age))
		pdf.Ln(7)
	}

	// Tinggi Badan
	if profileInfo.Height != nil {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(50, 7, "Tinggi Badan:")
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(120, 7, fmt.Sprintf("%d cm", *profileInfo.Height))
		pdf.Ln(7)
	}
	pdf.Ln(8)

	// Info identitas laporan
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)

	// Periode laporan
	pdf.Cell(50, 7, "Periode Laporan:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(120, 7, fmt.Sprintf("%s s/d %s", startDate.Format("02 Januari 2006"), endDate.Format("02 Januari 2006")))
	pdf.Ln(8)

	// Tanggal dibuat
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(50, 7, "Tanggal Dibuat:")
	pdf.SetFont("Arial", "B", 10)
		pdf.Cell(120, 7, timezoneUtils.NowInJakarta().Format("02 Januari 2006, 15:04:05 WIB"))
	pdf.Ln(20)

	// ========== RINGKASAN STATISTIK ==========
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(170, 10, "RINGKASAN STATISTIK")
	pdf.Ln(12)

	// Buat tabel ringkasan statistik
	var summaryRows [][]string

	// Tekanan Darah
	if historyResp.Summary.BloodPressure != nil {
		summaryRows = append(summaryRows, []string{
			"Tekanan Darah",
			fmt.Sprintf("Systolic: %.1f mmHg", historyResp.Summary.BloodPressure.AvgSystolic),
			fmt.Sprintf("Diastolic: %.1f mmHg", historyResp.Summary.BloodPressure.AvgDiastolic),
			fmt.Sprintf("%s / %s", historyResp.Summary.BloodPressure.SystolicStatus, historyResp.Summary.BloodPressure.DiastolicStatus),
		})
		summaryRows = append(summaryRows, []string{
			"",
			fmt.Sprintf("Rentang Normal: %s", historyResp.Summary.BloodPressure.NormalRange),
			fmt.Sprintf("Perubahan: %.1f%%", historyResp.Summary.BloodPressure.ChangePercent),
			"",
		})
		summaryRows = append(summaryRows, []string{"", "", "", ""}) // Spacer
	}

	// Gula Darah
	if historyResp.Summary.BloodSugar != nil {
		summaryRows = append(summaryRows, []string{
			"Gula Darah",
			fmt.Sprintf("Rata-rata: %.1f mg/dL", historyResp.Summary.BloodSugar.AvgValue),
			fmt.Sprintf("Status: %s", historyResp.Summary.BloodSugar.Status),
			fmt.Sprintf("Rentang Normal: %s", historyResp.Summary.BloodSugar.NormalRange),
		})
		summaryRows = append(summaryRows, []string{
			"",
			fmt.Sprintf("Perubahan: %.1f%%", historyResp.Summary.BloodSugar.ChangePercent),
			"",
			"",
		})
		summaryRows = append(summaryRows, []string{"", "", "", ""}) // Spacer
	}

	// Berat Badan
	if historyResp.Summary.Weight != nil {
		bmiText := ""
		if historyResp.Summary.Weight.BMI != nil {
			bmiText = fmt.Sprintf("BMI: %.1f", *historyResp.Summary.Weight.BMI)
		}
		summaryRows = append(summaryRows, []string{
			"Berat Badan",
			fmt.Sprintf("Rata-rata: %.1f kg", historyResp.Summary.Weight.AvgWeight),
			fmt.Sprintf("Tren: %s", historyResp.Summary.Weight.Trend),
			bmiText,
		})
		summaryRows = append(summaryRows, []string{
			"",
			fmt.Sprintf("Perubahan: %.1f%%", historyResp.Summary.Weight.ChangePercent),
			"",
			"",
		})
		summaryRows = append(summaryRows, []string{"", "", "", ""}) // Spacer
	}

	// Aktivitas
	if historyResp.Summary.Activity != nil {
		summaryRows = append(summaryRows, []string{
			"Aktivitas",
			fmt.Sprintf("Total Langkah: %s", formatNumber(historyResp.Summary.Activity.TotalSteps)),
			fmt.Sprintf("Total Kalori: %.0f kkal", historyResp.Summary.Activity.TotalCalories),
			fmt.Sprintf("Perubahan: %.1f%%", historyResp.Summary.Activity.ChangePercent),
		})
	}

	// Draw tabel ringkasan
	if len(summaryRows) > 0 {
		headers := []string{"Parameter", "Nilai", "Status/Tren", "Keterangan"}
		colWidths := []float64{45, 55, 50, 20} // Lebar kolom disesuaikan agar teks tidak bocor
		drawFormalTable(headers, summaryRows, colWidths)
		pdf.Ln(10)
	}

	// ========== CATATAN PEMBACAAN ==========
	if len(historyResp.ReadingHistory) > 0 {
		// Cek jika perlu halaman baru
		if pdf.GetY() > 250 {
			pdf.AddPage()
		} else {
			pdf.Ln(15)
		}

		pdf.SetFont("Arial", "B", 14)
		pdf.SetTextColor(0, 0, 0)
		pdf.Cell(170, 10, "CATATAN PEMBACAAN")
		pdf.Ln(12)

		// Siapkan data untuk tabel
		var readingRows [][]string
		for _, record := range historyResp.ReadingHistory {
			dateTime := record.DateTime.Format("02/01/2006 15:04")
			metricType := record.MetricType
			value := record.Value
			status := record.Status
			context := ""
			if record.Context != nil {
				context = *record.Context
			}

			notes := ""
			if record.Notes != nil && *record.Notes != "" {
				notes = *record.Notes
			}

			readingRows = append(readingRows, []string{
				dateTime,
				metricType,
				value,
				status,
				context,
				notes,
			})
		}

		// Draw tabel formal
		headers := []string{"Tanggal & Waktu", "Jenis Metrik", "Nilai", "Status", "Konteks", "Catatan"}
		colWidths := []float64{38, 32, 25, 28, 22, 25} // Lebar kolom disesuaikan agar teks tidak bocor
		drawFormalTable(headers, readingRows, colWidths)
	}

	// Footer di setiap halaman
	pdf.AliasNbPages("")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(0, 10, fmt.Sprintf("Halaman %d dari {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	// Output PDF ke buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, "", fmt.Errorf("gagal membuat PDF: %w", err)
	}

	return &buf, filename, nil
}

// formatNumber memformat angka dengan separator ribuan
func formatNumber(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(".")
		}
		result.WriteRune(char)
	}
	return result.String()
}

