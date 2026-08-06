<template>
  <Modal :is-open="isOpen" title="WhatsApp Session QR Connect" @close="$emit('close')">
    <template #icon>
      <QrCode class="w-5 h-5 text-[#34D399]" />
    </template>

    <div class="flex flex-col items-center justify-center p-4 space-y-5 text-center">
      <!-- Status Pill -->
      <div class="flex items-center gap-2 px-3 py-1 rounded-full text-xs font-mono font-semibold border" :class="statusBadgeClass">
        <span class="w-2 h-2 rounded-full inline-block" :class="statusDotClass"></span>
        <span>{{ statusText }}</span>
      </div>

      <!-- QR Code Container -->
      <div class="relative bg-white p-4 rounded-xl border border-[#26262A] shadow-xl w-56 h-56 flex items-center justify-center overflow-hidden">
        <!-- Connected Overlay -->
        <div v-if="status === 'connected'" class="absolute inset-0 bg-[#151517]/95 backdrop-blur rounded-xl flex flex-col items-center justify-center p-4 space-y-2 z-10">
          <CheckCircle2 class="w-12 h-12 text-[#34D399] animate-bounce" />
          <p class="text-xs font-bold text-white">Session Connected!</p>
          <p class="text-[10px] font-mono text-[#34D399]">{{ linkedNumber }}</p>
        </div>

        <!-- Expired Overlay -->
        <div v-else-if="status === 'expired'" class="absolute inset-0 bg-[#151517]/90 backdrop-blur rounded-xl flex flex-col items-center justify-center p-4 space-y-2 z-10">
          <RefreshCw class="w-8 h-8 text-amber-400 animate-spin" />
          <p class="text-xs font-semibold text-gray-200">QR Code Expired</p>
          <button @click="refreshQR" class="text-[11px] text-[#7B96F5] hover:underline">Click to refresh</button>
        </div>

        <!-- Rendered QR Code Image -->
        <img v-if="qrDataUrl && status !== 'offline'" :src="qrDataUrl" class="w-full h-full object-contain" alt="WhatsApp QR Code" />
        <div v-else-if="status === 'offline'" class="flex flex-col items-center gap-2 text-center p-2">
          <span class="text-2xl">⚠️</span>
          <span class="text-[10px] font-mono text-[#F16565]">Sidecar offline</span>
          <span class="text-[9px] text-gray-500 leading-tight">{{ errorMsg }}</span>
        </div>
        <div v-else class="flex flex-col items-center gap-2 text-gray-400">
          <RefreshCw class="w-6 h-6 animate-spin text-[#7B96F5]" />
          <span class="text-[10px] font-mono">Generating Live QR...</span>
        </div>
      </div>

      <!-- Instructions -->
      <div class="space-y-1 text-xs text-gray-400 max-w-sm">
        <p class="font-semibold text-gray-200">How to connect WhatsApp Gateway:</p>
        <p>1. Open WhatsApp on your phone &rarr; Linked Devices</p>
        <p>2. Tap "Link a Device" and point camera at this QR code</p>
      </div>
    </div>

    <template #footer>
      <div class="flex items-center justify-between w-full text-xs">
        <span class="text-gray-500 font-mono text-[10px]">Baileys Node.js HTTP Sidecar v2.4</span>
        <button
          @click="$emit('close')"
          class="px-4 py-2 rounded-lg border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-medium transition-colors"
        >
          Cancel
        </button>
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue';
import Modal from '../common/Modal.vue';
import { QrCode, CheckCircle2, RefreshCw } from 'lucide-vue-next';
import api from '../../api/client';
import { wsClient } from '../../ws/websocket';
import QRCode from 'qrcode';

const props = defineProps<{ isOpen: boolean }>();
const emit = defineEmits(['close', 'connected']);

const status = ref<'pending' | 'connected' | 'expired' | 'offline'>('pending');
const qrDataUrl = ref<string | null>(null);
const linkedNumber = ref<string | null>(null);
const errorMsg = ref<string | null>(null);
let pollTimer: any = null;
let unsubscribeWS: (() => void) | null = null;

