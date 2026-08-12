import api from './client';
import type {
  Device,
  DashboardSummary,
  Incident,
  StatusHistoryPoint,
  SystemSettings,
  User
} from '../types';

export const authApi = {
  login: async (credentials: { usernameOrEmail: string; password: string; rememberMe?: boolean }) => {
    const res = await api.post('/auth/login', credentials);
    return res.data;
  },
  verifyLoginMFA: async (mfaToken: string, code: string) => {
    const res = await api.post('/auth/verify-mfa', { mfaToken, code });
    return res.data;
  },
  setupMFA: async () => {
    const res = await api.post('/auth/mfa/setup');
    return res.data;
  },
  verifyMFA: async (secret: string, code: string) => {
    const res = await api.post('/auth/mfa/verify', { secret, code });
    return res.data;
  },
  disableMFA: async (password?: string) => {
    const res = await api.post('/auth/mfa/disable', { password });
    return res.data;
  },
  getMe: async () => {
    const res = await api.get('/auth/me');
    return res.data;
  },
  logout: async () => {
    const res = await api.post('/auth/logout');
    return res.data;
  },
  forgotPassword: async (email: string) => {
    const res = await api.post('/auth/forgot-password', { email });
    return res.data;
  },
  resetPassword: async (token: string, newPassword: string) => {
    const res = await api.post('/auth/reset-password', { token, newPassword });
    return res.data;
  },
  updateProfile: async (data: { username?: string; name: string; email: string; avatarUrl: string; currentPassword?: string; newPassword?: string }) => {
    const res = await api.put('/auth/profile', data);
    return res.data;
  },
  uploadAvatar: async (file: File) => {
    const formData = new FormData();
    formData.append('avatar', file);
    const res = await api.post('/auth/avatar', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    return res.data;
  }
};

export const dashboardApi = {
  getSummary: async (): Promise<DashboardSummary> => {
    const res = await api.get('/dashboard/summary');
    return res.data;
  },
  refreshNow: async (): Promise<{ success: boolean; polledCount: number; timestamp: string }> => {
    const res = await api.post('/monitoring/refresh-now');
    return res.data;
  }
};

export const locationsApi = {
  getLocations: async (search?: string): Promise<{ id: string; name: string; description?: string }[]> => {
    const res = await api.get('/locations', { params: { search } });
    return res.data;
  },
  createLocation: async (name: string, description?: string): Promise<{ id: string; name: string; description?: string }> => {
    const res = await api.post('/locations', { name, description });
    return res.data;
  }
};

export const permissionsApi = {
  getPermissions: async (): Promise<{ role: string; featureKey: string; enabled: boolean }[]> => {
    const res = await api.get('/permissions');
    return res.data;
  },
  updatePermissions: async (permissions: { role: string; featureKey: string; enabled: boolean }[]): Promise<{ success: boolean }> => {
    const res = await api.put('/permissions', permissions);
    return res.data;
  }
};

export const devicesApi = {
  getDevices: async (params?: { type?: string; status?: string; search?: string; page?: number; page_size?: number }): Promise<any> => {
    const res = await api.get('/devices', { params });
    return res.data;
  },
  getDeviceById: async (id: string): Promise<Device> => {
    const res = await api.get(`/devices/${id}`);
    return res.data;
  },
  createDevice: async (device: Partial<Device>): Promise<Device> => {
    const res = await api.post('/devices', device);
    return res.data;
  },
  updateDevice: async (id: string, device: Partial<Device>): Promise<Device> => {
    const res = await api.put(`/devices/${id}`, device);
    return res.data;
  },
  deleteDevice: async (id: string): Promise<{ success: boolean }> => {
    const res = await api.delete(`/devices/${id}`);
    return res.data;
  },
  getStatusHistory: async (id: string, range = '7d'): Promise<StatusHistoryPoint[]> => {
    const res = await api.get(`/devices/${id}/status-history`, { params: { range } });
    return res.data;
  }
};

export const incidentsApi = {
  getIncidents: async (params?: { status?: string; search?: string; deviceId?: string; page?: number; page_size?: number }): Promise<any> => {
    const res = await api.get('/incidents', { params });
    return res.data;
  },
  getIncidentById: async (id: string): Promise<Incident> => {
    const res = await api.get(`/incidents/${id}`);
    return res.data;
  }
};

export const settingsApi = {
  getSettings: async (): Promise<SystemSettings> => {
    const res = await api.get('/settings');
    return res.data;
  },
  updateThresholds: async (thresholds: any): Promise<any> => {
    const res = await api.put('/settings/thresholds', { thresholds });
    return res.data;
  },
  updatePolling: async (polling: { intervalSeconds: number; concurrencyBatchSize: number; flapReuseWindowMinutes?: number }): Promise<any> => {
    const res = await api.put('/settings/polling', polling);
    return res.data;
  },
  updateRateLimit: async (maxMsgPerMin: number): Promise<any> => {
    const res = await api.put('/settings/rate-limit', { maxMsgPerMin });
    return res.data;
  }
};

export const usersApi = {
  getUsers: async (): Promise<User[]> => {
    const res = await api.get('/users');
    return res.data;
  },
  inviteUser: async (user: { name: string; email: string; role: string }): Promise<User> => {
    const res = await api.post('/users/invite', user);
    return res.data;
  }
};
