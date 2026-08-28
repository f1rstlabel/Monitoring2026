<template>
  <Modal :is-open="isOpen" title="Configure WhatsApp Notification Targets" @close="$emit('close')">
    <div class="space-y-4">
      <!-- Add / Edit Target Form Card -->
      <div class="bg-card border border-subtle rounded-xl p-4 space-y-4 shadow-lg">
        <h4 class="text-xs font-bold text-text-main flex items-center justify-between uppercase tracking-wider font-mono">
          <span>{{ editingId ? 'Edit Target' : 'Add New Target' }}</span>
          <button v-if="editingId" @click="cancelEdit" class="text-[11px] text-text-secondary hover:text-text-main font-normal capitalize">
            Cancel Edit
          </button>
        </h4>

        <!-- Target Type Segmented Control -->
        <div class="space-y-1.5">
          <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">Target Type</label>
          <div class="grid grid-cols-2 gap-2 bg-surface p-1 rounded-lg border border-subtle">
            <button
              type="button"
              @click="form.targetType = 'individual'"
              class="py-1.5 px-3 rounded-md text-xs font-semibold font-mono transition-all flex items-center justify-center gap-2"
              :class="form.targetType === 'individual' ? 'bg-brand-periwinkle text-white shadow-sm' : 'text-text-secondary hover:text-text-main hover:bg-card'"
            >
              <User class="w-3.5 h-3.5" />
              <span>Individual Number</span>
            </button>

            <button
              type="button"
              @click="form.targetType = 'group'"
              class="py-1.5 px-3 rounded-md text-xs font-semibold font-mono transition-all flex items-center justify-center gap-2"
              :class="form.targetType === 'group' ? 'bg-brand-periwinkle text-white shadow-sm' : 'text-text-secondary hover:text-text-main hover:bg-card'"
            >
              <Users class="w-3.5 h-3.5" />
              <span>WhatsApp Group (JID)</span>
            </button>
          </div>
        </div>

        <!-- Form Inputs Grid -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">Target Label</label>
            <input
              v-model="form.label"
              type="text"
              placeholder="e.g. NOC On-Call Group"
              class="w-full bg-surface border border-subtle rounded-lg px-3 py-2 text-xs text-text-main focus:outline-none focus:border-brand-periwinkle"
            />
          </div>

          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">
              {{ form.targetType === 'group' ? 'Group JID' : 'Phone Number' }}
            </label>
            <input
              v-model="form.phoneNumber"
              type="text"
              :placeholder="form.targetType === 'group' ? 'e.g. 120363028192837128@g.us' : 'e.g. +6281234567890'"
              class="w-full bg-surface border border-subtle rounded-lg px-3 py-2 text-xs text-text-main font-mono focus:outline-none focus:border-brand-periwinkle"
            />
            <p class="text-[10px] text-text-muted font-mono">
              {{ form.targetType === 'group' ? 'Format: <group-id>@g.us' : 'Phone number must include country code (e.g. 628...)' }}
            </p>
          </div>
        </div>

        <!-- JID Preview & Action Button Footer -->
        <div class="flex items-center justify-between pt-3 border-t border-subtle gap-3">
          <div class="flex items-center gap-2 min-w-0 truncate">
            <span class="text-[10px] font-mono uppercase text-text-secondary font-semibold shrink-0">JID Preview:</span>
            <code class="px-2 py-0.5 rounded bg-surface border border-subtle text-amber-400 font-mono text-[11px] font-medium truncate">
              {{ computedJID || (form.targetType === 'group' ? 'Enter Group JID ending in @g.us' : 'Enter phone number with country code') }}
            </code>
          </div>

          <button
            @click="handleSubmit"
            :disabled="isSubmitting || !form.label || !form.phoneNumber"
            class="px-4 py-2 rounded-lg bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white text-xs font-semibold shadow-sm transition-all disabled:opacity-50 flex items-center gap-1.5 shrink-0"
          >
            <Plus v-if="!editingId" class="w-3.5 h-3.5" />
            <Save v-else class="w-3.5 h-3.5" />
            {{ editingId ? 'Save Changes' : 'Add Target' }}
          </button>
        </div>

        <!-- Banner Error -->
        <div v-if="actionError" class="p-2 rounded bg-red-500/10 border border-red-500/30 text-[11px] text-red-400 font-mono">
          {{ actionError }}
        </div>
      </div>

      <!-- Banner Test Message -->
      <div v-if="testStatusMessage" class="p-2.5 rounded bg-card border border-subtle text-xs font-mono" :class="testStatusSuccess ? 'text-status-up' : 'text-status-down'">
        {{ testStatusMessage }}
      </div>

      <!-- Targets List -->
      <div class="space-y-2">
        <div class="flex items-center justify-between text-[11px] font-mono text-text-secondary px-1">
          <span>Configured Targets ({{ targets.length }})</span>
          <span v-if="isLoading" class="animate-pulse">Loading...</span>
        </div>

        <div v-if="targets.length === 0 && !isLoading" class="text-center py-6 border border-dashed border-subtle rounded-lg">
          <p class="text-xs text-text-secondary">No WhatsApp notification targets configured.</p>
          <p class="text-[10px] text-text-muted mt-1">Add a recipient number above to start receiving alerts.</p>
        </div>

        <div
          v-for="target in targets"
          :key="target.id"
              class="flex items-center justify-between bg-surface border border-subtle rounded-lg p-3 hover:border-brand-periwinkle/50 transition-colors"
        >
          <div class="space-y-0.5">
            <div class="flex items-center gap-2">
              <span class="text-xs font-semibold text-text-main">{{ target.label }}</span>
            </div>
            <div class="flex items-center gap-3 text-[10px] font-mono text-text-secondary">
              <span>{{ target.phoneNumber }}</span>
              <span class="text-text-muted">({{ target.jid }})</span>
            </div>
          </div>

          <div class="flex items-center gap-2">
            <button
              @click="handleTest(target.id)"
              :disabled="testingId === target.id"
              class="px-2.5 py-1 rounded-lg bg-brand-periwinkle/15 border border-brand-periwinkle/30 text-brand-periwinkle hover:bg-brand-periwinkle/25 text-[11px] font-medium transition-colors flex items-center gap-1 disabled:opacity-50"
              title="Send test message to this target"
            >
              <Send class="w-3 h-3" />
              {{ testingId === target.id ? 'Sending...' : 'Test' }}
            </button>

            <button
              @click="startEdit(target)"
              class="px-2.5 py-1 rounded-lg bg-subtle hover:bg-hover text-text-secondary text-[11px] font-medium transition-colors flex items-center gap-1"
            >
              <Pencil class="w-3 h-3" />
              Edit
            </button>

            <button
              @click="handleDelete(target.id)"
              class="px-2.5 py-1 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 hover:bg-red-500/20 text-[11px] font-medium transition-colors flex items-center gap-1"
            >
              <Trash2 class="w-3 h-3" />
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>
  </Modal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue';
