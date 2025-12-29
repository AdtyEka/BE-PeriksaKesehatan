package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// HealthAlertService menangani business logic untuk health alerts
type HealthAlertService struct {
	healthAlertRepo *repository.HealthAlertRepository
	healthDataRepo  *repository.HealthDataRepository
}

// NewHealthAlertService membuat instance baru dari HealthAlertService
func NewHealthAlertService(healthAlertRepo *repository.HealthAlertRepository, healthDataRepo *repository.HealthDataRepository) *HealthAlertService {
	return &HealthAlertService{
		healthAlertRepo: healthAlertRepo,
		healthDataRepo:  healthDataRepo,
	}
}

// CheckHealthAlerts memeriksa data kesehatan dan menghasilkan alert jika diperlukan
func (s *HealthAlertService) CheckHealthAlerts(userID uint, req *request.HealthAlertRequest) (*response.CheckHealthAlertsResponse, error) {
	// Validasi data
	if err := s.validateHealthAlertData(req); err != nil {
		return nil, err
	}

	var alerts []response.HealthAlertResponse

	// 1. Pemeriksaan Tekanan Darah
	if req.Systolic > 140 || req.Diastolic > 90 {
		alert := s.createBloodPressureAlert(req)
		alerts = append(alerts, alert)
		
		// Simpan ke database
		if err := s.saveAlertToDB(userID, alert, req.RecordedAt); err != nil {
			// Log error tapi jangan gagalkan response
			fmt.Printf("Error menyimpan alert tekanan darah: %v\n", err)
		}
	}

	// 2. Pemeriksaan Gula Darah
	if req.BloodSugar < 70 {
		alert := s.createLowBloodSugarAlert(req)
		alerts = append(alerts, alert)
		
		if err := s.saveAlertToDB(userID, alert, req.RecordedAt); err != nil {
			fmt.Printf("Error menyimpan alert gula darah rendah: %v\n", err)
		}
	} else if req.BloodSugar > 100 {
		alert := s.createHighBloodSugarAlert(req)
		alerts = append(alerts, alert)
		
		if err := s.saveAlertToDB(userID, alert, req.RecordedAt); err != nil {
			fmt.Printf("Error menyimpan alert gula darah tinggi: %v\n", err)
		}
	}

	// 3. Pemeriksaan Penurunan Berat Badan
	weightAlert, err := s.checkWeightLoss(userID, req.Weight, req.RecordedAt)
	if err != nil {
		// Log error tapi jangan gagalkan response
		fmt.Printf("Error memeriksa penurunan berat badan: %v\n", err)
	} else if weightAlert != nil {
		alerts = append(alerts, *weightAlert)
		
		if err := s.saveAlertToDB(userID, *weightAlert, req.RecordedAt); err != nil {
			fmt.Printf("Error menyimpan alert penurunan berat badan: %v\n", err)
		}
	}

	return &response.CheckHealthAlertsResponse{
		Alerts: alerts,
	}, nil
}

// validateHealthAlertData melakukan validasi data kesehatan
func (s *HealthAlertService) validateHealthAlertData(req *request.HealthAlertRequest) error {
	// Validasi nilai numerik
	if req.Systolic <= 0 || req.Systolic > 300 {
		return errors.New("systolic harus valid (1-300 mmHg)")
	}
	if req.Diastolic <= 0 || req.Diastolic > 200 {
		return errors.New("diastolic harus valid (1-200 mmHg)")
	}
	if req.BloodSugar <= 0 || req.BloodSugar > 1000 {
		return errors.New("blood_sugar harus valid (1-1000 mg/dL)")
	}
	if req.Weight <= 0 || req.Weight > 500 {
		return errors.New("weight harus valid (1-500 kg)")
	}
	return nil
}

// createBloodPressureAlert membuat alert untuk tekanan darah tinggi
func (s *HealthAlertService) createBloodPressureAlert(req *request.HealthAlertRequest) response.HealthAlertResponse {
	value := fmt.Sprintf("%d / %d mmHg", req.Systolic, req.Diastolic)
	message := fmt.Sprintf("Tekanan Darah Anda %s — segera konsultasi dengan dokter.", value)
	
	recommendations := []string{
		"Konsultasikan dengan dokter.",
		"Kurangi konsumsi garam.",
		"Monitor tekanan darah secara rutin.",
	}

	return response.HealthAlertResponse{
		AlertType:      "Tekanan Darah Tinggi",
		Value:          value,
		Label:          "Hipertensi",
		Message:        message,
		Status:         string(entity.AlertStatusHigh),
		Recommendations: recommendations,
		RecordedAt:     req.RecordedAt,
	}
}

