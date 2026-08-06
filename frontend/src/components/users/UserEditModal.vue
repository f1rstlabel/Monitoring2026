<template>
  <Modal :is-open="isOpen" title="Edit User & Feature Access Permissions" max-width="2xl" @close="$emit('close')">
    <template #default>
      <form v-if="user" @submit.prevent="handleSubmit" class="space-y-5 text-xs">
        <!-- User Profile Bar -->
        <div class="flex items-center gap-3 bg-[#18181B] p-3 rounded-xl border border-[#26262A]">
          <img :src="user.avatarUrl || 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'" class="w-10 h-10 rounded-full object-cover border border-[#26262A]" />
          <div>
            <h3 class="font-bold text-sm text-white">{{ form.name }}</h3>
            <p class="text-xs font-mono text-[#7B96F5]">{{ form.email }}</p>
          </div>
          <span class="ml-auto px-2.5 py-0.5 rounded font-mono text-[10px] font-bold uppercase bg-[#7B96F5]/15 text-[#7B96F5] border border-[#7B96F5]/30">
            {{ form.role }}
          </span>
        </div>

        <!-- Identity Fields: Name, Email, Role, Status -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Name -->
          <div class="space-y-1">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Full Name *</label>
            <input
              v-model="form.name"
              type="text"
              required
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
            />
          </div>

          <!-- Email -->
          <div class="space-y-1">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Email Address *</label>
            <input
              v-model="form.email"
              type="email"
              required
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
            />
          </div>

          <!-- Role -->
          <div class="space-y-1">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">System Role *</label>
            <select
              v-model="form.role"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
            >
              <option value="superadmin">SUPER ADMIN (Full Unrestricted Access)</option>
              <option value="anggota">NOC STAFF (Operator - Technical Actions)</option>
              <option value="pimpinan">PIMPINAN (Executive Dashboard & Reports)</option>
            </select>
          </div>

          <!-- Account Status -->
          <div class="space-y-1">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Account Status *</label>
            <select
              v-model="form.status"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
            >
              <option value="Active">Active (Permitted to Log In)</option>
              <option value="Inactive">Inactive (Disabled / Deactivated)</option>
            </select>
          </div>
        </div>

        <!-- Role Feature Access Permissions Section -->
        <div class="border-t border-[#26262A] pt-4 space-y-3">
          <div class="flex items-center justify-between">
            <div>
              <h4 class="font-bold text-white text-xs uppercase tracking-wider font-mono flex items-center gap-2">
                <ShieldCheck class="w-4 h-4 text-[#7B96F5]" />
                Feature Access Permission Matrix
              </h4>
              <p class="text-[11px] text-gray-400 mt-0.5">Toggle explicit feature capabilities assigned to this user</p>
            </div>
            <span v-if="form.role === 'superadmin'" class="text-[10px] font-mono text-[#3ECF8E] bg-[#3ECF8E]/10 px-2 py-0.5 rounded border border-[#3ECF8E]/30">
              Super Admin Override Enabled
            </span>
          </div>

          <!-- Feature Permission Toggles Grid -->
          <div class="space-y-3 bg-[#18181B] p-4 rounded-xl border border-[#26262A]">
            <div v-for="group in featureGroups" :key="group.category" class="space-y-2">
              <p class="font-mono text-[10px] font-bold text-[#7B96F5] uppercase tracking-wider border-b border-[#26262A] pb-1">
                {{ group.category }}
              </p>
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                <label
                  v-for="feat in group.features"
                  :key="feat.key"
                  class="flex items-start gap-2.5 p-2 rounded-lg bg-[#151517] border border-[#26262A] hover:border-[#7B96F5]/50 cursor-pointer transition-colors"
                >
                  <input
                    type="checkbox"
                    :value="feat.key"
                    v-model="form.permissions"
                    :disabled="form.role === 'superadmin'"
                    class="mt-0.5 rounded border-[#26262A] bg-[#0A0A0B] text-[#7B96F5] focus:ring-[#7B96F5] disabled:opacity-50"
                  />
                  <div>
                    <p class="text-xs font-bold text-gray-200">{{ feat.label }}</p>
                    <p class="text-[10px] font-mono text-gray-500">{{ feat.key }}</p>
                  </div>
                </label>
              </div>
            </div>
          </div>
        </div>

        <div v-if="errorMsg" class="bg-[#F16565]/10 border border-[#F16565]/30 rounded-lg p-2.5 text-xs text-[#F16565] font-mono">
          {{ errorMsg }}
        </div>
      </form>
    </template>

    <template #footer>
      <div class="flex items-center justify-between w-full">
        <button
          type="button"
          @click="handleDeactivate"
          :disabled="isSubmitting"
          class="px-3 py-1.5 rounded-lg border border-[#F16565]/40 bg-[#F16565]/10 hover:bg-[#F16565]/20 text-[#F16565] text-xs font-mono font-medium transition-colors"
        >
          Deactivate User
        </button>

        <div class="flex items-center gap-2">
          <button
            type="button"
            @click="$emit('close')"
            class="px-4 py-1.5 rounded-lg border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs"
          >
            Cancel
          </button>
          <button
            type="button"
            @click="handleSubmit"
            :disabled="isSubmitting"
            class="px-4 py-1.5 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs flex items-center gap-1.5 shadow-md shadow-[#7B96F5]/20 disabled:opacity-50"
          >
            <RefreshCw v-if="isSubmitting" class="w-3.5 h-3.5 animate-spin" />
            <span>Save User &amp; Permissions</span>
          </button>
        </div>
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import Modal from '../common/Modal.vue';
import { RefreshCw, ShieldCheck } from 'lucide-vue-next';
import type { User, UserRole } from '../../types';
import api from '../../api/client';

