package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
	"errors"
	"math"
	"time"
)

// ProfileService menangani business logic untuk profile
type ProfileService struct {
	userRepo         *repository.UserRepository
	healthDataRepo   *repository.HealthDataRepository
	healthTargetRepo *repository.HealthTargetRepository
}

// NewProfileService membuat instance baru dari ProfileService
func NewProfileService(
	userRepo *repository.UserRepository,
	healthDataRepo *repository.HealthDataRepository,
	healthTargetRepo *repository.HealthTargetRepository,
) *ProfileService {
	return &ProfileService{
		userRepo:         userRepo,
		healthDataRepo:   healthDataRepo,
		healthTargetRepo: healthTargetRepo,
	}
}

// ==================== GET PROFILE ====================

// GetProfile mengambil data profil user lengkap dengan berat badan dari health_data terbaru
func (s *ProfileService) GetProfile(userID uint) (*response.ProfileResponse, error) {
	// Ambil user
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Ambil berat badan terbaru dari health_data
	latestHealthData, err := s.healthDataRepo.GetLatestHealthDataByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Hitung umur dari birth_date
	var age *int
	if user.BirthDate != nil {
		calculatedAge := s.calculateAge(*user.BirthDate)
		age = &calculatedAge
	}

	// Buat response
	resp := &response.ProfileResponse{
		Name:     user.Nama,
		Email:    user.Email,
		PhotoURL: user.PhotoURL,
		Height:   user.HeightCM,
		Age:      age,
	}

	// Set berat badan dari health_data terbaru (bisa null jika belum ada data)
	if latestHealthData != nil {
		resp.Weight = &latestHealthData.Weight
	}

	return resp, nil
}

// ==================== UPDATE PROFILE ====================

// UpdateProfile melakukan update profil dasar (nama, photo_url, height)
func (s *ProfileService) UpdateProfile(userID uint, req *request.UpdateProfileRequest) error {
	// Validasi user exists
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Buat map untuk update
	// Catatan: GORM akan otomatis convert camelCase ke snake_case
	// HeightCM -> height_c_m (default), tapi kita bisa override dengan nama kolom eksplisit
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["nama"] = *req.Name
	}
	if req.PhotoURL != nil {
		updates["photo_url"] = *req.PhotoURL
	}
	if req.Height != nil {
		// Gunakan nama field dari struct, GORM akan handle konversinya
		// Atau gunakan nama kolom eksplisit jika berbeda
		updates["height_cm"] = *req.Height
	}

	if len(updates) == 0 {
		return errors.New("tidak ada data untuk diupdate")
	}

	return s.userRepo.UpdateUserProfile(userID, updates)
}

// ==================== PERSONAL INFO ====================

// GetPersonalInfo mengambil informasi pribadi user
func (s *ProfileService) GetPersonalInfo(userID uint) (*response.PersonalInfoResponse, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	resp := &response.PersonalInfoResponse{
		Name: user.Nama,
	}

	if user.BirthDate != nil {
		birthDateStr := user.BirthDate.Format("2006-01-02")
		resp.BirthDate = &birthDateStr
	}
	if user.Phone != nil {
		resp.Phone = user.Phone
	}
	if user.Address != nil {
		resp.Address = user.Address
	}

	return resp, nil
}

// UpdatePersonalInfo melakukan update informasi pribadi
func (s *ProfileService) UpdatePersonalInfo(userID uint, req *request.UpdatePersonalInfoRequest) error {
	// Validasi user exists
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Parse birth_date
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return errors.New("format tanggal lahir tidak valid, gunakan format YYYY-MM-DD")
	}

	// Validasi tanggal lahir tidak boleh di masa depan
	if birthDate.After(time.Now()) {
		return errors.New("tanggal lahir tidak boleh di masa depan")
	}

	// Buat map untuk update
	updates := make(map[string]interface{})
	updates["nama"] = req.Name
	updates["birth_date"] = birthDate
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}

	return s.userRepo.UpdateUserPersonalInfo(userID, updates)
}

// ==================== HEALTH TARGETS ====================

// GetHealthTargets mengambil target kesehatan dengan progress
func (s *ProfileService) GetHealthTargets(userID uint) (*response.HealthTargetsResponse, error) {
	// Ambil health target
	healthTarget, err := s.healthTargetRepo.GetHealthTargetByUserID(userID)
	if err != nil && err.Error() != "health target tidak ditemukan" {
		return nil, err
	}

	// Ambil data kesehatan terbaru untuk progress
	latestHealthData, err := s.healthDataRepo.GetLatestHealthDataByUserID(userID)
	if err != nil {
		return nil, err
	}

	resp := &response.HealthTargetsResponse{}

	// Jika tidak ada health target, return response kosong
	if healthTarget == nil {
		return resp, nil
	}

	// Tekanan darah target & progress
	if healthTarget.TargetSystolic != nil || healthTarget.TargetDiastolic != nil {
		bpProgress := &response.BloodPressureTargetProgress{
			TargetSystolic:  healthTarget.TargetSystolic,
			TargetDiastolic: healthTarget.TargetDiastolic,
		}

		if latestHealthData != nil {
			bpProgress.CurrentSystolic = &latestHealthData.Systolic
			bpProgress.CurrentDiastolic = &latestHealthData.Diastolic

			// Hitung progress (persentase mendekati target)
			if healthTarget.TargetSystolic != nil && healthTarget.TargetDiastolic != nil {
				progress := s.calculateBloodPressureProgress(
					*healthTarget.TargetSystolic,
					*healthTarget.TargetDiastolic,
					latestHealthData.Systolic,
					latestHealthData.Diastolic,
				)
				bpProgress.ProgressPercent = &progress
			}
		}

		resp.BloodPressure = bpProgress
	}

	// Gula darah target & progress
	if healthTarget.TargetBloodSugar != nil {
		bsProgress := &response.BloodSugarTargetProgress{
			Target: healthTarget.TargetBloodSugar,
		}

		if latestHealthData != nil {
			bsProgress.Current = &latestHealthData.BloodSugar

			// Hitung progress
			progress := s.calculateBloodSugarProgress(
				*healthTarget.TargetBloodSugar,
				latestHealthData.BloodSugar,
			)
			bsProgress.ProgressPercent = &progress
		}

		resp.BloodSugar = bsProgress
	}

	// Berat badan target & progress
	if healthTarget.TargetWeight != nil {
		weightProgress := &response.WeightTargetProgress{
			Target: healthTarget.TargetWeight,
		}

		if latestHealthData != nil {
			weightProgress.Current = &latestHealthData.Weight

			// Hitung progress
			progress := s.calculateWeightProgress(
				*healthTarget.TargetWeight,
				latestHealthData.Weight,
			)
			weightProgress.ProgressPercent = &progress
		}

		resp.Weight = weightProgress
	}

	return resp, nil
}

