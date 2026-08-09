<template>
  <Teleport to="body">
    <div id="printable-incidents-list" class="hidden print:block bg-white text-slate-900 p-10 font-sans leading-normal text-xs max-w-4xl mx-auto border border-slate-300 shadow-none">
      <!-- Kop Surat Header -->
      <div class="border-b-2 border-slate-900 pb-4 mb-6 flex justify-between items-start">
        <div>
          <h2 class="text-sm font-extrabold tracking-wider uppercase text-slate-900">PEMERINTAH PROVINSI JAWA BARAT</h2>
          <h1 class="text-lg font-black tracking-tight text-indigo-900">DINAS KOMUNIKASI DAN INFORMATIKA — GOVMONITOR IT</h1>
          <p class="text-[10px] font-mono text-slate-600 mt-0.5">Daftar Antrean Kejadian &amp; Outage Infrastruktur NOC</p>
        </div>
        <div class="text-right font-mono text-[10px] text-slate-600 bg-slate-100 p-2.5 rounded border border-slate-200">
          <div><strong class="text-slate-900">JUMLAH TIKET:</strong> {{ incidents.length }}</div>
          <div><strong class="text-slate-900">TANGGAL CETAK:</strong> {{ new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) }}</div>
        </div>
      </div>

      <!-- Incidents list -->
      <div class="mb-6 break-inside-avoid page-break-inside-avoid">
        <h3 class="font-bold text-xs uppercase tracking-wider text-slate-800 font-mono mb-2">DETAIL DAFTAR TIKET OUTAGE</h3>
        
        <!-- Grouped Render -->
        <template v-if="groupingMode !== 'none'">
          <div v-for="group in groupedIncidents" :key="group.name" class="mb-4 break-inside-avoid page-break-inside-avoid">
            <h4 class="bg-slate-100 border border-slate-300 p-2 font-bold text-xs text-indigo-950 font-mono flex items-center justify-between mb-1">
              <span>{{ group.name }}</span>
              <span>({{ group.items.length }} Tiket)</span>
            </h4>
            <table class="w-full text-left border-collapse border border-slate-300 text-xs mb-3" style="table-layout: fixed; width: 100%;">
              <colgroup>
                <col style="width: 15%;" />
                <col style="width: 25%;" />
                <col style="width: 15%;" />
                <col style="width: 15%;" />
                <col style="width: 15%;" />
                <col style="width: 15%;" />
              </colgroup>
              <thead>
                <tr class="bg-slate-50 font-bold text-slate-700 font-mono uppercase text-[9px] border-b border-slate-300">
                  <th class="border border-slate-300 p-2">Ticket ID</th>
                  <th class="border border-slate-300 p-2">Nama Perangkat</th>
                  <th class="border border-slate-300 p-2">Tipe</th>
                  <th class="border border-slate-300 p-2">IP Address</th>
                  <th class="border border-slate-300 p-2 text-center">Durasi</th>
                  <th class="border border-slate-300 p-2 text-center">Status</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-200">
                <tr v-for="inc in group.items" :key="inc.id" class="break-inside-avoid page-break-inside-avoid text-[11px]">
                  <td class="border border-slate-300 p-2 font-mono font-bold text-indigo-900">{{ inc.id }}</td>
                  <td class="border border-slate-300 p-2 font-bold text-slate-800">{{ inc.deviceName }}</td>
                  <td class="border border-slate-300 p-2 font-mono text-slate-600">{{ inc.deviceType }}</td>
                  <td class="border border-slate-300 p-2 font-mono text-slate-600">{{ inc.deviceIp }}</td>
                  <td class="border border-slate-300 p-2 text-center font-mono text-red-700 font-semibold">{{ inc.duration }}</td>
                  <td class="border border-slate-300 p-2 text-center">
                    <span class="px-1.5 py-0.5 rounded font-mono font-bold text-[9px] uppercase border"
                      :class="inc.status === 'ACTIVE' ? 'bg-red-100 text-red-700 border-red-300' : 'bg-emerald-100 text-emerald-700 border-emerald-300'">
                      {{ inc.status }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>

        <!-- Flat Render -->
        <template v-else>
          <table class="w-full text-left border-collapse border border-slate-300 text-xs" style="table-layout: fixed; width: 100%;">
            <colgroup>
              <col style="width: 15%;" />
              <col style="width: 25%;" />
              <col style="width: 15%;" />
              <col style="width: 15%;" />
              <col style="width: 15%;" />
              <col style="width: 15%;" />
            </colgroup>
            <thead>
              <tr class="bg-slate-100 font-bold text-slate-700 font-mono uppercase text-[10px]">
                <th class="border border-slate-300 p-2">Ticket ID</th>
                <th class="border border-slate-300 p-2">Nama Perangkat</th>
                <th class="border border-slate-300 p-2">Tipe</th>
                <th class="border border-slate-300 p-2">IP Address</th>
                <th class="border border-slate-300 p-2 text-center">Durasi</th>
                <th class="border border-slate-300 p-2 text-center">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200">
              <tr v-for="inc in incidents" :key="inc.id" class="break-inside-avoid page-break-inside-avoid text-[11px]">
                <td class="border border-slate-300 p-2 font-mono font-bold text-indigo-900">{{ inc.id }}</td>
                <td class="border border-slate-300 p-2 font-bold text-slate-800">{{ inc.deviceName }}</td>
                <td class="border border-slate-300 p-2 font-mono text-slate-600">{{ inc.deviceType }}</td>
                <td class="border border-slate-300 p-2 font-mono text-slate-600">{{ inc.deviceIp }}</td>
                <td class="border border-slate-300 p-2 text-center font-mono text-red-700 font-semibold">{{ inc.duration }}</td>
                <td class="border border-slate-300 p-2 text-center">
                  <span class="px-1.5 py-0.5 rounded font-mono font-bold text-[9px] uppercase border"
                    :class="inc.status === 'ACTIVE' ? 'bg-red-100 text-red-700 border-red-300' : 'bg-emerald-100 text-emerald-700 border-emerald-300'">
                    {{ inc.status }}
                  </span>
                </td>
              </tr>
              <tr v-if="!incidents || incidents.length === 0">
                <td colspan="6" class="border border-slate-300 p-4 text-center text-slate-400 font-mono">Tidak ada data kejadian.</td>
              </tr>
            </tbody>
          </table>
        </template>
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
import type { Incident } from '../../types';

defineProps<{
  incidents: Incident[];
  groupingMode: string;
  groupedIncidents: { name: string; items: Incident[] }[];
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

  #printable-incidents-list {
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

  tr, .grid, .mb-4, .mb-6 {
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