const props = defineProps<{
  isOpen: boolean;
  user: User | null;
}>();

const emit = defineEmits(['close', 'saved']);

const isSubmitting = ref(false);
const errorMsg = ref('');

const featureGroups = [
  {
    category: 'Device Management',
    features: [
      { key: 'devices.view', label: 'View Devices Queue' },
      { key: 'devices.create', label: 'Add New Device' },
      { key: 'devices.edit', label: 'Edit Device Parameters' },
      { key: 'devices.delete', label: 'Delete Devices' },
      { key: 'devices.import', label: 'Bulk Import CSV/Excel' }
    ]
  },
  {
    category: 'Incidents Queue',
    features: [
      { key: 'incidents.view', label: 'View Active Incidents' },
      { key: 'incidents.resolve', label: 'Mark Incidents Resolved' }
    ]
  },
  {
    category: 'Reports & Audits',
    features: [
      { key: 'reports.view', label: 'View SLA Reports & Analytics' },
      { key: 'reports.export', label: 'Export PDF & Excel Audits' }
    ]
  },
  {
    category: 'Settings & Gateways',
    features: [
      { key: 'settings.notifications', label: 'Manage Alert Channels' },
      { key: 'settings.polling', label: 'Configure Polling Engine' },
      { key: 'settings.users', label: 'Manage Users & Permissions' }
    ]
  }
];

const form = reactive({
  name: '',
  email: '',
  role: 'anggota' as UserRole,
  status: 'Active',
  permissions: [] as string[]
});

const defaultRolePermissions: Record<UserRole, string[]> = {
  superadmin: [
    'devices.view', 'devices.create', 'devices.edit', 'devices.delete', 'devices.import',
    'incidents.view', 'incidents.resolve',
    'reports.view', 'reports.export',
    'settings.notifications', 'settings.polling', 'settings.users'
  ],
  anggota: [
    'devices.view', 'devices.create', 'devices.edit',
    'incidents.view', 'incidents.resolve',
    'reports.view', 'reports.export'
  ],
  pimpinan: [
    'devices.view', 'incidents.view', 'reports.view', 'reports.export'
  ]
};

watch(
  [() => props.user, () => props.isOpen],
  ([newUser, open]) => {
    if (newUser && open) {
      form.name = newUser.name || '';
      form.email = newUser.email || '';
      form.role = newUser.role || 'anggota';
      form.status = newUser.status || 'Active';
      form.permissions = newUser.permissions && newUser.permissions.length > 0
        ? [...newUser.permissions]
        : [...(defaultRolePermissions[newUser.role] || [])];
    }
  },
  { immediate: true }
);

watch(
  () => form.role,
  (newRole) => {
    if (newRole && defaultRolePermissions[newRole]) {
      form.permissions = [...defaultRolePermissions[newRole]];
    }
  }
);

async function handleSubmit() {
  if (!props.user) return;
  isSubmitting.value = true;
  errorMsg.value = '';
  try {
    await api.put(`/users/${props.user.id}`, {
      name: form.name,
      email: form.email,
      role: form.role,
      status: form.status,
      permissions: form.permissions
    });
    emit('saved');
    emit('close');
  } catch (e: any) {
    errorMsg.value = e.response?.data?.error || 'Failed to update user';
  } finally {
    isSubmitting.value = false;
  }
}

async function handleDeactivate() {
  if (!props.user) return;
  if (!confirm(`Are you sure you want to deactivate ${props.user.name}?`)) return;
  isSubmitting.value = true;
  try {
    await api.delete(`/users/${props.user.id}`);
    emit('saved');
    emit('close');
  } catch (e: any) {
    errorMsg.value = e.response?.data?.error || 'Failed to deactivate user';
  } finally {
    isSubmitting.value = false;
  }
}
</script>
