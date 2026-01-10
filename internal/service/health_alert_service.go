package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/repository"
	"fmt"
	"time"

	timezoneUtils "BE-PeriksaKesehatan/pkg/utils"
)

// Kategori konstan
const (
	CategoryDiabetes   = "diabetes"
	CategoryHipertensi = "hipertensi"
	CategoryJantung    = "jantung"
	CategoryBeratBadan = "berat_badan"
)

// Mapping kategori alert ke category_id untuk video edukasi
const (
	CategoryIDDiabetes   = uint(1)
	CategoryIDHipertensi = uint(2)
	CategoryIDJantung    = uint(3)
	CategoryIDBeratBadan = uint(4)
)

type HealthAlertService struct {
	healthAlertRepo      *repository.HealthAlertRepository
	healthDataRepo       *repository.HealthDataRepository
	educationalVideoRepo *repository.EducationalVideoRepository
	categoryRepo         *repository.CategoryRepository
}

func NewHealthAlertService(
	healthAlertRepo *repository.HealthAlertRepository,
	healthDataRepo *repository.HealthDataRepository,
	educationalVideoRepo *repository.EducationalVideoRepository,
	categoryRepo *repository.CategoryRepository,
) *HealthAlertService {
	return &HealthAlertService{
		healthAlertRepo:      healthAlertRepo,
		healthDataRepo:       healthDataRepo,
		educationalVideoRepo: educationalVideoRepo,
		categoryRepo:         categoryRepo,
	}
}

