<template>
  <div class="bg-surface border border-subtle rounded-xl p-5 space-y-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <span class="w-2 h-2 rounded-full bg-status-down animate-pulse"></span>
        <h3 class="text-xs font-bold uppercase tracking-wider text-text-secondary font-mono">Incident Scope & Types</h3>
      </div>
      <span class="text-[10px] font-mono text-red-400 bg-red-500/10 px-2 py-0.5 rounded border border-red-500/20">Live Queue</span>
    </div>

    <div v-if="incidents.length === 0" class="flex flex-col items-center justify-center h-44 text-center space-y-2">
      <div class="w-8 h-8 rounded-full bg-status-up/10 border border-status-up/20 flex items-center justify-center text-status-up">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><polyline points="9 12 11 14 15 10"/></svg>
      </div>
      <p class="text-xs font-semibold text-text-secondary">Incident Queue Clear</p>
      <p class="text-[10px] text-text-muted max-w-[200px]">No active outages currently affecting regional infrastructure.</p>
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="item in typeBreakdown"
        :key="item.type"
        class="flex items-center gap-3 group"
      >
        <!-- Label -->
        <div class="w-32 shrink-0 text-right">
          <p class="text-[11px] text-text-secondary font-medium truncate group-hover:text-red-400 transition-colors">
            {{ item.type }}
          </p>
          <p class="text-[9px] font-mono text-text-muted">{{ item.pct.toFixed(0) }}% share</p>
        </div>

        <!-- Bar + value -->
        <div class="flex-1 flex items-center gap-2">
          <div class="flex-1 h-5 bg-main rounded-md overflow-hidden border border-subtle">
            <div
              class="h-full rounded-md transition-all duration-700 ease-out relative bg-gradient-to-r from-[#F16565] to-[#E04040]"
              :style="{ width: item.barPct + '%' }"
            >
              <!-- shimmer overlay on bar -->
              <div class="absolute inset-0 opacity-30 bg-gradient-to-r from-transparent via-white to-transparent bg-[length:200%_100%] animate-[slide_2s_ease-in-out_infinite]" />
            </div>
          </div>
          <span class="text-xs font-mono font-bold w-6 text-right text-red-400">
            {{ item.count }}
          </span>
        </div>
      </div>
    </div>

    <!-- Quick Stats Cards in Chart -->
    <div v-if="incidents.length > 0" class="grid grid-cols-2 gap-2 pt-2 border-t border-subtle">
      <div class="bg-card p-2.5 rounded-lg border border-subtle">
        <p class="text-[9px] font-mono text-text-secondary uppercase">Active Tickets</p>
        <p class="text-base font-extrabold text-red-400 font-mono mt-0.5">{{ activeCount }}</p>
      </div>
      <div class="bg-card p-2.5 rounded-lg border border-subtle">
        <p class="text-[9px] font-mono text-text-secondary uppercase">Affected Nodes</p>
        <p class="text-base font-extrabold text-amber-400 font-mono mt-0.5">{{ totalAffected }}</p>
      </div>
    </div>

    <!-- Context note -->
    <div class="pt-1 text-[10px] font-mono text-text-muted flex items-center justify-between border-t border-subtle/60">
      <span>Auto-updated live</span>
      <span class="text-text-secondary font-semibold">{{ incidents.length }} Total Recorded</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { Incident } from '../../types';

const props = defineProps<{
  incidents: Incident[];
}>();

const activeCount = computed(() => {
  return props.incidents.filter(i => i.status === 'ACTIVE').length;
});

const totalAffected = computed(() => {
  return props.incidents.reduce((acc, i) => acc + (i.affectedDevicesCount || 1), 0);
});

const typeBreakdown = computed(() => {
  const counts: Record<string, number> = {};
  props.incidents.forEach(i => {
    const t = i.deviceType || 'Other';
    counts[t] = (counts[t] || 0) + 1;
  });

  const total = props.incidents.length || 1;
  const maxCount = Math.max(...Object.values(counts), 1);

  return Object.entries(counts)
    .map(([type, count]) => ({
      type,
      count,
      pct: (count / total) * 100,
      barPct: Math.max((count / maxCount) * 100, 10)
    }))
    .sort((a, b) => b.count - a.count);
});
</script>

<style scoped>
@keyframes slide {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
