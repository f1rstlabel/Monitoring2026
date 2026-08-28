<template>
  <span
    class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-[11px] font-semibold tracking-wide uppercase font-mono border"
    :class="[
      status === 'UP'
        ? 'bg-status-up/10 text-status-up border-status-up/30'
        : 'bg-status-down/15 text-status-down border-status-down/40 animate-pulse'
    ]"
  >
    <span
      class="w-1 h-1 rounded-full"
      :class="status === 'UP' ? 'bg-status-up' : 'bg-status-down'"
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
