<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-subtle pb-4">
      <div>
        <h1 class="text-xl font-extrabold text-text-main tracking-tight">Availability & SLA Reports</h1>
        <p class="text-xs text-text-secondary mt-1">Network performance analysis, device uptime metrics, MTTR, and downtime distribution</p>
      </div>

      <div class="flex items-center gap-3">
        <!-- Period Toggle -->
        <div class="flex items-center bg-card border border-subtle rounded-lg p-0.5">
          <button
            @click="setPeriod('daily')"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-all cursor-pointer"
            :class="reportStore.period === 'daily' ? 'bg-subtle text-text-main shadow-sm' : 'text-text-secondary hover:text-text-main'"
          >
            Today
          </button>
          <button
            @click="setPeriod('weekly')"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-all cursor-pointer"
            :class="reportStore.period === 'weekly' ? 'bg-subtle text-text-main shadow-sm' : 'text-text-secondary hover:text-text-main'"
          >
            Last 7 Days
          </button>
          <button
            @click="setPeriod('monthly')"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-all cursor-pointer"
            :class="reportStore.period === 'monthly' ? 'bg-subtle text-text-main shadow-sm' : 'text-text-secondary hover:text-text-main'"
          >
            Last 30 Days
          </button>
          <button
            @click="setPeriod('custom')"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-all cursor-pointer"
            :class="reportStore.period === 'custom' ? 'bg-subtle text-text-main shadow-sm' : 'text-text-secondary hover:text-text-main'"
          >
            Custom Range
          </button>
        </div>

        <!-- Export Excel / CSV Dropdown -->
        <div v-if="authStore.hasPermission('reports.export')" class="relative">
          <button
            @click="showExportDropdown = !showExportDropdown"
            class="px-3.5 py-1.5 rounded-lg border border-subtle bg-surface hover:bg-hover text-emerald-400 font-semibold text-xs transition-all flex items-center gap-1.5 shadow-sm"
            title="Export as Excel / CSV Spreadsheet"
          >
            <FileSpreadsheet class="w-3.5 h-3.5 text-emerald-400" />
            <span>Export Data</span>
            <ChevronDown class="w-3 h-3 text-emerald-500/70 ml-0.5" />
          </button>

          <!-- Dropdown Options -->
          <div
            v-if="showExportDropdown"
            @click="showExportDropdown = false"
            class="absolute right-0 mt-1.5 w-52 bg-card border border-subtle rounded-xl shadow-2xl py-1.5 z-50 animate-in fade-in slide-in-from-top-1 duration-150"
          >
            <div class="px-3 py-1.5 border-b border-subtle/60 font-mono text-[10px] uppercase font-bold text-text-secondary">
              Select Export Format
            </div>
            <button
              @click="reportStore.exportData('xls', activeTab, incidentStore.incidents)"
              class="w-full text-left px-3.5 py-2 text-xs font-mono text-text-main hover:bg-emerald-500/10 hover:text-emerald-400 flex items-center gap-2.5 transition-colors"
            >
              <FileSpreadsheet class="w-4 h-4 text-emerald-400 shrink-0" />
              <div>
                <div class="font-bold">Excel Spreadsheet (.xls)</div>
                <div class="text-[10px] text-text-secondary font-sans">Structured Excel workbook</div>
              </div>
            </button>
            <button
              @click="reportStore.exportData('csv', activeTab, incidentStore.incidents)"
              class="w-full text-left px-3.5 py-2 text-xs font-mono text-text-main hover:bg-emerald-500/10 hover:text-emerald-400 flex items-center gap-2.5 transition-colors"
            >
              <FileText class="w-4 h-4 text-sky-400 shrink-0" />
              <div>
                <div class="font-bold">CSV File (.csv)</div>
                <div class="text-[10px] text-text-secondary font-sans">Standard UTF-8 CSV document</div>
              </div>
            </button>
          </div>
        </div>

        <!-- Export PDF -->
        <button
          v-if="authStore.hasPermission('reports.export')"
          @click="handlePrintPDF()"
          class="px-3.5 py-1.5 rounded-lg border border-subtle bg-surface hover:bg-hover text-sky-400 font-semibold text-xs transition-all flex items-center gap-1.5"
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
      <div v-if="reportStore.period === 'custom'" class="flex items-center gap-2 bg-surface border border-subtle rounded-lg px-3 py-1.5 text-xs text-text-secondary">
        <Calendar class="w-3.5 h-3.5 text-brand-periwinkle" />
        <span class="text-[10px] uppercase font-mono text-text-secondary">From:</span>
        <input
          v-model="reportStore.startDate"
          type="date"
          @change="reportStore.fetchReport()"
          class="bg-card border border-subtle rounded px-2 py-1 text-xs text-text-main font-mono focus:outline-none focus:border-brand-periwinkle"
        />
        <span class="text-[10px] uppercase font-mono text-text-secondary">To:</span>
        <input
          v-model="reportStore.endDate"
          type="date"
          @change="reportStore.fetchReport()"
          class="bg-card border border-subtle rounded px-2 py-1 text-xs text-text-main font-mono focus:outline-none focus:border-brand-periwinkle"
        />
      </div>

      <select
        v-model="reportStore.deviceTypeFilter"
        class="bg-card border border-subtle rounded-lg px-3 py-2 text-xs text-text-main focus:outline-none focus:border-brand-periwinkle font-mono"
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
        class="bg-card border border-subtle rounded-lg px-3 py-2 text-xs text-text-main focus:outline-none focus:border-brand-periwinkle font-mono"
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
        <div class="bg-surface border border-subtle rounded-xl p-5 space-y-1">
          <p class="text-[10px] font-mono text-text-secondary uppercase tracking-wider">Avg SLA Uptime</p>
          <p class="text-2xl font-extrabold text-status-up font-mono">{{ reportStore.filteredAvgSlaUptime.toFixed(2) }}%</p>
          <p class="text-[11px] text-text-muted font-mono">Target: 99.50%</p>
        </div>

        <div class="bg-surface border border-subtle rounded-xl p-5 space-y-1">
          <p class="text-[10px] font-mono text-text-secondary uppercase tracking-wider">Total Outage Events</p>
          <p class="text-2xl font-extrabold text-amber-400 font-mono">{{ reportStore.filteredTotalOutageEvents }}</p>
          <p class="text-[11px] text-text-muted font-mono">Avg MTTR: {{ reportStore.avgMttrMinutes.toFixed(0) }}m</p>
        </div>

        <div class="bg-surface border border-subtle rounded-xl p-5 space-y-1">
          <p class="text-[10px] font-mono text-text-secondary uppercase tracking-wider">Alert Delivery</p>
          <p class="text-2xl font-extrabold text-brand-periwinkle font-mono">{{ reportStore.alertDeliveryRate }}%</p>
          <p class="text-[11px] text-text-muted font-mono">WA + Telegram</p>
        </div>

        <div class="bg-surface border border-status-warning/30 rounded-xl p-5 space-y-1">
          <p class="text-[10px] font-mono text-amber-400/80 uppercase tracking-wider">Recurring Issues</p>
          <p class="text-2xl font-extrabold text-amber-400 font-mono">{{ reportStore.filteredFlapDevices.length }}</p>
          <p class="text-[11px] text-text-muted font-mono">≥5 downs / 7 days</p>
        </div>
      </template>
    </div>
    <!-- Report Section Tabs Selector -->
    <div class="flex border-b border-subtle gap-6 text-sm font-mono pb-0.5">
      <button
        @click="activeTab = 'downtime'"
        class="pb-3 border-b-2 font-semibold transition-all relative text-xs"
        :class="activeTab === 'downtime' ? 'border-brand-periwinkle text-text-main' : 'border-transparent text-text-muted hover:text-text-secondary'"
      >
        Downtime by Device
      </button>
      <button
        @click="activeTab = 'recurring'"
        class="pb-3 border-b-2 font-semibold transition-all relative text-xs"
        :class="activeTab === 'recurring' ? 'border-amber-500 text-text-main' : 'border-transparent text-text-muted hover:text-text-secondary'"
      >
        Recurring Issues
      </button>
      <button
        @click="activeTab = 'active_incidents'"
        class="pb-3 border-b-2 font-semibold transition-all relative text-xs"
        :class="activeTab === 'active_incidents' ? 'border-red-500 text-text-main' : 'border-transparent text-text-muted hover:text-text-secondary'"
      >
        Active Incidents
      </button>
    </div>

    <!-- Tab 1: Downtime by Device -->
    <div v-if="activeTab === 'downtime'" class="grid grid-cols-1 xl:grid-cols-3 gap-6 items-start">
      <!-- Downtime Table (2/3 width) -->
      <div class="xl:col-span-2 space-y-4">
        <SkeletonTable v-if="reportStore.isLoading" :rows="7" :cols="5" />
        <div v-else class="bg-surface border border-subtle rounded-xl overflow-hidden">
          <div class="px-5 py-3.5 border-b border-subtle flex items-center justify-between">
            <h3 class="text-xs font-bold uppercase tracking-wider text-text-secondary font-mono">Downtime by Device</h3>
            <span class="text-[10px] text-text-muted font-mono">Sorted by most-down first</span>
          </div>
          <table class="w-full text-left text-xs text-text-secondary">
            <thead class="bg-card font-mono text-[10px] uppercase text-text-muted border-b border-subtle">
              <tr>
                <th class="py-3 px-4">Device</th>
                <th class="py-3 px-4">Location</th>
                <th class="py-3 px-4 text-center">Down Count</th>
                <th class="py-3 px-4">Total Downtime</th>
                <th class="py-3 px-4">Last Down</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-subtle">
              <tr v-if="reportStore.isLoading">
                <td colspan="5" class="p-0 border-0">
                  <SkeletonTable :rows="5" :cols="5" />
                </td>
              </tr>
              <template v-else-if="paginatedDowntimeRows.length > 0">
                <tr
                  v-for="(row, idx) in paginatedDowntimeRows"
                  :key="row.deviceId"
                class="hover:bg-card transition-colors"
                :class="{ 'border-l-2 border-l-[#F5A65B]': row.downCount >= 5 }"
              >
                <td class="py-3 px-4">
                  <div class="flex items-center gap-2">
                    <span
                      class="text-[10px] font-mono font-bold w-5 h-5 rounded-full flex items-center justify-center"
                      :class="idx === 0 ? 'bg-status-down/20 text-status-down' : idx === 1 ? 'bg-amber-500/20 text-amber-400' : 'bg-subtle text-text-secondary'"
                    >
                      {{ (downtimePage - 1) * downtimePageSize + idx + 1 }}
                    </span>
                    <div>
                      <p class="font-bold text-text-main">{{ row.deviceName }}</p>
                      <p class="text-[10px] font-mono text-text-muted">{{ row.deviceType }}</p>
                    </div>
                  </div>
                </td>
                <td class="py-3 px-4 text-text-secondary max-w-[150px] truncate text-[11px]">{{ row.location }}</td>
                <td class="py-3 px-4 text-center">
                  <span
                    class="text-sm font-extrabold font-mono"
                    :class="row.downCount >= 6 ? 'text-status-down' : row.downCount >= 3 ? 'text-amber-400' : 'text-text-secondary'"
                  >
                    {{ row.downCount }}
                  </span>
                </td>
                <td class="py-3 px-4 font-mono font-semibold text-status-down">
                  {{ reportStore.formatDowntime(row.totalDowntimeMinutes) }}
                </td>
                <td class="py-3 px-4 font-mono text-text-secondary text-[11px]">{{ row.lastDown }}</td>
              </tr>
              </template>
              <tr v-else-if="reportStore.filteredRows.length === 0">
                <td colspan="5" class="py-10 text-center text-text-muted text-xs">
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
          <div class="bg-surface border border-subtle rounded-xl p-5 space-y-3">
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
    <div v-if="activeTab === 'recurring'" class="grid grid-cols-1 xl:grid-cols-3 gap-6 items-start">
      <!-- Recurring Table (2/3 width) -->
      <div class="xl:col-span-2 space-y-4">
        <SkeletonTable v-if="reportStore.isLoading" :rows="2" :cols="5" />
        <template v-else>
          <RecurringIssuesTable :devices="paginatedFlapDevices" />
          <PaginationControl
            v-model:current-page="recurringPage"
            v-model:page-size="recurringPageSize"
            :total="reportStore.filteredFlapDevices.length"
          />
        </template>
      </div>

      <!-- Recurring Flap Chart (1/3 width) -->
      <div class="xl:col-span-1">
        <template v-if="reportStore.isLoading">
          <div class="bg-surface border border-subtle rounded-xl p-5 space-y-3">
            <Skeleton height="0.75rem" width="60%" />
            <Skeleton v-for="i in 5" :key="i" height="1.25rem" :width="`${85 - i * 10}%`" />
          </div>
        </template>
        <RecurringBarChart
          v-else
          :devices="reportStore.filteredFlapDevices"
        />
      </div>
    </div>

    <!-- Tab 3: Active Incidents -->
    <div v-if="activeTab === 'active_incidents'" class="grid grid-cols-1 xl:grid-cols-3 gap-6 items-start">
      <!-- Active Incidents Table (2/3 width) -->
      <div class="xl:col-span-2 space-y-4">
        <SkeletonTable v-if="reportStore.isLoading" :rows="4" :cols="8" />
        <div v-else class="bg-surface border border-subtle rounded-xl overflow-hidden">
          <div class="px-5 py-3.5 border-b border-subtle flex items-center justify-between">
            <h3 class="text-xs font-bold uppercase tracking-wider text-text-secondary font-mono">Active Incidents Report Table</h3>
            <span class="text-[10px] text-text-muted font-mono">Real-time incident ticket queue</span>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs text-text-secondary">
              <thead class="bg-card font-mono text-[10px] uppercase text-text-muted border-b border-subtle">
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
              <tbody class="divide-y divide-subtle">
                <tr v-if="incidentStore.isLoading || reportStore.isLoading">
                  <td colspan="8" class="p-0 border-0">
                    <SkeletonTable :rows="5" :cols="8" />
                  </td>
                </tr>
                <template v-else-if="paginatedIncidents.length > 0">
                  <tr
                    v-for="inc in paginatedIncidents"
                    :key="inc.id"
                    class="hover:bg-card transition-colors"
                  >
                    <td class="py-3 px-4 font-mono font-bold text-brand-periwinkle">{{ inc.id }}</td>
                    <td class="py-3 px-4">
                      <div>
                        <p class="font-bold text-text-main">{{ inc.deviceName }}</p>
                        <p class="text-[10px] text-text-muted font-mono">{{ inc.location }}</p>
                      </div>
                    </td>
                    <td class="py-3 px-4 text-text-secondary">{{ inc.deviceType }}</td>
                    <td class="py-3 px-4 font-mono text-text-secondary">{{ inc.deviceIp }}</td>
                    <td class="py-3 px-4 font-mono font-semibold" :class="inc.status === 'ACTIVE' ? 'text-red-400' : 'text-text-secondary'">
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
                        :class="inc.status === 'ACTIVE' ? 'bg-status-down/10 text-status-down border border-status-down/20' : 'bg-status-up/10 text-status-up border border-status-up/20'"
                      >
                        {{ inc.status }}
                      </span>
                    </td>
                    <td class="py-3 px-4 text-right">
                      <router-link
                        :to="`/incidents/${inc.id}`"
                        class="px-2.5 py-1 rounded-lg bg-subtle hover:bg-hover text-xs font-semibold text-text-main transition-colors inline-block"
                      >
                        View Details
                      </router-link>
                    </td>
                  </tr>
                </template>
                <tr v-else-if="filteredIncidents.length === 0">
                  <td colspan="8" class="py-10 text-center text-text-muted text-xs">
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
          :total="filteredIncidents.length"
        />
      </div>

      <!-- Active Incidents Chart (1/3 width) -->
      <div class="xl:col-span-1">
        <template v-if="reportStore.isLoading">
          <div class="bg-surface border border-subtle rounded-xl p-5 space-y-3">
            <Skeleton height="0.75rem" width="60%" />
            <Skeleton v-for="i in 4" :key="i" height="1.25rem" :width="`${85 - i * 10}%`" />
          </div>
        </template>
        <ActiveIncidentsChart
          v-else
          :incidents="filteredIncidents"
        />
      </div>
    </div>

    <PrintableSLAAudit
      v-if="isPrintRendered"
      :period="reportStore.period"
      :active-tab="activeTab"
      :avg-sla-uptime="reportStore.filteredAvgSlaUptime"
      :total-outage-events="reportStore.filteredTotalOutageEvents"
      :avg-mttr-minutes="reportStore.avgMttrMinutes"
      :alert-delivery-rate="reportStore.alertDeliveryRate"
      :rows="reportStore.filteredRows"
      :flap-devices="reportStore.filteredFlapDevices"
      :incidents="filteredIncidents"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from 'vue';
