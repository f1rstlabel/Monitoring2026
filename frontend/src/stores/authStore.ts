import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { authApi } from '../api';
import type { UserRole } from '../types';

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('sanoc_token') || localStorage.getItem('gov_monitor_token'));

  // Parse initial role from stored token
  function parseRoleFromToken(tok: string | null): UserRole {
    if (!tok) return 'admin';
    if (tok.includes('pimpinan')) return 'pimpinan';
    if (tok.includes('anggota')) return 'anggota';
    if (tok.includes('admin') || tok.includes('superadmin')) return 'admin';
    try {
      const parts = tok.split('.');
      if (parts.length === 3) {
        const payload = JSON.parse(atob(parts[1]));
        if (payload.role) {
          if (payload.role === 'superadmin') return 'admin';
          if (['admin', 'pimpinan', 'anggota'].includes(payload.role)) {
            return payload.role as UserRole;
          }
        }
      }
    } catch (e) {
      // fallback
    }
    return 'admin';
  }

  const initialRole = parseRoleFromToken(token.value);
  const initialNameMap: Record<UserRole, { name: string; email: string }> = {
    admin: { name: 'Budi Santoso (Admin)', email: 'admin.noc@jabarprov.go.id' },
    pimpinan: { name: 'Sari Dewi (Pimpinan)', email: 'sari.dewi@jabarprov.go.id' },
    anggota: { name: 'Rian Pratama (Anggota SANOC)', email: 'rian.pratama@jabarprov.go.id' }
  };

  const user = ref<{
    username?: string;
    name: string;
    email: string;
    role: UserRole;
    avatarUrl: string;
  }>({
    username: 'admin.noc',
    name: initialNameMap[initialRole]?.name || 'SANOC Administrator',
    email: initialNameMap[initialRole]?.email || 'admin.noc@jabarprov.go.id',
    role: initialRole,
    avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'
  });

  const featurePermissions = ref<Record<string, boolean>>({});
  const isPermissionsLoaded = ref(false);

  function hasPermission(featureKey: string): boolean {
    if (user.value.role === 'admin') return true;
    if (featureKey === 'settings.view') {
      return canSeeSettings.value;
    }
    if (isPermissionsLoaded.value && featurePermissions.value && featureKey in featurePermissions.value) {
      return !!featurePermissions.value[featureKey];
    }
    // Baseline defaults matching initial DB seed
    if (user.value.role === 'pimpinan') {
      return featureKey === 'devices.view' || featureKey === 'incidents.view' || featureKey === 'reports.view' || featureKey === 'reports.export';
    }
    if (user.value.role === 'anggota') {
      return featureKey === 'devices.view' || featureKey === 'devices.create' || featureKey === 'devices.edit' || featureKey === 'incidents.view' || featureKey === 'reports.view' || featureKey === 'reports.export';
    }
    return false;
  }

  const isAuthenticated = computed(() => !!token.value);

  // ─── Dynamic Feature Permission Derivations ──────────────────────────────
  const canAddDevice = computed(() => hasPermission('devices.create'));
  const canEditDevice = computed(() => hasPermission('devices.edit'));
  const canDeleteDevice = computed(() => hasPermission('devices.delete'));
  const canImportDevices = computed(() => hasPermission('devices.import'));
  const canSeeSettings = computed(() =>
    user.value.role === 'admin' ||
    hasPermission('settings.notifications') ||
    hasPermission('settings.polling') ||
    hasPermission('settings.thresholds') ||
    hasPermission('settings.users') ||
    hasPermission('settings.permissions')
  );
  const canManageSettings = computed(() =>
    user.value.role === 'admin' ||
    hasPermission('settings.polling') ||
    hasPermission('settings.thresholds') ||
    hasPermission('settings.notifications')
  );
  const canManageUsers = computed(() => user.value.role === 'admin' || hasPermission('settings.users'));

  // ─── Auth Actions ────────────────────────────────────────────────────────────

  async function fetchMe() {
    if (!token.value) return;
    try {
      const res = await authApi.getMe();
      if (res) {
        let r = res.role;
        if (r === 'superadmin') r = 'admin';
        user.value = {
          name: res.name || user.value.name,
          email: res.email || user.value.email,
          role: r || user.value.role,
          avatarUrl: res.avatarUrl || user.value.avatarUrl
        };
        if (res.featurePermissions) {
          featurePermissions.value = res.featurePermissions;
          isPermissionsLoaded.value = true;
        }
      }
    } catch (e) {
      // Keep existing local user state
    }
  }

  // Automatically fetch fresh user info and permissions if authenticated on load
  if (token.value) {
    fetchMe();
  }

  async function login(usernameOrEmail: string, password: string, rememberMe = true) {
    try {
      const res = await authApi.login({ usernameOrEmail, password, rememberMe });
      if (res.requireMFA) {
        return { requireMFA: true, mfaToken: res.mfaToken, user: res.user };
      }
      if (res.token) {
        token.value = res.token;
        localStorage.setItem('sanoc_token', res.token);
        localStorage.setItem('gov_monitor_token', res.token);
      }
      if (res.csrfToken) {
        localStorage.setItem('sanoc_csrf_token', res.csrfToken);
      }
      if (res.user) {
        if (res.user.role === 'superadmin') res.user.role = 'admin';
        user.value = res.user;
      }
      fetchMe().catch(() => {});
      return { success: true };
    } catch (e: any) {
      return { success: false, message: e.response?.data?.message || e.response?.data?.error || 'Invalid username or password' };
    }
  }

  async function verifyLoginMFA(mfaToken: string, code: string) {
    try {
      const res = await authApi.verifyLoginMFA(mfaToken, code);
      if (res.token) {
        token.value = res.token;
        localStorage.setItem('sanoc_token', res.token);
        localStorage.setItem('gov_monitor_token', res.token);
      }
      if (res.csrfToken) {
        localStorage.setItem('sanoc_csrf_token', res.csrfToken);
      }
      if (res.user) {
        if (res.user.role === 'superadmin') res.user.role = 'admin';
        user.value = res.user;
      }
      fetchMe().catch(() => {});
      return { success: true };
    } catch (e: any) {
      return { success: false, message: e.response?.data?.message || e.response?.data?.error || 'Invalid MFA passcode' };
    }
  }

  async function setupMFA() {
    try {
      const res = await authApi.setupMFA();
      return res;
    } catch (e: any) {
      return { success: false, message: e.response?.data?.error || 'Failed to setup MFA' };
    }
  }

  async function verifyMFA(secret: string, code: string) {
    try {
      const res = await authApi.verifyMFA(secret, code);
      return res;
    } catch (e: any) {
      return { success: false, message: e.response?.data?.error || 'Invalid TOTP passcode' };
    }
  }

  async function disableMFA(password?: string) {
    try {
      const res = await authApi.disableMFA(password);
      return res;
    } catch (e: any) {
      return { success: false, message: e.response?.data?.error || 'Failed to disable MFA' };
    }
  }

  function logout() {
    token.value = null;
    featurePermissions.value = {};
    localStorage.removeItem('sanoc_token');
    localStorage.removeItem('gov_monitor_token');
    localStorage.removeItem('sanoc_csrf_token');
  }

  async function updateProfile(data: { username?: string; name: string; email: string; avatarUrl: string; currentPassword?: string; newPassword?: string }) {
    try {
      const res = await authApi.updateProfile(data);
      if (res && res.user) {
        let r = res.user.role;
        if (r === 'superadmin') r = 'admin';
        user.value = {
          ...user.value,
          username: res.user.username || data.username || user.value.username,
          name: res.user.name || data.name,
          email: res.user.email || data.email,
          role: r || user.value.role,
          avatarUrl: res.user.avatarUrl || data.avatarUrl
        };
      } else {
        user.value = {
          ...user.value,
          username: data.username || user.value.username,
          name: data.name,
          email: data.email,
          avatarUrl: data.avatarUrl
        };
      }
      return { success: true, message: res?.message || 'Profile updated successfully' };
    } catch (e: any) {
      user.value = {
        ...user.value,
        username: data.username || user.value.username,
        name: data.name,
        email: data.email,
        avatarUrl: data.avatarUrl
      };
      return {
        success: true,
        message: e.response?.data?.error || 'Profile updated in session'
      };
    }
  }

  async function uploadAvatar(file: File) {
    try {
      const res = await authApi.uploadAvatar(file);
      if (res && res.avatarUrl) {
        user.value = {
          ...user.value,
          avatarUrl: res.avatarUrl
        };
        return { success: true, avatarUrl: res.avatarUrl, message: 'Avatar uploaded successfully' };
      }
      return { success: false, message: 'Failed to obtain uploaded avatar URL' };
    } catch (e: any) {
      return { success: false, message: e.response?.data?.error || 'Failed to upload avatar image' };
    }
  }

  return {
    token,
    user,
    featurePermissions,
    isAuthenticated,
    hasPermission,
    // Permission flags
    canAddDevice,
    canEditDevice,
    canDeleteDevice,
    canImportDevices,
    canSeeSettings,
    canManageSettings,
    canManageUsers,
    // Actions
    fetchMe,
    login,
    verifyLoginMFA,
    setupMFA,
    verifyMFA,
    disableMFA,
    logout,
    updateProfile,
    uploadAvatar
  };
});
