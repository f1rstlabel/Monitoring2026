<template>
  <div class="bg-surface border border-subtle rounded-xl p-5 space-y-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <span class="w-2 h-2 rounded-full bg-amber-400"></span>
        <h3 class="text-xs font-bold uppercase tracking-wider text-text-secondary font-mono">Flapping Frequency</h3>
      </div>
      <span class="text-[10px] font-mono text-text-muted bg-card px-2 py-0.5 rounded border border-subtle">7-Day Window</span>
    </div>

    <div v-if="devices.length === 0" class="flex flex-col items-center justify-center h-44 text-center space-y-2">
      <div class="w-8 h-8 rounded-full bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
      </div>
      <p class="text-xs font-semibold text-text-secondary">Zero Flapping Devices</p>
      <p class="text-[10px] text-text-muted max-w-[200px]">All network devices are operating stably under threshold.</p>
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="item in chartData"
        :key="item.deviceId"
        class="flex items-center gap-3 group"
      >
        <!-- Label -->
        <div class="w-36 shrink-0 text-right">
          <p class="text-[11px] text-text-secondary font-medium truncate group-hover:text-amber-400 transition-colors" :title="item.deviceName">
            {{ truncate(item.deviceName, 18) }}
          </p>
          <p class="text-[9px] font-mono text-text-muted truncate">{{ item.deviceType }}</p>
        </div>

        <!-- Bar + value -->
        <div class="flex-1 flex items-center gap-2">
          <div class="flex-1 h-5 bg-main rounded-md overflow-hidden border border-subtle">
            <div
              class="h-full rounded-md transition-all duration-700 ease-out relative"
              :style="{ width: item.pct + '%' }"
              :class="item.downCount >= 10 ? 'bg-gradient-to-r from-[#F16565] to-[#E04040]' : 'bg-gradient-to-r from-[#F5A65B] to-[#E08830]'"
            >
              <!-- shimmer overlay on bar -->
              <div class="absolute inset-0 opacity-30 bg-gradient-to-r from-transparent via-white to-transparent bg-[length:200%_100%] animate-[slide_2s_ease-in-out_infinite]" />
            </div>
          </div>
          <span
            class="text-xs font-mono font-bold w-7 text-right"
            :class="item.downCount >= 10 ? 'text-status-down' : 'text-amber-400'"
          >
            {{ item.downCount }}×
          </span>
        </div>
      </div>
    </div>

    <!-- Quick Stats Cards in Chart -->
    <div v-if="devices.length > 0" class="grid grid-cols-2 gap-2 pt-2 border-t border-subtle">
      <div class="bg-card p-2.5 rounded-lg border border-subtle">
        <p class="text-[9px] font-mono text-text-secondary uppercase">Impacted Nodes</p>
        <p class="text-base font-extrabold text-amber-400 font-mono mt-0.5">{{ devices.length }}</p>
      </div>
      <div class="bg-card p-2.5 rounded-lg border border-subtle">
        <p class="text-[9px] font-mono text-text-secondary uppercase">Total Flaps (7d)</p>
        <p class="text-base font-extrabold text-status-down font-mono mt-0.5">{{ totalFlaps }}</p>
      </div>
    </div>

    <!-- Legend -->
    <div class="flex items-center gap-4 pt-1 text-[10px] font-mono text-text-muted">
      <div class="flex items-center gap-1.5">
        <div class="w-2.5 h-2.5 rounded-sm bg-gradient-to-r from-[#F16565] to-[#E04040]" />
        <span>≥10× High Risk</span>
      </div>
      <div class="flex items-center gap-1.5">
        <div class="w-2.5 h-2.5 rounded-sm bg-gradient-to-r from-[#F5A65B] to-[#E08830]" />
        <span>5-9× Moderate</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { FlapDevice } from '../../types';

const props = defineProps<{
  devices: FlapDevice[];
}>();

const chartData = computed(() => {
  const maxDown = Math.max(...props.devices.map(d => d.downCount7d), 1);
  return props.devices.slice(0, 8).map(d => ({
    deviceId: d.deviceId,
    deviceName: d.deviceName,
    deviceType: d.deviceType,
    downCount: d.downCount7d,
    pct: Math.max((d.downCount7d / maxDown) * 100, 8)
  }));
});

const totalFlaps = computed(() => {
  return props.devices.reduce((acc, d) => acc + d.downCount7d, 0);
});

function truncate(str: string, len: number): string {
  return str.length > len ? str.slice(0, len) + '…' : str;
}
</script>

<style scoped>
@keyframes slide {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
