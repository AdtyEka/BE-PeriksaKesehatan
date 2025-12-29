package repository

import (
	"BE-PeriksaKesehatan/internal/model/entity"
	"errors"

	"gorm.io/gorm"
)

// UserRepository adalah struct yang menampung koneksi database
// File ini tidak peduli dari mana koneksi datang, dia hanya tahu "saya punya akses ke database"
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository membuat instance baru dari UserRepository
// Menerima koneksi database yang sudah aktif sebagai parameter
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// ==================== OPERASI CREATE (INSERT) ====================

// CreateUser melakukan INSERT data user baru ke database
func (r *UserRepository) CreateUser(user *entity.User) error {
	// Validasi input dasar
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

// ==================== OPERASI READ (SELECT) ====================

// GetUserByID melakukan SELECT user berdasarkan ID
// Error handling:
// - gorm.ErrRecordNotFound → return error "user tidak ditemukan"
// - Error database lain → return error asli
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

// GetUserByEmail melakukan SELECT user berdasarkan email
// Error handling:
// - gorm.ErrRecordNotFound → return error "user tidak ditemukan"
// - Error database lain → return error asli
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

// GetUserByUsername melakukan SELECT user berdasarkan username
// Error handling:
// - gorm.ErrRecordNotFound → return error "user tidak ditemukan"
// - Error database lain → return error asli
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

// GetUserByEmailOrUsername melakukan SELECT user berdasarkan email ATAU username
// Berguna untuk login karena bisa menggunakan email atau username
// Error handling:
// - gorm.ErrRecordNotFound → return error "user tidak ditemukan" (untuk 401)
// - Error database lain → return error asli (untuk 500)
func (r *UserRepository) GetUserByEmailOrUsername(account string) (*entity.User, error) {
	if account == "" {
		return nil, errors.New("identifier tidak boleh kosong")
	}

	var user entity.User
	
	// Gunakan query yang lebih stabil dengan context
	result := r.db.Where("email = ? OR username = ?", account, account).First(&user)
	
	if result.Error != nil {
		// Bedakan antara "user tidak ditemukan" dengan error database lain
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// User tidak ditemukan → return error khusus untuk 401
			return nil, errors.New("user tidak ditemukan")
		}
		// Error database lain (prepared statement, connection, dll) → return error asli untuk 500
		return nil, result.Error
	}
	
	return &user, nil
}

// GetAllUsers melakukan SELECT semua user dari database
func (r *UserRepository) GetAllUsers() ([]entity.User, error) {
	var users []entity.User
	result := r.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// ==================== OPERASI UPDATE ====================

// UpdateUser melakukan UPDATE data user berdasarkan ID
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

// UpdateUserPassword melakukan UPDATE password user berdasarkan ID
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

// ==================== OPERASI DELETE ====================

// DeleteUser melakukan DELETE user berdasarkan ID
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

// ==================== OPERASI VALIDASI ====================

// CheckEmailExists mengecek apakah email sudah terdaftar
func (r *UserRepository) CheckEmailExists(email string) (bool, error) {
	var count int64
	result := r.db.Model(&entity.User{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// CheckUsernameExists mengecek apakah username sudah terdaftar
func (r *UserRepository) CheckUsernameExists(username string) (bool, error) {
	var count int64
	result := r.db.Model(&entity.User{}).Where("username = ?", username).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// ==================== HELPER METHODS ====================

// GetDB mengembalikan koneksi database untuk digunakan oleh repository lain
func (r *UserRepository) GetDB() *gorm.DB {
	return r.db
}

