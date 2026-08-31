# SANOC - Frontend

Bagian frontend dari proyek SANOC dibangun menggunakan framework modern untuk memberikan antarmuka pengguna (UI) yang reaktif, elegan, dan *real-time*. Aplikasi SPA (Single Page Application) ini dirancang khusus untuk memenuhi standar dasbor *monitoring* jaringan skala *enterprise*.

## Tech Stack

- **Framework**: [Vue.js 3](https://vuejs.org/) (Composition API)
- **Build Tool**: [Vite](https://vitejs.dev/)
- **Bahasa**: TypeScript
- **State Management**: [Pinia](https://pinia.vuejs.org/)
- **Routing**: [Vue Router](https://router.vuejs.org/)
- **Styling**: [Tailwind CSS](https://tailwindcss.com/)
- **Grafik / Visualisasi**: Vue3-ApexCharts
- **Real-time Data**: Native WebSocket

## Instalasi dan Setup

1. **Persiapan Dependensi**
   Pastikan Node.js versi 18 atau yang lebih baru sudah terinstal di sistem Anda. Masuk ke direktori `frontend` dan instal dependensi menggunakan `npm`:
   ```bash
   cd frontend
   npm install
   ```

2. **Konfigurasi Environment (*.env*)**
   Salin file konfigurasi *.env* dan sesuaikan dengan *environment* lokal Anda.
   Variabel lingkungan utama yang dibutuhkan:
   ```env
   VITE_API_BASE_URL=http://localhost:8080/api/v1
   VITE_WS_URL=ws://localhost:8080/ws/live
   VITE_RECAPTCHA_SITE_KEY=your_recaptcha_site_key
   ```
   > **Catatan**: Jika menggunakan reCAPTCHA, pastikan `VITE_RECAPTCHA_SITE_KEY` terisi dan sesuai dengan yang ada di server backend, jika tidak halaman login akan mengalami error atau tidak menampilkan tombol submit dengan benar.

3. **Menjalankan Development Server**
   Gunakan perintah berikut untuk menjalankan server pengembangan lokal (dilengkapi *Hot Module Replacement* / HMR):
   ```bash
   npm run dev
   ```
   Aplikasi secara *default* akan berjalan pada `http://localhost:5173`.

4. **Kompilasi (Build) untuk Produksi**
   Jika ingin melakukan kompilasi rilis produksi, jalankan:
   ```bash
   npm run build
   ```
   Hasil *build* statis akan diletakkan di direktori `dist/` dan siap disajikan (*serve*) menggunakan Nginx atau web server lainnya.

## Struktur Direktori Utama

- `src/components/`: Komponen Vue UI yang dapat digunakan kembali (*reusable*), terbagi dalam kategori (ai, common, dashboard).
- `src/views/`: Halaman-halaman utama tingkat rute (*router views*).
- `src/stores/`: Konfigurasi *state management* (Pinia) untuk menyimpan status global seperti notifikasi, perangkat, dan preferensi AI.
- `src/api/`: Definisi layanan klien API Axios (HTTP Requests).
- `src/router/`: Pemetaan rute aplikasi.
