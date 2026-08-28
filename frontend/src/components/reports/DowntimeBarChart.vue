<template>
  <div class="bg-surface border border-subtle rounded-xl p-5 space-y-4">
    <div class="flex items-center justify-between">
      <h3 class="text-xs font-bold uppercase tracking-wider text-text-secondary font-mono">Down Count per Device</h3>
      <span class="text-[10px] font-mono text-text-muted">{{ period }} period</span>
    </div>

    <div v-if="data.length === 0" class="flex items-center justify-center h-32 text-text-muted text-xs">
      No data for selected period
    </div>

    <div v-else class="space-y-1.5">
      <div
        v-for="item in chartData"
        :key="item.deviceId"
        class="flex items-center gap-3 group"
      >
        <!-- Label -->
        <div class="w-40 shrink-0 text-right">
          <p class="text-[11px] text-text-secondary font-medium truncate group-hover:text-text-main transition-colors" :title="item.deviceName">
            {{ truncate(item.deviceName, 22) }}
          </p>
        </div>

        <!-- Bar + value -->
        <div class="flex-1 flex items-center gap-2">
          <div class="flex-1 h-5 bg-main rounded-md overflow-hidden border border-subtle">
            <div
              class="h-full rounded-md transition-all duration-700 ease-out relative"
              :style="{ width: item.pct + '%' }"
              :class="item.downCount >= 6 ? 'bg-gradient-to-r from-[#F16565] to-[#E04040]' : item.downCount >= 3 ? 'bg-gradient-to-r from-[#F5A65B] to-[#E08830]' : 'bg-gradient-to-r from-[#7B96F5] to-[#6070D0]'"
            >
              <!-- shimmer overlay on bar -->
              <div class="absolute inset-0 opacity-30 bg-gradient-to-r from-transparent via-white to-transparent bg-[length:200%_100%] animate-[slide_2s_ease-in-out_infinite]" />
            </div>
          </div>
          <span
            class="text-xs font-mono font-bold w-6 text-right"
            :class="item.downCount >= 6 ? 'text-status-down' : item.downCount >= 3 ? 'text-amber-400' : 'text-brand-periwinkle'"
          >
            {{ item.downCount }}
          </span>
        </div>
      </div>
    </div>

    <!-- Legend -->
    <div class="flex items-center gap-4 pt-2 border-t border-subtle">
      <div class="flex items-center gap-1.5">
        <div class="w-3 h-3 rounded-sm bg-gradient-to-r from-[#F16565] to-[#E04040]" />
        <span class="text-[10px] font-mono text-text-muted">≥6 (critical)</span>
      </div>
      <div class="flex items-center gap-1.5">
        <div class="w-3 h-3 rounded-sm bg-gradient-to-r from-[#F5A65B] to-[#E08830]" />
        <span class="text-[10px] font-mono text-text-muted">≥3 (warning)</span>
      </div>
      <div class="flex items-center gap-1.5">
        <div class="w-3 h-3 rounded-sm bg-gradient-to-r from-[#7B96F5] to-[#6070D0]" />
        <span class="text-[10px] font-mono text-text-muted">&lt;3 (normal)</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { ReportRow } from '../../types';

const props = defineProps<{
  data: ReportRow[];
  period: string;
}>();

const chartData = computed(() => {
  const maxDown = Math.max(...props.data.map(d => d.downCount), 1);
  return props.data.slice(0, 10).map(d => ({
    deviceId: d.deviceId,
    deviceName: d.deviceName,
    downCount: d.downCount,
    pct: Math.max((d.downCount / maxDown) * 100, 4)
  }));
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
