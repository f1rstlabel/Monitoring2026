<template>
  <div 
    class="rounded-xl border p-5 transition-all duration-200"
    :class="[
      isAlert 
        ? 'bg-red-950/20 border-red-500/40 text-red-200 shadow-lg shadow-red-500/5' 
        : 'bg-[#151517] border-[#26262A] hover:border-[#3A3A40]'
    ]"
  >
    <div class="flex items-start justify-between">
      <div>
        <p class="text-[11px] font-mono tracking-wider uppercase text-gray-400 font-medium">{{ title }}</p>
        <div class="flex items-baseline gap-2 mt-2">
          <h2 
            class="text-3xl font-extrabold tracking-tight"
            :class="[
              isAlert ? 'text-red-400' : 'text-white'
            ]"
          >
            {{ value }}
          </h2>
          <span 
            v-if="change"
            class="text-xs font-semibold px-1.5 py-0.5 rounded"
            :class="[
              changeType === 'increase-good' ? 'bg-[#3ECF8E]/15 text-[#3ECF8E]' :
              changeType === 'increase-bad' ? 'bg-red-500/15 text-red-400 border border-red-500/20' :
              'bg-amber-500/15 text-amber-400'
            ]"
          >
            {{ change }}
          </span>
        </div>
        <p v-if="subtitle" class="text-xs text-gray-500 mt-1">{{ subtitle }}</p>
      </div>

      <div 
        class="w-10 h-10 rounded-lg flex items-center justify-center shrink-0"
        :class="[
          isAlert 
            ? 'bg-red-500/15 text-red-400 border border-red-500/30' 
            : 'bg-[#18181B] border border-[#26262A] text-[#7B96F5]'
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
}>();
</script>
