package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// EducationalVideoService menangani business logic untuk educational videos
type EducationalVideoService struct {
	educationalVideoRepo *repository.EducationalVideoRepository
	categoryRepo         *repository.CategoryRepository
}

// NewEducationalVideoService membuat instance baru dari EducationalVideoService
func NewEducationalVideoService(educationalVideoRepo *repository.EducationalVideoRepository, categoryRepo *repository.CategoryRepository) *EducationalVideoService {
	return &EducationalVideoService{
		educationalVideoRepo: educationalVideoRepo,
		categoryRepo:         categoryRepo,
	}
}

// AddEducationalVideo menambahkan video edukasi baru dengan multiple categories
func (s *EducationalVideoService) AddEducationalVideo(req *request.EducationalVideoRequest) (*response.AddEducationalVideoResponse, error) {
	// Validasi
	if err := s.validateVideoRequest(req); err != nil {
		return nil, err
	}

	// Validasi category_ids tidak kosong
	if len(req.CategoryIDs) == 0 {
		return nil, errors.New("category_ids tidak boleh kosong")
	}

	// Validasi semua kategori exists
	categories, err := s.validateCategoriesExist(req.CategoryIDs)
	if err != nil {
		return nil, err
	}

	// Buat entity video (tanpa category_id untuk data baru)
	video := &entity.EducationalVideo{
		VideoTitle:      strings.TrimSpace(req.VideoTitle),
		VideoURL:        strings.TrimSpace(req.VideoURL),
		HealthCondition: categories[0].Kategori, // Gunakan kategori pertama untuk health_condition (backward compatibility)
	}

	// Simpan video beserta relasi kategori dengan transaksi atomic
	if err := s.educationalVideoRepo.CreateEducationalVideoWithCategories(video, req.CategoryIDs); err != nil {
		return nil, err
	}

	// Buat response
	resp := &response.AddEducationalVideoResponse{
		ID:          video.ID,
		VideoTitle:  video.VideoTitle,
		VideoURL:    video.VideoURL,
		CategoryIDs: req.CategoryIDs,
	}

	return resp, nil
}

// validateCategoriesExist memvalidasi bahwa semua category IDs ada di database
func (s *EducationalVideoService) validateCategoriesExist(categoryIDs []uint) ([]entity.Category, error) {
	if len(categoryIDs) == 0 {
		return nil, errors.New("category_ids tidak boleh kosong")
	}

	categories := make([]entity.Category, 0, len(categoryIDs))
	notFoundIDs := make([]uint, 0)

	for _, categoryID := range categoryIDs {
		category, err := s.categoryRepo.GetCategoryByID(categoryID)
		if err != nil {
			notFoundIDs = append(notFoundIDs, categoryID)
		} else {
			categories = append(categories, *category)
		}
	}

	if len(notFoundIDs) > 0 {
		return nil, errors.New("kategori tidak ditemukan")
	}

	return categories, nil
}

// GetAllEducationalVideos mengambil semua kategori beserta videonya
func (s *EducationalVideoService) GetAllEducationalVideos() (*response.GetAllEducationalVideosResponse, error) {
	// Ambil semua kategori
	categories, err := s.categoryRepo.GetAllCategories()
	if err != nil {
		return nil, err
	}

	// Ambil semua category IDs
	categoryIDs := make([]uint, 0, len(categories))
	for _, category := range categories {
		categoryIDs = append(categoryIDs, category.ID)
	}

	// Ambil semua videos berdasarkan category IDs (efisien, tidak N+1)
	videosByCategory, err := s.educationalVideoRepo.GetAllEducationalVideosByCategoryIDs(categoryIDs)
	if err != nil {
		return nil, err
	}

	// Build response
	result := make([]response.CategoryWithVideosResponse, 0, len(categories))
	for _, category := range categories {
		videos := videosByCategory[category.ID]
		videoItems := make([]response.EducationalVideoItem, 0, len(videos))
		for _, video := range videos {
			videoItems = append(videoItems, response.EducationalVideoItem{
				Title: video.VideoTitle,
				URL:   video.VideoURL,
			})
		}

		result = append(result, response.CategoryWithVideosResponse{
			ID:       category.ID,
			Kategori: category.Kategori,
			Videos:   videoItems,
		})
	}

	return &response.GetAllEducationalVideosResponse{
		Data: result,
	}, nil
}

// GetEducationalVideosByCategoryID mengambil video berdasarkan kategori ID
func (s *EducationalVideoService) GetEducationalVideosByCategoryID(categoryIDStr string) (*response.GetEducationalVideosByIDResponse, error) {
	// Validasi dan parse ID
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		return nil, errors.New("ID kategori tidak valid")
	}

	// Ambil kategori
	category, err := s.categoryRepo.GetCategoryByID(uint(categoryID))
	if err != nil {
		if err.Error() == "kategori tidak ditemukan" {
			return nil, err
		}
		return nil, err
	}

	// Ambil videos berdasarkan kategori ID
	videos, err := s.educationalVideoRepo.GetEducationalVideosByCategoryID(uint(categoryID))
	if err != nil {
		return nil, err
	}

	// Convert ke response format
	videoItems := make([]response.EducationalVideoItem, 0, len(videos))
	for _, video := range videos {
		videoItems = append(videoItems, response.EducationalVideoItem{
			Title: video.VideoTitle,
			URL:   video.VideoURL,
		})
	}

	return &response.GetEducationalVideosByIDResponse{
		ID:       category.ID,
		Kategori: category.Kategori,
		Videos:   videoItems,
	}, nil
}

// validateVideoRequest melakukan validasi request video
func (s *EducationalVideoService) validateVideoRequest(req *request.EducationalVideoRequest) error {
	// Validasi title tidak kosong
	title := strings.TrimSpace(req.VideoTitle)
	if title == "" {
		return errors.New("video_title tidak boleh kosong")
	}

	// Validasi URL tidak kosong
	urlStr := strings.TrimSpace(req.VideoURL)
	if urlStr == "" {
		return errors.New("video_url tidak boleh kosong")
	}

	// Validasi URL format
	parsedURL, err := url.Parse(urlStr)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return errors.New("video_url harus berupa URL yang valid")
	}

	// Validasi category_ids tidak kosong
	if len(req.CategoryIDs) == 0 {
		return errors.New("category_ids tidak boleh kosong")
	}

	return nil
}