// createLowBloodSugarAlert membuat alert untuk gula darah rendah
func (s *HealthAlertService) createLowBloodSugarAlert(req *request.HealthAlertRequest) response.HealthAlertResponse {
	value := fmt.Sprintf("%d mg/dL", req.BloodSugar)
	message := fmt.Sprintf("Gula Darah Anda %s — kondisi ini memerlukan perhatian segera.", value)
	
	recommendations := []string{
		"Segera konsumsi makanan atau minuman manis.",
		"Monitor gula darah secara rutin.",
		"Konsultasikan dengan dokter jika sering terjadi.",
	}

	return response.HealthAlertResponse{
		AlertType:      "Gula Darah Rendah",
		Value:          value,
		Label:          "Hipoglikemia",
		Message:        message,
		Status:         string(entity.AlertStatusCritical),
		Recommendations: recommendations,
		RecordedAt:     req.RecordedAt,
	}
}

// createHighBloodSugarAlert membuat alert untuk gula darah tinggi
func (s *HealthAlertService) createHighBloodSugarAlert(req *request.HealthAlertRequest) response.HealthAlertResponse {
	value := fmt.Sprintf("%d mg/dL", req.BloodSugar)
	message := fmt.Sprintf("Gula Darah Anda %s — perhatikan pola makan dan aktivitas fisik.", value)
	
	recommendations := []string{
		"Kurangi konsumsi gula dan karbohidrat.",
		"Tingkatkan aktivitas fisik.",
		"Konsultasikan dengan dokter.",
		"Monitor gula darah secara rutin.",
	}

	return response.HealthAlertResponse{
		AlertType:      "Gula Darah Tinggi",
		Value:          value,
		Label:          "Hiperglikemia",
		Message:        message,
		Status:         string(entity.AlertStatusHigh),
		Recommendations: recommendations,
		RecordedAt:     req.RecordedAt,
	}
}

// checkWeightLoss memeriksa apakah terjadi penurunan berat badan signifikan
func (s *HealthAlertService) checkWeightLoss(userID uint, currentWeight float64, recordedAt time.Time) (*response.HealthAlertResponse, error) {
	// Ambil data berat badan sebelumnya (7 hari terakhir)
	beforeTime := recordedAt.Add(-24 * time.Hour)
	recentData, err := s.healthAlertRepo.GetRecentWeightData(userID, beforeTime, 5)
	if err != nil {
		return nil, err
	}

	// Jika tidak ada data sebelumnya, tidak bisa menentukan penurunan
	if len(recentData) == 0 {
		return nil, nil
	}

	// Hitung rata-rata berat badan sebelumnya
	var sumWeight float64
	for _, data := range recentData {
		sumWeight += data.Weight
	}
	avgPreviousWeight := sumWeight / float64(len(recentData))

	// Hitung persentase penurunan
	weightLossPercent := ((avgPreviousWeight - currentWeight) / avgPreviousWeight) * 100

	// Jika penurunan lebih dari 5% dalam waktu singkat, buat alert
	if weightLossPercent > 5.0 {
		value := fmt.Sprintf("%.2f kg (dari %.2f kg)", currentWeight, avgPreviousWeight)
		message := fmt.Sprintf("Terjadi penurunan berat badan signifikan (%.2f%%) dalam waktu singkat — perhatikan kondisi kesehatan Anda.", weightLossPercent)
		
		recommendations := []string{
			"Konsultasikan dengan dokter untuk evaluasi.",
			"Pastikan asupan nutrisi yang cukup.",
			"Monitor berat badan secara rutin.",
		}

		return &response.HealthAlertResponse{
			AlertType:      "Penurunan Berat Badan",
			Value:          value,
			Label:          "Penurunan Berat Badan Signifikan",
			Message:        message,
			Status:         string(entity.AlertStatusModerate),
			Recommendations: recommendations,
			RecordedAt:     recordedAt,
		}, nil
	}

	return nil, nil
}

// saveAlertToDB menyimpan alert ke database
func (s *HealthAlertService) saveAlertToDB(userID uint, alert response.HealthAlertResponse, recordedAt time.Time) error {
	// Convert recommendations ke JSON string
	recommendationsJSON, err := json.Marshal(alert.Recommendations)
	if err != nil {
		return err
	}

	healthAlert := &entity.HealthAlert{
		UserID:         userID,
		AlertType:      alert.AlertType,
		Message:        alert.Message,
		Status:         entity.AlertStatus(alert.Status),
		Recommendations: string(recommendationsJSON),
		RecordedAt:     recordedAt,
	}

	return s.healthAlertRepo.CreateHealthAlert(healthAlert)
}

