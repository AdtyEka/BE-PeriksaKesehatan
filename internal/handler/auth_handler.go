package handler

import (
	"BE-PeriksaKesehatan/config"
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/repository"
	"BE-PeriksaKesehatan/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler menangani semua request terkait autentikasi
type AuthHandler struct {
	userRepo  *repository.UserRepository
	authRepo  *repository.AuthRepository
	jwtSecret string
}

// NewAuthHandler membuat instance baru dari AuthHandler
func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	// Ambil secret dari config (sudah divalidasi di main.go)
	cfg := config.LoadConfig()
	secret := cfg.JWTSecret

	// Buat AuthRepository untuk operasi blacklist token
	authRepo := repository.NewAuthRepository(userRepo.GetDB())

	return &AuthHandler{
		userRepo:  userRepo,
		authRepo:  authRepo,
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
// Error handling:
// - "user tidak ditemukan" → return 401 Unauthorized
// - Error database lain → return 500 Internal Server Error
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
		// Bedakan antara "user tidak ditemukan" dengan error database lain
		if err.Error() == "user tidak ditemukan" {
			// User tidak ditemukan → return 401
			utils.Unauthorized(c, "Email/Username atau password salah")
			return
		}
		// Error database lain (prepared statement, connection, dll) → return 500
		utils.InternalServerError(c, "Gagal memproses login", err.Error())
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

// extractTokenFromHeader mengambil token dari header Authorization
// Mengembalikan token string dan true jika berhasil, false jika tidak ada token
func extractTokenFromHeader(authHeader string) (string, bool) {
	if authHeader == "" {
		return "", false
	}

	// Parse token (format: "Bearer <token>")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:], true
	}
	return authHeader, true
}

// extractUserIDFromClaims mengambil user ID dari JWT claims dengan type assertion yang aman
func extractUserIDFromClaims(claims jwt.MapClaims) (uint, bool) {
	sub, ok := claims["sub"]
	if !ok {
		return 0, false
	}

	// Coba sebagai float64 (format standar JWT untuk angka)
	if userIDFloat, ok := sub.(float64); ok {
		return uint(userIDFloat), true
	}

	// Coba sebagai string (untuk kompatibilitas)
	if userIDStr, ok := sub.(string); ok {
		userIDUint, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			return 0, false
		}
		return uint(userIDUint), true
	}

	return 0, false
}

// extractExpiryFromClaims mengambil waktu kadaluarsa dari JWT claims
func extractExpiryFromClaims(claims jwt.MapClaims) time.Time {
	if exp, ok := claims["exp"].(float64); ok {
		return time.Unix(int64(exp), 0)
	}
	// Fallback: default 24 jam dari sekarang
	return time.Now().Add(24 * time.Hour)
}

// Logout menangani request logout user
// Method ini idempotent - selalu return sukses meskipun token tidak valid atau sudah di-blacklist
func (h *AuthHandler) Logout(c *gin.Context) {
	// Ambil token dari header
	authHeader := c.GetHeader("Authorization")
	tokenString, hasToken := extractTokenFromHeader(authHeader)
	if !hasToken {
		utils.SuccessResponse(c, http.StatusOK, "Sesi sudah berakhir", nil)
		return
	}

	// Parse token untuk mendapatkan claims (tanpa validasi penuh, karena mungkin sudah expired)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})

	// Jika token tidak valid atau sudah expired, tetap return sukses (idempotent)
	if err != nil || !token.Valid {
		utils.SuccessResponse(c, http.StatusOK, "Sesi sudah berakhir", nil)
		return
	}

	// Ambil claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.SuccessResponse(c, http.StatusOK, "Sesi sudah berakhir", nil)
		return
	}

	// Cek apakah token sudah di-blacklist sebelumnya
	isBlacklisted, err := h.authRepo.IsTokenBlacklisted(tokenString)
	if err != nil {
		utils.InternalServerError(c, "Gagal memeriksa status token", err.Error())
		return
	}
	if isBlacklisted {
		utils.SuccessResponse(c, http.StatusOK, "Sesi sudah berakhir", nil)
		return
	}

	// Ambil user ID dan expiry time dari claims
	userID, ok := extractUserIDFromClaims(claims)
	if !ok {
		utils.SuccessResponse(c, http.StatusOK, "Sesi sudah berakhir", nil)
		return
	}

	expiresAt := extractExpiryFromClaims(claims)

	// Tambahkan token ke blacklist
	if err := h.authRepo.BlacklistToken(tokenString, userID, expiresAt); err != nil {
		utils.InternalServerError(c, "Gagal melakukan logout", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Logout berhasil", nil)
}

