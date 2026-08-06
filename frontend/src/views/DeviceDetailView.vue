<template>
  <div v-if="device" class="space-y-6">
    <!-- Breadcrumb -->
    <nav class="flex items-center gap-2 text-xs font-mono text-gray-400">
      <router-link to="/devices" class="hover:text-[#7B96F5] transition-colors">Devices</router-link>
      <ChevronRight class="w-3.5 h-3.5 text-gray-600" />
      <span class="text-gray-200 font-semibold">{{ device.name }}</span>
    </nav>

    <!-- Header Section -->
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-6 flex flex-wrap items-center justify-between gap-4">
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
      <div class="lg:col-span-1 bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-6">
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
          <!-- New SNMP Info Section -->
          <template v-if="device.snmpEnabled">
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

      <!-- Right Card: Status History Chart & SNMP Metrics -->
      <div class="lg:col-span-2 space-y-6">
        <StatusHistoryChart :device-id="device.id" />
        <CCTVPictureCard v-if="device.type === 'CCTV' || device.type === 'NVR'" :device="device" />
      </div>
    </div>

    <!-- Bottom Section: Recent Incidents -->
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4">
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
          <tr v-for="inc in recentIncidents" :key="inc.id" class="hover:bg-[#18181B]">
            <td class="py-2.5 px-3 font-mono text-gray-400">{{ inc.date }}</td>
            <td class="py-2.5 px-3 font-mono text-red-400 font-semibold">{{ inc.duration }}</td>
            <td class="py-2.5 px-3">
              <span class="px-2 py-0.5 rounded text-[10px] bg-[#3ECF8E]/10 text-[#3ECF8E] font-mono font-medium">
                {{ inc.resolution }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
  <div v-else class="p-8 text-center bg-[#151517] border border-[#26262A] rounded-xl space-y-4">
    <div class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-amber-500/10 text-amber-400">
      <AlertTriangle class="w-6 h-6" />
    </div>
    <h3 class="text-sm font-bold text-white">Device Information Loading or Not Found</h3>
    <p class="text-xs text-gray-400 max-w-md mx-auto">
      The requested device ID could not be loaded from the inventory or is currently fetching.
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
import CCTVPictureCard from '../components/devices/CCTVPictureCard.vue';
import DeviceFormModal from '../components/devices/DeviceFormModal.vue';
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

const isFormModalOpen = ref(false);
const formModalMode = ref<'add' | 'edit'>('edit');

const deviceId = computed(() => route.params.id as string);
const device = computed(() => deviceStore.devices.find((d: Device) => d.id === deviceId.value) || null);

const recentIncidents = ref<{ id: string; date: string; duration: string; resolution: string }[]>([]);

async function fetchIncidentsForDevice() {
  if (!deviceId.value) return;
  try {
    const res = await api.get('/incidents', { params: { deviceId: deviceId.value } });
    if (Array.isArray(res.data)) {
      recentIncidents.value = res.data.map((inc: any) => ({
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
  if (deviceStore.devices.length === 0) {
    await deviceStore.fetchDevices();
  }
  fetchIncidentsForDevice();
});

watch(deviceId, () => {
  fetchIncidentsForDevice();
});
</script>