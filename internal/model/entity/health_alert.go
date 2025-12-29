package entity

import "time"

// HealthAlert = tabel health_alerts
type HealthAlert struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	HealthDataID uint    `gorm:"not null;index" json:"health_data_id"`

	Type       string    `gorm:"type:varchar(50);not null;index" json:"type"`     // blood_pressure, blood_sugar, heart_rate, dll
	Level      string    `gorm:"type:varchar(20);not null;index" json:"level"`    // Normal, Tinggi, Kritis
	Title      string    `gorm:"type:varchar(100);not null" json:"title"`
	Subtitle   string    `gorm:"type:varchar(100)" json:"subtitle"`
	Value      string    `gorm:"type:varchar(50);not null" json:"value"`
	Unit       string    `gorm:"type:varchar(20)" json:"unit"`
	StatusText string    `gorm:"type:varchar(200)" json:"status_text"`

	RecordedAt time.Time `gorm:"not null;index" json:"recorded_at"`
	CreatedAt  time.Time `json:"created_at"`
}

func (HealthAlert) TableName() string {
	return "health_alerts"
}
