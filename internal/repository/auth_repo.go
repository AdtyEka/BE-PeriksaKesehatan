package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"time"

	"gorm.io/gorm"
)

// AuthRepository untuk operasi repository terkait autentikasi
type AuthRepository struct {
	db *gorm.DB
}

// NewAuthRepository membuat instance baru dari AuthRepository
func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

// BlacklistToken menambahkan token ke daftar blacklist
// Method ini idempotent - jika token sudah di-blacklist, tidak akan error
func (r *AuthRepository) BlacklistToken(token string, userID uint, expiresAt time.Time) error {
	// Cek apakah token sudah di-blacklist
	isBlacklisted, err := r.IsTokenBlacklisted(token)
	if err != nil {
		return err
	}
	if isBlacklisted {
		// Token sudah di-blacklist, return nil (idempotent)
		return nil
	}

	// Token belum di-blacklist, tambahkan ke database
	blacklistedToken := &entity.BlacklistedToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	result := r.db.Create(blacklistedToken)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// IsTokenBlacklisted mengecek apakah token sudah di-blacklist
func (r *AuthRepository) IsTokenBlacklisted(token string) (bool, error) {
	var count int64
	result := r.db.Model(&entity.BlacklistedToken{}).
		Where("token = ?", token).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// CleanupExpiredTokens menghapus token yang sudah kadaluarsa dari blacklist
// Bisa dipanggil secara berkala untuk membersihkan database
func (r *AuthRepository) CleanupExpiredTokens() error {
	result := r.db.Where("expires_at < ?", time.Now()).Delete(&entity.BlacklistedToken{})
	return result.Error
}

