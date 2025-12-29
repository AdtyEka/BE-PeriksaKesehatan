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

type HealthAlertService struct {
	healthAlertRepo *repository.HealthAlertRepository
	healthDataRepo  *repository.HealthDataRepository
}

func NewHealthAlertService(healthAlertRepo *repository.HealthAlertRepository, healthDataRepo *repository.HealthDataRepository) *HealthAlertService {
	return &HealthAlertService{
		healthAlertRepo: healthAlertRepo,
		healthDataRepo:  healthDataRepo,
	}
}

func (s *HealthAlertService) CheckHealthAlerts(userID uint, req *request.HealthAlertRequest) (*response.CheckHealthAlertsResponse, error) {
	if err := s.validateHealthAlertData(req); err != nil {
		return nil, err
	}

	var alerts []response.HealthAlertResponse

	if req.Systolic > 140 || req.Diastolic > 90 {
		alert := s.createBloodPressureAlert(req)
		alerts = append(alerts, alert)
		
		if err := s.saveAlertToDB(userID, alert, req.RecordedAt); err != nil {
			fmt.Printf("Error menyimpan alert tekanan darah: %v\n", err)
		}
	}

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

	weightAlert, err := s.checkWeightLoss(userID, req.Weight, req.RecordedAt)
	if err != nil {
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

func (s *HealthAlertService) validateHealthAlertData(req *request.HealthAlertRequest) error {
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

func (s *HealthAlertService) checkWeightLoss(userID uint, currentWeight float64, recordedAt time.Time) (*response.HealthAlertResponse, error) {
	beforeTime := recordedAt.Add(-24 * time.Hour)
	recentData, err := s.healthAlertRepo.GetRecentWeightData(userID, beforeTime, 5)
	if err != nil {
		return nil, err
	}

	if len(recentData) == 0 {
		return nil, nil
	}

	var sumWeight float64
	var validCount int
	for _, data := range recentData {
		if data.Weight != nil {
			sumWeight += *data.Weight
			validCount++
		}
	}
	
	if validCount == 0 {
		return nil, nil
	}
	
	avgPreviousWeight := sumWeight / float64(validCount)
	if avgPreviousWeight == 0 {
		return nil, nil
	}
	
	weightLossPercent := ((avgPreviousWeight - currentWeight) / avgPreviousWeight) * 100

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

func (s *HealthAlertService) saveAlertToDB(userID uint, alert response.HealthAlertResponse, recordedAt time.Time) error {
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

