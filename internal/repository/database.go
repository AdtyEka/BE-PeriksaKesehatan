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

const (
	// Connection pool settings
	maxOpenConns    = 25
	maxIdleConns    = 5
	connMaxLifetime = 5 * time.Minute
	connMaxIdleTime = 10 * time.Minute

	// Schema settings
	defaultSchema = "public"
)

// InitDB menginisialisasi koneksi database dengan konfigurasi untuk production.
// PreferSimpleProtocol: true mengatasi error "prepared statement already exists" di Supabase/PostgreSQL.
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := openDatabaseConnection(cfg.DBURL)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan instance database: %w", err)
	}

	if err := setupConnectionPool(sqlDB); err != nil {
		return nil, fmt.Errorf("gagal setup connection pool: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("gagal melakukan ping ke database: %w", err)
	}

	log.Println("[DB] Koneksi database berhasil dibuat")

	// Jalankan migrations dan seeds
	if err := runMigrations(db); err != nil {
		log.Printf("[DB] Warning: Beberapa migration gagal: %v", err)
		// Tidak return error, karena beberapa migration bisa non-critical
	}

	if err := runSeeds(db); err != nil {
		log.Printf("[DB] Warning: Beberapa seed gagal: %v", err)
		// Tidak return error, karena seed bisa non-critical
	}

	return db, nil
}

// openDatabaseConnection membuka koneksi GORM ke PostgreSQL
func openDatabaseConnection(dbURL string) (*gorm.DB, error) {
	postgresConfig := postgres.Config{
		DSN:                  dbURL,
		PreferSimpleProtocol: true, // Mengatasi error "prepared statement already exists" di Supabase
	}

	gormConfig := &gorm.Config{
		PrepareStmt: false, // Disable prepared statements untuk Supabase compatibility
		Logger:      logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.New(postgresConfig), gormConfig)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// setupConnectionPool mengatur connection pool untuk database
func setupConnectionPool(sqlDB *sql.DB) error {
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)
	return nil
}

// GetDBConnection mengembalikan instance *sql.DB dari GORM DB
func GetDBConnection(db *gorm.DB) (*sql.DB, error) {
	return db.DB()
}

// runMigrations menjalankan semua database migrations
func runMigrations(db *gorm.DB) error {
	var migrationErrors []error

	// AutoMigrate untuk create tables dan add columns
	if err := autoMigrateTables(db); err != nil {
		migrationErrors = append(migrationErrors, fmt.Errorf("auto-migrate: %w", err))
	}

	// Manual migrations untuk alter constraints
	if err := migrateHealthDataNullable(db); err != nil {
		migrationErrors = append(migrationErrors, fmt.Errorf("migrate health_data nullable: %w", err))
	}

	if err := migrateEducationalVideosTable(db); err != nil {
		migrationErrors = append(migrationErrors, fmt.Errorf("migrate educational_videos: %w", err))
	}

	if len(migrationErrors) > 0 {
		return fmt.Errorf("migration errors: %v", migrationErrors)
	}

	return nil
}

// autoMigrateTables menjalankan GORM AutoMigrate untuk semua entities
func autoMigrateTables(db *gorm.DB) error {
	entities := []interface{}{
		&entity.User{},
		&entity.HealthData{},
		&entity.BlacklistedToken{},
		&entity.HealthAlert{},
		&entity.Category{},
		&entity.EducationalVideo{},
		&entity.HealthTarget{},
	}

	if err := db.AutoMigrate(entities...); err != nil {
		return err
	}

	log.Println("[DB] Auto-migrate berhasil")
	return nil
}

// migrateHealthDataNullable mengubah kolom health_data dari NOT NULL ke NULL.
// GORM AutoMigrate tidak mengubah constraint NOT NULL secara otomatis.
// Migration ini idempotent dan aman untuk Supabase/PostgreSQL.
func migrateHealthDataNullable(db *gorm.DB) error {
	migrator := db.Migrator()

	if !migrator.HasTable(&entity.HealthData{}) {
		log.Println("[DB] Tabel health_data belum ada, akan dibuat oleh AutoMigrate")
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
		if err := makeColumnNullable(db, "health_data", columnName); err != nil {
			log.Printf("[DB] Warning: Gagal migrate kolom %s.%s: %v", "health_data", columnName, err)
			// Continue dengan kolom berikutnya, tidak return error
		}
	}

	return nil
}

// makeColumnNullable mengubah kolom menjadi nullable jika belum nullable.
// Fungsi ini idempotent dan aman untuk dipanggil berulang kali.
func makeColumnNullable(db *gorm.DB, tableName, columnName string) error {
	migrator := db.Migrator()

	// Cek apakah kolom ada
	if !migrator.HasColumn(&entity.HealthData{}, columnName) {
		log.Printf("[DB] Kolom %s.%s tidak ditemukan, skip", tableName, columnName)
		return nil
	}

	// Cek apakah kolom sudah nullable
	isNullable, err := checkColumnNullable(db, tableName, columnName)
	if err != nil {
		// Jika gagal cek, coba langsung alter (untuk backward compatibility)
		log.Printf("[DB] Warning: Gagal mengecek nullable untuk %s.%s, mencoba alter langsung", tableName, columnName)
		return alterColumnNullable(db, tableName, columnName, true)
	}

	if isNullable {
		log.Printf("[DB] Kolom %s.%s sudah nullable, skip", tableName, columnName)
		return nil
	}

	return alterColumnNullable(db, tableName, columnName, true)
}

