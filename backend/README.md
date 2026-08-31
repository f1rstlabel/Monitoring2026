# SANOC - Backend

Bagian backend dari proyek SANOC dibangun menggunakan bahasa Go (Golang). Backend ini bertugas untuk melayani API Frontend, menjalankan proses *background polling* untuk mengecek status ICMP pada perangkat jaringan, mengelola notifikasi (Telegram & WhatsApp), dan mengintegrasikan Copilot (Gemini AI).

## Tech Stack

- **Bahasa Pemrograman**: Go 1.22+
- **Framework Web**: [Gin](https://gin-gonic.com/) / Standar net/http (bergantung implementasi routing)
- **Database Utama**: PostgreSQL (via *database/sql* atau ORM)
- **Message Broker / Cache**: Redis
- **WhatsApp Gateway**: *Custom Sidecar* berbasis Node.js (`wa-sidecar`)
- **AI Integration**: Google Gemini API SDK

## Instalasi dan Setup

1. **Persyaratan Sistem**
   - Go 1.22 atau lebih baru.
   - PostgreSQL (Database harus sudah berjalan).
   - Redis Server (Harus sudah berjalan).
   - Node.js (Khusus untuk menjalankan WA Sidecar).

2. **Konfigurasi Environment (*.env*)**
   Salin atau buat file `.env` di dalam direktori `backend/` dan sesuaikan nilainya:
   ```env
   # Database Configuration
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=your_password
   DB_NAME=sanoc
   DB_SSLMODE=disable

   # Redis Configuration
   REDIS_HOST=localhost
   REDIS_PORT=6379

   # Authentication & Security
   JWT_SECRET=your_jwt_secret_key
   JWT_EXPIRY=24h
   RECAPTCHA_ENABLED=true
   RECAPTCHA_SECRET_KEY=your_recaptcha_secret_key

   # WhatsApp & Telegram Notifications
   WHATSAPP_GATEWAY_URL=http://localhost:3001
   TELEGRAM_BOT_TOKEN=your_telegram_bot_token
   TELEGRAM_CHAT_ID=your_telegram_chat_id

   # AI Integration
   GEMINI_API_KEY=your_gemini_api_key
   ```

3. **Menjalankan Backend (Golang API)**
   Unduh modul Go yang diperlukan dan jalankan aplikasi:
   ```bash
   cd backend
   go mod tidy
   go run ./cmd/api
   ```
   Secara *default*, server backend berjalan di port `8080`.

4. **Menjalankan WhatsApp Sidecar**
   Buka jendela terminal terpisah untuk menjalankan *sidecar* pengirim pesan WhatsApp:
   ```bash
   cd backend/wa-sidecar
   npm install
   node index.js
   ```
   *Sidecar* WhatsApp secara *default* akan berjalan di port `3001`.

## Arsitektur & Direktori Utama

- `cmd/api/`: Entri (*entry-point*) utama dari aplikasi. Tempat *server listener* berjalan.
- `internal/handler/`: *Controller* logika aplikasi, menangani request HTTP dan WebSocket.
- `internal/domain/`: Struktur model (*structs*) yang menjadi entitas logika bisnis utama.
- `wa-sidecar/`: *Service* independen (Node.js) untuk menjembatani komunikasi ke ekosistem WhatsApp (menggunakan Baileys/Puppeteer).
