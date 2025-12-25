package service

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
)

// UserService menangani business logic untuk user
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService membuat instance baru dari UserService
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUserByID mengambil user berdasarkan ID
func (s *UserService) GetUserByID(id uint) (*entity.User, error) {
	return s.userRepo.GetUserByID(id)
}

// GetAllUsers mengambil semua user
func (s *UserService) GetAllUsers() ([]entity.User, error) {
	return s.userRepo.GetAllUsers()
}

