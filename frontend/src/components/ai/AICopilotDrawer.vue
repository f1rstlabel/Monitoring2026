<template>
  <div>
    <!-- Floating Copilot Trigger Button (Bottom-Right) -->
    <div class="fixed bottom-6 right-6 z-40 select-none">
      <button
        @click="toggleDrawer"
        class="group relative flex items-center gap-2.5 px-4 py-3 rounded-full bg-gradient-to-r from-[#18181B] to-[#202025] hover:from-[#202025] hover:to-[#2A2A32] border border-[#7B96F5]/40 text-white font-medium text-xs shadow-2xl shadow-[#7B96F5]/25 transition-all duration-300 hover:scale-105 active:scale-95 cursor-pointer"
        title="Open SANOC AI Copilot"
      >
        <!-- Ambient Glow Aura -->
        <span class="absolute -inset-1 rounded-full bg-gradient-to-r from-[#7B96F5]/30 to-[#3ECF8E]/20 blur-md opacity-75 group-hover:opacity-100 transition-opacity -z-10 animate-pulse"></span>

        <!-- AI Sparkles Icon -->
        <div class="w-6 h-6 rounded-full bg-[#7B96F5]/20 border border-[#7B96F5]/40 flex items-center justify-center text-[#7B96F5] shrink-0">
          <Sparkles class="w-3.5 h-3.5 animate-spin-slow" />
        </div>

        <span class="font-bold tracking-wide">AI Copilot</span>
        
        <!-- Live status pip -->
        <span
          class="w-2 h-2 rounded-full ring-2 ring-[#0D0D0F]"
          :class="aiStore.isConfigured ? 'bg-[#3ECF8E] animate-pulse' : 'bg-amber-400'"
        ></span>
      </button>
    </div>

    <!-- Backdrop Overlay on Mobile -->
    <div
      v-if="aiStore.isDrawerOpen"
      @click="aiStore.isDrawerOpen = false"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 md:hidden animate-fade-in"
    ></div>

    <!-- Slide-Over Drawer Panel -->
    <div
      v-if="aiStore.isDrawerOpen"
      class="fixed bottom-0 right-0 top-0 w-full sm:w-[460px] md:w-[500px] bg-[#0E0E11] border-l border-[#26262A] shadow-2xl z-50 flex flex-col justify-between animate-slide-left overflow-hidden select-none"
    >
      <!-- Drawer Header -->
      <div class="px-5 py-4 border-b border-[#26262A] bg-[#141418] flex items-center justify-between shrink-0">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-[#7B96F5]/20 to-[#3ECF8E]/20 border border-[#7B96F5]/40 flex items-center justify-center text-[#7B96F5] shadow-sm">
            <Sparkles class="w-4 h-4" />
          </div>
          <div>
            <div class="flex items-center gap-2">
              <h3 class="text-sm font-bold text-white font-mono">SANOC AI Copilot</h3>
              <span class="text-[9px] font-mono font-semibold px-1.5 py-0.5 rounded bg-[#7B96F5]/15 text-[#7B96F5] border border-[#7B96F5]/30">
                {{ aiStore.aiModel }}
              </span>
            </div>
            <p class="text-[10px] text-gray-400 flex items-center gap-1.5 mt-0.5">
              <span
                class="w-1.5 h-1.5 rounded-full inline-block"
                :class="aiStore.isConfigured ? 'bg-[#3ECF8E] pulsing-dot-green' : 'bg-amber-400'"
              ></span>
              <span>{{ aiStore.isConfigured ? 'Live Network Telemetry Connected' : 'Awaiting GEMINI_API_KEY in .env' }}</span>
            </p>
          </div>
        </div>

        <!-- Controls: Clear & Close -->
        <div class="flex items-center gap-1">
          <button
            @click="aiStore.clearMessages()"
            class="p-1.5 rounded-lg text-gray-400 hover:text-gray-200 hover:bg-[#202026] transition-colors cursor-pointer"
            title="Clear Chat History"
          >
            <RotateCcw class="w-3.5 h-3.5" />
          </button>
          <button
            @click="aiStore.isDrawerOpen = false"
            class="p-1.5 rounded-lg text-gray-400 hover:text-red-400 hover:bg-red-500/10 transition-colors cursor-pointer"
            title="Close Drawer"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- Drawer Body: Message Stream & Context -->
      <div ref="chatContainerRef" class="flex-1 overflow-y-auto p-4 space-y-4 select-text">
        <!-- Unconfigured Warning Banner -->
        <div
          v-if="!aiStore.isConfigured"
          class="p-4 rounded-xl bg-amber-500/10 border border-amber-500/30 text-xs text-amber-300 space-y-2 select-none"
        >
          <div class="flex items-center gap-2 font-bold font-mono text-[11px] text-amber-400">
            <AlertCircle class="w-4 h-4" />
            <span>KUNCI API BELUM TERKONFIGURASI</span>
          </div>
          <p class="text-[11px] text-gray-300 leading-relaxed">
            Untuk mengaktifkan AI Copilot, tambahkan baris berikut pada file <code class="bg-[#151517] px-1.5 py-0.5 rounded text-amber-400 font-mono">backend/.env</code>:
          </p>
          <div class="bg-[#0A0A0B] p-2.5 rounded-lg border border-[#26262A] font-mono text-[11px] text-gray-200 flex items-center justify-between">
            <span>GEMINI_API_KEY=AIzaSy...</span>
            <button
              @click="copyText('GEMINI_API_KEY=your_gemini_api_key_here')"
              class="text-[10px] text-[#7B96F5] hover:underline cursor-pointer"
            >
              Salin
            </button>
          </div>
          <p class="text-[10px] text-gray-400">
            Dapatkan API Key gratis di <a href="https://aistudio.google.com/app/apikey" target="_blank" class="text-[#7B96F5] underline">Google AI Studio</a>.
          </p>
        </div>

        <!-- Chat Messages -->
        <div
          v-for="msg in aiStore.messages"
          :key="msg.id"
          class="flex flex-col space-y-1.5"
          :class="msg.role === 'user' ? 'items-end' : 'items-start'"
        >
          <!-- Role Label & Timestamp -->
          <div class="flex items-center gap-2 px-1 text-[10px] font-mono text-gray-500 select-none">
            <span v-if="msg.role === 'assistant'" class="font-bold text-[#7B96F5] flex items-center gap-1">
              <Sparkles class="w-2.5 h-2.5" /> SANOC AI
            </span>
            <span v-else class="font-bold text-gray-300">You</span>
            <span>&bull;</span>
            <span>{{ msg.timestamp }}</span>
          </div>

          <!-- Message Bubble -->
          <div
            class="max-w-[92%] rounded-2xl p-3.5 text-xs leading-relaxed transition-all shadow-md relative group"
            :class="msg.role === 'user' 
              ? 'bg-gradient-to-r from-[#7B96F5]/20 to-[#6070D0]/20 border border-[#7B96F5]/40 text-gray-100 rounded-tr-sm' 
              : msg.isError 
                ? 'bg-red-500/10 border border-red-500/30 text-red-200 rounded-tl-sm' 
                : 'bg-[#151518] border border-[#26262A] text-gray-200 rounded-tl-sm'"
          >
            <!-- Formatted Markdown Content -->
            <div class="prose prose-invert max-w-none text-xs space-y-2 whitespace-pre-wrap font-sans break-words" v-html="formatMarkdown(msg.text)"></div>

            <!-- Quick Copy Button on Assistant Message -->
            <button
              v-if="msg.role === 'assistant' && !msg.isError"
              @click="copyText(msg.text)"
              class="absolute top-2 right-2 p-1 rounded bg-[#1C1C22] border border-[#26262A] text-gray-400 hover:text-white opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
              title="Salin Pesan"
            >
              <Copy class="w-3 h-3" />
            </button>
          </div>
        </div>

        <!-- AI Thinking / Streaming Indicator -->
        <div v-if="aiStore.isLoading" class="flex items-center gap-2 p-3 bg-[#151518] border border-[#26262A] rounded-2xl rounded-tl-sm max-w-[85%] text-xs text-gray-400">
          <div class="flex gap-1 items-center">
            <span class="w-1.5 h-1.5 rounded-full bg-[#7B96F5] animate-bounce"></span>
            <span class="w-1.5 h-1.5 rounded-full bg-[#7B96F5] animate-bounce [animation-delay:0.2s]"></span>
            <span class="w-1.5 h-1.5 rounded-full bg-[#7B96F5] animate-bounce [animation-delay:0.4s]"></span>
          </div>
          <span class="font-mono text-[11px] text-gray-400">Menganalisis telemetri & menyusun respons...</span>
        </div>
      </div>

      <!-- Quick Suggestion Chips (Above Input) -->
      <div v-if="aiStore.isConfigured" class="px-4 py-2 border-t border-[#26262A]/60 bg-[#121215] flex items-center gap-1.5 overflow-x-auto no-scrollbar select-none">
        <button
          v-for="chip in aiStore.quickPrompts"
          :key="chip"
          @click="submitPrompt(chip)"
          :disabled="aiStore.isLoading"
          class="shrink-0 px-2.5 py-1 rounded-full bg-[#18181D] hover:bg-[#202028] border border-[#26262A] hover:border-[#7B96F5]/40 text-[10px] font-mono text-gray-300 hover:text-white transition-all cursor-pointer disabled:opacity-50"
        >
          💡 {{ chip }}
        </button>
      </div>

      <!-- Drawer Footer: Input Form -->
      <div class="p-4 border-t border-[#26262A] bg-[#141418] shrink-0 space-y-2 select-none">
        <form @submit.prevent="handleSend" class="relative flex items-center gap-2">
          <input
            v-model="inputPrompt"
            type="text"
            placeholder="Tanyakan status perangkat, RCA insiden, atau draf laporan..."
            :disabled="aiStore.isLoading"
            class="flex-1 bg-[#18181D] border border-[#26262A] focus:border-[#7B96F5] rounded-xl pl-3.5 pr-10 py-2.5 text-xs text-gray-200 placeholder-gray-500 focus:outline-none transition-colors select-text"
          />
          <button
            type="submit"
            :disabled="!inputPrompt.trim() || aiStore.isLoading"
            class="p-2.5 rounded-xl bg-[#7B96F5] hover:bg-[#6885EB] disabled:opacity-40 disabled:hover:bg-[#7B96F5] text-white transition-all cursor-pointer shrink-0 shadow-md shadow-[#7B96F5]/20"
            title="Kirim Pesan"
          >
            <Send class="w-3.5 h-3.5" />
          </button>
        </form>
        <p class="text-[9px] font-mono text-gray-500 text-center">
          AI Copilot menganalisis data langsung dari telemetri ICMP, SNMP, & Incidents SANOC.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, watch } from 'vue';
