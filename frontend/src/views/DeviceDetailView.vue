<template>
  <!-- Skeleton loading state while device data fetches -->
  <div v-if="isDetailLoading || deviceStore.isLoading" class="space-y-6">
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-6 space-y-3">
      <Skeleton width="40%" height="1.5rem" />
      <div class="flex gap-4">
        <Skeleton width="120px" height="1rem" />
        <Skeleton width="140px" height="1rem" />
        <Skeleton width="100px" height="1rem" />
      </div>
    </div>
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-1 bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4">
        <Skeleton width="50%" height="1rem" />
        <Skeleton width="120px" height="120px" customClass="rounded-full mx-auto" />
        <Skeleton v-for="i in 5" :key="i" width="100%" height="0.8rem" />
      </div>
      <div class="lg:col-span-2 space-y-6">
        <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4 h-72">
          <Skeleton width="40%" height="1rem" />
          <Skeleton width="100%" height="180px" />
        </div>
        <SkeletonTable :rows="3" :cols="5" />
      </div>
    </div>
  </div>

  <!-- Real Device Detail Content -->
  <div v-else-if="device" class="space-y-6">
    <!-- Breadcrumb -->
    <nav class="flex items-center gap-2 text-xs font-mono text-gray-400">
      <router-link to="/devices" class="hover:text-[#7B96F5] transition-colors">Devices</router-link>
      <ChevronRight class="w-3.5 h-3.5 text-gray-600" />
      <span class="text-gray-200 font-semibold">{{ device.name }}</span>
    </nav>

    <!-- Header Section -->
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-6 flex flex-wrap items-center justify-between gap-4 shadow-xl">
      <div class="space-y-2">
        <div class="flex items-center gap-3">
          <h1 class="text-xl font-extrabold text-white tracking-tight">{{ device.name }}</h1>
          <StatusPill :status="device.status" />
        </div>

        <div class="flex flex-wrap items-center gap-4 text-xs font-mono text-gray-400">
          <span class="flex items-center gap-1.5 bg-[#18181B] px-2.5 py-1 rounded border border-[#26262A]">
            <Network class="w-3.5 h-3.5 text-[#7B96F5]" />
            IP: {{ device.ip }}
          </span>
          <span class="flex items-center gap-1.5 bg-[#18181B] px-2.5 py-1 rounded border border-[#26262A]">
            <Cpu class="w-3.5 h-3.5 text-gray-400" />
            MAC: {{ device.mac }}
          </span>
          <span v-if="device.latencyMs !== undefined && device.status === 'UP'" class="flex items-center gap-1.5 bg-[#3ECF8E]/10 text-[#3ECF8E] px-2.5 py-1 rounded border border-[#3ECF8E]/20 font-bold">
            <Zap class="w-3.5 h-3.5 text-[#3ECF8E]" />
            ICMP Ping: {{ device.latencyMs }} ms
          </span>
          <span class="flex items-center gap-1.5 text-gray-500">
            <MapPin class="w-3.5 h-3.5" />
            {{ device.location }} ({{ device.rack }})
          </span>
        </div>
      </div>

      <button
        v-if="authStore.canEditDevice"
        @click="openEditDevice()"
        class="px-4 py-2 rounded-lg bg-[#18181B] border border-[#26262A] hover:bg-[#26262A] text-gray-200 font-medium text-xs transition-all flex items-center gap-2"
      >
        <Edit3 class="w-4 h-4 text-[#7B96F5]" />
        Edit Configuration
      </button>
    </div>

    <!-- Top Grid: Device Overview & 7-Day History Chart -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Left Card: Device Overview & Circular Gauge -->
      <div class="lg:col-span-1 bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-6 shadow-xl">
        <h3 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono border-b border-[#26262A] pb-3">
          Device Overview
        </h3>

        <!-- Circular Uptime Gauge -->
        <UptimeGauge :uptime="device.uptime30d" />

        <!-- Metadata Rows -->
        <div class="space-y-3 pt-2 text-xs divide-y divide-[#26262A]/60">
          <div class="flex justify-between py-2">
            <span class="text-gray-400 font-mono">Addressing Mode</span>
            <span class="font-mono text-white font-semibold">{{ device.addressingMode }}</span>
          </div>

          <div class="flex justify-between py-2">
            <span class="text-gray-400 font-mono">Location</span>
            <span class="text-white truncate max-w-[160px]">{{ device.location }}</span>
          </div>

          <div class="flex justify-between py-2">
            <span class="text-gray-400 font-mono">Failure Threshold</span>
            <span class="font-mono text-white font-semibold">{{ device.useCustomThreshold && device.customFailureThreshold ? device.customFailureThreshold : device.failureThreshold }} fails</span>
          </div>

          <div v-if="device.createdByUserName || device.createdByUserId" class="flex justify-between py-2 items-center">
            <span class="text-gray-400 font-mono">Added By</span>
            <span class="font-mono text-gray-200 font-semibold">{{ device.createdByUserName || device.createdByUserId }}</span>
          </div>
          <!-- New SNMP Info Section -->
          <template v-if="device.snmpEnabled">
            <div class="flex justify-between py-2 items-center">
              <span class="text-gray-400 font-mono">OS / System Name</span>
              <span class="font-mono text-amber-400 font-bold truncate max-w-[160px]" :title="device.snmpSysName || 'Not Available'">{{ device.snmpSysName || 'Not Available' }}</span>
            </div>
            <div v-if="device.snmpSysDescr" class="flex justify-between py-2 items-start gap-2">
              <span class="text-gray-400 font-mono flex-shrink-0">OS / Firmware</span>
              <span class="font-mono text-gray-300 text-[11px] text-right truncate max-w-[170px]" :title="device.snmpSysDescr">{{ device.snmpSysDescr }}</span>
            </div>
            <div v-if="device.snmpSysUpTime" class="flex justify-between py-2 items-center">
              <span class="text-gray-400 font-mono" title="Device's self-reported uptime since last reboot">System Uptime (Reboot)</span>
              <span class="font-mono text-sky-400 font-semibold text-[11px]">{{ device.snmpSysUpTime }}</span>
            </div>
            <div v-if="device.snmpSysContact" class="flex justify-between py-2 items-center">
              <span class="text-gray-400 font-mono">SNMP Contact</span>
              <span class="font-mono text-gray-300 text-[11px] truncate max-w-[160px]" :title="device.snmpSysContact">{{ device.snmpSysContact }}</span>
            </div>
            <div v-if="device.snmpSysLocation" class="flex justify-between py-2 items-center">
              <span class="text-gray-400 font-mono">SNMP Location</span>
              <span class="font-mono text-gray-300 text-[11px] truncate max-w-[160px]" :title="device.snmpSysLocation">{{ device.snmpSysLocation }}</span>
            </div>
            <div class="flex justify-between py-2">
              <span class="text-gray-400 font-mono">SNMP Polling</span>
              <span class="font-mono text-[#3ECF8E] font-semibold">Enabled</span>
            </div>
            <div class="flex justify-between py-2">
              <span class="text-gray-400 font-mono">SNMP Port</span>
              <span class="font-mono text-white">{{ device.snmpPort || 161 }}</span>
            </div>
            <div v-if="device.snmpIfIndex" class="flex justify-between py-2">
              <span class="text-gray-400 font-mono">SNMP IfIndex</span>
              <span class="font-mono text-white">{{ device.snmpIfIndex }}</span>
            </div>
          </template>
        </div>
      </div>

      <!-- Right Card: Status History Chart & Location Siblings -->
      <div class="lg:col-span-2 space-y-6">
        <StatusHistoryChart :device-id="device.id" />

        <!-- Devices in this Location -->
        <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4 shadow-xl">
          <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
            <h3 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono flex items-center gap-2">
              <MapPin class="w-4 h-4 text-[#7B96F5]" />
              Devices in this Location ({{ siblingDevices.length }})
            </h3>
          </div>

          <div v-if="siblingDevices.length > 0" class="space-y-3">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <router-link
                v-for="sib in paginatedSiblings"
                :key="sib.id"
                :to="'/devices/' + sib.id"
                class="bg-[#18181B] border border-[#26262A] hover:border-[#7B96F5]/50 rounded-lg p-3 flex items-center justify-between transition-colors group"
              >
                <div>
                  <h4 class="font-mono text-xs font-bold text-gray-200 group-hover:text-[#7B96F5] flex items-center gap-2">
                    {{ sib.name }}
                    <span class="text-[10px] text-gray-500 font-normal">({{ sib.type }})</span>
                  </h4>
                  <p class="text-[11px] font-mono text-gray-400 mt-0.5">IP: {{ sib.ip }}</p>
                </div>
                <StatusPill :status="sib.status" />
              </router-link>
            </div>
            <PaginationControl
              v-if="siblingDevices.length > sibPageSize"
              v-model:current-page="sibPage"
              v-model:page-size="sibPageSize"
              :total="siblingDevices.length"
            />
          </div>
          <div v-else class="p-4 text-center text-xs font-mono text-gray-500">
            No other devices registered at {{ device.location || 'this location' }}
          </div>
        </div>
      </div>
    </div>

    <!-- Bottom Section: Recent Incidents -->
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4 shadow-xl">
      <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
        <h3 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono flex items-center gap-2">
          <AlertTriangle class="w-4 h-4 text-amber-500" />
          Recent Incidents
        </h3>
        <router-link to="/incidents" class="text-xs text-[#7B96F5] hover:underline font-mono">
          View All &rarr;
        </router-link>
      </div>

      <table class="w-full text-left text-xs text-gray-300">
        <thead class="bg-[#18181B] font-mono text-[10px] uppercase text-gray-500">
          <tr>
            <th class="py-2 px-3">Date</th>
            <th class="py-2 px-3">Duration</th>
            <th class="py-2 px-3">Resolution</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#26262A]/60">
          <tr v-for="inc in paginatedIncidents" :key="inc.id" class="hover:bg-[#18181B]">
            <td class="py-2.5 px-3 font-mono text-gray-400">{{ inc.date }}</td>
            <td class="py-2.5 px-3 font-mono text-red-400 font-semibold">{{ inc.duration }}</td>
            <td class="py-2.5 px-3">
              <span class="px-2 py-0.5 rounded text-[10px] bg-[#3ECF8E]/10 text-[#3ECF8E] font-mono font-medium">
                {{ inc.resolution }}
              </span>
            </td>
          </tr>
          <tr v-if="recentIncidents.length === 0">
            <td colspan="3" class="py-4 text-center text-gray-500 font-mono text-xs">No recent incidents recorded for this device</td>
          </tr>
        </tbody>
      </table>

      <PaginationControl
        v-if="recentIncidents.length > incPageSize"
        v-model:current-page="incPage"
        v-model:page-size="incPageSize"
        :total="recentIncidents.length"
      />
    </div>
  </div>
  <!-- 404 Not Found State -->
  <div v-else class="p-8 text-center bg-[#151517] border border-[#26262A] rounded-xl space-y-4">
    <div class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-amber-500/10 text-amber-400">
      <AlertTriangle class="w-6 h-6" />
    </div>
    <h3 class="text-sm font-bold text-white">Device Information Not Found</h3>
    <p class="text-xs text-gray-400 max-w-md mx-auto">
      The requested device ID could not be found in the inventory database.
    </p>
    <router-link to="/devices" class="inline-block px-4 py-2 text-xs font-semibold rounded-lg bg-[#7B96F5] text-white hover:bg-[#95ABF7]">
      Back to Devices List
    </router-link>
  </div>

  <!-- Device Form Modal -->
  <DeviceFormModal
    :is-open="isFormModalOpen"
    :mode="formModalMode"
    :device="device"
    @close="isFormModalOpen = false"
    @saved="deviceStore.fetchDevices(); fetchIncidentsForDevice(); isFormModalOpen = false"
  />
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useDeviceStore } from '../stores/deviceStore';
import { useAuthStore } from '../stores/authStore';
import api from '../api/client';
import type { Device } from '../types';
import StatusPill from '../components/common/StatusPill.vue';
import UptimeGauge from '../components/devices/UptimeGauge.vue';
import StatusHistoryChart from '../components/devices/StatusHistoryChart.vue';
import DeviceFormModal from '../components/devices/DeviceFormModal.vue';
import PaginationControl from '../components/common/PaginationControl.vue';
import Skeleton from '../components/common/Skeleton.vue';
import SkeletonTable from '../components/common/SkeletonTable.vue';
import {
  ChevronRight,
  Network,
  Cpu,
  MapPin,
  Edit3,
  AlertTriangle,
  Zap
} from 'lucide-vue-next';

