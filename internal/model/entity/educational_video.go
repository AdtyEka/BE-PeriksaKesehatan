package entity

import "time"

// EducationalVideo adalah representasi tabel educational_videos di database
type EducationalVideo struct {
	ID              uint                        `gorm:"primaryKey" json:"id"`
	VideoTitle      string                      `gorm:"type:varchar(255);not null" json:"video_title"`
	VideoURL        string                      `gorm:"type:text;not null" json:"video_url"`
	CategoryID      *uint                       `gorm:"index" json:"category_id,omitempty"` // Nullable untuk backward compatibility
	Category        *Category                   `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	HealthCondition string                      `gorm:"type:varchar(100);not null" json:"health_condition"`
	CreatedAt       time.Time                   `json:"created_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
	
	// Relasi many-to-many dengan Category melalui junction table
	Categories      []Category                  `gorm:"many2many:educational_video_categories;foreignKey:ID;joinForeignKey:EducationalVideoID;References:ID;joinReferences:CategoryID" json:"categories,omitempty"`
}

