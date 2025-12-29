package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"
	"time"

	"gorm.io/gorm"
)

// HealthAlertRepository adalah struct yang menampung koneksi database untuk health alerts
type HealthAlertRepository struct {
	db *gorm.DB
}

// NewHealthAlertRepository membuat instance baru dari HealthAlertRepository
func NewHealthAlertRepository(db *gorm.DB) *HealthAlertRepository {
	return &HealthAlertRepository{
		db: db,
	}
}

// CreateHealthAlert melakukan INSERT alert baru ke database
func (r *HealthAlertRepository) CreateHealthAlert(alert *entity.HealthAlert) error {
	result := r.db.Create(alert)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetHealthAlertsByUserID mengambil semua alert berdasarkan UserID
func (r *HealthAlertRepository) GetHealthAlertsByUserID(userID uint) ([]entity.HealthAlert, error) {
	var alerts []entity.HealthAlert
	result := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&alerts)
	if result.Error != nil {
		return nil, result.Error
	}
	return alerts, nil
}

// GetRecentWeightData mengambil data berat badan terakhir untuk perbandingan
func (r *HealthAlertRepository) GetRecentWeightData(userID uint, beforeTime time.Time, limit int) ([]entity.HealthData, error) {
	var healthDataList []entity.HealthData
	result := r.db.Where("user_id = ? AND created_at < ?", userID, beforeTime).
		Order("created_at DESC").
		Limit(limit).
		Find(&healthDataList)
	if result.Error != nil {
		return nil, result.Error
	}
	return healthDataList, nil
}

// GetHealthAlertByID mengambil alert berdasarkan ID
func (r *HealthAlertRepository) GetHealthAlertByID(id uint) (*entity.HealthAlert, error) {
	var alert entity.HealthAlert
	result := r.db.First(&alert, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("alert tidak ditemukan")
		}
		return nil, result.Error
	}
	return &alert, nil
}

