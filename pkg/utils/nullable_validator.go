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

// RequireAtLeastOneField memastikan minimal satu field dari slice field nullable tidak nil
// Berguna untuk memastikan request tidak kosong semua
func RequireAtLeastOneField(fieldNames []string, fields ...*interface{}) error {
	if len(fieldNames) != len(fields) {
		return errors.New("jumlah fieldNames harus sama dengan jumlah fields")
	}

	hasAtLeastOne := false
	for i, field := range fields {
		if field != nil && *field != nil {
			hasAtLeastOne = true
			break
		}
		// Skip check untuk field yang tidak relevan
		_ = fieldNames[i]
	}

	if !hasAtLeastOne {
		return errors.New("minimal satu field kesehatan harus diisi")
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

// SafeDerefInt safely dereference int pointer, return default jika nil
// Hanya digunakan untuk kasus di mana kita yakin field harus ada setelah validasi
// JANGAN gunakan untuk validasi - gunakan ValidateNullableInt
func SafeDerefInt(ptr *int, defaultValue int) int {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// SafeDerefFloat64 safely dereference float64 pointer, return default jika nil
// Hanya digunakan untuk kasus di mana kita yakin field harus ada setelah validasi
// JANGAN gunakan untuk validasi - gunakan ValidateNullableFloat64
func SafeDerefFloat64(ptr *float64, defaultValue float64) float64 {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// SafeDerefString safely dereference string pointer, return default jika nil
func SafeDerefString(ptr *string, defaultValue string) string {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// IsFieldProvided mengecek apakah field benar-benar dikirim (bukan nil)
// Berguna untuk membedakan antara "field tidak dikirim" vs "field dikirim null"
func IsFieldProvided[T any](field *T) bool {
	return field != nil
}

// MustHaveValue memastikan field tidak nil, return error jika nil
// Digunakan untuk field yang wajib setelah validasi business logic
func MustHaveValue[T any](field *T, fieldName string) error {
	if field == nil {
		return fmt.Errorf("%s wajib diisi", fieldName)
	}
	return nil
}

