<template>
  <Teleport to="body">
    <div v-if="incident" id="printable-incident-report" class="hidden print:block bg-white text-slate-900 p-10 font-sans leading-normal text-xs max-w-4xl mx-auto border border-slate-300 shadow-none">
      <!-- Official Header with Kop Surat Style -->
      <div class="border-b-2 border-slate-900 pb-4 mb-6 flex items-center justify-between">
        <div>
          <h2 class="text-sm font-extrabold tracking-wider uppercase text-slate-900">PEMERINTAH PROVINSI JAWA BARAT</h2>
          <h1 class="text-lg font-black tracking-tight text-indigo-900">DINAS KOMUNIKASI DAN INFORMATIKA — NOC IT</h1>
          <p class="text-[10px] font-mono text-slate-600">Laporan Resmi Insiden Infrastruktur &amp; Disaster Recovery Audit</p>
        </div>
        <div class="text-right font-mono text-[10px] text-slate-600 bg-slate-100 p-2.5 rounded border border-slate-200">
          <div><strong class="text-slate-900">TICKET ID:</strong> {{ incident.id }}</div>
          <div><strong class="text-slate-900">TANGGAL LAPORAN:</strong> {{ new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) }}</div>
          <div><strong class="text-slate-900">WAKTU:</strong> {{ new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }) }} WIB</div>
        </div>
      </div>

      <!-- Ringkasan Insiden (Grid Summary Box) -->
      <div class="mb-6 border border-slate-300 rounded overflow-hidden">
        <div class="bg-slate-800 text-white font-mono text-[11px] font-bold px-4 py-2 uppercase tracking-wide">
          1. RINGKASAN DATA NODES &amp; EFEK INSIDEN
        </div>
        <div class="p-4 grid grid-cols-2 gap-4 bg-slate-50 text-xs">
          <div class="space-y-1.5 border-r border-slate-200 pr-4">
            <div><span class="text-slate-500 font-mono">Nama Perangkat:</span> <strong class="text-slate-900 font-bold text-sm">{{ incident.deviceName }}</strong></div>
            <div><span class="text-slate-500 font-mono">Tipe &amp; Kategori:</span> <span class="font-mono text-slate-800">{{ incident.deviceType }}</span></div>
            <div><span class="text-slate-500 font-mono">Alamat IP:</span> <span class="font-mono font-bold text-indigo-900">{{ incident.deviceIp }}</span></div>
            <div><span class="text-slate-500 font-mono">Lokasi Fisik:</span> <span class="font-semibold text-slate-800">{{ incident.location || 'Gedung Sate / NOC Server Room' }}</span></div>
          </div>
          <div class="space-y-1.5">
            <div><span class="text-slate-500 font-mono">Status Insiden:</span> <span class="font-mono font-bold uppercase px-2 py-0.5 rounded text-[10px]" :class="incident.status === 'RESOLVED' ? 'bg-emerald-100 text-emerald-800 border border-emerald-300' : 'bg-red-100 text-red-800 border border-red-300'">{{ incident.status }}</span></div>
            <div><span class="text-slate-500 font-mono">Durasi Outage:</span> <strong class="font-mono text-red-700">{{ incident.duration }}</strong></div>
            <div><span class="text-slate-500 font-mono">Waktu Mulai:</span> <span class="font-mono text-slate-800">{{ incident.startTime }}</span></div>
            <div v-if="incident.resolvedAt"><span class="text-slate-500 font-mono">Waktu Pulih:</span> <span class="font-mono text-emerald-800">{{ incident.resolvedAt }}</span></div>
          </div>
        </div>
      </div>

      <!-- Tabel Detail Dampak & Audit -->
      <div class="mb-6 break-inside-avoid page-break-inside-avoid">
        <h3 class="font-bold text-xs uppercase tracking-wider text-slate-800 font-mono mb-2">2. ANALISIS DAMPAK &amp; AUDIT TRAIL NOTIFIKASI</h3>
        <table class="w-full text-left border-collapse border border-slate-300 text-xs" style="table-layout: fixed; width: 100%;">
          <colgroup>
            <col style="width: 35%;">
            <col style="width: 65%;">
          </colgroup>
          <thead>
            <tr class="bg-slate-100 font-bold text-slate-700 font-mono uppercase text-[10px]">
              <th class="border border-slate-300 p-2.5">Parameter Audit</th>
              <th class="border border-slate-300 p-2.5">Hasil Evaluasi / Deskripsi Sistem</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200">
            <tr class="break-inside-avoid page-break-inside-avoid">
              <td class="border border-slate-300 p-2.5 font-semibold text-slate-800">Jumlah Node Terdampak</td>
              <td class="border border-slate-300 p-2.5 font-mono">{{ incident.affectedDevicesCount || 1 }} Perangkat Infrastructure</td>
            </tr>
            <tr class="break-inside-avoid page-break-inside-avoid">
              <td class="border border-slate-300 p-2.5 font-semibold text-slate-800">Indikator Kegagalan ICMP</td>
              <td class="border border-slate-300 p-2.5">Ping Packet Loss 100% (Batas threshold kegagalan berturut-turut terlampaui).</td>
            </tr>
            <tr class="break-inside-avoid page-break-inside-avoid">
              <td class="border border-slate-300 p-2.5 font-semibold text-slate-800">Kanal Dispatched Alert</td>
              <td class="border border-slate-300 p-2.5">Notifikasi otomatis terkirim via WhatsApp API Gateway &amp; Telegram NOC Channel.</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Tanda Tangan & Pengesahan Dokumen -->
      <div class="mt-12 pt-6 border-t-2 border-slate-300 grid grid-cols-2 gap-12 text-center text-xs break-inside-avoid page-break-inside-avoid">
        <div>
          <p class="text-slate-600 font-mono text-[10px] mb-12">Petugas Operator NOC (Diskominfo Jabar)</p>
          <div class="border-b border-slate-900 w-44 mx-auto mb-1"></div>
          <p class="font-bold text-slate-900">Tim Operasional NOC</p>
          <p class="text-[10px] text-slate-500 font-mono">NIP: 19880412 201403 1 002</p>
        </div>
        <div>
          <p class="text-slate-600 font-mono text-[10px] mb-12">Penanggung Jawab Infrastruktur / Superadmin</p>
          <div class="border-b border-slate-900 w-44 mx-auto mb-1"></div>
          <p class="font-bold text-slate-900">Kepala Sub-Bagian IT &amp; Jaringan</p>
          <p class="text-[10px] text-slate-500 font-mono">NIP: 19820719 200801 1 005</p>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import type { Incident } from '../../types';

defineProps<{
  incident: Incident | null;
}>();
</script>

<style>
@media print {
  @page {
    size: A4;
    margin: 15mm 20mm 15mm 20mm;
  }

  body {
    background: white !important;
    color: #0f172a !important;
  }

  #printable-incident-report {
    display: block !important;
    width: 100% !important;
    max-width: 100% !important;
    padding: 0 !important;
    margin: 0 !important;
    border: none !important;
    box-shadow: none !important;
  }

  /* Prevent page split inside table rows/cards */
  .break-inside-avoid {
    break-inside: avoid !important;
    page-break-inside: avoid !important;
  }

  /* Wrap text for table cells */
  table {
    table-layout: fixed !important;
    width: 100% !important;
  }

  th, td {
    word-break: break-word !important;
    overflow-wrap: break-word !important;
    white-space: normal !important;
  }
}
</style>

