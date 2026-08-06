<template>
  <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-5 shadow-xl">
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-[#26262A] pb-4">
      <div>
        <h2 class="text-xs font-bold uppercase tracking-wider text-white font-mono flex items-center gap-2">
          <ShieldAlert class="w-4 h-4 text-[#7B96F5]" />
          Role Permission Access Control Matrix
        </h2>
        <p class="text-xs text-gray-400 mt-1">
          Configure granular feature capability access per role. Changes take effect instantly across frontend &amp; backend middleware.
        </p>
      </div>

      <button
        @click="savePermissions"
        :disabled="isSaving || isLoading"
        class="px-4 py-2 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 transition-all flex items-center gap-2 disabled:opacity-50"
      >
        <Save class="w-4 h-4" />
        {{ isSaveSuccess ? 'Permissions Saved!' : isSaving ? 'Saving Matrix...' : 'Save Matrix' }}
      </button>
    </div>

    <!-- Feedback Banner -->
    <div v-if="saveMessage" class="p-3 rounded-lg text-xs font-mono border" :class="saveSuccess ? 'bg-[#3ECF8E]/10 border-[#3ECF8E]/30 text-[#3ECF8E]' : 'bg-red-500/10 border-red-500/30 text-red-400'">
      {{ saveMessage }}
    </div>

    <!-- Matrix Table -->
    <div class="overflow-x-auto">
      <table class="w-full text-left text-xs text-gray-300">
        <thead class="bg-[#18181B] font-mono text-[10px] uppercase text-gray-400 border-b border-[#26262A]">
          <tr>
            <th class="py-3 px-4">Feature Module &amp; Action</th>
            <th class="py-3 px-4 text-center w-36">Pimpinan</th>
            <th class="py-3 px-4 text-center w-36">Anggota NOC</th>
            <th class="py-3 px-4 text-center w-32">Super Admin</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#26262A]">
          <template v-for="group in featureGroups" :key="group.category">
            <!-- Category Header Row -->
            <tr class="bg-[#18181B]/60 text-[#7B96F5] font-mono text-[11px] font-bold">
              <td colspan="4" class="py-2.5 px-4 uppercase tracking-wider bg-[#18181B]">
                {{ group.category }}
              </td>
            </tr>

            <!-- Feature Row -->
            <tr v-for="feat in group.features" :key="feat.key" class="hover:bg-[#18181B] transition-colors">
              <td class="py-3 px-4">
                <div>
                  <p class="font-bold text-white">{{ feat.label }}</p>
                  <p class="text-[10px] font-mono text-gray-500">{{ feat.key }} &bull; {{ feat.description }}</p>
                </div>
              </td>

              <!-- Pimpinan Toggle -->
              <td class="py-3 px-4 text-center">
                <button
                  type="button"
                  @click="togglePermission('pimpinan', feat.key)"
                  class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                  :class="getPermission('pimpinan', feat.key) ? 'bg-[#7B96F5]' : 'bg-[#26262A]'"
                >
                  <span
                    class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                    :class="getPermission('pimpinan', feat.key) ? 'translate-x-4' : 'translate-x-0'"
                  />
                </button>
              </td>

              <!-- Anggota Toggle -->
              <td class="py-3 px-4 text-center">
                <button
                  type="button"
                  @click="togglePermission('anggota', feat.key)"
                  class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                  :class="getPermission('anggota', feat.key) ? 'bg-[#7B96F5]' : 'bg-[#26262A]'"
                >
                  <span
                    class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                    :class="getPermission('anggota', feat.key) ? 'translate-x-4' : 'translate-x-0'"
                  />
                </button>
              </td>

              <!-- Super Admin (Locked Full Access) -->
              <td class="py-3 px-4 text-center">
                <span class="inline-flex items-center gap-1 text-[10px] font-mono font-bold text-[#3ECF8E] bg-[#3ECF8E]/10 px-2 py-0.5 rounded border border-[#3ECF8E]/30">
                  <Check class="w-3 h-3" /> FULL
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
import { ref, reactive, onMounted } from 'vue';
import { permissionsApi } from '../../api';
import { useAuthStore } from '../../stores/authStore';
import { ShieldAlert, Save, Check } from 'lucide-vue-next';

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
      { key: 'devices.import', label: 'Bulk Import CSV/Excel', description: 'Batch import devices' }
    ]
  },
  {
    category: 'Incidents Queue',
    features: [
      { key: 'incidents.view', label: 'View Active Incidents', description: 'Access incidents queue & timeline' },
      { key: 'incidents.resolve', label: 'Mark Incident Resolved', description: 'Change status to RESOLVED' }
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
      { key: 'settings.notifications', label: 'Manage Alert Gateways', description: 'Configure WhatsApp & Telegram' },
      { key: 'settings.polling', label: 'Configure Engine Polling', description: 'Update interval, batch size & debounce' },
      { key: 'settings.thresholds', label: 'Manage Failure Thresholds', description: 'Update consecutive failure rules' },
      { key: 'settings.users', label: 'Manage Users & Invites', description: 'Invite users & edit user roles' },
      { key: 'settings.permissions', label: 'Access Control Matrix', description: 'Configure per-role permissions' }
    ]
  }
];

const authStore = useAuthStore();
const isLoading = ref(false);
const isSaving = ref(false);
const isSaveSuccess = ref(false);
const saveMessage = ref('');
const saveSuccess = ref(true);

// State matrix: role -> featureKey -> boolean
const matrix = reactive<Record<string, Record<string, boolean>>>({
  pimpinan: {
    'devices.view': true,
    'devices.create': false,
    'devices.edit': false,
    'devices.delete': false,
    'devices.import': false,
    'incidents.view': true,
    'incidents.resolve': false,
    'reports.view': true,
    'reports.export': true,
    'settings.notifications': false,
    'settings.polling': false,
    'settings.thresholds': false,
    'settings.users': false,
    'settings.permissions': false
  },
  anggota: {
    'devices.view': true,
    'devices.create': true,
    'devices.edit': true,
    'devices.delete': false,
    'devices.import': false,
    'incidents.view': true,
    'incidents.resolve': true,
    'reports.view': true,
    'reports.export': true,
    'settings.notifications': false,
    'settings.polling': false,
    'settings.thresholds': false,
    'settings.users': false,
    'settings.permissions': false
  }
});

function getPermission(role: string, featureKey: string): boolean {
  return matrix[role]?.[featureKey] ?? false;
}

function togglePermission(role: string, featureKey: string) {
  if (!matrix[role]) matrix[role] = {};
  matrix[role][featureKey] = !matrix[role][featureKey];
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
    saveMessage.value = 'Per-role feature access permission matrix saved and updated across system.';
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
