package utils

// Nullable helper functions untuk operasi nullable yang aman

// IntValue mengembalikan nilai int dari pointer, atau default value jika nil
func IntValue(ptr *int, defaultValue int) int {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// IntPtr mengembalikan pointer ke int, atau nil jika value adalah zero value
func IntPtr(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}

// Float64Value mengembalikan nilai float64 dari pointer, atau default value jika nil
func Float64Value(ptr *float64, defaultValue float64) float64 {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// Float64Ptr mengembalikan pointer ke float64, atau nil jika value adalah zero value
func Float64Ptr(value float64) *float64 {
	if value == 0 {
		return nil
	}
	return &value
}

// StringValue mengembalikan nilai string dari pointer, atau default value jika nil
func StringValue(ptr *string, defaultValue string) string {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// StringPtr mengembalikan pointer ke string, atau nil jika value adalah empty string
func StringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// IsEmptyString mengecek apakah pointer string adalah nil atau empty string
// Safe: tidak akan panic karena check nil dulu
func IsEmptyString(ptr *string) bool {
	if ptr == nil {
		return true
	}
	return *ptr == ""
}

// IsNotEmptyString mengecek apakah pointer string tidak nil dan tidak empty
func IsNotEmptyString(ptr *string) bool {
	return ptr != nil && *ptr != ""
}

