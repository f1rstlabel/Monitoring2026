<template>
  <Teleport to="body">
    <div id="printable-sla-audit" class="hidden print:block bg-white text-slate-900 p-10 font-sans leading-normal text-xs max-w-4xl mx-auto border border-slate-300 shadow-none">
      <!-- Kop Surat Header -->
      <div class="border-b-2 border-slate-900 pb-4 mb-6 flex justify-between items-start">
        <div>
          <h2 class="text-sm font-extrabold tracking-wider uppercase text-slate-900">PEMERINTAH PROVINSI JAWA BARAT</h2>
          <h1 class="text-lg font-black tracking-tight text-indigo-900">DINAS KOMUNIKASI DAN INFORMATIKA — GOVMONITOR IT</h1>
          <p class="text-[10px] font-mono text-slate-600 mt-0.5">Laporan Kepatuhan SLA Uptime &amp; Audit Kesehatan Infrastruktur NOC</p>
        </div>
        <div class="text-right font-mono text-[10px] text-slate-600 bg-slate-100 p-2.5 rounded border border-slate-200">
          <div><strong class="text-slate-900">PERIODE AUDIT:</strong> {{ periodLabel }}</div>
          <div><strong class="text-slate-900">TANGGAL CETAK:</strong> {{ new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) }}</div>
        </div>
      </div>

      <!-- Ringkasan SLA Metrics (Grid) -->
      <!-- Ringkasan SLA Metrics (Grid) -->
      <div class="grid grid-cols-4 gap-4 mb-6 bg-slate-50 p-4 border border-slate-300 rounded text-center">
        <div class="border-r border-slate-200 pr-2">
          <div class="text-[9px] uppercase font-bold text-slate-500 font-mono">Uptime SLA Rata-Rata</div>
          <div class="text-lg font-black text-emerald-700 font-mono mt-1">{{ avgSlaUptime?.toFixed(2) }}%</div>
          <div class="text-[8px] text-slate-400 font-mono">Target: 99.50%</div>
        </div>
        <div class="border-r border-slate-200 pr-2">
          <div class="text-[9px] uppercase font-bold text-slate-500 font-mono">Total Kejadian Outage</div>
          <div class="text-lg font-black text-slate-900 font-mono mt-1">{{ totalOutageEvents }} Kali</div>
          <div class="text-[8px] text-slate-400 font-mono">Terdeteksi otomatis</div>
        </div>
        <div class="border-r border-slate-200 pr-2">
          <div class="text-[9px] uppercase font-bold text-slate-500 font-mono">Rata-Rata MTTR</div>
          <div class="text-lg font-black text-indigo-900 font-mono mt-1">{{ avgMttrMinutes?.toFixed(0) }} Menit</div>
          <div class="text-[8px] text-slate-400 font-mono">Respon pemulihan</div>
        </div>
        <div>
          <div class="text-[9px] uppercase font-bold text-slate-500 font-mono">Delivery Notifikasi</div>
          <div class="text-lg font-black text-sky-700 font-mono mt-1">{{ alertDeliveryRate }}%</div>
          <div class="text-[8px] text-slate-400 font-mono">WA + Telegram</div>
        </div>
      </div>

      <!-- Tabel Audit SLA per Perangkat -->
      <div class="mb-6 break-inside-avoid page-break-inside-avoid">
        <h3 class="font-bold text-xs uppercase tracking-wider text-slate-800 font-mono mb-2">2. DETAIL KINERJA &amp; DOWNTIME PER PERANGKAT</h3>
        <table class="w-full text-left border-collapse border border-slate-300 text-xs" style="table-layout: fixed; width: 100%;">
          <colgroup>
            <col style="width: 5%;">
            <col style="width: 30%;">
            <col style="width: 20%;">
            <col style="width: 25%;">
            <col style="width: 20%;">
          </colgroup>
          <thead>
            <tr class="bg-slate-100 font-bold text-slate-700 font-mono uppercase text-[10px]">
              <th class="border border-slate-300 p-2">#</th>
              <th class="border border-slate-300 p-2">Nama Perangkat</th>
              <th class="border border-slate-300 p-2">Tipe</th>
              <th class="border border-slate-300 p-2">Lokasi Fisik</th>
              <th class="border border-slate-300 p-2 text-center">Total Downtime</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200">
            <tr v-for="(row, idx) in rows" :key="row.deviceId" class="break-inside-avoid page-break-inside-avoid">
              <td class="border border-slate-300 p-2 font-mono text-[10px]">{{ idx + 1 }}</td>
              <td class="border border-slate-300 p-2 font-bold text-slate-800">{{ row.deviceName }}</td>
              <td class="border border-slate-300 p-2 font-mono text-slate-600">{{ row.deviceType }}</td>
              <td class="border border-slate-300 p-2 text-slate-600">{{ row.location }}</td>
              <td class="border border-slate-300 p-2 text-center font-mono font-bold text-red-700">
                {{ formatDowntime(row.totalDowntimeMinutes) }} ({{ row.downCount }}x down)
              </td>
            </tr>
            <tr v-if="!rows || rows.length === 0">
              <td colspan="5" class="border border-slate-300 p-4 text-center text-slate-400 font-mono">Tidak ada data performa perangkat.</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Tabel Pengesahan Dokumen -->
      <div class="mt-12 pt-6 border-t-2 border-slate-300 grid grid-cols-2 gap-12 text-center text-xs break-inside-avoid page-break-inside-avoid">
        <div>
          <p class="text-slate-600 font-mono text-[10px] mb-12">Disusun Oleh: Pengelola Infrastruktur NOC</p>
          <div class="border-b border-slate-900 w-44 mx-auto mb-1"></div>
          <p class="font-bold text-slate-900">Spesialis Jaringan &amp; Sistem</p>
        </div>
        <div>
          <p class="text-slate-600 font-mono text-[10px] mb-12">Disetujui Oleh: Kepala Biro / Sub-Bagian IT</p>
          <div class="border-b border-slate-900 w-44 mx-auto mb-1"></div>
          <p class="font-bold text-slate-900">Kepala Sub-Bagian IT Diskominfo</p>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { ReportRow } from '../../types';

