package service

import (
	"BE-PeriksaKesehatan/internal/model/dto/response"
	"BE-PeriksaKesehatan/internal/repository"
	"fmt"
	"time"
)

// Kategori konstan
const (
	CategoryDiabetes   = "diabetes"
	CategoryHipertensi = "hipertensi"
	CategoryJantung    = "jantung"
	CategoryBeratBadan = "berat_badan"
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

	return &response.CheckHealthAlertsResponse{
		Alerts: alerts,
	}, nil
}

// evaluateBloodPressure mengevaluasi tekanan darah dan mengembalikan alert jika tidak normal
func (s *HealthAlertService) evaluateBloodPressure(systolic, diastolic int, recordedAt time.Time) *response.HealthAlertResponse {
	status := s.getBloodPressureStatus(systolic, diastolic)
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

	// Ambil video edukasi
	videos := s.getEducationVideosByCategory(CategoryHipertensi)

	return &response.HealthAlertResponse{
		AlertType:        alertType,
		Category:         CategoryHipertensi,
		Value:            value,
		Label:            label,
		Status:           status,
		RecordedAt:       recordedAt,
		Explanation:      explanation,
		ImmediateActions: immediateActions,
		MedicalAttention: medicalAttention,
		ManagementTips:   managementTips,
		EducationVideos:  videos,
	}
}

// evaluateBloodSugar mengevaluasi gula darah dan mengembalikan alert jika tidak normal
func (s *HealthAlertService) evaluateBloodSugar(bloodSugar int, recordedAt time.Time) *response.HealthAlertResponse {
	status := s.getBloodSugarStatus(bloodSugar)
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

	// Ambil video edukasi
	videos := s.getEducationVideosByCategory(CategoryDiabetes)

	return &response.HealthAlertResponse{
		AlertType:        alertType,
		Category:         CategoryDiabetes,
		Value:            value,
		Label:            label,
		Status:           status,
		RecordedAt:       recordedAt,
		Explanation:      explanation,
		ImmediateActions: immediateActions,
		MedicalAttention: medicalAttention,
		ManagementTips:   managementTips,
		EducationVideos:  videos,
	}
}

// evaluateHeartRate mengevaluasi detak jantung dan mengembalikan alert jika tidak normal
func (s *HealthAlertService) evaluateHeartRate(heartRate int, recordedAt time.Time) *response.HealthAlertResponse {
	status := s.getHeartRateStatus(heartRate)
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

	// Ambil video edukasi
	videos := s.getEducationVideosByCategory(CategoryJantung)

	return &response.HealthAlertResponse{
		AlertType:        alertType,
		Category:         CategoryJantung,
		Value:            value,
		Label:            label,
		Status:           status,
		RecordedAt:       recordedAt,
		Explanation:      explanation,
		ImmediateActions: immediateActions,
		MedicalAttention: medicalAttention,
		ManagementTips:   managementTips,
		EducationVideos:  videos,
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

	videos := s.getEducationVideosByCategory(CategoryBeratBadan)

	return &response.HealthAlertResponse{
		AlertType:        alertType,
		Category:         CategoryBeratBadan,
		Value:            value,
		Label:            label,
		Status:           status,
		RecordedAt:       recordedAt,
		Explanation:      explanation,
		ImmediateActions: immediateActions,
		MedicalAttention: medicalAttention,
		ManagementTips:   managementTips,
		EducationVideos:  videos,
	}
}

// getBloodPressureStatus menentukan status tekanan darah berdasarkan kombinasi sistolik dan diastolik (WHO)
// RENDAH jika sistolik < 90 atau diastolik < 60
// NORMAL jika sistolik 90–139 dan diastolik 60–89
// TINGGI jika sistolik ≥ 140 atau diastolik ≥ 90
func (s *HealthAlertService) getBloodPressureStatus(systolic, diastolic int) string {
	if systolic < 90 || diastolic < 60 {
		return StatusRendah
	}
	if systolic >= 140 || diastolic >= 90 {
		return StatusTinggi
	}
	if systolic >= 90 && systolic <= 139 && diastolic >= 60 && diastolic <= 89 {
		return StatusNormal
	}
	// Fallback untuk kasus edge case
	return StatusNormal
}

// getBloodSugarStatus menentukan status gula darah sewaktu (WHO)
// RENDAH jika < 70 mg/dL
// NORMAL jika 70–140 mg/dL
// TINGGI jika > 140 mg/dL
func (s *HealthAlertService) getBloodSugarStatus(bloodSugar int) string {
	if bloodSugar < 70 {
		return StatusRendah
	}
	if bloodSugar > 140 {
		return StatusTinggi
	}
	return StatusNormal
}

// getHeartRateStatus menentukan status detak jantung
// Menggunakan logika yang sama untuk konsistensi
func (s *HealthAlertService) getHeartRateStatus(heartRate int) string {
	if heartRate >= 60 && heartRate <= 100 {
		return StatusNormal
	}
	if heartRate < 60 {
		return StatusRendah
	}
	return StatusTinggi
}

// getEducationVideosByCategory mengambil video edukasi berdasarkan kategori
func (s *HealthAlertService) getEducationVideosByCategory(category string) []response.EducationVideoItem {
	videos, err := s.educationalVideoRepo.GetEducationalVideosByCategoryKategori(category)
	if err != nil {
		// Jika error, return empty array (tidak mengganggu response utama)
		return []response.EducationVideoItem{}
	}

	result := make([]response.EducationVideoItem, 0, len(videos))
	for _, video := range videos {
		result = append(result, response.EducationVideoItem{
			ID:    video.ID,
			Title: video.VideoTitle,
			URL:   video.VideoURL,
		})
	}

	return result
}
