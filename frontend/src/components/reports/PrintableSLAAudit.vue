<template>
  <Teleport to="body">
    <div id="printable-sla-audit" class="hidden print:block bg-white text-slate-900 p-0 print:p-0 font-sans leading-normal text-xs w-full mx-auto border-0 shadow-none">
      <!-- Kop Surat Header -->
      <div class="border-b-2 border-slate-900 pb-4 mb-6 flex justify-between items-start">
        <div>
          <h2 class="text-sm font-extrabold tracking-wider uppercase text-slate-900">PEMERINTAH PROVINSI JAWA BARAT</h2>
          <h1 class="text-lg font-black tracking-tight text-indigo-900">DINAS KOMUNIKASI DAN INFORMATIKA — SANOC</h1>
          <p class="text-[10px] font-mono text-slate-600 mt-0.5">Laporan Kepatuhan SLA Uptime &amp; Audit Kesehatan Infrastruktur SANOC</p>
        </div>
        <div class="text-right font-mono text-[10px] text-slate-600 bg-slate-100 p-2.5 rounded border border-slate-200">
          <div><strong class="text-slate-900">PERIODE AUDIT:</strong> {{ periodLabel }}</div>
          <div><strong class="text-slate-900">TAB LAPORAN:</strong> {{ tabLabel }}</div>
          <div><strong class="text-slate-900">TANGGAL CETAK:</strong> {{ new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) }}</div>
        </div>
      </div>

      <!-- Ringkasan SLA Metrics (Grid) -->
      <div class="grid grid-cols-4 gap-4 mb-3 bg-slate-50 p-3 border border-slate-300 rounded text-center break-inside-avoid">
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

      <!-- TAB 1: Downtime by Device -->
      <div v-if="activeTab === 'downtime' || !activeTab" class="mb-6">
        <h3 class="font-bold text-xs uppercase tracking-wider text-slate-800 font-mono mb-2">DETAIL KINERJA &amp; DOWNTIME PER PERANGKAT</h3>
        <table class="w-full text-left border-collapse border border-slate-300 text-xs">
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
            <tr v-for="(row, idx) in rows" :key="row.deviceId">
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

      <!-- TAB 2: Recurring Issues -->
      <div v-else-if="activeTab === 'recurring'" class="mb-6">
        <h3 class="font-bold text-xs uppercase tracking-wider text-slate-800 font-mono mb-2">DAFTAR PERANGKAT RECURRING ISSUES (FLAPPING)</h3>
        <table class="w-full text-left border-collapse border border-slate-300 text-xs">
          <colgroup>
            <col style="width: 5%;">
            <col style="width: 30%;">
            <col style="width: 20%;">
            <col style="width: 25%;">
            <col style="width: 20%;">
          </colgroup>
          <thead>
            <tr class="bg-amber-100 font-bold text-slate-800 font-mono uppercase text-[10px]">
              <th class="border border-slate-300 p-2">#</th>
              <th class="border border-slate-300 p-2">Nama Perangkat</th>
              <th class="border border-slate-300 p-2">Tipe</th>
              <th class="border border-slate-300 p-2">Lokasi</th>
              <th class="border border-slate-300 p-2 text-center">Down (7 Hari)</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200">
            <tr v-for="(dev, idx) in flapDevices" :key="dev.deviceId">
              <td class="border border-slate-300 p-2 font-mono text-[10px]">{{ idx + 1 }}</td>
              <td class="border border-slate-300 p-2 font-bold text-slate-900">{{ dev.deviceName }}</td>
              <td class="border border-slate-300 p-2 font-mono text-slate-600">{{ dev.deviceType }}</td>
              <td class="border border-slate-300 p-2 text-slate-600">{{ dev.location }}</td>
              <td class="border border-slate-300 p-2 text-center font-mono font-extrabold text-amber-700">
                {{ dev.downCount7d }}x down
              </td>
            </tr>
            <tr v-if="!flapDevices || flapDevices.length === 0">
              <td colspan="5" class="border border-slate-300 p-4 text-center text-slate-400 font-mono">Tidak ada perangkat yang terdeteksi flapping / recurring issue.</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- TAB 3: Active Incidents -->
      <div v-else-if="activeTab === 'active_incidents'" class="mb-6">
        <h3 class="font-bold text-xs uppercase tracking-wider text-slate-800 font-mono mb-2">DAFTAR ANTREAN TIKET INCIDENT OUTAGE</h3>
        <table class="w-full text-left border-collapse border border-slate-300 text-xs">
          <colgroup>
            <col style="width: 15%;">
            <col style="width: 25%;">
            <col style="width: 15%;">
            <col style="width: 15%;">
            <col style="width: 15%;">
            <col style="width: 15%;">
          </colgroup>
          <thead>
            <tr class="bg-red-100 font-bold text-slate-800 font-mono uppercase text-[10px]">
              <th class="border border-slate-300 p-2">Ticket ID</th>
              <th class="border border-slate-300 p-2">Perangkat</th>
              <th class="border border-slate-300 p-2">Tipe</th>
              <th class="border border-slate-300 p-2">IP Address</th>
              <th class="border border-slate-300 p-2 text-center">Durasi</th>
              <th class="border border-slate-300 p-2 text-center">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200">
            <tr v-for="inc in incidents" :key="inc.id">
              <td class="border border-slate-300 p-2 font-mono font-bold text-indigo-900">{{ inc.id }}</td>
              <td class="border border-slate-300 p-2 font-bold text-slate-800">{{ inc.deviceName }}</td>
              <td class="border border-slate-300 p-2 font-mono text-slate-600">{{ inc.deviceType }}</td>
              <td class="border border-slate-300 p-2 font-mono text-slate-600">{{ inc.deviceIp }}</td>
              <td class="border border-slate-300 p-2 text-center font-mono font-semibold text-red-700">{{ inc.duration }}</td>
              <td class="border border-slate-300 p-2 text-center font-mono font-bold text-[10px]">
                <span class="px-1.5 py-0.5 rounded border" :class="inc.status === 'ACTIVE' ? 'bg-red-100 text-red-700 border-red-300' : 'bg-emerald-100 text-emerald-700 border-emerald-300'">
                  {{ inc.status }}
                </span>
              </td>
            </tr>
            <tr v-if="!incidents || incidents.length === 0">
              <td colspan="6" class="border border-slate-300 p-4 text-center text-slate-400 font-mono">Tidak ada antrean incident aktif.</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Tabel Pengesahan Dokumen -->
      <div class="mt-12 pt-6 border-t-2 border-slate-300 grid grid-cols-2 gap-12 text-center text-xs break-inside-avoid">
        <div>
          <p class="text-slate-600 font-mono text-[10px] mb-12">Disusun Oleh: Pengelola Infrastruktur SANOC</p>
          <div class="border-b border-slate-900 w-44 mx-auto mb-1"></div>
          <p class="font-bold text-slate-900">Spesialis Jaringan &amp; Sistem</p>
        </div>
        <div>
          <p class="text-slate-600 font-mono text-[10px] mb-12">Disetujui Oleh: Kepala Biro / Sub-Bagian IT</p>
          <div class="border-b border-slate-900 w-44 mx-auto mb-1"></div>
          <p class="font-bold text-slate-900">Kepala Sub-Bagian IT Diskominfo</p>
        </div>
      </div>
      <div class="mt-8 pt-2 border-t border-slate-200 text-center font-mono text-[9px] text-slate-500">
        © SANOC Team — UTB 2026.
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { ReportRow, FlapDevice, Incident } from '../../types';

const props = defineProps<{
  period?: string;
  activeTab?: 'downtime' | 'recurring' | 'active_incidents';
  avgSlaUptime?: number;
  totalOutageEvents?: number;
  avgMttrMinutes?: number;
  alertDeliveryRate?: number;
  rows?: ReportRow[];
  flapDevices?: FlapDevice[];
  incidents?: Incident[];
}>();

const periodLabel = computed(() => {
  if (props.period === 'daily') return 'Harian (24 Jam Terakhir)';
  if (props.period === 'weekly') return 'Mingguan (7 Hari Terakhir)';
  if (props.period === 'custom') return 'Rentang Tanggal Kustom';
  return 'Bulanan (30 Hari Terakhir)';
});

const tabLabel = computed(() => {
  if (props.activeTab === 'recurring') return 'Recurring Issues (Flapping)';
  if (props.activeTab === 'active_incidents') return 'Daftar Active Incidents';
  return 'Downtime by Device';
});

function formatDowntime(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h > 0) return `${h}j ${m}m`;
  return `${m}m`;
}
</script>
