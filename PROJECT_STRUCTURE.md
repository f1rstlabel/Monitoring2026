# 📂 Struktur Project - GovMonitor IT

**GovMonitor IT** adalah sistem pemantauan infrastruktur IT (*IT Infrastructure Monitoring & Alerting System*) enterprise untuk jaringan pemerintah/organisasi. Sistem ini terdiri dari 3 komponen utama:
1. **Backend** (Go / Golang) - API Server, SNMP/ICMP Poller Concurrent Engine, Notifier.
2. **Frontend** (Vue 3 + Vite + TypeScript + Tailwind CSS) - Dashboard SPA & Real-time Monitoring.
3. **WhatsApp Sidecar** (Node.js + Baileys API) - Layanan mikro pengiriman notifikasi via WhatsApp dengan Telegram Bot Fallback.

---

## 🚀 Ringkasan Struktur Utama

```
vue project/
├── backend/            # Sub-sistem Server & API Engine (Go & Node.js Sidecar)
├── frontend/           # Aplikasi Web SPA (Vue 3 + Vite + TypeScript)
├── docs/               # Dokumentasi Arsitektur, Flow, ERD, Use-Case & Panduan
├── bin/                # Binary executable bawaan / terkompilasi
├── dev-run.sh          # Script runner untuk lingkungan pengembangan
├── README.md           # Dokumentasi utama proyek (English & Bahasa Indonesia)
└── PROJECT_STRUCTURE.md # Dokumen pemetaan struktur proyek ini
```

---

## 📁 Detail Struktur Direktori

### 1. `backend/` — Go Backend & WA Sidecar

Merupakan inti sistem pemantauan dan penyedia REST API serta WebSocket.

```
backend/
├── cmd/                        # Entry points / aplikasi yang dapat dieksekusi
│   ├── api/                    # Application API Server main package
│   ├── git_stub/               # Utility / Stub tools
│   ├── hashgen/                # Utility untuk pembuat hash password/keamanan
│   ├── simulator/              # Simulator perangkat/peristiwa jaringan
│   └── verifier/               # Tools verifikasi data / koneksi
├── internal/                   # Core business logic & internal Go modules
│   ├── config/                 # Konfigurasi aplikasi & pemuatan environment variables
│   ├── domain/                 # Struct & model data domain (models.go)
│   ├── handler/                # REST API endpoints & route handlers
│   │   ├── handlers.go         # Main API Handlers (Devices, Incidents, Reports, Users)
│   │   ├── import.go           # Bulk import handler (CSV/Excel)
│   │   ├── integrations.go     # Integrasi WhatsApp, Telegram, SMTP, Webhook
│   │   └── notifications.go    # Handler notifikasi manual/sistem
│   ├── middleware/             # Middleware HTTP (JWT Auth, CORS, Logger, Rate Limiter)
│   ├── notifier/               # Engine & antrian notifikasi (Pipeline, Queue, Asynq)
│   ├── poller/                 # Engine pemantau status jaringan (ICMP Ping / SNMP)
│   │   ├── engine.go           # High-concurrency polling engine
│   │   └── arp.go              # Discovery & ARP check helper
│   ├── repository/             # Abstraksi database & query layer
│   │   ├── interfaces.go       # Kontrak interface repository
│   │   ├── pg_repo.go          # Implementasi PostgreSQL DB
│   │   ├── memory.go           # In-memory store untuk testing/caching cepat
│   │   ├── postgres.go         # Koneksi DB PostgreSQL
│   │   ├── redis.go            # Koneksi & Pub/Sub Redis
│   │   └── settings_repo.go    # Pengelolaan setting aplikasi di DB
│   ├── scheduler/              # Cron job & penjatwalan tugas periodik
│   ├── simulator/              # Logical simulation engine
│   ├── worker/                 # Worker task background
│   └── ws/                     # WebSocket Manager & Real-time Client Broadcast
├── migrations/                 # File migrasi skema SQL PostgreSQL
├── wa-sidecar/                 # Service Notifikasi WhatsApp (Node.js)
│   ├── index.js                # Baileys WhatsApp client & HTTP API bridge
│   ├── package.json            # Dependensi Node.js (baileys, express, pino)
│   └── auth_sessions/          # Sesi login & kredensial WhatsApp Multi-Device
├── .env / .env.example         # File variabel lingkungan backend
├── go.mod / go.sum             # Modul & dependensi Go
└── *.exe                       # Executable binary terkompilasi (api.exe, server.exe, dll)
```

---

### 2. `frontend/` — Vue 3 SPA Dashboard

Antarmuka pengguna berbasis web modern berbasis Vue 3, Vite, TypeScript, dan Tailwind CSS.

