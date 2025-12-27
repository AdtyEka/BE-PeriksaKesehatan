package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
	"errors"
)

// HealthDataService menangani business logic untuk data kesehatan
type HealthDataService struct {
	healthDataRepo *repository.HealthDataRepository
}

// NewHealthDataService membuat instance baru dari HealthDataService
func NewHealthDataService(healthDataRepo *repository.HealthDataRepository) *HealthDataService {
	return &HealthDataService{
		healthDataRepo: healthDataRepo,
	}
}

// ValidateHealthData melakukan validasi range nilai data kesehatan
func (s *HealthDataService) ValidateHealthData(req *request.HealthDataRequest) error {
	// Validasi systolic: 90 – 180
	if req.Systolic < 90 || req.Systolic > 180 {
		return errors.New("systolic harus berada dalam range 90-180 mmHg")
	}

	// Validasi diastolic: 60 – 120
	if req.Diastolic < 60 || req.Diastolic > 120 {
		return errors.New("diastolic harus berada dalam range 60-120 mmHg")
	}

	// Validasi blood_sugar: 60 – 300
	if req.BloodSugar < 60 || req.BloodSugar > 300 {
		return errors.New("blood_sugar harus berada dalam range 60-300 mg/dL")
	}

	// Validasi weight: 20 – 200
	if req.Weight < 20 || req.Weight > 200 {
		return errors.New("weight harus berada dalam range 20-200 kg")
	}

	// Validasi heart_rate: 40 – 180
	if req.HeartRate < 40 || req.HeartRate > 180 {
		return errors.New("heart_rate harus berada dalam range 40-180 bpm")
	}

	return nil
}

// CreateHealthData membuat data kesehatan baru dengan validasi
func (s *HealthDataService) CreateHealthData(userID uint, req *request.HealthDataRequest) (*response.HealthDataResponse, error) {
	// Validasi data
	if err := s.ValidateHealthData(req); err != nil {
		return nil, err
	}

	// Buat entity dari request
	healthData := &entity.HealthData{
		UserID:     userID,
		Systolic:   req.Systolic,
		Diastolic:  req.Diastolic,
		BloodSugar: req.BloodSugar,
		Weight:     req.Weight,
		HeartRate:  req.HeartRate,
		Activity:   req.Activity,
	}

	// Simpan ke database
	if err := s.healthDataRepo.CreateHealthData(healthData); err != nil {
		return nil, err
	}

	// Buat response
	resp := &response.HealthDataResponse{
		ID:         healthData.ID,
		UserID:     healthData.UserID,
		Systolic:   healthData.Systolic,
		Diastolic:  healthData.Diastolic,
		BloodSugar: healthData.BloodSugar,
		Weight:     healthData.Weight,
		HeartRate:  healthData.HeartRate,
		Activity:   healthData.Activity,
		CreatedAt:  healthData.CreatedAt,
	}

	return resp, nil
}

// GetHealthDataByUserID mengambil semua data kesehatan berdasarkan UserID
func (s *HealthDataService) GetHealthDataByUserID(userID uint) ([]entity.HealthData, error) {
	return s.healthDataRepo.GetHealthDataByUserID(userID)
}

