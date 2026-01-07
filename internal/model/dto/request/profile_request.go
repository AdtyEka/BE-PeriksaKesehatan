package request

// UpdateProfileRequest untuk update profil dasar (foto, tinggi, nama tampilan)
// Deprecated: Gunakan UpdateProfileMultipartRequest untuk form-data
type UpdateProfileRequest struct {
	Name     *string `json:"name" binding:"omitempty,min=1,max=100"`
	PhotoURL *string `json:"photo_url" binding:"omitempty,url"`
	Height   *int    `json:"height" binding:"omitempty,min=50,max=250"` // cm
}

// UpdateProfileMultipartRequest untuk update profil dengan form-data (multipart/form-data)
// Semua field optional, hanya update field yang dikirim
type UpdateProfileMultipartRequest struct {
	Name      *string `form:"name" binding:"omitempty,min=1,max=100"`
	BirthDate *string `form:"birth_date" binding:"omitempty"` // format: YYYY-MM-DD, akan divalidasi di handler
	Phone     *string `form:"phone" binding:"omitempty,min=6,max=30"`
	Address   *string `form:"address" binding:"omitempty,min=5"`
	PhotoURL  *string `form:"photo_url" binding:"omitempty,url"` // untuk URL foto eksternal (opsional jika upload file)
	// Photo akan di-handle sebagai *multipart.FileHeader di handler
}

// UpdatePersonalInfoRequest untuk update informasi pribadi (JSON)
// Semua field optional, hanya update field yang dikirim
type UpdatePersonalInfoRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=1,max=100"`
	BirthDate *string `json:"birth_date" binding:"omitempty"` // format: YYYY-MM-DD, akan divalidasi di handler
	Phone     *string `json:"phone" binding:"omitempty,min=6,max=30"`
	Address   *string `json:"address" binding:"omitempty,min=5"`
}

// UpdatePersonalInfoMultipartRequest untuk update informasi pribadi dengan file upload (multipart/form-data)
// Semua field optional, hanya update field yang dikirim
type UpdatePersonalInfoMultipartRequest struct {
	Name      *string `form:"name" binding:"omitempty,min=1,max=100"`
	BirthDate *string `form:"birth_date" binding:"omitempty"` // format: YYYY-MM-DD, akan divalidasi di handler
	Phone     *string `form:"phone" binding:"omitempty,min=6,max=30"`
	Address   *string `form:"address" binding:"omitempty,min=5"`
	// Photo akan di-handle sebagai *multipart.FileHeader di handler
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

// CreatePersonalInfoRequest untuk POST /api/profile (multipart/form-data)
// Note: File upload tidak bisa di-validate dengan binding tag, harus divalidasi manual di handler
// Note: Name tidak lagi diterima di POST, akan diambil dari data user (register/auth)
type CreatePersonalInfoRequest struct {
	BirthDate string  `form:"birth_date" binding:"required"` // format: YYYY-MM-DD
	Phone     *string `form:"phone" binding:"required"`       // akan divalidasi manual untuk numeric dan panjang 10-15 digit
	Address   *string `form:"address" binding:"omitempty"`   // optional
	// Photo akan di-handle sebagai *multipart.FileHeader di handler
}


