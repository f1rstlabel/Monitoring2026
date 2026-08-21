<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-[#26262A] pb-4">
      <div>
        <h1 class="text-xl font-extrabold text-white tracking-tight">Outage List & Incident Tickets</h1>
        <p class="text-xs text-gray-400 mt-1">Downtime history, notification escalation logs, and recovery metrics</p>
      </div>

      <div class="flex items-center gap-3">
        <span class="px-3 py-1 rounded-full bg-red-500/15 border border-red-500/30 text-red-400 text-xs font-mono font-bold">
          {{ incidentStore.incidents.filter((i: Incident) => i.status === 'ACTIVE').length }} Active Outages
        </span>

        <!-- Export Excel / CSV Dropdown -->
        <div v-if="authStore.hasPermission('reports.export')" class="relative">
          <button
            @click="showExportDropdown = !showExportDropdown"
            class="px-3.5 py-1.5 rounded-lg border border-[#26262A] bg-[#151517] hover:bg-[#1E1E22] text-emerald-400 font-semibold text-xs transition-all flex items-center gap-1.5 shadow-sm cursor-pointer"
            title="Export Incidents to Excel / CSV"
          >
            <FileSpreadsheet class="w-3.5 h-3.5 text-emerald-400" />
            <span>Export Data</span>
            <ChevronDown class="w-3 h-3 text-emerald-500/70 ml-0.5" />
          </button>

          <!-- Dropdown Options -->
          <div
            v-if="showExportDropdown"
            @click="showExportDropdown = false"
            class="absolute right-0 mt-1.5 w-52 bg-[#1A1A1E] border border-[#26262A] rounded-xl shadow-2xl py-1.5 z-50 animate-in fade-in slide-in-from-top-1 duration-150"
          >
            <div class="px-3 py-1.5 border-b border-[#26262A]/60 font-mono text-[10px] uppercase font-bold text-gray-400">
              Select Export Format
            </div>
            <button
              @click="exportIncidentsData('xls')"
              class="w-full text-left px-3.5 py-2 text-xs font-mono text-gray-200 hover:bg-emerald-500/10 hover:text-emerald-400 flex items-center gap-2.5 transition-colors cursor-pointer"
            >
              <FileSpreadsheet class="w-4 h-4 text-emerald-400 shrink-0" />
              <div>
                <div class="font-bold">Excel Spreadsheet (.xls)</div>
                <div class="text-[10px] text-gray-400 font-sans">Structured Excel workbook</div>
              </div>
            </button>
            <button
              @click="exportIncidentsData('csv')"
              class="w-full text-left px-3.5 py-2 text-xs font-mono text-gray-200 hover:bg-emerald-500/10 hover:text-emerald-400 flex items-center gap-2.5 transition-colors cursor-pointer"
            >
              <FileText class="w-4 h-4 text-sky-400 shrink-0" />
              <div>
                <div class="font-bold">CSV File (.csv)</div>
                <div class="text-[10px] text-gray-400 font-sans">Standard UTF-8 CSV document</div>
              </div>
            </button>
          </div>
        </div>

        <button
          @click="exportIncidentsPDF"
          class="px-3 py-1.5 rounded-lg border border-[#26262A] bg-[#151517] hover:bg-[#1E1E22] text-sky-400 font-semibold text-xs transition-all flex items-center gap-1.5 cursor-pointer"
          title="Print / Save Incidents as PDF"
        >
          <Printer class="w-3.5 h-3.5 text-sky-400" />
          Export PDF
        </button>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-4 flex flex-wrap items-center justify-between gap-4">
      <div class="flex items-center gap-3 flex-1 min-w-[320px]">
        <div class="relative flex-1">
          <Search class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by device name or IP address..."
            class="w-full bg-[#18181B] border border-[#26262A] rounded-lg pl-9 pr-4 py-1.5 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5]"
          />
        </div>
        <select
          v-model="statusFilter"
          class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-1.5 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
        >
          <option value="ALL">All Statuses</option>
          <option value="ACTIVE">Active (Open)</option>
          <option value="RESOLVED">Resolved</option>
        </select>
        <select
          v-model="groupingMode"
          class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-1.5 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
        >
          <option value="none">No Grouping / Flat List</option>
          <option value="location">Group by Location</option>
          <option value="device">Group by Device</option>
          <option value="status">Group by Status</option>
        </select>
      </div>

      <span class="text-xs font-mono text-gray-400">
        Showing <span class="text-white font-bold">{{ filteredIncidents.length }}</span> of {{ incidentStore.totalCount || incidentStore.incidents.length }} incidents
      </span>
    </div>

    <!-- Skeleton while loading -->
    <template v-if="incidentStore.isLoading">
      <div v-if="groupingMode !== 'none'" class="space-y-4">
        <div v-for="g in 3" :key="g" class="bg-[#151517] border border-[#26262A] rounded-xl p-4 space-y-3">
          <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
            <Skeleton width="35%" height="1.1rem" />
            <Skeleton width="15%" height="0.8rem" />
          </div>
          <SkeletonTable :rows="2" :cols="6" />
        </div>
      </div>
      <SkeletonTable v-else :rows="6" :cols="7" />
    </template>

    <!-- Grouped Accordion List (by Location, Device, Status) -->
    <div v-else-if="groupingMode !== 'none'" class="space-y-4">
      <div v-if="paginatedGroupedIncidents.length === 0" class="bg-[#151517] border border-[#26262A] rounded-xl p-12 text-center">
        <p class="text-sm font-semibold text-gray-300 font-mono">No incidents match your filters</p>
      </div>

      <div
        v-for="group in paginatedGroupedIncidents"
        :key="group.name"
        class="bg-[#151517] border rounded-xl overflow-hidden shadow-xl transition-all"
        :class="group.activeCount > 0 ? 'border-red-500/30' : 'border-[#26262A]'"
      >
        <!-- Group Header -->
        <div
          @click="toggleGroupExpand(group.name)"
          class="flex items-center justify-between p-4 cursor-pointer hover:bg-[#1C1C20] transition-colors"
        >
          <div class="flex items-center gap-3">
            <span
              class="w-2.5 h-2.5 rounded-full"
              :class="group.activeCount > 0 ? 'bg-red-500 pulsing-dot-red' : 'bg-emerald-500'"
            ></span>
            <div>
              <h3 class="text-sm font-bold text-gray-100 flex items-center gap-2">
                {{ group.name }}
                <span class="text-xs font-mono font-normal text-gray-500">
                  ({{ group.items.length }} Incidents)
                </span>
              </h3>
              <p class="text-[10px] font-mono text-gray-500 mt-0.5">
                {{ group.activeCount }} ACTIVE OUTAGE(S) &bull; {{ group.resolvedCount }} RESOLVED
              </p>
            </div>
          </div>

          <div class="flex items-center gap-3">
            <span
              class="px-2 py-0.5 rounded text-[9px] font-mono font-bold uppercase border"
              :class="group.activeCount > 0 ? 'bg-red-500/15 text-red-400 border-red-500/30' : 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30'"
            >
              {{ group.activeCount > 0 ? `${group.activeCount} OUTAGE` : 'RESOLVED' }}
            </span>
            <ChevronRight
              class="w-4 h-4 text-gray-400 transition-transform duration-200"
              :class="expandedGroups[group.name] !== false ? 'rotate-180' : ''"
            />
          </div>
        </div>

        <!-- Group Table Content -->
        <div v-if="expandedGroups[group.name] !== false" class="overflow-x-auto border-t border-[#26262A]">
          <table class="w-full text-left text-xs text-gray-300">
            <thead class="bg-[#18181B] border-b border-[#26262A] font-mono text-[10px] uppercase text-gray-400">
              <tr>
                <th class="py-3.5 px-4">Ticket ID</th>
                <th class="py-3.5 px-4">Device Name</th>
                <th class="py-3.5 px-4">Type</th>
                <th class="py-3.5 px-4">IP Address</th>
                <th class="py-3.5 px-4">Duration</th>
                <th class="py-3.5 px-4">Affected</th>
                <th class="py-3.5 px-4">Status</th>
                <th class="py-3.5 px-4 text-right">Action</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[#26262A]">
              <tr
                v-for="inc in group.items"
                :key="inc.id"
                @click="router.push(`/incidents/${inc.id}`)"
                class="hover:bg-[#18181B] transition-colors cursor-pointer group"
                :class="{
                  'border-l-2 border-l-[#F16565] bg-red-500/5': inc.status === 'ACTIVE'
                }"
              >
                <td class="py-3.5 px-4 font-mono font-bold text-[#7B96F5] group-hover:underline">
                  {{ inc.id }}
                </td>
                <td class="py-3.5 px-4 font-bold text-white">{{ inc.deviceName }}</td>
                <td class="py-3.5 px-4 font-mono text-gray-400">{{ inc.deviceType }}</td>
                <td class="py-3.5 px-4 font-mono text-gray-300">{{ inc.deviceIp }}</td>
                <td class="py-3.5 px-4 font-mono text-red-400 font-semibold">{{ inc.duration }}</td>
                <td class="py-3.5 px-4 font-mono text-amber-400">{{ inc.affectedDevicesCount }} Nodes</td>
                <td class="py-3.5 px-4">
                  <span
                    class="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold uppercase border"
                    :class="inc.status === 'ACTIVE'
                      ? 'bg-red-500/15 text-red-400 border-red-500/30 animate-pulse'
                      : 'bg-[#3ECF8E]/15 text-[#3ECF8E] border-[#3ECF8E]/30'"
                  >
                    {{ inc.status }}
                  </span>
                </td>
                <td class="py-3.5 px-4 text-right" @click.stop>
                  <div class="flex items-center justify-end gap-1">
                    <button
                      @click.stop="router.push(`/incidents/${inc.id}`)"
                      class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-[#26262A] transition-colors"
                    >
                      <ChevronRight class="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Incident List Table (Flat View) -->
    <div v-else class="bg-[#151517] border border-[#26262A] rounded-xl overflow-hidden shadow-xl">
      <table class="w-full text-left text-xs text-gray-300">
        <thead class="bg-[#18181B] border-b border-[#26262A] font-mono text-[10px] uppercase text-gray-400">
          <tr>
            <th class="py-3.5 px-4">Ticket ID</th>
            <th class="py-3.5 px-4">Device Name</th>
            <th class="py-3.5 px-4">Type</th>
            <th class="py-3.5 px-4">IP Address</th>
            <th class="py-3.5 px-4">Duration</th>
            <th class="py-3.5 px-4">Affected</th>
            <th class="py-3.5 px-4">Status</th>
            <th class="py-3.5 px-4 text-right">Action</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#26262A]">
          <tr v-if="incidentStore.isLoading">
            <td colspan="8" class="p-0 border-0">
              <SkeletonTable :rows="6" :cols="8" />
            </td>
          </tr>
          <template v-else-if="filteredIncidents.length > 0">
            <tr
              v-for="inc in filteredIncidents"
              :key="inc.id"
              @click="router.push(`/incidents/${inc.id}`)"
              class="hover:bg-[#18181B] transition-colors cursor-pointer group"
              :class="{
                'border-l-2 border-l-[#F16565] bg-red-500/5': inc.status === 'ACTIVE'
              }"
            >
              <td class="py-3.5 px-4 font-mono font-bold text-[#7B96F5] group-hover:underline">
                {{ inc.id }}
              </td>
              <td class="py-3.5 px-4 font-bold text-white">{{ inc.deviceName }}</td>
              <td class="py-3.5 px-4 font-mono text-gray-400">{{ inc.deviceType }}</td>
              <td class="py-3.5 px-4 font-mono text-gray-300">{{ inc.deviceIp }}</td>
              <td class="py-3.5 px-4 font-mono text-red-400 font-semibold">{{ inc.duration }}</td>
              <td class="py-3.5 px-4 font-mono text-amber-400">{{ inc.affectedDevicesCount }} Nodes</td>
              <td class="py-3.5 px-4">
                <span
                  class="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold uppercase border"
                  :class="inc.status === 'ACTIVE'
                    ? 'bg-red-500/15 text-red-400 border-red-500/30 animate-pulse'
                    : 'bg-[#3ECF8E]/15 text-[#3ECF8E] border-[#3ECF8E]/30'"
                >
                  {{ inc.status }}
                </span>
              </td>
              <td class="py-3.5 px-4 text-right" @click.stop>
                <div class="flex items-center justify-end gap-1">
                  <button
                    @click.stop="router.push(`/incidents/${inc.id}`)"
                    class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-[#26262A] transition-colors"
                  >
                    <ChevronRight class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </template>
          <tr v-else>
            <td colspan="8" class="py-14 text-center">
              <div class="flex flex-col items-center gap-3 text-gray-600">
                <CheckCircle class="w-8 h-8" />
                <p class="text-sm font-semibold">All Clear — No Active Incidents</p>
                <p class="text-xs">All monitored nodes are operating normally</p>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination Control -->
    <PaginationControl
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :total="paginatedTotal"
    />

    <!-- Hidden Printable Report element for PDF output -->
    <PrintableIncidentsList
      v-if="isPrintRendered"
      :incidents="filteredIncidents"
      :groupingMode="groupingMode"
      :groupedIncidents="groupedIncidents"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useIncidentStore } from '../stores/incidentStore';
import { useAuthStore } from '../stores/authStore';
import SkeletonTable from '../components/common/SkeletonTable.vue';
import Skeleton from '../components/common/Skeleton.vue';
import PaginationControl from '../components/common/PaginationControl.vue';
import PrintableIncidentsList from '../components/reports/PrintableIncidentsList.vue';
import { ChevronRight, CheckCircle, FileText, Printer, Search, ChevronDown, FileSpreadsheet } from 'lucide-vue-next';
import { downloadCSV, downloadXLS } from '../utils/exportUtils';
import type { Incident } from '../types';

const currentPage = ref(1);
const pageSize = ref(10);

const router = useRouter();
const route = useRoute();
const incidentStore = useIncidentStore();
const authStore = useAuthStore();

const searchQuery = ref('');
const statusFilter = ref('ALL');
const groupingMode = ref<'none' | 'location' | 'device' | 'status'>('none');
const expandedGroups = ref<Record<string, boolean>>({});
const showExportDropdown = ref(false);
const isPrintRendered = ref(false);

function exportIncidentsData(format: 'csv' | 'xls' = 'xls') {
  const headers = ['Ticket ID', 'Nama Perangkat', 'Tipe Perangkat', 'IP Address', 'Durasi Outage', 'Jumlah Terdampak', 'Status Tiket'];
  const rows = filteredIncidents.value.map((i: Incident) => [
    i.id,
    i.deviceName,
    i.deviceType,
    i.deviceIp,
    i.duration,
    i.affectedDevicesCount,
    i.status
  ]);
  const dateStr = new Date().toISOString().slice(0, 10);
  const typePrefix = format === 'csv' ? 'sanoc-csv' : 'sanoc-excel';
  const filename = `${typePrefix}-incidents-list-${dateStr}`;

  if (format === 'csv') {
    downloadCSV(filename, headers, rows);
  } else {
    downloadXLS(filename, 'Daftar Antrean Outage Kejadian Infrastruktur NOC', headers, rows);
  }
}

async function exportIncidentsPDF() {
  const originalTitle = document.title;
  const dateStr = new Date().toISOString().slice(0, 10);
  document.title = `sanoc-pdf-incidents-list-${dateStr}`;
  try {
    isPrintRendered.value = true;
    await nextTick();
    await new Promise((resolve) => setTimeout(resolve, 100)); // allow rendering thread sync
    window.print();
  } finally {
    isPrintRendered.value = false;
    document.title = originalTitle;
  }
}

const filteredIncidents = computed(() => {
  return incidentStore.incidents.filter((inc: Incident) => {
    if (statusFilter.value !== 'ALL' && inc.status !== statusFilter.value) {
      return false;
    }
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.trim().toLowerCase();
      const matchId = (inc.id || '').toLowerCase().includes(q);
      const matchName = (inc.deviceName || '').toLowerCase().includes(q);
      const matchIp = (inc.deviceIp || '').toLowerCase().includes(q);
      if (!matchId && !matchName && !matchIp) return false;
    }
    return true;
  });
});