import { useReportStore } from '../stores/reportStore';
import { useIncidentStore } from '../stores/incidentStore';
import { useAuthStore } from '../stores/authStore';
import DowntimeBarChart from '../components/reports/DowntimeBarChart.vue';
import RecurringBarChart from '../components/reports/RecurringBarChart.vue';
import ActiveIncidentsChart from '../components/reports/ActiveIncidentsChart.vue';
import RecurringIssuesTable from '../components/reports/RecurringIssuesTable.vue';
import PrintableSLAAudit from '../components/reports/PrintableSLAAudit.vue';
import PaginationControl from '../components/common/PaginationControl.vue';
import SkeletonCard from '../components/common/SkeletonCard.vue';
import SkeletonTable from '../components/common/SkeletonTable.vue';
import Skeleton from '../components/common/Skeleton.vue';
import { FileText, Printer, Calendar, ChevronDown, FileSpreadsheet } from 'lucide-vue-next';

const reportStore = useReportStore();
const incidentStore = useIncidentStore();
const authStore = useAuthStore();

const showExportDropdown = ref(false);
const activeTab = ref<'downtime' | 'recurring' | 'active_incidents'>('downtime');

const downtimePage = ref(1);
const downtimePageSize = ref(10);

const recurringPage = ref(1);
const recurringPageSize = ref(10);

