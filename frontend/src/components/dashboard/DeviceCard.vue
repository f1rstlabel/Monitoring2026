<template>
  <div
    @click="navigateToDetail"
    class="group relative rounded-xl border p-4 transition-all duration-200 cursor-pointer select-none"
    :class="[
      device.status === 'DOWN'
        ? 'bg-[#18181B]/90 border-[#F16565]/40 border-l-4 border-l-[#F16565] shadow-lg shadow-red-950/20'
        : 'bg-[#151517] border-[#26262A] hover:border-[#7B96F5]/50 hover:bg-[#18181B]'
    ]"
  >
    <div class="flex items-start justify-between">
      <div class="flex items-center gap-3">
        <div
          class="w-9 h-9 rounded-lg flex items-center justify-center shrink-0 transition-colors"
          :class="[
            device.status === 'DOWN'
              ? 'bg-[#F16565]/15 text-[#F16565] border border-[#F16565]/30'
              : 'bg-[#18181B] text-[#7B96F5] border border-[#26262A] group-hover:border-[#7B96F5]/30'
          ]"
        >
          <component :is="getIcon(device.type)" class="w-4 h-4" />
        </div>

        <div class="overflow-hidden">
          <h3 class="text-xs font-bold text-gray-100 truncate group-hover:text-[#7B96F5] transition-colors">
            {{ device.name }}
          </h3>
          <p class="text-[11px] font-mono text-gray-400 mt-0.5">{{ device.ip }}</p>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <span v-if="device.latencyMs !== undefined && device.status === 'UP'" class="px-2 py-0.5 rounded text-[10px] font-mono font-semibold bg-[#3ECF8E]/10 text-[#3ECF8E] border border-[#3ECF8E]/20">
          {{ device.latencyMs }} ms
        </span>
        <StatusPill :status="device.status" />
      </div>
    </div>

    <!-- Additional Metadata -->
    <div class="mt-4 pt-3 border-t border-[#26262A]/60 flex items-center justify-between text-[11px]">
      <span class="text-gray-500 font-mono text-[10px] uppercase truncate max-w-[150px]">{{ device.location }}</span>
      <span class="text-gray-400 font-mono text-[10px] flex items-center gap-1">
        <Clock class="w-3 h-3 text-gray-500" />
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
