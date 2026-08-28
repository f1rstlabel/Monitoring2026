<template>
  <div class="bg-surface border border-subtle rounded-xl p-5 space-y-5 shadow-xl">
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-subtle pb-4">
      <div>
        <h2 class="text-xs font-bold uppercase tracking-wider text-text-main font-mono flex items-center gap-2">
          <ShieldAlert class="w-4 h-4 text-brand-periwinkle" />
          Role Permission Access Control Matrix
        </h2>
        <p class="text-xs text-text-secondary mt-1">
          Configure explicit feature permission matrix per role using radio controls. Admin bypasses all checks.
        </p>
      </div>

      <button
        @click="savePermissions"
        :disabled="isSaving || isLoading"
        class="px-4 py-2 rounded-lg bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-semibold text-xs shadow-md shadow-brand-periwinkle/20 transition-all flex items-center gap-2 disabled:opacity-50"
      >
        <Save class="w-4 h-4" />
        {{ isSaveSuccess ? 'Permissions Saved!' : isSaving ? 'Saving Matrix...' : 'Save Matrix' }}
      </button>
    </div>

    <!-- Role-Wide Notice Banner -->
    <div class="bg-brand-periwinkle/10 border border-brand-periwinkle/30 rounded-lg p-3 text-xs font-mono text-text-main flex items-center gap-2.5">
      <ShieldCheck class="w-4 h-4 text-brand-periwinkle flex-shrink-0" />
      <div>
        <span class="font-bold text-brand-periwinkle">Role-Based Access Control:</span>
        <span class="text-text-secondary">
          Feature access rules are configured per role. Editing permissions for <strong>Anggota SANOC</strong> or <strong>Pimpinan</strong> immediately updates feature capabilities for <strong>all users</strong> assigned that role.
        </span>
      </div>
    </div>

    <!-- Feedback Banner -->
    <div v-if="saveMessage" class="p-3 rounded-lg text-xs font-mono border" :class="saveSuccess ? 'bg-status-up/10 border-status-up/30 text-status-up' : 'bg-red-500/10 border-red-500/30 text-red-400'">
      {{ saveMessage }}
    </div>

    <!-- Matrix Table -->
    <div class="overflow-x-auto">
      <table class="w-full text-left text-xs text-text-secondary">
        <thead class="bg-card font-mono text-[10px] uppercase text-text-secondary border-b border-subtle">
          <tr>
            <th class="py-3 px-4">Feature Module &amp; Action</th>
            <th class="py-3 px-4 text-center w-48">
              <div>Pimpinan</div>
              <div class="text-[9px] text-brand-periwinkle font-semibold lowercase">({{ pimpinanCount }} user{{ pimpinanCount === 1 ? '' : 's' }})</div>
            </th>
            <th class="py-3 px-4 text-center w-48">
              <div>Anggota SANOC</div>
              <div class="text-[9px] text-brand-periwinkle font-semibold lowercase">({{ anggotaCount }} user{{ anggotaCount === 1 ? '' : 's' }})</div>
            </th>
            <th class="py-3 px-4 text-center w-36">
              <div>Admin</div>
              <div class="text-[9px] text-status-up font-semibold lowercase">({{ adminCount }} user{{ adminCount === 1 ? '' : 's' }})</div>
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-subtle">
          <tr v-if="isLoading">
            <td colspan="4" class="p-0 border-0">
              <SkeletonTable :rows="6" :cols="4" />
            </td>
          </tr>
          <template v-else v-for="group in featureGroups" :key="group.category">
            <!-- Category Header Row -->
            <tr class="bg-card/60 text-brand-periwinkle font-mono text-[11px] font-bold">
              <td colspan="4" class="py-2.5 px-4 uppercase tracking-wider bg-card">
                {{ group.category }}
              </td>
            </tr>

            <!-- Feature Row -->
            <tr v-for="feat in group.features" :key="feat.key" class="hover:bg-card transition-colors">
              <td class="py-3 px-4">
                <div>
                  <p class="font-bold text-text-main">{{ feat.label }}</p>
                  <p class="text-[10px] font-mono text-text-muted">{{ feat.key }} &bull; {{ feat.description }}</p>
                </div>
              </td>

              <!-- Pimpinan ON / OFF Switch Controls -->
              <td class="py-3 px-4 text-center">
                <div class="inline-flex p-0.5 rounded-lg bg-main border border-subtle">
                  <button
                    type="button"
                    @click="setPermission('pimpinan', feat.key, true)"
                    :class="getPermission('pimpinan', feat.key) === true ? 'bg-status-up/20 text-status-up border-status-up/40 font-bold shadow-sm' : 'text-text-muted hover:text-text-secondary border-transparent'"
                    class="px-2.5 py-1 text-[11px] font-mono rounded-md border transition-all flex items-center gap-1.5"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="getPermission('pimpinan', feat.key) === true ? 'bg-status-up pulsing-dot-green' : 'bg-gray-600'"></span>
                    ON
                  </button>
                  <button
                    type="button"
                    @click="setPermission('pimpinan', feat.key, false)"
                    :class="getPermission('pimpinan', feat.key) === false ? 'bg-red-500/20 text-red-400 border-red-500/40 font-bold shadow-sm' : 'text-text-muted hover:text-text-secondary border-transparent'"
                    class="px-2.5 py-1 text-[11px] font-mono rounded-md border transition-all flex items-center gap-1.5"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="getPermission('pimpinan', feat.key) === false ? 'bg-red-400' : 'bg-gray-600'"></span>
                    OFF
                  </button>
                </div>
              </td>

              <!-- Anggota ON / OFF Switch Controls -->
              <td class="py-3 px-4 text-center">
                <div class="inline-flex p-0.5 rounded-lg bg-main border border-subtle">
                  <button
                    type="button"
                    @click="setPermission('anggota', feat.key, true)"
                    :class="getPermission('anggota', feat.key) === true ? 'bg-status-up/20 text-status-up border-status-up/40 font-bold shadow-sm' : 'text-text-muted hover:text-text-secondary border-transparent'"
                    class="px-2.5 py-1 text-[11px] font-mono rounded-md border transition-all flex items-center gap-1.5"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="getPermission('anggota', feat.key) === true ? 'bg-status-up pulsing-dot-green' : 'bg-gray-600'"></span>
                    ON
                  </button>
                  <button
                    type="button"
                    @click="setPermission('anggota', feat.key, false)"
                    :class="getPermission('anggota', feat.key) === false ? 'bg-red-500/20 text-red-400 border-red-500/40 font-bold shadow-sm' : 'text-text-muted hover:text-text-secondary border-transparent'"
                    class="px-2.5 py-1 text-[11px] font-mono rounded-md border transition-all flex items-center gap-1.5"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="getPermission('anggota', feat.key) === false ? 'bg-red-400' : 'bg-gray-600'"></span>
                    OFF
                  </button>
                </div>
              </td>

              <!-- Admin (Locked Full Access) -->
              <td class="py-3 px-4 text-center">
                <span class="inline-flex items-center gap-1 text-[10px] font-mono font-bold text-status-up bg-status-up/10 px-2 py-1 rounded border border-status-up/30">
                  <Check class="w-3 h-3" /> ADMIN (FULL)
                </span>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { permissionsApi } from '../../api';
