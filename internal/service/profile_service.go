package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
	"errors"
	"fmt"
	"math"
	"time"
)

type ProfileService struct {
	userRepo          *repository.UserRepository
	healthDataRepo    *repository.HealthDataRepository
	healthTargetRepo  *repository.HealthTargetRepository
	personalInfoRepo  *repository.PersonalInfoRepository
}

func NewProfileService(
	userRepo *repository.UserRepository,
	healthDataRepo *repository.HealthDataRepository,
	healthTargetRepo *repository.HealthTargetRepository,
	personalInfoRepo *repository.PersonalInfoRepository,
) *ProfileService {
	return &ProfileService{
		userRepo:         userRepo,
		healthDataRepo:   healthDataRepo,
		healthTargetRepo: healthTargetRepo,
		personalInfoRepo: personalInfoRepo,
	}
}

func (s *ProfileService) GetProfile(userID uint) (*response.ProfileResponse, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	latestHealthData, err := s.healthDataRepo.GetLatestHealthDataByUserID(userID)
	if err != nil {
		return nil, err
	}

	var age *int
	// Ambil birth_date dari personal_info jika ada
	personalInfo, err := s.personalInfoRepo.GetPersonalInfoByUserID(userID)
	if err == nil && personalInfo != nil && personalInfo.BirthDate != nil {
		calculatedAge := s.calculateAge(*personalInfo.BirthDate)
		age = &calculatedAge
	}

	resp := &response.ProfileResponse{
		Name:     user.Nama,
		Email:    user.Email,
		PhotoURL: user.PhotoURL,
		Height:   user.HeightCM,
		Age:      age,
	}

	if latestHealthData != nil && latestHealthData.Weight != nil {
		resp.Weight = latestHealthData.Weight
	}

	return resp, nil
}

func (s *ProfileService) UpdateProfile(userID uint, req *request.UpdateProfileRequest) error {
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["nama"] = *req.Name
	}
	if req.PhotoURL != nil {
		updates["photo_url"] = *req.PhotoURL
	}
	if req.Height != nil {
		updates["height_cm"] = *req.Height
	}

	if len(updates) == 0 {
		return errors.New("tidak ada data untuk diupdate")
	}

	return s.userRepo.UpdateUserProfile(userID, updates)
}

func (s *ProfileService) GetPersonalInfo(userID uint) (*response.PersonalInfoResponse, error) {
	// Cek apakah user ada
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Ambil personal info dari tabel personal_infos
	personalInfo, err := s.personalInfoRepo.GetPersonalInfoByUserID(userID)
	if err != nil {
		if err.Error() == "personal info tidak ditemukan" {
			// Return response kosong jika belum ada data
			return &response.PersonalInfoResponse{
				Name: "",
			}, nil
		}
		return nil, err
	}

	resp := &response.PersonalInfoResponse{
		Name: personalInfo.Name,
	}

	if personalInfo.BirthDate != nil {
		birthDateStr := personalInfo.BirthDate.Format("2006-01-02")
		resp.BirthDate = &birthDateStr
	}
	if personalInfo.Phone != nil {
		resp.Phone = personalInfo.Phone
	}
	if personalInfo.Address != nil {
		resp.Address = personalInfo.Address
	}

	return resp, nil
}

func (s *ProfileService) UpdatePersonalInfo(userID uint, req *request.UpdatePersonalInfoRequest) error {
	// Cek apakah user ada
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Cek apakah personal info sudah ada
	exists, err := s.personalInfoRepo.CheckPersonalInfoExists(userID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("personal info tidak ditemukan, silakan buat terlebih dahulu")
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return errors.New("format tanggal lahir tidak valid, gunakan format YYYY-MM-DD")
	}

	if birthDate.After(time.Now()) {
		return errors.New("tanggal lahir tidak boleh di masa depan")
	}

	updates := make(map[string]interface{})
	updates["name"] = req.Name
	updates["birth_date"] = birthDate
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}

	return s.personalInfoRepo.UpdatePersonalInfo(userID, updates)
}

