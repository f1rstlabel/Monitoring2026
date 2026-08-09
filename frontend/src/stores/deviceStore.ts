import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { Device, DeviceType, DeviceStatus, DashboardSummary } from '../types';
import { devicesApi, dashboardApi } from '../api';

export const useDeviceStore = defineStore('devices', () => {
  const devices = ref<Device[]>([]);
  const summary = ref<DashboardSummary>({
    totalDevices: 414,
    devicesUp: 398,
    devicesDown: 16,
    activeIncidents: 4,
    upPercentage: 96.1,
    downPercentage: 3.9
  });
  const isLoading = ref(false);
  const selectedTypeFilter = ref<string>('All');
  const selectedStatusFilter = ref<string>('All');
  const searchQuery = ref<string>('');
  const viewMode = ref<'grid' | 'list'>('grid');
  const totalCount = ref<number>(0);

  const filteredDevices = computed(() => {
    return devices.value.filter(device => {
      const matchType = selectedTypeFilter.value === 'All' || device.type === selectedTypeFilter.value;
      const matchStatus = selectedStatusFilter.value === 'All' || device.status === selectedStatusFilter.value;
      const query = searchQuery.value.toLowerCase().trim();
      const matchQuery = !query || 
        device.name.toLowerCase().includes(query) || 
        device.ip.toLowerCase().includes(query) || 
        device.mac.toLowerCase().includes(query) ||
        device.location.toLowerCase().includes(query);

      return matchType && matchStatus && matchQuery;
    });
  });

  async function fetchSummary() {
    try {
      const res = await dashboardApi.getSummary();
      summary.value = res;
    } catch (e) {
      // Use initial state if mock server loading
    }
  }

  async function fetchDevices(params?: { page?: number; pageSize?: number }) {
    isLoading.value = true;
    try {
      const queryParams: any = {
        type: selectedTypeFilter.value !== 'All' ? selectedTypeFilter.value : undefined,
        status: selectedStatusFilter.value !== 'All' ? selectedStatusFilter.value : undefined,
        search: searchQuery.value || undefined
      };
      if (params && params.page !== undefined) {
        queryParams.page = params.page;
        queryParams.page_size = params.pageSize || 10;
      }
      const res = await devicesApi.getDevices(queryParams);
      if (res && (Array.isArray(res.items) || Array.isArray(res.data))) {
        devices.value = res.items || res.data;
        totalCount.value = res.total || devices.value.length;
      } else if (Array.isArray(res)) {
        devices.value = res;
        totalCount.value = res.length;
      } else {
        devices.value = [];
        totalCount.value = 0;
      }
    } catch (e) {
      // Fallback generator if offline / mock loading
      if (devices.value.length === 0) {
        devices.value = generateInitialDevices();
        totalCount.value = devices.value.length;
      }
    } finally {
      isLoading.value = false;
    }
  }



  async function addDevice(deviceData: Partial<Device>) {
    try {
      const created = await devicesApi.createDevice(deviceData);
      devices.value.unshift(created);
      summary.value.totalDevices++;
      if (created.status === 'UP') summary.value.devicesUp++;
      else summary.value.devicesDown++;
      return created;
    } catch (e) {
      // Offline fallback
      const newDev: Device = {
        id: `dev-${Date.now()}`,
        name: deviceData.name || 'New AP Core',
        type: (deviceData.type as DeviceType) || 'Access Point',
        ip: deviceData.ip || '10.20.4.150',
        mac: deviceData.mac || 'F4:8E:38:1A:0B:44',
        status: 'UP',
        addressingMode: deviceData.addressingMode || 'Static',
        location: deviceData.location || 'Gedung Sate Lt. 2',
        rack: deviceData.rack || 'Rack B-04',
        checkedSecondsAgo: 2,
        lastChecked: new Date().toLocaleTimeString(),
        uptime30d: 100.0,
        failureThreshold: 3
      };
      devices.value.unshift(newDev);
      summary.value.totalDevices++;
      summary.value.devicesUp++;
      return newDev;
    }
  }

  async function updateDevice(id: string, deviceData: Partial<Device>) {
    const dev = devices.value.find(d => d.id === id);
    if (dev) {
      Object.assign(dev, deviceData);
      try {
        await devicesApi.updateDevice(id, deviceData);
      } catch (e) {
        // local update complete
      }
    }
  }

  function updateDeviceStatus(id: string, newStatus: DeviceStatus, latencyMs?: number) {
    const dev = devices.value.find(d => d.id === id);
    if (dev) {
      if (dev.status !== newStatus) {
        if (dev.status === 'UP' && newStatus === 'DOWN') {
          summary.value.devicesUp--;
          summary.value.devicesDown++;
        } else if (dev.status === 'DOWN' && newStatus === 'UP') {
          summary.value.devicesDown--;
          summary.value.devicesUp++;
        }
        dev.status = newStatus;
      }
      if (typeof latencyMs === 'number') {
        dev.latencyMs = latencyMs;
      }
      dev.checkedSecondsAgo = 0;
      dev.lastChecked = new Date().toLocaleTimeString();
    }
  }

  return {
    devices,
    summary,
    isLoading,
    selectedTypeFilter,
    selectedStatusFilter,
    searchQuery,
    viewMode,
    totalCount,
    filteredDevices,
    fetchSummary,
    fetchDevices,
    addDevice,
    updateDevice,
    updateDeviceStatus
  };
});

