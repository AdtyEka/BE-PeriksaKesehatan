package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"

	"gorm.io/gorm"
)

type HealthAlertRepository struct {
	db *gorm.DB
}

func NewHealthAlertRepository(db *gorm.DB) *HealthAlertRepository {
	return &HealthAlertRepository{db: db}
}

// SaveMany insert banyak alert sekaligus
func (r *HealthAlertRepository) SaveMany(alerts []entity.HealthAlert) error {
	if len(alerts) == 0 {
		return nil
	}
	return r.db.Create(&alerts).Error
}

// GetLatestByUser ambil alert terbaru (opsional)
func (r *HealthAlertRepository) GetLatestByUser(userID uint, limit int) ([]entity.HealthAlert, error) {
	var alerts []entity.HealthAlert
	q := r.db.Where("user_id = ?", userID).Order("recorded_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}
