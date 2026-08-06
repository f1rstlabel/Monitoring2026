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
      <div class="grid grid-cols-3 gap-4 mb-6 bg-slate-50 p-4 border border-slate-300 rounded text-center">
        <div class="border-r border-slate-200 pr-2">
          <div class="text-[10px] uppercase font-bold text-slate-500 font-mono">Rata-Rata Uptime SLA</div>
          <div class="text-xl font-black text-emerald-700 font-mono mt-1">99.94%</div>
          <div class="text-[9px] text-slate-400 font-mono">Target Standar: 99.50%</div>
        </div>
        <div class="border-r border-slate-200 pr-2">
          <div class="text-[10px] uppercase font-bold text-slate-500 font-mono">Total Kejadian Outage</div>
          <div class="text-xl font-black text-slate-900 font-mono mt-1">3 Kejadian</div>
          <div class="text-[9px] text-slate-400 font-mono">Terdeteksi otomatis</div>
        </div>
        <div>
          <div class="text-[10px] uppercase font-bold text-slate-500 font-mono">Rata-Rata Waktu Pulih (MTTR)</div>
          <div class="text-xl font-black text-indigo-900 font-mono mt-1">08m 45s</div>
          <div class="text-[9px] text-slate-400 font-mono">Respon penanganan cepat</div>
        </div>
      </div>

      <!-- Tabel Pengesahan Dokumen -->
      <div class="mt-12 pt-6 border-t-2 border-slate-300 grid grid-cols-2 gap-12 text-center text-xs">
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

const props = defineProps<{
  period?: string;
}>();

const periodLabel = computed(() => {
  if (props.period === 'daily') return 'Harian (24 Jam Terakhir)';
  if (props.period === 'weekly') return 'Mingguan (7 Hari Terakhir)';
  if (props.period === 'custom') return 'Rentang Tanggal Kustom';
  return 'Bulanan (30 Hari Terakhir)';
});
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
  .grid, .mb-6 {
    break-inside: avoid !important;
    page-break-inside: avoid !important;
  }

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

