<template>
  <div class="bg-[#151517] border border-[#F5A65B]/30 rounded-xl overflow-hidden">
    <!-- Header -->
    <div class="flex items-center gap-3 px-5 py-3.5 border-b border-[#26262A] bg-amber-500/5">
      <div class="w-8 h-8 rounded-lg bg-amber-500/15 border border-amber-500/30 flex items-center justify-center shrink-0">
        <AlertTriangle class="w-4 h-4 text-amber-400" />
      </div>
      <div>
        <p class="text-xs font-bold text-amber-400">Recurring Issues — Needs Technician</p>
        <p class="text-[10px] text-gray-400 mt-0.5">Devices with ≥5 downs in the last 7 days — possible hardware fault</p>
      </div>
    </div>

    <div v-if="isLoading" class="p-4">
      <SkeletonTable :rows="4" :cols="6" />
    </div>
    <table v-else-if="devices.length > 0" class="w-full text-xs text-gray-300">
      <thead class="bg-[#18181B] font-mono text-[10px] uppercase text-gray-500 border-b border-[#26262A]">
        <tr>
          <th class="py-2.5 px-4">Device</th>
          <th class="py-2.5 px-4">Location</th>
          <th class="py-2.5 px-4">IP</th>
          <th class="py-2.5 px-4">7-Day Downs</th>
          <th class="py-2.5 px-4">Total Downtime</th>
          <th class="py-2.5 px-4">Recommendation</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-[#26262A]">
        <tr
          v-for="d in devices"
          :key="d.deviceId"
          class="hover:bg-amber-500/5 transition-colors"
        >
          <td class="py-3 px-4">
            <p class="font-bold text-white">{{ d.deviceName }}</p>
            <p class="text-[10px] font-mono text-gray-500 mt-0.5">{{ d.deviceType }}</p>
          </td>
          <td class="py-3 px-4 text-gray-400 max-w-[160px] truncate">{{ d.location }}</td>
          <td class="py-3 px-4 font-mono text-gray-400 text-[11px]">{{ d.ip }}</td>
          <td class="py-3 px-4">
            <div class="flex items-center gap-2">
              <span class="text-base font-extrabold font-mono text-amber-400">{{ d.downCount7d }}×</span>
            </div>
          </td>
          <td class="py-3 px-4 font-mono text-[#F16565] font-semibold">{{ formatTime(d.totalDowntimeMinutes) }}</td>
          <td class="py-3 px-4">
            <span class="px-2 py-0.5 rounded-full text-[10px] font-mono bg-amber-500/15 text-amber-400 border border-amber-500/30">
              🔧 On-site Inspection
            </span>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-else class="py-12 px-6 text-center flex flex-col items-center justify-center space-y-3">
      <div class="w-10 h-10 rounded-full bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-check-circle"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
      </div>
      <div>
        <p class="text-sm font-bold text-gray-200">No Recurring Issues Detected</p>
        <p class="text-xs text-gray-500 mt-1 max-w-md mx-auto">All infrastructure nodes are operating stably. No devices have exceeded the flapping threshold of 5+ outages in the past 7 days.</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { FlapDevice } from '../../types';
import { AlertTriangle } from 'lucide-vue-next';
import SkeletonTable from '../common/SkeletonTable.vue';

defineProps<{
  devices: FlapDevice[];
  isLoading?: boolean;
}>();

function formatTime(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return h > 0 ? `${h}h ${m}m` : `${m}m`;
}
</script>
