package repository

import "gorm.io/gorm"

// AuthRepository untuk operasi repository terkait autentikasi
// File ini siap untuk repository auth di masa depan jika diperlukan
type AuthRepository struct {
	db *gorm.DB
}

// NewAuthRepository membuat instance baru dari AuthRepository
func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

