package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/model/entity"
)

// getRequiredVideoCategories menentukan kategori video yang diperlukan berdasarkan status kesehatan
// Mapping:
// - Tekanan darah (blood_pressure): jika systolic_status atau diastolic_status != "Normal" → kategori_id 2 (Hipertensi)
// - Gula darah (blood_sugar): jika status != "Normal" → kategori_id 1 (Diabetes)
// - Detak jantung (heart_rate): jika status != "Normal" → kategori_id 3 (Jantung)
// Menggunakan map untuk memastikan tidak ada duplikasi kategori
func (s *HealthDataService) getRequiredVideoCategories(summary response.HealthSummaryResponse) []uint {
	categoryMap := make(map[uint]bool)

	// Cek tekanan darah
	if summary.BloodPressure != nil {
		if summary.BloodPressure.SystolicStatus != "Normal" || summary.BloodPressure.DiastolicStatus != "Normal" {
			categoryMap[2] = true // Hipertensi
		}
	}

	// Cek gula darah
	if summary.BloodSugar != nil {
		if summary.BloodSugar.Status != "Normal" {
			categoryMap[1] = true // Diabetes
		}
	}

	// Cek detak jantung (jika ada di summary)
	// Note: Saat ini heart_rate tidak ada di summary, tapi kita siapkan untuk future
	// Jika nanti ditambahkan, bisa ditambahkan logika di sini

	// Convert map ke slice
	categoryIDs := make([]uint, 0, len(categoryMap))
	for categoryID := range categoryMap {
		categoryIDs = append(categoryIDs, categoryID)
	}

	return categoryIDs
}

// getEducationVideos mengambil dan memformat video edukasi berdasarkan kategori yang diperlukan
func (s *HealthDataService) getEducationVideos(summary response.HealthSummaryResponse) ([]response.EducationVideoCategory, error) {
	// Tentukan kategori yang diperlukan
	categoryIDs := s.getRequiredVideoCategories(summary)

	// Jika tidak ada kategori yang diperlukan, kembalikan array kosong
	if len(categoryIDs) == 0 {
		return []response.EducationVideoCategory{}, nil
	}

	// Ambil kategori dari database
	categories, err := s.categoryRepo.GetAllCategories()
	if err != nil {
		// Fail-safe: jika error, kembalikan array kosong
		return []response.EducationVideoCategory{}, nil
	}

	// Buat map untuk lookup kategori
	categoryMap := make(map[uint]entity.Category)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat
	}

	// Ambil video berdasarkan kategori IDs
	videosByCategory, err := s.educationalVideoRepo.GetAllEducationalVideosByCategoryIDs(categoryIDs)
	if err != nil {
		// Fail-safe: jika error, kembalikan array kosong
		return []response.EducationVideoCategory{}, nil
	}

	// Build response
	result := make([]response.EducationVideoCategory, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		category, exists := categoryMap[categoryID]
		if !exists {
			continue // Skip jika kategori tidak ditemukan
		}

		videos := videosByCategory[categoryID]
		if len(videos) == 0 {
			continue // Skip jika tidak ada video
		}

		// Format video items
		videoItems := make([]response.EducationVideoItem, 0, len(videos))
		for _, video := range videos {
			videoItems = append(videoItems, response.EducationVideoItem{
				ID:        video.ID,
				Title:     video.VideoTitle,
				URL:       video.VideoURL,
				Thumbnail: "", // Tidak ada di entity, set empty string
				Duration:  0,  // Tidak ada di entity, set 0
			})
		}

		result = append(result, response.EducationVideoCategory{
			KategoriID: category.ID,
			Kategori:   category.Kategori,
			Videos:     videoItems,
		})
	}

	return result, nil
}

