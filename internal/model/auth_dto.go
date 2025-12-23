package model

// RegisterRequest untuk menangkap input JSON saat pendaftaran
type RegisterRequest struct {
	Nama            string `json:"nama" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=6"`
}

// LoginRequest untuk menangkap input JSON saat login
type LoginRequest struct {
    Account  string `json:"account" binding:"required"` 
    Password string `json:"password" binding:"required"`
}

// RegisterResponse untuk mengirim balik data setelah pendaftaran
type RegisterResponse struct {
	Message string `json:"message"`
	Nama    string `json:"nama"`
	Email   string `json:"email"`
}

// LoginResponse untuk mengirim balik data ke Flutter
type LoginResponse struct {
    Token    string `json:"token"`
    Nama     string `json:"nama"`
    Username string `json:"username"`
    Email    string `json:"email"`
}