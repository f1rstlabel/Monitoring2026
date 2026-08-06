<template>
  <div class="flex flex-col items-center justify-center p-6 bg-[#18181B] rounded-xl border border-[#26262A]">
    <div class="relative w-40 h-40 flex items-center justify-center">
      <!-- SVG Radial Progress Gauge -->
      <svg class="w-full h-full transform -rotate-90" viewBox="0 0 100 100">
        <!-- Background Track -->
        <circle
          cx="50"
          cy="50"
          r="42"
          stroke="#26262A"
          stroke-width="8"
          fill="transparent"
        />
        <!-- Progress Bar -->
        <circle
          cx="50"
          cy="50"
          r="42"
          stroke="#3ECF8E"
          stroke-width="8"
          stroke-linecap="round"
          fill="transparent"
          :stroke-dasharray="263.89"
          :stroke-dashoffset="263.89 * (1 - (uptime / 100))"
          class="transition-all duration-1000 ease-out"
        />
      </svg>
      <div class="absolute inset-0 flex flex-col items-center justify-center text-center">
        <span class="text-2xl font-extrabold text-white font-mono tracking-tight">{{ uptime.toFixed(1) }}%</span>
        <span class="text-[10px] font-mono text-gray-400 uppercase tracking-widest mt-0.5">30D Uptime</span>
      </div>
    </div>
    <p class="text-xs text-gray-400 mt-4 text-center">
      Target SLA: <span class="font-mono text-[#3ECF8E] font-semibold">99.5%</span> (Exceeded by +0.3%)
    </p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  uptime: number;
}>();
</script>
