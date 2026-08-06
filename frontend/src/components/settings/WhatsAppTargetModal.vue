<template>
  <Modal :is-open="isOpen" title="Configure WhatsApp Notification Targets" @close="$emit('close')">
    <div class="space-y-4">
      <!-- Add / Edit Target Form -->
      <div class="bg-[#18181B] border border-[#26262A] rounded-lg p-3.5 space-y-3">
        <h4 class="text-xs font-semibold text-white flex items-center justify-between">
          <span>{{ editingId ? 'Edit Target' : 'Add New Target' }}</span>
          <button v-if="editingId" @click="cancelEdit" class="text-[11px] text-gray-400 hover:text-white font-normal">
            Cancel Edit
          </button>
        </h4>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <div>
            <label class="block font-mono uppercase text-[9px] text-gray-400 font-medium mb-1">Target Label</label>
            <input
              v-model="form.label"
              type="text"
              placeholder="e.g. NOC Group / Admin Backup"
              class="w-full bg-[#151517] border border-[#26262A] rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
          <div>
            <label class="block font-mono uppercase text-[9px] text-gray-400 font-medium mb-1">Phone Number</label>
            <input
              v-model="form.phoneNumber"
              type="text"
              placeholder="e.g. +6281234567890"
              class="w-full bg-[#151517] border border-[#26262A] rounded-lg px-3 py-1.5 text-xs text-white font-mono focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
        </div>

        <div class="flex items-center justify-between pt-1">
          <p class="text-[10px] text-gray-400 font-mono">
            JID preview: {{ computedJID || 'Enter phone number with country code' }}
          </p>
          <button
            @click="handleSubmit"
            :disabled="isSubmitting || !form.label || !form.phoneNumber"
            class="px-3.5 py-1.5 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-semibold shadow-sm transition-all disabled:opacity-50 flex items-center gap-1.5"
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
      <div v-if="testStatusMessage" class="p-2.5 rounded bg-[#18181B] border border-[#26262A] text-xs font-mono" :class="testStatusSuccess ? 'text-[#34D399]' : 'text-[#F16565]'">
        {{ testStatusMessage }}
      </div>

      <!-- Targets List -->
      <div class="space-y-2">
        <div class="flex items-center justify-between text-[11px] font-mono text-gray-400 px-1">
          <span>Configured Targets ({{ targets.length }})</span>
          <span v-if="isLoading" class="animate-pulse">Loading...</span>
        </div>

        <div v-if="targets.length === 0 && !isLoading" class="text-center py-6 border border-dashed border-[#26262A] rounded-lg">
          <p class="text-xs text-gray-400">No WhatsApp notification targets configured.</p>
          <p class="text-[10px] text-gray-500 mt-1">Add a recipient number above to start receiving alerts.</p>
        </div>

        <div
          v-for="target in targets"
          :key="target.id"
          class="flex items-center justify-between bg-[#151517] border border-[#26262A] rounded-lg p-3 hover:border-[#3A3A40] transition-colors"
        >
          <div class="space-y-0.5">
            <div class="flex items-center gap-2">
              <span class="text-xs font-semibold text-white">{{ target.label }}</span>
            </div>
            <div class="flex items-center gap-3 text-[10px] font-mono text-gray-400">
              <span>{{ target.phoneNumber }}</span>
              <span class="text-gray-500">({{ target.jid }})</span>
            </div>
          </div>

          <div class="flex items-center gap-2">
            <button
              @click="handleTest(target.id)"
              :disabled="testingId === target.id"
              class="px-2.5 py-1 rounded-lg bg-[#7B96F5]/15 border border-[#7B96F5]/30 text-[#7B96F5] hover:bg-[#7B96F5]/25 text-[11px] font-medium transition-colors flex items-center gap-1 disabled:opacity-50"
              title="Send test message to this target"
            >
              <Send class="w-3 h-3" />
              {{ testingId === target.id ? 'Sending...' : 'Test' }}
            </button>

            <button
              @click="startEdit(target)"
              class="px-2.5 py-1 rounded-lg bg-[#26262A] hover:bg-[#323238] text-gray-300 text-[11px] font-medium transition-colors flex items-center gap-1"
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
import { Plus, Save, Send, Pencil, Trash2 } from 'lucide-vue-next';
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
  phoneNumber: ''
});

const computedJID = computed(() => {
  if (!form.phoneNumber) return '';
  let digits = form.phoneNumber.replace(/[^0-9]/g, '');
  if (digits.startsWith('0') && digits.length > 1) {
    digits = '62' + digits.slice(1);
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
