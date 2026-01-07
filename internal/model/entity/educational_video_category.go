package entity

import "time"

// EducationalVideoCategory adalah junction table untuk relasi many-to-many
// antara EducationalVideo dan Category
type EducationalVideoCategory struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	EducationalVideoID uint      `gorm:"not null;index" json:"educational_video_id"`
	CategoryID        uint      `gorm:"not null;index" json:"category_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	
	// Relasi
	EducationalVideo EducationalVideo `gorm:"foreignKey:EducationalVideoID" json:"educational_video,omitempty"`
	Category         Category         `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// TableName mengembalikan nama tabel untuk GORM
func (EducationalVideoCategory) TableName() string {
	return "educational_video_categories"
}