const route = useRoute();
const deviceStore = useDeviceStore();
const authStore = useAuthStore();

const isDetailLoading = ref(true);
const isFormModalOpen = ref(false);
const formModalMode = ref<'add' | 'edit'>('edit');

const deviceId = computed(() => route.params.id as string);
const device = computed(() => deviceStore.devices.find((d: Device) => d.id === deviceId.value) || null);

const siblingDevices = computed(() => {
  if (!device.value) return [];
  return deviceStore.devices.filter((d: Device) => {
    if (d.id === device.value?.id) return false;
    if (d.locationId && device.value?.locationId && d.locationId === device.value.locationId) return true;
    if (d.location && device.value?.location && d.location.toLowerCase() === device.value.location.toLowerCase()) return true;
    return false;
  });
});

const sibPage = ref(1);
const sibPageSize = ref(4);
const paginatedSiblings = computed(() => {
  const start = (sibPage.value - 1) * sibPageSize.value;
  return siblingDevices.value.slice(start, start + sibPageSize.value);
});

const recentIncidents = ref<{ id: string; date: string; duration: string; resolution: string }[]>([]);
const incPage = ref(1);
const incPageSize = ref(5);
const paginatedIncidents = computed(() => {
  const start = (incPage.value - 1) * incPageSize.value;
  return recentIncidents.value.slice(start, start + incPageSize.value);
});

async function fetchIncidentsForDevice() {
  if (!deviceId.value) return;
  try {
    const res = await api.get('/incidents', { params: { deviceId: deviceId.value } });
    const items = Array.isArray(res.data) ? res.data : (res.data?.items || res.data?.data || []);
    if (Array.isArray(items)) {
      recentIncidents.value = items.map((inc: any) => ({
        id: inc.id,
        date: `${inc.startTime || 'Recent'}`,
        duration: inc.duration || 'N/A',
        resolution: inc.status === 'RESOLVED' ? `Resolved (${inc.resolvedAt || 'Recovered'})` : 'Active Outage (Ongoing)'
      }));
    }
  } catch (e) {
    recentIncidents.value = [];
  }
}

function openEditDevice() {
  isFormModalOpen.value = true;
}

onMounted(async () => {
  isDetailLoading.value = true;
  try {
    if (deviceStore.devices.length === 0) {
      await deviceStore.fetchDevices();
    }
    await fetchIncidentsForDevice();
  } finally {
    isDetailLoading.value = false;
  }
});

watch(deviceId, async () => {
  isDetailLoading.value = true;
  try {
    await fetchIncidentsForDevice();
  } finally {
    isDetailLoading.value = false;
  }
});
</script>