<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-[#26262A] pb-4">
      <div>
        <h1 class="text-xl font-extrabold text-white tracking-tight">Downtime &amp; Infrastructure Reports</h1>
        <p class="text-xs text-gray-400 mt-1">Exportable SLA performance audits and recurring issue tracking for Biro Umum</p>
      </div>

      <div class="flex items-center gap-3">
        <!-- Period Toggle -->
        <div class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5">
          <button
            @click="setPeriod('daily')"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-all"
            :class="reportStore.period === 'daily' ? 'bg-[#26262A] text-white shadow-sm' : 'text-gray-400 hover:text-gray-200'"
          >
            Harian (24h)
          </button>
          <button
            @click="setPeriod('weekly')"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-all"
            :class="reportStore.period === 'weekly' ? 'bg-[#26262A] text-white shadow-sm' : 'text-gray-400 hover:text-gray-200'"
          >
            Mingguan (7d)
          </button>
          <button
            @click="setPeriod('monthly')"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-all"
            :class="reportStore.period === 'monthly' ? 'bg-[#26262A] text-white shadow-sm' : 'text-gray-400 hover:text-gray-200'"
          >
            Bulanan (30d)
          </button>
          <button
            @click="setPeriod('custom')"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-all"
            :class="reportStore.period === 'custom' ? 'bg-[#26262A] text-white shadow-sm' : 'text-gray-400 hover:text-gray-200'"
          >
            Custom Range
          </button>
        </div>

        <!-- Export Excel / CSV -->
        <button
          @click="reportStore.exportCSV()"
          class="px-3.5 py-1.5 rounded-lg border border-[#26262A] bg-[#151517] hover:bg-[#1E1E22] text-emerald-400 font-semibold text-xs transition-all flex items-center gap-1.5"
          title="Export as Excel / CSV Spreadsheet"
        >
          <FileText class="w-3.5 h-3.5 text-emerald-400" />
          Export Excel
        </button>

        <!-- Export PDF -->
        <button
          @click="handlePrintPDF()"
          class="px-3.5 py-1.5 rounded-lg border border-[#26262A] bg-[#151517] hover:bg-[#1E1E22] text-sky-400 font-semibold text-xs transition-all flex items-center gap-1.5"
          title="Export SLA Audit as PDF"
        >
          <Printer class="w-3.5 h-3.5 text-sky-400" />
          Export PDF
        </button>
      </div>
    </div>

    <!-- Custom Date Range Bar & Type/Location Filters -->
    <div class="flex flex-wrap items-center gap-3">
      <!-- Custom Date Inputs -->
      <div v-if="reportStore.period === 'custom'" class="flex items-center gap-2 bg-[#151517] border border-[#26262A] rounded-lg px-3 py-1.5 text-xs text-gray-300">
        <Calendar class="w-3.5 h-3.5 text-[#7B96F5]" />
        <span class="text-[10px] uppercase font-mono text-gray-400">From:</span>
        <input
          v-model="reportStore.startDate"
          type="date"
          @change="reportStore.fetchReport()"
          class="bg-[#18181B] border border-[#26262A] rounded px-2 py-1 text-xs text-gray-200 font-mono focus:outline-none focus:border-[#7B96F5]"
        />
        <span class="text-[10px] uppercase font-mono text-gray-400">To:</span>
        <input
          v-model="reportStore.endDate"
          type="date"
          @change="reportStore.fetchReport()"
          class="bg-[#18181B] border border-[#26262A] rounded px-2 py-1 text-xs text-gray-200 font-mono focus:outline-none focus:border-[#7B96F5]"
        />
      </div>

      <select
        v-model="reportStore.deviceTypeFilter"
        class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
      >
        <option value="All">All Device Types</option>
        <option value="Access Point">Access Point</option>
        <option value="Switch">Switch</option>
        <option value="Router">Router</option>
        <option value="SmartPower">SmartPower</option>
        <option value="CCTV">CCTV</option>
        <option value="NVR">NVR</option>
      </select>

      <select
        v-model="reportStore.locationFilter"
        class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
      >
        <option v-for="loc in reportStore.uniqueLocations" :key="loc" :value="loc">
          {{ loc === 'All' ? 'All Locations' : loc }}
        </option>
      </select>
    </div>

    <!-- Overview Stat Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-4 gap-4">
      <template v-if="reportStore.isLoading">
        <SkeletonCard v-for="i in 4" :key="i" />
      </template>
      <template v-else>
        <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-1">
          <p class="text-[10px] font-mono text-gray-400 uppercase tracking-wider">Avg SLA Uptime</p>
          <p class="text-2xl font-extrabold text-[#34D399] font-mono">{{ reportStore.avgSlaUptime.toFixed(2) }}%</p>
          <p class="text-[11px] text-gray-500 font-mono">Target: 99.50%</p>
        </div>

        <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-1">
          <p class="text-[10px] font-mono text-gray-400 uppercase tracking-wider">Total Outage Events</p>
          <p class="text-2xl font-extrabold text-amber-400 font-mono">{{ reportStore.totalOutageEvents }}</p>
          <p class="text-[11px] text-gray-500 font-mono">Avg MTTR: {{ reportStore.avgMttrMinutes.toFixed(0) }}m</p>
        </div>

        <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-1">
          <p class="text-[10px] font-mono text-gray-400 uppercase tracking-wider">Alert Delivery</p>
          <p class="text-2xl font-extrabold text-[#7B96F5] font-mono">{{ reportStore.alertDeliveryRate }}%</p>
          <p class="text-[11px] text-gray-500 font-mono">WA + Telegram</p>
        </div>

        <div class="bg-[#151517] border border-[#F5A65B]/30 rounded-xl p-5 space-y-1">
          <p class="text-[10px] font-mono text-amber-400/80 uppercase tracking-wider">Recurring Issues</p>
          <p class="text-2xl font-extrabold text-amber-400 font-mono">{{ reportStore.flapDevices.length }}</p>
          <p class="text-[11px] text-gray-500 font-mono">≥5 downs / 7 days</p>
        </div>
      </template>
    </div>
    <!-- Report Section Tabs Selector -->
    <div class="flex border-b border-[#26262A] gap-6 text-sm font-mono pb-0.5">
      <button
        @click="activeTab = 'downtime'"
        class="pb-3 border-b-2 font-semibold transition-all relative text-xs"
        :class="activeTab === 'downtime' ? 'border-[#7B96F5] text-white' : 'border-transparent text-gray-500 hover:text-gray-300'"
      >
        Downtime by Device
      </button>
      <button
        @click="activeTab = 'recurring'"
        class="pb-3 border-b-2 font-semibold transition-all relative text-xs"
        :class="activeTab === 'recurring' ? 'border-amber-500 text-white' : 'border-transparent text-gray-500 hover:text-gray-300'"
      >
        Recurring Issues
      </button>
      <button
        @click="activeTab = 'active_incidents'"
        class="pb-3 border-b-2 font-semibold transition-all relative text-xs"
        :class="activeTab === 'active_incidents' ? 'border-red-500 text-white' : 'border-transparent text-gray-500 hover:text-gray-300'"
      >
        Active Incidents
      </button>
    </div>

    <!-- Tab 1: Downtime by Device -->
    <div v-if="activeTab === 'downtime'" class="grid grid-cols-1 xl:grid-cols-3 gap-6 items-start">
      <!-- Downtime Table (2/3 width) -->
      <div class="xl:col-span-2 space-y-4">
        <SkeletonTable v-if="reportStore.isLoading" :rows="7" :cols="5" />
        <div v-else class="bg-[#151517] border border-[#26262A] rounded-xl overflow-hidden">
          <div class="px-5 py-3.5 border-b border-[#26262A] flex items-center justify-between">
            <h3 class="text-xs font-bold uppercase tracking-wider text-gray-300 font-mono">Downtime by Device</h3>
            <span class="text-[10px] text-gray-500 font-mono">Sorted by most-down first</span>
          </div>
          <table class="w-full text-left text-xs text-gray-300">
            <thead class="bg-[#18181B] font-mono text-[10px] uppercase text-gray-500 border-b border-[#26262A]">
              <tr>
                <th class="py-3 px-4">Device</th>
                <th class="py-3 px-4">Location</th>
                <th class="py-3 px-4 text-center">Down Count</th>
                <th class="py-3 px-4">Total Downtime</th>
                <th class="py-3 px-4">Last Down</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[#26262A]">
              <tr
                v-for="(row, idx) in paginatedDowntimeRows"
                :key="row.deviceId"
                class="hover:bg-[#18181B] transition-colors"
                :class="{ 'border-l-2 border-l-[#F5A65B]': row.downCount >= 5 }"
              >
                <td class="py-3 px-4">
                  <div class="flex items-center gap-2">
                    <span
                      class="text-[10px] font-mono font-bold w-5 h-5 rounded-full flex items-center justify-center"
                      :class="idx === 0 ? 'bg-[#F16565]/20 text-[#F16565]' : idx === 1 ? 'bg-amber-500/20 text-amber-400' : 'bg-[#26262A] text-gray-400'"
                    >
                      {{ (downtimePage - 1) * downtimePageSize + idx + 1 }}
                    </span>
                    <div>
                      <p class="font-bold text-white">{{ row.deviceName }}</p>
                      <p class="text-[10px] font-mono text-gray-500">{{ row.deviceType }}</p>
                    </div>
                  </div>
                </td>
                <td class="py-3 px-4 text-gray-400 max-w-[150px] truncate text-[11px]">{{ row.location }}</td>
                <td class="py-3 px-4 text-center">
                  <span
                    class="text-sm font-extrabold font-mono"
                    :class="row.downCount >= 6 ? 'text-[#F16565]' : row.downCount >= 3 ? 'text-amber-400' : 'text-gray-300'"
                  >
                    {{ row.downCount }}
                  </span>
                </td>
                <td class="py-3 px-4 font-mono font-semibold text-[#F16565]">
                  {{ reportStore.formatDowntime(row.totalDowntimeMinutes) }}
                </td>
                <td class="py-3 px-4 font-mono text-gray-400 text-[11px]">{{ row.lastDown }}</td>
              </tr>
              <tr v-if="reportStore.filteredRows.length === 0">
                <td colspan="5" class="py-10 text-center text-gray-600 text-xs">
                  No downtime events for selected filters
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <PaginationControl
          v-model:current-page="downtimePage"
          v-model:page-size="downtimePageSize"
          :total="reportStore.filteredRows.length"
        />
      </div>

      <!-- Bar Chart (1/3 width) -->
      <div class="xl:col-span-1">
        <template v-if="reportStore.isLoading">
          <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-3">
            <Skeleton height="0.75rem" width="60%" />
            <Skeleton v-for="i in 7" :key="i" height="1.25rem" :width="`${90 - i * 8}%`" />
          </div>
        </template>
        <DowntimeBarChart
          v-else
          :data="reportStore.filteredRows"
          :period="reportStore.period"
        />
      </div>
    </div>

    <!-- Tab 2: Recurring Issues -->
    <div v-if="activeTab === 'recurring'" class="space-y-4">
      <SkeletonTable v-if="reportStore.isLoading" :rows="2" :cols="5" />
      <template v-else>
        <RecurringIssuesTable :devices="paginatedFlapDevices" />
        <PaginationControl
          v-model:current-page="recurringPage"
          v-model:page-size="recurringPageSize"
          :total="reportStore.flapDevices.length"
        />
      </template>
    </div>

    <!-- Tab 3: Active Incidents -->
    <div v-if="activeTab === 'active_incidents'" class="space-y-4">
      <SkeletonTable v-if="reportStore.isLoading" :rows="4" :cols="8" />
      <div v-else class="bg-[#151517] border border-[#26262A] rounded-xl overflow-hidden">
        <div class="px-5 py-3.5 border-b border-[#26262A] flex items-center justify-between">
          <h3 class="text-xs font-bold uppercase tracking-wider text-gray-300 font-mono">Active Incidents Report Table</h3>
          <span class="text-[10px] text-gray-500 font-mono">Real-time incident ticket queue</span>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs text-gray-300">
            <thead class="bg-[#18181B] font-mono text-[10px] uppercase text-gray-500 border-b border-[#26262A]">
              <tr>
                <th class="py-3 px-4">Ticket ID</th>
                <th class="py-3 px-4">Device Name</th>
                <th class="py-3 px-4">Type</th>
                <th class="py-3 px-4">IP Address</th>
                <th class="py-3 px-4">Duration</th>
                <th class="py-3 px-4 text-center">Affected</th>
                <th class="py-3 px-4">Status</th>
                <th class="py-3 px-4 text-right">Action</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[#26262A]">
              <tr
                v-for="inc in paginatedIncidents"
                :key="inc.id"
                class="hover:bg-[#18181B] transition-colors"
              >
                <td class="py-3 px-4 font-mono font-bold text-[#7B96F5]">{{ inc.id }}</td>
                <td class="py-3 px-4">
                  <div>
                    <p class="font-bold text-white">{{ inc.deviceName }}</p>
                    <p class="text-[10px] text-gray-500 font-mono">{{ inc.location }}</p>
                  </div>
                </td>
                <td class="py-3 px-4 text-gray-400">{{ inc.deviceType }}</td>
                <td class="py-3 px-4 font-mono text-gray-400">{{ inc.deviceIp }}</td>
                <td class="py-3 px-4 font-mono font-semibold" :class="inc.status === 'ACTIVE' ? 'text-red-400' : 'text-gray-400'">
                  {{ formatLiveDuration(inc.startedAt, inc.status, inc.duration) }}
                </td>
                <td class="py-3 px-4 text-center font-mono">
                  <span class="px-2 py-0.5 rounded text-[10px] bg-red-500/10 text-red-400 border border-red-500/20 font-bold">
                    {{ inc.affectedDevicesCount }}
                  </span>
                </td>
                <td class="py-3 px-4">
                  <span
                    class="px-2 py-0.5 rounded text-[10px] font-mono font-semibold"
                    :class="inc.status === 'ACTIVE' ? 'bg-[#F16565]/10 text-[#F16565] border border-[#F16565]/20' : 'bg-[#3ECF8E]/10 text-[#3ECF8E] border border-[#3ECF8E]/20'"
                  >
                    {{ inc.status }}
                  </span>
                </td>
                <td class="py-3 px-4 text-right">
                  <router-link
                    :to="`/incidents/${inc.id}`"
                    class="px-2.5 py-1 rounded-lg bg-[#26262A] hover:bg-[#323236] text-xs font-semibold text-gray-200 transition-colors inline-block"
                  >
                    View Details
                  </router-link>
                </td>
              </tr>
              <tr v-if="incidentStore.incidents.length === 0">
                <td colspan="8" class="py-10 text-center text-gray-600 text-xs">
                  No active or recent incidents reported
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <PaginationControl
        v-model:current-page="incidentsPage"
        v-model:page-size="incidentsPageSize"
        :total="incidentStore.incidents.length"
      />
    </div>

    <PrintableSLAAudit
      v-if="isPrintRendered"
      :period="reportStore.period"
      :avg-sla-uptime="reportStore.avgSlaUptime"
      :total-outage-events="reportStore.totalOutageEvents"
      :avg-mttr-minutes="reportStore.avgMttrMinutes"
      :alert-delivery-rate="reportStore.alertDeliveryRate"
      :rows="reportStore.filteredRows"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from 'vue';
import { useReportStore } from '../stores/reportStore';
import { useIncidentStore } from '../stores/incidentStore';
import DowntimeBarChart from '../components/reports/DowntimeBarChart.vue';
import RecurringIssuesTable from '../components/reports/RecurringIssuesTable.vue';
import PrintableSLAAudit from '../components/reports/PrintableSLAAudit.vue';
import PaginationControl from '../components/common/PaginationControl.vue';
import SkeletonCard from '../components/common/SkeletonCard.vue';
import SkeletonTable from '../components/common/SkeletonTable.vue';
import Skeleton from '../components/common/Skeleton.vue';
import { FileText, Printer, Calendar } from 'lucide-vue-next';

const activeTab = ref<'downtime' | 'recurring' | 'active_incidents'>('downtime');

const downtimePage = ref(1);
const downtimePageSize = ref(10);

const recurringPage = ref(1);
const recurringPageSize = ref(10);

const incidentsPage = ref(1);
const incidentsPageSize = ref(10);

const reportStore = useReportStore();
const incidentStore = useIncidentStore();

const paginatedDowntimeRows = computed(() => {
  const start = (downtimePage.value - 1) * downtimePageSize.value;
  return reportStore.filteredRows.slice(start, start + downtimePageSize.value);
});

const paginatedFlapDevices = computed(() => {
  const start = (recurringPage.value - 1) * recurringPageSize.value;
  return reportStore.flapDevices.slice(start, start + recurringPageSize.value);
});

const paginatedIncidents = computed(() => {
  const start = (incidentsPage.value - 1) * incidentsPageSize.value;
  return incidentStore.incidents.slice(start, start + incidentsPageSize.value);
});

// Reset page numbers to 1 when filters or tabs change
watch(
  () => [
    activeTab.value,
    reportStore.period,
    reportStore.deviceTypeFilter,
    reportStore.locationFilter
  ],
  () => {
    downtimePage.value = 1;
    recurringPage.value = 1;
    incidentsPage.value = 1;
  }
);

const now = ref(Date.now());
let durationInterval: any = null;
const isPrintRendered = ref(false);

function setPeriod(p: 'daily' | 'weekly' | 'monthly' | 'custom') {
  reportStore.period = p;
  reportStore.fetchReport();
}

function formatLiveDuration(startedAtStr: string | undefined, status: string, durationStr?: string) {
  if (status === 'RESOLVED') {
    return durationStr || 'Resolved';
  }
  if (!startedAtStr) return durationStr || 'Ongoing';
  const start = new Date(startedAtStr);
  const diffMs = now.value - start.getTime();
  if (diffMs < 0) return '0m 0s';
  const mins = Math.floor(diffMs / 60000);
  const secs = Math.floor((diffMs % 60000) / 1000);
  return `${mins}m ${secs}s (ongoing)`;
}

async function handlePrintPDF() {
  reportStore.isLoading = true;
  try {
    await reportStore.fetchReport();
    isPrintRendered.value = true;
    await nextTick();
    await new Promise((resolve) => setTimeout(resolve, 50));
    window.print();
  } catch (err) {
    console.error('Failed to print SLA report:', err);
  } finally {
    reportStore.isLoading = false;
    isPrintRendered.value = false;
  }
}

onMounted(() => {
  reportStore.fetchReport();
  incidentStore.fetchIncidents();
  durationInterval = setInterval(() => {
    now.value = Date.now();
  }, 1000);
});

onUnmounted(() => {
  if (durationInterval) clearInterval(durationInterval);
});
</script>
