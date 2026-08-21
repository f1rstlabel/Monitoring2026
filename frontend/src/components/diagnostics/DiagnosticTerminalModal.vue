<template>
  <Modal :is-open="isOpen" title="Network Diagnostic Terminal" @close="$emit('close')">
    <template #default>
      <div class="space-y-4">
        <!-- Mode Tabs & Target Input Bar -->
        <div class="p-3 bg-[#18181B] border border-[#26262A] rounded-xl space-y-3">
          <!-- Diagnostics Mode Tabs -->
          <div class="flex flex-wrap items-center gap-1.5 p-1 bg-[#111113] rounded-lg border border-[#26262A]">
            <button
              v-for="mode in modes"
              :key="mode.id"
              type="button"
              @click="activeMode = mode.id"
              class="flex-1 py-1.5 px-3 rounded-md text-xs font-mono font-medium transition-all flex items-center justify-center gap-2 cursor-pointer"
              :class="[
                activeMode === mode.id
                  ? 'bg-[#7B96F5] text-white shadow-sm font-bold'
                  : 'text-gray-400 hover:text-gray-200 hover:bg-[#18181B]'
              ]"
            >
              <component :is="mode.icon" class="w-3.5 h-3.5" />
              <span>{{ mode.label }}</span>
            </button>
          </div>

          <!-- Target & Options Input Grid -->
          <div class="grid grid-cols-1 sm:grid-cols-12 gap-2.5 items-end">
            <!-- Target Input + Device Selector -->
            <div :class="activeMode === 'port' ? 'sm:col-span-6' : (activeMode === 'ping' ? 'sm:col-span-7' : 'sm:col-span-9')" class="space-y-1">
              <div class="flex items-center justify-between">
                <label class="font-mono uppercase text-[10px] text-gray-400 font-semibold">Target Host / IP</label>
                <button
                  v-if="deviceStore.devices.length > 0"
                  type="button"
                  @click="showDeviceDropdown = !showDeviceDropdown"
                  class="text-[10px] text-[#7B96F5] hover:underline font-mono flex items-center gap-1 cursor-pointer"
                >
                  <Server class="w-3 h-3" />
                  <span>Pick Monitored Device</span>
                </button>
              </div>

              <div class="relative">
                <input
                  v-model="targetInput"
                  type="text"
                  placeholder="e.g. 10.20.1.18 or 1.1.1.1"
                  class="w-full bg-[#111113] border border-[#26262A] rounded-lg px-3 py-2 text-xs font-mono text-white focus:outline-none focus:border-[#7B96F5]"
                  @keydown.enter="runDiagnostic"
                />

                <!-- Quick Pick Device Dropdown -->
                <div
                  v-if="showDeviceDropdown"
                  class="absolute left-0 right-0 top-full mt-1 bg-[#18181B] border border-[#26262A] rounded-xl shadow-2xl max-h-48 overflow-y-auto z-50 p-1.5 space-y-1"
                >
                  <div
                    v-for="dev in deviceStore.devices"
                    :key="dev.id"
                    @click="selectDevice(dev)"
                    class="p-2 rounded-lg hover:bg-[#26262A] text-xs flex items-center justify-between cursor-pointer transition-colors"
                  >
                    <div class="flex items-center gap-2">
                      <span
                        class="w-2 h-2 rounded-full"
                        :class="dev.status === 'UP' ? 'bg-[#3ECF8E]' : 'bg-red-500'"
                      ></span>
                      <span class="font-bold text-white">{{ dev.name }}</span>
                    </div>
                    <span class="font-mono text-[11px] text-gray-400">{{ dev.ip || dev.mac }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Ping Count Option -->
            <div v-if="activeMode === 'ping'" class="sm:col-span-2 space-y-1">
              <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Packets</label>
              <select
                v-model="pingCount"
                class="w-full bg-[#111113] border border-[#26262A] rounded-lg px-2.5 py-2 text-xs font-mono text-white focus:outline-none focus:border-[#7B96F5]"
              >
                <option :value="4">4 Echo</option>
                <option :value="8">8 Echo</option>
                <option :value="10">10 Echo</option>
              </select>
            </div>

            <!-- Port Option -->
            <div v-if="activeMode === 'port'" class="sm:col-span-3 space-y-1">
              <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Port (TCP)</label>
              <input
                v-model.number="portInput"
                type="number"
                min="1"
                max="65535"
                placeholder="80"
                class="w-full bg-[#111113] border border-[#26262A] rounded-lg px-3 py-2 text-xs font-mono text-white focus:outline-none focus:border-[#7B96F5]"
                @keydown.enter="runDiagnostic"
              />
            </div>

            <!-- Execute Button -->
            <div class="sm:col-span-3">
              <button
                type="button"
                @click="runDiagnostic"
                :disabled="isRunning || !targetInput.trim()"
                class="w-full py-2 px-4 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs font-mono shadow-md shadow-[#7B96F5]/20 flex items-center justify-center gap-2 disabled:opacity-50 transition-all cursor-pointer"
              >
                <RefreshCw v-if="isRunning" class="w-3.5 h-3.5 animate-spin" />
                <Play v-else class="w-3.5 h-3.5" />
                <span>{{ isRunning ? 'Running...' : 'Execute Test' }}</span>
              </button>
            </div>
          </div>

          <!-- Quick Port Presets (Port Mode only) -->
          <div v-if="activeMode === 'port'" class="flex items-center gap-2 pt-1">
            <span class="text-[10px] font-mono text-gray-500 uppercase">Quick Ports:</span>
            <div class="flex flex-wrap gap-1.5">
              <button
                v-for="p in [80, 443, 22, 161, 8080, 53]"
                :key="p"
                type="button"
                @click="portInput = p"
                class="px-2 py-0.5 rounded text-[10px] font-mono border transition-colors cursor-pointer"
                :class="portInput === p ? 'bg-[#7B96F5]/20 border-[#7B96F5] text-[#7B96F5]' : 'bg-[#111113] border-[#26262A] text-gray-400 hover:text-white'"
              >
                {{ p }} ({{ getPortLabel(p) }})
              </button>
            </div>
          </div>
        </div>

        <!-- Terminal Console Display -->
        <div class="rounded-xl border border-[#26262A] overflow-hidden bg-[#0A0A0C] shadow-2xl">
          <!-- Terminal Titlebar -->
          <div class="bg-[#151517] px-4 py-2.5 border-b border-[#26262A] flex items-center justify-between select-none">
            <div class="flex items-center gap-2">
              <div class="flex items-center gap-1.5">
                <span class="w-2.5 h-2.5 rounded-full bg-red-500/80 inline-block"></span>
                <span class="w-2.5 h-2.5 rounded-full bg-amber-500/80 inline-block"></span>
                <span class="w-2.5 h-2.5 rounded-full bg-emerald-500/80 inline-block"></span>
              </div>
              <span class="text-[11px] font-mono text-gray-400 ml-2">sanoc-diagnostic-terminal — bash</span>
            </div>

            <!-- Terminal Actions -->
            <div class="flex items-center gap-2">
              <button
                v-if="consoleLines.length > 0"
                type="button"
                @click="copyOutput"
                class="px-2 py-1 rounded bg-[#26262A] hover:bg-[#323238] text-gray-300 text-[10px] font-mono flex items-center gap-1 transition-colors cursor-pointer"
                title="Copy output to clipboard"
              >
                <Check v-if="copied" class="w-3 h-3 text-emerald-400" />
                <Copy v-else class="w-3 h-3" />
                <span>{{ copied ? 'Copied!' : 'Copy' }}</span>
              </button>

              <button
                v-if="consoleLines.length > 0"
                type="button"
                @click="clearConsole"
                class="px-2 py-1 rounded bg-[#26262A] hover:bg-red-500/20 hover:text-red-400 text-gray-400 text-[10px] font-mono flex items-center gap-1 transition-colors cursor-pointer"
                title="Clear console"
              >
                <Trash2 class="w-3 h-3" />
                <span>Clear</span>
              </button>
            </div>
          </div>

          <!-- Terminal Content Area -->
          <div
            ref="terminalRef"
            class="p-4 font-mono text-[11px] leading-relaxed h-72 overflow-y-auto space-y-1 select-text scrollbar-thin scrollbar-thumb-[#26262A]"
          >
            <div v-if="consoleLines.length === 0" class="text-gray-600 italic">
              SANOC Network Diagnostics Console v2.6.0<br />
              Enter an IP address or hostname above and click "Execute Test" to initiate live ICMP ping, traceroute, or TCP socket checks from the monitoring server.
            </div>

            <div
              v-for="(line, idx) in consoleLines"
              :key="idx"
              :class="getLineClass(line)"
              class="break-all"
            >
              {{ line }}
            </div>

            <!-- Running indicator cursor -->
            <div v-if="isRunning" class="text-[#7B96F5] flex items-center gap-2 pt-1 animate-pulse">
              <span>●</span>
              <span>Executing test on server...</span>
            </div>
          </div>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="flex items-center justify-between w-full">
        <span class="text-[10px] font-mono text-gray-500">
          Source: Monitoring Server Gateway (Local OS Execution)
        </span>
        <button
          type="button"
          @click="$emit('close')"
          class="px-4 py-2 rounded-xl bg-[#26262A] hover:bg-[#323238] text-gray-200 text-xs font-mono cursor-pointer"
        >
          Close Terminal
        </button>
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue';
import Modal from '../common/Modal.vue';
import {
  Activity,
  GitCommit,
  Radio,
  Server,
  Play,
  RefreshCw,
  Copy,
  Trash2,
  Check
} from 'lucide-vue-next';
import { diagnosticsApi } from '../../api';
import { useDeviceStore } from '../../stores/deviceStore';
import type { Device } from '../../types';

const props = defineProps<{
  isOpen: boolean;
  initialTarget?: string;
  initialMode?: 'ping' | 'traceroute' | 'port';
}>();

const emit = defineEmits(['close']);
const deviceStore = useDeviceStore();

const modes = [
  { id: 'ping', label: 'Ping (ICMP)', icon: Activity },
  { id: 'traceroute', label: 'Traceroute (Hops)', icon: GitCommit },
  { id: 'port', label: 'Port Probe (TCP)', icon: Radio }
] as const;

type DiagnosticMode = typeof modes[number]['id'];

const activeMode = ref<DiagnosticMode>('ping');
const targetInput = ref('');
const pingCount = ref(4);
const portInput = ref(80);
const isRunning = ref(false);
const showDeviceDropdown = ref(false);
const copied = ref(false);
const consoleLines = ref<string[]>([]);
const terminalRef = ref<HTMLDivElement | null>(null);

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      if (props.initialTarget) {
        targetInput.value = props.initialTarget;
      }
      if (props.initialMode) {
        activeMode.value = props.initialMode;
      }
    }
  },
  { immediate: true }
);

