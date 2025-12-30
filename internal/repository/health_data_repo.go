package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"
	"time"

	"gorm.io/gorm"
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
	query := r.db.Where("user_id = ?", userID).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Order("created_at DESC")
	
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

