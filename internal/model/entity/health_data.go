package entity

import "time"

// HealthData adalah representasi tabel health_data di database
// Setiap record merepresentasikan snapshot kondisi kesehatan pada satu waktu
//
// MIGRASI NULLABLE-READY:
// Field-field berikut sekarang nullable-ready untuk mendukung partial data input.
// 
// CATATAN PENTING - DB MIGRATION REQUIRED:
// Untuk mengaktifkan nullable di database, perlu menjalankan migration:
//   ALTER TABLE health_data 
//     MODIFY COLUMN systolic INT NULL,
//     MODIFY COLUMN diastolic INT NULL,
//     MODIFY COLUMN blood_sugar INT NULL,
//     MODIFY COLUMN weight DOUBLE NULL,
//     MODIFY COLUMN heart_rate INT NULL;
//
// Sebelum migration, field nullable akan di-handle di application layer.
// Setelah validasi, field yang dikirim akan di-assign, field yang tidak dikirim akan nil.
//
// BUSINESS RULE:
// - Minimal satu metrik kesehatan harus diisi (divalidasi di service layer)
// - Systolic dan Diastolic harus dikirim bersamaan jika salah satu dikirim
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
	HeartRate  *int      `gorm:"not null" json:"heart_rate"`             // Detak jantung (bpm) - nullable-ready
	Activity   *string   `gorm:"type:text" json:"activity"`               // Aktivitas terbaru (opsional) - sudah nullable dari awal
	
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