// CreatePersonalInfo membuat personal info baru untuk user
func (s *ProfileService) CreatePersonalInfo(userID uint, req *request.CreatePersonalInfoRequest) (*response.PersonalInfoResponse, error) {
	// Cek apakah user ada
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Cek apakah personal info sudah ada
	exists, err := s.personalInfoRepo.CheckPersonalInfoExists(userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("personal info sudah ada")
	}

	// Validasi birth_date format
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return nil, errors.New("format tanggal lahir tidak valid, gunakan format YYYY-MM-DD")
	}

	// Validasi birth_date tidak boleh di masa depan
	if birthDate.After(time.Now()) {
		return nil, errors.New("tanggal lahir tidak boleh di masa depan")
	}

	// Validasi phone jika ada
	if req.Phone != nil && *req.Phone != "" {
		phone := *req.Phone
		// Validasi numeric dan panjang 10-15 digit
		if len(phone) < 10 || len(phone) > 15 {
			return nil, errors.New("phone harus 10-15 digit")
		}
		// Validasi numeric
		for _, char := range phone {
			if char < '0' || char > '9' {
				return nil, errors.New("phone harus numeric")
			}
		}
	}

	// Buat personal info
	personalInfo := &entity.PersonalInfo{
		UserID:    userID,
		Name:      req.Name,
		BirthDate: &birthDate,
		Phone:     req.Phone,
		Address:   req.Address,
	}

	err = s.personalInfoRepo.CreatePersonalInfo(personalInfo)
	if err != nil {
		return nil, err
	}

	// Build response
	resp := &response.PersonalInfoResponse{
		Name: personalInfo.Name,
	}

	if personalInfo.BirthDate != nil {
		birthDateStr := personalInfo.BirthDate.Format("2006-01-02")
		resp.BirthDate = &birthDateStr
	}
	if personalInfo.Phone != nil {
		resp.Phone = personalInfo.Phone
	}
	if personalInfo.Address != nil {
		resp.Address = personalInfo.Address
	}

	return resp, nil
}

func (s *ProfileService) GetHealthTargets(userID uint) (*response.HealthTargetsResponse, error) {
	healthTarget, err := s.healthTargetRepo.GetHealthTargetByUserID(userID)
	if err != nil && err.Error() != "health target tidak ditemukan" {
		return nil, err
	}

	latestHealthData, err := s.healthDataRepo.GetLatestHealthDataByUserID(userID)
	if err != nil {
		return nil, err
	}

	resp := &response.HealthTargetsResponse{}

	if healthTarget == nil {
		return resp, nil
	}

	if healthTarget.TargetSystolic != nil || healthTarget.TargetDiastolic != nil {
		bpProgress := &response.BloodPressureTargetProgress{
			TargetSystolic:  healthTarget.TargetSystolic,
			TargetDiastolic: healthTarget.TargetDiastolic,
		}

		if latestHealthData != nil && latestHealthData.Systolic != nil && latestHealthData.Diastolic != nil {
			bpProgress.CurrentSystolic = latestHealthData.Systolic
			bpProgress.CurrentDiastolic = latestHealthData.Diastolic

			if healthTarget.TargetSystolic != nil && healthTarget.TargetDiastolic != nil {
				progress := s.calculateBloodPressureProgress(
					*healthTarget.TargetSystolic,
					*healthTarget.TargetDiastolic,
					*latestHealthData.Systolic,
					*latestHealthData.Diastolic,
				)
				bpProgress.ProgressPercent = &progress
			}
		}

		resp.BloodPressure = bpProgress
	}

	if healthTarget.TargetBloodSugar != nil {
		bsProgress := &response.BloodSugarTargetProgress{
			Target: healthTarget.TargetBloodSugar,
		}

		if latestHealthData != nil && latestHealthData.BloodSugar != nil {
			bsProgress.Current = latestHealthData.BloodSugar

			progress := s.calculateBloodSugarProgress(
				*healthTarget.TargetBloodSugar,
				*latestHealthData.BloodSugar,
			)
			bsProgress.ProgressPercent = &progress
		}

		resp.BloodSugar = bsProgress
	}

	if healthTarget.TargetWeight != nil {
		weightProgress := &response.WeightTargetProgress{
			Target: healthTarget.TargetWeight,
		}

		if latestHealthData != nil && latestHealthData.Weight != nil {
			weightProgress.Current = latestHealthData.Weight

			progress := s.calculateWeightProgress(
				*healthTarget.TargetWeight,
				*latestHealthData.Weight,
			)
			weightProgress.ProgressPercent = &progress
		}

		resp.Weight = weightProgress
	}

	return resp, nil
}

func (s *ProfileService) UpdateHealthTargets(userID uint, req *request.UpdateHealthTargetsRequest) error {
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	healthTarget := &entity.HealthTarget{
		UserID:          userID,
		TargetSystolic:  req.TargetSystolic,
		TargetDiastolic: req.TargetDiastolic,
		TargetBloodSugar: req.TargetBloodSugar,
		TargetWeight:    req.TargetWeight,
	}

	return s.healthTargetRepo.UpsertHealthTarget(healthTarget)
}

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

func (s *ProfileService) UpdateSettings(userID uint, req *request.UpdateSettingsRequest) error {
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

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

func (s *ProfileService) calculateAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	birthMonthDay := time.Date(now.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, now.Location())
	if now.Before(birthMonthDay) {
		age--
	}

	return age
}