const statusText = computed(() => {
  switch (status.value) {
    case 'connected': return 'Connected to WhatsApp';
    case 'expired': return 'QR Expired — Regenerating...';
    case 'offline': return 'Sidecar Service Offline';
    default: return 'Waiting for scan...';
  }
});

const statusBadgeClass = computed(() => {
  switch (status.value) {
    case 'connected': return 'bg-[#34D399]/15 text-[#34D399] border-[#34D399]/30';
    case 'expired': return 'bg-amber-500/15 text-amber-400 border-amber-500/30';
    case 'offline': return 'bg-[#F16565]/15 text-[#F16565] border-[#F16565]/30';
    default: return 'bg-[#7B96F5]/15 text-[#7B96F5] border-[#7B96F5]/30';
  }
});

const statusDotClass = computed(() => {
  switch (status.value) {
    case 'connected': return 'bg-[#34D399]';
    case 'expired': return 'bg-amber-400';
    case 'offline': return 'bg-[#F16565]';
    default: return 'bg-[#7B96F5] animate-ping';
  }
});

async function renderQR(qrString: string) {
  if (!qrString || typeof qrString !== 'string') return;
  try {
    qrDataUrl.value = await QRCode.toDataURL(qrString, {
      margin: 1,
      width: 240,
      color: { dark: '#000000', light: '#ffffff' }
    });
  } catch (e) {
    console.error('Failed to generate QR code canvas:', e);
  }
}

function setupWSSubscription() {
  cleanupWSSubscription();
  unsubscribeWS = wsClient.subscribe((data) => {
    if (data.type === 'whatsapp_qr' && data.qr) {
      renderQR(data.qr);
    } else if (data.type === 'whatsapp_status' && data.status === 'connected') {
      status.value = 'connected';
      linkedNumber.value = data.linkedNumber || null;
      stopPolling();
      setTimeout(() => {
        emit('connected', linkedNumber.value || '');
        emit('close');
      }, 1800);
    }
  });
}

function cleanupWSSubscription() {
  if (unsubscribeWS) {
    unsubscribeWS();
    unsubscribeWS = null;
  }
}

async function initiateConnect() {
  status.value = 'pending';
  errorMsg.value = null;
  qrDataUrl.value = null;
  setupWSSubscription();
  try {
    const res = await api.post('/integrations/whatsapp/connect');
    if (res.data.error) {
      // Sidecar is offline
      status.value = 'offline';
      errorMsg.value = res.data.error;
      return;
    }
    if (res.data.status === 'connected') {
      // Already connected — no QR needed
      status.value = 'connected';
      linkedNumber.value = res.data.linkedNumber || null;
      return;
    }
    // Pending — start polling for QR
    if (res.data.qr) {
      await renderQR(res.data.qr);
    }
  } catch (e: any) {
    // Real network error — sidecar not running
    status.value = 'offline';
    errorMsg.value = 'WhatsApp sidecar is not running. Start it with: cd backend/wa-sidecar && node index.js';
    console.error('[WhatsApp QR] Connect failed:', e.message);
    return;
  }
  startPolling();
}

function startPolling() {
  stopPolling();
  pollTimer = setInterval(async () => {
    try {
      const res = await api.get('/integrations/whatsapp/qr');
      if (res.data.error) {
        // Sidecar went offline mid-session
        status.value = 'offline';
        errorMsg.value = res.data.error;
        stopPolling();
        return;
      }
      if (res.data.status === 'connected') {
        status.value = 'connected';
        linkedNumber.value = res.data.linkedNumber || null;
        stopPolling();
        setTimeout(() => {
          emit('connected', linkedNumber.value || '');
          emit('close');
        }, 1800);
        return;
      } else if (res.data.qr) {
        await renderQR(res.data.qr);
      }
    } catch (e: any) {
      // Transient network blip — log but don't fake success
      console.warn('[WhatsApp QR] Poll error (backend may be restarting):', e.message);
    }
  }, 3000);
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

function refreshQR() {
  initiateConnect();
}

watch(() => props.isOpen, (open) => {
  if (open) {
    initiateConnect();
  } else {
    stopPolling();
    cleanupWSSubscription();
  }
});

onUnmounted(() => {
  stopPolling();
  cleanupWSSubscription();
});
</script>
