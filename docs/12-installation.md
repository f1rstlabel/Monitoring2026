# 12 — Local Installation Guide / Panduan Instalasi Lokal

This document provides step-by-step instructions for installing, configuring, and running **SANOC v2.6.0** locally for development and testing.  
*Panduan instalasi dan menjalankan SANOC di lingkungan lokal.*

---

## 1. System Prerequisites / Prasyarat Sistem

| Software | Minimum Version | Verification Command |
|---|---|---|
| **Go** | v1.22+ | `go version` |
| **Node.js** | v18.0+ | `node -v` |
| **npm** | v9.0+ | `npm -v` |
| **PostgreSQL** | v14.0+ | `psql --version` |
| **Redis** | v6.0+ | `redis-cli ping` |

---

## 2. Environment Configuration (`.env`)

Copy `.env.example` templates to `.env` in both backend and frontend directories:

### 2.1 Backend Environment (`backend/.env`)
```bash
cd backend
cp .env.example .env
```
Ensure key environment variables are set in `backend/.env`:
```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_local_db_password
DB_NAME=sanoc
DB_SSLMODE=disable

REDIS_HOST=localhost
REDIS_PORT=6379

JWT_SECRET=your_local_jwt_secret_key
CORS_ALLOWED_ORIGIN=http://localhost:5173

# WhatsApp Baileys Sidecar
WHATSAPP_GATEWAY_URL=http://localhost:3001
WA_SIDECAR_TOKEN=sanoc-sidecar-secret-2026

# Email / SMTP Gateway (Gmail Global SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_16_char_app_password
SMTP_FROM=SANOC Jabar Monitoring <sanoc.noreply@gmail.com>
```

### 2.2 Frontend Environment (`frontend/.env`)
```bash
cd frontend
cp .env.example .env
```
Set Vite API endpoint:
```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

---

## 3. Database Migration & Execution / Migrasi & Jalankan

### 3.1 Run Database Migrations
```bash
cd backend
go run ./cmd/api --migrate
```

### 3.2 Start Services
1. **WhatsApp Web Sidecar**:
   ```bash
   cd backend/wa-sidecar
   npm install
   node index.js
   ```
2. **Backend API (with hot reload)**:
   ```bash
   cd backend
   air
   ```
3. **Frontend Vue 3 App**:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
4. Open browser at `http://localhost:5173`.
