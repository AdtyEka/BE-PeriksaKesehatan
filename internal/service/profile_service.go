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
	var photoURL *string
	// Ambil birth_date dari personal_info jika ada
	personalInfo, err := s.personalInfoRepo.GetPersonalInfoByUserID(userID)
	if err == nil && personalInfo != nil {
		if personalInfo.BirthDate != nil {
			calculatedAge := s.calculateAge(*personalInfo.BirthDate)
			age = &calculatedAge
		}
		if personalInfo.PhotoURL != nil {
			photoURL = personalInfo.PhotoURL
		}
	}

	resp := &response.ProfileResponse{
		Name:     user.Nama,
		Email:    user.Email,
		PhotoURL: photoURL,
		Age:      age,
	}

	// Set weight dan height dari health_data jika ada
	if latestHealthData != nil {
		if latestHealthData.Weight != nil {
			resp.Weight = latestHealthData.Weight
		} else {
			// Default 0 jika tidak ada
			zeroWeight := 0.0
			resp.Weight = &zeroWeight
		}
		if latestHealthData.HeightCM != nil {
			resp.Height = latestHealthData.HeightCM
		} else {
			// Default 0 jika tidak ada
			zeroHeight := 0
			resp.Height = &zeroHeight
		}
	} else {
		// Default 0 jika tidak ada health_data
		zeroWeight := 0.0
		zeroHeight := 0
		resp.Weight = &zeroWeight
		resp.Height = &zeroHeight
	}

	return resp, nil
}

func (s *ProfileService) UpdateProfile(userID uint, req *request.UpdateProfileRequest) error {
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Update user profile (nama)
	userUpdates := make(map[string]interface{})
	if req.Name != nil {
		userUpdates["nama"] = *req.Name
	}
	if len(userUpdates) > 0 {
		err = s.userRepo.UpdateUserProfile(userID, userUpdates)
		if err != nil {
			return err
		}
	}

	// Update personal info (photo_url)
	personalInfoUpdates := make(map[string]interface{})
	if req.PhotoURL != nil {
		personalInfoUpdates["photo_url"] = *req.PhotoURL
	}
	if len(personalInfoUpdates) > 0 {
		err = s.personalInfoRepo.UpdatePersonalInfo(userID, personalInfoUpdates)
		if err != nil {
			return err
		}
	}

	// Update height di health_data
	if req.Height != nil {
		latestHealthData, err := s.healthDataRepo.GetLatestHealthDataByUserID(userID)
		if err != nil {
			return err
		}

		if latestHealthData != nil {
			// Update health_data yang sudah ada
			latestHealthData.HeightCM = req.Height
			err = s.healthDataRepo.UpdateHealthData(latestHealthData)
			if err != nil {
				return err
			}
		} else {
			// Buat health_data baru jika belum ada
			healthData := &entity.HealthData{
				UserID:   userID,
				HeightCM: req.Height,
			}
			err = s.healthDataRepo.CreateHealthData(healthData)
			if err != nil {
				return err
			}
		}
	}

	// Validasi bahwa ada data untuk diupdate
	if len(userUpdates) == 0 && len(personalInfoUpdates) == 0 && req.Height == nil {
		return errors.New("tidak ada data untuk diupdate")
	}

	return nil
}

