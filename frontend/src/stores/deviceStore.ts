import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { Device, DeviceStatus, DashboardSummary } from '../types';
import { devicesApi, dashboardApi } from '../api';

export const useDeviceStore = defineStore('devices', () => {
  const devices = ref<Device[]>([]);
  const summary = ref<DashboardSummary>({
    totalDevices: 0,
    devicesUp: 0,
    devicesDown: 0,
    activeIncidents: 0,
    upPercentage: 0,
    downPercentage: 0
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
      // Keep state as is on error
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
      devices.value = [];
      totalCount.value = 0;
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
      throw e;
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
