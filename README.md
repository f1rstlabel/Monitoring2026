# SANOC — Smart Agent and Network Operations Center

> **Enterprise System Monitoring & SLA Compliance Audit Platform — Infrastructure NOC**  
> *Sistem Monitoring Infrastruktur Jaringan & Audit Kepatuhan SLA Real-time*  
> **Diskominfo Pemerintah Daerah Provinsi Jawa Barat — 2026**

---

## 🌐 Dokumentasi & Panduan Lengkap (Documentation)

Daftar modul panduan teknis yang tersedia di folder [`/docs`](./docs/README.md):

- 🏛️ **[01 — Arsitektur Sistem (System Architecture)](./docs/01-architecture.md)**
- 📖 **[02 — Panduan Operasional & Penggunaan (Usage Guide)](./docs/02-usage-guide.md)**
- ⚡ **[03 — Kebutuhan Non-Fungsional & Keamanan (NFR/NFSR)](./docs/03-nfr-nfsr.md)**
- 👥 **[04 — Diagram Use Case (Use Case Diagram)](./docs/04-use-case-diagram.md)**
- 🔀 **[05 — Diagram Alur Polling & Notifikasi (Flowchart)](./docs/05-flowchart.md)**
- 🔄 **[06 — Alur Proses Bisnis NOC (Business Flow)](./docs/06-business-flow.md)**
- 🏊 **[07 — Diagram Aktivitas Swimlane (Activity Diagram)](./docs/07-activity-diagram-swimlane.md)**
- ⏱️ **[08 — Diagram Sekuensial (Sequence Diagram)](./docs/08-sequence-diagram.md)**
- 🧱 **[09 — Diagram Kelas Backend (Class Diagram)](./docs/09-class-diagram.md)**
- 🗄️ **[10 — Skema Relasi Database (ERD)](./docs/10-erd.md)**
- 💾 **[11 — Manajemen Data & Memori (Data Management)](./docs/11-data-and-memory-management.md)**
- 🛠️ **[12 — Panduan Instalasi Lokal (Installation)](./docs/12-installation.md)**
- 🐳 **[13 — Pengerahan Menggunakan Docker (Docker Deployment)](./docs/13-docker-deployment.md)**
- 🌿 **[14 — Alur Kerja Git Branching (Git Workflow)](./docs/14-git-branching-workflow.md)**
- 🔐 **[15 — Keamanan & Konfigurasi TOTP MFA (Security & MFA)](./docs/15-recaptcha-setup.md)**
- 🚀 **[16 — Panduan Pengerahan Server Produksi (Production Deployment)](./docs/16-deployment.md)**
- 📧 **[17 — Konfigurasi Email SMTP & OTP Produksi (SMTP Gateway)](./docs/17-smtp-email-configuration.md)**

---

## 🇬🇧 English Documentation

### 📌 Project Overview
**SANOC (Smart Agent and Network Operations Center)** is an enterprise-grade network infrastructure monitoring and SLA compliance auditing system developed for **Diskominfo Pemerintah Daerah Provinsi Jawa Barat**. It provides real-time automated ICMP ping and SNMP polling across regional network nodes, switches, routers, access points, and server devices.

Key capabilities include:
- **Real-Time Monitoring**: Automated ICMP & SNMP polling engine with sub-minute failure detection and debounce logic.
- **Incident & Outage Management**: Automated incident ticket generation, live outage duration tracking, and incident resolution workflows.
- **Multi-Channel Alert Gateway**: Real-time broadcast alerts via Baileys WhatsApp Sidecar and Telegram Bot.
- **AI Copilot Assistant**: In-system Google Gemini AI agent assisting technicians with root cause analysis (RCA), diagnostic guidance, and executive summaries.
- **Live SLA & Availability Analytics**: Dynamic SLA metrics, Mean Time to Recovery (MTTR), and scheduled compliance reports.
- **Enterprise Security**: Role-based access control (RBAC), Time-based One-Time Password (TOTP) MFA, and email verification.

---

### 📦 Technology Stack