// checkColumnNullable mengecek apakah kolom sudah nullable
func checkColumnNullable(db *gorm.DB, tableName, columnName string) (bool, error) {
	var isNullable string
	query := `
		SELECT is_nullable 
		FROM information_schema.columns 
		WHERE table_schema = $1 
		AND table_name = $2 
		AND column_name = $3
	`

	err := db.Raw(query, defaultSchema, tableName, columnName).Scan(&isNullable).Error
	if err != nil {
		return false, err
	}

	return isNullable == "YES", nil
}

// alterColumnNullable mengubah constraint kolom menjadi nullable atau not null
func alterColumnNullable(db *gorm.DB, tableName, columnName string, nullable bool) error {
	var alterSQL string
	if nullable {
		alterSQL = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", tableName, columnName)
	} else {
		alterSQL = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", tableName, columnName)
	}

	if err := db.Exec(alterSQL).Error; err != nil {
		// Cek jika error karena constraint sudah sesuai
		if isAlreadyNullableError(err) {
			log.Printf("[DB] Kolom %s.%s sudah nullable, skip", tableName, columnName)
			return nil
		}
		return fmt.Errorf("gagal mengubah kolom %s.%s: %w", tableName, columnName, err)
	}

	log.Printf("[DB] Kolom %s.%s berhasil diubah menjadi nullable", tableName, columnName)
	return nil
}

// migrateEducationalVideosTable menambahkan kolom category_id ke tabel educational_videos.
// Migration ini idempotent dan aman untuk Supabase/PostgreSQL.
func migrateEducationalVideosTable(db *gorm.DB) error {
	migrator := db.Migrator()

	if !migrator.HasTable(&entity.EducationalVideo{}) {
		log.Println("[DB] Tabel educational_videos belum ada, akan dibuat oleh AutoMigrate")
		return nil
	}

	if migrator.HasColumn(&entity.EducationalVideo{}, "category_id") {
		log.Println("[DB] Kolom category_id sudah ada di educational_videos, skip")
		return nil
	}

	// Tambahkan kolom category_id (nullable untuk backward compatibility)
	alterSQL := "ALTER TABLE educational_videos ADD COLUMN category_id INTEGER"
	if err := db.Exec(alterSQL).Error; err != nil {
		if isAlreadyExistsError(err) {
			log.Println("[DB] Kolom category_id sudah ada")
			return nil
		}
		return fmt.Errorf("gagal menambahkan kolom category_id: %w", err)
	}

	log.Println("[DB] Kolom category_id berhasil ditambahkan ke educational_videos")

	// Tambahkan index untuk category_id
	if err := createIndexIfNotExists(db, "educational_videos", "category_id", "idx_educational_videos_category_id"); err != nil {
		log.Printf("[DB] Warning: Gagal menambahkan index untuk category_id: %v", err)
		// Tidak return error, karena index bukan critical
	}

	return nil
}

// createIndexIfNotExists membuat index jika belum ada
func createIndexIfNotExists(db *gorm.DB, tableName, columnName, indexName string) error {
	indexSQL := fmt.Sprintf(
		"CREATE INDEX IF NOT EXISTS %s ON %s(%s)",
		indexName, tableName, columnName,
	)

	if err := db.Exec(indexSQL).Error; err != nil {
		return err
	}

	log.Printf("[DB] Index %s berhasil ditambahkan", indexName)
	return nil
}

// runSeeds menjalankan semua database seeds
func runSeeds(db *gorm.DB) error {
	if err := seedDefaultCategories(db); err != nil {
		return fmt.Errorf("seed default categories: %w", err)
	}
	return nil
}

// seedDefaultCategories melakukan seed kategori default jika belum ada.
// Fungsi ini idempotent dan aman untuk dipanggil berulang kali.
func seedDefaultCategories(db *gorm.DB) error {
	defaultCategories := []entity.Category{
		{Kategori: "Diabetes"},
		{Kategori: "Hipertensi"},
		{Kategori: "Jantung"},
	}

	for _, category := range defaultCategories {
		if err := seedCategoryIfNotExists(db, category); err != nil {
			return fmt.Errorf("gagal seed kategori %s: %w", category.Kategori, err)
		}
	}

	log.Println("[DB] Seed default categories berhasil")
	return nil
}

// seedCategoryIfNotExists menambahkan kategori jika belum ada
func seedCategoryIfNotExists(db *gorm.DB, category entity.Category) error {
	var existingCategory entity.Category
	result := db.Where("kategori = ?", category.Kategori).First(&existingCategory)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := db.Create(&category).Error; err != nil {
				return err
			}
			log.Printf("[DB] Kategori %s berhasil dibuat", category.Kategori)
			return nil
		}
		return result.Error
	}

	log.Printf("[DB] Kategori %s sudah ada, skip", category.Kategori)
	return nil
}

// Helper functions untuk error checking

// isAlreadyNullableError mengecek apakah error karena kolom sudah nullable
func isAlreadyNullableError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "does not exist") ||
		strings.Contains(errMsg, "already") ||
		strings.Contains(errMsg, "duplicate")
}

// isAlreadyExistsError mengecek apakah error karena resource sudah ada
func isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "already exists") ||
		strings.Contains(errMsg, "duplicate")
}

