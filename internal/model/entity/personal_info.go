package entity

import "time"

// PersonalInfo merepresentasikan informasi pribadi user (1:1 dengan User)
type PersonalInfo struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID uint `gorm:"not null;uniqueIndex" json:"user_id"`
	User   User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`

	// Informasi pribadi
	Name      string     `gorm:"type:varchar(100);not null" json:"name"`
	BirthDate *time.Time `gorm:"type:date" json:"birth_date,omitempty"`
	Phone     *string    `gorm:"type:varchar(15)" json:"phone,omitempty"`
	Address   *string    `gorm:"type:text" json:"address,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName mengembalikan nama tabel untuk GORM
func (PersonalInfo) TableName() string {
	return "personal_infos"
}

