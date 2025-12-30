package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"

	"gorm.io/gorm"
)

// CategoryRepository adalah struct yang menampung koneksi database untuk categories
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository membuat instance baru dari CategoryRepository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

// GetAllCategories mengambil semua kategori dari database
func (r *CategoryRepository) GetAllCategories() ([]entity.Category, error) {
	var categories []entity.Category
	result := r.db.Order("id ASC").Find(&categories)
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}

// GetCategoryByID mengambil kategori berdasarkan ID
func (r *CategoryRepository) GetCategoryByID(id uint) (*entity.Category, error) {
	var category entity.Category
	result := r.db.First(&category, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("kategori tidak ditemukan")
		}
		return nil, result.Error
	}
	return &category, nil
}

