package request

// UpdateProfileRequest untuk update profil dasar (foto, tinggi, nama tampilan)
type UpdateProfileRequest struct {
	Name     *string `json:"name" binding:"omitempty,min=1,max=100"`
	PhotoURL *string `json:"photo_url" binding:"omitempty,url"`
	Height   *int    `json:"height" binding:"omitempty,min=50,max=250"` // cm
}

// UpdatePersonalInfoRequest untuk update informasi pribadi
type UpdatePersonalInfoRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	BirthDate   string  `json:"birth_date" binding:"required"`           // format: YYYY-MM-DD
	Phone       *string `json:"phone" binding:"omitempty,min=6,max=30"`
	Address     *string `json:"address" binding:"omitempty,min=5"`
}

// UpdateHealthTargetsRequest untuk update target kesehatan
type UpdateHealthTargetsRequest struct {
	TargetSystolic    *int     `json:"target_systolic" binding:"omitempty,min=90,max=180"`
	TargetDiastolic   *int     `json:"target_diastolic" binding:"omitempty,min=60,max=120"`
	TargetBloodSugar  *int     `json:"target_blood_sugar" binding:"omitempty,min=70,max=250"`
	TargetWeight      *float64 `json:"target_weight" binding:"omitempty,min=30,max=300"`
}

// UpdateSettingsRequest untuk update pengaturan aplikasi
type UpdateSettingsRequest struct {
	NotificationEnabled *bool   `json:"notification_enabled" binding:"omitempty"`
	Language            *string `json:"language" binding:"omitempty,oneof=id en"`
}

// CreateProfileRequest untuk POST /api/profile (multipart/form-data)
// Note: File upload tidak bisa di-validate dengan binding tag, harus divalidasi manual di handler
type CreateProfileRequest struct {
	Name   string  `form:"name" binding:"required,min=1,max=100"`
	Email  string  `form:"email" binding:"required,email"`
	Weight *float64 `form:"weight"` // optional, akan divalidasi manual
	Height *int    `form:"height"`  // optional, akan divalidasi manual
	Age    *int    `form:"age"`     // optional, akan divalidasi manual
	// Photo akan di-handle sebagai *multipart.FileHeader di handler
}


