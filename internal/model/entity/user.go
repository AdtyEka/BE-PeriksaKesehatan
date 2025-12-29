package entity

import "time"

// User adalah representasi tabel users di Supabase
type User struct {
	ID          uint         `gorm:"primaryKey" json:"id"`                    // Primary key User (auto increment: 1, 2, 3, ...)
	Nama        string       `gorm:"type:varchar(100);not null" json:"nama"`
	Username    string       `gorm:"type:varchar(50);unique;not null" json:"username"`
	Email       string       `gorm:"type:varchar(100);unique;not null" json:"email"`
	Password    string       `gorm:"type:varchar(255);not null" json:"-"`
	// Profil & informasi pribadi
	PhotoURL  *string     `gorm:"type:text" json:"photo_url,omitempty"`                    // URL foto profil
	HeightCM  *int        `gorm:"type:int;column:height_cm" json:"height,omitempty"`        // Tinggi badan dalam cm
	BirthDate *time.Time  `gorm:"column:birth_date" json:"birth_date,omitempty"`            // Tanggal lahir
	Phone     *string     `gorm:"type:varchar(30)" json:"phone,omitempty"`
	Address   *string     `gorm:"type:text" json:"address,omitempty"`

	// Pengaturan aplikasi
	NotificationEnabled *bool   `gorm:"default:true;column:notification_enabled" json:"notification_enabled,omitempty"`
	Language            *string `gorm:"type:varchar(10);default:'id'" json:"language,omitempty"`

	HealthData  []HealthData `gorm:"foreignKey:UserID" json:"health_data,omitempty"` // Relasi one-to-many ke HealthData (opsional untuk query)
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

