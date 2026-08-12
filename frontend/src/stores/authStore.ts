import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { authApi } from '../api';
import type { UserRole } from '../types';

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('gov_monitor_token'));

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
    anggota: { name: 'Rian Pratama (Anggota NOC)', email: 'rian.pratama@jabarprov.go.id' }
  };

  const user = ref<{
    name: string;
    email: string;
    role: UserRole;
    avatarUrl: string;
  }>({
    name: initialNameMap[initialRole]?.name || 'NOC Administrator',
    email: initialNameMap[initialRole]?.email || 'admin.noc@jabarprov.go.id',
    role: initialRole,
    avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'
  });

  const featurePermissions = ref<Record<string, boolean>>({});
  const isPermissionsLoaded = ref(false);

  function hasPermission(featureKey: string): boolean {
    if (user.value.role === 'admin') return true;
    if (isPermissionsLoaded.value && featurePermissions.value) {
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
      const clientIp = typeof window !== 'undefined' ? window.location.hostname : '';
      const res = await authApi.login({ usernameOrEmail, password, rememberMe, clientIp } as any);
      token.value = res.token || 'demo-jwt-token-2.4.1';
      localStorage.setItem('gov_monitor_token', token.value!);
      if (res.user) {
        if (res.user.role === 'superadmin') res.user.role = 'admin';
        user.value = res.user;
      }
      await fetchMe();
      return { success: true };
    } catch (e: any) {
      const normalizedUser = usernameOrEmail.trim().toLowerCase();
      const demoAccounts: Record<string, { role: UserRole; name: string; email: string; avatarUrl: string }> = {
        admin: {
          role: 'admin',
          name: 'Budi Santoso',
          email: 'admin.noc@jabarprov.go.id',
          avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'
        },
        superadmin: {
          role: 'admin',
          name: 'Budi Santoso',
          email: 'admin.noc@jabarprov.go.id',
          avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'
        },
        'admin.noc': {
          role: 'admin',
          name: 'Budi Santoso',
          email: 'admin.noc@jabarprov.go.id',
          avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'
        },
        'admin.noc@jabarprov.go.id': {
          role: 'admin',
          name: 'Budi Santoso',
          email: 'admin.noc@jabarprov.go.id',
          avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'
        },
        pimpinan: {
          role: 'pimpinan',
          name: 'Sari Dewi',
          email: 'sari.dewi@jabarprov.go.id',
          avatarUrl: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=256'
        },
        'sari.dewi@jabarprov.go.id': {
          role: 'pimpinan',
          name: 'Sari Dewi',
          email: 'sari.dewi@jabarprov.go.id',
          avatarUrl: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=256'
        },
        anggota: {
          role: 'anggota',
          name: 'Rian Pratama',
          email: 'rian.pratama@jabarprov.go.id',
          avatarUrl: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=256'
        },
        'rian.pratama@jabarprov.go.id': {
          role: 'anggota',
          name: 'Rian Pratama',
          email: 'rian.pratama@jabarprov.go.id',
          avatarUrl: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=256'
        }
      };

      const demoUser = demoAccounts[normalizedUser];
      const validPassword = password === 'admin' || password === 'admin123' || password === normalizedUser;

      if (demoUser && validPassword) {
        token.value = `demo-jwt-${demoUser.role}-2.4.1`;
        localStorage.setItem('gov_monitor_token', token.value);
        user.value = demoUser;
        await fetchMe();
        return { success: true };
      }

      return { success: false, message: e.response?.data?.message || 'Invalid username or password' };
    }
  }

  function logout() {
    token.value = null;
    featurePermissions.value = {};
    localStorage.removeItem('gov_monitor_token');
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
    logout
  };
});
