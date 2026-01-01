package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
	"errors"

	"BE-PeriksaKesehatan/pkg/utils"
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

func (s *HealthDataService) CreateHealthData(userID uint, req *request.HealthDataRequest) (*response.HealthDataResponse, error) {
	// Cek apakah ada record terakhir milik user
	existingData, err := s.healthDataRepo.GetLatestHealthDataByUserID(userID)
	if err != nil {
		return nil, err
	}

	var healthData *entity.HealthData
	var isUpdate bool

	if existingData != nil {
		// UPDATE: Merge data lama + data baru
		// Field yang tidak dikirim (nil) akan tetap menggunakan nilai lama
		healthData = existingData
		isUpdate = true

		// Untuk update, validasi lebih fleksibel: Activity saja sudah cukup
		// Tapi tetap validasi field yang dikirim jika ada
		if req.Systolic != nil || req.Diastolic != nil {
			if (req.Systolic != nil && req.Diastolic == nil) ||
				(req.Systolic == nil && req.Diastolic != nil) {
				return nil, errors.New("systolic dan diastolic harus dikirim bersamaan")
			}
			// Setelah cek di atas, kita tahu keduanya tidak nil atau keduanya nil
			// Validasi hanya jika keduanya tidak nil
			if req.Systolic != nil && req.Diastolic != nil {
				if err := utils.ValidateNullableInt(req.Systolic, "systolic", 90, 180); err != nil {
					return nil, err
				}
				if err := utils.ValidateNullableInt(req.Diastolic, "diastolic", 60, 120); err != nil {
					return nil, err
				}
			}
		}
		if req.BloodSugar != nil {
			if err := utils.ValidateNullableInt(req.BloodSugar, "blood_sugar", 60, 300); err != nil {
				return nil, err
			}
		}
		if req.Weight != nil {
			if err := utils.ValidateNullableFloat64(req.Weight, "weight", 20.0, 200.0); err != nil {
				return nil, err
			}
		}
		if req.Height != nil {
			if err := utils.ValidateNullableInt(req.Height, "height", 50, 250); err != nil {
				return nil, err
			}
		}
		if req.HeartRate != nil {
			if err := utils.ValidateNullableInt(req.HeartRate, "heart_rate", 40, 180); err != nil {
				return nil, err
			}
		}

		// Update hanya field yang dikirim (tidak nil)
		if req.Systolic != nil {
			healthData.Systolic = req.Systolic
		}
		if req.Diastolic != nil {
			healthData.Diastolic = req.Diastolic
		}
		if req.BloodSugar != nil {
			healthData.BloodSugar = req.BloodSugar
		}
		if req.Weight != nil {
			healthData.Weight = req.Weight
		}
		if req.Height != nil {
			healthData.HeightCM = req.Height
		}
		if req.HeartRate != nil {
			healthData.HeartRate = req.HeartRate
		}
		if req.Activity != nil {
			healthData.Activity = req.Activity
		}

		// Validasi: minimal ada 1 field yang akan di-update
		hasUpdate := req.Systolic != nil || req.Diastolic != nil || req.BloodSugar != nil ||
			req.Weight != nil || req.Height != nil || req.HeartRate != nil || req.Activity != nil
		if !hasUpdate {
			return nil, errors.New("minimal satu field harus diisi untuk update")
		}

		// Lakukan update
		if err := s.healthDataRepo.UpdateHealthData(healthData); err != nil {
			return nil, err
		}

		// Reload data untuk mendapatkan updated_at terbaru
		healthData, err = s.healthDataRepo.GetHealthDataByID(healthData.ID)
		if err != nil {
			return nil, err
		}
	} else {
		// INSERT: Buat record baru
		// Untuk insert, validasi lebih ketat: minimal satu metrik kesehatan harus diisi
		if err := s.ValidateHealthData(req); err != nil {
			return nil, err
		}

		healthData = &entity.HealthData{
			UserID: userID,
		}

		if req.Systolic != nil {
			healthData.Systolic = req.Systolic
		}
		if req.Diastolic != nil {
			healthData.Diastolic = req.Diastolic
		}
		if req.BloodSugar != nil {
			healthData.BloodSugar = req.BloodSugar
		}
		if req.Weight != nil {
			healthData.Weight = req.Weight
		}
		if req.Height != nil {
			healthData.HeightCM = req.Height
		}
		if req.HeartRate != nil {
			healthData.HeartRate = req.HeartRate
		}
		if req.Activity != nil {
			healthData.Activity = req.Activity
		}

		if err := s.healthDataRepo.CreateHealthData(healthData); err != nil {
			return nil, err
		}
	}

	// Build response
	resp := &response.HealthDataResponse{
		ID:        healthData.ID,
		UserID:    healthData.UserID,
		CreatedAt: healthData.CreatedAt,
	}

	// Set semua field (termasuk yang nil) untuk response
	resp.Systolic = healthData.Systolic
	resp.Diastolic = healthData.Diastolic
	resp.BloodSugar = healthData.BloodSugar
	resp.Weight = healthData.Weight
	resp.Height = healthData.HeightCM
	resp.HeartRate = healthData.HeartRate
	resp.Activity = healthData.Activity

	_ = isUpdate // Mark as used (untuk debugging jika diperlukan)

	return resp, nil
}

// GetHealthDataByUserID mengembalikan 1 record health data terbaru milik user
// Menggunakan logic inkremental: selalu mengembalikan record aktif terbaru
func (s *HealthDataService) GetHealthDataByUserID(userID uint) (*entity.HealthData, error) {
	return s.healthDataRepo.GetLatestHealthDataByUserID(userID)
}