// UpdateProfileWithMultipart mengupdate profil dengan support form-data dan file upload
func (s *ProfileService) UpdateProfileWithMultipart(userID uint, req *request.UpdateProfileMultipartRequest, photoURL *string) error {
	// Cek apakah user ada
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Update user profile (nama)
	userUpdates := make(map[string]interface{})
	if req.Name != nil && *req.Name != "" {
		userUpdates["nama"] = *req.Name
	}
	if len(userUpdates) > 0 {
		err = s.userRepo.UpdateUserProfile(userID, userUpdates)
		if err != nil {
			return err
		}
	}

	// Update personal info (birth_date, phone, address, photo_url)
	personalInfoUpdates := make(map[string]interface{})
	
	// Update birth_date jika dikirim
	if req.BirthDate != nil && *req.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			return errors.New("format tanggal lahir tidak valid, gunakan format YYYY-MM-DD")
		}

		if birthDate.After(time.Now()) {
			return errors.New("tanggal lahir tidak boleh di masa depan")
		}

		personalInfoUpdates["birth_date"] = birthDate
	}

	// Update phone jika dikirim
	if req.Phone != nil && *req.Phone != "" {
		personalInfoUpdates["phone"] = *req.Phone
	}

	// Update address jika dikirim
	if req.Address != nil && *req.Address != "" {
		personalInfoUpdates["address"] = *req.Address
	}

	// Update photoURL jika ada file baru yang diupload atau URL yang dikirim
	if photoURL != nil {
		personalInfoUpdates["photo_url"] = *photoURL
	} else if req.PhotoURL != nil && *req.PhotoURL != "" {
		// Jika tidak ada file upload tapi ada photo_url di form
		personalInfoUpdates["photo_url"] = *req.PhotoURL
	}

	// Update personal info jika ada perubahan
	if len(personalInfoUpdates) > 0 {
		// Cek apakah personal info sudah ada, jika belum buat baru
		exists, err := s.personalInfoRepo.CheckPersonalInfoExists(userID)
		if err != nil {
			return err
		}

		if !exists {
			// Ambil user untuk mendapatkan name
			user, err := s.userRepo.GetUserByID(userID)
			if err != nil {
				return err
			}

			// Buat personal info baru
			personalInfo := &entity.PersonalInfo{
				UserID:   userID,
				Name:     user.Nama,
			}

			// Set field yang dikirim
			if birthDate, ok := personalInfoUpdates["birth_date"].(time.Time); ok {
				personalInfo.BirthDate = &birthDate
			}
			if phone, ok := personalInfoUpdates["phone"].(string); ok {
				personalInfo.Phone = &phone
			}
			if address, ok := personalInfoUpdates["address"].(string); ok {
				personalInfo.Address = &address
			}
			if photoURL, ok := personalInfoUpdates["photo_url"].(string); ok {
				personalInfo.PhotoURL = &photoURL
			}

			err = s.personalInfoRepo.CreatePersonalInfo(personalInfo)
			if err != nil {
				return err
			}
		} else {
			// Update personal info yang sudah ada
			err = s.personalInfoRepo.UpdatePersonalInfo(userID, personalInfoUpdates)
			if err != nil {
				return err
			}
		}
	}

	// Validasi bahwa ada data untuk diupdate
	if len(userUpdates) == 0 && len(personalInfoUpdates) == 0 {
		return errors.New("tidak ada data untuk diupdate")
	}

	return nil
}

func (s *ProfileService) GetPersonalInfo(userID uint) (*response.PersonalInfoResponse, error) {
	// Cek apakah user ada dan ambil data user untuk mendapatkan name
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Ambil personal info dari tabel personal_infos
	personalInfo, err := s.personalInfoRepo.GetPersonalInfoByUserID(userID)
	if err != nil {
		if err.Error() == "personal info tidak ditemukan" {
			// Return response dengan name dari user jika personal info belum ada
			return &response.PersonalInfoResponse{
				Name: user.Nama, // Ambil name dari user, bukan dari personal_info
			}, nil
		}
		return nil, err
	}

	// Name diambil dari user.Nama (data dari register/auth), bukan dari personal_info
	resp := &response.PersonalInfoResponse{
		Name: user.Nama, // Ambil name dari user, bukan dari personalInfo.Name
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
	if personalInfo.PhotoURL != nil {
		resp.PhotoURL = personalInfo.PhotoURL
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

	updates := make(map[string]interface{})

	// Update name jika dikirim
	if req.Name != nil && *req.Name != "" {
		updates["name"] = *req.Name
	}

	// Update birth_date jika dikirim
	if req.BirthDate != nil && *req.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			return errors.New("format tanggal lahir tidak valid, gunakan format YYYY-MM-DD")
		}

		if birthDate.After(time.Now()) {
			return errors.New("tanggal lahir tidak boleh di masa depan")
		}

		updates["birth_date"] = birthDate
	}

	// Update phone jika dikirim
	if req.Phone != nil && *req.Phone != "" {
		updates["phone"] = *req.Phone
	}

	// Update address jika dikirim
	if req.Address != nil && *req.Address != "" {
		updates["address"] = *req.Address
	}

	// Jika tidak ada field yang akan diupdate, return nil (no-op, tidak error)
	if len(updates) == 0 {
		return nil
	}

	return s.personalInfoRepo.UpdatePersonalInfo(userID, updates)
}

