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
	userRepo         *repository.UserRepository
	healthDataRepo   *repository.HealthDataRepository
	healthTargetRepo *repository.HealthTargetRepository
}

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
	if user.BirthDate != nil {
		calculatedAge := s.calculateAge(*user.BirthDate)
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

func (s *ProfileService) UpdatePersonalInfo(userID uint, req *request.UpdatePersonalInfoRequest) error {
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return errors.New("format tanggal lahir tidak valid, gunakan format YYYY-MM-DD")
	}

	if birthDate.After(time.Now()) {
		return errors.New("tanggal lahir tidak boleh di masa depan")
	}

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

