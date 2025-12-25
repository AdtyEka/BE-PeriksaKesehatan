package handler

import "BE-PeriksaKesehatan/internal/repository"

// UserHandler menangani semua request terkait user
// File ini siap untuk handler user di masa depan
type UserHandler struct {
	userRepo *repository.UserRepository
}

// NewUserHandler membuat instance baru dari UserHandler
func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

