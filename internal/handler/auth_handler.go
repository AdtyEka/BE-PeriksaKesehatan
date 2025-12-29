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

type AuthHandler struct {
	userRepo  *repository.UserRepository
	authRepo  *repository.AuthRepository
	jwtSecret string
}

func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	cfg := config.LoadConfig()
	secret := cfg.JWTSecret
	authRepo := repository.NewAuthRepository(userRepo.GetDB())

	return &AuthHandler{
		userRepo:  userRepo,
		authRepo:  authRepo,
		jwtSecret: secret,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	if req.Password != req.ConfirmPassword {
		utils.BadRequest(c, "Password dan konfirmasi password tidak sama", nil)
		return
	}

	emailExists, err := h.userRepo.CheckEmailExists(req.Email)
	if err != nil {
		utils.InternalServerError(c, "Gagal memeriksa email", err.Error())
		return
	}
	if emailExists {
		utils.BadRequest(c, "Email sudah terdaftar", nil)
		return
	}

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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.InternalServerError(c, "Gagal mengenkripsi password", err.Error())
		return
	}

	user := &entity.User{
		Nama:     req.Nama,
		Username: username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := h.userRepo.CreateUser(user); err != nil {
		utils.InternalServerError(c, "Gagal mendaftarkan user", err.Error())
		return
	}

	resp := response.RegisterResponse{
		Message: "Pendaftaran berhasil",
		Nama:    user.Nama,
		Email:   user.Email,
	}

	utils.SuccessResponse(c, http.StatusCreated, resp.Message, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Data tidak valid", err.Error())
		return
	}

	user, err := h.userRepo.GetUserByEmailOrUsername(req.Identifier)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			utils.Unauthorized(c, "Email/Username atau password salah")
			return
		}
		utils.InternalServerError(c, "Gagal memproses login", err.Error())
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.Unauthorized(c, "Email/Username atau password salah")
		return
	}

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"name":  user.Nama,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenObj.SignedString([]byte(h.jwtSecret))
	if err != nil {
		utils.InternalServerError(c, "Gagal membuat token", err.Error())
		return
	}

	resp := response.LoginResponse{
		Token:    token,
		Nama:     user.Nama,
		Username: user.Username,
		Email:    user.Email,
	}

	utils.SuccessResponse(c, http.StatusOK, "Login berhasil", resp)
}

func extractTokenFromHeader(authHeader string) (string, bool) {
	if authHeader == "" {
		return "", false
	}

	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:], true
	}
	return authHeader, true
}

func extractUserIDFromClaims(claims jwt.MapClaims) (uint, bool) {
	sub, ok := claims["sub"]
	if !ok {
		return 0, false
	}

	if userIDFloat, ok := sub.(float64); ok {
		return uint(userIDFloat), true
	}

	if userIDStr, ok := sub.(string); ok {
		userIDUint, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			return 0, false
		}
		return uint(userIDUint), true
	}

	return 0, false
}

func extractExpiryFromClaims(claims jwt.MapClaims) time.Time {
	if exp, ok := claims["exp"].(float64); ok {
		return time.Unix(int64(exp), 0)
	}
	return time.Now().Add(24 * time.Hour)
}

// Logout menangani request logout user.
// Method ini idempotent - selalu return sukses meskipun token tidak valid atau sudah di-blacklist.
func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	tokenString, hasToken := extractTokenFromHeader(authHeader)
	if !hasToken {
		utils.SuccessResponse(c, http.StatusOK, "Sesi sudah berakhir", nil)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		utils.SuccessResponse(c, http.StatusOK, "Sesi sudah berakhir", nil)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.SuccessResponse(c, http.StatusOK, "Sesi sudah berakhir", nil)
		return
	}

	isBlacklisted, err := h.authRepo.IsTokenBlacklisted(tokenString)
	if err != nil {
		utils.InternalServerError(c, "Gagal memeriksa status token", err.Error())
		return
	}
	if isBlacklisted {
		utils.SuccessResponse(c, http.StatusOK, "Sesi sudah berakhir", nil)
		return
	}

	userID, ok := extractUserIDFromClaims(claims)
	if !ok {
		utils.SuccessResponse(c, http.StatusOK, "Sesi sudah berakhir", nil)
		return
	}

	expiresAt := extractExpiryFromClaims(claims)

	if err := h.authRepo.BlacklistToken(tokenString, userID, expiresAt); err != nil {
		utils.InternalServerError(c, "Gagal melakukan logout", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Logout berhasil", nil)
}

