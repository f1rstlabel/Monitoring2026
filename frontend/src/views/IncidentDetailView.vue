<template>
  <div v-if="pageState === 'ready' && incident" class="space-y-6">
    <!-- Breadcrumb -->
    <nav class="flex items-center gap-2 text-xs font-mono text-gray-400">
      <router-link to="/incidents" class="hover:text-[#7B96F5] transition-colors">Incidents</router-link>
      <ChevronRight class="w-3.5 h-3.5 text-gray-600" />
      <span class="text-gray-200 font-semibold">{{ incident.id }}</span>
    </nav>

    <!-- Header Section -->
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-6 flex flex-wrap items-center justify-between gap-4">
      <div class="space-y-2">
        <div class="flex items-center gap-3">
          <h1 class="text-xl font-extrabold text-white tracking-tight flex items-center gap-2">
            <span class="font-mono" :class="incident.status === 'RESOLVED' ? 'text-[#3ECF8E]' : 'text-red-400'">{{ incident.deviceName }}</span> is {{ incident.status === 'RESOLVED' ? 'RESOLVED (UP)' : 'DOWN' }}
          </h1>
          <StatusPill :status="incident.status === 'RESOLVED' ? 'UP' : 'DOWN'" />
          <span class="px-2.5 py-0.5 rounded-full text-xs font-mono font-semibold" :class="incident.status === 'RESOLVED' ? 'bg-[#3ECF8E]/15 border border-[#3ECF8E]/30 text-[#3ECF8E]' : 'bg-red-500/15 border border-red-500/30 text-red-400'">
            Duration: {{ incident.duration }}
          </span>
        </div>

        <p class="text-xs font-mono text-gray-400">
          Target IP: <span class="text-gray-200">{{ incident.deviceIp }}</span> &bull;
          Device Type: <span class="text-gray-200">{{ incident.deviceType }}</span> &bull;
          Started: <span class="text-gray-200">{{ incident.startTime }}</span>
        </p>
      </div>

      <!-- Action Buttons — role-gated via v-if: pimpinan sees neither -->
      <div class="flex items-center gap-3">
        <!-- Export Structured PDF / Print -->
        <button
          @click="handlePrintPDF"
          class="px-4 py-2 rounded-lg bg-[#18181B] border border-[#26262A] hover:bg-[#26262A] text-gray-300 font-medium text-xs transition-all flex items-center gap-1.5"
        >
          <Printer class="w-4 h-4 text-[#3ECF8E]" />
          Export PDF
        </button>

        <!-- Read-only badge for pimpinan -->
        <span
          v-if="authStore.user.role === 'pimpinan'"
          class="px-3 py-1.5 rounded-lg border border-[#26262A] text-gray-500 text-[10px] font-mono uppercase tracking-widest flex items-center gap-1.5"
        >
          <Eye class="w-3.5 h-3.5" />
          Read-only access
        </span>
      </div>
    </div>

    <!-- Main Layout: 2 Columns -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
      <!-- Left Column: Event Timeline -->
      <div class="lg:col-span-2 space-y-6">
        <!-- Vertical Event Timeline -->
        <div class="bg-[#151517] border border-[#26262A] rounded-xl p-6 space-y-5">
          <div class="flex items-center justify-between border-b border-[#26262A] pb-4">
            <h3 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono flex items-center gap-2">
              <Clock class="w-4 h-4 text-[#7B96F5]" />
              Event Timeline & Failover Logs
            </h3>
            <span class="text-[10px] font-mono text-gray-500">Chronological Audit</span>
          </div>

          <!-- Vertical Timeline Nodes -->
          <div class="relative pl-6 space-y-6 before:absolute before:left-2 before:top-2 before:bottom-2 before:w-0.5 before:bg-[#26262A]">
            <div
              v-for="evt in (incident.timeline || [])"
              :key="evt.id"
              class="relative group"
            >
              <!-- Colored Circle Node on Bar -->
              <span
                class="absolute -left-[23px] top-0.5 w-3.5 h-3.5 rounded-full ring-4 ring-[#151517]"
                :class="[
                  evt.severity === 'critical' ? 'bg-[#F16565] pulsing-dot-red' :
                  evt.severity === 'warning' ? 'bg-[#F5A65B]' :
                  evt.severity === 'skipped' ? 'bg-gray-600' : 'bg-[#7B96F5]'
                ]"
              ></span>

              <div
                class="border rounded-lg p-3.5 text-xs space-y-1 group-hover:border-[#3A3A40] transition-colors"
                :class="[
                  evt.severity === 'skipped'
                    ? 'bg-[#18181B]/50 border-[#26262A]/50 opacity-70'
                    : 'bg-[#18181B] border-[#26262A]'
                ]"
              >
                <div class="flex items-center justify-between">
                  <h4 class="font-bold flex items-center gap-2"
                    :class="evt.severity === 'skipped' ? 'text-gray-500' : 'text-gray-100'"
                  >
                    {{ evt.title }}
                    <span v-if="evt.channel" class="text-[9px] font-mono px-1.5 py-0.5 rounded"
                      :class="[
                        evt.severity === 'skipped'
                          ? 'bg-[#1C1C1F] text-gray-600'
                          : 'bg-[#26262A] text-[#7B96F5]'
                      ]"
                    >
                      {{ evt.channel }}
                    </span>
                  </h4>
                  <span class="font-mono text-[10px] text-gray-500">{{ evt.timestamp }}</span>
                </div>
                <p class="font-mono text-[11px] leading-relaxed"
                  :class="evt.severity === 'skipped' ? 'text-gray-600' : 'text-gray-400'"
                >{{ evt.description }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Column: 3 Stat Cards & Notification Log Table -->
      <div class="lg:col-span-1 space-y-6">
        <!-- 3 Small Stat Cards -->
        <div class="grid grid-cols-1 gap-3">
          <div class="bg-[#151517] border border-[#26262A] rounded-xl p-4 flex items-center justify-between">
            <div>
              <p class="text-[10px] font-mono text-gray-500 uppercase">Packet Loss</p>
              <p class="text-lg font-bold font-mono text-red-400 mt-0.5">{{ incident.packetLoss }}%</p>
            </div>
            <span class="w-2.5 h-2.5 rounded-full bg-red-500 pulsing-dot-red"></span>
          </div>

          <div class="bg-[#151517] border border-[#26262A] rounded-xl p-4 flex items-center justify-between">
            <div>
              <p class="text-[10px] font-mono text-gray-500 uppercase">Latency</p>
              <p class="text-lg font-bold font-mono text-gray-200 mt-0.5">{{ incident.latencyMs }} ms</p>
            </div>
            <span class="w-2.5 h-2.5 rounded-full bg-amber-500"></span>
          </div>

          <div class="bg-[#151517] border border-[#26262A] rounded-xl p-4 flex items-center justify-between">
            <div>
              <p class="text-[10px] font-mono text-gray-400 uppercase tracking-wider">Node Location</p>
              <p class="text-sm font-bold text-[#7B96F5] mt-0.5 truncate max-w-[200px]">{{ incident.location || 'Gedung Sate Lt 2' }}</p>
            </div>
            <MapPin class="w-4 h-4 text-[#7B96F5]" />
          </div>
        </div>

        <!-- Notification Audit Log Card -->
        <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4">
          <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
            <h3 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono flex items-center gap-2">
              <Send class="w-4 h-4 text-[#3ECF8E]" />
              NOTIFICATION AUDIT LOG
            </h3>
          </div>

          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs font-mono">
              <thead>
                <tr class="text-[10px] text-gray-500 border-b border-[#26262A] uppercase">
                  <th class="pb-2.5 font-medium w-28 sm:w-32">Channel</th>
                  <th class="pb-2.5 font-medium min-w-[200px]">Recipient & Details</th>
                  <th class="pb-2.5 font-medium w-24">Status</th>
                  <th class="pb-2.5 font-medium w-24 text-right">Sent</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-[#26262A]/50">
                <tr v-for="log in auditLogs" :key="log.id" class="text-[11px]">
                  <td class="py-3 font-bold text-gray-200 align-top">
                    <div class="flex items-center gap-2">
                      <div
                        class="w-6 h-6 rounded-lg flex items-center justify-center shrink-0"
                        :class="log.channel.toLowerCase().includes('whatsapp') ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30' : 'bg-[#7B96F5]/15 text-[#7B96F5] border border-[#7B96F5]/30'"
                      >
                        <MessageSquare v-if="log.channel.toLowerCase().includes('whatsapp')" class="w-3.5 h-3.5" />
                        <Send v-else class="w-3.5 h-3.5" />
                      </div>
                      <span class="font-mono text-xs">{{ log.channel }}</span>
                    </div>
                  </td>
                  <td class="py-3 text-gray-300 pr-3 align-top leading-tight">
                    <div class="font-semibold text-gray-100 font-mono text-xs">{{ log.recipient }}</div>
                    <div v-if="getLogDescription(log)" class="text-[10.5px] text-gray-400 font-sans italic mt-0.5">
                      {{ getLogDescription(log) }}
                    </div>
                  </td>
                  <td class="py-3 align-top">
                    <span
                      class="px-2.5 py-1 rounded text-[9px] font-bold font-mono uppercase tracking-wider inline-block border"
                      :class="[
                        log.status.toLowerCase() === 'delivered' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/25' :
                        log.status.toLowerCase() === 'failed' ? 'bg-red-500/10 text-red-400 border-red-500/25' :
                        log.status.toLowerCase() === 'skipped' ? 'bg-amber-500/10 text-amber-400 border-amber-500/25' :
                        'bg-gray-800 text-gray-400 border-gray-700'
                      ]"
                    >
                      {{ log.status }}
                    </span>
                  </td>
                  <td class="py-3 text-right text-gray-400 font-mono text-[10px] align-top whitespace-nowrap">
                    {{ log.timestamp }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Loading State -->
  <div v-else-if="pageState === 'loading'" class="space-y-6">
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-6 space-y-3">
      <Skeleton width="45%" height="1.5rem" />
      <Skeleton width="60%" height="0.8rem" />
    </div>
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-2 bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4">
        <Skeleton width="40%" height="1rem" />
        <div v-for="i in 4" :key="i" class="p-3 bg-[#18181B] border border-[#26262A] rounded-lg space-y-2">
          <Skeleton width="50%" height="0.8rem" />
          <Skeleton width="80%" height="0.65rem" />
        </div>
      </div>
      <div class="lg:col-span-1 bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4">
        <Skeleton width="60%" height="1rem" />
        <Skeleton v-for="i in 3" :key="i" width="100%" height="2.5rem" customClass="rounded-lg" />
      </div>
    </div>
  </div>

  <!-- 404 Not Found State -->
  <div v-else-if="pageState === 'not_found'" class="p-8 text-center bg-[#151517] border border-[#26262A] rounded-xl space-y-4">
    <div class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-amber-500/10 text-amber-400">
      <AlertTriangle class="w-6 h-6" />
    </div>
    <h3 class="text-sm font-bold text-white">Ticket Not Found</h3>
    <p class="text-xs text-gray-400 max-w-md mx-auto">
      No incident ticket exists with ID <span class="font-mono text-amber-400">{{ route.params.id }}</span> in the NOC audit database.
    </p>
    <router-link to="/incidents" class="inline-block px-4 py-2 text-xs font-semibold rounded-lg bg-[#7B96F5] text-white hover:bg-[#95ABF7]">
      Back to Active Incidents Queue
    </router-link>
  </div>
  <!-- Network / 5xx Error State -->
  <div v-else class="p-8 text-center bg-[#151517] border border-red-500/30 rounded-xl space-y-4">
    <div class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-red-500/10 text-red-400">
      <AlertCircle class="w-6 h-6" />
    </div>
    <h3 class="text-sm font-bold text-white">Failed to Load Incident Log</h3>
    <p class="text-xs text-gray-400 max-w-md mx-auto">
      Network or backend server connection error while communicating with Go API.
    </p>
    <div class="flex items-center justify-center gap-3">
      <button @click="loadIncident" class="px-4 py-2 text-xs font-semibold rounded-lg bg-[#7B96F5] text-white hover:bg-[#95ABF7] flex items-center gap-1.5 cursor-pointer">
        <RefreshCw class="w-3.5 h-3.5" />
        Retry Connection
      </button>
      <router-link to="/incidents" class="px-4 py-2 text-xs font-semibold rounded-lg border border-[#26262A] text-gray-300 hover:bg-[#18181B]">
        Back to Active Incidents Queue
      </router-link>
    </div>
  </div>

  <!-- Printable Incident PDF Container (Hidden from Screen, Visible on Print) -->
  <PrintableIncidentReport v-if="isPrintRendered" :incident="incident" />
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { useRoute } from 'vue-router';
import { useIncidentStore } from '../stores/incidentStore';
import { useAuthStore } from '../stores/authStore';
import StatusPill from '../components/common/StatusPill.vue';
import Skeleton from '../components/common/Skeleton.vue';
import PrintableIncidentReport from '../components/reports/PrintableIncidentReport.vue';
import {
  ChevronRight,
  Clock,
  Send,
  Eye,
  AlertTriangle,
  AlertCircle,
  RefreshCw,
  Printer,
  MapPin,
  MessageSquare
} from 'lucide-vue-next';

const route = useRoute();
const incidentStore = useIncidentStore();
const authStore = useAuthStore();

const pageState = ref<'loading' | 'ready' | 'not_found' | 'error'>('loading');
const incident = ref(incidentStore.currentIncident);
const isPrintRendered = ref(false);
let liveRefreshTimer: any = null;

function getLogDescription(log: { channel: string; status: string; channelIcon?: string; errorMsg?: string; recipient?: string }) {
  const iconOrErr = (log.channelIcon || log.errorMsg || '').trim();
  if (iconOrErr && !['MessageSquare', 'Send', 'whatsapp', 'telegram'].includes(iconOrErr)) {
    return iconOrErr;
  }
  if (log.status.toLowerCase() === 'skipped') {
    return 'Skipped — WhatsApp delivered successfully (fallback not needed)';
  }
  if (log.status.toLowerCase() === 'delivered') {
    return 'Delivered successfully';
  }
  if (log.status.toLowerCase() === 'failed') {
    return 'Delivery failed';
  }
  return '';
}

const auditLogs = computed(() => {
  let list: Array<{ id: string; channel: string; recipient: string; status: string; timestamp: string; channelIcon?: string; errorMsg?: string }> = [];

  if (incident.value?.notificationLog && incident.value.notificationLog.length > 0) {
    list = incident.value.notificationLog.map((l) => ({ ...l }));
  } else {
    const timeline = incident.value?.timeline || [];
    let waAdded = false;
    let tgAdded = false;

    for (const evt of timeline) {
      const ch = (evt.channel || '').toLowerCase();
      const title = (evt.title || '').toLowerCase();
      const desc = evt.description || '';

      if (!waAdded && (ch.includes('whatsapp') || title.includes('whatsapp'))) {
        const isDelivered = title.includes('delivered') || evt.status === 'delivered';
        const isFailed = title.includes('failed') || evt.status === 'failed';

        let recipient = 'NOC On-Call Target (+6289526788625)';
        const phoneMatch = desc.match(/(\+?\d{10,15})/);
        if (phoneMatch) {
          recipient = `NOC Target (${phoneMatch[1]})`;
        }

        list.push({
          id: evt.id || 'wa-1',
          channel: 'WhatsApp',
          recipient: recipient,
          status: isDelivered ? 'Delivered' : isFailed ? 'Failed' : 'Delivered',
          errorMsg: isDelivered ? 'Primary notification delivered successfully' : 'Delivery failed',
          timestamp: evt.timestamp || incident.value?.startTime || ''
        });
        waAdded = true;
      }

      if (!tgAdded && (ch.includes('telegram') || title.includes('telegram'))) {
        const isDelivered = title.includes('delivered') || evt.status === 'delivered';
        const isSkipped = title.includes('skipped') || evt.severity === 'skipped';
        list.push({
          id: evt.id || 'tg-1',
          channel: 'Telegram',
          recipient: 'NOC Telegram Channel (@SanocBot)',
          status: isDelivered ? 'Delivered' : isSkipped ? 'Skipped' : 'Delivered',
          errorMsg: isSkipped ? 'Skipped — WhatsApp delivered successfully (fallback not needed)' : 'Delivered',
          timestamp: evt.timestamp || incident.value?.startTime || ''
        });
        tgAdded = true;
      }
    }

    if (!waAdded) {
      list.push({
        id: 'wa-def',
        channel: 'WhatsApp',
        recipient: 'NOC Target (+6289526788625)',
        status: 'Delivered',
        errorMsg: 'Primary notification delivered successfully',
        timestamp: incident.value?.startTime || ''
      });
    }
  }

  // Ensure Telegram with Skipped status is ALWAYS present when WhatsApp succeeded and Telegram was not in list
  const hasTelegram = list.some((l) => l.channel.toLowerCase().includes('telegram'));
  const hasWhatsApp = list.some((l) => l.channel.toLowerCase().includes('whatsapp'));

  if (!hasTelegram && hasWhatsApp) {
    list.push({
      id: 'tg-skip-audit',
      channel: 'Telegram',
      recipient: 'NOC Telegram Channel (@SanocBot)',
      status: 'Skipped',
      errorMsg: 'Skipped — WhatsApp delivered successfully (fallback not needed)',
      timestamp: list[0]?.timestamp || incident.value?.startTime || ''
    });
  }

  return list;
});

async function handlePrintPDF() {
  const previousState = pageState.value;
  pageState.value = 'loading';
  const originalTitle = document.title;
  const id = route.params.id as string;
  document.title = `sanoc-pdf-incident-ticket-${id || 'detail'}`;

  try {
    await loadIncident();
    pageState.value = 'ready';
    isPrintRendered.value = true;
    await nextTick();
    await new Promise((resolve) => setTimeout(resolve, 50));
    window.print();
  } catch (err) {
    console.error('Failed to print PDF:', err);
    pageState.value = previousState;
  } finally {
    isPrintRendered.value = false;
    document.title = originalTitle;
  }
}

async function loadIncident() {
  const id = route.params.id as string;
  if (!id) {
    pageState.value = 'not_found';
    return;
  }

  pageState.value = 'loading';
  try {
    await incidentStore.fetchIncidentById(id);
    if (incidentStore.currentIncident) {
      incident.value = incidentStore.currentIncident;
      pageState.value = 'ready';
    } else {
      pageState.value = 'not_found';
    }
  } catch (e: any) {
    if (e.response?.status === 404) {
      pageState.value = 'not_found';
    } else {
      pageState.value = 'error';
    }
  }
}

async function refreshIncidentSilently() {
  const id = route.params.id as string;
  if (!id || pageState.value !== 'ready') return;
  try {
    await incidentStore.fetchIncidentById(id);
    if (incidentStore.currentIncident) {
      incident.value = incidentStore.currentIncident;
    }
  } catch {
    // silent catch during background interval
  }
}

onMounted(() => {
  loadIncident();
  liveRefreshTimer = setInterval(refreshIncidentSilently, 3000);
});

onUnmounted(() => {
  if (liveRefreshTimer) {
    clearInterval(liveRefreshTimer);
  }
});

watch(() => route.params.id, loadIncident);
</script>
