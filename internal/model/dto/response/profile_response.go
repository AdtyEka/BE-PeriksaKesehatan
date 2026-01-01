package response

// ProfileResponse untuk GET /profile
type ProfileResponse struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	PhotoURL *string  `json:"photo_url,omitempty"`
	Weight   *float64 `json:"weight"` // dari health_data terbaru (bisa null)
	Height   *int     `json:"height"` // dari health_data terbaru (bisa null)
	Age      *int     `json:"age"`    // dihitung dari birth_date (bisa null)
}

// PersonalInfoResponse untuk menampilkan informasi pribadi
type PersonalInfoResponse struct {
	Name      string  `json:"name"`
	BirthDate *string `json:"birth_date,omitempty"` // YYYY-MM-DD
	Phone     *string `json:"phone,omitempty"`
	Address   *string `json:"address,omitempty"`
}

// BloodPressureTargetProgress ringkasan target tekanan darah
type BloodPressureTargetProgress struct {
	TargetSystolic  *int `json:"target_systolic,omitempty"`
	TargetDiastolic *int `json:"target_diastolic,omitempty"`
	CurrentSystolic *int `json:"current_systolic,omitempty"`
	CurrentDiastolic *int `json:"current_diastolic,omitempty"`
	// progress dalam persen (mendekati target), bisa negatif jika jauh dari target
	ProgressPercent *float64 `json:"progress_percent,omitempty"`
}

// BloodSugarTargetProgress ringkasan target gula darah
type BloodSugarTargetProgress struct {
	Target      *int     `json:"target,omitempty"`
	Current     *int     `json:"current,omitempty"`
	ProgressPercent *float64 `json:"progress_percent,omitempty"`
}

// WeightTargetProgress ringkasan target berat badan
type WeightTargetProgress struct {
	Target      *float64 `json:"target,omitempty"`
	Current     *float64 `json:"current,omitempty"` // dari health_data terbaru
	ProgressPercent *float64 `json:"progress_percent,omitempty"`
}

// HealthTargetsResponse untuk GET /profile/health-targets
type HealthTargetsResponse struct {
	BloodPressure *BloodPressureTargetProgress `json:"blood_pressure,omitempty"`
	BloodSugar    *BloodSugarTargetProgress    `json:"blood_sugar,omitempty"`
	Weight        *WeightTargetProgress        `json:"weight,omitempty"`
}

// SettingsResponse untuk GET /profile/settings
type SettingsResponse struct {
	NotificationEnabled *bool   `json:"notification_enabled,omitempty"`
	Language            *string `json:"language,omitempty"`

	// Informasi statis yang bisa ditampilkan di frontend
	PrivacyInfo string `json:"privacy_info"`
	SupportInfo string `json:"support_info"`
	AppVersion  string `json:"app_version"`
}


