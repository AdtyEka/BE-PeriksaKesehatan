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

// GetEducationalVideosByCategoryID mengambil semua video berdasarkan kategori ID
func (r *EducationalVideoRepository) GetEducationalVideosByCategoryID(categoryID uint) ([]entity.EducationalVideo, error) {
	var videos []entity.EducationalVideo
	result := r.db.Where("category_id = ?", categoryID).
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

// GetAllEducationalVideosByCategoryIDs mengambil semua video berdasarkan list kategori ID
func (r *EducationalVideoRepository) GetAllEducationalVideosByCategoryIDs(categoryIDs []uint) (map[uint][]entity.EducationalVideo, error) {
	var videos []entity.EducationalVideo
	result := r.db.Where("category_id IN ?", categoryIDs).
		Order("category_id ASC, created_at DESC").
		Find(&videos)
	if result.Error != nil {
		return nil, result.Error
	}

	// Group videos by category_id
	videosByCategory := make(map[uint][]entity.EducationalVideo)
	for _, video := range videos {
		videosByCategory[video.CategoryID] = append(videosByCategory[video.CategoryID], video)
	}

	return videosByCategory, nil
}

