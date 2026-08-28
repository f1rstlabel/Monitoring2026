<template>
  <header class="h-16 bg-main/80 backdrop-blur border-b border-subtle flex items-center justify-between px-6 sticky top-0 z-20">
    <!-- Click Outside Backdrop for Dropdowns -->
    <div
      v-if="isNotifOpen"
      @click="closeDropdowns"
      class="fixed inset-0 z-30 bg-black/10 backdrop-blur-[1px]"
    ></div>

    <!-- Dynamic Page Breadcrumbs & Module Title -->
    <div class="flex items-center gap-3 select-none">
      <div class="flex items-center gap-2 text-xs font-mono">
        <span class="text-text-muted font-semibold tracking-wider uppercase text-[10px] bg-surface px-2 py-0.5 rounded border border-subtle">
          {{ pageMeta.category }}
        </span>
        <span class="text-text-muted">/</span>
        <div class="flex items-center gap-1.5 text-text-main font-bold">
          <component :is="pageMeta.icon" class="w-3.5 h-3.5 text-brand-periwinkle" />
          <span>{{ pageMeta.title }}</span>
        </div>
      </div>
    </div>

    <!-- Right Controls -->
    <div class="flex items-center gap-4 relative z-40">
      <!-- Live Status Indicator & Refresh Now -->
      <div class="flex items-center gap-2">
        <div class="flex items-center gap-2 px-3 py-1 rounded-full bg-surface border border-subtle text-xs">
          <span 
            class="w-2 h-2 rounded-full inline-block"
            :class="liveStore.isConnected ? 'bg-status-up pulsing-dot-green' : 'bg-amber-500'"
          ></span>
          <span class="text-text-secondary font-mono text-[11px]">
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
          class="px-2.5 py-1 rounded-full bg-surface border border-subtle hover:border-brand-periwinkle text-brand-periwinkle hover:text-brand-periwinkle-hover text-[11px] font-mono font-semibold transition-all flex items-center gap-1 disabled:opacity-50 cursor-pointer"
          title="Poll Now"
        >
          <RefreshCw class="w-3 h-3" :class="isRefreshing ? 'animate-spin' : ''" />
          <span>{{ isRefreshing ? 'Polling...' : 'Poll Now' }}</span>
        </button>

        <!-- Network Diagnostic Terminal Button -->
        <button
          type="button"
          @click="isTerminalOpen = true"
          class="px-2.5 py-1 rounded-full bg-surface border border-subtle hover:border-status-up text-status-up hover:text-status-up text-[11px] font-mono font-semibold transition-all flex items-center gap-1.5 cursor-pointer shadow-sm shadow-status-up/10"
          title="Open Network Diagnostic Terminal (Ping & Traceroute)"
        >
          <Terminal class="w-3 h-3" />
          <span>Diagnostics</span>
        </button>
      </div>

      <!-- Active Role Badge & Profile Quick Link -->
      <router-link
        to="/profile"
        title="View & Edit Profile"
        class="flex items-center gap-1.5 px-3 py-1 rounded-full border bg-surface text-xs font-mono font-semibold hover:ring-1 hover:ring-brand-periwinkle transition-all cursor-pointer"
        :class="roleBadgeClass"
      >
        <ShieldCheck v-if="authStore.user.role === 'admin'" class="w-3.5 h-3.5" />
        <Eye v-else-if="authStore.user.role === 'pimpinan'" class="w-3.5 h-3.5" />
        <Cpu v-else class="w-3.5 h-3.5" />
        <span class="uppercase text-[10px] tracking-wider">{{ roleBadgeLabel }}</span>
      </router-link>

      <!-- Theme Toggle Button -->
      <button
        @click="themeStore.toggleTheme"
        class="relative p-2 rounded-lg bg-surface border border-subtle text-text-secondary hover:text-text-main hover:bg-card transition-colors cursor-pointer"
        title="Toggle Theme"
      >
        <Sun v-if="themeStore.currentTheme === 'dark'" class="w-4 h-4" />
        <Moon v-else class="w-4 h-4" />
      </button>

      <!-- Notifications Bell Button & Dropdown -->
      <div class="relative">
        <button
          @click.stop="toggleNotif"
          class="relative p-2 rounded-lg bg-surface border border-subtle text-text-secondary hover:text-text-main hover:bg-card transition-colors cursor-pointer"
          title="System Notifications"
        >
          <Bell class="w-4 h-4" />
          <span v-if="notifStore.unreadCount > 0" class="absolute top-1.5 right-1.5 w-2 h-2 bg-status-down rounded-full ring-2 ring-main"></span>
        </button>

        <!-- Notifications Dropdown Panel -->
        <div
          v-if="isNotifOpen"
          class="absolute right-0 mt-2 w-80 sm:w-96 bg-surface border border-subtle rounded-xl shadow-2xl z-50 overflow-hidden"
        >
          <div class="p-3 border-b border-subtle flex items-center justify-between bg-card">
            <div class="flex items-center gap-2">
              <Bell class="w-4 h-4 text-brand-periwinkle" />
              <h3 class="text-xs font-bold text-text-main font-mono">System Notifications</h3>
              <span v-if="notifStore.unreadCount > 0" class="px-2 py-0.5 rounded-full text-[10px] font-mono bg-status-down/20 text-status-down border border-status-down/30">
                {{ notifStore.unreadCount }} new
              </span>
            </div>
            <button
              v-if="notifStore.unreadCount > 0"
              @click="notifStore.markAllAsRead"
              class="text-[11px] font-mono text-brand-periwinkle hover:underline cursor-pointer"
            >
              Mark all as read
            </button>
          </div>

          <div class="max-h-80 overflow-y-auto divide-y divide-subtle">
            <div
              v-for="item in notifStore.notifications"
              :key="item.id"
              @click="onNotifClick(item)"
              class="p-3 hover:bg-card cursor-pointer transition-colors flex items-start gap-3"
              :class="item.isUnread ? 'bg-brand-periwinkle/5' : ''"
            >
              <div class="p-2 rounded-lg bg-card border border-subtle shrink-0 mt-0.5">
                <AlertTriangle v-if="item.type === 'INCIDENT_NEW'" class="w-3.5 h-3.5 text-status-down" />
                <Activity v-else-if="item.type === 'FLAP_ALERT'" class="w-3.5 h-3.5 text-amber-400" />
                <CheckCircle2 v-else-if="item.type === 'INCIDENT_RESOLVED'" class="w-3.5 h-3.5 text-status-up" />
                <MessageSquare v-else class="w-3.5 h-3.5 text-brand-periwinkle" />
              </div>

              <div class="flex-1 space-y-0.5">
                <div class="flex items-center justify-between text-xs">
                  <span class="font-bold text-text-main" :class="item.isUnread ? 'font-extrabold' : ''">{{ item.title }}</span>
                  <span v-if="item.isUnread" class="w-1.5 h-1.5 rounded-full bg-brand-periwinkle"></span>
                </div>
                <p class="text-[11px] text-text-secondary leading-snug">{{ item.message }}</p>
                <p class="text-[10px] font-mono text-text-muted pt-1">{{ formatRelativeTime(item.timestamp) }}</p>
              </div>
            </div>
            <div v-if="notifStore.notifications.length === 0" class="p-8 text-center text-xs text-text-muted font-mono">
              No new notifications
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Diagnostic Terminal Modal -->
    <DiagnosticTerminalModal
      :is-open="isTerminalOpen"
      @close="isTerminalOpen = false"
    />
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import {
  Bell,
  ShieldCheck,
  Eye,
  Cpu,
  AlertTriangle,
  Activity,
  CheckCircle2,
  MessageSquare,
  RefreshCw,
  Terminal,
  LayoutDashboard,
  Server,
  Network,
  BarChart3,
  Settings,
  Users,
  User,
  Sun,
  Moon
} from 'lucide-vue-next';
import DiagnosticTerminalModal from '../diagnostics/DiagnosticTerminalModal.vue';
import { useDeviceStore } from '../../stores/deviceStore';
import { useLiveStore } from '../../stores/liveStore';
import { useAuthStore } from '../../stores/authStore';
import { useThemeStore } from '../../stores/themeStore';
import { useNotificationStore, type AppNotification } from '../../stores/notificationStore';
import { dashboardApi } from '../../api';

