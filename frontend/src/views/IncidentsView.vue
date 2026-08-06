<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-[#26262A] pb-4">
      <div>
        <h1 class="text-xl font-extrabold text-white tracking-tight">Active Incident Queue</h1>
        <p class="text-xs text-gray-400 mt-1">Real-time infrastructure outage alerts requiring NOC intervention</p>
      </div>

      <div class="flex items-center gap-3">
        <span class="px-3 py-1 rounded-full bg-red-500/15 border border-red-500/30 text-red-400 text-xs font-mono font-bold">
          {{ incidentStore.incidents.filter((i: Incident) => i.status === 'ACTIVE').length }} Active
        </span>

        <button
          @click="exportIncidentsCSV"
          class="px-3 py-1.5 rounded-lg border border-[#26262A] bg-[#151517] hover:bg-[#1E1E22] text-emerald-400 font-semibold text-xs transition-all flex items-center gap-1.5"
          title="Export Incidents to Excel / CSV"
        >
          <FileText class="w-3.5 h-3.5 text-emerald-400" />
          Export Excel
        </button>

        <button
          @click="exportIncidentsPDF"
          class="px-3 py-1.5 rounded-lg border border-[#26262A] bg-[#151517] hover:bg-[#1E1E22] text-sky-400 font-semibold text-xs transition-all flex items-center gap-1.5"
          title="Print / Save Incidents as PDF"
        >
          <Printer class="w-3.5 h-3.5 text-sky-400" />
          Export PDF
        </button>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-4 flex flex-wrap items-center justify-between gap-4">
      <div class="flex items-center gap-3 flex-1 min-w-[240px]">
        <div class="relative flex-1">
          <Search class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Filter by Ticket ID, Device Name or IP..."
            class="w-full bg-[#18181B] border border-[#26262A] rounded-lg pl-9 pr-4 py-1.5 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5]"
          />
        </div>
        <select
          v-model="statusFilter"
          class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-1.5 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
        >
          <option value="ALL">All Statuses</option>
          <option value="ACTIVE">ACTIVE Outages Only</option>
          <option value="RESOLVED">RESOLVED Only</option>
        </select>
      </div>

      <span class="text-xs font-mono text-gray-400">
        Showing <span class="text-white font-bold">{{ filteredIncidents.length }}</span> of {{ incidentStore.incidents.length }} incidents
      </span>
    </div>

    <!-- Skeleton while loading -->
    <SkeletonTable v-if="incidentStore.isLoading" :rows="6" :cols="7" />

    <!-- Incident List Table -->
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
          <template v-if="filteredIncidents.length > 0">
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useIncidentStore } from '../stores/incidentStore';
import SkeletonTable from '../components/common/SkeletonTable.vue';
import { ChevronRight, CheckCircle, FileText, Printer, Search } from 'lucide-vue-next';
import type { Incident } from '../types';

const router = useRouter();
const incidentStore = useIncidentStore();

const searchQuery = ref('');
const statusFilter = ref('ALL');

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

function exportIncidentsCSV() {
  const headers = ['Ticket ID', 'Device Name', 'Type', 'IP Address', 'Duration', 'Affected Count', 'Status'];
  const rows = incidentStore.incidents.map((i: Incident) => [
    i.id,
    i.deviceName,
    i.deviceType,
    i.deviceIp,
    i.duration,
    i.affectedDevicesCount,
    i.status
  ]);
  const csv = [headers, ...rows].map(r => r.map((v: any) => `"${v}"`).join(',')).join('\n');
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = `govmonitor-incidents-${new Date().toISOString().slice(0, 10)}.csv`;
  link.click();
  URL.revokeObjectURL(url);
}

function exportIncidentsPDF() {
  window.print();
}

onMounted(() => {
  incidentStore.fetchIncidents();
});
</script>
