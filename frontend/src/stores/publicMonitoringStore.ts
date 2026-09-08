import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import type {
  PublicMonitor,
  PublicMonitorDetail,
  PublicMonitorCheck,
  PublicMonitorStatus,
  PublicMonitorIncident,
  PublicMonitorIncidentDetail,
  PublicMonitorReport,
  PublicMonitorGroup
} from '../types';
import { publicMonitoringApi } from '../api';

export const usePublicMonitoringStore = defineStore('publicMonitoring', () => {
  const monitors = ref<PublicMonitor[]>([]);
  const archivedMonitors = ref<PublicMonitor[]>([]);
  const archivedLoading = ref(false);
  const archivedTotal = ref(0);
  const archivedTotalPages = ref(1);
  const selectedMonitor = ref<PublicMonitorDetail | null>(null);
  const checks = ref<PublicMonitorCheck[]>([]);
  const checksLoading = ref(false);
  const checksTotal = ref(0);
  const checksPage = ref(1);
  const checksPageSize = ref(10);
  const checksTotalPages = ref(1);
  const isLoading = ref(false);
  const isDetailLoading = ref(false);
  const error = ref('');
  const search = ref('');
  const statusFilter = ref<'ALL' | PublicMonitorStatus>('ALL');
  const groupFilter = ref('ALL');
  const page = ref(1);
  const pageSize = ref(100);
  const total = ref(0);
  const totalPages = ref(1);
  const incidents = ref<PublicMonitorIncident[]>([]);
  const selectedIncident = ref<PublicMonitorIncidentDetail | null>(null);
  const incidentsLoading = ref(false);
  const incidentsTotal = ref(0);
  const incidentsTotalPages = ref(1);
  const report = ref<PublicMonitorReport | null>(null);
  const reportLoading = ref(false);
  const groups = ref<PublicMonitorGroup[]>([]);
  const groupsLoading = ref(false);

  const summary = computed(() => ({
    total: total.value,
    up: monitors.value.filter(m => m.status === 'UP').length,
    down: monitors.value.filter(m => m.status === 'DOWN').length,
    paused: monitors.value.filter(m => m.status === 'PAUSED').length
  }));

  async function fetchMonitors() {
    isLoading.value = true;
    error.value = '';
    try {
      const response = await publicMonitoringApi.getMonitors({
        search: search.value || undefined,
        status: statusFilter.value === 'ALL' ? undefined : statusFilter.value,
        group: groupFilter.value === 'ALL' ? undefined : groupFilter.value,
        page: page.value,
        page_size: pageSize.value
      });
      monitors.value = response.items || response.data || [];
      total.value = response.total || 0;
      totalPages.value = response.totalPages || 1;
      if (selectedMonitor.value && !monitors.value.some(m => m.id === selectedMonitor.value?.monitor.id)) {
        selectedMonitor.value = null;
      }
    } catch (e: any) {
      error.value = e?.response?.data?.error || 'Unable to load public monitors';
    } finally {
      isLoading.value = false;
    }
  }

  async function fetchDetail(id: string) {
    isDetailLoading.value = true;
    try {
      selectedMonitor.value = await publicMonitoringApi.getMonitorById(id);
    } catch (e: any) {
      error.value = e?.response?.data?.error || 'Unable to load monitor details';
    } finally {
      isDetailLoading.value = false;
    }
  }

  async function fetchArchivedMonitors(searchValue = '') {
    archivedLoading.value = true;
    try {
      const response = await publicMonitoringApi.getArchivedMonitors({ search: searchValue || undefined, page: 1, page_size: 100 });
      archivedMonitors.value = response.items || response.data || [];
      archivedTotal.value = response.total || 0;
      archivedTotalPages.value = response.totalPages || 1;
    } catch (e: any) {
      error.value = e?.response?.data?.error || 'Unable to load archived monitors';
    } finally {
      archivedLoading.value = false;
    }
  }

  async function fetchChecks(id: string, page = 1, pageSize = 10) {
    checksLoading.value = true;
    try {
      const response = await publicMonitoringApi.getMonitorChecks(id, { page, page_size: pageSize });
      checks.value = response.items || response.data || [];
      checksTotal.value = response.total || 0;
      checksPage.value = response.page || page;
      checksPageSize.value = response.pageSize || pageSize;
      checksTotalPages.value = response.totalPages || 1;
    } catch (e: any) {
      error.value = e?.response?.data?.error || 'Unable to load monitor check history';
    } finally {
      checksLoading.value = false;
    }
  }

  async function createMonitor(payload: Partial<PublicMonitor>) {
    const created = await publicMonitoringApi.createMonitor(payload);
    page.value = 1;
    await fetchMonitors();
    await fetchDetail(created.id);
    return created;
  }

  async function updateMonitor(id: string, payload: Partial<PublicMonitor>) {
    const updated = await publicMonitoringApi.updateMonitor(id, payload);
    await fetchMonitors();
    await fetchDetail(updated.id);
    return updated;
  }

  async function deleteMonitor(id: string) {
    await publicMonitoringApi.deleteMonitor(id);
    selectedMonitor.value = null;
    await fetchMonitors();
  }

  async function archiveMonitor(id: string, reason = '') {
    await publicMonitoringApi.deleteMonitor(id, reason);
    selectedMonitor.value = null;
    await Promise.all([fetchMonitors(), fetchArchivedMonitors()]);
  }

  async function restoreMonitor(id: string) {
    const restored = await publicMonitoringApi.restoreMonitor(id);
    await Promise.all([fetchMonitors(), fetchArchivedMonitors()]);
    return restored;
  }

  async function purgeMonitor(id: string, reason: string, confirmation: string) {
    await publicMonitoringApi.purgeMonitor(id, reason, confirmation);
    selectedMonitor.value = null;
    await Promise.all([fetchMonitors(), fetchArchivedMonitors()]);
  }

  async function checkNow(id: string) {
    const result = await publicMonitoringApi.checkNow(id);
    await fetchMonitors();
    await fetchDetail(id);
    return result;
  }

  function applyLiveUpdate(update: { monitorId: string; status: string; timestamp: string; latencyMs: number; statusCode?: number; error?: string }) {
    const status = update.status as PublicMonitorStatus;
    const apply = (monitor: PublicMonitor) => {
      if (monitor.id !== update.monitorId) return monitor;
      return {
        ...monitor,
        status,
        lastChecked: update.timestamp,
        lastLatencyMs: update.latencyMs,
        lastStatusCode: update.statusCode || monitor.lastStatusCode,
        lastError: update.error || undefined,
        enabled: status !== 'PAUSED'
      };
    };
    monitors.value = monitors.value.map(apply);
    if (selectedMonitor.value?.monitor.id === update.monitorId) {
      selectedMonitor.value = {
        ...selectedMonitor.value,
        monitor: apply(selectedMonitor.value.monitor)
      };
    }
  }

  async function fetchIncidents(params: { monitorId?: string; status?: string; search?: string; period?: string; startDate?: string; endDate?: string; page?: number; page_size?: number } = {}) {
    incidentsLoading.value = true;
    try {
      const response = await publicMonitoringApi.getIncidents({ period: 'monthly', page: 1, page_size: 50, ...params });
      incidents.value = response.items || response.data || [];
      incidentsTotal.value = response.total || 0;
      incidentsTotalPages.value = response.totalPages || 1;
    } catch (e: any) {
      error.value = e?.response?.data?.error || 'Unable to load public monitor incidents';
    } finally {
      incidentsLoading.value = false;
    }
  }

  async function fetchIncidentDetail(id: string) {
    try {
      selectedIncident.value = await publicMonitoringApi.getIncidentById(id);
    } catch (e: any) {
      error.value = e?.response?.data?.error || 'Unable to load public incident details';
    }
  }

  async function fetchReport(params: { monitorId?: string; period?: string; startDate?: string; endDate?: string } = {}) {
    reportLoading.value = true;
    try {
      report.value = await publicMonitoringApi.getReport({ period: 'daily', ...params });
    } catch (e: any) {
      error.value = e?.response?.data?.error || 'Unable to load public monitoring report';
    } finally {
      reportLoading.value = false;
    }
  }

  async function fetchGroups() {
    groupsLoading.value = true;
    try {
      const response = await publicMonitoringApi.getGroups();
      groups.value = response.items || response.data || [];
    } catch (e: any) {
      error.value = e?.response?.data?.error || 'Unable to load public monitor groups';
    } finally {
      groupsLoading.value = false;
    }
  }

  async function createGroup(payload: { name: string; description?: string; displayOrder?: number; enabled?: boolean }) {
    const created = await publicMonitoringApi.createGroup(payload);
    await fetchGroups();
    return created;
  }

  async function updateGroup(id: string, payload: { name: string; description?: string; displayOrder?: number; enabled?: boolean }) {
    const updated = await publicMonitoringApi.updateGroup(id, payload);
    await fetchGroups();
    return updated;
  }

  async function deleteGroup(id: string) {
    await publicMonitoringApi.deleteGroup(id);
    await fetchGroups();
  }

  function resetFilters() {
    search.value = '';
    statusFilter.value = 'ALL';
    groupFilter.value = 'ALL';
    page.value = 1;
  }

  return {
    monitors,
    archivedMonitors,
    archivedLoading,
    archivedTotal,
    archivedTotalPages,
    selectedMonitor,
    checks,
    checksLoading,
    checksTotal,
    checksPage,
    checksPageSize,
    checksTotalPages,
    isLoading,
    isDetailLoading,
    error,
    search,
    statusFilter,
    groupFilter,
    page,
    pageSize,
    total,
    totalPages,
    summary,
    incidents,
    selectedIncident,
    incidentsLoading,
    incidentsTotal,
    incidentsTotalPages,
    report,
    reportLoading,
    groups,
    groupsLoading,
    fetchMonitors,
    fetchDetail,
    fetchArchivedMonitors,
    fetchChecks,
    createMonitor,
    updateMonitor,
    deleteMonitor,
    archiveMonitor,
    restoreMonitor,
    purgeMonitor,
    checkNow,
    applyLiveUpdate,
    fetchIncidents,
    fetchIncidentDetail,
    fetchReport,
    fetchGroups,
    createGroup,
    updateGroup,
    deleteGroup,
    resetFilters
  };
});