const router = useRouter();
const route = useRoute();
const deviceStore = useDeviceStore();
const liveStore = useLiveStore();
const authStore = useAuthStore();
const notifStore = useNotificationStore();
const themeStore = useThemeStore();

const pageMeta = computed(() => {
  const path = route.path;
  if (path === '/' || path === '/dashboard') {
    return { title: 'Live Dashboard', category: 'Monitoring', icon: LayoutDashboard };
  }
  if (path.startsWith('/devices/')) {
    return { title: 'Device Telemetry & Analytics', category: 'Infrastructure', icon: Cpu };
  }
  if (path === '/devices') {
    return { title: 'Device Inventory', category: 'Infrastructure', icon: Server };
  }
  if (path.startsWith('/incidents/')) {
    return { title: 'Incident Investigation', category: 'Operations', icon: AlertTriangle };
  }
  if (path === '/incidents') {
    return { title: 'Incident Center', category: 'Operations', icon: AlertTriangle };
  }
  if (path === '/topology') {
    return { title: 'Network Topology Map', category: 'Infrastructure', icon: Network };
  }
  if (path === '/reports') {
    return { title: 'Availability & SLA Reports', category: 'Analytics', icon: BarChart3 };
  }
  if (path === '/settings') {
    return { title: 'System Settings', category: 'Administration', icon: Settings };
  }
  if (path === '/users') {
    return { title: 'User & Role Management', category: 'Administration', icon: Users };
  }
  if (path === '/profile') {
    return { title: 'User Profile', category: 'Account', icon: User };
  }
  return { title: 'Network Operations Center', category: 'SANOC', icon: Activity };
});

const isNotifOpen = ref(false);
const isRefreshing = ref(false);
const isTerminalOpen = ref(false);

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
}

function toggleNotif() {
  isNotifOpen.value = !isNotifOpen.value;
  if (isNotifOpen.value) {
    notifStore.fetchNotifications();
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
    admin: 'ADMIN',
    pimpinan: 'LEADERSHIP',
    anggota: 'OPERATOR'
  };
  return map[authStore.user.role] ?? authStore.user.role;
});

const roleBadgeClass = computed(() => {
  const map: Record<string, string> = {
    admin: 'border-brand-periwinkle/40 text-brand-periwinkle bg-brand-periwinkle/10',
    pimpinan: 'border-amber-500/40 text-amber-400 bg-amber-500/10',
    anggota: 'border-status-up/40 text-status-up bg-status-up/10'
  };
  return map[authStore.user.role] ?? 'border-gray-500/40 text-text-secondary';
});

onMounted(() => {
  notifStore.fetchNotifications();
});
</script>
