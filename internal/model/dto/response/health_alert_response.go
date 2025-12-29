package response

import "time"

// HealthAlertResponse merepresentasikan satu kartu alert di UI.
type HealthAlertResponse struct {
	Type       string     `json:"type"`                 // "blood_pressure", "blood_sugar"
	Level      string     `json:"level"`                // "Normal", "Tinggi", "Kritis"
	Title      string     `json:"title"`                // "Tekanan Darah Anda"
	Subtitle   string     `json:"subtitle"`             // "Perhatian!"
	Value      string     `json:"value"`                // "150 / 95"
	Unit       string     `json:"unit"`                 // "mmHg", "mg/dL"
	Extra      *string    `json:"extra,omitempty"`      // optional (misal penurunan berat badan)
	StatusText string     `json:"status_text"`          // "Tekanan Darah Tinggi (Hipertensi)"
	RecordedAt time.Time  `json:"recorded_at"`          // waktu pengukuran
}