function selectDevice(dev: Device) {
  targetInput.value = dev.ip || dev.mac;
  showDeviceDropdown.value = false;
}

function getPortLabel(p: number): string {
  const map: Record<number, string> = {
    80: 'HTTP',
    443: 'HTTPS',
    22: 'SSH',
    161: 'SNMP',
    8080: 'WebGUI',
    53: 'DNS'
  };
  return map[p] || 'TCP';
}

function clearConsole() {
  consoleLines.value = [];
}

async function copyOutput() {
  if (consoleLines.value.length === 0) return;
  const text = consoleLines.value.join('\n');
  try {
    await navigator.clipboard.writeText(text);
    copied.value = true;
    setTimeout(() => {
      copied.value = false;
    }, 2000);
  } catch (e) {
    // fallback
  }
}

function getLineClass(line: string): string {
  const lower = line.toLowerCase();
  if (lower.startsWith('$ ') || lower.startsWith('---')) {
    return 'text-gray-400 font-bold';
  }
  if (lower.includes('reply from') || lower.includes('bytes from') || lower.includes('is open') || lower.includes('0% packet loss') || lower.includes('0% loss')) {
    return 'text-emerald-400 font-medium';
  }
  if (lower.includes('timed out') || lower.includes('unreachable') || lower.includes('closed') || lower.includes('100% loss') || lower.includes('100% packet loss') || lower.includes('failed')) {
    return 'text-red-400 font-semibold';
  }
  if (lower.includes('packets: sent') || lower.includes('approximate round trip') || lower.includes('trace complete') || lower.includes('ms')) {
    return 'text-cyan-300';
  }
  return 'text-gray-300';
}