```
frontend/
├── src/                        # Kode sumber frontend
│   ├── api/                    # Client HTTP (Axios) & Service API Calls
│   ├── components/             # Komponen Reusable UI
│   │   ├── common/             # Button, Modal, Badge, Card, Navbar, Sidebar
│   │   ├── dashboard/          # DeviceCard, LiveFeedPanel, MetricWidget
│   │   ├── devices/            # Form tambah/edit device, filter, status table
│   │   ├── notifications/      # Panel notifikasi & log insiden
│   │   ├── reports/            # Grafik, tabel ekspor laporan, SLA chart
│   │   ├── settings/           # Form konfigurasi notifikasi, user management
│   │   └── users/              # Manajemen profil & peran pengguna
│   ├── router/                 # Vue Router (Navigasi & Auth Guard)
│   ├── stores/                 # State Management (Pinia Stores)
│   │   ├── authStore.ts        # Status login, token, & hak akses user
│   │   ├── deviceStore.ts      # Data daftar perangkat & status polling
│   │   ├── incidentStore.ts    # Management log insiden & alert active
│   │   ├── liveStore.ts        # Real-time WebSocket event feed
│   │   ├── notificationStore.ts# Log notifikasi WA/Telegram/Email
│   │   ├── reportStore.ts      # Data analitik, SLA & statistik
│   │   └── settingStore.ts     # Konfigurasi sistem & integrasi kanal
│   ├── styles/                 # Global CSS & Custom Tailwind utilities
│   ├── types/                  # TypeScript Interfaces & Data Types
│   ├── views/                  # Halaman Utama (Pages/Screens)
│   │   ├── DashboardView.vue   # Halaman ringkasan status utama & peta/grid
│   │   ├── DevicesView.vue     # Daftar & inventaris perangkat
│   │   ├── DeviceDetailView.vue# Detail metrics, latency & history perangkat
│   │   ├── IncidentsView.vue   # Log insiden & histori outage
│   │   ├── IncidentDetailView.vue # Detail analisis & resolusi insiden
│   │   ├── ReportsView.vue     # Laporan kinerja, uptime SLA & PDF/CSV export
│   │   ├── SettingsView.vue    # Pengaturan umum, saluran notifikasi & pengguna
│   │   └── LoginView.vue       # Halaman Masuk / Autentikasi
│   ├── ws/                     # WebSocket Client Handler & reconnect logic
│   ├── App.vue                 # Komponen Root Aplikasi Vue
│   └── main.ts                 # Entry point JavaScript/TypeScript Vue
├── index.html                  # Template HTML Utama
├── package.json                # Dependensi NPM & Script (vite, vue, pinia, tailwind)
├── tailwind.config.js          # Konfigurasi Desain & Tema Tailwind CSS
├── tsconfig.json               # Konfigurasi TypeScript Compiler
└── vite.config.ts              # Konfigurasi Vite Bundler & Proxy Server
```

---

### 3. `docs/` — Dokumentasi Teknis Sistem

Kumpulan dokumen perancangan dan standar arsitektur terstruktur.

```
docs/
├── 01-architecture.md           # Arsitektur sistem menyeluruh & diagram komponen
├── 02-usage-guide.md            # Panduan pengoperasian & penggunaan aplikasi
├── 03-nfr-nfsr.md              # Non-Functional Requirements & Security
├── 04-use-case-diagram.md      # Diagram Use Case aktor & fungsi sistem
├── 05-flowchart.md             # Flowchart alur kerja jaringan & alert
├── 06-business-flow.md         # Process flow bisnis & eskalasi insiden
├── 07-activity-diagram-swimlane.md # Activity diagram per peran (Poller, Admin, WA)
├── 08-sequence-diagram.md      # Sequence diagram interaksi API & Polling
├── 09-class-diagram.md         # Class & Interface diagram backend Go
├── 10-erd.md                   # Entity Relationship Diagram (Database PostgreSQL)
├── 11-data-and-memory-management.md # Pengelolaan memori & cache Redis
├── 12-installation.md          # Panduan instalasi lokal & prasyarat
├── 13-docker-deployment.md     # Panduan kompilasi & kontainerisasi Docker
├── 14-git-branching-workflow.md# Standar percabangan Git & alur kerja tim
└── README.md                   # Indeks dokumentasi teknis
```

---

## 🛠️ Alur Komunikasi Antar Komponen

1. **Polling Engine (`backend/internal/poller`)**: Secara berkala melakukan cek ICMP Ping & SNMP ke IP Perangkat.
2. **Database & Cache (`PostgreSQL` & `Redis`)**: Menyimpan status perangkat, riwayat insiden, serta log latensi.
3. **Notification Pipeline (`backend/internal/notifier`)**: Jika perangkat terdeteksi *Down*, insiden dibuat dan dikirim ke `wa-sidecar` (Node.js) untuk pengiriman WhatsApp. Jika WhatsApp gagal, otomatis fallback ke Telegram Bot API.
4. **WebSocket & REST API (`backend/cmd/api` & `frontend/src/ws`)**: Memancarkan update status secara real-time ke UI Dashboard Vue 3.
