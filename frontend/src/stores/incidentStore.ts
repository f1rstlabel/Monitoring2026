import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Incident } from '../types';
import { incidentsApi } from '../api';

export const useIncidentStore = defineStore('incidents', () => {
  const incidents = ref<Incident[]>([]);
  const currentIncident = ref<Incident | null>(null);
  const isLoading = ref(false);

  async function fetchIncidents() {
    isLoading.value = true;
    try {
      incidents.value = await incidentsApi.getIncidents();
    } catch (e) {
      if (incidents.value.length === 0) {
        incidents.value = generateInitialIncidents();
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
        currentIncident.value = generateInitialIncidents()[0];
      }
    } finally {
      isLoading.value = false;
    }
  }

  return {
    incidents,
    currentIncident,
    isLoading,
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
          title: 'Incident Detected',
          description: 'ICMP Poller received 100% packet loss (0/5 replies received) for IP 10.20.1.18',
          severity: 'critical'
        },
        {
          id: 't-2',
          timestamp: '14:02:25 WIB',
          title: 'Debounce Threshold Reached',
          description: 'Consecutive ICMP check failure count hit threshold = 3. Marking status to DOWN.',
          severity: 'critical'
        },
        {
          id: 't-3',
          timestamp: '14:02:28 WIB',
          title: 'Dependency Check Executed',
          description: 'Parent device "Distribution Switch Core-01" (10.20.0.2) is UP. Isolated issue confirmed.',
          severity: 'info'
        },
        {
          id: 't-4',
          timestamp: '14:02:30 WIB',
          title: 'WhatsApp Alert Attempted',
          description: 'Sending alert template via WhatsApp Gateway to NOC On-Call Group (+62 812-9000-8888)...',
          severity: 'info',
          channel: 'WhatsApp API'
        },
        {
          id: 't-5',
          timestamp: '14:02:35 WIB',
          title: 'WhatsApp Delivery Failed',
          description: 'Error: Gateway timeout (504). WhatsApp session temporarily unresponsive.',
          severity: 'warning',
          channel: 'WhatsApp API'
        },
        {
          id: 't-6',
          timestamp: '14:02:36 WIB',
          title: 'Telegram Fallback Triggered',
          description: 'Automatic failover mechanism redirected dispatch to Telegram Bot (@GovMonitorJabarBot)...',
          severity: 'info',
          channel: 'Telegram Bot'
        },
        {
          id: 't-7',
          timestamp: '14:02:37 WIB',
          title: 'Telegram Alert Delivered',
          description: 'Message successfully delivered to NOC Channel ID -1001982736412.',
          severity: 'info',
          channel: 'Telegram Bot'
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
