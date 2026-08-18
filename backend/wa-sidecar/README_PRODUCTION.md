# 🚀 Panduan Konfigurasi WhatsApp Sidecar di Production

Dokumen ini berisi panduan lengkap variabel lingkungan (*environment variables*), keamanan, persistensi sesi, dan langkah-langkah deployment untuk **WhatsApp Baileys Sidecar** pada lingkungan Production.

---

## 📋 1. Variabel Lingkungan (`.env`) yang Wajib Disesuaikan

Buat atau perbarui file `.env` di dalam folder `backend/wa-sidecar/.env` (atau set variabel lingkungan server/Docker Anda) dengan nilai produksi:

```ini
# ─── SANOC WhatsApp Sidecar — Production Config ─────────────────────────

# 1. Port internal HTTP untuk WA Sidecar
PORT=3001

# 2. Path Penyimpanan Sesi Login WhatsApp (Wajib Persisten)
WHATSAPP_SESSION_PATH=./auth_sessions/whatsapp

# 3. Secret Token Komunikasi Internal dengan Go Backend
# (WAJIB SAMA dengan WA_SIDECAR_TOKEN pada backend/.env.production)
INTERNAL_TOKEN=sanoc-prod-wa-secret-key-2026-secure!
```

---

## 🔄 2. Sinkronisasi dengan Go Backend (`backend/.env.production`)

Pastikan file `backend/.env.production` milik backend Go memiliki konfigurasi yang **identik** untuk token internalnya:

```ini
# ─── WhatsApp Baileys Sidecar Integration ──────────────────────────────
WA_SIDECAR_URL=http://127.0.0.1:3001
WA_SIDECAR_TOKEN=sanoc-prod-wa-secret-key-2026-secure!
```

> ⚠️ **PENTING**: Nilai `INTERNAL_TOKEN` di WA Sidecar dan `WA_SIDECAR_TOKEN` di Go Backend **HARUS SAMA**. Jika tidak sama, backend akan mendapatkan error `401 Unauthorized`.

---

## 🛡️ 3. Keamanan Port & Firewall

1. **Port Internal Only**: Port `3001` adalah API internal yang **HANYA** digunakan oleh Go Backend.
2. **Jangan Buka ke Internet Publik**:
   - Apabila menggunakan `ufw` di Ubuntu/Debian:
     ```bash
     sudo ufw deny 3001
     ```
   - Apabila menggunakan Nginx reverse proxy, **JANGAN** membuat `location` proxy ke port 3001 untuk domain publik. Port 3001 cukup dapat diakses dari `127.0.0.1` local server.

---

## 💾 4. Persistensi Sesi WhatsApp (`auth_sessions`)

Folder `auth_sessions/whatsapp` menyimpan kredensial autentikasi WhatsApp (`creds.json` dan *app-state keys*).

- **Jangan Hapus Sesi**: Jangan menghapus atau menimpa folder `auth_sessions` saat melakukan CI/CD deployment atau *update code*.
- **Hak Akses Folder**: Pastikan user yang menjalankan Node.js memiliki hak akses baca & tulis:
  ```bash
  mkdir -p backend/wa-sidecar/auth_sessions/whatsapp
  chmod -R 755 backend/wa-sidecar/auth_sessions
  ```
- **Jika Menggunakan Docker**: Wajib gunakan **Docker Volume** untuk folder sesi agar login WhatsApp tidak hilang ketika container di-restart:
  ```yaml
  volumes:
    - ./wa-sessions:/app/auth_sessions
  ```

---

## 🛠️ 5. Cara Menjalankan di Production

### Opsi A: Menggunakan PM2 (Rekomendasi untuk Linux/VPS Server)

1. Masuk ke direktori sidecar:
   ```bash
   cd backend/wa-sidecar
   ```
2. Install dependensi (hanya produksi):
   ```bash
   npm ci --only=production
   ```
3. Jalankan menggunakan PM2:
   ```bash
   pm2 start index.js --name "sanoc-wa-sidecar"
   pm2 save
   pm2 startup
   ```

### Opsi B: Menggunakan Docker Compose

```yaml
version: '3.8'
services:
  wa-sidecar:
    build: ./backend/wa-sidecar
    container_name: sanoc-wa-sidecar
    restart: always
    environment:
      - PORT=3001
      - INTERNAL_TOKEN=sanoc-prod-wa-secret-key-2026-secure!
      - WHATSAPP_SESSION_PATH=/app/auth_sessions/whatsapp
    volumes:
      - wa_auth_data:/app/auth_sessions
    ports:
      - "127.0.0.1:3001:3001"

volumes:
  wa_auth_data:
```

---

## ✅ 6. Checklist Verifikasi Setelah Deploy

1. **Cek Status Health Check**:
   ```bash
   curl http://127.0.0.1:3001/health
   ```
   *Respon sukses:* `{"status":"disconnected","healthy":false,"hasSavedAuth":false}` (atau `"connected"` jika sudah terhubung).

2. **Cek Koneksi WhatsApp via Web UI**:
   - Buka menu **System Configuration > Notification Channels** di Dashboard SANOC.
   - Klik **QR Reconnect** dan scan QR Code menggunakan WhatsApp HP Anda.
   - Setelah status menjadi `CONNECTED`, klik **Send Test Notification** untuk menguji pengiriman pesan WhatsApp.
