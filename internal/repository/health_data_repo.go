package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"
	"time"

	"gorm.io/gorm"
	timezoneUtils "BE-PeriksaKesehatan/pkg/utils"
)

type HealthDataRepository struct {
	db *gorm.DB
}

func NewHealthDataRepository(db *gorm.DB) *HealthDataRepository {
	return &HealthDataRepository{
		db: db,
	}
}

func (r *HealthDataRepository) CreateHealthData(healthData *entity.HealthData) error {
	result := r.db.Create(healthData)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *HealthDataRepository) GetHealthDataByID(id uint) (*entity.HealthData, error) {
	var healthData entity.HealthData
	result := r.db.First(&healthData, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("data kesehatan tidak ditemukan")
		}
		return nil, result.Error
	}
	return &healthData, nil
}

func (r *HealthDataRepository) GetHealthDataByUserID(userID uint) ([]entity.HealthData, error) {
	var healthDataList []entity.HealthData
	result := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&healthDataList)
	if result.Error != nil {
		return nil, result.Error
	}
	return healthDataList, nil
}

func (r *HealthDataRepository) GetAllHealthData() ([]entity.HealthData, error) {
	var healthDataList []entity.HealthData
	result := r.db.Order("created_at DESC").Find(&healthDataList)
	if result.Error != nil {
		return nil, result.Error
	}
	return healthDataList, nil
}

func (r *HealthDataRepository) GetHealthDataByUserIDWithFilter(userID uint, startDate, endDate time.Time) ([]entity.HealthData, error) {
	var healthDataList []entity.HealthData
	// Normalisasi startDate dan endDate ke awal hari untuk perbandingan (dalam timezone Asia/Jakarta)
	startDateJakarta := timezoneUtils.ToJakarta(startDate)
	endDateJakarta := timezoneUtils.ToJakarta(endDate)
	startDateNormalized := timezoneUtils.DateInJakarta(startDateJakarta.Year(), startDateJakarta.Month(), startDateJakarta.Day(), 0, 0, 0, 0)
	endDateNormalized := timezoneUtils.DateInJakarta(endDateJakarta.Year(), endDateJakarta.Month(), endDateJakarta.Day(), 0, 0, 0, 0)
	
	// Format tanggal sebagai string untuk perbandingan yang lebih reliable
	startDateStr := startDateNormalized.Format("2006-01-02")
	endDateStr := endDateNormalized.Format("2006-01-02")
	
	// Filter berdasarkan record_date (bukan created_at) karena summary dan trend charts menggunakan record_date
	// Gunakan DATE() untuk memastikan perbandingan hanya berdasarkan tanggal, bukan waktu
	// Range inklusif: DATE(record_date) >= startDate AND DATE(record_date) <= endDate
	query := r.db.Where("user_id = ?", userID).
		Where("DATE(record_date) >= ? AND DATE(record_date) <= ?", startDateStr, endDateStr).
		Order("record_date DESC, created_at DESC")
	
	result := query.Find(&healthDataList)
	if result.Error != nil {
		return nil, result.Error
	}
	return healthDataList, nil
}

func (r *HealthDataRepository) GetHealthDataForComparison(userID uint, startDate, endDate time.Time, periodDuration time.Duration) ([]entity.HealthData, error) {
	periodLength := endDate.Sub(startDate)
	prevEndDate := startDate.Add(-24 * time.Hour)
	prevStartDate := prevEndDate.Add(-periodLength)
	
	var healthDataList []entity.HealthData
	query := r.db.Where("user_id = ?", userID).
		Where("created_at >= ? AND created_at <= ?", prevStartDate, prevEndDate).
		Order("created_at DESC")
	
	result := query.Find(&healthDataList)
	if result.Error != nil {
		return nil, result.Error
	}
	return healthDataList, nil
}

func (r *HealthDataRepository) GetLatestHealthDataByUserID(userID uint) (*entity.HealthData, error) {
	var healthData entity.HealthData
	result := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&healthData)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &healthData, nil
}

// GetHealthDataByUserIDAndDate mencari record berdasarkan user_id dan record_date
// Digunakan untuk daily record system (1 record per hari per user)
func (r *HealthDataRepository) GetHealthDataByUserIDAndDate(userID uint, date time.Time) (*entity.HealthData, error) {
	var healthData entity.HealthData
	// Gunakan DATE() untuk membandingkan hanya bagian tanggal, bukan waktu
	result := r.db.Where("user_id = ? AND DATE(record_date) = DATE(?)", userID, date).
		First(&healthData)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &healthData, nil
}

// UpdateHealthData melakukan partial update pada health data
// Hanya field yang tidak nil yang akan di-update
// Field yang nil akan diabaikan (tidak di-overwrite dengan NULL)
func (r *HealthDataRepository) UpdateHealthData(healthData *entity.HealthData) error {
	// Buat map untuk menyimpan field yang akan di-update
	updates := make(map[string]interface{})
	
	// Hanya tambahkan field yang tidak nil ke updates
	if healthData.Systolic != nil {
		updates["systolic"] = *healthData.Systolic
	}
	if healthData.Diastolic != nil {
		updates["diastolic"] = *healthData.Diastolic
	}
	if healthData.BloodSugar != nil {
		updates["blood_sugar"] = *healthData.BloodSugar
	}
	if healthData.Weight != nil {
		updates["weight"] = *healthData.Weight
	}
	if healthData.HeightCM != nil {
		updates["height_cm"] = *healthData.HeightCM
	}
	if healthData.HeartRate != nil {
		updates["heart_rate"] = *healthData.HeartRate
	}
	if healthData.Activity != nil {
		updates["activity"] = *healthData.Activity
	}
	
	// Update hanya jika ada field yang akan di-update
	if len(updates) > 0 {
		result := r.db.Model(healthData).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
	}
	
	return nil
}

