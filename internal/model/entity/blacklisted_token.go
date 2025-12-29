package entity

import "time"

// BlacklistedToken adalah representasi tabel blacklisted_tokens di database
// Menyimpan token yang sudah di-blacklist untuk logout
type BlacklistedToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"type:text;unique;not null;index" json:"-"` // Token JWT yang di-blacklist
	UserID    uint      `gorm:"not null;index" json:"user_id"`            // User ID pemilik token
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`         // Waktu kadaluarsa token (dari claim exp)
	CreatedAt time.Time `json:"created_at"`
}

// TableName mengembalikan nama tabel untuk GORM
func (BlacklistedToken) TableName() string {
	return "blacklisted_tokens"
}

