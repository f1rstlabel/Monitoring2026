# SANOC — Smart Agent & Network Operations Center

> **System Monitoring & SLA Compliance Audit Platform — Infrastructure NOC**  
> *Sistem Monitoring Infrastruktur Jaringan & Audit Kepatuhan SLA Real-time*  
> **Diskominfo Provinsi Jawa Barat — UTB 2026**

---

## 🌐 Language Options / Pilihan Bahasa
- 🇬🇧 [English Documentation](#-english-documentation)
- 🇮🇩 [Dokumentasi Bahasa Indonesia](#-dokumentasi-bahasa-indonesia)

---

## 🇬🇧 English Documentation

### 📌 Project Overview
**SANOC (Smart Agent & Network Operations Center)** is an enterprise-grade network infrastructure monitoring and SLA compliance auditing system. It provides real-time automated ICMP ping and SNMP polling across regional network nodes, switches, routers, access points, and server devices.

Key capabilities include:
- **Real-Time Monitoring**: Automated ICMP & SNMP polling engine with sub-minute failure detection.
- **Incident & Outage Management**: Automated incident ticket generation, live outage duration tracking, and incident resolution workflows.
- **Multi-Channel Alerting**: Instant dispatching of outage notifications via WhatsApp API Gateway and Telegram Bot Channel.
- **SLA Compliance Reporting**: Automated SLA uptime calculations, Mean Time to Recovery (MTTR) tracking, and customizable export options (PDF, Excel `.xls`, CSV `.csv`).
- **Enterprise Security**: Two-Factor Authentication (TOTP MFA via Google Authenticator / Authy), Role-Based Access Control (RBAC), and secure avatar file uploads.

---

### 🚀 Key Features

1. **Interactive Dashboard**: Real-time SLA gauges, total active outages, device up/down counters, and live event feed.
2. **Device Inventory Management**: Single device creation, bulk CSV/Excel import, customizable failure thresholds, and SNMP SysName/Location discovery.
3. **Active Incidents Queue**: Ticket management, status filtering (ACTIVE / RESOLVED), grouping by location/device, and printable outage incident reports.
4. **SLA Audit & Reports**: Tab-specific report generation for:
   - **Downtime by Device**: Total downtime duration and outage counts per node.
   - **Recurring Issues**: Identification of flapping devices (≥5 outages in 7 days).
   - **Active Incidents**: Current outage ticket queue summary.
5. **User Profile & Security**: Profile photo uploads with security validation (magic-byte MIME checking, CSP headers), password updates, and TOTP 2FA setup.

---

### 📦 Technology Stack

- **Backend**: Go (Golang 1.22+), Gin Web Framework, PostgreSQL 14+, Redis (for rate limiting and caching), JWT Authentication, TOTP MFA (`pquerna/otp`).
- **Frontend**: Vue 3 (Composition API), Vite, TypeScript, Pinia State Store, TailwindCSS, Lucide Icons.
- **Database Migrations**: `golang-migrate`.

---

### 📝 Recent Improvements & Today's Changelog (Hari Ini)

The following major enhancements and bug fixes were completed today before pushing to GitHub:

1. **TOTP MFA Verification Stuck Loading Fix**:
   - Expanded TOTP time-skew tolerance window to $\pm 90$ seconds (7 consecutive 30-second windows) to absorb system clock drift between mobile authenticators and server time.
   - Made Pinia `fetchMe()` call non-blocking during MFA verification to allow immediate UI navigation to the dashboard without hanging on "Verifying...".
   - Added automatic whitespace and hyphen stripping from input passcodes.

2. **User Profile Username Population**:
   - Removed `omitempty` from `Username` JSON tag in backend `domain.User` struct so the field is always returned in API responses.
   - Added automatic fallback to email prefix (`name` from `name@domain.com`) in PostgreSQL repository when username is blank.
   - Added deep reactivity watchers in Vue `UserProfileView.vue` so the username field is populated automatically.

3. **Multi-Page PDF Export & Pagination Layout Overhaul**:
   - Removed restrictive `max-w-4xl` and `p-8` padding constraints from printable components (`PrintableSLAAudit.vue` and `PrintableIncidentsList.vue`), enabling 100% full-width A4 printable layouts.
   - Enforced `break-before: avoid !important;` on table headings and elements to eliminate empty white spaces on Page 1 and fill Page 1 top-to-bottom.
   - Added `thead { display: table-header-group !important; }` so table column headers automatically repeat at the top of every printed page.
   - Removed numbered prefixes (`2.`) from section headings in PDF layouts for a cleaner professional appearance.

4. **Tab-Specific & Distinct File Export Naming**:
   - Updated report export handlers so Excel (`.xls`), CSV (`.csv`), and PDF (`.pdf`) exports strictly export the dataset of the currently selected tab (**Downtime by Device**, **Recurring Issues**, **Active Incidents**).
   - Added dynamic filename prefixes matching format and active tab (e.g., `sanoc-excel-downtime-by-device-monthly-2026-08-12.xls`, `sanoc-pdf-active-incidents-monthly-2026-08-12.pdf`).

5. **Profile Avatar Upload & Security Hardening**:
   - Configured Vite `/uploads` proxy to local backend static server.
   - Implemented MIME magic-byte validation, extension whitelist, safe filename sanitization (`filepath.Base`), and orphan file cleanup.
   - Applied `X-Content-Type-Options: nosniff` and `Content-Security-Policy` headers to static routes.

---

### 💻 Quick Start & Local Setup

#### Prerequisites
- Go 1.22+
- Node.js 18+ & npm
- PostgreSQL 14+
- Redis 6+

#### 1. Backend Setup
```bash
cd backend

# Copy environment template (Do NOT commit real passwords)
cp .env.example .env

# Run database migrations
go run ./cmd/api --migrate

# Start Backend API Server
go run ./cmd/api
```

#### 2. Frontend Setup
```bash
cd frontend

# Install dependencies
npm install

# Start Vite Development Server
npm run dev
```
Open browser at `http://localhost:5173`.

---

<br/>

---

## 🇮🇩 Dokumentasi Bahasa Indonesia

### 📌 Penjelasan Project
**SANOC (Smart Agent & Network Operations Center)** adalah sistem pemantauan infrastruktur jaringan berbasis *enterprise* dan pengauditan kepatuhan SLA (*Service Level Agreement*). Sistem ini secara otomatis memonitor perangkat jaringan (router, switch, access point, server, CCTV) menggunakan protokol ICMP Ping dan SNMP Polling secara real-time.

Fitur Unggulan:
- **Monitoring Real-time**: Engine polling ICMP & SNMP otomatis dengan deteksi gangguan dalam hitungan detik.
- **Manajemen Incident & Outage**: Pembuatan tiket kejadian otomatis, pencatatan durasi outage *live*, dan workflow penyelesaian incident.
- **Pengiriman Notifikasi Multi-Kanal**: Pengiriman pesan peringatan seketika via WhatsApp API Gateway dan Telegram Bot.
- **Laporan Kepatuhan SLA**: Perhitungan persentase Uptime SLA otomatis, pencatatan MTTR (*Mean Time to Recovery*), serta pilihan ekspor dokumen (PDF, Excel `.xls`, CSV `.csv`).
- **Keamanan Tingkat Tinggi**: Autentikasi Dua Faktor (TOTP MFA via Google Authenticator / Authy), Hak Akses Berbasis Role (RBAC), serta validasi berkas upload yang aman.

---

### 📝 Ringkasan Perubahan Hari Ini (Today's Changelog)

Berikut adalah daftar perbaikan dan peningkatan utama yang telah diselesaikan hari ini sebelum dilakukan *push* ke GitHub:

1. **Perbaikan MFA Verification Stuck Loading**:
   - Memperluas toleransi waktu TOTP menjadi $\pm 90$ detik (7 jendela waktu 30-detik) untuk menangani perbedaan jam antara server dan aplikasi Google Authenticator HP.
   - Mengubah proses `fetchMe()` pada Pinia `authStore` menjadi non-blocking saat verifikasi MFA agar halaman langsung beralih ke Dashboard tanpa tertahan (*stuck*).
   - Otomatis membersihkan spasi dan karakter strip pada input kode 6-digit.

2. **Perbaikan Tampilan Username pada Profil User**:
   - Menghapus tag `omitempty` dari field `Username` pada struct domain Go di backend agar nilai username selalu disertakan dalam respon JSON API.
   - Menambahkan *fallback* otomatis dari awalan email (`budiono` dari `budiono@gmail.com`) pada layer repository PostgreSQL ketika username di database kosong.
   - Menambahkan reaktivitas *watcher* di `UserProfileView.vue` agar form profil langsung terisi username pengguna yang aktif.

3. **Perapihan Layout Cetak PDF Multi-Halaman & Responsif**:
   - Menghapus pembatasan `max-w-4xl` dan padding `p-8` pada komponen cetak (`PrintableSLAAudit.vue` dan `PrintableIncidentsList.vue`) sehingga dokumen PDF menggunakan 100% lebar kertas A4 secara rapi.
   - Menambahkan aturan `break-before: avoid !important;` pada judul section dan tabel untuk mengeliminasi area kosong (*empty gap*) pada Halaman 1 dan mengisi Halaman 1 dari atas ke bawah.
   - Menambahkan `thead { display: table-header-group !important; }` agar header kolom tabel otomatis berulang di bagian atas setiap halaman baru (untuk data 10, 25, 50, 100+ baris).
   - Menghapus nomor prefiks (`2.`) pada subjudul PDF agar tampilan laporan terlihat lebih bersih dan profesional.

4. **Ekspor Data Spesifik Tab & Pembedaan Nama File**:
   - Memastikan ekspor Excel (`.xls`), CSV (`.csv`), dan PDF (`.pdf`) pada halaman Reports strictly mengekspor data yang sesuai dengan tab yang sedang dipilih (**Downtime by Device**, **Recurring Issues**, **Active Incidents**).
   - Memberikan penamaan file otomatis yang spesifik sesuai format dan nama tab (contoh: `sanoc-excel-downtime-by-device-monthly-2026-08-12.xls`, `sanoc-pdf-active-incidents-monthly-2026-08-12.pdf`).

5. **Keamanan Upload Foto Profil**:
   - Mengonfigurasi proxy `/uploads` pada Vite dev server ke backend Go static file handler.
   - Menerapkan validasi header MIME magic-byte, daftar putih ekstensi (JPG, PNG, WEBP), sanitasi nama file (`filepath.Base`), dan pembersihan file yatim (*orphan files*).
   - Menerapkan header keamanan `X-Content-Type-Options: nosniff` dan `Content-Security-Policy`.

---

### 🛡️ Keamanan & Data Sensitif

Dokumen ini dan seluruh repositori kode **bebas dari data sensitif**:
- Seluruh kata sandi, kunci rahasia JWT, token bot Telegram, dan kredensial database disimpan secara terpisah dalam berkas `.env` lokal.
- Berkas `.env`, sertifikat SSL/TLS (`*.pem`, `*.key`), file biner (`*.exe`, `bin/`), serta folder upload pengguna telah dimasukkan ke dalam `.gitignore` sehingga tidak akan ikut terdorong (*pushed*) ke repositori GitHub.

---

### 📄 Lisensi & Hak Cipta
© **SANOC Team — Diskominfo Provinsi Jawa Barat (UTB 2026)**. All rights reserved.
