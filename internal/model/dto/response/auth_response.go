package response

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