// UpdateHealthTargets melakukan update target kesehatan
func (s *ProfileService) UpdateHealthTargets(userID uint, req *request.UpdateHealthTargetsRequest) error {
	// Validasi user exists
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Buat atau update health target
	healthTarget := &entity.HealthTarget{
		UserID:          userID,
		TargetSystolic:  req.TargetSystolic,
		TargetDiastolic: req.TargetDiastolic,
		TargetBloodSugar: req.TargetBloodSugar,
		TargetWeight:    req.TargetWeight,
	}

	return s.healthTargetRepo.UpsertHealthTarget(healthTarget)
}

// ==================== SETTINGS ====================

// GetSettings mengambil pengaturan aplikasi
func (s *ProfileService) GetSettings(userID uint) (*response.SettingsResponse, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	resp := &response.SettingsResponse{
		NotificationEnabled: user.NotificationEnabled,
		Language:            user.Language,
		PrivacyInfo:        "Data Anda dilindungi dan hanya digunakan untuk keperluan aplikasi kesehatan.",
		SupportInfo:        "Hubungi support@periksakesehatan.com untuk bantuan.",
		AppVersion:         "1.0.0",
	}

	return resp, nil
}

// UpdateSettings melakukan update pengaturan aplikasi
func (s *ProfileService) UpdateSettings(userID uint, req *request.UpdateSettingsRequest) error {
	// Validasi user exists
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Buat map untuk update
	updates := make(map[string]interface{})
	if req.NotificationEnabled != nil {
		updates["notification_enabled"] = *req.NotificationEnabled
	}
	if req.Language != nil {
		updates["language"] = *req.Language
	}

	if len(updates) == 0 {
		return errors.New("tidak ada data untuk diupdate")
	}

	return s.userRepo.UpdateUserSettings(userID, updates)
}

// ==================== HELPER FUNCTIONS ====================

// calculateAge menghitung umur dari tanggal lahir
func (s *ProfileService) calculateAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	// Cek apakah ulang tahun sudah lewat tahun ini
	// Bandingkan bulan dan hari (bukan YearDay karena bisa berbeda tahun)
	birthMonthDay := time.Date(now.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, now.Location())
	if now.Before(birthMonthDay) {
		age--
	}

	return age
}

// calculateBloodPressureProgress menghitung progress tekanan darah (persentase mendekati target)
// Progress dihitung sebagai: 100 - (selisih absolut / target * 100)
// Nilai bisa negatif jika jauh dari target
func (s *ProfileService) calculateBloodPressureProgress(
	targetSystolic, targetDiastolic, currentSystolic, currentDiastolic int,
) float64 {
	// Hitung rata-rata progress untuk systolic dan diastolic
	systolicDiff := math.Abs(float64(currentSystolic - targetSystolic))
	diastolicDiff := math.Abs(float64(currentDiastolic - targetDiastolic))

	// Hindari division by zero (target sudah divalidasi di request, tapi tetap aman)
	var systolicProgress, diastolicProgress float64
	if targetSystolic > 0 {
		systolicProgress = 100 - (systolicDiff/float64(targetSystolic))*100
	} else {
		systolicProgress = 0
	}
	if targetDiastolic > 0 {
		diastolicProgress = 100 - (diastolicDiff/float64(targetDiastolic))*100
	} else {
		diastolicProgress = 0
	}

	// Rata-rata dari kedua progress
	avgProgress := (systolicProgress + diastolicProgress) / 2

	return math.Round(avgProgress*100) / 100
}

// calculateBloodSugarProgress menghitung progress gula darah
func (s *ProfileService) calculateBloodSugarProgress(target, current int) float64 {
	diff := math.Abs(float64(current - target))
	// Hindari division by zero
	if target == 0 {
		return 0
	}
	progress := 100 - (diff/float64(target))*100
	return math.Round(progress*100) / 100
}

// calculateWeightProgress menghitung progress berat badan
func (s *ProfileService) calculateWeightProgress(target, current float64) float64 {
	diff := math.Abs(current - target)
	// Hindari division by zero
	if target == 0 {
		return 0
	}
	progress := 100 - (diff/target)*100
	return math.Round(progress*100) / 100
}

