package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"

	"gorm.io/gorm"
)

type PersonalInfoRepository struct {
	db *gorm.DB
}

func NewPersonalInfoRepository(db *gorm.DB) *PersonalInfoRepository {
	return &PersonalInfoRepository{
		db: db,
	}
}

// CreatePersonalInfo membuat personal info baru untuk user
func (r *PersonalInfoRepository) CreatePersonalInfo(personalInfo *entity.PersonalInfo) error {
	if personalInfo == nil {
		return errors.New("personal info tidak boleh nil")
	}
	if personalInfo.UserID == 0 {
		return errors.New("user_id harus diisi")
	}
	if personalInfo.Name == "" {
		return errors.New("name harus diisi")
	}

	result := r.db.Create(personalInfo)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetPersonalInfoByUserID mengambil personal info berdasarkan user_id
func (r *PersonalInfoRepository) GetPersonalInfoByUserID(userID uint) (*entity.PersonalInfo, error) {
	if userID == 0 {
		return nil, errors.New("user_id tidak valid")
	}

	var personalInfo entity.PersonalInfo
	result := r.db.Where("user_id = ?", userID).First(&personalInfo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("personal info tidak ditemukan")
		}
		return nil, result.Error
	}
	return &personalInfo, nil
}

// UpdatePersonalInfo mengupdate personal info berdasarkan user_id
func (r *PersonalInfoRepository) UpdatePersonalInfo(userID uint, updates map[string]interface{}) error {
	if userID == 0 {
		return errors.New("user_id tidak valid")
	}
	if len(updates) == 0 {
		return errors.New("tidak ada data untuk diupdate")
	}

	result := r.db.Model(&entity.PersonalInfo{}).Where("user_id = ?", userID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("personal info tidak ditemukan")
	}
	return nil
}

// CheckPersonalInfoExists mengecek apakah personal info sudah ada untuk user_id tertentu
func (r *PersonalInfoRepository) CheckPersonalInfoExists(userID uint) (bool, error) {
	if userID == 0 {
		return false, errors.New("user_id tidak valid")
	}

	var count int64
	result := r.db.Model(&entity.PersonalInfo{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

