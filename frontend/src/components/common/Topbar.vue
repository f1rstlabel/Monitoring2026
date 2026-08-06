<template>
  <header class="h-16 bg-[#0D0D0F]/80 backdrop-blur border-b border-[#26262A] flex items-center justify-between px-6 sticky top-0 z-20">
    <!-- Click Outside Backdrop for Dropdowns -->
    <div
      v-if="isNotifOpen || isHelpOpen"
      @click="closeDropdowns"
      class="fixed inset-0 z-30 bg-black/10 backdrop-blur-[1px]"
    ></div>

    <!-- Search Input -->
    <div class="relative w-80 z-20">
      <Search class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
      <input
        v-model="deviceStore.searchQuery"
        type="text"
        placeholder="Search devices, IPs, or incidents..."
        class="w-full bg-[#151517] border border-[#26262A] rounded-lg pl-9 pr-3 py-1.5 text-xs text-gray-200 placeholder-gray-500 focus:outline-none focus:border-[#7B96F5] transition-colors"
      />
      <span v-if="deviceStore.searchQuery" @click="deviceStore.searchQuery = ''" class="absolute right-2.5 top-1/2 -translate-y-1/2 text-xs text-gray-500 cursor-pointer hover:text-gray-300">✕</span>
    </div>

    <!-- Right Controls -->
    <div class="flex items-center gap-4 relative z-40">
      <!-- Live Status Indicator & Refresh Now -->
      <div class="flex items-center gap-2">
        <div class="flex items-center gap-2 px-3 py-1 rounded-full bg-[#151517] border border-[#26262A] text-xs">
          <span 
            class="w-2 h-2 rounded-full inline-block"
            :class="liveStore.isConnected ? 'bg-[#3ECF8E] pulsing-dot-green' : 'bg-amber-500'"
          ></span>
          <span class="text-gray-300 font-mono text-[11px]">
            <template v-if="liveStore.isConnected">
              Live Feed &bull; Updated {{ liveStore.lastUpdatedAgo }}s ago
            </template>
            <template v-else>
              Connecting WebSocket...
            </template>
          </span>
        </div>

        <button
          @click="handleManualRefresh"
          :disabled="isRefreshing"
          class="px-2.5 py-1 rounded-full bg-[#151517] border border-[#26262A] hover:border-[#7B96F5] text-[#7B96F5] hover:text-[#95ABF7] text-[11px] font-mono font-semibold transition-all flex items-center gap-1 disabled:opacity-50"
          title="Force immediate ICMP poll cycle across all devices"
        >
          <RefreshCw class="w-3 h-3" :class="isRefreshing ? 'animate-spin' : ''" />
          <span>{{ isRefreshing ? 'Polling...' : 'Refresh Now' }}</span>
        </button>
      </div>

      <!-- Active Role Badge -->
      <div class="flex items-center gap-1.5 px-3 py-1 rounded-full border bg-[#151517] text-xs font-mono font-semibold" :class="roleBadgeClass">
        <ShieldCheck v-if="authStore.user.role === 'superadmin'" class="w-3.5 h-3.5" />
        <Eye v-else-if="authStore.user.role === 'pimpinan'" class="w-3.5 h-3.5" />
        <Cpu v-else class="w-3.5 h-3.5" />
        <span class="uppercase text-[10px] tracking-wider">{{ roleBadgeLabel }}</span>
      </div>

      <!-- Notifications Bell Button & Dropdown -->
      <div class="relative">
        <button
          @click.stop="toggleNotif"
          class="relative p-2 rounded-lg bg-[#151517] border border-[#26262A] text-gray-400 hover:text-gray-200 hover:bg-[#18181B] transition-colors"
          title="Notifikasi Incident & Gateways"
        >
          <Bell class="w-4 h-4" />
          <span v-if="notifStore.unreadCount > 0" class="absolute top-1.5 right-1.5 w-2 h-2 bg-[#F16565] rounded-full ring-2 ring-[#0D0D0F]"></span>
        </button>

        <!-- Notifications Dropdown Panel -->
        <div
          v-if="isNotifOpen"
          class="absolute right-0 mt-2 w-80 sm:w-96 bg-[#151517] border border-[#26262A] rounded-xl shadow-2xl z-50 overflow-hidden"
        >
          <div class="p-3 border-b border-[#26262A] flex items-center justify-between bg-[#18181B]">
            <div class="flex items-center gap-2">
              <Bell class="w-4 h-4 text-[#7B96F5]" />
              <h3 class="text-xs font-bold text-white font-mono">Notifikasi Systems</h3>
              <span v-if="notifStore.unreadCount > 0" class="px-2 py-0.5 rounded-full text-[10px] font-mono bg-[#F16565]/20 text-[#F16565] border border-[#F16565]/30">
                {{ notifStore.unreadCount }} new
              </span>
            </div>
            <button
              v-if="notifStore.unreadCount > 0"
              @click="notifStore.markAllAsRead"
              class="text-[11px] font-mono text-[#7B96F5] hover:underline"
            >
              Mark all as read
            </button>
          </div>

          <div class="max-h-80 overflow-y-auto divide-y divide-[#26262A]">
            <div
              v-for="item in notifStore.notifications"
              :key="item.id"
              @click="onNotifClick(item)"
              class="p-3 hover:bg-[#18181B] cursor-pointer transition-colors flex items-start gap-3"
              :class="item.isUnread ? 'bg-[#7B96F5]/5' : ''"
            >
              <div class="p-2 rounded-lg bg-[#18181B] border border-[#26262A] shrink-0 mt-0.5">
                <AlertTriangle v-if="item.type === 'INCIDENT_NEW'" class="w-3.5 h-3.5 text-[#F16565]" />
                <Activity v-else-if="item.type === 'FLAP_ALERT'" class="w-3.5 h-3.5 text-amber-400" />
                <CheckCircle2 v-else-if="item.type === 'INCIDENT_RESOLVED'" class="w-3.5 h-3.5 text-[#34D399]" />
                <MessageSquare v-else class="w-3.5 h-3.5 text-[#7B96F5]" />
              </div>

              <div class="flex-1 space-y-0.5">
                <div class="flex items-center justify-between text-xs">
                  <span class="font-bold text-gray-200" :class="item.isUnread ? 'text-white font-extrabold' : ''">{{ item.title }}</span>
                  <span v-if="item.isUnread" class="w-1.5 h-1.5 rounded-full bg-[#7B96F5]"></span>
                </div>
                <p class="text-[11px] text-gray-400 leading-snug">{{ item.message }}</p>
                <p class="text-[10px] font-mono text-gray-500 pt-1">{{ formatRelativeTime(item.timestamp) }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>



      <!-- Help Button & Modal (Bantuan) -->
      <div class="relative">
        <button
          @click.stop="toggleHelp"
          class="p-2 rounded-lg bg-[#151517] border border-[#26262A] text-gray-400 hover:text-gray-200 hover:bg-[#18181B] transition-colors"
          title="Bantuan & Documentation"
        >
          <HelpCircle class="w-4 h-4" />
        </button>

        <!-- Help Dropdown Modal -->
        <div
          v-if="isHelpOpen"
          class="absolute right-0 mt-2 w-72 bg-[#151517] border border-[#26262A] rounded-xl shadow-2xl p-4 z-50 space-y-3"
        >
          <div class="flex items-center justify-between border-b border-[#26262A] pb-2">
            <h3 class="text-xs font-bold text-white flex items-center gap-2 font-mono">
              <HelpCircle class="w-4 h-4 text-[#7B96F5]" />
              Bantuan &amp; Dukungan NOC
            </h3>
            <span class="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-[#7B96F5]/10 text-[#7B96F5] border border-[#7B96F5]/30">
              v2.4.1-stable
            </span>
          </div>

          <div class="space-y-2 text-xs text-gray-300">
            <a
              href="https://diskominfo.jabarprov.go.id"
              target="_blank"
              class="flex items-center justify-between p-2.5 rounded-lg bg-[#18181B] border border-[#26262A] hover:border-[#7B96F5] text-gray-200 transition-colors"
            >
              <span>Dokumentasi Sistem NOC</span>
              <ExternalLink class="w-3.5 h-3.5 text-gray-400" />
            </a>

            <div class="p-3 rounded-lg bg-[#18181B] border border-[#26262A] space-y-1">
              <p class="font-mono text-[10px] uppercase text-gray-400 font-bold">Layanan Helpdesk NOC</p>
              <p class="text-xs font-mono text-[#7B96F5]">noc.alerts@jabarprov.go.id</p>
              <p class="text-[11px] text-gray-400">Ext: 4401 / 4402 (Hotline Biro Umum)</p>
              <p class="text-[10px] text-gray-500 pt-1 border-t border-[#26262A] mt-2">Jam Operasional: 24/7 Monitoring</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import {
  Search,
  Bell,
  HelpCircle,
  ShieldCheck,
  Eye,
  Cpu,
  AlertTriangle,
  Activity,
  CheckCircle2,
  MessageSquare,
  ExternalLink,
  RefreshCw
} from 'lucide-vue-next';
import { useDeviceStore } from '../../stores/deviceStore';
import { useLiveStore } from '../../stores/liveStore';
import { useAuthStore } from '../../stores/authStore';
import { useNotificationStore, type AppNotification } from '../../stores/notificationStore';
import { dashboardApi } from '../../api';

const router = useRouter();
const deviceStore = useDeviceStore();
const liveStore = useLiveStore();
const authStore = useAuthStore();
const notifStore = useNotificationStore();

const isNotifOpen = ref(false);
const isHelpOpen = ref(false);
const isRefreshing = ref(false);

async function handleManualRefresh() {
  isRefreshing.value = true;
  try {
    await dashboardApi.refreshNow();
    liveStore.lastUpdatedAgo = 0;
    await deviceStore.fetchDevices();
  } catch (e) {
    // offline
  } finally {
    setTimeout(() => {
      isRefreshing.value = false;
    }, 600);
  }
}

function closeDropdowns() {
  isNotifOpen.value = false;
  isHelpOpen.value = false;
}

function toggleNotif() {
  isNotifOpen.value = !isNotifOpen.value;
  if (isNotifOpen.value) {
    isHelpOpen.value = false;
    notifStore.fetchNotifications();
  }
}

function toggleHelp() {
  isHelpOpen.value = !isHelpOpen.value;
  if (isHelpOpen.value) {
    isNotifOpen.value = false;
  }
}

function onNotifClick(item: AppNotification) {
  item.isUnread = false;
  closeDropdowns();
  if (item.targetUrl) {
    router.push(item.targetUrl);
  }
}

function formatRelativeTime(isoString: string) {
  const diffMs = Date.now() - new Date(isoString).getTime();
  const mins = Math.floor(diffMs / 60000);
  if (mins < 1) return 'Just now';
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  return `${hours}h ago`;
}

const roleBadgeLabel = computed(() => {
  const map: Record<string, string> = {
    superadmin: 'SUPERADMIN',
    pimpinan: 'PIMPINAN',
    anggota: 'ANGGOTA NOC'
  };
  return map[authStore.user.role] ?? authStore.user.role;
});

const roleBadgeClass = computed(() => {
  const map: Record<string, string> = {
    superadmin: 'border-[#7B96F5]/40 text-[#7B96F5] bg-[#7B96F5]/10',
    pimpinan: 'border-amber-500/40 text-amber-400 bg-amber-500/10',
    anggota: 'border-[#34D399]/40 text-[#34D399] bg-[#34D399]/10'
  };
  return map[authStore.user.role] ?? 'border-gray-500/40 text-gray-400';
});

onMounted(() => {
  notifStore.fetchNotifications();
});
</script>
