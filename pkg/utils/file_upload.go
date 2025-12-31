package utils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// MaxFileSize adalah maksimal ukuran file dalam bytes (2MB)
	MaxFileSize = 2 * 1024 * 1024
	// AllowedImageTypes adalah tipe file image yang diizinkan
	AllowedImageTypes = "image/jpeg,image/jpg,image/png,image/webp"
	// UploadDir adalah direktori untuk menyimpan file upload
	UploadDir = "uploads/profile"
)

// ValidateImageFile memvalidasi file image yang diupload
// Returns: error jika file tidak valid, nil jika valid
func ValidateImageFile(fileHeader *multipart.FileHeader) error {
	if fileHeader == nil {
		return nil // File tidak wajib, skip validasi
	}

	// Validasi ukuran file
	if fileHeader.Size > MaxFileSize {
		return fmt.Errorf("ukuran file terlalu besar, maksimal %d MB", MaxFileSize/(1024*1024))
	}

	// Validasi tipe file
	contentType := fileHeader.Header.Get("Content-Type")
	allowedTypes := []string{"image/jpeg", "image/jpg", "image/png", "image/webp"}
	
	isAllowed := false
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("tipe file tidak didukung, hanya jpg, jpeg, png, dan webp yang diizinkan")
	}

	// Validasi ekstensi file
	filename := fileHeader.Filename
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".webp"}
	
	isExtAllowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isExtAllowed = true
			break
		}
	}

	if !isExtAllowed {
		return fmt.Errorf("ekstensi file tidak didukung, hanya .jpg, .jpeg, .png, dan .webp yang diizinkan")
	}

	return nil
}

// UploadProfileImage mengupload file image ke local storage
// Returns: filepath relatif (untuk disimpan di DB) dan error
func UploadProfileImage(fileHeader *multipart.FileHeader, userID uint) (string, error) {
	if fileHeader == nil {
		return "", nil // File tidak wajib
	}

	// Validasi file
	if err := ValidateImageFile(fileHeader); err != nil {
		return "", err
	}

	// Buka file
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %w", err)
	}
	defer file.Close()

	// Buat direktori jika belum ada
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori: %w", err)
	}

	// Generate filename: userID_timestamp.ext
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%d%s", userID, timestamp, ext)
	destPath := filepath.Join(UploadDir, filename)

	// Buat file baru
	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file: %w", err)
	}
	defer dst.Close()

	// Copy file
	if _, err := io.Copy(dst, file); err != nil {
		// Jika gagal copy, hapus file yang sudah dibuat
		os.Remove(destPath)
		return "", fmt.Errorf("gagal menyimpan file: %w", err)
	}

	// Return relative path untuk disimpan di database
	// Format: uploads/profile/userID_timestamp.ext
	return destPath, nil
}

// DeleteProfileImage menghapus file image dari local storage
func DeleteProfileImage(filepath string) error {
	if filepath == "" {
		return nil // Tidak ada file untuk dihapus
	}

	// Hanya hapus jika file ada di direktori uploads
	if !strings.HasPrefix(filepath, UploadDir) {
		return errors.New("path file tidak valid")
	}

	if err := os.Remove(filepath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("gagal menghapus file: %w", err)
		}
		// File sudah tidak ada, tidak perlu error
	}

	return nil
}

// GetFileExtension mendapatkan ekstensi file dari filename
func GetFileExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