// CheckHealthAlerts mengambil data kesehatan terbaru dari database dan mengevaluasi alerts
func (s *HealthAlertService) CheckHealthAlerts(userID uint) (*response.CheckHealthAlertsResponse, error) {
	// Ambil data kesehatan terbaru dari database
	latestHealthData, err := s.healthDataRepo.GetLatestHealthDataByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data kesehatan: %w", err)
	}

	// Jika tidak ada data kesehatan, return empty alerts
	if latestHealthData == nil {
		return &response.CheckHealthAlertsResponse{
			Alerts: []response.HealthAlertResponse{},
		}, nil
	}

	// Inisialisasi slice agar tidak bernilai nil saat tidak ada alert
	alerts := make([]response.HealthAlertResponse, 0)

	// Evaluasi kategori hipertensi
	if latestHealthData.Systolic != nil && latestHealthData.Diastolic != nil {
		alert := s.evaluateBloodPressure(*latestHealthData.Systolic, *latestHealthData.Diastolic, latestHealthData.CreatedAt)
		if alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	// Evaluasi kategori diabetes
	if latestHealthData.BloodSugar != nil {
		alert := s.evaluateBloodSugar(*latestHealthData.BloodSugar, latestHealthData.CreatedAt)
		if alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	// Evaluasi kategori jantung
	if latestHealthData.HeartRate != nil {
		alert := s.evaluateHeartRate(*latestHealthData.HeartRate, latestHealthData.CreatedAt)
		if alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	// Evaluasi kategori berat badan berbasis BMI
	if latestHealthData.Weight != nil && latestHealthData.HeightCM != nil {
		alert := s.evaluateBMI(*latestHealthData.Weight, *latestHealthData.HeightCM, latestHealthData.CreatedAt)
		if alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	// Batch query education videos untuk semua kategori yang memerlukan (status RENDAH atau TINGGI)
	// Kumpulkan kategori unik terlebih dahulu
	categorySet := make(map[string]bool)
	for i := range alerts {
		if alerts[i].Status == StatusRendah || alerts[i].Status == StatusTinggi {
			categorySet[alerts[i].Category] = true
		}
	}

	// Batch query semua video untuk kategori yang diperlukan
	videosByCategory := s.getEducationVideosByCategories(categorySet)

	// Isi education_videos untuk setiap alert
	for i := range alerts {
		if alerts[i].Status == StatusRendah || alerts[i].Status == StatusTinggi {
			alerts[i].EducationVideos = videosByCategory[alerts[i].Category]
		}
	}

	return &response.CheckHealthAlertsResponse{
		Alerts: alerts,
	}, nil
}

// evaluateBloodPressure mengevaluasi tekanan darah dan mengembalikan alert jika tidak normal
func (s *HealthAlertService) evaluateBloodPressure(systolic, diastolic int, recordedAt time.Time) *response.HealthAlertResponse {
	status := getBloodPressureStatusValue(systolic, diastolic)
	if status == StatusNormal {
		return nil
	}

	value := fmt.Sprintf("%d / %d mmHg", systolic, diastolic)
	var alertType, label, explanation string
	var immediateActions, medicalAttention, managementTips []string

	if systolic >= 140 || diastolic >= 90 {
		// Hipertensi
		alertType = "Tekanan Darah Tinggi"
		label = "Hipertensi"
		explanation = "Tekanan darah Anda berada di atas batas normal dan dapat meningkatkan risiko stroke, serangan jantung, dan penyakit ginjal."
		immediateActions = []string{
			"Duduk atau berbaring dengan tenang",
			"Hindari garam dan kafein",
			"Lakukan relaksasi pernapasan dalam",
		}
		medicalAttention = []string{
			"Jika tekanan darah tetap tinggi setelah 1 jam",
			"Jika disertai nyeri dada atau pusing berat",
			"Konsultasi dengan dokter dalam 24 jam",
		}
		managementTips = []string{
			"Batasi konsumsi garam maksimal 5 gram per hari",
			"Olahraga rutin minimal 30 menit per hari",
			"Pertahankan berat badan ideal",
			"Hindari merokok dan alkohol",
		}
	} else if systolic < 90 || diastolic < 60 {
		// Hipotensi
		alertType = "Tekanan Darah Rendah"
		label = "Hipotensi"
		explanation = "Tekanan darah Anda berada di bawah batas normal. Kondisi ini dapat menyebabkan pusing, lemas, atau pingsan."
		immediateActions = []string{
			"Berbaring dengan kaki lebih tinggi dari kepala",
			"Minum air putih yang cukup",
			"Hindari berdiri terlalu cepat",
		}
		medicalAttention = []string{
			"Jika disertai pingsan atau kehilangan kesadaran",
			"Jika terjadi secara tiba-tiba dan berulang",
			"Konsultasi dengan dokter jika gejala menetap",
		}
		managementTips = []string{
			"Tingkatkan asupan cairan",
			"Konsumsi makanan bergizi seimbang",
			"Hindari berdiri terlalu lama",
			"Tidur dengan bantal lebih tinggi",
		}
	} else {
		return nil
	}

	return &response.HealthAlertResponse{
		AlertType:        alertType,
		Category:         CategoryHipertensi,
		Value:            value,
		Label:            label,
		Status:           status,
		RecordedAt:       timezoneUtils.ToJakarta(recordedAt),
		Explanation:      explanation,
		ImmediateActions: immediateActions,
		MedicalAttention: medicalAttention,
		ManagementTips:   managementTips,
		EducationVideos:  []response.EducationVideoItem{}, // Akan diisi di CheckHealthAlerts
	}
}

// evaluateBloodSugar mengevaluasi gula darah dan mengembalikan alert jika tidak normal
func (s *HealthAlertService) evaluateBloodSugar(bloodSugar int, recordedAt time.Time) *response.HealthAlertResponse {
	status := getBloodSugarStatusValue(bloodSugar)
	if status == StatusNormal {
		return nil
	}

	value := fmt.Sprintf("%d mg/dL", bloodSugar)
	var alertType, label, explanation string
	var immediateActions, medicalAttention, managementTips []string

	if bloodSugar < 70 {
		// Hipoglikemia (RENDAH)
		alertType = "Gula Darah Rendah"
		label = "Gula Darah Rendah"
		explanation = "Gula darah Anda berada di bawah batas normal (WHO: < 70 mg/dL). Kondisi ini memerlukan perhatian segera karena dapat menyebabkan pingsan, kejang, atau koma."
		immediateActions = []string{
			"Segera konsumsi 15-20 gram gula sederhana (permen, jus buah, atau tablet glukosa)",
			"Tunggu 15 menit dan periksa kembali gula darah",
			"Jika masih rendah, ulangi konsumsi gula",
		}
		medicalAttention = []string{
			"Jika tidak sadar atau tidak bisa menelan",
			"Jika gula darah tidak naik setelah 2 kali konsumsi gula",
			"Jika disertai kejang atau kehilangan kesadaran",
		}
		managementTips = []string{
			"Makan teratur dengan porsi kecil tapi sering",
			"Selalu siapkan camilan manis untuk keadaan darurat",
			"Monitor gula darah secara rutin",
			"Konsultasi dengan dokter untuk penyesuaian obat",
		}
	} else if bloodSugar > 140 {
		// Hiperglikemia (TINGGI) - sesuai standar WHO untuk gula darah sewaktu
		alertType = "Gula Darah Tinggi"
		label = "Gula Darah Tinggi"
		explanation = "Gula darah Anda berada di atas batas normal (WHO: > 140 mg/dL untuk gula darah sewaktu). Jika berlangsung lama, dapat meningkatkan risiko komplikasi kesehatan."
		immediateActions = []string{
			"Hindari makanan dan minuman manis",
			"Lakukan aktivitas fisik ringan jika memungkinkan",
			"Minum air putih yang cukup",
		}
		medicalAttention = []string{
			"Jika gula darah tetap tinggi setelah beberapa hari",
			"Jika disertai gejala seperti sering haus, sering buang air kecil, atau lemas",
			"Konsultasi dengan dokter untuk evaluasi",
		}
		managementTips = []string{
			"Batasi konsumsi karbohidrat dan gula",
			"Pilih karbohidrat kompleks (nasi merah, roti gandum)",
			"Olahraga rutin minimal 30 menit per hari",
			"Pertahankan berat badan ideal",
			"Monitor gula darah secara rutin",
		}
	} else {
		return nil
	}

	return &response.HealthAlertResponse{
		AlertType:        alertType,
		Category:         CategoryDiabetes,
		Value:            value,
		Label:            label,
		Status:           status,
		RecordedAt:       timezoneUtils.ToJakarta(recordedAt),
		Explanation:      explanation,
		ImmediateActions: immediateActions,
		MedicalAttention: medicalAttention,
		ManagementTips:   managementTips,
		EducationVideos:  []response.EducationVideoItem{}, // Akan diisi di CheckHealthAlerts
	}
}

// evaluateHeartRate mengevaluasi detak jantung dan mengembalikan alert jika tidak normal
func (s *HealthAlertService) evaluateHeartRate(heartRate int, recordedAt time.Time) *response.HealthAlertResponse {
	status := getHeartRateStatusValue(heartRate)
	if status == StatusNormal {
		return nil
	}

	value := fmt.Sprintf("%d bpm", heartRate)
	var alertType, label, explanation string
	var immediateActions, medicalAttention, managementTips []string

	if heartRate < 60 {
		// Bradikardia
		alertType = "Detak Jantung Lambat"
		label = "Bradikardia"
		explanation = "Detak jantung Anda berada di bawah batas normal. Kondisi ini dapat menyebabkan kelelahan, pusing, atau pingsan karena jantung tidak memompa cukup darah ke seluruh tubuh."
		immediateActions = []string{
			"Berbaring atau duduk dengan tenang",
			"Hindari aktivitas fisik yang berat",
			"Monitor gejala seperti pusing atau sesak napas",
		}
		medicalAttention = []string{
			"Jika disertai pingsan atau kehilangan kesadaran",
			"Jika disertai nyeri dada atau sesak napas",
			"Jika terjadi secara tiba-tiba",
			"Konsultasi dengan dokter untuk evaluasi",
		}
		managementTips = []string{
			"Konsultasi dengan dokter untuk pemeriksaan jantung",
			"Hindari konsumsi kafein berlebihan",
			"Olahraga ringan secara teratur",
			"Monitor detak jantung secara rutin",
		}
	} else if heartRate > 100 {
		// Takikardia
		alertType = "Detak Jantung Cepat"
		label = "Takikardia"
		explanation = "Detak jantung Anda berada di atas batas normal. Kondisi ini dapat menyebabkan palpitasi, pusing, atau sesak napas."
		immediateActions = []string{
			"Duduk atau berbaring dengan tenang",
			"Lakukan relaksasi pernapasan dalam",
			"Hindari kafein dan stimulan lainnya",
		}
		medicalAttention = []string{
			"Jika disertai nyeri dada atau sesak napas berat",
			"Jika detak jantung tidak kembali normal setelah istirahat",
			"Jika terjadi secara tiba-tiba dan berulang",
			"Segera ke unit gawat darurat jika disertai gejala serius",
		}
		managementTips = []string{
			"Kurangi konsumsi kafein dan alkohol",
			"Kelola stres dengan baik",
			"Olahraga teratur dengan intensitas sedang",
			"Tidur cukup minimal 7-8 jam per hari",
			"Monitor detak jantung secara rutin",
		}
	} else {
		return nil
	}

	return &response.HealthAlertResponse{
		AlertType:        alertType,
		Category:         CategoryJantung,
		Value:            value,
		Label:            label,
		Status:           status,
		RecordedAt:       timezoneUtils.ToJakarta(recordedAt),
		Explanation:      explanation,
		ImmediateActions: immediateActions,
		MedicalAttention: medicalAttention,
		ManagementTips:   managementTips,
		EducationVideos:  []response.EducationVideoItem{}, // Akan diisi di CheckHealthAlerts
	}
}

// evaluateBMI mengevaluasi status BMI dan mengembalikan alert jika tidak normal
func (s *HealthAlertService) evaluateBMI(weightKg float64, heightCM int, recordedAt time.Time) *response.HealthAlertResponse {
	if weightKg <= 0 || heightCM <= 0 {
		return nil
	}

	bmiValue := calculateBMI(weightKg, heightCM)
	if bmiValue <= 0 {
		return nil
	}

	bmiRounded := roundTo2Decimals(bmiValue)
	status := getBMIStatusValue(bmiRounded)
	if status == StatusNormal {
		return nil
	}

	value := fmt.Sprintf("BMI %.2f", bmiRounded)

	var alertType, label, explanation string
	var immediateActions, medicalAttention, managementTips []string

	if status == StatusRendah {
		alertType = "Berat Badan Tidak Normal"
		label = "Kurus"
		explanation = "Indeks Massa Tubuh Anda berada di bawah batas normal. Kondisi ini dapat menandakan kurangnya asupan nutrisi dan energi."
		immediateActions = []string{
			"Tingkatkan asupan kalori dengan makanan bergizi seimbang.",
			"Konsumsi camilan sehat di antara waktu makan.",
			"Perbanyak konsumsi protein dan karbohidrat kompleks.",
		}
		medicalAttention = []string{
			"Jika penurunan berat badan terjadi cepat tanpa sebab jelas.",
			"Jika disertai lemas, pusing, atau gejala lain.",
			"Konsultasikan dengan tenaga kesehatan untuk evaluasi menyeluruh.",
		}
		managementTips = []string{
			"Atur jadwal makan teratur dengan porsi cukup.",
			"Tambah porsi protein seperti telur, ikan, atau kacang-kacangan.",
			"Lakukan aktivitas fisik ringan untuk menjaga nafsu makan.",
		}
	} else {
		alertType = "Berat Badan Tidak Normal"
		label = "Obesitas"
		explanation = "Indeks Massa Tubuh Anda berada di atas batas normal dan dapat meningkatkan risiko gangguan kesehatan."
		immediateActions = []string{
			"Kurangi konsumsi makanan tinggi lemak dan gula.",
			"Perbanyak aktivitas fisik ringan.",
			"Pilih porsi makan lebih kecil namun sering.",
		}
		medicalAttention = []string{
			"Jika berat badan meningkat cepat dalam waktu singkat.",
			"Jika disertai sesak napas atau kelelahan berlebih.",
			"Konsultasikan dengan tenaga kesehatan untuk rencana penurunan berat badan.",
		}
		managementTips = []string{
			"Terapkan pola makan seimbang dengan sayur dan buah.",
			"Lakukan olahraga rutin minimal 30 menit per hari.",
			"Pantau berat badan secara berkala.",
			"Hindari minuman manis dan pilih air putih.",
		}
	}

	return &response.HealthAlertResponse{
		AlertType:        alertType,
		Category:         CategoryBeratBadan,
		Value:            value,
		Label:            label,
		Status:           status,
		RecordedAt:       timezoneUtils.ToJakarta(recordedAt),
		Explanation:      explanation,
		ImmediateActions: immediateActions,
		MedicalAttention: medicalAttention,
		ManagementTips:   managementTips,
		EducationVideos:  []response.EducationVideoItem{}, // Akan diisi di CheckHealthAlerts
	}
}

// getCategoryIDByCategory mengembalikan category_id berdasarkan kategori alert
func (s *HealthAlertService) getCategoryIDByCategory(category string) (uint, bool) {
	switch category {
	case CategoryDiabetes:
		return CategoryIDDiabetes, true
	case CategoryHipertensi:
		return CategoryIDHipertensi, true
	case CategoryJantung:
		return CategoryIDJantung, true
	case CategoryBeratBadan:
		return CategoryIDBeratBadan, true
	default:
		return 0, false
	}
}

// getEducationVideosByCategories mengambil video edukasi untuk multiple kategori sekaligus (batch query)
// Mengembalikan map category -> []EducationVideoItem untuk menghindari N+1 query
func (s *HealthAlertService) getEducationVideosByCategories(categorySet map[string]bool) map[string][]response.EducationVideoItem {
	result := make(map[string][]response.EducationVideoItem)

	// Kumpulkan category IDs yang valid
	var categoryIDs []uint
	categoryToID := make(map[string]uint)
	for category := range categorySet {
		categoryID, ok := s.getCategoryIDByCategory(category)
		if ok {
			categoryIDs = append(categoryIDs, categoryID)
			categoryToID[category] = categoryID
		} else {
			// Jika kategori tidak dikenal, set empty array
			result[category] = []response.EducationVideoItem{}
		}
	}

	// Jika tidak ada category ID yang valid, return empty map
	if len(categoryIDs) == 0 {
		return result
	}

	// Batch query semua video untuk semua kategori sekaligus
	videosByCategoryID, err := s.educationalVideoRepo.GetAllEducationalVideosByCategoryIDs(categoryIDs)
	if err != nil {
		// Jika error, return empty arrays untuk semua kategori
		for category := range categorySet {
			result[category] = []response.EducationVideoItem{}
		}
		return result
	}

	// Convert ke response format dan map berdasarkan category string
	for category, categoryID := range categoryToID {
		videos, exists := videosByCategoryID[categoryID]
		if !exists || len(videos) == 0 {
			result[category] = []response.EducationVideoItem{}
			continue
		}

		items := make([]response.EducationVideoItem, 0, len(videos))
		for _, video := range videos {
			items = append(items, response.EducationVideoItem{
				ID:         video.ID,
				VideoTitle: video.VideoTitle,
				VideoURL:   video.VideoURL,
				CategoryID: categoryID,
			})
		}
		result[category] = items
	}

	return result
}

// getEducationVideosByCategory mengambil video edukasi berdasarkan kategori alert (single category)
// Menggunakan category_id untuk mengambil video dari API education
// NOTE: Untuk multiple categories, gunakan getEducationVideosByCategories untuk menghindari N+1 query
func (s *HealthAlertService) getEducationVideosByCategory(category string) []response.EducationVideoItem {
	// Map kategori ke category_id
	categoryID, ok := s.getCategoryIDByCategory(category)
	if !ok {
		// Jika kategori tidak dikenal, return empty array
		return []response.EducationVideoItem{}
	}

	// Ambil video berdasarkan category_id
	videos, err := s.educationalVideoRepo.GetEducationalVideosByCategoryID(categoryID)
	if err != nil {
		// Jika error, return empty array (tidak mengganggu response utama)
		return []response.EducationVideoItem{}
	}

	// Jika tidak ada video, return empty array
	if len(videos) == 0 {
		return []response.EducationVideoItem{}
	}

	// Convert ke response format dengan category_id
	result := make([]response.EducationVideoItem, 0, len(videos))
	for _, video := range videos {
		result = append(result, response.EducationVideoItem{
			ID:         video.ID,
			VideoTitle: video.VideoTitle,
			VideoURL:   video.VideoURL,
			CategoryID: categoryID,
		})
	}

	return result
}
