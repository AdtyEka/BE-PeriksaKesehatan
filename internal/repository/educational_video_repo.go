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

// CreateEducationalVideoWithCategories membuat video baru beserta relasi kategori dalam transaksi
func (r *EducationalVideoRepository) CreateEducationalVideoWithCategories(video *entity.EducationalVideo, categoryIDs []uint) error {
	// Mulai transaksi
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Simpan video
	if err := tx.Create(video).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Simpan relasi many-to-many
	if len(categoryIDs) > 0 {
		// Load categories terlebih dahulu
		var categories []entity.Category
		if err := tx.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Validasi: pastikan semua category IDs ditemukan
		// (Ini double-check, karena sudah divalidasi di service layer)
		if len(categories) != len(categoryIDs) {
			tx.Rollback()
			return errors.New("beberapa kategori tidak ditemukan")
		}

		// Associate categories ke video
		if err := tx.Model(video).Association("Categories").Append(categories); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaksi
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// GetEducationalVideosByCategoryID mengambil semua video berdasarkan kategori ID
// Mendukung data lama (category_id) dan data baru (junction table)
func (r *EducationalVideoRepository) GetEducationalVideosByCategoryID(categoryID uint) ([]entity.EducationalVideo, error) {
	var videos []entity.EducationalVideo
	
	// Query untuk mengambil video dari:
	// 1. Junction table (many-to-many) - data baru
	// 2. category_id langsung - data lama (backward compatibility)
	result := r.db.Where(`
		(id IN (
			SELECT educational_video_id 
			FROM educational_video_categories 
			WHERE category_id = ?
		)) OR (category_id = ?)
	`, categoryID, categoryID).
		Preload("Categories").
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
// Mendukung data lama (category_id) dan data baru (junction table)
func (r *EducationalVideoRepository) GetAllEducationalVideosByCategoryIDs(categoryIDs []uint) (map[uint][]entity.EducationalVideo, error) {
	if len(categoryIDs) == 0 {
		return make(map[uint][]entity.EducationalVideo), nil
	}

	// Query untuk mengambil video dari:
	// 1. Junction table (many-to-many) - data baru
	// 2. category_id langsung - data lama (backward compatibility)
	var videos []entity.EducationalVideo
	result := r.db.Where(`
		(id IN (
			SELECT educational_video_id 
			FROM educational_video_categories 
			WHERE category_id IN ?
		)) OR (category_id IN ?)
	`, categoryIDs, categoryIDs).
		Preload("Categories").
		Order("created_at DESC").
		Find(&videos)
	
	if result.Error != nil {
		return nil, result.Error
	}

	// Group videos by category_id
	// Untuk video dengan many-to-many, kita perlu cek semua kategorinya
	videosByCategory := make(map[uint][]entity.EducationalVideo)
	
	// Map untuk tracking video yang sudah ditambahkan per kategori (untuk menghindari duplikasi)
	videoAddedToCategory := make(map[uint]map[uint]bool) // categoryID -> videoID -> bool
	
	for _, video := range videos {
		// Jika video punya categories (many-to-many), tambahkan ke semua kategorinya
		if len(video.Categories) > 0 {
			for _, category := range video.Categories {
				// Hanya tambahkan jika category ID ada di list yang diminta
				for _, catID := range categoryIDs {
					if category.ID == catID {
						// Inisialisasi map jika belum ada
						if videoAddedToCategory[catID] == nil {
							videoAddedToCategory[catID] = make(map[uint]bool)
						}
						// Cek apakah video sudah ditambahkan ke kategori ini
						if !videoAddedToCategory[catID][video.ID] {
							videosByCategory[catID] = append(videosByCategory[catID], video)
							videoAddedToCategory[catID][video.ID] = true
						}
						break
					}
				}
			}
		}
		
		// Jika video punya category_id langsung (data lama), tambahkan juga
		if video.CategoryID != nil {
			for _, catID := range categoryIDs {
				if *video.CategoryID == catID {
					// Inisialisasi map jika belum ada
					if videoAddedToCategory[catID] == nil {
						videoAddedToCategory[catID] = make(map[uint]bool)
					}
					// Cek apakah video sudah ditambahkan ke kategori ini
					if !videoAddedToCategory[catID][video.ID] {
						videosByCategory[catID] = append(videosByCategory[catID], video)
						videoAddedToCategory[catID][video.ID] = true
					}
					break
				}
			}
		}
	}

	return videosByCategory, nil
}

// GetEducationalVideosByCategoryKategori mengambil video berdasarkan nama kategori
func (r *EducationalVideoRepository) GetEducationalVideosByCategoryKategori(kategori string) ([]entity.EducationalVideo, error) {
	var videos []entity.EducationalVideo
	result := r.db.Joins("JOIN categories ON categories.id = educational_videos.category_id").
		Where("categories.kategori = ?", kategori).
		Order("educational_videos.created_at DESC").
		Find(&videos)
	if result.Error != nil {
		return nil, result.Error
	}
	return videos, nil
}
