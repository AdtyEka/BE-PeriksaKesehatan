package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"
	"strings"
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
// Menggunakan database constraint untuk menghindari race condition
func (r *AuthRepository) BlacklistToken(token string, userID uint, expiresAt time.Time) error {
	// Validasi input
	if token == "" {
		return gorm.ErrInvalidValue
	}

	// Token belum di-blacklist, tambahkan ke database
	// Database constraint unique akan mencegah duplikasi (idempotent)
	blacklistedToken := &entity.BlacklistedToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	result := r.db.Create(blacklistedToken)
	if result.Error != nil {
		// Jika error karena duplicate (token sudah ada), return nil (idempotent)
		// PostgreSQL mengembalikan error dengan kode 23505 untuk unique constraint violation
		errMsg := strings.ToLower(result.Error.Error())
		if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique") || 
		   strings.Contains(errMsg, "23505") {
			return nil // Token sudah di-blacklist, anggap sukses
		}
		return result.Error
	}
	return nil
}

// IsTokenBlacklisted mengecek apakah token sudah di-blacklist
func (r *AuthRepository) IsTokenBlacklisted(token string) (bool, error) {
	// Validasi input
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

// CleanupExpiredTokens menghapus token yang sudah kadaluarsa dari blacklist
// Bisa dipanggil secara berkala untuk membersihkan database
// Menggunakan soft delete tidak diperlukan karena ini data cleanup
func (r *AuthRepository) CleanupExpiredTokens() error {
	now := time.Now()
	result := r.db.Where("expires_at < ?", now).Delete(&entity.BlacklistedToken{})
	if result.Error != nil {
		return result.Error
	}
	// Log jumlah token yang dihapus (opsional, untuk monitoring)
	// log.Printf("Info: Menghapus %d token yang sudah kadaluarsa", result.RowsAffected)
	return nil
}

