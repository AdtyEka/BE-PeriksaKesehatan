package entity

import "time"

// User adalah representasi tabel users di Supabase
type User struct {
	ID          uint         `gorm:"primaryKey" json:"id"`                    // Primary key User (auto increment: 1, 2, 3, ...)
	Nama        string       `gorm:"type:varchar(100);not null" json:"nama"`
	Username    string       `gorm:"type:varchar(50);unique;not null" json:"username"`
	Email       string       `gorm:"type:varchar(100);unique;not null" json:"email"`
	Password    string       `gorm:"type:varchar(255);not null" json:"-"`
	HealthData  []HealthData `gorm:"foreignKey:UserID" json:"health_data,omitempty"` // Relasi one-to-many ke HealthData (opsional untuk query)
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

