package service

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (s *AuthService) RegisterUser(nama, username, email, password string) (*entity.User, error) {
	if nama == "" || username == "" || email == "" || password == "" {
		return nil, errors.New("semua field harus diisi")
	}
	if len(password) < 6 {
		return nil, errors.New("password minimal 6 karakter")
	}

	emailExists, err := s.userRepo.CheckEmailExists(email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, errors.New("email sudah terdaftar")
	}

	usernameExists, err := s.userRepo.CheckUsernameExists(username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, errors.New("username sudah terdaftar")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Nama:     nama,
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) LoginUser(identifier, password string) (*entity.User, error) {
	if identifier == "" || password == "" {
		return nil, errors.New("identifier dan password harus diisi")
	}

	user, err := s.userRepo.GetUserByEmailOrUsername(identifier)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			return nil, errors.New("email/username atau password salah")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("email/username atau password salah")
	}

	return user, nil
}

