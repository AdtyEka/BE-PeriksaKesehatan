package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(user *entity.User) error {
	if user == nil {
		return errors.New("user tidak boleh nil")
	}
	if user.Email == "" || user.Username == "" || user.Password == "" {
		return errors.New("email, username, dan password harus diisi")
	}

	result := r.db.Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepository) GetUserByID(id uint) (*entity.User, error) {
	if id == 0 {
		return nil, errors.New("ID tidak valid")
	}

	var user entity.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user tidak ditemukan")
		}
		// Error database lain → return error asli
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*entity.User, error) {
	if email == "" {
		return nil, errors.New("email tidak boleh kosong")
	}

	var user entity.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user tidak ditemukan")
		}
		// Error database lain → return error asli
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) GetUserByUsername(username string) (*entity.User, error) {
	if username == "" {
		return nil, errors.New("username tidak boleh kosong")
	}

	var user entity.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user tidak ditemukan")
		}
		// Error database lain → return error asli
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) GetUserByEmailOrUsername(account string) (*entity.User, error) {
	if account == "" {
		return nil, errors.New("identifier tidak boleh kosong")
	}

	var user entity.User
	result := r.db.Where("email = ? OR username = ?", account, account).First(&user)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, result.Error
	}
	
	return &user, nil
}

func (r *UserRepository) GetAllUsers() ([]entity.User, error) {
	var users []entity.User
	result := r.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *UserRepository) UpdateUser(id uint, user *entity.User) error {
	result := r.db.Model(&entity.User{}).Where("id = ?", id).Updates(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}
	return nil
}

func (r *UserRepository) UpdateUserPassword(id uint, hashedPassword string) error {
	result := r.db.Model(&entity.User{}).Where("id = ?", id).Update("password", hashedPassword)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}
	return nil
}

func (r *UserRepository) UpdateUserProfile(id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.New("tidak ada data untuk diupdate")
	}
	result := r.db.Model(&entity.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}
	return nil
}

func (r *UserRepository) UpdateUserPersonalInfo(id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.New("tidak ada data untuk diupdate")
	}
	result := r.db.Model(&entity.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}
	return nil
}

func (r *UserRepository) UpdateUserSettings(id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.New("tidak ada data untuk diupdate")
	}
	result := r.db.Model(&entity.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}
	return nil
}

func (r *UserRepository) DeleteUser(id uint) error {
	result := r.db.Delete(&entity.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}
	return nil
}

func (r *UserRepository) CheckEmailExists(email string) (bool, error) {
	var count int64
	result := r.db.Model(&entity.User{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func (r *UserRepository) CheckUsernameExists(username string) (bool, error) {
	var count int64
	result := r.db.Model(&entity.User{}).Where("username = ?", username).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func (r *UserRepository) GetDB() *gorm.DB {
	return r.db
}