import Modal from '../common/Modal.vue';
import { Plus, Save, Send, Pencil, Trash2, User, Users } from 'lucide-vue-next';
import api from '../../api/client';
import type { WhatsAppTarget } from '../../types';

const props = defineProps<{ isOpen: boolean }>();
const emit = defineEmits(['close']);

const targets = ref<WhatsAppTarget[]>([]);
const isLoading = ref(false);
const isSubmitting = ref(false);
const editingId = ref<string | null>(null);
const testingId = ref<string | null>(null);
const actionError = ref('');
const testStatusMessage = ref('');
const testStatusSuccess = ref(true);

const form = reactive({
  label: '',
  phoneNumber: '',
  targetType: 'individual' as 'individual' | 'group'
});

const computedJID = computed(() => {
  if (!form.phoneNumber) return '';
  const val = form.phoneNumber.trim();
  if (val.endsWith('@g.us') || val.endsWith('@s.whatsapp.net')) return val;
  let digits = val.replace(/[^0-9]/g, '');
  if (digits.startsWith('0') && digits.length > 1) {
    digits = '62' + digits.slice(1);
  }
  if (form.targetType === 'group') {
    return digits ? `${digits}@g.us` : '';
  }
  return digits ? `${digits}@s.whatsapp.net` : '';
});

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      fetchTargets();
      testStatusMessage.value = '';
      actionError.value = '';
    }
  }
);

async function fetchTargets() {
  isLoading.value = true;
  try {
    const res = await api.get('/integrations/whatsapp/targets');
    targets.value = res.data || [];
  } catch (e: any) {
    actionError.value = 'Failed to load targets';
  } finally {
    isLoading.value = false;
  }
}

function startEdit(target: WhatsAppTarget) {
  editingId.value = target.id;
  form.label = target.label;
  form.phoneNumber = target.phoneNumber;
  form.targetType = target.jid?.endsWith('@g.us') ? 'group' : 'individual';
  actionError.value = '';
}

function cancelEdit() {
  editingId.value = null;
  form.label = '';
  form.phoneNumber = '';
  actionError.value = '';
}

async function handleSubmit() {
  if (!form.label || !form.phoneNumber) return;
  actionError.value = '';
  isSubmitting.value = true;

  try {
    if (editingId.value) {
      await api.put(`/integrations/whatsapp/targets/${editingId.value}`, form);
    } else {
      await api.post('/integrations/whatsapp/targets', form);
    }
    cancelEdit();
    await fetchTargets();
  } catch (e: any) {
    actionError.value = e.response?.data?.error || 'Failed to save target';
  } finally {
    isSubmitting.value = false;
  }
}

async function handleDelete(id: string) {
  actionError.value = '';
  try {
    await api.delete(`/integrations/whatsapp/targets/${id}`);
    if (editingId.value === id) {
      cancelEdit();
    }
    await fetchTargets();
  } catch (e: any) {
    actionError.value = e.response?.data?.error || 'Failed to delete target';
  }
}

async function handleTest(id: string) {
  testingId.value = id;
  testStatusMessage.value = '';
  try {
    const res = await api.post('/integrations/whatsapp/test', { targetId: id });
    testStatusSuccess.value = res.data.success;
    testStatusMessage.value = res.data.message || 'Test message sent successfully!';
  } catch (e: any) {
    testStatusSuccess.value = false;
    testStatusMessage.value = e.response?.data?.error || 'Failed to send test message';
  } finally {
    testingId.value = null;
  }
}

onMounted(fetchTargets);
</script>
