<template>
  <aside class="w-60 bg-[#0D0D0F] border-r border-[#26262A] flex flex-col justify-between h-screen fixed left-0 top-0 z-30 select-none">
    <!-- Top Branding & Navigation -->
    <div>
      <!-- Header / Logo -->
      <div class="px-5 py-5 border-b border-[#26262A]/60 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-lg bg-[#7B96F5]/15 border border-[#7B96F5]/30 flex items-center justify-center text-[#7B96F5] shadow-sm shadow-[#7B96F5]/10">
            <Shield class="w-5 h-5" />
          </div>
          <div>
            <h1 class="text-sm font-bold text-white tracking-tight leading-none">GovMonitor <span class="text-[#7B96F5]">IT</span></h1>
            <p class="text-[10px] font-mono text-gray-400 mt-1 uppercase tracking-wider">Jabar Regional NOC</p>
          </div>
        </div>
        <span class="text-[10px] font-mono bg-[#18181B] text-[#7B96F5] px-1.5 py-0.5 rounded border border-[#26262A]">v2.4.1</span>
      </div>

      <!-- Navigation Items — rendered via v-if per role, not v-show -->
      <nav class="p-3 space-y-1 mt-2">
        <!-- Dashboard: all roles -->
        <router-link
          to="/dashboard"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-medium transition-all duration-150"
          :class="navLinkClass('/dashboard')"
        >
          <LayoutGrid class="w-4 h-4 shrink-0" />
          <span>Dashboard</span>
        </router-link>

        <!-- Devices: all roles -->
        <router-link
          to="/devices"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-medium transition-all duration-150"
          :class="navLinkClass('/devices')"
        >
          <Server class="w-4 h-4 shrink-0" />
          <span>Devices</span>
          <span class="ml-auto text-[10px] font-bold px-1.5 py-0.5 rounded-full bg-[#26262A] text-gray-300">
            {{ deviceStore.summary.totalDevices }}
          </span>
        </router-link>

        <!-- Incidents: all roles -->
        <router-link
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

        <!-- Reports: all roles -->
        <router-link
          to="/reports"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-medium transition-all duration-150"
          :class="navLinkClass('/reports')"
        >
          <BarChart3 class="w-4 h-4 shrink-0" />
          <span>Reports</span>
        </router-link>

        <!-- Settings: superadmin + anggota ONLY — removed from render tree for pimpinan -->
        <router-link
          v-if="authStore.canSeeSettings"
          to="/settings"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-medium transition-all duration-150"
          :class="navLinkClass('/settings')"
        >
          <Settings class="w-4 h-4 shrink-0" />
          <span>Settings</span>
        </router-link>
      </nav>

      <!-- Role indicator badge -->
      <div class="px-4 mt-1">
        <div class="flex items-center gap-2 px-2 py-1.5 rounded-lg bg-[#0A0A0B] border border-[#26262A]/50">
          <component :is="roleIcon" class="w-3 h-3 shrink-0" :class="roleColor" />
          <span class="text-[10px] font-mono uppercase tracking-widest" :class="roleColor">
            {{ roleLabel }}
          </span>
          <span class="ml-auto inline-block w-1.5 h-1.5 rounded-full bg-[#3ECF8E]" />
        </div>
      </div>
    </div>

    <!-- Bottom User Section -->
    <div class="p-3 border-t border-[#26262A]/60 bg-[#0A0A0B]/50">
      <div class="flex items-center justify-between p-2 rounded-lg bg-[#151517] border border-[#26262A]">
        <div class="flex items-center gap-2.5 overflow-hidden">
          <img
            :src="authStore.user.avatarUrl"
            alt="User Avatar"
            class="w-8 h-8 rounded-full object-cover border border-[#7B96F5]/40 shrink-0"
          />
          <div class="overflow-hidden">
            <p class="text-xs font-medium text-gray-200 truncate leading-tight">{{ authStore.user.name }}</p>
            <p class="text-[10px] text-gray-500 truncate mt-0.5">{{ authStore.user.email }}</p>
          </div>
        </div>
        <button
          @click="handleLogout"
          title="Sign Out"
          class="p-1.5 rounded text-gray-400 hover:text-red-400 hover:bg-red-500/10 transition-colors shrink-0"
        >
          <LogOut class="w-4 h-4" />
        </button>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '../../stores/authStore';
import { useDeviceStore } from '../../stores/deviceStore';
import {
  Shield,
  LayoutGrid,
  Server,
  AlertTriangle,
  BarChart3,
  Settings,
  LogOut,
  ShieldCheck,
  Eye,
  Cpu
} from 'lucide-vue-next';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const deviceStore = useDeviceStore();

function navLinkClass(path: string) {
  const isActive = route.path === path || (route.path.startsWith(path) && path !== '/dashboard');
  return isActive
    ? 'bg-[#7B96F5]/15 text-[#7B96F5] border border-[#7B96F5]/30 font-semibold shadow-sm shadow-[#7B96F5]/10'
    : 'text-gray-400 hover:text-gray-200 hover:bg-[#151517] border border-transparent';
}

// Role display metadata
const roleLabel = computed(() => {
  const map: Record<string, string> = {
    superadmin: 'Super Admin',
    pimpinan: 'Pimpinan',
    anggota: 'Anggota NOC'
  };
  return map[authStore.user.role] ?? authStore.user.role;
});

const roleColor = computed(() => {
  const map: Record<string, string> = {
    superadmin: 'text-[#7B96F5]',
    pimpinan: 'text-amber-400',
    anggota: 'text-[#34D399]'
  };
  return map[authStore.user.role] ?? 'text-gray-400';
});

const roleIcon = computed(() => {
  const map: Record<string, any> = {
    superadmin: ShieldCheck,
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
