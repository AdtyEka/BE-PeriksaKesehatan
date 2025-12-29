package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config menampung semua variabel lingkungan
type Config struct {
	DBURL     string
	Port      string
	JWTSecret string
}

// LoadConfig akan membaca file .env dan memasukkannya ke struct Config
func LoadConfig() *Config {
	// 1. Load file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Info: File .env tidak ditemukan, menggunakan environment variable sistem")
	}

	// 2. Ambil nilai dari os.Getenv
	dbURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	jwtSecret := os.Getenv("JWT_SECRET")

	// 3. Beri nilai default jika PORT kosong
	if port == "" {
		port = "8080"
	}

	// 4. Validasi kritikal: Jika DATABASE_URL kosong, aplikasi harus berhenti
	if dbURL == "" {
		log.Fatal("Fatal: DATABASE_URL tidak ditemukan di .env atau environment variable")
	}

	// 5. Validasi kritikal: Jika JWT_SECRET kosong, aplikasi harus berhenti
	if jwtSecret == "" {
		log.Fatal("Fatal: JWT_SECRET tidak ditemukan di .env atau environment variable")
	}

	return &Config{
		DBURL:     dbURL,
		Port:      port,
		JWTSecret: jwtSecret,
	}
}