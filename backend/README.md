# SANOC Backend Service (Golang)

Backend service untuk aplikasi **SANOC (Sanditel Network Operations Center)** yang dibangun menggunakan bahasa pemrograman Go (Golang 1.22+) dengan framework Gin, PostgreSQL, Redis, Asynq task queue, ICMP/SNMP Polling engine, dan Gmail Global SMTP Gateway.

---

## 🛠️ Tech Stack & Key Components

- **Language**: Go 1.22+
- **HTTP Framework**: [Gin Web Framework](https://github.com/gin-gonic/gin)
- **Database**: PostgreSQL (Driver: `lib/pq` / `sql`)
- **Caching & Rate Limiting**: Redis
- **Background Worker**: Asynq (Go Distributed Task Queue)
- **SMTP Gateway**: Gmail Global SMTP (`smtp.gmail.com:587`) & Local Mailpit support
- **Authentication**: JWT (JSON Web Tokens), HttpOnly Secure Cookies, RFC 6238 TOTP (2FA/MFA), Real Email OTP Verification
- **Security Protections**: Rate Limiting per IP, Security Headers, CSRF Token Verification, MIME Detection File Upload Validation.

---

## 📂 Backend Project Structure

```text
backend/
├── cmd/
│   └── api/
│       └── main.go           # Entry point aplikasi backend
├── internal/
│   ├── config/               # Struct & pemuat konfigurasi (.env)
│   ├── domain/               # Struct model data (User, Device, Incident, dll)
│   ├── handler/              # HTTP Request Handlers (Auth, Devices, Incidents, Reports, Integrations, Users)
│   ├── mailer/               # SMTP Mailer Service (OTP Email Verification & Password Reset)
│   ├── middleware/           # Security, Auth JWT, CSRF, Rate Limiter, & Logging
│   ├── notifier/             # Asynq Task Queue Dispatcher & Formatter
│   ├── poller/               # ICMP Ping & SNMP Core Switch Polling Engine
│   └── repository/           # Interface & implementasi PostgreSQL (pg_repo.go)
├── migrations/               # File migrasi SQL (golang-migrate)
└── wa-sidecar/               # Node.js Baileys WhatsApp Gateway Sidecar
```

---

## 🚀 Quick Start (Development)

### 1. Prasyarat System
- Go 1.22 atau lebih baru
- PostgreSQL 14+ running di `localhost:5432`
- Redis Server running di `localhost:6379`

### 2. Konfigurasi Environment File
Salin file `.env.example` ke `.env`:

```bash
cp .env.example .env
```

Sesuaikan kredensial PostgreSQL dan JWT Secret di `.env`.

### 3. Jalankan Migrasi Database
Jalankan file SQL yang berada di folder `migrations/` pada database PostgreSQL `sanoc`.

### 4. Jalankan Server API
```bash
go run ./cmd/api
```

Server API backend akan berjalan secara lokal pada port `http://localhost:8080`.

---

## 🧪 Testing & Security Audits

Jalankan unit tests & handler tests:
```bash
go test -v ./internal/middleware ./internal/handler
```

Jalankan audit kerentanan kode:
```bash
govulncheck ./...
```
