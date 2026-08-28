<template>
  <div 
    class="rounded-xl border p-5 transition-all duration-150"
    :class="[
      isAlert 
        ? (clickable ? 'bg-status-down/10 border-status-down/40 text-status-down shadow-lg shadow-status-down/5 hover:border-status-down/70 hover:bg-status-down/20 cursor-pointer hover:scale-[1.01]' : 'bg-status-down/10 border-status-down/40 text-status-down shadow-lg shadow-status-down/5') 
        : (clickable ? 'bg-card border-subtle hover:border-brand-periwinkle/50 hover:bg-card hover:shadow-lg hover:shadow-brand-periwinkle/5 cursor-pointer hover:scale-[1.01] shadow-sm' : 'bg-card border-subtle shadow-sm')
    ]"
    @click="clickable && $emit('click')"
  >
    <div class="flex items-start justify-between">
      <div>
        <p class="text-[11px] font-mono tracking-wider uppercase text-text-secondary font-medium">{{ title }}</p>
        <div class="flex items-baseline gap-2 mt-2">
          <h2 
            class="text-3xl font-extrabold tracking-tight font-mono"
            :class="[
              isAlert ? 'text-status-down' : 'text-text-main'
            ]"
          >
            {{ value }}
          </h2>
          <span 
            v-if="change"
            class="text-xs font-semibold px-1.5 py-0.5 rounded"
            :class="[
              changeType === 'increase-good' ? 'bg-status-up/15 text-status-up' :
              changeType === 'increase-bad' ? 'bg-status-down/15 text-status-down border border-status-down/20' :
              'bg-amber-500/15 text-amber-500 dark:text-amber-400'
            ]"
          >
            {{ change }}
          </span>
        </div>
        <p v-if="subtitle" class="text-xs text-text-muted mt-1">{{ subtitle }}</p>
      </div>

      <div 
        class="w-10 h-10 rounded-lg flex items-center justify-center shrink-0"
        :class="[
          isAlert 
            ? 'bg-status-down/15 text-status-down border border-status-down/30' 
            : 'bg-card border border-subtle text-brand-periwinkle'
        ]"
      >
        <component :is="icon" class="w-5 h-5" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue';

defineProps<{
  title: string;
  value: string | number;
  icon: Component;
  change?: string;
  changeType?: 'increase-good' | 'increase-bad' | 'warning';
  subtitle?: string;
  isAlert?: boolean;
  clickable?: boolean;
}>();

defineEmits<{
  (e: 'click'): void;
}>();
</script>
