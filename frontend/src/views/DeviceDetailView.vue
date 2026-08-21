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
          <span v-if="device.snmpEnabled" class="flex items-center gap-1.5 bg-[#3ECF8E]/10 text-[#3ECF8E] px-2.5 py-1 rounded border border-[#3ECF8E]/20 font-bold">
            <Radio class="w-3.5 h-3.5 text-[#3ECF8E] animate-pulse" />
            SNMP v2c Active
          </span>
          <span class="flex items-center gap-1.5 text-gray-500">
            <MapPin class="w-3.5 h-3.5" />
            {{ device.location }} ({{ device.rack }})
          </span>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <button
          type="button"
          @click="isTerminalOpen = true"
          class="px-3.5 py-2 rounded-lg bg-[#18181B] border border-[#26262A] hover:border-[#3ECF8E] text-[#3ECF8E] font-medium text-xs transition-all flex items-center gap-2 cursor-pointer shadow-sm shadow-[#3ECF8E]/10"
        >
          <Terminal class="w-4 h-4" />
          Run Diagnostics
        </button>
        <button
          v-if="authStore.canEditDevice"
          @click="openEditDevice()"
          class="px-4 py-2 rounded-lg bg-[#18181B] border border-[#26262A] hover:bg-[#26262A] text-gray-200 font-medium text-xs transition-all flex items-center gap-2 cursor-pointer"
        >
          <Edit3 class="w-4 h-4 text-[#7B96F5]" />
          Edit Configuration
        </button>
      </div>
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
          <div v-if="device.snmpEnabled" class="flex justify-between py-2 items-center">
            <span class="text-gray-400 font-mono">SNMP Telemetry</span>
            <span class="font-mono text-[#3ECF8E] font-semibold flex items-center gap-1.5">
              <span class="w-1.5 h-1.5 rounded-full bg-[#3ECF8E] animate-ping"></span>
              Active (Port {{ device.snmpPort || 161 }})
            </span>
          </div>
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

    <!-- Dedicated SNMP Live Telemetry & Hardware Diagnostics Section -->
    <div v-if="device.snmpEnabled" class="bg-[#151517] border border-[#26262A] rounded-xl p-6 space-y-5 shadow-xl">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-[#26262A] pb-4">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-lg bg-[#3ECF8E]/10 border border-[#3ECF8E]/30 flex items-center justify-center text-[#3ECF8E]">
            <Radio class="w-4 h-4 animate-pulse" />
          </div>
          <div>
            <h3 class="text-sm font-bold text-white font-mono flex items-center gap-2">
              SNMP Live Telemetry &amp; Hardware Diagnostics
              <span class="text-[10px] px-2 py-0.5 rounded-full bg-[#3ECF8E]/10 border border-[#3ECF8E]/30 text-[#3ECF8E] font-semibold uppercase">
                Active v2c
              </span>
            </h3>
            <p class="text-xs text-gray-400 font-mono mt-0.5">Real-time MIB-2 telemetry and hardware operational parameters</p>
          </div>
        </div>

        <div class="flex items-center gap-2 text-xs font-mono">
          <span class="px-2.5 py-1 rounded-md bg-[#18181B] border border-[#26262A] text-gray-300">
            Port: <strong class="text-white">{{ device.snmpPort || 161 }}</strong>
          </span>
          <span class="px-2.5 py-1 rounded-md bg-[#18181B] border border-[#26262A] text-gray-300">
            Community: <strong class="text-white">{{ device.snmpCommunity || 'public' }}</strong>
          </span>
          <span v-if="device.snmpIfIndex" class="px-2.5 py-1 rounded-md bg-[#18181B] border border-[#26262A] text-gray-300">
            IfIndex: <strong class="text-sky-400">{{ device.snmpIfIndex }}</strong>
          </span>
        </div>
      </div>

      <!-- 4 Telemetry Metric Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- Hostname / System Name -->
        <div class="bg-[#18181B] border border-[#26262A] rounded-xl p-4 space-y-2">
          <div class="flex items-center justify-between text-gray-400 text-xs font-mono">
            <span class="uppercase tracking-wider">System Hostname</span>
            <Server class="w-4 h-4 text-amber-400" />
          </div>
          <p class="text-sm font-bold font-mono text-amber-400 truncate" :title="device.snmpSysName || 'Not Reported'">
            {{ device.snmpSysName || 'Not Reported' }}
          </p>
          <p class="text-[10px] font-mono text-gray-500">OID .1.3.6.1.2.1.1.5.0 (sysName)</p>
        </div>

        <!-- System Uptime (Since Last Reboot) -->
        <div class="bg-[#18181B] border border-[#26262A] rounded-xl p-4 space-y-2">
          <div class="flex items-center justify-between text-gray-400 text-xs font-mono">
            <span class="uppercase tracking-wider">Uptime (Since Reboot)</span>
            <Clock class="w-4 h-4 text-sky-400" />
          </div>
          <p class="text-sm font-bold font-mono text-sky-400 truncate">
            {{ device.snmpSysUpTime || 'N/A' }}
          </p>
          <p class="text-[10px] font-mono text-gray-500">OID .1.3.6.1.2.1.1.3.0 (sysUpTime)</p>
        </div>

        <!-- SNMP Physical Location -->
        <div class="bg-[#18181B] border border-[#26262A] rounded-xl p-4 space-y-2">
          <div class="flex items-center justify-between text-gray-400 text-xs font-mono">
            <span class="uppercase tracking-wider">Reported Location</span>
            <MapPin class="w-4 h-4 text-[#7B96F5]" />
          </div>
          <p class="text-sm font-bold font-mono text-gray-200 truncate" :title="device.snmpSysLocation || 'Not Configured'">
            {{ device.snmpSysLocation || 'Not Configured' }}
          </p>
          <p class="text-[10px] font-mono text-gray-500">OID .1.3.6.1.2.1.1.6.0 (sysLocation)</p>
        </div>

        <!-- SNMP Contact Admin -->
        <div class="bg-[#18181B] border border-[#26262A] rounded-xl p-4 space-y-2">
          <div class="flex items-center justify-between text-gray-400 text-xs font-mono">
            <span class="uppercase tracking-wider">Sys Contact / Admin</span>
            <UserCheck class="w-4 h-4 text-emerald-400" />
          </div>
          <p class="text-sm font-bold font-mono text-gray-200 truncate" :title="device.snmpSysContact || 'Not Configured'">
            {{ device.snmpSysContact || 'Not Configured' }}
          </p>
          <p class="text-[10px] font-mono text-gray-500">OID .1.3.6.1.2.1.1.4.0 (sysContact)</p>
        </div>
      </div>

      <!-- Full OS / Firmware Description (sysDescr) without truncation -->
      <div v-if="device.snmpSysDescr" class="space-y-2">
        <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold tracking-wider flex items-center gap-1.5">
          <Layers class="w-3.5 h-3.5 text-[#7B96F5]" />
          Full System &amp; Firmware Description (sysDescr)
        </label>
        <div class="bg-[#18181B] border border-[#26262A] rounded-xl p-3.5 font-mono text-xs text-gray-300 leading-relaxed break-words select-all shadow-inner">
          {{ device.snmpSysDescr }}
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

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs text-gray-300">
          <thead class="bg-[#18181B] font-mono text-[10px] uppercase text-gray-500">
            <tr>
              <th class="py-2.5 px-3">Date &amp; Time</th>
              <th class="py-2.5 px-3">Downtime Duration</th>
              <th class="py-2.5 px-3">Resolution Status</th>
              <th class="py-2.5 px-3 text-right">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#26262A]/60">
            <tr
              v-for="inc in paginatedIncidents"
              :key="inc.id"
              @click="$router.push(`/incidents/${inc.id}`)"
              class="hover:bg-[#18181B] cursor-pointer group transition-colors"
            >
              <td class="py-3 px-3 font-mono text-gray-300 group-hover:text-white font-medium flex items-center gap-2">
                <AlertTriangle class="w-3.5 h-3.5 text-amber-500 shrink-0" />
                <span>{{ inc.date }}</span>
              </td>
              <td class="py-3 px-3 font-mono text-red-400 font-semibold">{{ inc.duration }}</td>
              <td class="py-3 px-3">
                <span
                  class="px-2 py-0.5 rounded text-[10px] font-mono font-medium"
                  :class="inc.status === 'RESOLVED' ? 'bg-[#3ECF8E]/10 text-[#3ECF8E] border border-[#3ECF8E]/30' : 'bg-red-500/10 text-red-400 border border-red-500/30'"
                >
                  {{ inc.resolution }}
                </span>
              </td>
              <td class="py-3 px-3 text-right">
                <span class="text-xs font-mono text-[#7B96F5] group-hover:text-[#95ABF7] inline-flex items-center gap-1 font-semibold">
                  View Incident &rarr;
                </span>
              </td>
            </tr>
            <tr v-if="recentIncidents.length === 0">
              <td colspan="4" class="py-6 text-center text-gray-500 font-mono text-xs">No recent incidents recorded for this device</td>
            </tr>
          </tbody>
        </table>
      </div>

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
    @saved="deviceStore.fetchDevices()"
  />

  <!-- Diagnostic Terminal Modal -->
  <DiagnosticTerminalModal
    :is-open="isTerminalOpen"
    :initial-target="device?.ip || device?.mac || ''"
    @close="isTerminalOpen = false"
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
import DiagnosticTerminalModal from '../components/diagnostics/DiagnosticTerminalModal.vue';
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
  Zap,
  Radio,
  Server,
  Clock,
  UserCheck,
  Layers,
  Terminal
} from 'lucide-vue-next';

