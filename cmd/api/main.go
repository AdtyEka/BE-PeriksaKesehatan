package main

import (
	"BE-PeriksaKesehatan/config"
	"BE-PeriksaKesehatan/internal/handler"
	"BE-PeriksaKesehatan/internal/repository"
	"log"
)

func main() {
	// 1. Load konfigurasi dari .env
	cfg := config.LoadConfig()
	log.Println("Info: Konfigurasi berhasil dimuat")

	// 2. Initialize koneksi database
	db, err := repository.InitDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: Gagal menginisialisasi database: %v", err)
	}

	// 3. Setup repository
	userRepo := repository.NewUserRepository(db)

	// 4. Setup router dengan dependency injection
	router := handler.SetupRouter(userRepo)

	// 5. Start server
	port := cfg.Port
	log.Printf("Info: Server berjalan di port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Fatal: Gagal menjalankan server: %v", err)
	}
}

