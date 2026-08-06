<template>
  <span
    class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold tracking-wide uppercase font-mono border"
    :class="[
      status === 'UP'
        ? 'bg-[#3ECF8E]/10 text-[#3ECF8E] border-[#3ECF8E]/30'
        : 'bg-[#F16565]/15 text-[#F16565] border-[#F16565]/40 animate-pulse'
    ]"
  >
    <span
      class="w-1.5 h-1.5 rounded-full"
      :class="status === 'UP' ? 'bg-[#3ECF8E]' : 'bg-[#F16565]'"
    ></span>
    {{ status }}
  </span>
</template>

<script setup lang="ts">
import type { DeviceStatus } from '../../types';

defineProps<{
  status: DeviceStatus;
}>();
</script>
