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

// AddEducationalVideo menambahkan video edukasi baru
func (s *EducationalVideoService) AddEducationalVideo(req *request.EducationalVideoRequest) (*response.AddEducationalVideoResponse, error) {
	// Validasi
	if err := s.validateVideoRequest(req); err != nil {
		return nil, err
	}

	// Validasi kategori exists
	category, err := s.categoryRepo.GetCategoryByID(req.CategoryID)
	if err != nil {
		return nil, errors.New("kategori tidak ditemukan")
	}

	// Buat entity
	video := &entity.EducationalVideo{
		VideoTitle:      strings.TrimSpace(req.VideoTitle),
		VideoURL:        strings.TrimSpace(req.VideoURL),
		CategoryID:      req.CategoryID,
		HealthCondition: category.Kategori,
	}

	// Simpan ke database
	if err := s.educationalVideoRepo.CreateEducationalVideo(video); err != nil {
		return nil, err
	}

	// Buat response
	resp := &response.AddEducationalVideoResponse{
		ID:         video.ID,
		VideoTitle: video.VideoTitle,
		VideoURL:   video.VideoURL,
		CategoryID: video.CategoryID,
	}

	return resp, nil
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

	// Validasi category_id tidak kosong
	if req.CategoryID == 0 {
		return errors.New("category_id tidak boleh kosong")
	}

	return nil
}

