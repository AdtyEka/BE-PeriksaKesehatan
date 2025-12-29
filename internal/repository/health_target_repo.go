package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"

	"gorm.io/gorm"
)

type HealthTargetRepository struct {
	db *gorm.DB
}

func NewHealthTargetRepository(db *gorm.DB) *HealthTargetRepository {
	return &HealthTargetRepository{
		db: db,
	}
}

func (r *HealthTargetRepository) CreateHealthTarget(healthTarget *entity.HealthTarget) error {
	result := r.db.Create(healthTarget)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *HealthTargetRepository) GetHealthTargetByUserID(userID uint) (*entity.HealthTarget, error) {
	var healthTarget entity.HealthTarget
	result := r.db.Where("user_id = ?", userID).First(&healthTarget)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("health target tidak ditemukan")
		}
		return nil, result.Error
	}
	return &healthTarget, nil
}

func (r *HealthTargetRepository) UpdateHealthTarget(userID uint, healthTarget *entity.HealthTarget) error {
	result := r.db.Model(&entity.HealthTarget{}).Where("user_id = ?", userID).Updates(healthTarget)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("health target tidak ditemukan")
	}
	return nil
}

func (r *HealthTargetRepository) UpsertHealthTarget(healthTarget *entity.HealthTarget) error {
	existing, err := r.GetHealthTargetByUserID(healthTarget.UserID)
	if err != nil && err.Error() == "health target tidak ditemukan" {
		return r.CreateHealthTarget(healthTarget)
	}
	if err != nil {
		return err
	}
	healthTarget.ID = existing.ID
	return r.UpdateHealthTarget(healthTarget.UserID, healthTarget)
}

