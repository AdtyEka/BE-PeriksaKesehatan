package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/pkg/utils"
	"errors"
)

// ValidateHealthData melakukan validasi range nilai data kesehatan dengan nullable-aware.
// Minimal satu metrik kesehatan harus diisi. Jika systolic dikirim, diastolic juga harus dikirim.
func (s *HealthDataService) ValidateHealthData(req *request.HealthDataRequest) error {
	if err := utils.RequireAtLeastOneHealthMetric(
		req.Systolic, req.Diastolic, req.BloodSugar, nil, req.HeartRate, req.Weight, req.Height,
	); err != nil {
		return err
	}

	if (req.Systolic != nil && req.Diastolic == nil) ||
		(req.Systolic == nil && req.Diastolic != nil) {
		return errors.New("systolic dan diastolic harus dikirim bersamaan")
	}

	if err := utils.ValidateNullableInt(req.Systolic, "systolic", 90, 180); err != nil {
		return err
	}

	if err := utils.ValidateNullableInt(req.Diastolic, "diastolic", 60, 120); err != nil {
		return err
	}

	if err := utils.ValidateNullableInt(req.BloodSugar, "blood_sugar", 60, 300); err != nil {
		return err
	}

	if err := utils.ValidateNullableFloat64(req.Weight, "weight", 20.0, 200.0); err != nil {
		return err
	}

	if err := utils.ValidateNullableInt(req.Height, "height", 50, 250); err != nil {
		return err
	}

	if err := utils.ValidateNullableInt(req.HeartRate, "heart_rate", 40, 180); err != nil {
		return err
	}

	return nil
}

