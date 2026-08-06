<template>
  <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4 shadow-xl">
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
      <div class="flex items-center gap-2">
        <Video class="w-4 h-4 text-[#7B96F5]" />
        <h3 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono">Camera Live View</h3>
      </div>
      
      <!-- Mode Tabs -->
      <div v-if="device.snapshotUrl || device.streamUrl" class="flex bg-[#18181B] border border-[#26262A] rounded-lg p-0.5">
        <button
          v-if="device.snapshotUrl"
          @click="activeMode = 'snapshot'"
          class="px-2.5 py-1 rounded text-[10px] font-mono font-bold transition-colors"
          :class="activeMode === 'snapshot' ? 'bg-[#26262A] text-white' : 'text-gray-400 hover:text-gray-200'"
        >
          Snapshot
        </button>
        <button
          v-if="device.streamUrl"
          @click="activeMode = 'stream'"
          class="px-2.5 py-1 rounded text-[10px] font-mono font-bold transition-colors"
          :class="activeMode === 'stream' ? 'bg-[#26262A] text-white' : 'text-gray-400 hover:text-gray-200'"
        >
          Live Stream
        </button>
      </div>
    </div>

    <!-- Main View Area -->
    <div class="relative w-full aspect-video rounded-lg overflow-hidden bg-black border border-[#26262A] flex items-center justify-center">
      <!-- Status Badge Overlay -->
      <div v-if="device.snapshotUrl || device.streamUrl" class="absolute top-3 left-3 z-10 flex items-center gap-1.5 px-2 py-0.5 rounded bg-black/60 backdrop-blur-md border border-white/10 text-[9px] font-mono font-medium">
        <span class="w-1.5 h-1.5 rounded-full" :class="isOffline ? 'bg-red-500 animate-pulse' : 'bg-[#3ECF8E] animate-pulse'" />
        <span :class="isOffline ? 'text-red-400' : 'text-[#3ECF8E]'">{{ isOffline ? 'OFFLINE' : 'LIVE' }}</span>
      </div>

      <!-- No configuration placeholder -->
      <div v-if="!device.snapshotUrl && !device.streamUrl" class="text-center p-6 space-y-2.5">
        <VideoOff class="w-8 h-8 text-gray-600 mx-auto" />
        <div>
          <p class="text-xs font-bold text-gray-300">Live Feed Not Configured</p>
          <p class="text-[10px] text-gray-500 max-w-sm">Configure stream or snapshot URLs in device settings to enable CCTV live views.</p>
        </div>
      </div>

      <!-- Snapshot Mode -->
      <template v-else-if="activeMode === 'snapshot' && device.snapshotUrl">
        <img
          v-if="!isOffline"
          :src="snapshotSrc"
          @error="handleLoadError"
          @load="handleLoadSuccess"
          alt="CCTV Snapshot"
          class="w-full h-full object-cover transition-all duration-300"
        />
        <div v-else class="text-center p-6 space-y-2.5">
          <AlertCircle class="w-8 h-8 text-red-400 mx-auto animate-bounce" />
          <div>
            <p class="text-xs font-bold text-gray-300">Camera Feed Unreachable</p>
            <p class="text-[10px] text-gray-500">Failed to load snapshot from camera address. Retrying...</p>
          </div>
          <button @click="retryLoad" class="px-2.5 py-1 rounded bg-[#26262A] hover:bg-[#323236] text-[10px] font-semibold text-gray-200">
            Retry Connection
          </button>
        </div>
      </template>

      <!-- Live Stream Mode (go2rtc Bridge) -->
      <template v-else-if="activeMode === 'stream' && device.streamUrl">
        <iframe
          :src="bridgeUrl"
          class="w-full h-full border-0 bg-black"
          allow="autoplay; fullscreen"
          sandbox="allow-scripts allow-same-origin"
        ></iframe>
      </template>
    </div>

    <!-- Metadata & Details footer -->
    <div v-if="device.snapshotUrl || device.streamUrl" class="flex items-center justify-between text-[10px] font-mono text-gray-500">
      <div class="flex items-center gap-1.5">
        <Camera class="w-3.5 h-3.5 text-gray-400" />
        <span class="truncate max-w-[200px]">{{ activeMode === 'snapshot' ? device.snapshotUrl : device.streamUrl }}</span>
      </div>
      <span>{{ activeMode === 'snapshot' ? 'Refreshes every 3s' : 'Low Latency WebRTC' }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import { Video, VideoOff, Camera, AlertCircle } from 'lucide-vue-next';
import type { Device } from '../../types';

const props = defineProps<{
  device: Device;
}>();

const activeMode = ref<'snapshot' | 'stream'>('snapshot');
const isOffline = ref(false);
const cacheBuster = ref(Date.now());
let refreshInterval: any = null;

// Determine initial mode based on available configurations
onMounted(() => {
  if (!props.device.snapshotUrl && props.device.streamUrl) {
    activeMode.value = 'stream';
  }
  startRefresh();
});

onUnmounted(() => {
  stopRefresh();
});

// Watch device changes to update active modes
watch(
  () => props.device,
  (newDev) => {
    if (!newDev.snapshotUrl && newDev.streamUrl) {
      activeMode.value = 'stream';
    } else {
      activeMode.value = 'snapshot';
    }
    isOffline.value = false;
    cacheBuster.value = Date.now();
  }
);

// Dynamic source URLs
const snapshotSrc = computed(() => {
  if (!props.device.snapshotUrl) return '';
  const delimiter = props.device.snapshotUrl.includes('?') ? '&' : '?';
  return `${props.device.snapshotUrl}${delimiter}t=${cacheBuster.value}`;
});

const bridgeUrl = computed(() => {
  if (!props.device.streamUrl) return '';
  // Point to the go2rtc restreaming container (port 1984 standard API)
  const host = window.location.hostname;
  return `http://${host}:1984/stream.html?src=${encodeURIComponent(props.device.streamUrl)}&mode=webrtc,mse`;
});

// Load event handlers
function handleLoadError() {
  isOffline.value = true;
}

function handleLoadSuccess() {
  isOffline.value = false;
}

function retryLoad() {
  isOffline.value = false;
  cacheBuster.value = Date.now();
}

// Polling scheduler
function startRefresh() {
  stopRefresh();
  refreshInterval = setInterval(() => {
    if (activeMode.value === 'snapshot') {
      cacheBuster.value = Date.now();
    }
  }, 3000);
}

function stopRefresh() {
  if (refreshInterval) {
    clearInterval(refreshInterval);
    refreshInterval = null;
  }
}
</script>
