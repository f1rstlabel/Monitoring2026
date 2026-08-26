import { defineStore } from 'pinia';
import { ref } from 'vue';
import { aiApi } from '../api';

export interface AIMessage {
  id: string;
  role: 'user' | 'assistant';
  text: string;
  timestamp: string;
  isError?: boolean;
}

export const useAIStore = defineStore('aiStore', () => {
  const isDrawerOpen = ref(false);

  const networkCondition = ref('healthy');
  const networkMessage = ref('');
  const isConfigured = ref(false);
  const aiModel = ref('gemini-1.5-flash');
  const isLoading = ref(false);
  const quickPrompts = ref<string[]>([]);
  const messages = ref<AIMessage[]>([
    {
      id: 'welcome-1',
      role: 'assistant',
      text: 'Halo! Saya **SANOC AI Copilot**, asisten cerdas pemantauan infrastruktur dan jaringan Sanditel Jabar. Anda dapat bertanya tentang status perangkat, akar masalah insiden (RCA), analisis latensi, maupun draf laporan pimpinan.',
      timestamp: new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
    }
  ]);

  const isAnalyzingIncident = ref(false);

  async function checkStatus() {
    try {
      const res = await aiApi.getStatus();
      isConfigured.value = res.configured;
      if (res.model) aiModel.value = res.model;
    } catch (e) {
      isConfigured.value = false;
    }
  }

  async function fetchQuickPrompts() {
    try {
      const res = await aiApi.getQuickPrompts();
      if (res && res.prompts) {
        quickPrompts.value = res.prompts;
      }
      if (res && res.condition) {
        networkCondition.value = res.condition || 'healthy';
        networkMessage.value = res.message || '';
      }
    } catch (e) {
      quickPrompts.value = [
        'Rangkum status jaringan hari ini',
        'Cek utilisasi CPU tertinggi',
        'Buatkan draf format laporan WA pimpinan'
      ];
    }
  }

  async function sendMessage(promptText: string) {
    const text = promptText.trim();
    if (!text || isLoading.value) return;

    const userMsgId = `user-${Date.now()}`;
    const userTimestamp = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });

    messages.value.push({
      id: userMsgId,
      role: 'user',
      text: text,
      timestamp: userTimestamp
    });

    isLoading.value = true;

    // Convert existing message history to Gemini API format (limit last 6 for efficiency)
    const history = messages.value
      .slice(-7, -1)
      .map(m => ({
        role: m.role === 'assistant' ? 'model' : 'user',
        content: m.text
      }));

    try {
      const res = await aiApi.chat(text, history);
      messages.value.push({
        id: `ai-${Date.now()}`,
        role: 'assistant',
        text: res.reply,
        timestamp: new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
      });
    } catch (err: any) {
      const errorMsg = err.response?.data?.error || err.message || 'Gagal menghubungi AI Copilot';
      messages.value.push({
        id: `err-${Date.now()}`,
        role: 'assistant',
        text: `⚠️ **Gagal memproses permintaan:**\n${errorMsg}`,
        timestamp: new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }),
        isError: true
      });
    } finally {
      isLoading.value = false;
    }
  }

  async function analyzeIncident(incidentId: string) {
    isDrawerOpen.value = true;
    isAnalyzingIncident.value = true;

    const userMsg = `Analisis akar masalah (RCA) dan panduan penanganan untuk insiden #${incidentId}`;
    messages.value.push({
      id: `user-${Date.now()}`,
      role: 'user',
      text: userMsg,
      timestamp: new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
    });

    isLoading.value = true;

    try {
      const res = await aiApi.analyzeIncident(incidentId);
      messages.value.push({
        id: `ai-${Date.now()}`,
        role: 'assistant',
        text: res.analysis,
        timestamp: new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
      });
    } catch (err: any) {
      const errorMsg = err.response?.data?.error || err.message || 'Gagal menganalisis insiden.';
      messages.value.push({
        id: `err-${Date.now()}`,
        role: 'assistant',
        text: `⚠️ **Gagal menganalisis insiden #${incidentId}:**\n${errorMsg}`,
        timestamp: new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }),
        isError: true
      });
    } finally {
      isLoading.value = false;
      isAnalyzingIncident.value = false;
    }
  }

  function clearMessages() {
    messages.value = [
      {
        id: 'welcome-reset',
        role: 'assistant',
        text: 'Percakapan telah direset. Silakan tanyakan hal lain terkait telemetri jaringan, root cause analysis, atau instruksi penanganan SANOC.',
        timestamp: new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
      }
    ];
  }

  return {
    isDrawerOpen,
    isConfigured,
    networkCondition,
    networkMessage,
    aiModel,
    isLoading,
    quickPrompts,
    messages,
    isAnalyzingIncident,
    checkStatus,
    fetchQuickPrompts,
    sendMessage,
    analyzeIncident,
    clearMessages
  };
});
