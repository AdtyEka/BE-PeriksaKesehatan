package entity

import "time"

// AlertStatus adalah enum untuk status alert
type AlertStatus string

const (
	AlertStatusCritical AlertStatus = "Critical"
	AlertStatusHigh     AlertStatus = "High"
	AlertStatusModerate AlertStatus = "Moderate"
	AlertStatusLow      AlertStatus = "Low"
)

// HealthAlert adalah representasi tabel health_alerts di database
type HealthAlert struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	User           User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	AlertType      string     `gorm:"type:varchar(100);not null" json:"alert_type"`      // Jenis alert (e.g., "Tekanan Darah Tinggi")
	Message        string     `gorm:"type:text;not null" json:"message"`                // Pesan alert
	Status         AlertStatus `gorm:"type:varchar(20);not null" json:"status"`          // Status: Critical, High, Moderate, Low
	Recommendations string    `gorm:"type:text" json:"recommendations"`                 // Rekomendasi (JSON array sebagai string)
	RecordedAt     time.Time  `gorm:"not null" json:"recorded_at"`                      // Waktu pengukuran data kesehatan
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

