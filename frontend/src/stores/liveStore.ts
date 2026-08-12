import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { LiveFeedItem } from '../types';
import { wsClient } from '../ws/websocket';
import { useDeviceStore } from './deviceStore';
import { useAuthStore } from './authStore';

let isInitialized = false;
let unsubscribeFn: (() => void) | null = null;

export const useLiveStore = defineStore('live', () => {
  const liveFeed = ref<LiveFeedItem[]>([]);
  const batchStatusMessage = ref<string>('');
  const isConnected = wsClient.isConnected;
  const lastUpdatedAgo = wsClient.lastUpdatedAgo;

  function formatWIBTime(date: Date = new Date()): string {
    return date.toLocaleTimeString('id-ID', {
      timeZone: 'Asia/Jakarta',
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    }) + ' WIB';
  }

  function handleWSMessage(data: any) {
    const deviceStore = useDeviceStore();

    if (data.type === 'STATUS_CHANGE') {
      deviceStore.updateDeviceStatus(data.deviceId, data.status, data.latencyMs);
    }

    if (data.type === 'STATUS_CHANGE' || data.type === 'DASHBOARD_UPDATE' || data.type === 'INCIDENT_ALERT') {
      deviceStore.fetchSummary();
    }

    if (data.type === 'ROLE_PERMISSIONS_UPDATED') {
      const authStore = useAuthStore();
      authStore.fetchMe();
    }

    if (data.type === 'BATCH_PROGRESS') {
      batchStatusMessage.value = data.title || '';
      let formattedTs = formatWIBTime();
      pushLiveFeed({
        id: `batch-${Date.now()}-${Math.random().toString(36).substring(2, 5)}`,
        timestamp: formattedTs,
        title: data.title || 'Batch Scanning Progress',
        description: data.description || 'Scanning inventory...',
        severity: 'info'
      });
    }

    if (data.type === 'LIVE_FEED' || data.type === 'STATUS_CHANGE' || data.type === 'INCIDENT_ALERT') {
      let formattedTs = formatWIBTime();
      if (data.timestamp) {
        try {
          const parsed = new Date(data.timestamp);
          if (!isNaN(parsed.getTime())) {
            formattedTs = parsed.toLocaleTimeString('id-ID', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' }) + ' WIB';
          }
        } catch (e) {
          // fallback
        }
      }

      pushLiveFeed({
        id: `feed-${data.deviceId || 'sys'}-${Date.now()}-${Math.random().toString(36).substr(2, 4)}`,
        timestamp: formattedTs,
        title: data.title || (data.deviceName ? `${data.deviceName} — ${data.status}` : 'System Log'),
        description: data.description || `IP: ${data.ip || 'N/A'}`,
        severity: data.severity || (data.status === 'DOWN' ? 'critical' : 'info'),
        deviceId: data.deviceId
      });
    }
  }

  function initWebSocket() {
    wsClient.connect();
    if (isInitialized) return;
    isInitialized = true;

    if (unsubscribeFn) {
      unsubscribeFn();
    }
    unsubscribeFn = wsClient.subscribe(handleWSMessage);
  }

  function pushLiveFeed(item: LiveFeedItem) {
    liveFeed.value.unshift(item);
    if (liveFeed.value.length > 100) {
      liveFeed.value.pop();
    }
  }

  return {
    liveFeed,
    batchStatusMessage,
    isConnected,
    lastUpdatedAgo,
    initWebSocket,
    pushLiveFeed
  };
});