const paginatedTotal = computed(() => {
  if (groupingMode.value === 'none') {
    return incidentStore.totalCount;
  }
  return groupedIncidents.value.length;
});

const paginatedGroupedIncidents = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  return groupedIncidents.value.slice(start, start + pageSize.value);
});

const groupedIncidents = computed(() => {
  if (groupingMode.value === 'none') {
    return [];
  }

  const map: Record<string, { name: string; items: Incident[]; activeCount: number; resolvedCount: number }> = {};
  
  for (const inc of filteredIncidents.value) {
    let key = '';
    if (groupingMode.value === 'location') {
      key = (inc.location || 'Unassigned').trim();
    } else if (groupingMode.value === 'device') {
      key = (inc.deviceName || 'Unknown Device').trim();
    } else if (groupingMode.value === 'status') {
      key = (inc.status || 'ACTIVE').trim();
    }

    if (!map[key]) {
      map[key] = { name: key, items: [], activeCount: 0, resolvedCount: 0 };
    }
    map[key].items.push(inc);
    if (inc.status === 'ACTIVE') {
      map[key].activeCount++;
    } else {
      map[key].resolvedCount++;
    }
  }

  return Object.values(map).sort((a, b) => {
    if (b.activeCount !== a.activeCount) {
      return b.activeCount - a.activeCount;
    }
    return a.name.localeCompare(b.name);
  });
});

