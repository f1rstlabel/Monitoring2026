<template>
  <Modal :is-open="isOpen" title="Configure Telegram Bot Integration" @close="$emit('close')">
    <template #icon>
      <Send class="w-5 h-5 text-[#7B96F5]" />
    </template>

    <form @submit.prevent="handleSave" class="space-y-4 text-xs">
      <!-- Bot Token -->
      <div class="space-y-1.5">
        <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium">Telegram Bot Token *</label>
        <div class="relative">
          <input
            v-model="form.botToken"
            :type="showToken ? 'text' : 'password'"
            required
            placeholder="e.g. 7129847123:AAH3k891k1zL0P921..."
            class="w-full bg-[#18181B] border border-[#26262A] rounded-lg pl-3 pr-10 py-2 font-mono text-gray-200 focus:outline-none focus:border-[#7B96F5]"
          />
          <button
            type="button"
            @click="showToken = !showToken"
            class="absolute right-2.5 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-200"
          >
            <Eye v-if="!showToken" class="w-4 h-4" />
            <EyeOff v-else class="w-4 h-4" />
          </button>
        </div>
        <p class="text-[10px] text-gray-500">Obtain your bot token from @BotFather on Telegram</p>
      </div>

      <!-- Chat / Channel ID -->
      <div class="space-y-1.5">
        <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium">Chat or Channel ID *</label>
        <input
          v-model="form.chatId"
          type="text"
          required
          placeholder="e.g. -1001982736412 or @GovMonitorAlerts"
          class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 font-mono text-gray-200 focus:outline-none focus:border-[#7B96F5]"
          :class="chatIdError ? 'border-[#F16565]' : ''"
        />
        <!-- Bug 1: inline validation error -->
        <p v-if="chatIdError" class="text-[10px] text-[#F16565] font-semibold">{{ chatIdError }}</p>
        <p class="text-[10px] text-gray-500">
          To get your chat ID: message <b>@userinfobot</b> on Telegram for a personal chat ID.
          For a group/channel, add your bot to it, then check <b>@RawDataBot</b> or use
          <code>getUpdates</code> to find the negative group ID (e.g., <b>-1001234567890</b>).
        </p>
        <p class="text-[10px] text-gray-500">Ensure the bot is added as an Administrator in your target channel</p>
      </div>

      <!-- Test Connection Button & Inline Result Banner -->
      <div class="pt-2">
        <div class="flex items-center justify-between">
          <button
            type="button"
            :disabled="isTesting || !form.botToken || !form.chatId"
            @click="handleTest"
            class="px-3.5 py-1.5 rounded-lg border border-[#7B96F5]/40 bg-[#7B96F5]/10 text-[#7B96F5] hover:bg-[#7B96F5]/20 font-medium text-xs transition-colors flex items-center gap-1.5 disabled:opacity-50"
          >
            <Send class="w-3.5 h-3.5" />
            {{ isTesting ? 'Testing...' : 'Test Connection' }}
          </button>
          <span v-if="testResult" class="text-[11px] font-mono" :class="testResult.success ? 'text-[#34D399]' : 'text-[#F16565]'">
            {{ testResult.success ? '✓ Verified' : '✗ Failed' }}
          </span>
        </div>

        <!-- Result Banner -->
        <div v-if="testResult" class="mt-2 p-3 rounded-lg border text-xs" :class="testResult.success ? 'bg-[#34D399]/10 border-[#34D399]/30 text-[#34D399]' : 'bg-[#F16565]/10 border-[#F16565]/30 text-[#F16565]'">
          <div class="flex items-start gap-2">
            <CheckCircle2 v-if="testResult.success" class="w-4 h-4 shrink-0 mt-0.5" />
            <AlertCircle v-else class="w-4 h-4 shrink-0 mt-0.5" />
            <div>
              <p class="font-bold">{{ testResult.title }}</p>
              <p class="text-[11px] mt-0.5 opacity-90">{{ testResult.message }}</p>
            </div>
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex items-center justify-end gap-3 w-full">
        <button
          type="button"
          @click="$emit('close')"
          class="px-4 py-2 rounded-lg border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-medium transition-colors"
        >
          Cancel
        </button>
        <button
          type="button"
          :disabled="isSaving || !form.botToken || !form.chatId"
          @click="handleSave"
          class="px-5 py-2 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 transition-all flex items-center gap-2 disabled:opacity-50"
        >
          <Save class="w-4 h-4" />
          {{ isSaving ? 'Saving...' : 'Save Configuration' }}
        </button>
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import Modal from '../common/Modal.vue';
import { Send, Eye, EyeOff, CheckCircle2, AlertCircle, Save } from 'lucide-vue-next';
import api from '../../api/client';

const props = defineProps<{
  isOpen: boolean;
  botToken?: string;
  chatId?: string;
}>();

const emit = defineEmits(['close', 'saved']);

const form = ref({
  botToken: '',
  chatId: ''
});

const showToken = ref(false);
const isTesting = ref(false);
const isSaving = ref(false);
const testResult = ref<{ success: boolean; title: string; message: string } | null>(null);

// Bug 1: computed validation — chatId must not equal the numeric prefix of botToken
const chatIdError = computed(() => {
  const { botToken, chatId } = form.value;
  if (!botToken || !chatId) return '';
  const parts = botToken.trim().split(':');
  if (parts.length < 2) return '';
  const botNumericId = parts[0];
  const cleaned = chatId.trim().replace(/[()[\]\s]/g, '');
  const stripped = cleaned.replace(/^-/, '');
  if (stripped === botNumericId) {
    return `Chat ID matches your bot's own ID (${botNumericId}). A bot cannot send messages to itself! Please enter your personal Telegram User ID (from @userinfobot) or a Group/Channel ID (e.g. -1001982736412).`;
  }
  return '';
});

// On modal open: load saved config from DB via GET /integrations/telegram/config
watch(() => props.isOpen, async (open) => {
  if (open) {
    testResult.value = null;
    form.value.botToken = props.botToken || '';
    form.value.chatId = props.chatId || '';
    try {
      const res = await api.get('/integrations/telegram/config');
      if (res.data.botToken) form.value.botToken = res.data.botToken;
      if (res.data.chatId) form.value.chatId = res.data.chatId;
    } catch {
      // offline — keep whatever is in props
    }
  }
});

async function handleTest() {
  if (!form.value.botToken || !form.value.chatId) return;
  const cleanedChatId = form.value.chatId.trim().replace(/[()[\]\s]/g, '');
  form.value.chatId = cleanedChatId;

  if (chatIdError.value) {
    testResult.value = {
      success: false,
      title: 'Invalid Chat ID',
      message: chatIdError.value
    };
    return;
  }

  isTesting.value = true;
  testResult.value = null;

  try {
    await api.post('/integrations/telegram/test', {
      botToken: form.value.botToken.trim(),
      chatId: cleanedChatId
    });
    testResult.value = {
      success: true,
      title: 'Test Message Sent Successfully',
      message: 'Check your Telegram channel/chat to confirm arrival of the test notification.'
    };
  } catch (e: any) {
    const errMsg = e.response?.data?.error || 'Failed to communicate with Telegram Bot API';
    testResult.value = {
      success: false,
      title: 'Telegram Test Failed',
      message: errMsg
    };
  } finally {
    isTesting.value = false;
  }
}

async function handleSave() {
  if (!form.value.botToken || !form.value.chatId) return;
  const cleanedChatId = form.value.chatId.trim().replace(/[()[\]\s]/g, '');
  form.value.chatId = cleanedChatId;

  if (chatIdError.value) {
    testResult.value = {
      success: false,
      title: 'Invalid Chat ID',
      message: chatIdError.value
    };
    return;
  }

  isSaving.value = true;

  try {
    await api.post('/integrations/telegram/config', {
      botToken: form.value.botToken.trim(),
      chatId: cleanedChatId
    });
    emit('saved', { botToken: form.value.botToken.trim(), chatId: cleanedChatId });
    emit('close');
  } catch (e: any) {
    const errMsg = e.response?.data?.error || 'Failed to save Telegram configuration.';
    testResult.value = {
      success: false,
      title: 'Save Failed',
      message: errMsg
    };
  } finally {
    isSaving.value = false;
  }
}
</script>