const incidentsPage = ref(1);
const incidentsPageSize = ref(10);

const paginatedDowntimeRows = computed(() => {
  const start = (downtimePage.value - 1) * downtimePageSize.value;
  return reportStore.filteredRows.slice(start, start + downtimePageSize.value);
});

const paginatedFlapDevices = computed(() => {
  const start = (recurringPage.value - 1) * recurringPageSize.value;
  return reportStore.filteredFlapDevices.slice(start, start + recurringPageSize.value);
});

const filteredIncidents = computed(() => {
  let filtered = incidentStore.incidents;

  // Filter by date range based on reportStore.period
  const now = new Date().getTime();
  let fromTime = 0;
  if (reportStore.period === 'daily') fromTime = now - 24 * 3600 * 1000;
  else if (reportStore.period === 'weekly') fromTime = now - 7 * 24 * 3600 * 1000;
  else if (reportStore.period === 'monthly') fromTime = now - 30 * 24 * 3600 * 1000;
  else if (reportStore.period === 'custom' && reportStore.startDate) {
    fromTime = new Date(reportStore.startDate).getTime();
  }
  let toTime = now;
  if (reportStore.period === 'custom' && reportStore.endDate) {
    const d = new Date(reportStore.endDate);
    d.setHours(23, 59, 59, 999);
    toTime = d.getTime();
  }
  
  if (fromTime > 0) {
    filtered = filtered.filter(inc => {
      const incTime = new Date(inc.startedAt || inc.startTime).getTime();
      return incTime >= fromTime && incTime <= toTime;
    });
  }

  if (reportStore.locationFilter !== 'All') {
    filtered = filtered.filter(inc => inc.location && inc.location.includes(reportStore.locationFilter));
  }
  if (reportStore.deviceTypeFilter !== 'All') {
    filtered = filtered.filter(inc => inc.deviceType === reportStore.deviceTypeFilter);
  }
  // Always filter only ACTIVE incidents for this tab
  filtered = filtered.filter(inc => inc.status === 'ACTIVE');
  return filtered;
});

