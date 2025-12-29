package entity

import "time"

// HealthTarget merepresentasikan target kesehatan per user (1:1)
type HealthTarget struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID uint `gorm:"not null;uniqueIndex" json:"user_id"`
	User   User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`

	// Target tekanan darah (mmHg)
	TargetSystolic  *int `gorm:"type:int" json:"target_systolic,omitempty"`
	TargetDiastolic *int `gorm:"type:int" json:"target_diastolic,omitempty"`

	// Target gula darah (mg/dL)
	TargetBloodSugar *int `gorm:"type:int" json:"target_blood_sugar,omitempty"`

	// Target berat badan (kg)
	TargetWeight *float64 `gorm:"type:decimal(5,2)" json:"target_weight,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


