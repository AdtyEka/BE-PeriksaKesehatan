package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"
	"time"

	"gorm.io/gorm"
)

// HealthDataRepository adalah struct yang menampung koneksi database untuk health data
type HealthDataRepository struct {
	db *gorm.DB
}

// NewHealthDataRepository membuat instance baru dari HealthDataRepository
func NewHealthDataRepository(db *gorm.DB) *HealthDataRepository {
	return &HealthDataRepository{
		db: db,
	}
}

// ==================== OPERASI CREATE (INSERT) ====================

// CreateHealthData melakukan INSERT data kesehatan baru ke database
func (r *HealthDataRepository) CreateHealthData(healthData *entity.HealthData) error {
	result := r.db.Create(healthData)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ==================== OPERASI READ (SELECT) ====================

// GetHealthDataByID melakukan SELECT data kesehatan berdasarkan ID
func (r *HealthDataRepository) GetHealthDataByID(id uint) (*entity.HealthData, error) {
	var healthData entity.HealthData
	result := r.db.First(&healthData, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("data kesehatan tidak ditemukan")
		}
		return nil, result.Error
	}
	return &healthData, nil
}

// GetHealthDataByUserID melakukan SELECT semua data kesehatan berdasarkan UserID
// Berguna untuk melihat riwayat kesehatan user
func (r *HealthDataRepository) GetHealthDataByUserID(userID uint) ([]entity.HealthData, error) {
	var healthDataList []entity.HealthData
	result := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&healthDataList)
	if result.Error != nil {
		return nil, result.Error
	}
	return healthDataList, nil
}

// GetAllHealthData melakukan SELECT semua data kesehatan dari database
func (r *HealthDataRepository) GetAllHealthData() ([]entity.HealthData, error) {
	var healthDataList []entity.HealthData
	result := r.db.Order("created_at DESC").Find(&healthDataList)
	if result.Error != nil {
		return nil, result.Error
	}
	return healthDataList, nil
}

// ==================== OPERASI UNTUK RIWAYAT KESEHATAN ====================

// GetHealthDataByUserIDWithFilter melakukan SELECT data kesehatan dengan filter waktu dan metrik
func (r *HealthDataRepository) GetHealthDataByUserIDWithFilter(userID uint, startDate, endDate time.Time) ([]entity.HealthData, error) {
	var healthDataList []entity.HealthData
	query := r.db.Where("user_id = ?", userID).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Order("created_at DESC")
	
	result := query.Find(&healthDataList)
	if result.Error != nil {
		return nil, result.Error
	}
	return healthDataList, nil
}

// GetHealthDataForComparison mengambil data periode sebelumnya untuk perbandingan
// periodDuration adalah durasi periode sebelumnya (misalnya 7 hari, 30 hari)
func (r *HealthDataRepository) GetHealthDataForComparison(userID uint, startDate, endDate time.Time, periodDuration time.Duration) ([]entity.HealthData, error) {
	// Hitung periode sebelumnya
	periodLength := endDate.Sub(startDate)
	prevEndDate := startDate.Add(-24 * time.Hour) // 1 hari sebelum startDate
	prevStartDate := prevEndDate.Add(-periodLength)
	
	var healthDataList []entity.HealthData
	query := r.db.Where("user_id = ?", userID).
		Where("created_at >= ? AND created_at <= ?", prevStartDate, prevEndDate).
		Order("created_at DESC")
	
	result := query.Find(&healthDataList)
	if result.Error != nil {
		return nil, result.Error
	}
	return healthDataList, nil
}

// GetLatestHealthDataByUserID mengambil 1 data kesehatan terbaru berdasarkan UserID
// Digunakan untuk mengambil berat badan terbaru untuk profile
func (r *HealthDataRepository) GetLatestHealthDataByUserID(userID uint) (*entity.HealthData, error) {
	var healthData entity.HealthData
	result := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&healthData)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Tidak ada data, return nil tanpa error (bukan error, hanya belum ada data)
			return nil, nil
		}
		return nil, result.Error
	}
	return &healthData, nil
}

