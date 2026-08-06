<template>
  <div class="bg-[#151517] border border-[#26262A] rounded-xl flex flex-col h-full overflow-hidden">
    <!-- Header -->
    <div class="px-5 py-4 border-b border-[#26262A] flex items-center justify-between bg-[#18181B]">
      <div class="flex items-center gap-2">
        <Activity class="w-4 h-4 text-[#7B96F5]" />
        <h3 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono">Live Feed Events</h3>
      </div>
      <span class="w-2 h-2 rounded-full bg-[#3ECF8E] pulsing-dot-green"></span>
    </div>

    <!-- Feed Event List -->
    <div class="p-4 overflow-y-auto space-y-3 flex-1 max-h-[580px]">
      <div
        v-for="item in liveStore.liveFeed"
        :key="item.id"
        class="p-3 rounded-lg border text-xs transition-all duration-150 relative group"
        :class="[
          item.severity === 'critical'
            ? 'bg-red-950/20 border-red-500/30 text-red-200 hover:border-red-500/50'
            : item.severity === 'warning'
            ? 'bg-amber-950/20 border-amber-500/30 text-amber-200 hover:border-amber-500/50'
            : 'bg-[#18181B] border-[#26262A] text-gray-300 hover:border-[#3A3A40]'
        ]"
      >
        <div class="flex items-start justify-between gap-2">
          <div class="flex items-center gap-2">
            <span
              class="w-2 h-2 rounded-full shrink-0"
              :class="[
                item.severity === 'critical' ? 'bg-[#F16565] pulsing-dot-red' :
                item.severity === 'warning' ? 'bg-[#F5A65B]' : 'bg-[#3ECF8E]'
              ]"
            ></span>
            <h4 class="font-bold tracking-tight text-white">{{ item.title }}</h4>
          </div>
          <span class="text-[10px] font-mono text-gray-500 shrink-0">{{ item.timestamp }}</span>
        </div>

        <p class="mt-1 text-[11px] text-gray-400 pl-4 font-mono leading-relaxed">
          {{ item.description }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { Activity } from 'lucide-vue-next';
import { useLiveStore } from '../../stores/liveStore';

const liveStore = useLiveStore();

onMounted(() => {
  liveStore.initWebSocket();
});
</script>
