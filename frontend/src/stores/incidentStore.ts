import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Incident } from '../types';
import { incidentsApi } from '../api';

export const useIncidentStore = defineStore('incidents', () => {
  const incidents = ref<Incident[]>([]);
  const currentIncident = ref<Incident | null>(null);
  const isLoading = ref(false);
  const totalCount = ref<number>(0);

  async function fetchIncidents(params?: { page?: number; pageSize?: number; status?: string; search?: string }) {
    isLoading.value = true;
    try {
      const queryParams: any = {
        status: params?.status,
        search: params?.search
      };
      if (params && params.page !== undefined) {
        queryParams.page = params.page;
        queryParams.page_size = params.pageSize || 10;
      }
      const res = await incidentsApi.getIncidents(queryParams);
      if (res && (Array.isArray(res.items) || Array.isArray(res.data))) {
        incidents.value = res.items || res.data;
        totalCount.value = res.total || incidents.value.length;
      } else if (Array.isArray(res)) {
        incidents.value = res;
        totalCount.value = res.length;
      } else {
        incidents.value = [];
        totalCount.value = 0;
      }
    } catch (e) {
      if (incidents.value.length === 0) {
        incidents.value = generateInitialIncidents();
        totalCount.value = incidents.value.length;
      }
    } finally {
      isLoading.value = false;
    }
  }

  async function fetchIncidentById(id: string) {
    isLoading.value = true;
    try {
      currentIncident.value = await incidentsApi.getIncidentById(id);
    } catch (e) {
      const found = incidents.value.find((i: Incident) => i.id === id);
      if (found) {
        currentIncident.value = found;
      } else {
        throw e;
      }
    } finally {
      isLoading.value = false;
    }
  }

  return {
    incidents,
    currentIncident,
    isLoading,
    totalCount,
    fetchIncidents,
    fetchIncidentById
  };
});

function generateInitialIncidents(): Incident[] {
  return [
    {
      id: 'INC-2026-089',
      deviceId: 'dev-5',
      deviceName: 'AP Biro Umum Lt 2 Conference',
      deviceType: 'Access Point',
      deviceIp: '10.20.1.18',
      status: 'ACTIVE',
      startTime: '14 min ago',
      duration: '00:14:22',
      affectedDevicesCount: 14,
      packetLoss: 100,
      latencyMs: 0,
      dependenciesCount: 0,
      timeline: [
        {
          id: 't-1',
          timestamp: '14:02:10 WIB',
          title: 'Polling Check Failed',
          description: 'Check 1/3 failed — ICMP ping timeout for IP 10.20.1.18',
          severity: 'warning'
        },
        {
          id: 't-2',
          timestamp: '14:02:15 WIB',
          title: 'Polling Check Failed',
          description: 'Check 2/3 failed — ICMP ping timeout for IP 10.20.1.18',
          severity: 'warning'
        },
        {
          id: 't-3',
          timestamp: '14:02:20 WIB',
          title: 'Incident Created',
          description: 'Check 3/3 failed — failure threshold reached, incident created',
          severity: 'critical'
        },
        {
          id: 't-4',
          timestamp: '14:02:22 WIB',
          title: 'Aggregation Phase',
          description: 'Single device alert in Gedung Sate Lt 2 — sent individually',
          severity: 'info'
        },
        {
          id: 't-5',
          timestamp: '14:02:25 WIB',
          title: 'Rate Limit Check',
          description: 'Rate limit OK — dispatching notification now',
          severity: 'info'
        },
        {
          id: 't-6',
          timestamp: '14:02:28 WIB',
          title: 'Attempting Notification (WhatsApp)',
          description: 'Attempting WhatsApp notification to NOC On-Call Group...',
          severity: 'info',
          channel: 'WhatsApp'
        },
        {
          id: 't-7',
          timestamp: '14:02:30 WIB',
          title: '✅ Notification Delivered (WhatsApp)',
          description: 'WhatsApp delivered successfully to +6281290008888',
          severity: 'info',
          channel: 'WhatsApp'
        },
        {
          id: 't-8',
          timestamp: '14:02:31 WIB',
          title: '⏭️ Telegram Skipped (Not Needed)',
          description: 'Telegram skipped — WhatsApp delivered successfully to +6281290008888 (no fallback needed)',
          severity: 'skipped',
          channel: 'Telegram'
        }
      ],
      notificationLog: [
        { id: 'nl-1', channel: 'WhatsApp API', channelIcon: 'message-square', recipient: 'NOC On-Call (+6281290008888)', status: 'Failed', timestamp: '14:02:35' },
        { id: 'nl-2', channel: 'Telegram Bot', channelIcon: 'send', recipient: 'NOC Telegram Channel (-10019827364)', status: 'Delivered', timestamp: '14:02:37' },
        { id: 'nl-3', channel: 'Email Gateway', channelIcon: 'mail', recipient: 'noc.alerts@jabarprov.go.id', status: 'Sent', timestamp: '14:02:38' }
      ]
    },
    {
      id: 'INC-2026-088',
      deviceId: 'dev-3',
      deviceName: 'Distribution Switch Core-02',
      deviceType: 'Switch',
      deviceIp: '10.20.0.3',
      status: 'ACTIVE',
      startTime: '42 min ago',
      duration: '00:42:10',
      affectedDevicesCount: 28,
      packetLoss: 100,
      latencyMs: 0,
      dependenciesCount: 0,
      timeline: [
        { id: 't-10', timestamp: '13:34:00 WIB', title: 'Incident Detected', description: 'Switch unreachable on ping & SNMP query timeout', severity: 'critical' }
      ],
      notificationLog: [
        { id: 'nl-10', channel: 'Telegram Bot', channelIcon: 'send', recipient: 'NOC Telegram Channel', status: 'Delivered', timestamp: '13:34:05' }
      ]
    }
  ];
}
