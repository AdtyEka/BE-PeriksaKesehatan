package main

import (
	"BE-PeriksaKesehatan/config"
	"BE-PeriksaKesehatan/internal/handler"
	"BE-PeriksaKesehatan/internal/repository"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	log.Println("Info: Konfigurasi berhasil dimuat")

	db, err := repository.InitDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: Gagal menginisialisasi database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	router := handler.SetupRouter(cfg, userRepo)

	port := cfg.Port
	log.Printf("Info: Server berjalan di port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Fatal: Gagal menjalankan server: %v", err)
	}
}

