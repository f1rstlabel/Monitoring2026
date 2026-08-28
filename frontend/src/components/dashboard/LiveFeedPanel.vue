<template>
  <div class="bg-surface border border-subtle rounded-xl flex flex-col h-full overflow-hidden">
    <!-- Header -->
    <div class="px-5 py-4 border-b border-subtle flex items-center justify-between bg-card">
      <div class="flex items-center gap-2">
        <Activity class="w-4 h-4 text-brand-periwinkle" />
        <h3 class="text-xs font-bold uppercase tracking-wider text-text-main font-mono">Live Feed Events</h3>
      </div>
      <span class="w-2 h-2 rounded-full bg-status-up pulsing-dot-green"></span>
    </div>

    <!-- Feed Event List -->
    <div class="p-4 overflow-y-auto space-y-3 flex-1 max-h-[580px]">
      <template v-if="liveStore.liveFeed.length === 0">
        <div v-for="i in 5" :key="i" class="p-3 bg-card border border-subtle rounded-lg space-y-2">
          <div class="flex items-center justify-between">
            <Skeleton width="60%" height="0.8rem" />
            <Skeleton width="20%" height="0.65rem" />
          </div>
          <Skeleton width="85%" height="0.65rem" />
        </div>
      </template>
      <template v-else>
        <div
          v-for="item in liveStore.liveFeed"
          :key="item.id"
          class="p-3 rounded-lg border text-xs transition-all duration-150 relative group"
          :class="[
            item.severity === 'critical'
              ? 'bg-status-down/10 border-status-down/30 text-status-down hover:border-status-down/50'
              : item.severity === 'warning'
              ? 'bg-status-warning/10 border-status-warning/30 text-status-warning hover:border-status-warning/50'
              : 'bg-card border-subtle text-text-secondary hover:border-text-muted'
          ]"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="flex items-center gap-2">
              <span
                class="w-2 h-2 rounded-full shrink-0"
                :class="[
                  item.severity === 'critical' ? 'bg-status-down pulsing-dot-red' :
                  item.severity === 'warning' ? 'bg-status-warning' : 'bg-status-up'
                ]"
              ></span>
              <h4 class="font-bold tracking-tight text-text-main">{{ item.title }}</h4>
            </div>
            <span class="text-[10px] font-mono text-text-muted shrink-0">{{ item.timestamp }}</span>
          </div>

          <p class="mt-1 text-[11px] text-text-secondary pl-4 font-mono leading-relaxed">
            {{ item.description }}
          </p>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { Activity } from 'lucide-vue-next';
import { useLiveStore } from '../../stores/liveStore';
import Skeleton from '../common/Skeleton.vue';

const liveStore = useLiveStore();

onMounted(() => {
  liveStore.initWebSocket();
});
</script>
