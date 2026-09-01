<template>
  <aside class="w-60 bg-main border-r border-subtle flex flex-col justify-between h-screen fixed left-0 top-0 z-30 select-none">
    <!-- Top Branding & Navigation -->
    <div>
      <!-- Header / Logo -->
      <div class="px-4 py-4 border-b border-subtle/60 flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-surface border border-subtle flex items-center justify-center shrink-0 overflow-hidden relative shadow-sm">
          <img
            v-if="settingStore.branding.logoUrl"
            :src="settingStore.branding.logoUrl"
            alt="System Logo"
            class="w-full h-full block"
            :class="(settingStore.branding.logoFit || 'cover') === 'cover' ? 'object-cover' : 'object-contain p-1'"
          />
          <img
            v-else
            src="../../assets/logo-sanoc-mark.svg"
            alt="SANOC Logo"
            class="w-full h-full object-contain p-1.5"
          />
        </div>
        <div class="min-w-0 flex-1 flex flex-col justify-center">
          <div class="flex items-center justify-between gap-1">
            <h1 class="text-sm font-black text-text-main tracking-wider leading-none truncate">
              {{ settingStore.branding.appTitle || 'SANOC' }}
            </h1>
            <span class="text-[9px] font-mono text-brand-periwinkle font-semibold bg-brand-periwinkle/10 px-1 py-0.5 rounded border border-brand-periwinkle/20 shrink-0">v2.6.0</span>
          </div>
          <p
            class="text-[10px] font-mono text-text-secondary mt-1 uppercase tracking-tight truncate leading-tight"
            :title="settingStore.branding.appSubtitle || 'Jabar Regional SANOC'"
          >
            {{ settingStore.branding.appSubtitle || 'Jabar Regional SANOC' }}
          </p>
        </div>
      </div>

      <!-- Navigation Items — conditionally rendered per feature permissions -->
      <nav class="p-3 space-y-1 mt-2">
        <!-- Dashboard -->
        <router-link
          to="/dashboard"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-medium transition-all duration-150"
          :class="navLinkClass('/dashboard')"
        >
          <LayoutGrid class="w-4 h-4 shrink-0" />
          <span>Dashboard</span>
        </router-link>

        <!-- Devices -->
        <router-link
          v-if="authStore.hasPermission('devices.view')"
          to="/devices"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-medium transition-all duration-150"
          :class="navLinkClass('/devices')"
        >
          <Server class="w-4 h-4 shrink-0" />
          <span>Devices</span>
          <span class="ml-auto text-[10px] font-bold px-1.5 py-0.5 rounded-full bg-subtle text-text-secondary">
            {{ deviceStore.summary.totalDevices }}
          </span>
        </router-link>

        <!-- Incidents -->
        <router-link
          v-if="authStore.hasPermission('incidents.view')"
          to="/incidents"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-medium transition-all duration-150"
          :class="navLinkClass('/incidents')"
        >
          <AlertTriangle class="w-4 h-4 shrink-0" />
          <span>Incidents</span>
          <span
            v-if="deviceStore.summary.activeIncidents > 0"
            class="ml-auto text-[10px] font-bold px-1.5 py-0.5 rounded-full bg-red-500/20 text-red-400 border border-red-500/30"
          >
            {{ deviceStore.summary.activeIncidents }}
          </span>
        </router-link>

        <!-- Reports -->
        <router-link
          v-if="authStore.hasPermission('reports.view')"
          to="/reports"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-medium transition-all duration-150"
          :class="navLinkClass('/reports')"
        >
          <BarChart3 class="w-4 h-4 shrink-0" />
          <span>Reports</span>
        </router-link>

        <!-- Settings -->
        <router-link
          v-if="authStore.hasPermission('settings.view') || authStore.canSeeSettings"
          to="/settings"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-medium transition-all duration-150"
          :class="navLinkClass('/settings')"
        >
          <Settings class="w-4 h-4 shrink-0" />
          <span>Settings</span>
        </router-link>

        <!-- Help Center & Guides -->
        <router-link
          to="/pusat-bantuan"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-medium transition-all duration-150"
          :class="navLinkClass('/pusat-bantuan')"
        >
          <HelpCircle class="w-4 h-4 shrink-0" />
          <span>Help Center</span>
        </router-link>
      </nav>

      <!-- Role indicator badge -->
      <div class="px-4 mt-1">
        <div class="flex items-center gap-2 px-2 py-1.5 rounded-lg bg-main border border-subtle/50">
          <component :is="roleIcon" class="w-3 h-3 shrink-0" :class="roleColor" />
          <span class="text-[10px] font-mono uppercase tracking-widest" :class="roleColor">
            {{ roleLabel }}
          </span>
          <span class="ml-auto inline-block w-1.5 h-1.5 rounded-full bg-status-up" />
        </div>
      </div>
    </div>

    <!-- Bottom User Section -->
    <div class="p-3 border-t border-subtle/60 bg-main/50">
      <div class="flex items-center justify-between p-2 rounded-lg bg-surface border border-subtle">
        <router-link to="/profile" class="flex items-center gap-2.5 overflow-hidden group flex-1 mr-1" title="View & Edit Profile">
          <img
            :src="authStore.user.avatarUrl"
            alt="User Avatar"
            @error="(e: Event) => (e.target as HTMLImageElement).src = 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'"
            class="w-8 h-8 rounded-full object-cover border border-brand-periwinkle/40 shrink-0 group-hover:ring-2 group-hover:ring-brand-periwinkle transition-all"
          />
          <div class="overflow-hidden">
            <p class="text-xs font-medium text-text-main truncate leading-tight group-hover:text-brand-periwinkle transition-colors">{{ authStore.user.name }}</p>
            <p class="text-[10px] text-text-muted truncate mt-0.5">{{ authStore.user.email }}</p>
          </div>
        </router-link>
        <button
          @click="handleLogout"
          title="Logout"
          class="p-1.5 rounded text-text-secondary hover:text-red-400 hover:bg-red-500/10 transition-colors shrink-0 cursor-pointer"
        >
          <LogOut class="w-4 h-4" />
        </button>
      </div>
      <p class="mt-2 text-center text-[9px] font-mono text-text-muted tracking-tight">
        © Tim SANOC — UTB 2026.
      </p>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '../../stores/authStore';
