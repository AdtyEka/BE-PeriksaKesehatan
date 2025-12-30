package repository

import (
	"BE-PeriksaKesehatan/config"
	"BE-PeriksaKesehatan/internal/model/entity"
	"database/sql"
	"errors"
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
		&entity.Category{},
		&entity.EducationalVideo{},
		&entity.HealthTarget{},
	)
	if err != nil {
		log.Printf("Warning: Gagal melakukan auto-migrate: %v", err)
	} else {
		log.Println("Info: Auto-migrate berhasil!")
	}

	// Seed default categories
	if err := seedDefaultCategories(db); err != nil {
		log.Printf("Warning: Gagal melakukan seed default categories: %v", err)
	} else {
		log.Println("Info: Seed default categories berhasil!")
	}

	// Migrate educational_videos table: add category_id column
	if err := migrateEducationalVideosTable(db); err != nil {
		log.Printf("Warning: Gagal melakukan migration educational_videos: %v", err)
	} else {
		log.Println("Info: Migration educational_videos berhasil!")
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

// migrateEducationalVideosTable menambahkan kolom category_id ke tabel educational_videos
// Migration ini idempotent dan aman untuk Supabase/PostgreSQL.
func migrateEducationalVideosTable(db *gorm.DB) error {
	migrator := db.Migrator()

	if !migrator.HasTable(&entity.EducationalVideo{}) {
		log.Println("Info: Tabel educational_videos belum ada, akan dibuat oleh AutoMigrate")
		return nil
	}

	// Cek apakah kolom category_id sudah ada
	if migrator.HasColumn(&entity.EducationalVideo{}, "category_id") {
		log.Println("Info: Kolom category_id sudah ada di tabel educational_videos, skip migration")
		return nil
	}

	log.Println("Info: Menambahkan kolom category_id ke tabel educational_videos...")

	// Tambahkan kolom category_id (nullable dulu untuk backward compatibility)
	alterSQL := `
		ALTER TABLE educational_videos 
		ADD COLUMN category_id INTEGER
	`

	if err := db.Exec(alterSQL).Error; err != nil {
		if contains(err.Error(), "already exists") || contains(err.Error(), "duplicate") {
			log.Println("Info: Kolom category_id sudah ada")
			return nil
		}
		return fmt.Errorf("gagal menambahkan kolom category_id: %v", err)
	}

	log.Println("Info: Kolom category_id berhasil ditambahkan")

	// Tambahkan index untuk category_id
	indexSQL := `
		CREATE INDEX IF NOT EXISTS idx_educational_videos_category_id 
		ON educational_videos(category_id)
	`

	if err := db.Exec(indexSQL).Error; err != nil {
		log.Printf("Warning: Gagal menambahkan index untuk category_id: %v", err)
		// Tidak return error, karena index bukan critical
	} else {
		log.Println("Info: Index untuk category_id berhasil ditambahkan")
	}

	// Catatan: Foreign key constraint bisa ditambahkan nanti setelah memastikan semua data valid
	// Untuk sekarang, kita biarkan nullable dulu agar tidak error jika ada data lama
	log.Println("Info: Kolom category_id ditambahkan (nullable dengan index). Set NOT NULL dan foreign key bisa dilakukan setelah data terisi.")

	return nil
}

// seedDefaultCategories melakukan seed kategori default jika belum ada
func seedDefaultCategories(db *gorm.DB) error {
	defaultCategories := []entity.Category{
		{Kategori: "Diabetes"},
		{Kategori: "Hipertensi"},
		{Kategori: "Jantung"},
	}

	for _, category := range defaultCategories {
		var existingCategory entity.Category
		result := db.Where("kategori = ?", category.Kategori).First(&existingCategory)
		
		// Jika kategori belum ada, insert
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				if err := db.Create(&category).Error; err != nil {
					return fmt.Errorf("gagal membuat kategori %s: %v", category.Kategori, err)
				}
				log.Printf("Info: Kategori %s berhasil dibuat", category.Kategori)
			} else {
				return fmt.Errorf("error saat mengecek kategori %s: %v", category.Kategori, result.Error)
			}
		} else {
			log.Printf("Info: Kategori %s sudah ada, skip", category.Kategori)
		}
	}

	return nil
}

