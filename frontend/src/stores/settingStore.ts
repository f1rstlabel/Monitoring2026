import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { SystemSettings, User } from '../types';
import { settingsApi, usersApi } from '../api';

export const useSettingStore = defineStore('settings', () => {
  const settings = ref<SystemSettings>({
    channels: [
      { id: 'ch-1', name: 'WhatsApp API Gateway', type: 'whatsapp', connected: true, handleOrNumber: '+62 812-9000-8888 (NOC Gateway)', lastSync: '2 min ago' },
      { id: 'ch-2', name: 'Telegram Emergency Bot', type: 'telegram', connected: true, handleOrNumber: '@GovMonitorJabarBot', lastSync: 'Just now' }
    ],
    rateLimitMaxMsgPerMin: 60,
    thresholds: [
      { type: 'Access Point', consecutiveFailures: 3 },
      { type: 'Switch', consecutiveFailures: 2 },
      { type: 'Router', consecutiveFailures: 2 },
      { type: 'SmartPower', consecutiveFailures: 4 },
      { type: 'CCTV', consecutiveFailures: 5 },
      { type: 'NVR', consecutiveFailures: 3 }
    ],
    polling: {
      intervalSeconds: 15,
      concurrencyBatchSize: 50
    }
  });

  const users = ref<User[]>([
    { id: 'u-1', name: 'Budi Santoso', email: 'budi.santoso@jabarprov.go.id', role: 'superadmin', status: 'Active', avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256', lastActive: 'Active Now' },
    { id: 'u-2', name: 'Rian Pratama', email: 'rian.pratama@jabarprov.go.id', role: 'anggota', status: 'Active', avatarUrl: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=256', lastActive: '12 min ago' },
    { id: 'u-3', name: 'Sari Dewi', email: 'sari.dewi@jabarprov.go.id', role: 'pimpinan', status: 'Active', avatarUrl: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=256', lastActive: '1 hr ago' }
  ]);

  const isLoading = ref(false);

  async function fetchSettings() {
    try {
      const res = await settingsApi.getSettings();
      if (res) settings.value = res;
    } catch (e) {
      // offline default
    }
  }

  async function fetchUsers() {
    try {
      const res = await usersApi.getUsers();
      if (res && res.length) users.value = res;
    } catch (e) {
      // offline default
    }
  }

  async function updateThreshold(type: string, val: number) {
    const item = settings.value.thresholds.find(t => t.type === type);
    if (item) {
      item.consecutiveFailures = val;
    }
    try {
      await settingsApi.updateThresholds(settings.value.thresholds);
    } catch (e) {
      // offline fallback
    }
  }

  async function updatePolling(intervalSec: number, batchSize: number) {
    settings.value.polling.intervalSeconds = intervalSec;
    settings.value.polling.concurrencyBatchSize = batchSize;
    try {
      await settingsApi.updatePolling({ intervalSeconds: intervalSec, concurrencyBatchSize: batchSize });
    } catch (e) {
      // offline
    }
  }

  async function updateRateLimit(val: number) {
    settings.value.rateLimitMaxMsgPerMin = val;
    try {
      await settingsApi.updateRateLimit(val);
    } catch (e) {
      // offline
    }
  }

  async function inviteUser(name: string, email: string, role: string) {
    const newUser: User = {
      id: `u-${Date.now()}`,
      name,
      email,
      role: role as any,
      status: 'Active',
      avatarUrl: `https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&q=80&w=256`,
      lastActive: 'Invited'
    };
    users.value.push(newUser);
    try {
      await usersApi.inviteUser({ name, email, role });
    } catch (e) {
      // offline
    }
    return newUser;
  }

  async function saveSettings() {
    try {
      await settingsApi.updateThresholds(settings.value.thresholds);
      await settingsApi.updatePolling(settings.value.polling);
      await settingsApi.updateRateLimit(settings.value.rateLimitMaxMsgPerMin);
    } catch (e) {
      // offline
    }
  }

  return {
    settings,
    users,
    isLoading,
    fetchSettings,
    fetchUsers,
    saveSettings,
    updateThreshold,
    updatePolling,
    updateRateLimit,
    inviteUser
  };
});
