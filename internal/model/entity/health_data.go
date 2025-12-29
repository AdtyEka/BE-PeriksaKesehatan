package entity

import "time"

// HealthData adalah representasi tabel health_data di database
// Setiap record merepresentasikan snapshot kondisi kesehatan pada satu waktu
type HealthData struct {
	ID         uint      `gorm:"primaryKey" json:"id"`                    // Primary key HealthData (auto increment: 1, 2, 3, ...)
	UserID     uint      `gorm:"not null;index" json:"user_id"`          // Foreign key ke users (referensi ke User.ID)
	User       User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"` // Relasi ke User (opsional untuk query)
	Systolic   *int      `gorm:"type:int" json:"systolic"`                // Tekanan darah sistolik (mmHg)
	Diastolic  *int      `gorm:"type:int" json:"diastolic"`              // Tekanan darah diastolik (mmHg)
	BloodSugar *int      `gorm:"type:int" json:"blood_sugar"`            // Gula darah (mg/dL)
	Weight     *float64  `gorm:"type:decimal(10,2)" json:"weight"`       // Berat badan (kg)
	HeartRate  *int      `gorm:"type:int" json:"heart_rate"`             // Detak jantung (bpm)
	Activity   *string   `gorm:"type:text" json:"activity"`              // Aktivitas terbaru (opsional)
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

