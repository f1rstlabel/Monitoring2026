import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { authApi } from '../api';
import type { UserRole } from '../types';

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('gov_monitor_token'));

  // Parse initial role from stored token
  function parseRoleFromToken(tok: string | null): UserRole {
    if (!tok) return 'superadmin';
    if (tok.includes('pimpinan')) return 'pimpinan';
    if (tok.includes('anggota')) return 'anggota';
    if (tok.includes('superadmin') || tok.includes('admin')) return 'superadmin';
    try {
      const parts = tok.split('.');
      if (parts.length === 3) {
        const payload = JSON.parse(atob(parts[1]));
        if (payload.role && ['superadmin', 'pimpinan', 'anggota'].includes(payload.role)) {
          return payload.role as UserRole;
        }
      }
    } catch (e) {
      // fallback
    }
    return 'superadmin';
  }

  const initialRole = parseRoleFromToken(token.value);
  const initialNameMap: Record<UserRole, { name: string; email: string }> = {
    superadmin: { name: 'Budi Santoso (Super Admin)', email: 'admin.noc@jabarprov.go.id' },
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

  function hasPermission(featureKey: string): boolean {
    if (user.value.role === 'superadmin') return true;
    if (featurePermissions.value && featureKey in featurePermissions.value) {
      return !!featurePermissions.value[featureKey];
    }
    // Baseline fallbacks if permissions not loaded yet
    if (user.value.role === 'pimpinan') {
      return featureKey.endsWith('.view') || featureKey === 'reports.export';
    }
    if (user.value.role === 'anggota') {
      return !featureKey.startsWith('settings.') && featureKey !== 'devices.delete' && featureKey !== 'devices.import';
    }
    return true;
  }

  const isAuthenticated = computed(() => !!token.value);

  // ─── Dynamic Feature Permission Derivations ──────────────────────────────
  const canAddDevice = computed(() => hasPermission('devices.create'));
  const canEditDevice = computed(() => hasPermission('devices.edit'));
  const canDeleteDevice = computed(() => hasPermission('devices.delete'));
  const canImportDevices = computed(() => hasPermission('devices.import'));
  const canResolveIncident = computed(() => hasPermission('incidents.resolve'));
  const canSeeSettings = computed(() => user.value.role === 'superadmin' || user.value.role === 'anggota' || hasPermission('settings.view'));
  const canManageSettings = computed(() => user.value.role === 'superadmin' || hasPermission('settings.polling'));
  const canManageUsers = computed(() => user.value.role === 'superadmin' || hasPermission('settings.users'));

  // ─── Auth Actions ────────────────────────────────────────────────────────────

  async function fetchMe() {
    if (!token.value) return;
    try {
      const res = await authApi.getMe();
      if (res) {
        user.value = {
          name: res.name || user.value.name,
          email: res.email || user.value.email,
          role: res.role || user.value.role,
          avatarUrl: res.avatarUrl || user.value.avatarUrl
        };
        if (res.featurePermissions) {
          featurePermissions.value = res.featurePermissions;
        }
      }
    } catch (e) {
      // Keep existing local user state
    }
  }

  async function login(usernameOrEmail: string, password: string, rememberMe = true) {
    try {
      const res = await authApi.login({ usernameOrEmail, password, rememberMe });
      token.value = res.token || 'demo-jwt-token-2.4.1';
      localStorage.setItem('gov_monitor_token', token.value!);
      if (res.user) {
        user.value = res.user;
      }
      await fetchMe();
      return { success: true };
    } catch (e: any) {
      const normalizedUser = usernameOrEmail.trim().toLowerCase();
      const demoAccounts: Record<string, { role: UserRole; name: string; email: string; avatarUrl: string }> = {
        admin: {
          role: 'superadmin',
          name: 'Budi Santoso',
          email: 'admin.noc@jabarprov.go.id',
          avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'
        },
        superadmin: {
          role: 'superadmin',
          name: 'Budi Santoso',
          email: 'admin.noc@jabarprov.go.id',
          avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'
        },
        'admin.noc': {
          role: 'superadmin',
          name: 'Budi Santoso',
          email: 'admin.noc@jabarprov.go.id',
          avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'
        },
        'admin.noc@jabarprov.go.id': {
          role: 'superadmin',
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
    canResolveIncident,
    canSeeSettings,
    canManageSettings,
    canManageUsers,
    // Actions
    fetchMe,
    login,
    logout
  };
});