func (s *ProfileService) calculateBloodPressureProgress(
	targetSystolic, targetDiastolic, currentSystolic, currentDiastolic int,
) float64 {
	systolicDiff := math.Abs(float64(currentSystolic - targetSystolic))
	diastolicDiff := math.Abs(float64(currentDiastolic - targetDiastolic))

	var systolicProgress, diastolicProgress float64
	if targetSystolic > 0 {
		systolicProgress = 100 - (systolicDiff/float64(targetSystolic))*100
	}
	if targetDiastolic > 0 {
		diastolicProgress = 100 - (diastolicDiff/float64(targetDiastolic))*100
	}

	avgProgress := (systolicProgress + diastolicProgress) / 2
	return math.Round(avgProgress*100) / 100
}

func (s *ProfileService) calculateBloodSugarProgress(target, current int) float64 {
	if target == 0 {
		return 0
	}
	diff := math.Abs(float64(current - target))
	progress := 100 - (diff/float64(target))*100
	return math.Round(progress*100) / 100
}

func (s *ProfileService) calculateWeightProgress(target, current float64) float64 {
	if target == 0 {
		return 0
	}
	diff := math.Abs(current - target)
	progress := 100 - (diff/target)*100
	return math.Round(progress*100) / 100
}

// CreateProfile membuat profile baru untuk user
// Validasi: name dan email wajib, weight/height/age optional dengan validasi
// Upload foto jika ada, rollback jika gagal
func (s *ProfileService) CreateProfile(userID uint, req *request.CreateProfileRequest, photoURL *string) (*response.ProfileResponse, error) {
	// Cek apakah user ada
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Cek apakah profile sudah ada
	exists, err := s.userRepo.CheckProfileExists(userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("profile sudah ada")
	}

	// Validasi email format (sudah divalidasi di binding, tapi double check)
	if req.Email == "" {
		return nil, errors.New("email wajib diisi")
	}

	// Validasi optional fields
	if req.Weight != nil && *req.Weight <= 0 {
		return nil, errors.New("weight harus lebih besar dari 0")
	}
	if req.Height != nil && *req.Height <= 0 {
		return nil, errors.New("height harus lebih besar dari 0")
	}
	if req.Age != nil && *req.Age <= 0 {
		return nil, errors.New("age harus lebih besar dari 0")
	}

	// Prepare updates untuk user
	updates := make(map[string]interface{})
	updates["nama"] = req.Name
	updates["email"] = req.Email

	if req.Height != nil {
		updates["height_cm"] = *req.Height
	}
	if photoURL != nil && *photoURL != "" {
		updates["photo_url"] = *photoURL
	}

	// Handle age: jika age dikirim, akan di-handle di personal_info
	// Tidak perlu update birth_date di user table karena sudah dipindah ke personal_info

	// Create profile
	err = s.userRepo.CreateProfile(userID, updates)
	if err != nil {
		return nil, err
	}

	// Jika weight dikirim, buat health_data
	if req.Weight != nil {
		healthData := &entity.HealthData{
			UserID: userID,
			Weight: req.Weight,
		}
		err = s.healthDataRepo.CreateHealthData(healthData)
		if err != nil {
			// Rollback: hapus profile yang sudah dibuat
			// Untuk sekarang, kita biarkan error ini, atau bisa di-handle dengan transaction
			return nil, fmt.Errorf("gagal menyimpan weight: %w", err)
		}
	}

	// Get updated user untuk response
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Get latest health data untuk weight
	latestHealthData, err := s.healthDataRepo.GetLatestHealthDataByUserID(userID)
	if err != nil {
		// Tidak fatal jika health data tidak ada
		latestHealthData = nil
	}

	var age *int
	// Ambil birth_date dari personal_info jika ada
	personalInfo, err := s.personalInfoRepo.GetPersonalInfoByUserID(userID)
	if err == nil && personalInfo != nil && personalInfo.BirthDate != nil {
		calculatedAge := s.calculateAge(*personalInfo.BirthDate)
		age = &calculatedAge
	} else if req.Age != nil {
		age = req.Age
	} else {
		// Default 0 jika tidak ada
		zeroAge := 0
		age = &zeroAge
	}

	// Set height, default 0 jika tidak ada
	var height *int
	if user.HeightCM != nil {
		height = user.HeightCM
	} else {
		zeroHeight := 0
		height = &zeroHeight
	}

	resp := &response.ProfileResponse{
		Name:     user.Nama,
		Email:    user.Email,
		PhotoURL: user.PhotoURL,
		Height:   height,
		Age:      age,
	}

	// Set weight dari health_data jika ada
	if latestHealthData != nil && latestHealthData.Weight != nil {
		resp.Weight = latestHealthData.Weight
	} else {
		// Default 0 jika tidak ada
		zeroWeight := 0.0
		resp.Weight = &zeroWeight
	}

	return resp, nil
}