const route = useRoute();
const deviceStore = useDeviceStore();
const authStore = useAuthStore();

const isDetailLoading = ref(true);
const isFormModalOpen = ref(false);
const isTerminalOpen = ref(false);
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

interface IncidentItem {
  id: string;
  date: string;
  duration: string;
  status: string;
  resolution: string;
}

const recentIncidents = ref<IncidentItem[]>([]);
const incPage = ref(1);
const incPageSize = ref(5);
const paginatedIncidents = computed(() => {
  const start = (incPage.value - 1) * incPageSize.value;
  return recentIncidents.value.slice(start, start + incPageSize.value);
});

function formatIncidentDateTime(dateStr: string | undefined): string {
  if (!dateStr) return 'Recent';
  try {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) {
      return dateStr;
    }
    return (
      d.toLocaleString('id-ID', {
        day: '2-digit',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
      }) + ' WIB'
    );
  } catch (e) {
    return dateStr;
  }
}

async function fetchIncidentsForDevice() {
  if (!deviceId.value) return;
  try {
    const res = await api.get('/incidents', { params: { deviceId: deviceId.value } });
    const items = Array.isArray(res.data) ? res.data : (res.data?.items || res.data?.data || []);
    if (Array.isArray(items)) {
      recentIncidents.value = items.map((inc: any) => ({
        id: inc.id,
        date: formatIncidentDateTime(inc.startedAt || inc.startTime || inc.createdAt),
        duration: inc.duration || 'N/A',
        status: inc.status || 'RESOLVED',
        resolution:
          inc.status === 'RESOLVED'
            ? `Resolved (${formatIncidentDateTime(inc.resolvedAtRaw || inc.resolvedAt)})`
            : 'Active Outage (Ongoing)'
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