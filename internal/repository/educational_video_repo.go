package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"

	"gorm.io/gorm"
)

// EducationalVideoRepository adalah struct yang menampung koneksi database untuk educational videos
type EducationalVideoRepository struct {
	db *gorm.DB
}

// NewEducationalVideoRepository membuat instance baru dari EducationalVideoRepository
func NewEducationalVideoRepository(db *gorm.DB) *EducationalVideoRepository {
	return &EducationalVideoRepository{
		db: db,
	}
}

// CreateEducationalVideo melakukan INSERT video edukasi baru ke database
func (r *EducationalVideoRepository) CreateEducationalVideo(video *entity.EducationalVideo) error {
	result := r.db.Create(video)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetEducationalVideosByHealthCondition mengambil semua video berdasarkan kondisi kesehatan
func (r *EducationalVideoRepository) GetEducationalVideosByHealthCondition(healthCondition string) ([]entity.EducationalVideo, error) {
	var videos []entity.EducationalVideo
	result := r.db.Where("health_condition = ?", healthCondition).
		Order("created_at DESC").
		Find(&videos)
	if result.Error != nil {
		return nil, result.Error
	}
	return videos, nil
}

// GetEducationalVideoByID mengambil video berdasarkan ID
func (r *EducationalVideoRepository) GetEducationalVideoByID(id uint) (*entity.EducationalVideo, error) {
	var video entity.EducationalVideo
	result := r.db.First(&video, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("video tidak ditemukan")
		}
		return nil, result.Error
	}
	return &video, nil
}

// GetAllEducationalVideos mengambil semua video edukasi
func (r *EducationalVideoRepository) GetAllEducationalVideos() ([]entity.EducationalVideo, error) {
	var videos []entity.EducationalVideo
	result := r.db.Order("created_at DESC").Find(&videos)
	if result.Error != nil {
		return nil, result.Error
	}
	return videos, nil
}

