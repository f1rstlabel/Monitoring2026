<template>
  <div v-if="isLoginPage" class="min-h-screen bg-[var(--bg-main)] text-[var(--text-primary)]">
    <router-view />
  </div>

  <div v-else class="min-h-screen overflow-x-hidden bg-[var(--bg-main)] text-[var(--text-primary)] flex">
    <!-- Sidebar Navigation -->
    <Sidebar :is-mobile-open="isSidebarOpen" @close="isSidebarOpen = false" />

    <!-- Main Content Area -->
    <div class="flex-1 min-w-0 min-h-screen flex flex-col lg:ml-60">
      <!-- Topbar Header -->
      <Topbar @toggle-sidebar="isSidebarOpen = !isSidebarOpen" />

      <!-- Page Container -->
      <main class="p-3 sm:p-4 lg:p-6 flex-1 min-w-0 bg-[var(--bg-main)]">
        <router-view />
      </main>

      <!-- AI Copilot Floating Drawer -->
      <AICopilotDrawer />
    </div>
  </div>
  
  <!-- Global Toast Notifications -->
  <ToastContainer />
  
  <!-- Cookie Consent Banner -->
  <CookieConsent />
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import Sidebar from './components/common/Sidebar.vue';
import Topbar from './components/common/Topbar.vue';
import AICopilotDrawer from './components/ai/AICopilotDrawer.vue';
import CookieConsent from './components/common/CookieConsent.vue';
import ToastContainer from './components/common/ToastContainer.vue';
import { useLiveStore } from './stores/liveStore';
import { useThemeStore } from './stores/themeStore';
import { useSettingStore } from './stores/settingStore';
import { useDeviceStore } from './stores/deviceStore';

const route = useRoute();
const liveStore = useLiveStore();
const themeStore = useThemeStore();
const settingStore = useSettingStore();
const deviceStore = useDeviceStore();
const isSidebarOpen = ref(false);

const isLoginPage = computed(() => route.path === '/login');

watch(() => route.path, () => {
  isSidebarOpen.value = false;
});

onMounted(() => {
  themeStore.initTheme();
  liveStore.initWebSocket();
  settingStore.fetchBranding();
  deviceStore.fetchSummary();
});
</script>
