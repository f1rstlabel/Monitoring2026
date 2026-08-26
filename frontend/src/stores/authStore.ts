import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { authApi } from '../api';
import type { UserRole } from '../types';

export const useAuthStore = defineStore('auth', () => {
  // Use user state to determine authentication instead of token, since token is now in HttpOnly cookie
  const token = ref<string | null>(null);



  const user = ref<{
    id?: string;
    username?: string;
    name: string;
    email: string;
    role: UserRole;
    avatarUrl: string;
    mfaEnabled?: boolean;
  }>({
    username: '',
    name: '',
    email: '',
    role: 'anggota',
    avatarUrl: ''
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

  const isAuthenticated = computed(() => !!user.value.email);

  // ─── Dynamic Feature Permission Derivations ──────────────────────────────
  const canAddDevice = computed(() => hasPermission('devices.create'));
  const canEditDevice = computed(() => hasPermission('devices.edit'));
  const canDeleteDevice = computed(() => hasPermission('devices.delete'));
  const canImportDevices = computed(() => hasPermission('devices.import'));
  const canBulkManageDevices = computed(() => hasPermission('devices.bulk'));
  const canSeeSettings = computed(() =>
    user.value.role === 'admin' ||
    hasPermission('settings.notifications') ||
    hasPermission('settings.polling') ||
    hasPermission('settings.network') ||
    hasPermission('settings.retention') ||
    hasPermission('settings.locations') ||
    hasPermission('settings.users') ||
    hasPermission('settings.audit')
  );
  const canManageSettings = computed(() =>
    user.value.role === 'admin' ||
    hasPermission('settings.notifications') ||
    hasPermission('settings.polling') ||
    hasPermission('settings.network') ||
    hasPermission('settings.retention') ||
    hasPermission('settings.locations')
  );
  const canManageUsers = computed(() => user.value.role === 'admin' || hasPermission('settings.users'));

  // ─── Auth Actions ────────────────────────────────────────────────────────────

  const isInitialized = ref(false);
  let initPromise: Promise<void> | null = null;

  async function fetchMe() {
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
    } finally {
      isInitialized.value = true;
    }
  }

  function init() {
    if (!initPromise) {
      initPromise = fetchMe();
    }
    return initPromise;
  }

  // Automatically fetch fresh user info and permissions if authenticated on load
  init();

  async function login(usernameOrEmail: string, password: string, rememberMe = true, recaptchaToken?: string) {
    try {
      const res = await authApi.login({ usernameOrEmail, password, rememberMe, recaptchaToken });
      if (res.requireMFA) {
        return { requireMFA: true, mfaToken: res.mfaToken, user: res.user };
      }
      if (res.token) {
        token.value = res.token;
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
    user.value = { username: '', name: '', email: '', role: 'anggota', avatarUrl: '' };
    featurePermissions.value = {};
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
    canBulkManageDevices,
    canSeeSettings,
    canManageSettings,
    canManageUsers,
    // Actions
    init,
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
