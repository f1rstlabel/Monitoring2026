<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-[#26262A] pb-4">
      <div>
        <h1 class="text-xl font-extrabold text-white tracking-tight">Network Control Center & Monitoring</h1>
        <p class="text-xs text-gray-400 mt-1">Real-time IT infrastructure & network health telemetry</p>
      </div>

    </div>

    <!-- Top Stat Cards (4 in a row) — Skeleton during first load -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <template v-if="deviceStore.isLoading && deviceStore.devices.length === 0">
        <SkeletonCard v-for="i in 4" :key="i" />
      </template>
      <template v-else>
        <StatCard
          title="TOTAL NODES"
          :value="deviceStore.summary.totalDevices"
          :icon="Server"
          subtitle="Registered devices"
          :clickable="true"
          @click="navigateToDevices()"
        />

        <StatCard
          title="NODES OPERATIONAL (UP)"
          :value="deviceStore.summary.devicesUp"
          :icon="CheckCircle2"
          change="+96.1%"
          change-type="increase-good"
          :subtitle="`${deviceStore.summary.devicesUp} nodes online`"
          :clickable="true"
          @click="navigateToDevices('UP')"
        />

        <StatCard
          title="ACTIVE OUTAGES"
          :value="deviceStore.summary.devicesDown"
          :icon="XCircle"
          change="+3.9%"
          change-type="increase-bad"
          :subtitle="`${deviceStore.summary.devicesDown} nodes offline`"
          :is-alert="true"
          :clickable="true"
          @click="navigateToDevices('DOWN')"
        />

        <StatCard
          title="ACTIVE INCIDENTS"
          :value="deviceStore.summary.activeIncidents"
          :icon="AlertTriangle"
          change="Requiring action"
          change-type="warning"
          :subtitle="`${deviceStore.summary.activeIncidents} issues logged`"
          :clickable="true"
          @click="navigateToIncidents('ACTIVE')"
        />
      </template>
    </div>

    <!-- Main Content Area: Device Grid + Live Feed -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
      <!-- Left 2-Columns: Filter Toolbar & Device Grid/List -->
      <div class="lg:col-span-2 space-y-4">
        <!-- Toolbar & Filter Bar -->
        <div class="bg-[#151517] border border-[#26262A] rounded-xl p-3.5 flex flex-wrap items-center justify-between gap-3">
          <!-- Left Controls: Filter Search & Dropdowns -->
          <div class="flex flex-wrap items-center gap-2.5 flex-1 min-w-[280px]">
            <div class="relative flex-1 min-w-[180px]">
              <Search class="w-3.5 h-3.5 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
              <input
                v-model="deviceStore.searchQuery"
                type="text"
                placeholder="Search by name, IP, or location..."
                class="w-full bg-[#18181B] border border-[#26262A] rounded-lg pl-8 pr-3 py-1.5 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5]"
              />
            </div>

            <!-- Type Filter Dropdown -->
            <select
              v-model="deviceStore.selectedTypeFilter"
              class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-1.5 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
            >
              <option value="All">All Types</option>
              <option value="Access Point">Access Point</option>
              <option value="Switch">Switch</option>
              <option value="Router">Router</option>
              <option value="SmartPower">SmartPower</option>
              <option value="CCTV">CCTV</option>
              <option value="NVR">NVR</option>
            </select>

            <!-- Status Filter Dropdown -->
            <select
              v-model="deviceStore.selectedStatusFilter"
              class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-1.5 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
            >
              <option value="All">All Statuses</option>
              <option value="UP">UP Only</option>
              <option value="DOWN">DOWN Only</option>
            </select>
          </div>

          <!-- Right Controls: View Mode Toggle -->
          <div class="flex items-center gap-2">
            <div class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5">
              <button
                @click="deviceStore.viewMode = 'grid'"
                class="p-1.5 rounded text-xs transition-colors cursor-pointer"
                :class="deviceStore.viewMode === 'grid' ? 'bg-[#26262A] text-white' : 'text-gray-400 hover:text-gray-200'"
                title="Grid View"
              >
                <LayoutGrid class="w-3.5 h-3.5" />
              </button>
              <button
                @click="deviceStore.viewMode = 'list'"
                class="p-1.5 rounded text-xs transition-colors cursor-pointer"
                :class="deviceStore.viewMode === 'list' ? 'bg-[#26262A] text-white' : 'text-gray-400 hover:text-gray-200'"
                title="List View"
              >
                <List class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>

        <!-- Grid View -->
        <template v-if="deviceStore.viewMode === 'grid'">
          <!-- Skeleton grid while loading -->
          <div v-if="deviceStore.isLoading && deviceStore.devices.length === 0" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div v-for="i in 6" :key="i" class="bg-[#151517] border border-[#26262A] rounded-xl p-4 space-y-3">
              <div class="flex items-center justify-between">
                <Skeleton width="55%" height="0.85rem" />
                <Skeleton width="2.5rem" height="1.25rem" customClass="rounded-full" />
              </div>
              <Skeleton width="40%" height="0.65rem" />
              <Skeleton width="70%" height="0.65rem" />
              <div class="flex gap-2 pt-1">
                <Skeleton width="45%" height="0.65rem" />
                <Skeleton width="35%" height="0.65rem" />
              </div>
            </div>
          </div>

          <div v-else-if="paginatedDevices.length > 0" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div
              v-for="device in paginatedDevices"
              :key="device.id"
              @click="router.push(`/devices/${device.id}`)"
              class="bg-[#151517] border rounded-xl p-4 space-y-3 cursor-pointer hover:border-[#7B96F5] transition-all group"
              :class="[
                device.status === 'DOWN'
                  ? 'border-[#F16565]/50 bg-red-500/5 shadow-lg shadow-red-500/5'
                  : 'border-[#26262A]'
              ]"
            >
              <div class="flex items-center justify-between gap-2">
                <h4 class="font-bold text-xs text-white truncate group-hover:text-[#7B96F5] transition-colors">
                  {{ device.name }}
                </h4>
                <StatusPill :status="device.status" />
              </div>

              <div class="text-[11px] font-mono text-gray-400 space-y-1">
                <p class="flex items-center justify-between">
                  <span>IP:</span>
                  <span class="text-gray-200 font-semibold">{{ device.ip }}</span>
                </p>
                <p class="flex items-center justify-between">
                  <span>Type:</span>
                  <span class="text-gray-300">{{ device.type }}</span>
                </p>
                <p class="flex items-center justify-between truncate">
                  <span>Location:</span>
                  <span class="text-gray-400 truncate max-w-[140px]">{{ device.location }}</span>
                </p>
              </div>

              <div class="pt-2 border-t border-[#26262A] flex items-center justify-between text-[10px] font-mono text-gray-500">
                <span>Checked {{ device.checkedSecondsAgo }}s ago</span>
                <span class="text-[#7B96F5]">30d Uptime: {{ device.uptime30d }}%</span>
              </div>
            </div>
          </div>

          <!-- Empty state -->
          <div v-else class="p-12 text-center bg-[#151517] border border-[#26262A] rounded-xl text-gray-400 font-mono text-xs">
            No devices matching filter criteria
          </div>
        </template>

        <!-- List View -->
        <template v-else-if="deviceStore.viewMode === 'list'">
          <SkeletonTable v-if="deviceStore.isLoading && deviceStore.devices.length === 0" :rows="6" :cols="6" />
          <div v-else class="bg-[#151517] border border-[#26262A] rounded-xl overflow-hidden shadow-xl">
            <table class="w-full text-left text-xs text-gray-300">
              <thead class="bg-[#18181B] border-b border-[#26262A] font-mono text-[10px] uppercase text-gray-400">
                <tr>
                  <th class="py-3 px-4">Device Name</th>
                  <th class="py-3 px-4">Type</th>
                  <th class="py-3 px-4">IP Address</th>
                  <th class="py-3 px-4">Location</th>
                  <th class="py-3 px-4">Status</th>
                  <th class="py-3 px-4 text-right">Checked</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-[#26262A]">
                <template v-if="paginatedDevices.length > 0">
                  <tr
                    v-for="device in paginatedDevices"
                    :key="device.id"
                    @click="router.push(`/devices/${device.id}`)"
                    class="hover:bg-[#18181B] cursor-pointer transition-colors"
                    :class="{ 'bg-[#F16565]/5': device.status === 'DOWN' }"
                  >
                    <td class="py-3 px-4 font-semibold text-white">{{ device.name }}</td>
                    <td class="py-3 px-4 font-mono text-gray-400">{{ device.type }}</td>
                    <td class="py-3 px-4 font-mono text-gray-300">{{ device.ip }}</td>
                    <td class="py-3 px-4 text-gray-400">{{ device.location }}</td>
                    <td class="py-3 px-4"><StatusPill :status="device.status" /></td>
                    <td class="py-3 px-4 text-right font-mono text-gray-500">{{ device.checkedSecondsAgo }}s ago</td>
                  </tr>
                </template>
                <tr v-else>
                  <td colspan="6" class="py-12 text-center text-gray-500 font-mono text-xs">
                    No devices matching filter criteria
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>

        <!-- Pagination Control -->
        <div class="mt-4">
          <PaginationControl
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="deviceStore.filteredDevices.length"
          />
        </div>
      </div>

      <!-- Right 1-Column: Live Feed Panel -->
      <div class="lg:col-span-1">
        <LiveFeedPanel />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useDeviceStore } from '../stores/deviceStore';
