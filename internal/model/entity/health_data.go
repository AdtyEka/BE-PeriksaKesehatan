package entity

import "time"

type HealthData struct {
	ID         uint      `gorm:"primaryKey" json:"id"`                    // Primary key HealthData (auto increment: 1, 2, 3, ...)
	UserID     uint      `gorm:"not null;index" json:"user_id"`          // Foreign key ke users (referensi ke User.ID) - TETAP WAJIB
	User       User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"` // Relasi ke User (opsional untuk query)
	
	// Field nullable-ready (akan nullable di DB setelah migration)
	// TODO: Update gorm tag setelah DB migration: gorm:"type:int" (remove "not null")
	Systolic   *int      `gorm:"not null" json:"systolic"`                // Tekanan darah sistolik (mmHg) - nullable-ready
	Diastolic  *int      `gorm:"not null" json:"diastolic"`              // Tekanan darah diastolik (mmHg) - nullable-ready
	BloodSugar *int      `gorm:"not null" json:"blood_sugar"`           // Gula darah (mg/dL) - nullable-ready
	Weight     *float64  `gorm:"not null" json:"weight"`                 // Berat badan (kg) - nullable-ready
	HeightCM   *int      `gorm:"type:int;column:height_cm" json:"height,omitempty"` // Tinggi badan dalam cm
	HeartRate  *int      `gorm:"not null" json:"heart_rate"`             // Detak jantung (bpm) - nullable-ready
	Activity   *string   `gorm:"type:text" json:"activity"`               // Aktivitas terbaru (opsional) - sudah nullable dari awal
	
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