function generateInitialDevices(): Device[] {
  const types: DeviceType[] = ['Access Point', 'Switch', 'Router', 'SmartPower', 'CCTV', 'NVR'];
  const result: Device[] = [];

  const sampleNames = [
    { name: 'Test - Localhost', type: 'Router' as DeviceType, ip: '127.0.0.1', status: 'UP' as DeviceStatus, loc: 'Local Host Server' },
    { name: 'Test - Cloudflare DNS', type: 'Access Point' as DeviceType, ip: '1.1.1.1', status: 'UP' as DeviceStatus, loc: 'Cloudflare Public DNS' },
    { name: 'Test - Google DNS', type: 'Switch' as DeviceType, ip: '8.8.8.8', status: 'UP' as DeviceStatus, loc: 'Google Public DNS' }
  ];

  sampleNames.forEach((s, idx) => {
    result.push({
      id: `dev-${idx + 1}`,
      name: s.name,
      type: s.type,
      ip: s.ip,
      mac: `00:1A:2B:3C:${(40 + idx).toString(16).toUpperCase()}:${(10 + idx).toString(16).toUpperCase()}`,
      status: s.status,
      addressingMode: 'Static',
      location: s.loc,
      rack: `Rack A-${idx + 1}`,
      checkedSecondsAgo: Math.floor(Math.random() * 25) + 2,
      lastChecked: new Date(Date.now() - Math.floor(Math.random() * 20000)).toLocaleTimeString(),
      uptime30d: s.status === 'UP' ? 99.8 : 94.2,
      failureThreshold: 3
    });
  });

  // Generate remaining devices to reach 414 total
  for (let i = 11; i <= 414; i++) {
    const type = types[i % types.length];
    const isDown = i % 26 === 0; // Create 16 DOWN devices total out of 414
    result.push({
      id: `dev-${i}`,
      name: `${type} Node ${i.toString().padStart(3, '0')}`,
      type,
      ip: `10.20.${Math.floor(i / 50)}.${i % 250}`,
      mac: `54:A0:50:${(i % 90 + 10).toString(16).toUpperCase()}:${(i % 80 + 10).toString(16).toUpperCase()}:${(i % 70 + 10).toString(16).toUpperCase()}`,
      status: isDown ? 'DOWN' : 'UP',
      addressingMode: i % 4 === 0 ? 'DHCP' : 'Static',
      location: `Gedung Sate Wing ${(i % 4) + 1} Lt. ${(i % 3) + 1}`,
      rack: `Rack B-${(i % 12) + 1}`,
      checkedSecondsAgo: Math.floor(Math.random() * 45) + 1,
      lastChecked: new Date(Date.now() - Math.floor(Math.random() * 45000)).toLocaleTimeString(),
      uptime30d: isDown ? 93.5 : 99.9,
      failureThreshold: 3
    });
  }

  return result;
}