const props = defineProps<{
  period?: string;
  avgSlaUptime?: number;
  totalOutageEvents?: number;
  avgMttrMinutes?: number;
  alertDeliveryRate?: number;
  rows?: ReportRow[];
}>();

const periodLabel = computed(() => {
  if (props.period === 'daily') return 'Harian (24 Jam Terakhir)';
  if (props.period === 'weekly') return 'Mingguan (7 Hari Terakhir)';
  if (props.period === 'custom') return 'Rentang Tanggal Kustom';
  return 'Bulanan (30 Hari Terakhir)';
});

function formatDowntime(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h > 0) return `${h}j ${m}m`;
  return `${m}m`;
}
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

  #printable-sla-audit {
    display: block !important;
    width: 100% !important;
    max-width: 100% !important;
    padding: 0 !important;
    margin: 0 !important;
    border: none !important;
    box-shadow: none !important;
  }

  /* Prevent page split inside elements */
  .break-inside-avoid {
    break-inside: avoid !important;
    page-break-inside: avoid !important;
  }

  tr, .grid, .mb-6 {
    break-inside: avoid !important;
    page-break-inside: avoid !important;
  }

  h1, h2, h3, h4, h5, h6 {
    break-after: avoid !important;
    page-break-after: avoid !important;
  }

  /* Wrap text for table cells */
  table {
    table-layout: fixed !important;
    width: 100% !important;
    border-collapse: collapse !important;
  }

  th, td {
    word-wrap: break-word !important;
    word-break: break-all !important;
    overflow-wrap: break-word !important;
    white-space: normal !important;
  }
}
</style>

