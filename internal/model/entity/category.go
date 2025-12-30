package entity

import "time"

// Category adalah representasi tabel categories di database
type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Kategori  string    `gorm:"type:varchar(100);not null;unique" json:"kategori"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

