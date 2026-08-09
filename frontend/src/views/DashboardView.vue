<template>
  <div class="space-y-6">
    <!-- Top Stat Cards (4 in a row) — Skeleton during first load -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <template v-if="deviceStore.isLoading && deviceStore.devices.length === 0">
        <SkeletonCard v-for="i in 4" :key="i" />
      </template>
      <template v-else>
        <StatCard
          title="TOTAL DEVICES"
          :value="deviceStore.summary.totalDevices"
          :icon="Server"
          subtitle="Monitored Infrastructure"
        />

        <StatCard
          title="DEVICES UP"
          :value="deviceStore.summary.devicesUp"
          :icon="CheckCircle2"
          change="+96.1%"
          change-type="increase-good"
          subtitle="Operational & Responsive"
        />

        <StatCard
          title="DEVICES DOWN"
          :value="deviceStore.summary.devicesDown"
          :icon="XCircle"
          change="+3.9%"
          change-type="increase-bad"
          subtitle="Requires Attention"
          :is-alert="true"
        />

        <StatCard
          title="ACTIVE INCIDENTS"
          :value="deviceStore.summary.activeIncidents"
          :icon="AlertTriangle"
          change="Requiring action"
          change-type="warning"
          subtitle="Open Ticket Queue"
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
                placeholder="Filter Name or IP..."
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

          <!-- Right Controls: Add Device & View Mode Toggle -->
          <div class="flex items-center gap-2">
            <!-- Add Device: hidden for pimpinan via v-if -->
            <button
              v-if="authStore.canAddDevice"
              @click="isAddModalOpen = true"
              class="px-3 py-1.5 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-medium text-xs shadow-sm shadow-[#7B96F5]/20 transition-all flex items-center gap-1.5"
            >
              <Plus class="w-4 h-4" />
              Add Device
            </button>

            <div class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5">
              <button
                @click="deviceStore.viewMode = 'grid'"
                class="p-1.5 rounded text-xs transition-colors"
                :class="deviceStore.viewMode === 'grid' ? 'bg-[#26262A] text-white' : 'text-gray-400 hover:text-gray-200'"
              >
                <LayoutGrid class="w-3.5 h-3.5" />
              </button>
              <button
                @click="deviceStore.viewMode = 'list'"
                class="p-1.5 rounded text-xs transition-colors"
                :class="deviceStore.viewMode === 'list' ? 'bg-[#26262A] text-white' : 'text-gray-400 hover:text-gray-200'"
              >
                <List class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>

        <!-- Grid View -->
        <template v-if="deviceStore.viewMode === 'grid'">
          <!-- Skeleton grid while loading -->
          <div v-if="deviceStore.isLoading" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
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
          <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div
              v-for="device in deviceStore.devices"
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
                  <span>IP Address:</span>
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
        </template>

        <!-- List View -->
        <div v-else class="bg-[#151517] border border-[#26262A] rounded-xl overflow-hidden">
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
              <tr
                v-for="device in deviceStore.devices"
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
            </tbody>
          </table>
        </div>

        <!-- Pagination Control -->
        <div class="mt-4">
          <PaginationControl
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="deviceStore.totalCount"
          />
        </div>
      </div>

      <!-- Right 1-Column: Live Feed Panel -->
      <div class="lg:col-span-1">
        <LiveFeedPanel />
      </div>
    </div>

    <!-- Add Device Modal -->
    <DeviceFormModal :is-open="isAddModalOpen" mode="add" @close="isAddModalOpen = false" @saved="deviceStore.fetchDevices(); deviceStore.fetchSummary();" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useDeviceStore } from '../stores/deviceStore';
import { useAuthStore } from '../stores/authStore';
import StatCard from '../components/common/StatCard.vue';
import SkeletonCard from '../components/common/SkeletonCard.vue';
import Skeleton from '../components/common/Skeleton.vue';
import LiveFeedPanel from '../components/dashboard/LiveFeedPanel.vue';
import StatusPill from '../components/common/StatusPill.vue';
import DeviceFormModal from '../components/devices/DeviceFormModal.vue';
import PaginationControl from '../components/common/PaginationControl.vue';
import {
  Server,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Search,
  Plus,
  LayoutGrid,
  List
} from 'lucide-vue-next';

const router = useRouter();
const deviceStore = useDeviceStore();
const authStore = useAuthStore();
const isAddModalOpen = ref(false);

const currentPage = ref(1);
const pageSize = ref(10);

function loadDevices() {
  deviceStore.fetchDevices({
    page: currentPage.value,
    pageSize: pageSize.value
  });
}

onMounted(() => {
  loadDevices();
  deviceStore.fetchSummary();
});

watch([currentPage, pageSize], () => {
  loadDevices();
});

let searchTimeout: any = null;
watch(
  () => deviceStore.searchQuery,
  () => {
    if (searchTimeout) clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      currentPage.value = 1;
      loadDevices();
    }, 300);
  }
);

watch(
  () => [deviceStore.selectedTypeFilter, deviceStore.selectedStatusFilter],
  () => {
    currentPage.value = 1;
    loadDevices();
  }
);
</script>