function toggleGroupExpand(name: string) {
  if (expandedGroups.value[name] === undefined) {
    expandedGroups.value[name] = true; // default expand when clicked first time
  } else {
    expandedGroups.value[name] = !expandedGroups.value[name];
  }
}

function loadIncidents() {
  if (groupingMode.value === 'none') {
    incidentStore.fetchIncidents({
      page: currentPage.value,
      pageSize: pageSize.value,
      status: statusFilter.value !== 'ALL' ? statusFilter.value : undefined,
      search: searchQuery.value || undefined
    });
  } else {
    // Grouped mode requires the full matching dataset
    incidentStore.fetchIncidents({
      status: statusFilter.value !== 'ALL' ? statusFilter.value : undefined,
      search: searchQuery.value || undefined
    });
  }
}

onMounted(() => {
  statusFilter.value = (route.query.status as string) || 'ALL';
  loadIncidents();
});

watch([currentPage, pageSize], () => {
  loadIncidents();
});

watch(
  () => route.query,
  (newQ) => {
    statusFilter.value = (newQ.status as string) || 'ALL';
  }
);

watch(
  () => [
    groupingMode.value,
    statusFilter.value,
    searchQuery.value
  ],
  () => {
    currentPage.value = 1;
    loadIncidents();
  }
);
</script>
