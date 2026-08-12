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
Ensure key environment variables are set:
```env
SANOC_PORT=8080
SANOC_DB_HOST=localhost
SANOC_DB_PORT=5432
SANOC_DB_USER=postgres
SANOC_DB_PASSWORD=your_local_db_password
SANOC_DB_NAME=sanoc
SANOC_JWT_SECRET=your_local_jwt_secret_key
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
1. **Backend API**:
   ```bash
   cd backend
   go run ./cmd/api
   ```
2. **Frontend Vue 3 App**:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
3. Open browser at `http://localhost:5173`.
