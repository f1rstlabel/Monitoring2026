# 15 — SMTP Email Gateway & Production Configuration Guide
*Panduan Konfigurasi Gateway Email SMTP & OTP untuk Lingkungan Produksi*

---

## 📌 Ringkasan / Overview
Layanan SMTP pada **SANOC (Sanditel Network Operations Center)** bertanggung jawab atas pengiriman email transaksional penting dengan tingkat keterandalan tinggi:
1. **Verifikasi OTP Pendaftaran Akun Baru (User Registration)**: Mengirimkan 6-digit kode OTP ke email real pengguna sebelum akun dibuat.
2. **Reset Kata Sandi Akun (Password Reset)**: Mengirimkan kode verifikasi OTP ke email pengguna yang terdaftar untuk otentikasi mandiri.
3. **Audit Notifikasi & Keamanan**: Memastikan setiap aktivitas autentikasi terverifikasi dan tercatat pada log sistem.

---

## 🏗️ Pilihan Konfigurasi SMTP untuk Production

### Opsi 1: Google Workspace / Gmail SMTP (Rekomendasi Cepat & Handal)
Jika menggunakan akun Gmail atau Google Workspace (`@jabarprov.go.id` via Google):

1. **Aktifkan 2-Step Verification (Verifikasi 2 Langkah)** pada Akun Google pengirim.
2. **Buat App Password (Sandi Aplikasi)**:
   - Buka: [Google Account Security → 2-Step Verification → App Passwords](https://myaccount.google.com/apppasswords).
   - Masukkan nama aplikasi: `SANOC Monitoring`.
   - Google akan memberikan **16 karakter kode sandi** (contoh: `abcd efgh ijkl mnop`).
3. **Konfigurasi pada `.env` Server Production**:
   ```env
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USER=akun.resmi.sanoc@gmail.com
   SMTP_PASSWORD=abcdefghijklmnop
   SMTP_FROM=SANOC Jabar Monitoring <sanoc.noreply@gmail.com>
   ```

> [!IMPORTANT]
> **Kebijakan SPF / DMARC Google:**  
> Untuk mencegah email masuk ke folder *Spam*, pastikan `SMTP_USER` adalah email Google yang memiliki kredensial App Password tersebut. Sistem SANOC secara otomatis mengatur *envelope sender* yang valid agar selalu lolos filter reputasi email.

---

### Opsi 2: Mail Server Internal Pemerintah / Enterprise (Zimbra / Postfix / Exchange)
Jika instansi memiliki mail server lokal (misal `mail.jabarprov.go.id`):

1. **Buka Port Outbound pada Firewall Server SANOC**:
   - Pastikan port **587 (STARTTLS)** atau **465 (Direct SSL)** diizinkan keluar (*egress*) menuju Mail Server.
2. **Daftarkan Whitelist / Relay IP**:
   - Minta administrator mail server untuk mendaftarkan IP Public server SANOC pada *relay whitelist* atau buatkan akun layanan (*service account*).
3. **Konfigurasi pada `.env`**:
   ```env
   SMTP_HOST=mail.jabarprov.go.id
   SMTP_PORT=587
   SMTP_USER=sanoc-noreply@jabarprov.go.id
   SMTP_PASSWORD=PasswordKuatMailServer2026!
   SMTP_FROM=SANOC Jabar Monitoring <sanoc-noreply@jabarprov.go.id>
   ```

---

### Opsi 3: Layanan Email Transaksional Cloud (SendGrid, Mailgun, Amazon SES, Brevo)
Untuk pengiriman skala besar dengan analitik keterkiriman:

1. Buat API Key / SMTP Credentials di dashboard provider (SendGrid / Mailgun / Brevo).
2. Lakukan verifikasi domain (menambahkan record DNS CNAME, TXT untuk SPF dan DKIM).
3. **Konfigurasi pada `.env`**:
   ```env
   # Contoh SendGrid
   SMTP_HOST=smtp.sendgrid.net
   SMTP_PORT=587
   SMTP_USER=apikey
   SMTP_PASSWORD=SG.xxxxxxxxxxxxxxxxxxxx
   SMTP_FROM=SANOC Jabar Monitoring <noreply@monitoring.jabarprov.go.id>

   # Contoh Amazon SES
   SMTP_HOST=email-smtp.ap-southeast-1.amazonaws.com
   SMTP_PORT=587
   SMTP_USER=AKIAIOSFODNN7EXAMPLE
   SMTP_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
   SMTP_FROM=SANOC Jabar Monitoring <noreply@monitoring.jabarprov.go.id>
   ```

---

### Opsi 4: Lingkungan Development / Uji Coba Lokal (Mailpit Container)
Untuk keperluan pengujian lokal tanpa internet atau tanpa mengirim email sungguhan:

```yaml
# Pada docker-compose.yml
services:
  mailpit:
    image: axllent/mailpit:latest
    container_name: sanoc-mailpit
    restart: always
    ports:
      - "8025:8025" # Web UI untuk membaca email yang masuk
      - "1025:1025" # Port SMTP lokal
```

Konfigurasi `.env` dev:
```env
SMTP_HOST=127.0.0.1
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=SANOC Dev <noreply@sanoc.local>
```

---

## ⚙️ Fitur Keterandalan Mailer SANOC (Dual-Port Fallback)
Backend Go pada SANOC dilengkapi dengan sistem pengiriman cerdas:
1. **STARTTLS Otomatis (Port 587)**: Pengiriman standar terenkripsi TLS.
2. **Auto-Fallback ke SSL (Port 465)**: Jika Port 587 diblokir oleh ISP / penyedia VPS / Cloud Provider (beberapa cloud memblokir port 587/25), mailer secara otomatis mencoba kembali via direct SSL Port 465 tanpa menyebabkan error pada pengguna.
3. **Fallback Logging Console**: Setiap kali OTP digenerate, kode 6-digit juga otomatis tercatat pada log backend untuk memudahkan penanganan darurat oleh NOC Administrator:
   ```text
   [EMAIL_VERIFICATION] 6-Digit Registration OTP for user@jabarprov.go.id: 541298 (Expires: 2026-08-18T15:00:00Z)
   ```

---

## 🛠️ Panduan Pemecahan Masalah (Troubleshooting)

| Gejala Masalah | Penyebab Umum | Solusi Perbaikan |
|---|---|---|
| `530 5.7.0 Must issue a STARTTLS command first` | `SMTP_USER` atau `SMTP_PASSWORD` pada `.env` kosong / salah nama variabel. | Periksa file `.env` di direktori server backend, pastikan `SMTP_USER` dan `SMTP_PASSWORD` terisi lengkap. |
| `535 5.7.8 Username and Password not accepted` | Password Gmail biasa digunakan alih-alih **App Password** 16 karakter. | Buat **App Password** baru di Google Account Security dan gunakan kode tersebut tanpa spasi. |
| `Connection timed out` pada port 587 | Provider VPS / Cloud membatasi port email keluar. | Ubah `SMTP_PORT=465` di `.env` agar menggunakan jalur direct SSL. |
| Email masuk ke folder **Spam / Promosi** | Domain pengirim tidak memiliki SPF / DKIM valid. | Gunakan akun pengirim yang sesuai dengan kredensial login SMTP, atau selaraskan DNS SPF domain. |
