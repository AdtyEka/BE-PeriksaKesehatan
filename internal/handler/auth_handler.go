package handler

import (
	"BE-PeriksaKesehatan/config"
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/repository"
	"BE-PeriksaKesehatan/pkg/utils"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler menangani semua request terkait autentikasi
type AuthHandler struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

// NewAuthHandler membuat instance baru dari AuthHandler
func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	// Ambil secret dari environment (langsung) agar handler tidak perlu bergantung pada struct config
	// Jika suatu saat ingin lebih rapi, bisa diubah ke dependency injection dari main.
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// fallback: coba ambil dari config untuk memastikan kalau LoadConfig sudah dijalankan
		cfg := config.LoadConfig()
		secret = cfg.JWTSecret
	}

	return &AuthHandler{
		userRepo:  userRepo,
		jwtSecret: secret,
	}
}

// Register menangani request pendaftaran user baru
func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest

	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	// Validasi: Pastikan password dan confirm_password sama
	if req.Password != req.ConfirmPassword {
		utils.BadRequest(c, "Password dan konfirmasi password tidak sama", nil)
		return
	}

	// Validasi: Cek apakah email sudah terdaftar
	emailExists, err := h.userRepo.CheckEmailExists(req.Email)
	if err != nil {
		utils.InternalServerError(c, "Gagal memeriksa email", err.Error())
		return
	}
	if emailExists {
		utils.BadRequest(c, "Email sudah terdaftar", nil)
		return
	}

	// Tentukan username: sama persis dengan nama lengkap (hanya di-trim spasi di depan/belakang)
	username := strings.TrimSpace(req.Nama)
	if username == "" {
		utils.BadRequest(c, "Username tidak boleh kosong", nil)
		return
	}

	usernameExists, err := h.userRepo.CheckUsernameExists(username)
	if err != nil {
		utils.InternalServerError(c, "Gagal memeriksa username", err.Error())
		return
	}
	if usernameExists {
		utils.BadRequest(c, "Username sudah terdaftar", nil)
		return
	}

	// Hash password menggunakan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.InternalServerError(c, "Gagal mengenkripsi password", err.Error())
		return
	}

	// Buat user baru
	user := &entity.User{
		Nama:     req.Nama,
		Username: username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	// Simpan ke database
	if err := h.userRepo.CreateUser(user); err != nil {
		utils.InternalServerError(c, "Gagal mendaftarkan user", err.Error())
		return
	}

	// Response sukses
	resp := response.RegisterResponse{
		Message: "Pendaftaran berhasil",
		Nama:    user.Nama,
		Email:   user.Email,
	}

	utils.SuccessResponse(c, http.StatusCreated, resp.Message, resp)
}

// Login menangani request login user
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest

	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	// Cari user berdasarkan email atau username
	user, err := h.userRepo.GetUserByEmailOrUsername(req.Identifier)
	if err != nil {
		utils.Unauthorized(c, "Email/Username atau password salah")
		return
	}

	// Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.Unauthorized(c, "Email/Username atau password salah")
		return
	}
	// Generate JWT token sungguhan
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"name":  user.Nama,
		"exp":   time.Now().Add(24 * time.Hour).Unix(), // token berlaku 24 jam
		"iat":   time.Now().Unix(),
	}

	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenObj.SignedString([]byte(h.jwtSecret))
	if err != nil {
		utils.InternalServerError(c, "Gagal membuat token", err.Error())
		return
	}

	// Response sukses
	resp := response.LoginResponse{
		Token:    token,
		Nama:     user.Nama,
		Username: user.Username,
		Email:    user.Email,
	}

	utils.SuccessResponse(c, http.StatusOK, "Login berhasil", resp)
}