import { Sparkles, X, RotateCcw, Send, Copy, AlertCircle } from 'lucide-vue-next';
import { useAIStore } from '../../stores/aiStore';

const aiStore = useAIStore();
const inputPrompt = ref('');
const chatContainerRef = ref<HTMLDivElement | null>(null);

function toggleDrawer() {
  aiStore.isDrawerOpen = !aiStore.isDrawerOpen;
  if (aiStore.isDrawerOpen) {
    aiStore.checkStatus();
    aiStore.fetchQuickPrompts();
    scrollToBottom();
  }
}

async function handleSend() {
  const text = inputPrompt.value.trim();
  if (!text || aiStore.isLoading) return;
  inputPrompt.value = '';
  await aiStore.sendMessage(text);
  scrollToBottom();
}

async function submitPrompt(chipText: string) {
  if (aiStore.isLoading) return;
  await aiStore.sendMessage(chipText);
  scrollToBottom();
}

function scrollToBottom() {
  nextTick(() => {
    if (chatContainerRef.value) {
      chatContainerRef.value.scrollTop = chatContainerRef.value.scrollHeight;
    }
  });
}

watch(
  () => aiStore.messages.length,
  () => {
    scrollToBottom();
  }
);

function copyText(text: string) {
  navigator.clipboard.writeText(text);
}