const paginatedIncidents = computed(() => {
  const start = (incidentsPage.value - 1) * incidentsPageSize.value;
  return filteredIncidents.value.slice(start, start + incidentsPageSize.value);
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
  // Refetch all incidents so we can filter them by date locally.
  // User requested to ONLY show active incidents in this tab.
  incidentStore.fetchIncidents({ status: 'ACTIVE' });
}

function formatLiveDuration(startedAtStr: string | undefined, status: string, durationStr?: string) {
  if (status === 'RESOLVED') {
    return durationStr || 'Resolved';
  }
  if (!startedAtStr) return durationStr || 'Ongoing';
  const start = new Date(startedAtStr);
  const diffMs = now.value - start.getTime();
  if (diffMs < 0) return '0s (ongoing)';
  
  const totalSecs = Math.floor(diffMs / 1000);
  const days = Math.floor(totalSecs / 86400);
  const hours = Math.floor((totalSecs % 86400) / 3600);
  const mins = Math.floor((totalSecs % 3600) / 60);
  const secs = totalSecs % 60;
  
  const parts = [];
  if (days > 0) parts.push(`${days}d`);
  if (hours > 0) parts.push(`${hours}h`);
  if (mins > 0) parts.push(`${mins}m`);
  if (secs > 0 || parts.length === 0) parts.push(`${secs}s`);
  
  return parts.join(' ') + ' (ongoing)';
}

async function handlePrintPDF() {
  reportStore.isLoading = true;
  const originalTitle = document.title;
  const dateStr = new Date().toISOString().slice(0, 10);
  const tabSlug = activeTab.value.replace(/_/g, '-');
  document.title = `sanoc-pdf-${tabSlug}-${reportStore.period}-${dateStr}`;

  try {
    await reportStore.fetchReport();
    if (activeTab.value === 'active_incidents') {
      await incidentStore.fetchIncidents();
    }
    isPrintRendered.value = true;
    await nextTick();
    await new Promise((resolve) => setTimeout(resolve, 50));
    window.print();
  } catch (err) {
    console.error('Failed to print SLA report:', err);
  } finally {
    reportStore.isLoading = false;
    isPrintRendered.value = false;
    document.title = originalTitle;
  }
}

onMounted(() => {
  reportStore.fetchReport();
  incidentStore.fetchIncidents({ status: 'ACTIVE' });
  durationInterval = setInterval(() => {
    now.value = Date.now();
  }, 1000);
});

onUnmounted(() => {
  if (durationInterval) clearInterval(durationInterval);
});
</script>
