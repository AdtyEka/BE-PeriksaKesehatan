package service

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// AuthService menangani business logic untuk autentikasi
type AuthService struct {
	userRepo *repository.UserRepository
}

// NewAuthService membuat instance baru dari AuthService
func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// RegisterUser melakukan pendaftaran user baru
func (s *AuthService) RegisterUser(nama, username, email, password string) (*entity.User, error) {
	// Validasi: Cek apakah email sudah terdaftar
	emailExists, err := s.userRepo.CheckEmailExists(email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, errors.New("email sudah terdaftar")
	}

	// Validasi: Cek apakah username sudah terdaftar
	usernameExists, err := s.userRepo.CheckUsernameExists(username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, errors.New("username sudah terdaftar")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Buat user baru
	user := &entity.User{
		Nama:     nama,
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	// Simpan ke database
	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// LoginUser melakukan login dan mengembalikan user jika berhasil
func (s *AuthService) LoginUser(identifier, password string) (*entity.User, error) {
	// Cari user berdasarkan email atau username
	user, err := s.userRepo.GetUserByEmailOrUsername(identifier)
	if err != nil {
		return nil, errors.New("email/username atau password salah")
	}

	// Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("email/username atau password salah")
	}

	return user, nil
}

