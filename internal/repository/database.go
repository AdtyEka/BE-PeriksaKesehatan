package repository

import (
	"BE-PeriksaKesehatan/config"
	"BE-PeriksaKesehatan/internal/model/entity"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDB membuka koneksi fisik (kabel) ke database menggunakan GORM/Postgres

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	// Mengambil alamat database dari config.go
	dbURL := cfg.DBURL

	// Membuka koneksi menggunakan driver GORM/Postgres
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Printf("Error: Gagal membuka koneksi ke database: %v", err)
		return nil, err
	}

	// Test koneksi dengan melakukan ping
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Error: Gagal mendapatkan instance database: %v", err)
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		log.Printf("Error: Gagal melakukan ping ke database: %v", err)
		return nil, err
	}

	// AutoMigrate
	err = db.AutoMigrate(
		&entity.User{},
		&entity.HealthData{},
	)
	if err != nil {
		log.Printf("Warning: Gagal melakukan auto-migrate: %v", err)
	} else {
		log.Println("Info: Auto-migrate berhasil!")
	}

	log.Println("Info: Koneksi ke database berhasil dibuat!")
	return db, nil
}