// UpdatePersonalInfoWithPhoto mengupdate personal info dengan support upload foto
// photoURL adalah path file yang sudah diupload (jika ada)
// Jika update gagal, rollback file sudah di-handle di handler
func (s *ProfileService) UpdatePersonalInfoWithPhoto(userID uint, req *request.UpdatePersonalInfoMultipartRequest, photoURL *string) error {
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

	updates := make(map[string]interface{})

	// Update name jika dikirim
	if req.Name != nil && *req.Name != "" {
		updates["name"] = *req.Name
	}

	// Update birth_date jika dikirim
	if req.BirthDate != nil && *req.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			return errors.New("format tanggal lahir tidak valid, gunakan format YYYY-MM-DD")
		}

		if birthDate.After(time.Now()) {
			return errors.New("tanggal lahir tidak boleh di masa depan")
		}

		updates["birth_date"] = birthDate
	}

	// Update phone jika dikirim
	if req.Phone != nil && *req.Phone != "" {
		updates["phone"] = *req.Phone
	}

	// Update address jika dikirim
	if req.Address != nil && *req.Address != "" {
		updates["address"] = *req.Address
	}

	// Update photoURL jika ada file baru
	if photoURL != nil {
		updates["photo_url"] = *photoURL
	}

	// Jika tidak ada field yang akan diupdate, return nil (no-op, tidak error)
	if len(updates) == 0 {
		return nil
	}

	// Update database
	err = s.personalInfoRepo.UpdatePersonalInfo(userID, updates)
	if err != nil {
		return err
	}

	// Note: Hapus foto lama akan dilakukan di handler setelah update berhasil
	// untuk memastikan rollback yang benar jika ada error

	return nil
}

// CreatePersonalInfo membuat personal info baru untuk user
// photoURL adalah path file yang sudah diupload (jika ada)
// Jika insert gagal, rollback file sudah di-handle di handler
// Name akan diambil dari data user (register/auth), bukan dari request
func (s *ProfileService) CreatePersonalInfo(userID uint, req *request.CreatePersonalInfoRequest, photoURL *string) (*response.PersonalInfoResponse, error) {
	// Cek apakah user ada dan ambil data user untuk mendapatkan name
	user, err := s.userRepo.GetUserByID(userID)
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

	// Validasi phone wajib (sudah divalidasi di handler, tapi double check)
	if req.Phone == nil || *req.Phone == "" {
		return nil, errors.New("phone wajib diisi")
	}

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

	// Buat personal info
	// Name diambil dari user.Nama (data dari register/auth), bukan dari request
	personalInfo := &entity.PersonalInfo{
		UserID:    userID,
		Name:      user.Nama, // Ambil name dari user, bukan dari request
		BirthDate: &birthDate,
		Phone:     req.Phone,
		Address:   req.Address,
		PhotoURL:  photoURL,
	}

	err = s.personalInfoRepo.CreatePersonalInfo(personalInfo)
	if err != nil {
		return nil, err
	}

	// Build response
	// Name diambil dari user.Nama untuk konsistensi
	resp := &response.PersonalInfoResponse{
		Name: user.Nama, // Gunakan name dari user, bukan dari personal_info
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
	if personalInfo.PhotoURL != nil {
		resp.PhotoURL = personalInfo.PhotoURL
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

