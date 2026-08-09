# GovMonitor IT — Enterprise IT Monitoring & Alerting System

🌎 **Language / Bahasa:** [English](#english) | [Bahasa Indonesia](#bahasa-indonesia)

---

<a name="english"></a>
# English Version

GovMonitor IT is an enterprise-grade IT infrastructure monitoring and incident alert system designed specifically for government networks. It features a concurrent, high-frequency poller engine that monitors network devices, detects outages, and dispatches real-time alerts.

The system is split into three main components:
1. **Go Backend**: High-performance API server and concurrent SNMP/ICMP poller engine.
2. **Vue 3 Frontend**: Modern, dynamic Single Page Application (SPA) dashboard.
3. **WhatsApp Sidecar**: A Node.js microservice utilizing the Baileys WhatsApp library for real-time dispatch, with automatic Telegram fallback.

---

## 🏛️ System Architecture Overview

```
                  ┌──────────────────────────────────────────────────┐
                  │                 Vue 3 Frontend                   │
                  │     (Vite + Pinia + TypeScript + Tailwind)       │
                  └──────────────┬───────────────────┬───────────────┘
                                 │ REST API          │ WebSockets (/ws)
                                 ▼                   ▼
                  ┌──────────────────────────────────────────────────┐
                  │                    Go Backend                    │
                  │          (API Handlers + Poller Engine)          │
                  └──────────────┬─────────────────┬─────────────────┘
                                 │                 │
                      ┌──────────┴───────┐     ┌───┴──────────────┐
                      │PostgreSQL Database│     │   Redis Cache    │
                      │(Persisted State) │     │& Pub/Sub Queue   │
                      └──────────────────┘     └──────────────────┘
                                 │
                                 ▼ [Primary Attempt]
                  ┌───────────────────────────────┐
                  │      Node.js wa-sidecar       │
                  │  (Baileys WhatsApp Web API)   │
                  └──────────────┬────────────────┘
                                 │ [If WhatsApp Fails / Fallback]
                                 ▼
                  ┌───────────────────────────────┐
                  │       Telegram Bot API        │
                  │      (Fallback Channel)       │
                  └───────────────────────────────┘
```

For in-depth explanations, diagrams (UML sequence, ERD, and use case diagrams), and operational details, refer to the [System Documentation Directory](./docs).

---

## 📂 Project Structure

```text
├── backend/                  # Go core source code
│   ├── cmd/                  # CLI commands & service entry points (API server, simulators)
│   ├── internal/             # Private application packages (handlers, worker pools, etc.)
│   ├── migrations/           # PostgreSQL migration files (sql schema versions)
│   ├── wa-sidecar/           # Node.js Baileys WhatsApp integration service
│   ├── go.mod                # Go module file
│   └── .gitignore            # Go & backend-specific ignores
│
├── frontend/                 # Vue 3 client application
│   ├── src/                  # Component and view source code
│   ├── index.html            # Entry HTML template
│   ├── vite.config.ts        # Vite configuration
│   └── .gitignore            # Frontend-specific ignores (dist, node_modules, envs)
│
├── docs/                     # Technical and architectural documentation set
├── dev-run.sh                # Shell script to start all 3 services concurrently
├── .gitignore                # Global workspace & IDE gitignore configuration
└── README.md                 # Project root documentation (this file)
```

---

## ⚙️ System Prerequisites

Make sure the following dependencies are installed on your host system:

| Software | Minimum Version | Verification Command |
|---|---|---|
| **Go** | v1.22+ | `go version` |
| **Node.js** | v18.0+ | `node -v` |
| **npm** | v9.0+ | `npm -v` |
| **PostgreSQL** | v14.0+ | `psql --version` |
| **Redis** | v6.0+ | `redis-cli ping` |

---

## 🚀 Setup & Local Installation

### 1. Environment Configurations (`.env`)

You need to configure the environment files for each of the three services. Copy the corresponding `.env.example` templates to `.env` in each respective folder and fill in your variables:

#### 1.1 Backend Configuration (`backend/.env`)
Create `backend/.env`:
```env
PORT=8080
ENV=development

# Database Settings
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=govmonitor
DB_SSLMODE=disable

# Redis Settings
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Security
JWT_SECRET=super-secret-jwt-key-govmonitor-2026

# WhatsApp Sidecar
SIDECAR_URL=http://localhost:3001
SIDECAR_SECRET=sidecar-internal-auth-token-2026

# Telegram Configuration (Fallback alert channel)
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
TELEGRAM_CHAT_ID=your_telegram_chat_id
```

#### 1.2 Frontend Configuration (`frontend/.env`)
Create `frontend/.env`:
```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/ws
```

#### 1.3 WhatsApp Sidecar Configuration (`backend/wa-sidecar/.env`)
Create `backend/wa-sidecar/.env`:
```env
PORT=3001
SIDECAR_SECRET=sidecar-internal-auth-token-2026
```

---

### 2. Database Migration & Seeding

Before running the application, prepare your PostgreSQL instance and run migrations.

#### Option A: Automatic Migration (Recommended)
On startup, the Go Backend automatically checks and runs any outstanding migrations located under `backend/migrations/` using `golang-migrate`.
1. Start your local PostgreSQL server and create a database named `govmonitor`:
   ```bash
   createdb -h localhost -p 5432 -U postgres govmonitor
   ```
2. Once the backend server starts (see running instructions below), it will apply the schema changes automatically.

#### Option B: Manual CLI Migration
Alternatively, run the migrations using `golang-migrate` command line:
```bash
migrate -path backend/migrations -database "postgres://postgres:postgres@localhost:5432/govmonitor?sslmode=disable" up
```

#### Default Seed Credentials
Migrations automatically seed the following accounts for immediate login:
* **Superadmin**: `admin.noc@jabarprov.go.id` / Password: `admin123`
* **Pimpinan (Executive)**: `sari.dewi@jabarprov.go.id` / Password: `admin123`
* **Anggota (NOC Staff)**: `rian.pratama@jabarprov.go.id` / Password: `admin123`

---

## 🏃 Running the Application

### Option 1: Concurrent Developer Script (Single Command)
If you are on Linux, macOS, or using Git Bash on Windows:
```bash
chmod +x dev-run.sh
./dev-run.sh
```
This script automatically verifies if `concurrently` is installed, installs it if missing, and boots all three servers simultaneously in one terminal window.

### Option 2: Step-by-Step Manual Run (Three Terminals)

If you prefer starting each service individually to inspect logs:

#### Terminal 1: Node.js WhatsApp Sidecar
Responsible for handling the WhatsApp session client and sending alerts.
```bash
cd backend/wa-sidecar
npm install
node index.js
```
*Expected console output: Listening on `http://localhost:3001`*

#### Terminal 2: Go Backend API & Poller
Responsible for serving REST API endpoints, WebSockets, and running the background device poller.
```bash
cd backend
go mod download
go run ./cmd/api
```
*Expected console output: Migrations applied successfully. Server listening on `:8080`, poller engine worker pool started.*

#### Terminal 3: Vue 3 Frontend
Serves the web client application.
```bash
cd frontend
npm install
npm run dev
```
*Expected console output: Vite dev server running at `http://localhost:5173`*

---

## 🔍 How it Works (Under the Hood)

1. **Poller Worker Pool**: The backend features a configurable concurrency worker pool. Every cycle, it triggers ICMP (ping) and SNMP queries for the devices registered in the inventory database.
2. **SNMP OS/System Name Discovery**: During SNMP check cycles, the poller queries OID `.1.3.6.1.2.1.1.5.0` to resolve and record the host device's OS/System Name (`snmpSysName`).
3. **Debounce and State Transitions**: If a device fails to respond, it enters a "Down" candidate state. If it remains down for a configured threshold of consecutive polls, it triggers a confirmed `DOWN` incident state transition.
4. **Bounded Flap-Reuse Window**: If a resolved device goes DOWN again within the configured `flapReuseWindowMinutes` setting, the engine re-opens the most recent resolved incident ticket instead of generating a duplicate, keeping the initial `started_at` outage timestamp pinned.
5. **Smart Notification Pipeline**: State transitions trigger the dispatch pipeline which queues notification tasks in Redis (`asynq`). The tasks are executed with spacing and rate limits. If WhatsApp Sidecar delivery fails, the pipeline automatically falls back to Telegram Bot API.
6. **Live Latency & Outage Visualizations**: Real-time metrics are streamed via WebSockets. On the Device Detail page, the latency chart leaves visual gaps (null values) during DOWN states and overlays red vertical band annotations (`xaxis` range annotations) labeled "DOWN" for clear outage visualization.
7. **Paging & Slicing Mechanics**: List tables (Dashboard, Devices, Incidents, Reports) enforce pagination controls. Flat lists query the backend directly using `?page=&page_size=`, whereas grouped views query the full dataset and perform pagination client-side. Searches and filters reactively reset the current page back to 1.
8. **Switchable Reports**: The Reports page segments data into switchable sections (Downtime by Device, Recurring Issues, Active Incidents) with independent pagination controls.
9. **WebSocket Updates**: Real-time polling events and device statuses are emitted on the WebSocket channel (`/ws`), updating the Vue 3 dashboard live without manual page reloads.

---

## 🛡️ Verification & Testing

1. Open your browser and navigate to `http://localhost:5173`.
2. Login with `admin.noc@jabarprov.go.id` and password `admin123`.
3. Check the WebSocket status indicator in the frontend top bar to ensure it shows `Connected`.
4. Register a test IP address (e.g. `127.0.0.1` or a known network target) in the device inventory and watch the live feed updates.

---
---

<a name="bahasa-indonesia"></a>
# Versi Bahasa Indonesia

GovMonitor IT adalah sistem pemantauan infrastruktur IT tingkat enterprise dan peringatan insiden yang dirancang khusus untuk jaringan pemerintah. Sistem ini memiliki mesin poller (pemantau) konkuren berfrekuensi tinggi yang memantau perangkat jaringan, mendeteksi gangguan, dan mengirimkan peringatan secara real-time.

Sistem ini dibagi menjadi tiga komponen utama:
1. **Go Backend**: API server berkinerja tinggi dan mesin poller SNMP/ICMP konkuren.
2. **Vue 3 Frontend**: Dashboard Single Page Application (SPA) yang modern dan dinamis.
3. **WhatsApp Sidecar**: Layanan mikro Node.js menggunakan pustaka Baileys WhatsApp untuk pengiriman real-time, dilengkapi dengan fallback otomatis ke Telegram.

---

## 🏛️ Gambaran Umum Arsitektur Sistem

```
                  ┌──────────────────────────────────────────────────┐
                  │                 Vue 3 Frontend                   │
                  │     (Vite + Pinia + TypeScript + Tailwind)       │
                  └──────────────┬───────────────────┬───────────────┘
                                 │ REST API          │ WebSockets (/ws)
                                 ▼                   ▼
                  ┌──────────────────────────────────────────────────┐
                  │                    Go Backend                    │
                  │          (API Handlers + Poller Engine)          │
                  └──────────────┬─────────────────┬─────────────────┘
                                 │                 │
                      ┌──────────┴───────┐     ┌───┴──────────────┐
                      │PostgreSQL Database│     │   Redis Cache    │
                      │(Persisted State) │     │& Pub/Sub Queue   │
                      └──────────────────┘     └───────────────────┘
                                 │
                                 ▼ [Upaya Utama (Primary)]
                  ┌───────────────────────────────┐
                  │      Node.js wa-sidecar       │
                  │  (Baileys WhatsApp Web API)   │
                  └──────────────┬────────────────┘
                                 │ [Jika WhatsApp Gagal / Fallback]
                                 ▼
                  ┌───────────────────────────────┐
                  │       Telegram Bot API        │
                  │      (Fallback Channel)       │
                  └───────────────────────────────┘
```

Untuk penjelasan mendalam, diagram (UML sequence, ERD, dan diagram use case), serta detail operasional, silakan merujuk ke [Direktori Dokumentasi Sistem](./docs).

---

## 📂 Struktur Proyek

```text
├── backend/                  # Kode sumber utama Go
│   ├── cmd/                  # CLI commands & service entry points (API server, simulator)
│   ├── internal/             # Paket privat aplikasi (handler, worker pool, dll.)
│   ├── migrations/           # File migrasi database PostgreSQL
│   ├── wa-sidecar/           # Layanan integrasi WhatsApp dengan Node.js Baileys
│   ├── go.mod                # File modul Go
│   └── .gitignore            # Gitignore khusus backend & Go
│
├── frontend/                 # Aplikasi klien Vue 3
│   ├── src/                  # Kode sumber komponen dan view
│   ├── index.html            # Templat HTML utama
│   ├── vite.config.ts        # Konfigurasi Vite
│   └── .gitignore            # Gitignore khusus frontend (dist, node_modules, dll.)
│
├── docs/                     # Kumpulan dokumentasi teknis dan arsitektur sistem
├── dev-run.sh                # Skrip shell untuk menjalankan ketiga layanan secara bersamaan
├── .gitignore                # Konfigurasi gitignore global workspace & IDE
└── README.md                 # Dokumentasi root proyek (file ini)
```

---

## ⚙️ Prasyarat Sistem

Pastikan dependensi berikut terinstal pada sistem host Anda:

| Perangkat Lunak | Versi Minimal | Perintah Verifikasi |
|---|---|---|
| **Go** | v1.22+ | `go version` |
| **Node.js** | v18.0+ | `node -v` |
| **npm** | v9.0+ | `npm -v` |
| **PostgreSQL** | v14.0+ | `psql --version` |
| **Redis** | v6.0+ | `redis-cli ping` |

---

## 🚀 Setup & Instalasi Lokal

### 1. Konfigurasi Lingkungan (`.env`)

Anda perlu mengonfigurasi file environment untuk masing-masing dari ketiga layanan. Salin templat `.env.example` yang sesuai menjadi `.env` di setiap folder masing-masing dan isi variabel Anda:

#### 1.1 Konfigurasi Backend (`backend/.env`)
Buat file `backend/.env`:
```env
PORT=8080
ENV=development

# Database Settings
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=govmonitor
DB_SSLMODE=disable

# Redis Settings
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Security
JWT_SECRET=super-secret-jwt-key-govmonitor-2026

# WhatsApp Sidecar
SIDECAR_URL=http://localhost:3001
SIDECAR_SECRET=sidecar-internal-auth-token-2026

# Telegram Configuration (Fallback alert channel)
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
TELEGRAM_CHAT_ID=your_telegram_chat_id
```

#### 1.2 Konfigurasi Frontend (`frontend/.env`)
Buat file `frontend/.env`:
```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/ws
```

#### 1.3 Konfigurasi WhatsApp Sidecar (`backend/wa-sidecar/.env`)
Buat file `backend/wa-sidecar/.env`:
```env
PORT=3001
SIDECAR_SECRET=sidecar-internal-auth-token-2026
```

---

### 2. Migrasi & Seeding Database

Sebelum menjalankan aplikasi, persiapkan instansi PostgreSQL Anda dan jalankan migrasi.

#### Opsi A: Migrasi Otomatis (Direkomendasikan)
Saat dijalankan pertama kali, Go Backend secara otomatis memeriksa dan menjalankan migrasi database yang tertunda di bawah folder `backend/migrations/` menggunakan `golang-migrate`.
1. Jalankan PostgreSQL lokal Anda dan buat database bernama `govmonitor`:
   ```bash
   createdb -h localhost -p 5432 -U postgres govmonitor
   ```
2. Setelah server backend dimulai (lihat instruksi menjalankan di bawah), server akan menerapkan perubahan skema secara otomatis.

#### Opsi B: Migrasi Manual via CLI
Alternatif lain, jalankan migrasi menggunakan command line `golang-migrate`:
```bash
migrate -path backend/migrations -database "postgres://postgres:postgres@localhost:5432/govmonitor?sslmode=disable" up
```

#### Kredensial Seed Default
Migrasi secara otomatis menyemai akun berikut agar dapat langsung digunakan untuk masuk:
* **Superadmin**: `admin.noc@jabarprov.go.id` / Sandi: `admin123`
* **Pimpinan (Eksekutif)**: `sari.dewi@jabarprov.go.id` / Sandi: `admin123`
* **Anggota (Staf NOC)**: `rian.pratama@jabarprov.go.id` / Sandi: `admin123`

---

## 🏃 Menjalankan Aplikasi

### Opsi 1: Skrip Developer Konkuren (Satu Perintah)
Jika Anda menggunakan Linux, macOS, atau menggunakan Git Bash di Windows:
```bash
chmod +x dev-run.sh
./dev-run.sh
```
Skrip ini secara otomatis memverifikasi apakah `concurrently` sudah terinstal, menginstalnya jika belum ada, dan menjalankan ketiga server secara bersamaan dalam satu jendela terminal.

### Opsi 2: Menjalankan Manual Langkah demi Langkah (Tiga Terminal)

Jika Anda lebih memilih menjalankan setiap layanan secara terpisah untuk memeriksa log:

#### Terminal 1: Node.js WhatsApp Sidecar
Bertanggung jawab untuk menangani sesi klien WhatsApp dan mengirimkan peringatan.
```bash
cd backend/wa-sidecar
npm install
node index.js
```
*Output konsol yang diharapkan: Listening on `http://localhost:3001`*

#### Terminal 2: Go Backend API & Mesin Pemantau (Poller)
Bertanggung jawab untuk melayani REST API endpoints, WebSockets, dan menjalankan mesin poller perangkat di latar belakang.
```bash
cd backend
go mod download
go run ./cmd/api
```
*Output konsol yang diharapkan: Migrations applied successfully. Server listening on `:8080`, poller engine worker pool started.*

#### Terminal 3: Vue 3 Frontend
Menyediakan aplikasi web klien.
```bash
cd frontend
npm install
npm run dev
```
*Output konsol yang diharapkan: Vite dev server running at `http://localhost:5173`*

---

## 🔍 Cara Kerja (Di Balik Layar)

1. **Poller Worker Pool**: Backend memiliki worker pool dengan konkurensi yang dapat dikonfigurasi. Setiap siklus, sistem memicu kueri ICMP (ping) dan SNMP untuk perangkat yang terdaftar di database inventaris.
2. **Pencarian Nama Sistem SNMP**: Selama siklus pemindaian SNMP, poller menanyakan OID `.1.3.6.1.2.1.1.5.0` untuk mendeteksi dan merekam OS / Nama Sistem perangkat host (`snmpSysName`).
3. **Debounce dan Transisi Status**: Jika perangkat gagal merespons, ia masuk ke status kandidat "Down". Jika tetap mati selama ambang batas polling berturut-turut yang dikonfigurasi, sistem memicu transisi status insiden `DOWN` yang terkonfirmasi.
4. **Jendela Flap Bounded**: Jika perangkat yang telah pulih mati kembali dalam jangka waktu **Flap Reuse Window (`flapReuseWindowMinutes`)**, mesin pemantau akan **membuka kembali** tiket insiden terdekat yang sudah terselesaikan daripada membuat tiket duplikat baru, menjaga `started_at` tetap tertambat ke awal insiden.
5. **Pipa Notifikasi Pintar**: Transisi status memicu pipa pengiriman yang memasukkan tugas ke antrean Redis (`asynq`). Pipa notifikasi ini berjalan dengan batasan kecepatan pesan (rate-limiting). Jika pengiriman WhatsApp Sidecar gagal, sistem otomatis mengirim pesan bot Telegram sebagai fallback.
6. **Visualisasi Latensi Live & Outage**: Metrik real-time dialirkan melalui WebSockets. Pada halaman detail perangkat, grafik latensi menampilkan celah kosong (nilai `null`) ketika perangkat mati (DOWN) dan menampilkan pita vertikal merah bertuliskan "DOWN" pada rentang waktu gangguan.
7. **Paginasi & Reset Halaman**: Tabel daftar perangkat/insiden/laporan dilengkapi paginasi. Tampilan flat (rata) meminta data langsung ke backend memakai parameter `?page=&page_size=`, sedangkan tampilan grup (grouped view) mengambil seluruh data lalu memotongnya di sisi klien. Reset otomatis ke halaman 1 berlaku saat pencarian/filter diubah.
8. **Laporan Tersegmentasi (Tab)**: Halaman laporan membagi visualisasi menjadi bagian-bagian yang dapat dialihkan lewat tab (Downtime by Device, Recurring Issues, Active Incidents) dengan kontrol paginasi independen untuk setiap tabel.
9. **Pembaruan WebSocket**: Event polling real-time dan status perangkat dikirimkan melalui saluran WebSocket (`/ws`), memperbarui dashboard Vue 3 secara langsung tanpa perlu memuat ulang halaman secara manual.

---

## 🛡️ Verifikasi & Pengujian

1. Buka browser Anda dan navigasikan ke `http://localhost:5173`.
2. Masuk menggunakan `admin.noc@jabarprov.go.id` dan sandi `admin123`.
3. Periksa indikator status WebSocket di bar atas frontend untuk memastikan statusnya menunjukkan `Connected`.
4. Daftarkan alamat IP uji (misalnya `127.0.0.1` atau target jaringan yang dikenal) di inventaris perangkat dan saksikan pembaruan feed secara langsung.