import { useDeviceStore } from '../../stores/deviceStore';
import { useSettingStore } from '../../stores/settingStore';
import {
  LayoutGrid,
  Server,
  AlertTriangle,
  BarChart3,
  Settings,
  HelpCircle,
  LogOut,
  ShieldCheck,
  Eye,
  Cpu
} from 'lucide-vue-next';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const deviceStore = useDeviceStore();
const settingStore = useSettingStore();

onMounted(() => {
  deviceStore.fetchSummary();
});

function navLinkClass(path: string) {
  const isActive = route.path === path || (route.path.startsWith(path) && path !== '/dashboard');
  return isActive
    ? 'bg-brand-periwinkle/15 text-brand-periwinkle border border-brand-periwinkle/30 font-semibold shadow-sm shadow-brand-periwinkle/10'
    : 'text-text-secondary hover:text-text-main hover:bg-surface border border-transparent';
}

// Role display metadata
const roleLabel = computed(() => {
  const r = authStore.user.role;
  if (r === 'admin') return 'ADMIN';
  if (r === 'pimpinan') return 'LEADERSHIP';
  return 'OPERATOR';
});

const roleColor = computed(() => {
  const map: Record<string, string> = {
    admin: 'text-brand-periwinkle',
    pimpinan: 'text-amber-400',
    anggota: 'text-status-up'
  };
  return map[authStore.user.role] ?? 'text-text-secondary';
});

const roleIcon = computed(() => {
  const map: Record<string, any> = {
    admin: ShieldCheck,
    pimpinan: Eye,
    anggota: Cpu
  };
  return map[authStore.user.role] ?? ShieldCheck;
});

function handleLogout() {
  authStore.logout();
  router.push('/login');
}
</script>