// Rich structured Markdown Formatter for AI chat bubbles
function formatMarkdown(text: string): string {
  if (!text) return '';
  let str = text;

  // 1. Code blocks ```lang\ncode\n```
  const codeBlocks: string[] = [];
  str = str.replace(/```([a-zA-Z0-9_-]*)\n([\s\S]*?)```/g, (_, lang, code) => {
    const idx = codeBlocks.length;
    const escapedCode = code
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
    const langLabel = lang
      ? `<div class="text-[10px] uppercase font-mono text-gray-400 pb-1 border-b border-[#26262A] mb-1.5 flex justify-between"><span>${lang}</span><span>code snippet</span></div>`
      : '';
    codeBlocks.push(
      `<div class="my-2 rounded-lg bg-[#0D0D10] border border-[#26262A] p-2.5 font-mono text-[11px] text-[#3ECF8E] overflow-x-auto select-text">${langLabel}<pre class="m-0 whitespace-pre font-mono"><code>${escapedCode}</code></pre></div>`
    );
    return `__CODE_BLOCK_${idx}__`;
  });

  // 2. Escape basic HTML for the rest
  str = str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');

  // 3. Headers
  str = str.replace(/^### (.*$)/gim, '<h4 class="text-xs font-bold text-white mt-3 mb-1.5 flex items-center gap-1.5 font-mono text-[#7B96F5]">$1</h4>');
  str = str.replace(/^## (.*$)/gim, '<h3 class="text-sm font-bold text-white mt-3.5 mb-1.5 font-mono border-b border-[#26262A] pb-1">$1</h3>');
  str = str.replace(/^# (.*$)/gim, '<h2 class="text-base font-bold text-white mt-4 mb-2 font-mono">$1</h2>');

  // 4. Horizontal rule
  str = str.replace(/^---$/gim, '<hr class="my-2.5 border-[#26262A]"/>');

  // 5. Bold & Italic
  str = str.replace(/\*\*\*(.*?)\*\*\*/g, '<strong class="font-bold italic text-white">$1</strong>');
  str = str.replace(/\*\*(.*?)\*\*/g, '<strong class="font-bold text-white">$1</strong>');
  str = str.replace(/\*(.*?)\*/g, '<em class="italic text-gray-300">$1</em>');

  // 6. Inline code `code`
  str = str.replace(/`([^`]+)`/g, '<code class="px-1.5 py-0.5 rounded bg-[#1C1C22] border border-[#2A2A32] text-[#7B96F5] font-mono text-[11px]">$1</code>');

  // 7. Blockquotes
  str = str.replace(/^> (.*$)/gim, '<blockquote class="border-l-2 border-[#7B96F5] pl-2.5 my-1.5 text-gray-400 italic text-[11px]">$1</blockquote>');

  // 8. Numbered lists (1. 2. 3.)
  str = str.replace(/^(\d+)\.\s+(.*)$/gim, '<div class="flex items-start gap-2 my-1 text-gray-200"><span class="font-mono text-[10px] font-bold text-[#7B96F5] bg-[#7B96F5]/10 px-1.5 py-0.5 rounded shrink-0 mt-0.5">$1</span><span>$2</span></div>');

  // 9. Bullet lists (- or • or *)
  str = str.replace(/^[\s]*[-•*]\s+(.*)$/gim, '<div class="flex items-start gap-2 my-1 text-gray-300"><span class="text-[#7B96F5] text-xs shrink-0 leading-tight">•</span><span>$1</span></div>');

  // 10. Paragraphs / Line breaks
  str = str.replace(/\n\n/g, '<div class="h-2"></div>');
  str = str.replace(/\n/g, '<br/>');

  // 11. Restore code blocks
  str = str.replace(/__CODE_BLOCK_(\d+)__/g, (_, idx) => {
    return codeBlocks[parseInt(idx, 10)] || '';
  });

  return str;
}

onMounted(() => {
  aiStore.checkStatus();
});
</script>

<style scoped>
@keyframes slide-left {
  from {
    transform: translateX(100%);
  }
  to {
    transform: translateX(0);
  }
}

@keyframes fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.animate-slide-left {
  animation: slide-left 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.animate-fade-in {
  animation: fade-in 0.2s ease-out;
}

.animate-spin-slow {
  animation: spin 8s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
