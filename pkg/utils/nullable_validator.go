package utils

import (
	"errors"
	"fmt"
)

//   - error jika field tidak nil dan nilainya di luar range
//   - nil jika field nil (tidak perlu validasi) atau nilainya valid
func ValidateNullableInt(field *int, fieldName string, min, max int) error {
	if field == nil {
		return nil // Field tidak dikirim, skip validasi
	}

	value := *field
	if value < min || value > max {
		return fmt.Errorf("%s harus berada dalam range %d-%d", fieldName, min, max)
	}

	return nil
}

//   - error jika field tidak nil dan nilainya di luar range
//   - nil jika field nil (tidak perlu validasi) atau nilainya valid
func ValidateNullableFloat64(field *float64, fieldName string, min, max float64) error {
	if field == nil {
		return nil // Field tidak dikirim, skip validasi
	}

	value := *field
	if value < min || value > max {
		return fmt.Errorf("%s harus berada dalam range %.2f-%.2f", fieldName, min, max)
	}

	return nil
}

// ValidateNullableString memvalidasi string nullable (opsional, untuk Activity)
// Returns:
//   - error jika field tidak nil dan string kosong (jika required)
//   - nil jika field nil atau string valid
func ValidateNullableString(field *string, fieldName string, required bool) error {
	if field == nil {
		if required {
			return fmt.Errorf("%s wajib diisi", fieldName)
		}
		return nil // Field tidak dikirim dan tidak wajib
	}

	value := *field
	if required && value == "" {
		return fmt.Errorf("%s tidak boleh kosong", fieldName)
	}

	return nil
}


// RequireAtLeastOneHealthMetric memastikan minimal satu metrik kesehatan diisi
// Helper khusus untuk health data yang lebih type-safe
func RequireAtLeastOneHealthMetric(
	systolic, diastolic, bloodSugar, weight, heartRate *int,
	weightFloat *float64,
	height *int,
) error {
	hasAtLeastOne := false

	if systolic != nil || diastolic != nil {
		hasAtLeastOne = true
	}
	if bloodSugar != nil {
		hasAtLeastOne = true
	}
	if weight != nil || weightFloat != nil {
		hasAtLeastOne = true
	}
	if height != nil {
		hasAtLeastOne = true
	}
	if heartRate != nil {
		hasAtLeastOne = true
	}

	if !hasAtLeastOne {
		return errors.New("minimal satu metrik kesehatan harus diisi (systolic/diastolic, blood_sugar, weight, height, atau heart_rate)")
	}

	return nil
}


