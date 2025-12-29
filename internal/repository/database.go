package repository

import (
	"BE-PeriksaKesehatan/config"
	"BE-PeriksaKesehatan/internal/model/entity"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB menginisialisasi koneksi database dengan konfigurasi untuk production.
// PreferSimpleProtocol: true mengatasi error "prepared statement already exists" di Supabase/PostgreSQL.
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dbURL := cfg.DBURL

	postgresConfig := postgres.Config{
		DSN:                  dbURL,
		PreferSimpleProtocol: true,
	}

	gormConfig := &gorm.Config{
		PrepareStmt: false,
		Logger:      logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.New(postgresConfig), gormConfig)
	if err != nil {
		log.Printf("Error: Gagal membuka koneksi ke database: %v", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Error: Gagal mendapatkan instance database: %v", err)
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		log.Printf("Error: Gagal melakukan ping ke database: %v", err)
		return nil, err
	}

	err = db.AutoMigrate(
		&entity.User{},
		&entity.HealthData{},
		&entity.BlacklistedToken{},
		&entity.HealthAlert{},
		&entity.EducationalVideo{},
		&entity.HealthTarget{},
	)
	if err != nil {
		log.Printf("Warning: Gagal melakukan auto-migrate: %v", err)
	} else {
		log.Println("Info: Auto-migrate berhasil!")
	}

	if err := migrateHealthDataNullable(db); err != nil {
		log.Printf("Warning: Gagal melakukan migration nullable untuk health_data: %v", err)
	} else {
		log.Println("Info: Migration nullable untuk health_data berhasil!")
	}

	log.Println("Info: Koneksi ke database berhasil dibuat!")
	return db, nil
}

func GetDBConnection(db *gorm.DB) (*sql.DB, error) {
	return db.DB()
}

// migrateHealthDataNullable mengubah kolom health_data dari NOT NULL ke NULL.
// GORM AutoMigrate tidak mengubah constraint NOT NULL secara otomatis.
// Migration ini idempotent dan aman untuk Supabase/PostgreSQL.
func migrateHealthDataNullable(db *gorm.DB) error {
	migrator := db.Migrator()

	if !migrator.HasTable(&entity.HealthData{}) {
		log.Println("Info: Tabel health_data belum ada, akan dibuat oleh AutoMigrate")
		return nil
	}

	columnsToMigrate := []string{
		"systolic",
		"diastolic",
		"blood_sugar",
		"weight",
		"heart_rate",
	}

	for _, columnName := range columnsToMigrate {
		if !migrator.HasColumn(&entity.HealthData{}, columnName) {
			log.Printf("Info: Kolom %s tidak ditemukan, skip migration", columnName)
			continue
		}

		var isNullable string
		err := db.Raw(`
			SELECT is_nullable 
			FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = 'health_data' 
			AND column_name = $1
		`, columnName).Scan(&isNullable).Error

		if err != nil {
			log.Printf("Warning: Gagal mengecek constraint untuk kolom %s: %v", columnName, err)
			log.Printf("Info: Mencoba langsung mengubah kolom %s menjadi nullable...", columnName)
			
			alterSQL := fmt.Sprintf(`
				ALTER TABLE health_data 
				ALTER COLUMN %s DROP NOT NULL
			`, columnName)

			if err := db.Exec(alterSQL).Error; err != nil {
				if contains(err.Error(), "does not exist") || contains(err.Error(), "already") {
					log.Printf("Info: Kolom %s sudah nullable atau constraint tidak ada", columnName)
				} else {
					log.Printf("Warning: Gagal mengubah kolom %s menjadi nullable: %v", columnName, err)
				}
			} else {
				log.Printf("Info: Kolom %s berhasil diubah menjadi nullable", columnName)
			}
			continue
		}

		if isNullable == "NO" {
			alterSQL := fmt.Sprintf(`
				ALTER TABLE health_data 
				ALTER COLUMN %s DROP NOT NULL
			`, columnName)

			if err := db.Exec(alterSQL).Error; err != nil {
				log.Printf("Warning: Gagal mengubah kolom %s menjadi nullable: %v", columnName, err)
				log.Printf("Info: Pastikan user database di Supabase punya permission ALTER TABLE")
				continue
			}

			log.Printf("Info: Kolom %s berhasil diubah menjadi nullable", columnName)
		} else {
			log.Printf("Info: Kolom %s sudah nullable, skip migration", columnName)
		}
	}

	return nil
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

