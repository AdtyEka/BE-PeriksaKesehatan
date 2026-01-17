# BE-PeriksaKesehatan

Backend API untuk aplikasi Periksa Kesehatan yang dibangun dengan Go (Golang) menggunakan framework Gin. Aplikasi ini menyediakan berbagai fitur untuk manajemen data kesehatan pengguna, termasuk pencatatan data kesehatan, riwayat kesehatan, alert kesehatan, video edukasi, dan manajemen profil.

## 📋 Daftar Isi

- [Fitur](#fitur)
- [Teknologi yang Digunakan](#teknologi-yang-digunakan)
- [Persyaratan](#persyaratan)
- [Instalasi](#instalasi)
- [Konfigurasi](#konfigurasi)
- [Menjalankan Aplikasi](#menjalankan-aplikasi)
- [Struktur Proyek](#struktur-proyek)
- [API Endpoints](#api-endpoints)
- [Database Schema](#database-schema)
- [Testing](#testing)
- [Kontribusi](#kontribusi)

## ✨ Fitur

### Autentikasi
- **Registrasi** - Pendaftaran pengguna baru dengan validasi email dan username
- **Login** - Autentikasi pengguna dengan JWT token
- **Logout** - Logout dengan token blacklisting untuk keamanan

### Data Kesehatan
- **Input Data Kesehatan** - Pencatatan data kesehatan (tekanan darah, gula darah, berat badan, tinggi badan, detak jantung, aktivitas)
- **Lihat Data Terbaru** - Mengambil data kesehatan terbaru pengguna
- **Riwayat Kesehatan** - Melihat riwayat data kesehatan dengan filter waktu (7 hari, 1 bulan, 3 bulan, custom range)
- **Download Laporan PDF** - Mengunduh laporan kesehatan dalam format PDF
- **Analisis Data** - Summary, trend charts, dan status kesehatan

### Health Alerts
- **Pengecekan Alert** - Sistem otomatis mengecek kondisi kesehatan dan memberikan alert jika diperlukan
- **Kategori Alert** - Alert berdasarkan kategori (Diabetes, Hipertensi, Jantung, Berat Badan)

### Video Edukasi
- **Manajemen Video** - Menambah dan melihat video edukasi kesehatan
- **Kategori Video** - Video dikelompokkan berdasarkan kategori

### Profil Pengguna
- **Informasi Pribadi** - Manajemen data pribadi (nama, tanggal lahir, nomor telepon, alamat)
- **Foto Profil** - Upload dan update foto profil
- **Health Targets** - Set dan update target kesehatan
- **Settings** - Pengaturan akun pengguna

## 🛠 Teknologi yang Digunakan

- **Go 1.25.5** - Bahasa pemrograman
- **Gin** - Web framework
- **GORM** - ORM untuk database
- **PostgreSQL** - Database
- **JWT** - Autentikasi token
- **bcrypt** - Enkripsi password
- **gofpdf** - Generate PDF reports
- **godotenv** - Environment variable management

## 📦 Persyaratan

- Go 1.25.5 atau lebih tinggi
- PostgreSQL 12 atau lebih tinggi
- Git

## 🚀 Instalasi

1. **Clone repository**
```bash
git clone <repository-url>
cd BE-PeriksaKesehatan
```

2. **Install dependencies**
```bash
go mod download
```

3. **Setup database**
   - Buat database PostgreSQL baru
   - Atau gunakan database yang sudah ada

## ⚙️ Konfigurasi

1. **Buat file `.env` di root directory**
```env
DATABASE_URL=postgres://username:password@localhost:5432/dbname?sslmode=disable
PORT=8080
JWT_SECRET=your-secret-key-here-minimum-32-characters
```

2. **Penjelasan variabel environment:**
   - `DATABASE_URL` - Connection string untuk PostgreSQL (wajib)
   - `PORT` - Port untuk menjalankan server (default: 8080)
   - `JWT_SECRET` - Secret key untuk JWT token (wajib, minimal 32 karakter)

## ▶️ Menjalankan Aplikasi

1. **Pastikan database sudah berjalan dan file `.env` sudah dikonfigurasi**

2. **Jalankan aplikasi**
```bash
go run cmd/api/main.go
```

3. **Aplikasi akan berjalan di `http://localhost:8080`**

4. **API base URL:** `http://localhost:8080/api`

## 📁 Struktur Proyek

```
BE-PeriksaKesehatan/
├── cmd/
│   └── api/
│       └── main.go              # Entry point aplikasi
├── config/
│   └── config.go                # Konfigurasi aplikasi
├── internal/
│   ├── handler/                 # HTTP handlers
│   │   ├── auth_handler.go
│   │   ├── health_data_handler.go
│   │   ├── health_alert_handler.go
│   │   ├── educational_video_handler.go
│   │   ├── profile_handler.go
│   │   └── router.go            # Route definitions
│   ├── model/
│   │   ├── dto/                 # Data Transfer Objects
│   │   │   ├── request/         # Request DTOs
│   │   │   └── response/        # Response DTOs
│   │   └── entity/              # Database entities
│   ├── repository/              # Data access layer
│   │   ├── database.go          # Database initialization & migrations
│   │   ├── auth_repo.go
│   │   ├── user_repo.go
│   │   ├── health_data_repo.go
│   │   └── ...
│   └── service/                 # Business logic layer
│       ├── health_data_service.go
│       ├── health_alert_service.go
│       ├── educational_video_service.go
│       ├── profile_service.go
│       └── ...
├── pkg/
│   ├── middleware/              # HTTP middleware
│   │   └── auth_middleware.go
│   └── utils/                   # Utility functions
│       ├── response.go
│       ├── timezone.go
│       ├── file_upload.go
│       └── nullable_validator.go
├── uploads/                     # Uploaded files directory
│   └── profile/                 # Profile pictures
├── go.mod
├── go.sum
└── README.md
```

## 🔌 API Endpoints

### Autentikasi

#### Register
```
POST /api/auth/register
Content-Type: application/json

{
  "nama": "John Doe",
  "email": "john@example.com",
  "username": "johndoe",
  "password": "password123",
  "confirm_password": "password123"
}
```

#### Login
```
POST /api/auth/login
Content-Type: application/json

{
  "identifier": "john@example.com", // atau username
  "password": "password123"
}
```

#### Logout
```
POST /api/auth/logout
Authorization: Bearer <token>
```

### Data Kesehatan

#### Input Data Kesehatan
```
POST /api/health/data
Authorization: Bearer <token>
Content-Type: application/json

{
  "systolic": 120,
  "diastolic": 80,
  "blood_sugar": 100,
  "weight": 70,
  "height": 170,
  "heart_rate": 72,
  "activity": "Jalan pagi"
}
```

#### Get Data Kesehatan Terbaru
```
GET /api/health/data
Authorization: Bearer <token>
```

#### Get Riwayat Kesehatan
```
GET /api/health/history?time_range=7days
GET /api/health/history?time_range=1month
GET /api/health/history?time_range=3months
GET /api/health/history?time_range=custom&start_date=2025-01-01&end_date=2025-01-31
Authorization: Bearer <token>
```

#### Download Laporan PDF
```
GET /api/health/history/download?time_range=7days
Authorization: Bearer <token>
```

#### Check Health Alerts
```
GET /api/health/check-health-alerts
Authorization: Bearer <token>
```

### Video Edukasi

#### Tambah Video Edukasi
```
POST /api/education/add-educational-video
Content-Type: application/json

{
  "title": "Tips Menjaga Kesehatan Jantung",
  "description": "Video tentang cara menjaga kesehatan jantung",
  "video_url": "https://youtube.com/watch?v=...",
  "category_ids": [1, 2]
}
```

#### Get Semua Video Edukasi
```
GET /api/education/get-educational-videos
```

#### Get Video Edukasi by ID
```
GET /api/education/get-educational-videos/:id
```

### Profil

#### Get Profil
```
GET /api/profile
Authorization: Bearer <token>
```

#### Create/Update Profil
```
POST /api/profile
PUT /api/profile
Authorization: Bearer <token>
Content-Type: application/json

{
  "nama": "John Doe",
  "birth_date": "1990-01-01",
  "phone": "081234567890",
  "address": "Jl. Contoh No. 123",
  "profile_picture": "base64_encoded_image" // optional
}
```

#### Get Health Targets
```
GET /api/profile/health-targets
Authorization: Bearer <token>
```

#### Create/Update Health Targets
```
POST /api/profile/health-targets
PUT /api/profile/health-targets
Authorization: Bearer <token>
Content-Type: application/json

{
  "target_weight": 65,
  "target_blood_sugar": 100,
  "target_systolic": 120,
  "target_diastolic": 80
}
```

#### Get Settings
```
GET /api/profile/settings
Authorization: Bearer <token>
```

#### Update Settings
```
PUT /api/profile/settings
Authorization: Bearer <token>
Content-Type: application/json

{
  "notification_enabled": true,
  "language": "id"
}
```

## 🗄️ Database Schema

Aplikasi menggunakan PostgreSQL dengan tabel-tabel berikut:

- **users** - Data pengguna
- **health_data** - Data kesehatan pengguna
- **health_alerts** - Alert kesehatan
- **health_targets** - Target kesehatan pengguna
- **personal_infos** - Informasi pribadi pengguna
- **educational_videos** - Video edukasi
- **categories** - Kategori untuk alert dan video
- **educational_video_categories** - Relasi many-to-many video dan kategori
- **blacklisted_tokens** - Token yang sudah di-blacklist

Database migration akan berjalan otomatis saat aplikasi pertama kali dijalankan.

## 🔒 Keamanan

- Password di-hash menggunakan bcrypt
- JWT token untuk autentikasi
- Token blacklisting untuk logout
- Middleware autentikasi untuk protected routes
- Validasi input data
- Timezone handling (Asia/Jakarta)

## 📝 Catatan

- Aplikasi menggunakan timezone Asia/Jakarta untuk semua timestamp
- File upload disimpan di direktori `uploads/`
- PDF reports di-generate menggunakan gofpdf
- Database migrations berjalan otomatis saat startup
- Default categories (Diabetes, Hipertensi, Jantung, Berat Badan) akan di-seed otomatis

## 🤝 Kontribusi

Kontribusi sangat diterima! Silakan buat issue atau pull request.

## 📄 Lisensi

[Tambahkan informasi lisensi di sini]
