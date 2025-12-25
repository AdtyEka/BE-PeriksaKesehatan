package request

// RegisterRequest untuk menangkap input JSON saat pendaftaran
type RegisterRequest struct {
	Nama            string `json:"nama" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=6"`
}

// LoginRequest untuk menangkap input JSON saat login
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // bisa email atau username
	Password   string `json:"password" binding:"required"`
}