async function scrollToBottom() {
  await nextTick();
  if (terminalRef.value) {
    terminalRef.value.scrollTop = terminalRef.value.scrollHeight;
  }
}

async function runDiagnostic() {
  const target = targetInput.value.trim();
  if (!target || isRunning.value) return;

  showDeviceDropdown.value = false;
  isRunning.value = true;

  const timestamp = new Date().toLocaleTimeString('id-ID', { hour12: false });

  if (activeMode.value === 'ping') {
    consoleLines.value.push(`$ [${timestamp}] ping -n ${pingCount.value} ${target}`);
    await scrollToBottom();

    try {
      const res = await diagnosticsApi.runPing(target, pingCount.value);
      if (res && res.output) {
        for (const line of res.output) {
          consoleLines.value.push(line);
        }
      } else if (res && res.raw) {
        consoleLines.value.push(res.raw);
      }
      consoleLines.value.push(`--- Ping completed in ${res?.durationMs ?? 0}ms ---\n`);
    } catch (e: any) {
      consoleLines.value.push(`Error executing ping: ${e.response?.data?.error || e.message}`);
    }
  } else if (activeMode.value === 'traceroute') {
    consoleLines.value.push(`$ [${timestamp}] traceroute -h 15 ${target}`);
    consoleLines.value.push(`Tracing route to ${target} (max 15 hops)...`);
    await scrollToBottom();

    try {
      const res = await diagnosticsApi.runTraceroute(target);
      if (res && res.output) {
        for (const line of res.output) {
          consoleLines.value.push(line);
        }
      } else if (res && res.raw) {
        consoleLines.value.push(res.raw);
      }
      consoleLines.value.push(`--- Trace complete (${res?.durationMs ?? 0}ms) ---\n`);
    } catch (e: any) {
      consoleLines.value.push(`Error executing traceroute: ${e.response?.data?.error || e.message}`);
    }
  } else if (activeMode.value === 'port') {
    const port = portInput.value || 80;
    consoleLines.value.push(`$ [${timestamp}] nc -zv ${target} ${port}`);
    await scrollToBottom();

    try {
      const res = await diagnosticsApi.runPortProbe(target, port);
      if (res && res.message) {
        consoleLines.value.push(res.message);
      }
      consoleLines.value.push(`--- Port check finished in ${res?.latencyMs ?? 0}ms ---\n`);
    } catch (e: any) {
      consoleLines.value.push(`Error checking port ${port}: ${e.response?.data?.error || e.message}`);
    }
  }

  isRunning.value = false;
  await scrollToBottom();
}
</script>
