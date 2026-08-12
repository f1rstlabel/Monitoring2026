# SANOC Frontend SPA (Vue 3 + TypeScript + Tailwind CSS)

Aplikasi antarmuka pengguna (Single Page Application) untuk **SANOC (Smart Alert & Network Operation Center)** yang dibangun menggunakan Vue 3, Vite, TypeScript, Pinia, dan Tailwind CSS.

---

## 🛠️ Tech Stack & Utilities

- **Framework**: Vue 3 (Composition API `<script setup lang="ts">`)
- **Build Tool**: Vite 6
- **State Management**: Pinia
- **Routing**: Vue Router 4
- **Styling**: Tailwind CSS
- **Icons**: Lucide Vue Next
- **HTTP Client**: Axios dengan Interceptor (CSRF & JWT token auto-attachment)
- **Charts & Visualization**: Chart.js / Vue-Chartjs / ECharts

---

## 📂 Frontend Project Structure

```text
frontend/
├── src/
│   ├── api/                  # Axios HTTP client & API service endpoints
│   ├── assets/               # Assets statis (SVG logo, icons)
│   ├── components/           # UI Reusable Components (Common, Devices, Incidents, Reports, Settings)
│   ├── router/               # Route definitions & Navigation Guards
│   ├── stores/               # Pinia Stores (authStore, deviceStore, incidentStore, settingStore)
│   ├── views/                # Views / Top-level Pages (Dashboard, Devices, Incidents, Reports, Profile, Login)
│   ├── App.vue               # Root App Layout
│   └── main.ts               # Entry point Vue app
├── index.html                # HTML entry point
├── package.json              # Dependencies & NPM scripts
└── vite.config.ts            # Configuration Vite
```

---

## 🚀 Quick Start (Development)

### 1. Install Dependencies
```bash
npm install
```

### 2. Konfigurasi Environment File
Salin file `.env.example` ke `.env`:

```bash
cp .env.example .env
```

Isikan nilai variabel:
```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

### 3. Jalankan Dev Server
```bash
npm run dev
```
Aplikasi akan tersedia di `http://localhost:5173`.

### 4. Build Kompilasi Produksi
```bash
npm run build
```
File hasil kompilasi statis akan disimpan di folder `dist/`.