import StatCard from '../components/common/StatCard.vue';
import SkeletonCard from '../components/common/SkeletonCard.vue';
import SkeletonTable from '../components/common/SkeletonTable.vue';
import Skeleton from '../components/common/Skeleton.vue';
import LiveFeedPanel from '../components/dashboard/LiveFeedPanel.vue';
import StatusPill from '../components/common/StatusPill.vue';
import PaginationControl from '../components/common/PaginationControl.vue';
import {
  Server,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Search,
  LayoutGrid,
  List
} from 'lucide-vue-next';

const router = useRouter();
const deviceStore = useDeviceStore();

const currentPage = ref(1);
const pageSize = ref(10);

const paginatedDevices = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  return deviceStore.filteredDevices.slice(start, start + pageSize.value);
});

function navigateToDevices(statusFilter?: string) {
  if (statusFilter) {
    deviceStore.selectedStatusFilter = statusFilter;
    router.push({ path: '/devices', query: { status: statusFilter } });
  } else {
    deviceStore.selectedStatusFilter = 'All';
    deviceStore.selectedTypeFilter = 'All';
    deviceStore.searchQuery = '';
    router.push('/devices');
  }
}

function navigateToIncidents(statusFilter?: string) {
  if (statusFilter) {
    router.push({ path: '/incidents', query: { status: statusFilter } });
  } else {
    router.push('/incidents');
  }
}

onMounted(() => {
  deviceStore.fetchDevices();
  deviceStore.fetchSummary();
});

watch(
  () => [deviceStore.searchQuery, deviceStore.selectedTypeFilter, deviceStore.selectedStatusFilter],
  () => {
    currentPage.value = 1;
  }
);
</script>
