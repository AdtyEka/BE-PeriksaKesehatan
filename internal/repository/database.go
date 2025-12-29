package repository

import (
	"BE-PeriksaKesehatan/config"
	"BE-PeriksaKesehatan/internal/model/entity"
	"database/sql"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB membuka koneksi fisik (kabel) ke database menggunakan GORM/Postgres
// Konfigurasi ini aman untuk production dan mengatasi masalah prepared statement cache

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	// Mengambil alamat database dari config.go
	dbURL := cfg.DBURL

	// Konfigurasi driver PostgreSQL dengan PreferSimpleProtocol: true
	// Ini menonaktifkan prepared statement di level driver PostgreSQL
	// Solusi untuk error "prepared statement already exists" di Supabase/PostgreSQL
	// Penting: Gunakan postgres.New() dengan Config, bukan postgres.Open()
	postgresConfig := postgres.Config{
		DSN:                  dbURL,
		PreferSimpleProtocol: true, // Gunakan protokol sederhana untuk menghindari prepared statement
	}

	// Konfigurasi GORM dengan prepared statement disabled
	// Ini mengatasi error "prepared statement already exists" di PostgreSQL/Supabase
	gormConfig := &gorm.Config{
		// Nonaktifkan prepared statement untuk menghindari konflik di PostgreSQL
		PrepareStmt: false,
		// Logger untuk development (bisa diubah ke Silent untuk production)
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Membuka koneksi menggunakan driver GORM/Postgres dengan konfigurasi eksplisit
	db, err := gorm.Open(postgres.New(postgresConfig), gormConfig)
	if err != nil {
		log.Printf("Error: Gagal membuka koneksi ke database: %v", err)
		return nil, err
	}

	// Dapatkan instance *sql.DB untuk konfigurasi connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Error: Gagal mendapatkan instance database: %v", err)
		return nil, err
	}

	// Konfigurasi connection pool untuk production
	// SetMaxOpenConns: jumlah maksimum koneksi terbuka ke database
	sqlDB.SetMaxOpenConns(25)

	// SetMaxIdleConns: jumlah maksimum koneksi idle di pool
	sqlDB.SetMaxIdleConns(5)

	// SetConnMaxLifetime: waktu maksimum koneksi bisa digunakan sebelum di-recycle
	// Ini penting untuk database managed seperti Supabase yang mungkin mematikan koneksi lama
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// SetConnMaxIdleTime: waktu maksimum koneksi idle sebelum di-close
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// Test koneksi dengan melakukan ping
	if err := sqlDB.Ping(); err != nil {
		log.Printf("Error: Gagal melakukan ping ke database: %v", err)
		return nil, err
	}

	// AutoMigrate
	err = db.AutoMigrate(
		&entity.User{},
		&entity.HealthData{},
		&entity.BlacklistedToken{},
		&entity.HealthAlert{}, // ✅ tambah ini
	)
	if err != nil {
		log.Printf("Warning: Gagal melakukan auto-migrate: %v", err)
	} else {
		log.Println("Info: Auto-migrate berhasil!")
	}

	log.Println("Info: Koneksi ke database berhasil dibuat!")
	return db, nil
}

// GetDBConnection mengembalikan *sql.DB untuk operasi low-level jika diperlukan
func GetDBConnection(db *gorm.DB) (*sql.DB, error) {
	return db.DB()
}

