<template>
  <Modal :is-open="isOpen" title="Edit User Profile & Account Status" max-width="lg" @close="$emit('close')">
    <template #default>
      <form v-if="user" @submit.prevent="handleSubmit" class="space-y-5 text-xs">
        <!-- User Profile Bar -->
        <div class="flex items-center gap-3 bg-[#18181B] p-3 rounded-xl border border-[#26262A]">
          <img :src="user.avatarUrl || 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'" @error="(e: Event) => (e.target as HTMLImageElement).src = 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'" class="w-10 h-10 rounded-full object-cover border border-[#26262A]" />
          <div>
            <h3 class="font-bold text-sm text-white">{{ form.name }}</h3>
            <p class="text-xs font-mono text-[#7B96F5]">{{ form.email }}</p>
          </div>
          <span class="ml-auto px-2.5 py-0.5 rounded font-mono text-[10px] font-bold uppercase bg-[#7B96F5]/15 text-[#7B96F5] border border-[#7B96F5]/30">
            {{ form.role }}
          </span>
        </div>

        <!-- Identity Fields: Username, Name, Email, Role, Status -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Username -->
          <div class="space-y-1">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Username *</label>
            <input
              v-model="form.username"
              type="text"
              required
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
            />
          </div>

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
          <div class="space-y-1 md:col-span-2">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Email Address *</label>
            <input
              v-model="form.email"
              type="email"
              required
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
            />
          </div>

          <!-- Role -->
          <div class="space-y-1 md:col-span-2">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Assigned System Role *</label>
            <select
              v-model="form.role"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
            >
              <option value="admin">ADMIN (Full Unrestricted Access)</option>
              <option value="anggota">SANOC STAFF (Operator - Technical Actions)</option>
              <option value="pimpinan">PIMPINAN (Executive Dashboard &amp; Reports)</option>
            </select>
            <p class="text-[10px] font-mono text-gray-500 mt-1">
              Role assignment determines user permissions. Feature access matrix per role can be configured in the Roles &amp; Permissions section below.
            </p>
          </div>

          <!-- Account Status -->
          <div class="space-y-1 md:col-span-2">
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

        <!-- Reset Password Direct Action -->
        <div class="border-t border-[#26262A] pt-4 space-y-2">
          <h4 class="font-bold text-white text-xs uppercase tracking-wider font-mono flex items-center gap-2">
            <KeyRound class="w-4 h-4 text-amber-400" />
            Direct Password Reset
          </h4>
          <div class="flex items-center gap-2">
            <input
              v-model="newPassword"
              type="password"
              placeholder="Enter new password for this user"
              class="flex-1 bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono text-xs"
            />
            <button
              type="button"
              @click="handleResetPassword"
              :disabled="!newPassword.trim() || isSubmitting"
              class="px-3.5 py-2 rounded-lg bg-amber-500/20 border border-amber-500/30 hover:bg-amber-500/30 text-amber-300 font-bold text-xs font-mono disabled:opacity-40 transition-colors"
            >
              Reset Password
            </button>
          </div>
          <p v-if="resetSuccessMsg" class="text-[11px] font-mono text-emerald-400 mt-1">{{ resetSuccessMsg }}</p>
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
            <span>Save User Profile</span>
          </button>
        </div>
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import Modal from '../common/Modal.vue';
import { RefreshCw, KeyRound } from 'lucide-vue-next';
import type { User, UserRole } from '../../types';
import api from '../../api/client';

const props = defineProps<{
  isOpen: boolean;
  user: User | null;
}>();

const emit = defineEmits(['close', 'saved']);

const isSubmitting = ref(false);
const errorMsg = ref('');
const newPassword = ref('');
const resetSuccessMsg = ref('');

const form = reactive({
  username: '',
  name: '',
  email: '',
  role: 'anggota' as UserRole,
  status: 'Active'
});

watch(
  [() => props.user, () => props.isOpen],
  ([newUser, open]) => {
    if (newUser && open) {
      const fallback = newUser.email ? newUser.email.split('@')[0] : '';
      form.username = newUser.username || fallback || '';
      form.name = newUser.name || '';
      form.email = newUser.email || '';
      form.role = newUser.role || 'anggota';
      form.status = newUser.status || 'Active';
      newPassword.value = '';
      resetSuccessMsg.value = '';
      errorMsg.value = '';
    }
  },
  { immediate: true }
);

async function handleResetPassword() {
  if (!props.user || !newPassword.value.trim()) return;
  isSubmitting.value = true;
  resetSuccessMsg.value = '';
  errorMsg.value = '';
  try {
    await api.put(`/users/${props.user.id}/reset-password`, {
      password: newPassword.value.trim()
    });
    resetSuccessMsg.value = 'Password reset successfully!';
    newPassword.value = '';
    setTimeout(() => { resetSuccessMsg.value = ''; }, 3000);
  } catch (e: any) {
    errorMsg.value = e.response?.data?.error || 'Failed to reset password';
  } finally {
    isSubmitting.value = false;
  }
}

async function handleSubmit() {
  if (!props.user) return;
  isSubmitting.value = true;
  errorMsg.value = '';
  try {
    await api.put(`/users/${props.user.id}`, {
      username: form.username,
      name: form.name,
      email: form.email,
      role: form.role,
      status: form.status
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
