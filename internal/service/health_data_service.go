package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
	"errors"
	"time"

	"BE-PeriksaKesehatan/pkg/utils"
)

// HealthDataService menangani business logic untuk data kesehatan
type HealthDataService struct {
	healthDataRepo   *repository.HealthDataRepository
	personalInfoRepo *repository.PersonalInfoRepository
}

// NewHealthDataService membuat instance baru dari HealthDataService
func NewHealthDataService(healthDataRepo *repository.HealthDataRepository, personalInfoRepo *repository.PersonalInfoRepository) *HealthDataService {
	return &HealthDataService{
		healthDataRepo:   healthDataRepo,
		personalInfoRepo: personalInfoRepo,
	}
}

// getIndonesiaTime mengembalikan waktu saat ini dengan timezone Indonesia (WIB - Asia/Jakarta)
func (s *HealthDataService) getIndonesiaTime() time.Time {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback jika gagal load timezone
		location = time.FixedZone("WIB", 7*60*60) // UTC+7
	}
	return time.Now().In(location)
}

// convertToIndonesiaTime mengkonversi waktu ke timezone Indonesia (WIB - Asia/Jakarta)
func (s *HealthDataService) convertToIndonesiaTime(t time.Time) time.Time {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback jika gagal load timezone
		location = time.FixedZone("WIB", 7*60*60) // UTC+7
	}
	return t.In(location)
}

func (s *HealthDataService) CreateHealthData(userID uint, req *request.HealthDataRequest) (*response.HealthDataResponse, error) {
	// Ambil waktu sekarang dengan timezone Indonesia
	now := s.getIndonesiaTime()
	
	// Normalisasi ke tanggal saja (00:00:00) untuk record_date
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	// Cari record dengan record_date = hari ini
	existingData, err := s.healthDataRepo.GetHealthDataByUserIDAndDate(userID, currentDate)
	if err != nil {
		return nil, err
	}

	// Validasi field yang dikirim (hanya sekali untuk INSERT dan UPDATE)
	if err := s.validateHealthDataFields(req); err != nil {
		return nil, err
	}

	var healthData *entity.HealthData

	if existingData != nil {
		// UPDATE: Partial update pada record hari ini
		// Field yang tidak dikirim (nil) TIDAK DIUBAH
		healthData = existingData

		// Update hanya field yang dikirim (tidak nil)
		s.updateHealthDataFields(healthData, req)

		// Lakukan update (partial update - hanya field yang tidak nil)
		if err := s.healthDataRepo.UpdateHealthData(healthData); err != nil {
			return nil, err
		}

		// Reload data untuk mendapatkan updated_at terbaru
		healthData, err = s.healthDataRepo.GetHealthDataByID(healthData.ID)
		if err != nil {
			return nil, err
		}
	} else {
		// INSERT: Buat record baru untuk hari ini
		// User boleh mengirim 1 field saja atau bahkan kosong (semua NULL)

		// Set expired_at = hari ini 23:59:59 dengan timezone Indonesia
		expiredAt := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

		healthData = &entity.HealthData{
			UserID:     userID,
			RecordDate: currentDate, // Sudah dinormalisasi ke 00:00:00
			ExpiredAt:  &expiredAt,
		}

		// Set field yang dikirim (field yang tidak dikirim tetap NULL)
		s.updateHealthDataFields(healthData, req)

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

	return resp, nil
}

// GetHealthDataByUserID mengembalikan 1 record health data untuk hari ini milik user
// Menggunakan daily record system: mengembalikan record dengan record_date = CURRENT_DATE
func (s *HealthDataService) GetHealthDataByUserID(userID uint) (*entity.HealthData, error) {
	now := s.getIndonesiaTime()
	// Normalisasi ke tanggal saja (00:00:00)
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return s.healthDataRepo.GetHealthDataByUserIDAndDate(userID, currentDate)
}

// validateHealthDataFields melakukan validasi field health data yang dikirim
// Digunakan untuk menghindari duplikasi validasi antara INSERT dan UPDATE
func (s *HealthDataService) validateHealthDataFields(req *request.HealthDataRequest) error {
	// Validasi systolic dan diastolic harus bersamaan
	if req.Systolic != nil || req.Diastolic != nil {
		if (req.Systolic != nil && req.Diastolic == nil) ||
			(req.Systolic == nil && req.Diastolic != nil) {
			return errors.New("systolic dan diastolic harus dikirim bersamaan")
		}
		if req.Systolic != nil && req.Diastolic != nil {
			if err := utils.ValidateNullableInt(req.Systolic, "systolic", 0, 180); err != nil {
				return err
			}
			if err := utils.ValidateNullableInt(req.Diastolic, "diastolic", 0, 120); err != nil {
				return err
			}
		}
	}
	
	// Validasi field lainnya
	if req.BloodSugar != nil {
		if err := utils.ValidateNullableInt(req.BloodSugar, "blood_sugar", 0, 300); err != nil {
			return err
		}
	}
	if req.Weight != nil {
		if err := utils.ValidateNullableFloat64(req.Weight, "weight", 20.0, 200.0); err != nil {
			return err
		}
	}
	if req.Height != nil {
		if err := utils.ValidateNullableInt(req.Height, "height", 50, 250); err != nil {
			return err
		}
	}
	if req.HeartRate != nil {
		if err := utils.ValidateNullableInt(req.HeartRate, "heart_rate", 0, 180); err != nil {
			return err
		}
	}
	
	return nil
}

// updateHealthDataFields mengupdate field health data dari request
// Hanya field yang tidak nil yang akan diupdate
func (s *HealthDataService) updateHealthDataFields(healthData *entity.HealthData, req *request.HealthDataRequest) {
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
}
