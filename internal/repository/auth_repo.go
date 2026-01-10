package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	timezoneUtils "BE-PeriksaKesehatan/pkg/utils"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

// BlacklistToken menambahkan token ke daftar blacklist.
// Method ini idempotent - jika token sudah di-blacklist, tidak akan error.
func (r *AuthRepository) BlacklistToken(token string, userID uint, expiresAt time.Time) error {
	if token == "" {
		return gorm.ErrInvalidValue
	}

	blacklistedToken := &entity.BlacklistedToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	result := r.db.Create(blacklistedToken)
	if result.Error != nil {
		errMsg := strings.ToLower(result.Error.Error())
		if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique") || 
		   strings.Contains(errMsg, "23505") {
			return nil
		}
		return result.Error
	}
	return nil
}

func (r *AuthRepository) IsTokenBlacklisted(token string) (bool, error) {
	if token == "" {
		return false, errors.New("token tidak boleh kosong")
	}

	var count int64
	result := r.db.Model(&entity.BlacklistedToken{}).
		Where("token = ?", token).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// CleanupExpiredTokens menghapus token yang sudah kadaluarsa dari blacklist.
// Bisa dipanggil secara berkala untuk membersihkan database.
func (r *AuthRepository) CleanupExpiredTokens() error {
	now := timezoneUtils.NowInJakarta()
	result := r.db.Where("expires_at < ?", now).Delete(&entity.BlacklistedToken{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

