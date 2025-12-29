package entity

import "time"

// EducationalVideo adalah representasi tabel educational_videos di database
type EducationalVideo struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	VideoTitle      string    `gorm:"type:varchar(255);not null" json:"video_title"`
	VideoURL        string    `gorm:"type:text;not null" json:"video_url"`
	HealthCondition string    `gorm:"type:varchar(100);not null;index" json:"health_condition"` // Kondisi kesehatan terkait (e.g., "Tekanan Darah Tinggi")
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

