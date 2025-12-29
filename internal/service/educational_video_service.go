package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/request"
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
	"BE-PeriksaKesehatan/internal/repository"
	"errors"
	"net/url"
	"strings"
)

// EducationalVideoService menangani business logic untuk educational videos
type EducationalVideoService struct {
	educationalVideoRepo *repository.EducationalVideoRepository
}

// NewEducationalVideoService membuat instance baru dari EducationalVideoService
func NewEducationalVideoService(educationalVideoRepo *repository.EducationalVideoRepository) *EducationalVideoService {
	return &EducationalVideoService{
		educationalVideoRepo: educationalVideoRepo,
	}
}

// AddEducationalVideo menambahkan video edukasi baru
func (s *EducationalVideoService) AddEducationalVideo(req *request.EducationalVideoRequest) (*response.AddEducationalVideoResponse, error) {
	// Validasi
	if err := s.validateVideoRequest(req); err != nil {
		return nil, err
	}

	// Buat entity
	video := &entity.EducationalVideo{
		VideoTitle:      strings.TrimSpace(req.VideoTitle),
		VideoURL:        strings.TrimSpace(req.VideoURL),
		HealthCondition: strings.TrimSpace(req.HealthCondition),
	}

	// Simpan ke database
	if err := s.educationalVideoRepo.CreateEducationalVideo(video); err != nil {
		return nil, err
	}

	// Buat response
	resp := &response.AddEducationalVideoResponse{
		ID:              video.ID,
		VideoTitle:      video.VideoTitle,
		VideoURL:        video.VideoURL,
		HealthCondition: video.HealthCondition,
	}

	return resp, nil
}

// GetEducationalVideosByHealthCondition mengambil video berdasarkan kondisi kesehatan
func (s *EducationalVideoService) GetEducationalVideosByHealthCondition(healthCondition string) (*response.GetEducationalVideosResponse, error) {
	// Validasi health condition tidak kosong
	healthCondition = strings.TrimSpace(healthCondition)
	if healthCondition == "" {
		return nil, errors.New("health_condition tidak boleh kosong")
	}

	// Ambil video dari database
	videos, err := s.educationalVideoRepo.GetEducationalVideosByHealthCondition(healthCondition)
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

	return &response.GetEducationalVideosResponse{
		HealthCondition: healthCondition,
		Videos:          videoItems,
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

	// Validasi health condition tidak kosong
	healthCondition := strings.TrimSpace(req.HealthCondition)
	if healthCondition == "" {
		return errors.New("health_condition tidak boleh kosong")
	}

	return nil
}