import { useAuthStore } from '../../stores/authStore';
import { useSettingStore } from '../../stores/settingStore';
import { ShieldAlert, ShieldCheck, Save, Check } from 'lucide-vue-next';
import SkeletonTable from '../common/SkeletonTable.vue';

interface FeatureDef {
  key: string;
  label: string;
  description: string;
}

interface FeatureGroup {
  category: string;
  features: FeatureDef[];
}

const featureGroups: FeatureGroup[] = [
  {
    category: 'Device Management',
    features: [
      { key: 'devices.view', label: 'View Devices Queue', description: 'Access devices page and detail view' },
      { key: 'devices.create', label: 'Add New Device', description: 'Register single network node' },
      { key: 'devices.edit', label: 'Edit Device Settings', description: 'Update IP, MAC, SNMP & thresholds' },
      { key: 'devices.delete', label: 'Delete Devices', description: 'Permanently remove node from inventory' },
      { key: 'devices.import', label: 'Bulk Import CSV/Excel', description: 'Batch upload multiple devices via spreadsheet' },
      { key: 'devices.bulk', label: 'Bulk Operations (Kelola Massal)', description: 'Batch update device attributes and bulk operations' },
      { key: 'diagnostics.run', label: 'Network Diagnostic Terminal', description: 'Run live ICMP ping, traceroute & port probing' }
    ]
  },
  {
    category: 'Incidents & Outages',
    features: [
      { key: 'incidents.view', label: 'View Incident Stream', description: 'Read-only access to alerts and outage logs' },
      { key: 'incidents.resolve', label: 'Acknowledge & Resolve Incidents', description: 'Mark alerts resolved or silenced' }
    ]
  },
  {
    category: 'Reports & Audits',
    features: [
      { key: 'reports.view', label: 'View Reports & Charts', description: 'Access downtime & SLA analytics' },
      { key: 'reports.export', label: 'Export PDF & Excel Reports', description: 'Download SLA audit reports' }
    ]
  },
  {
    category: 'Settings & Administration',
    features: [
      { key: 'settings.branding', label: 'Branding & Appearance', description: 'Change system logo, favicon and application titles' },
      { key: 'settings.notifications', label: 'Gateways & Alerts', description: 'Configure WhatsApp, Telegram & rate limits' },
      { key: 'settings.polling', label: 'Engine & Thresholds', description: 'Configure ICMP interval, debounce & failure rules' },
      { key: 'settings.dhcp_sync', label: 'DHCP Sync Engine', description: 'Configure Kea MySQL integration and view IP history logs' },
      { key: 'settings.network', label: 'Core Switch & SNMP', description: 'Configure cross-subnet L3 ARP SNMP target' },
      { key: 'settings.retention', label: 'Retention Policy', description: 'Configure incident archiving & housekeeping' },
      { key: 'settings.locations', label: 'Location Management', description: 'Add, edit and delete site locations' },
      { key: 'settings.users', label: 'Users & Roles', description: 'Manage accounts, passwords and privileges' },
      { key: 'settings.audit', label: 'Audit Logs & RBAC', description: 'View system audit logs and permission matrix' }
    ]
  }
];

