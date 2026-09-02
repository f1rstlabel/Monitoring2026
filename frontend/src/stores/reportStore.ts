import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { ReportRow, FlapDevice } from '../types';

import api from '../api/client';
import { downloadCSV, downloadXLS } from '../utils/exportUtils';

export const useReportStore = defineStore('reports', () => {
  const period = ref<'daily' | 'weekly' | 'monthly' | 'custom'>('monthly');
  const startDate = ref<string>('');
  const endDate = ref<string>('');
  const locationFilter = ref<string>('All');
  const deviceTypeFilter = ref<string>('All');
  const isLoading = ref(false);
  const rows = ref<ReportRow[]>([]);
  const flapDevices = ref<FlapDevice[]>([]);

  // Stat card values
  const avgSlaUptime = ref(99.82);
  const totalOutageEvents = ref(0);
  const avgMttrMinutes = ref(8.75);
  const alertDeliveryRate = ref(99.4);

  const filteredRows = computed(() => {
    return rows.value
      .filter(r => (r.downCount || 0) > 0)
      .filter(r => locationFilter.value === 'All' || (r.location && r.location.includes(locationFilter.value)))
      .filter(r => deviceTypeFilter.value === 'All' || r.deviceType === deviceTypeFilter.value)
      .sort((a, b) => b.downCount - a.downCount);
  });

  const filteredFlapDevices = computed(() => {
    return flapDevices.value
      .filter(d => locationFilter.value === 'All' || (d.location && d.location.includes(locationFilter.value)))
      .filter(d => deviceTypeFilter.value === 'All' || d.deviceType === deviceTypeFilter.value);
  });

  // Filtered stats for header cards — must respect location/type filters (bug fix: previously used unfiltered totals)
  const filteredTotalOutageEvents = computed(() => filteredRows.value.reduce((acc, r) => acc + (r.downCount || 0), 0));
  const filteredAvgSlaUptime = computed(() => {
    if (filteredRows.value.length === 0) return avgSlaUptime.value;
    const totalDT = filteredRows.value.reduce((acc, r) => acc + (r.totalDowntimeMinutes || 0), 0);
    const totalDevices = filteredRows.value.length || 1;
    let periodMinutes = 30 * 24 * 60;
    if (period.value === 'daily') periodMinutes = 24 * 60;
    else if (period.value === 'weekly') periodMinutes = 7 * 24 * 60;
    else if (period.value === 'custom' && startDate.value && endDate.value) {
      const diffMs = new Date(endDate.value).getTime() - new Date(startDate.value).getTime();
      periodMinutes = Math.max(60, Math.floor(diffMs / 60000));
    }
    const uptimePct = 100.0 - ((totalDT / (totalDevices * periodMinutes)) * 100.0);
    return Math.max(0.0, Math.min(100.0, Number(uptimePct.toFixed(2)))) || 100.0;
  });

  const uniqueLocations = computed(() => {
    const locs = new Set(rows.value.map(r => r.location).filter(Boolean));
    return ['All', ...Array.from(locs)];
  });

  function formatDowntime(minutes: number): string {
    const h = Math.floor(minutes / 60);
    const m = minutes % 60;
    if (h > 0) return `${h}h ${m}m`;
    return `${m}m`;
  }

  async function fetchReport() {
    isLoading.value = true;
    try {
      const params: any = { period: period.value, _t: Date.now() };
      if (period.value === 'custom') {
        params.startDate = startDate.value;
        params.endDate = endDate.value;
      }
      const [resReport, resFlap] = await Promise.allSettled([
        api.get('/reports', { params }),
        api.get('/reports/flap-devices', { params })
      ]);

      if (resReport.status === 'fulfilled') {
        rows.value = Array.isArray(resReport.value.data) ? resReport.value.data : [];
        if (rows.value.length > 0) {
          const totalDown = rows.value.reduce((acc, r) => acc + (r.downCount || 0), 0);
          totalOutageEvents.value = totalDown;
          const totalDT = rows.value.reduce((acc, r) => acc + (r.totalDowntimeMinutes || 0), 0);
          const totalDevices = rows.value.length || 1;
          let periodMinutes = 30 * 24 * 60;
          if (period.value === 'daily') periodMinutes = 24 * 60;
          else if (period.value === 'weekly') periodMinutes = 7 * 24 * 60;
          else if (period.value === 'custom' && startDate.value && endDate.value) {
            const diffMs = new Date(endDate.value).getTime() - new Date(startDate.value).getTime();
            periodMinutes = Math.max(60, Math.floor(diffMs / 60000));
          }
          const uptimePct = 100.0 - ((totalDT / (totalDevices * periodMinutes)) * 100.0);
          avgSlaUptime.value = Math.max(0.0, Math.min(100.0, Number(uptimePct.toFixed(2)))) || 100.0;
        }
      }

      if (resFlap.status === 'fulfilled') {
        flapDevices.value = Array.isArray(resFlap.value.data) ? resFlap.value.data : [];
      }
    } catch (e) {
      console.error('Failed to fetch real report data:', e);
    } finally {
      isLoading.value = false;
    }
  }

  function exportData(
    format: 'csv' | 'xls' = 'xls',
    activeTab: 'downtime' | 'recurring' | 'active_incidents' = 'downtime',
    incidentsData: any[] = []
  ) {
    let headers: string[] = [];
    let rowsData: (string | number)[][] = [];
    let reportTitle = '';
    let filePrefix = 'sanoc-report';

    if (activeTab === 'recurring') {
      reportTitle = `Laporan Audit SLA — Recurring Issues (Flapping Devices)`;
      filePrefix = `recurring-issues-${period.value}`;
      headers = ['#', 'Nama Perangkat', 'Tipe', 'Lokasi', 'Jumlah Down (Periode)'];
      rowsData = filteredFlapDevices.value.map((d, idx) => [
        idx + 1,
        d.deviceName,
        d.deviceType,
        d.location,
        `${d.downCount7d}x`
      ]);
    } else if (activeTab === 'active_incidents') {
      reportTitle = `Laporan Audit SLA — Antrean Active Incidents`;
      filePrefix = `active-incidents-${period.value}`;
      headers = ['Ticket ID', 'Nama Perangkat', 'Tipe', 'IP Address', 'Durasi', 'Jumlah Terdampak', 'Status'];
      rowsData = incidentsData.map((inc) => [
        inc.id,
        inc.deviceName,
        inc.deviceType,
        inc.deviceIp,
        inc.duration || '0m',
        inc.affectedDevicesCount || 1,
        inc.status
      ]);
    } else {
      reportTitle = `Laporan Audit SLA — Downtime per Perangkat`;
      filePrefix = `downtime-by-device-${period.value}`;
      headers = ['#', 'Nama Perangkat', 'Tipe', 'Lokasi', 'Jumlah Downtime', 'Total Durasi Downtime', 'Terakhir Down'];
      rowsData = filteredRows.value.map((r, idx) => [
        idx + 1,
        r.deviceName,
        r.deviceType,
        r.location,
        `${r.downCount}x`,
        formatDowntime(r.totalDowntimeMinutes),
        r.lastDown || '-'
      ]);
    }

    const dateStr = new Date().toISOString().slice(0, 10);
    const typePrefix = format === 'csv' ? 'sanoc-csv' : 'sanoc-excel';
    const filename = `${typePrefix}-${filePrefix}-${dateStr}`;

    if (format === 'csv') {
      downloadCSV(filename, headers, rowsData);
    } else {
      downloadXLS(filename, reportTitle, headers, rowsData);
    }
  }

  function exportCSV() {
    exportData('csv', 'downtime');
  }

  return {
    period,
    startDate,
    endDate,
    locationFilter,
    deviceTypeFilter,
    isLoading,
    rows,
    flapDevices,
    avgSlaUptime,
    totalOutageEvents,
    avgMttrMinutes,
    alertDeliveryRate,
    filteredRows,
    filteredFlapDevices,
    filteredTotalOutageEvents,
    filteredAvgSlaUptime,
    uniqueLocations,
    formatDowntime,
    fetchReport,
    exportData,
    exportCSV,
    exportPDF() {
      window.print();
    }
  };
});