- **Backend**: Go (Golang 1.22+), Gin Web Framework, PostgreSQL 14+, Redis & Asynq (queue & rate limiting), JWT Authentication, TOTP MFA (`pquerna/otp`), Gmail SMTP Gateway.
- **Sidecar**: Node.js 18+ with `@whiskeysockets/baileys` for native WhatsApp Web multi-device connectivity.
- **Frontend**: Vue 3 (Composition API), Vite, TypeScript, Pinia State Store, TailwindCSS, Lucide Icons.
- **Database Migrations**: `golang-migrate`.

---

### 💻 Quick Start & Local Setup

#### 1. Backend Setup
```bash
cd backend

# Copy environment template
cp .env.example .env

# Run database migrations
go run ./cmd/api --migrate

# Start Backend API Server
air
```

#### 2. WhatsApp Sidecar Setup
```bash
cd backend/wa-sidecar
npm install
node index.js
```

#### 3. Frontend Setup
```bash
cd frontend
npm install
npm run dev
```
Open browser at `http://localhost:5173`.

---

<br/>

---

## 🇮🇩 Dokumentasi Bahasa Indonesia

### 📌 Penjelasan Project
**SANOC (Smart Agent and Network Operations Center)** adalah sistem pemantauan infrastruktur jaringan berbasis *enterprise* dan pengauditan kepatuhan SLA (*Service Level Agreement*) resmi **Diskominfo Pemerintah Daerah Provinsi Jawa Barat**. Sistem ini secara otomatis memonitor ketersediaan perangkat jaringan (router, switch, access point, server, CCTV) menggunakan protokol ICMP Ping dan SNMP Polling secara real-time.

Fitur Unggulan:
- **Monitoring Real-time**: Engine polling ICMP & SNMP otomatis dengan goroutines performa tinggi dan deteksi gangguan dalam hitungan detik.
- **Manajemen Incident & Outage**: Pembuatan tiket kejadian otomatis, pencatatan durasi outage *live*, dan workflow penyelesaian incident.
- **Pengiriman Notifikasi Multi-Kanal**: Pengiriman pesan peringatan seketika via WhatsApp API Gateway (broadcast ke semua target nomor teknisi & grup) dan fallback ke Telegram Bot.
- **Verifikasi Email Asli & OTP**: Validasi alamat email real dengan pengiriman 6-digit kode OTP untuk pembuatan akun baru dan reset password profil.
- **Laporan Kepatuhan SLA**: Perhitungan persentase Uptime SLA otomatis, pencatatan MTTR (*Mean Time to Recovery*), serta pilihan ekspor dokumen (PDF A4 full-width, Excel `.xls`, CSV `.csv`).
- **Keamanan Tingkat Tinggi**: Autentikasi Dua Faktor (TOTP MFA via Google Authenticator / Authy), Hak Akses Berbasis Role (Admin, Pimpinan, Anggota SANOC), serta validasi berkas upload yang aman.

---

### 📝 Ringkasan Peningkatan Sistem Terbaru

1. **Standarisasi Bahasa Default Sistem ke English**:
   - Seluruh modul operasional (Dashboard, Devices, Incidents, Reports, Settings, User Profile) distandarkan menggunakan Bahasa Inggris.
   - **Pusat Bantuan (Help Center)** menjadi satu-satunya halaman yang memiliki tombol switcher dwibahasa (ID / EN) interaktif.

2. **Verifikasi Akun & Reset Kata Sandi via Email Asli**:
   - Setiap pembuatan akun baru oleh Admin mewajibkan alamat email aktif (disertai pemblokiran email dummy) dan verifikasi OTP 6-digit.
   - UI Kode Verifikasi menggunakan 6 kotak input terpisah (1 kolom 1 digit) dengan dukungan auto-advance dan paste clipboard.
   - Reset kata sandi di halaman Profil mendukung opsi Kata Sandi Saat Ini atau Verifikasi Email OTP.

3. **Integrasi Global Gmail SMTP**:
   - Pengiriman email notifikasi dan OTP menggunakan gateway Gmail SMTP (`smtp.gmail.com:587`) dengan identitas pengirim resmi `SANOC Jabar Monitoring <sanoc.noreply@gmail.com>` dan header `Reply-To`.

4. **Notifikasi WhatsApp Multi-Target Broadcast**:
   - Tombol pengujian di kartu gateway WhatsApp secara otomatis menyiarkan pesan uji coba ke seluruh target yang terdaftar (baik nomor personal teknisi maupun grup koordinasi WhatsApp).
