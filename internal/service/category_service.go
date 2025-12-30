package service

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
)

// CategoryService menangani business logic untuk categories
type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

// NewCategoryService membuat instance baru dari CategoryService
func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// GetAllCategories mengambil semua kategori
func (s *CategoryService) GetAllCategories() ([]entity.Category, error) {
	return s.categoryRepo.GetAllCategories()
}

// GetCategoryByID mengambil kategori berdasarkan ID
func (s *CategoryService) GetCategoryByID(id uint) (*entity.Category, error) {
	return s.categoryRepo.GetCategoryByID(id)
}

