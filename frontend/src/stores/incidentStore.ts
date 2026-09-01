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
      incidents.value = [];
      totalCount.value = 0;
    } finally {
      isLoading.value = false;
    }
  }

  async function fetchIncidentById(id: string) {
    isLoading.value = true;
    try {
      currentIncident.value = await incidentsApi.getIncidentById(id);
    } catch (e) {
      throw e;
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
