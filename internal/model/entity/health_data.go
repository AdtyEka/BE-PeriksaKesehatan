package entity

import "time"

type HealthData struct {
	ID         uint      `gorm:"primaryKey" json:"id"`                    // Primary key HealthData (auto increment: 1, 2, 3, ...)
	UserID     uint      `gorm:"not null;index" json:"user_id"`          // Foreign key ke users (referensi ke User.ID) - TETAP WAJIB
	User       User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"` // Relasi ke User (opsional untuk query)
	
	// Field kesehatan - SEMUA NULLABLE (tidak wajib)
	Systolic   *int      `gorm:"type:int" json:"systolic"`                // Tekanan darah sistolik (mmHg) - nullable
	Diastolic  *int      `gorm:"type:int" json:"diastolic"`              // Tekanan darah diastolik (mmHg) - nullable
	BloodSugar *int      `gorm:"type:int" json:"blood_sugar"`           // Gula darah (mg/dL) - nullable
	Weight     *float64  `gorm:"type:double precision" json:"weight"`                 // Berat badan (kg) - nullable
	HeightCM   *int      `gorm:"type:int;column:height_cm" json:"height,omitempty"` // Tinggi badan dalam cm - nullable
	HeartRate  *int      `gorm:"type:int" json:"heart_rate"`             // Detak jantung (bpm) - nullable
	Activity   *string   `gorm:"type:text" json:"activity"`               // Aktivitas terbaru - nullable
	
	// Field untuk daily record system
	RecordDate time.Time `gorm:"type:date;not null;index" json:"record_date"` // Tanggal record (1 record per hari per user)
	ExpiredAt  *time.Time `gorm:"type:timestamp" json:"expired_at,omitempty"` // Waktu expired (23:59:59 hari ini) - nullable
	
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