const authStore = useAuthStore();
const settingStore = useSettingStore();

const pimpinanCount = computed(() => settingStore.users.filter(u => u.role === 'pimpinan').length);
const anggotaCount = computed(() => settingStore.users.filter(u => u.role === 'anggota').length);
const adminCount = computed(() => settingStore.users.filter(u => u.role === 'admin').length);
const isLoading = ref(false);
const isSaving = ref(false);
const isSaveSuccess = ref(false);
const saveMessage = ref('');
const saveSuccess = ref(true);

const matrix = reactive<Record<string, Record<string, boolean>>>({
  pimpinan: {
    'devices.view': true,
    'devices.create': false,
    'devices.edit': false,
    'devices.delete': false,
    'devices.import': false,
    'devices.bulk': false,
    'diagnostics.run': false,
    'incidents.view': true,
    'reports.view': true,
    'reports.export': true,
    'settings.branding': false,
    'settings.notifications': false,
    'settings.polling': false,
    'settings.dhcp_sync': false,
    'settings.network': false,
    'settings.retention': false,
    'settings.locations': false,
    'settings.users': false,
    'settings.audit': false
  },
  anggota: {
    'devices.view': true,
    'devices.create': true,
    'devices.edit': true,
    'devices.delete': false,
    'devices.import': false,
    'devices.bulk': false,
    'diagnostics.run': true,
    'incidents.view': true,
    'reports.view': true,
    'reports.export': true,
    'settings.branding': false,
    'settings.notifications': false,
    'settings.polling': false,
    'settings.dhcp_sync': false,
    'settings.network': false,
    'settings.retention': false,
    'settings.locations': false,
    'settings.users': false,
    'settings.audit': false
  }
});

function getPermission(role: string, featureKey: string): boolean {
  return matrix[role]?.[featureKey] ?? false;
}

function setPermission(role: string, featureKey: string, val: boolean) {
  if (!matrix[role]) matrix[role] = {};
  matrix[role][featureKey] = val;
}

async function loadPermissions() {
  isLoading.value = true;
  try {
    const list = await permissionsApi.getPermissions();
    if (list && list.length > 0) {
      for (const item of list) {
        if (!matrix[item.role]) matrix[item.role] = {};
        matrix[item.role][item.featureKey] = item.enabled;
      }
    }
  } catch (e) {
    console.error('Failed to fetch permissions matrix:', e);
  } finally {
    isLoading.value = false;
  }
}

async function savePermissions() {
  isSaving.value = true;
  saveMessage.value = '';
  try {
    const payload: { role: string; featureKey: string; enabled: boolean }[] = [];
    for (const role of ['pimpinan', 'anggota']) {
      for (const group of featureGroups) {
        for (const feat of group.features) {
          payload.push({
            role,
            featureKey: feat.key,
            enabled: getPermission(role, feat.key)
          });
        }
      }
    }

    await permissionsApi.updatePermissions(payload);
    await authStore.fetchMe();
    isSaveSuccess.value = true;
    saveSuccess.value = true;
    saveMessage.value = 'Per-role feature access permission matrix saved successfully.';
    setTimeout(() => { isSaveSuccess.value = false; saveMessage.value = ''; }, 4000);
  } catch (e: any) {
    saveSuccess.value = false;
    saveMessage.value = e.response?.data?.error || 'Failed to save permission matrix';
  } finally {
    isSaving.value = false;
  }
}

onMounted(() => {
  loadPermissions();
});
</script>
