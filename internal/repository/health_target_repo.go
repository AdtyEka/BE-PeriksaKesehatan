package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"

	"gorm.io/gorm"
)

// HealthTargetRepository adalah struct yang menampung koneksi database untuk health targets
type HealthTargetRepository struct {
	db *gorm.DB
}

// NewHealthTargetRepository membuat instance baru dari HealthTargetRepository
func NewHealthTargetRepository(db *gorm.DB) *HealthTargetRepository {
	return &HealthTargetRepository{
		db: db,
	}
}

// ==================== OPERASI CREATE (INSERT) ====================

// CreateHealthTarget melakukan INSERT health target baru ke database
func (r *HealthTargetRepository) CreateHealthTarget(healthTarget *entity.HealthTarget) error {
	result := r.db.Create(healthTarget)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ==================== OPERASI READ (SELECT) ====================

// GetHealthTargetByUserID melakukan SELECT health target berdasarkan UserID
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

// ==================== OPERASI UPDATE ====================

// UpdateHealthTarget melakukan UPDATE health target berdasarkan UserID
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

// UpsertHealthTarget melakukan INSERT atau UPDATE health target (jika belum ada, insert; jika sudah ada, update)
func (r *HealthTargetRepository) UpsertHealthTarget(healthTarget *entity.HealthTarget) error {
	// Cek apakah sudah ada
	existing, err := r.GetHealthTargetByUserID(healthTarget.UserID)
	if err != nil && err.Error() == "health target tidak ditemukan" {
		// Belum ada, insert baru
		return r.CreateHealthTarget(healthTarget)
	}
	if err != nil {
		return err
	}
	// Sudah ada, update
	healthTarget.ID = existing.ID
	return r.UpdateHealthTarget(healthTarget.UserID, healthTarget)
}

