package utils

import (
	"time"
)

// TimezoneAsiaJakarta adalah timezone untuk Indonesia (WIB)
var TimezoneAsiaJakarta *time.Location

func init() {
	// Load timezone Asia/Jakarta sekali saat package di-load
	var err error
	TimezoneAsiaJakarta, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback ke UTC+7 jika LoadLocation gagal (untuk kompatibilitas)
		TimezoneAsiaJakarta = time.FixedZone("WIB", 7*60*60)
	}
}

// NowInJakarta mengembalikan waktu saat ini dalam timezone Asia/Jakarta
// Menggantikan time.Now() untuk konsistensi timezone di seluruh aplikasi
func NowInJakarta() time.Time {
	return time.Now().In(TimezoneAsiaJakarta)
}

// DateInJakarta membuat time.Time dengan timezone Asia/Jakarta
// Menggantikan time.Date untuk konsistensi timezone
func DateInJakarta(year int, month time.Month, day, hour, min, sec, nsec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, nsec, TimezoneAsiaJakarta)
}

// ToJakarta mengkonversi time.Time ke timezone Asia/Jakarta
// Digunakan untuk konversi timezone pada data yang sudah ada
func ToJakarta(t time.Time) time.Time {
	return t.In(TimezoneAsiaJakarta)
}
