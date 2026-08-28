<template>
  <div
    @click="navigateToDetail"
    class="group relative rounded-xl border p-4 transition-all duration-200 cursor-pointer select-none"
    :class="[
      device.status === 'DOWN'
        ? 'bg-card/90 border-status-down/40 border-l-4 border-l-[#F16565] shadow-lg shadow-red-950/20'
        : 'bg-surface border-subtle hover:border-brand-periwinkle/50 hover:bg-card'
    ]"
  >
    <div class="flex items-start justify-between">
      <div class="flex items-center gap-3">
        <div
          class="w-9 h-9 rounded-lg flex items-center justify-center shrink-0 transition-colors"
          :class="[
            device.status === 'DOWN'
              ? 'bg-status-down/15 text-status-down border border-status-down/30'
              : 'bg-card text-brand-periwinkle border border-subtle group-hover:border-brand-periwinkle/30'
          ]"
        >
          <component :is="getIcon(device.type)" class="w-4 h-4" />
        </div>

        <div class="overflow-hidden">
          <h3 class="text-xs font-bold text-text-main truncate group-hover:text-brand-periwinkle transition-colors">
            {{ device.name }}
          </h3>
          <p class="text-[11px] font-mono text-text-secondary mt-0.5">{{ device.ip }}</p>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <span v-if="device.latencyMs !== undefined && device.status === 'UP'" class="px-2 py-0.5 rounded text-[10px] font-mono font-semibold bg-status-up/10 text-status-up border border-status-up/20">
          {{ device.latencyMs }} ms
        </span>
        <StatusPill :status="device.status" />
      </div>
    </div>

    <!-- Additional Metadata -->
    <div class="mt-4 pt-3 border-t border-subtle/60 flex items-center justify-between text-[11px]">
      <span class="text-text-muted font-mono text-[10px] uppercase truncate max-w-[150px]">{{ device.location }}</span>
      <span class="text-text-secondary font-mono text-[10px] flex items-center gap-1">
        <Clock class="w-3 h-3 text-text-muted" />
        Checked {{ device.checkedSecondsAgo }}s ago
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import type { Device, DeviceType } from '../../types';
import StatusPill from '../common/StatusPill.vue';
import {
  Wifi,
  Network,
  Router as RouterIcon,
  Zap,
  Camera,
  HardDrive,
  Clock
} from 'lucide-vue-next';

const props = defineProps<{
  device: Device;
}>();

const router = useRouter();

function getIcon(type: DeviceType) {
  switch (type) {
    case 'Access Point': return Wifi;
    case 'Switch': return Network;
    case 'Router': return RouterIcon;
    case 'SmartPower': return Zap;
    case 'CCTV': return Camera;
    case 'NVR': return HardDrive;
    default: return Network;
  }
}

function navigateToDetail() {
  router.push(`/devices/${props.device.id}`);
}
</script>
