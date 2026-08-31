# SANOC (Smart Agent and Network Operation Centers)

## Overview
SANOC adalah sistem pemantauan dan kendali jaringan terpadu (Network Monitoring System) yang dikembangkan untuk regional Jawa Barat. Sistem ini dibangun dengan pendekatan arsitektur *Full-Stack* (Vue.js dan Golang) untuk memantau ketersediaan perangkat jaringan secara *real-time*, mengelola penyebaran notifikasi insiden, serta mengintegrasikan kecerdasan buatan untuk membantu administrator melakukan investigasi masalah jaringan.

## Problem
Infrastruktur jaringan skala besar sering mengalami gangguan yang terlambat terdeteksi. Ketika terjadi insiden (misalnya *node* terputus atau *timeout* berulang), tim *Network Engineer* sering kesulitan melacak runtutan kejadian secara pasti, merangkum *Root Cause Analysis* (RCA) dengan cepat, serta menghadapi kendala dalam mendistribusikan notifikasi insiden secara sistematis kepada tim piket melalui saluran yang tepat (seperti WhatsApp atau Telegram).

## Solution
Membangun sebuah sistem pintar (*Smart Agent*) yang melakukan *polling* status perangkat keras jaringan secara berkala, mengirimkan notifikasi peringatan multi-saluran secara dinamis ketika ambang batas kegagalan tercapai, merekam garis waktu (*timeline*) insiden secara rinci, dan menyediakan asisten kecerdasan buatan (AI Copilot) untuk membantu memandu dan mendiagnosa akar permasalahan secara instan.

## Fitur-fitur Utama
- **Dashboard Real-time**: Memantau seluruh *uptime* dan *downtime* infrastruktur jaringan secara visual dengan indikator yang diperbarui seketika menggunakan protokol WebSocket.
- **Incident Investigation & Timeline**: Garis waktu yang merekam jejak kronologis setiap insiden dengan presisi. Meliputi kegagalan *ping* (ICMP) tahap awal, proses pengiriman notifikasi, kegagalan (*fallback* ke saluran alternatif), hingga detik di mana perangkat dinyatakan kembali *online*.
- **Smart Notification Gateway**: Pengiriman peringatan insiden cerdas yang mampu mencoba menghubungi teknisi NOC melalui WhatsApp terlebih dahulu, dan secara otomatis melakukan *fallback* ke Telegram jika WhatsApp terdeteksi gagal terkirim atau luring.
- **AI Copilot (Gemini Integration)**: Modul asisten kecerdasan buatan yang siap membantu *Network Engineer*. AI ini mempermudah pembuatan *Root Cause Analysis* (RCA), meringkas status kinerja jaringan harian, dan bahkan memberikan panduan draf pelaporan insiden kepada pimpinan.
- **Perlindungan Login Aman**: Antarmuka masuk (Login) dilindungi dengan otentikasi JWT (JSON Web Token) dan dilengkapi oleh validasi bot/anti-spam menggunakan **Google reCAPTCHA**.

## Implementation
- **Dibangun dengan arsitektur terpisah (*decoupled*)**: **Vue.js 3** untuk Frontend dan **Golang** (Go) untuk Backend
- **Database Terintegrasi**: Menggunakan **PostgreSQL** untuk menyimpan data perangkat dan histori insiden, serta **Redis** untuk *caching* dan optimasi antrian
- **Sistem *Polling***: Mesin internal melakukan pemantauan berbasis ICMP secara *real-time* ke perangkat-perangkat terdaftar
- **Notifikasi Pintar**: Distribusi pesan dikelola secara hierarkis menggunakan *Sidecar* berbasis **Node.js** khusus untuk WhatsApp Gateway dan native API untuk Telegram Bot
- **Integrasi AI**: **Google Gemini API** dihubungkan ke sistem sebagai AI Copilot interaktif untuk *Root Cause Analysis* (RCA) dan rangkuman harian
- **Keamanan**: Sistem dilindungi dengan otentikasi JWT dan anti-spam **reCAPTCHA**

## Result
Sistem ini memungkinkan tim NOC (Network Operations Center) memantau infrastruktur berskala besar secara proaktif dari satu dasbor *real-time*. Ketika suatu jaringan atau perangkat terputus, insiden langsung terekam lengkap dengan *timeline*-nya, notifikasi terdistribusi ke ponsel teknisi dalam hitungan detik (via WA/Telegram), dan asisten AI secara signifikan mempercepat proses analisis sehingga *downtime* jaringan dapat diminimalisir dengan efektif.

## Cara Menggunakan Aplikasi Web
1. **Login ke Sistem**
   Akses antarmuka web melalui peramban web (browser). Masukkan kredensial administrator Anda (Email dan Kata Sandi), dan klik/centang kotak verifikasi reCAPTCHA jika diminta, kemudian tekan **Masuk**.
2. **Pantau Dasbor & Inventaris Perangkat**
   Setelah masuk, Anda akan melihat *Dashboard* yang merangkum *uptime* sistem serta peta penyebaran insiden. Beralih ke menu **Devices** (Perangkat) di bilah navigasi kiri untuk menambahkan perangkat jaringan baru (IP, Nama, Grup) yang ingin Anda pantau. Sistem *backend* akan otomatis mulai melakukan *ping* (ICMP) ke *IP Address* tersebut setiap saat.
3. **Memonitor dan Menginvestigasi Insiden**
   Jika perangkat gagal merespons, sistem akan menghasilkan lansiran baru di halaman **Incidents**.
   - Buka menu **Incidents**.
   - Klik salah satu ID insiden untuk membuka mode investigasi.
   - Di sini Anda dapat melihat bagan **Timeline**, mulai dari perangkat gagal *ping* tahap 1, 2, dan 3, proses agregasi peringatan, hingga sistem melaporkan pengiriman peringatan melalui WhatsApp (atau Telegram).
4. **Berinteraksi dengan AI Copilot**
   Di layar investigasi mana pun (atau melalui ikon *AI Copilot* di sudut kanan bawah), Anda dapat meluncurkan asisten pintar. 
   - Klik saran cepat (contoh: *"Lakukan Analisis Akar Masalah untuk insiden ini"*).
   - Atau ketikkan pertanyaan manual seperti: *"Buatkan saya draf WhatsApp untuk dilaporkan ke pimpinan mengenai matinya server Lantai 2"*. AI akan langsung merumuskan pesannya untuk Anda.

---
> **Informasi Modul Spesifik**:
> - Lihat **[Frontend README](./frontend/README.md)** untuk panduan teknis UI, konfigurasi `.env`, dan instalasi Node.js (Vite).
> - Lihat **[Backend README](./backend/README.md)** untuk panduan instalasi Golang, konfigurasi Database, dan *startup* WA Sidecar.
