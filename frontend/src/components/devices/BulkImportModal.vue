<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div
        v-if="isOpen"
        class="fixed inset-0 bg-black/75 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        @click.self="handleClose"
      >
        <div class="bg-[#151517] border border-[#26262A] rounded-2xl w-full max-w-3xl shadow-2xl shadow-black/50 flex flex-col max-h-[90vh] overflow-hidden">

          <!-- Modal Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-[#26262A] shrink-0">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-lg bg-[#7B96F5]/15 border border-[#7B96F5]/30 flex items-center justify-center">
                <FileSpreadsheet class="w-4 h-4 text-[#7B96F5]" />
              </div>
              <div>
                <h2 class="text-sm font-bold text-white">Bulk Import Devices</h2>
                <p class="text-[10px] text-gray-400 font-mono mt-0.5">CSV / Excel spreadsheet import wizard</p>
              </div>
            </div>
            <button @click="handleClose" class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-[#26262A] transition-colors">
              <X class="w-4 h-4" />
            </button>
          </div>

          <!-- Step Indicator -->
          <div class="flex items-center gap-0 px-6 pt-4 shrink-0">
            <div
              v-for="(s, i) in steps"
              :key="i"
              class="flex items-center"
            >
              <div class="flex items-center gap-2">
                <div
                  class="w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-bold font-mono transition-all"
                  :class="step > i
                    ? 'bg-[#34D399] text-black'
                    : step === i
                      ? 'bg-[#7B96F5] text-white'
                      : 'bg-[#26262A] text-gray-500'"
                >
                  <Check v-if="step > i" class="w-3 h-3" />
                  <span v-else>{{ i + 1 }}</span>
                </div>
                <span
                  class="text-[10px] font-mono uppercase tracking-wider transition-colors"
                  :class="step === i ? 'text-gray-200' : 'text-gray-500'"
                >{{ s }}</span>
              </div>
              <div v-if="i < steps.length - 1" class="mx-3 flex-1 h-px bg-[#26262A] w-8" />
            </div>
          </div>

          <!-- Step Content (scrollable) -->
          <div class="flex-1 overflow-y-auto px-6 py-4">

            <!-- ── Step 0: File Upload ──────────────────────────────────── -->
            <div v-if="step === 0" class="space-y-4">
              <div
                class="border-2 border-dashed rounded-xl p-10 flex flex-col items-center gap-3 transition-colors cursor-pointer"
                :class="isDragOver
                  ? 'border-[#7B96F5] bg-[#7B96F5]/5'
                  : 'border-[#26262A] hover:border-[#7B96F5]/50 hover:bg-[#18181B]'"
                @dragover.prevent="isDragOver = true"
                @dragleave="isDragOver = false"
                @drop.prevent="handleDrop"
                @click="fileInputRef?.click()"
              >
                <div class="w-14 h-14 rounded-xl bg-[#7B96F5]/10 border border-[#7B96F5]/20 flex items-center justify-center">
                  <Upload class="w-6 h-6 text-[#7B96F5]" />
                </div>
                <div class="text-center">
                  <p class="text-sm font-semibold text-gray-200">Drop your file here or click to browse</p>
                  <p class="text-xs text-gray-500 mt-1">Supports <span class="font-mono text-gray-400">.csv</span> and <span class="font-mono text-gray-400">.xlsx</span> — max 10MB</p>
                </div>
                <input ref="fileInputRef" type="file" accept=".csv,.xlsx" class="hidden" @change="handleFileSelect" />
              </div>

              <!-- Format hint -->
              <div class="bg-[#0A0A0B] border border-[#26262A] rounded-xl p-4 space-y-2">
                <p class="text-[10px] font-mono uppercase tracking-widest text-gray-500">Expected CSV format</p>
                <pre class="text-[11px] font-mono text-gray-300 leading-5">Name, Model (UniFi), Status, MAC Address, IP Address
02.01, U6 IW, Up to date, e4:38:83:46:f0:25, 10.11.11.122
2.1 SETDA, U6 Enterprise, Up to date, 9c:05:d6:a5:4b:65, 10.11.9.100</pre>
              </div>

              <!-- Selected file info -->
              <div v-if="selectedFile" class="flex items-center gap-3 bg-[#18181B] border border-[#34D399]/30 rounded-xl p-3">
                <FileCheck class="w-4 h-4 text-[#34D399] shrink-0" />
                <div class="flex-1 min-w-0">
                  <p class="text-xs font-semibold text-white truncate">{{ selectedFile.name }}</p>
                  <p class="text-[10px] text-gray-500 font-mono">{{ (selectedFile.size / 1024).toFixed(1) }} KB</p>
                </div>
                <button @click.stop="clearFile" class="text-gray-400 hover:text-gray-200">
                  <X class="w-4 h-4" />
                </button>
              </div>
            </div>

            <!-- ── Step 1: Column Mapping ────────────────────────────────── -->
            <div v-else-if="step === 1" class="space-y-4">
              <p class="text-xs text-gray-400">
                Map each column from your file to the corresponding GovMonitor field.
                Unrecognized columns can be skipped.
              </p>
              <div class="space-y-2">
                <div
                  v-for="(mapping, i) in columnMappings"
                  :key="i"
                  class="flex items-center gap-4 bg-[#18181B] border border-[#26262A] rounded-xl px-4 py-3"
                >
                  <div class="flex-1 min-w-0">
                    <p class="text-[10px] font-mono text-gray-500 uppercase tracking-wider mb-0.5">Source Column</p>
                    <p class="text-xs font-semibold text-gray-200 truncate">{{ mapping.sourceColumn }}</p>
                  </div>
                  <ArrowRight class="w-4 h-4 text-gray-600 shrink-0" />
                  <div class="flex-1">
                    <p class="text-[10px] font-mono text-gray-500 uppercase tracking-wider mb-0.5">Map To</p>
                    <select
                      v-model="mapping.targetField"
                      class="w-full bg-[#0A0A0B] border border-[#26262A] rounded-lg px-2 py-1.5 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
                    >
                      <option value="">— Skip this column —</option>
                      <option value="name">name</option>
                      <option value="model">model</option>
                      <option value="mac">mac_address</option>
                      <option value="ip">ip_address</option>
                      <option value="firmwareStatus">firmware_status</option>
                      <option value="type">device_type</option>
                      <option value="location">location</option>
                      <option value="addressingMode">addressing_mode</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>

            <!-- ── Step 2: Preview & Validate ──────────────────────────── -->
            <div v-else-if="step === 2" class="space-y-3">
              <div class="flex items-center justify-between">
                <p class="text-xs text-gray-400">
                  <span class="text-white font-semibold">{{ parsedRows.length }}</span> rows parsed.
                  Review validation results before importing.
                </p>
                <div class="flex items-center gap-2 text-[10px] font-mono">
                  <span class="px-2 py-0.5 rounded-full bg-[#34D399]/15 text-[#34D399] border border-[#34D399]/30">
                    {{ validCount }} valid
                  </span>
                  <span class="px-2 py-0.5 rounded-full bg-amber-500/15 text-amber-400 border border-amber-500/30">
                    {{ dupCount }} duplicate
                  </span>
                  <span class="px-2 py-0.5 rounded-full bg-[#F16565]/15 text-[#F16565] border border-[#F16565]/30">
                    {{ errorCount }} error
                  </span>
                </div>
              </div>

              <div class="bg-[#18181B] border border-[#26262A] rounded-xl overflow-hidden">
                <table class="w-full text-xs text-gray-300">
                  <thead class="bg-[#0A0A0B] border-b border-[#26262A] font-mono text-[10px] uppercase text-gray-500">
                    <tr>
                      <th class="py-2.5 px-3">#</th>
                      <th class="py-2.5 px-3">Name</th>
                      <th class="py-2.5 px-3">MAC Address</th>
                      <th class="py-2.5 px-3">IP Address</th>
                      <th class="py-2.5 px-3">Model</th>
                      <th class="py-2.5 px-3">Status</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-[#26262A]">
                    <tr v-for="row in parsedRows" :key="row.rowIndex" class="hover:bg-[#1A1A1E]">
                      <td class="py-2 px-3 font-mono text-gray-600">{{ row.rowIndex + 1 }}</td>
                      <td class="py-2 px-3 font-semibold text-white">{{ row.mapped.name || '—' }}</td>
                      <td class="py-2 px-3 font-mono text-gray-400 text-[11px]">{{ row.mapped.mac || '—' }}</td>
                      <td class="py-2 px-3 font-mono text-gray-400">{{ row.mapped.ip || '—' }}</td>
                      <td class="py-2 px-3 text-gray-400">{{ row.mapped.model || '—' }}</td>
                      <td class="py-2 px-3">
                        <span
                          class="px-2 py-0.5 rounded-full text-[10px] font-mono font-semibold border"
                          :class="rowStatusClass(row.status)"
                        >
                          {{ rowStatusLabel(row.status) }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- ── Step 3: Result ──────────────────────────────────────── -->
            <div v-else-if="step === 3" class="space-y-5">
              <div class="text-center py-4">
                <div
                  class="w-16 h-16 rounded-full mx-auto flex items-center justify-center border-2 mb-4"
                  :class="importSummary.failed === 0
                    ? 'bg-[#34D399]/15 border-[#34D399]/40'
                    : 'bg-amber-500/15 border-amber-500/40'"
                >
                  <CheckCircle v-if="importSummary.failed === 0" class="w-7 h-7 text-[#34D399]" />
                  <AlertCircle v-else class="w-7 h-7 text-amber-400" />
                </div>
                <p class="text-base font-bold text-white">Import Complete</p>
                <p class="text-xs text-gray-400 mt-1">Results for your CSV import</p>
              </div>

              <div class="grid grid-cols-3 gap-3">
                <div class="bg-[#0A0A0B] border border-[#34D399]/20 rounded-xl p-4 text-center">
                  <p class="text-2xl font-extrabold text-[#34D399] font-mono">{{ importSummary.imported }}</p>
                  <p class="text-[10px] text-gray-400 uppercase font-mono mt-1">Imported</p>
                </div>
                <div class="bg-[#0A0A0B] border border-amber-500/20 rounded-xl p-4 text-center">
                  <p class="text-2xl font-extrabold text-amber-400 font-mono">{{ importSummary.skipped }}</p>
                  <p class="text-[10px] text-gray-400 uppercase font-mono mt-1">Skipped</p>
                </div>
                <div class="bg-[#0A0A0B] border border-[#F16565]/20 rounded-xl p-4 text-center">
                  <p class="text-2xl font-extrabold text-[#F16565] font-mono">{{ importSummary.failed }}</p>
                  <p class="text-[10px] text-gray-400 uppercase font-mono mt-1">Failed</p>
                </div>
              </div>

              <div v-if="importSummary.failed > 0" class="bg-[#18181B] border border-[#26262A] rounded-xl p-4 space-y-2">
                <p class="text-[10px] font-mono uppercase text-gray-500">Failed rows</p>
                <div v-for="r in importSummary.results.filter(x => x.status === 'failed')" :key="r.rowIndex" class="flex items-center gap-3">
                  <span class="text-[10px] font-mono text-gray-600">Row {{ r.rowIndex + 1 }}</span>
                  <span class="text-xs text-[#F16565]">{{ r.reason }}</span>
                </div>
                <button @click="downloadErrorLog" class="mt-2 text-xs text-[#7B96F5] hover:underline flex items-center gap-1">
                  <Download class="w-3.5 h-3.5" />
                  Download error log
                </button>
              </div>
            </div>
          </div>

          <!-- Footer Actions -->
          <div class="flex items-center justify-between px-6 py-4 border-t border-[#26262A] shrink-0 bg-[#0D0D0F]">
            <button
              v-if="step > 0 && step < 3"
              @click="step--"
              class="px-4 py-2 rounded-lg border border-[#26262A] text-gray-400 hover:text-gray-200 hover:bg-[#26262A] transition-colors text-xs font-medium flex items-center gap-1.5"
            >
              <ChevronLeft class="w-4 h-4" />
              Back
            </button>
            <div v-else />

            <!-- Step 0: Next requires file -->
            <button
              v-if="step === 0"
              :disabled="!selectedFile"
              @click="proceedFromUpload"
              class="px-5 py-2 rounded-lg text-white font-semibold text-xs transition-all flex items-center gap-1.5 disabled:opacity-40 disabled:cursor-not-allowed"
              :class="selectedFile ? 'bg-[#7B96F5] hover:bg-[#95ABF7] shadow-md shadow-[#7B96F5]/20' : 'bg-[#26262A]'"
            >
              Next: Map Columns
              <ChevronRight class="w-4 h-4" />
            </button>

            <!-- Step 1: Next confirms mapping -->
            <button
              v-else-if="step === 1"
              @click="proceedFromMapping"
              class="px-5 py-2 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 transition-all flex items-center gap-1.5"
            >
              Next: Preview
              <ChevronRight class="w-4 h-4" />
            </button>

            <!-- Step 2: Confirm import -->
            <button
              v-else-if="step === 2"
              :disabled="validCount === 0"
              @click="executeImport"
              class="px-5 py-2 rounded-lg text-white font-semibold text-xs transition-all flex items-center gap-1.5 disabled:opacity-40"
              :class="validCount > 0 ? 'bg-[#34D399] hover:bg-emerald-400 shadow-md shadow-[#34D399]/20' : 'bg-[#26262A]'"
            >
              <Upload class="w-4 h-4" />
              Import {{ validCount }} Device{{ validCount !== 1 ? 's' : '' }}
            </button>

            <!-- Step 3: Done -->
            <button
              v-else-if="step === 3"
              @click="handleClose"
              class="px-5 py-2 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 transition-all"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useDeviceStore } from '../../stores/deviceStore';
import type { ImportRow, ImportSummary, ColumnMapping, Device } from '../../types';
import {
  FileSpreadsheet, X, Upload, Check, ArrowRight, FileCheck,
  ChevronLeft, ChevronRight, CheckCircle, AlertCircle, Download
} from 'lucide-vue-next';

const props = defineProps<{ isOpen: boolean }>();
const emit = defineEmits(['close']);

const deviceStore = useDeviceStore();

// ─── State ───────────────────────────────────────────────────────────────────
const step = ref(0);
const steps = ['Upload File', 'Map Columns', 'Preview', 'Result'];
const isDragOver = ref(false);
const selectedFile = ref<File | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);
const columnMappings = ref<ColumnMapping[]>([]);
const parsedRows = ref<ImportRow[]>([]);
const importSummary = ref<ImportSummary>({ imported: 0, skipped: 0, failed: 0, results: [] });

// ─── Computed ────────────────────────────────────────────────────────────────
const validCount = computed(() => parsedRows.value.filter(r => r.status === 'valid').length);
const dupCount = computed(() => parsedRows.value.filter(r => r.status === 'duplicate_mac').length);
const errorCount = computed(() => parsedRows.value.filter(r => r.status === 'missing_field' || r.status === 'invalid_mac' || r.status === 'failed').length);

// ─── File Handling ───────────────────────────────────────────────────────────
function handleDrop(e: DragEvent) {
  isDragOver.value = false;
  const file = e.dataTransfer?.files[0];
  if (file) setFile(file);
}

function handleFileSelect(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0];
  if (file) setFile(file);
}

function setFile(file: File) {
  if (!file.name.match(/\.(csv|xlsx)$/i)) {
    alert('Only .csv and .xlsx files are supported.');
    return;
  }
  selectedFile.value = file;
}

function clearFile() {
  selectedFile.value = null;
  if (fileInputRef.value) fileInputRef.value.value = '';
}

// ─── CSV Parsing ─────────────────────────────────────────────────────────────
async function parseCSV(file: File): Promise<{ headers: string[]; rows: Record<string, string>[] }> {
  const text = await file.text();
  const lines = text.trim().split('\n').filter(l => l.trim());
  const headers = lines[0].split(',').map(h => h.trim().replace(/^"|"$/g, ''));
  const rows = lines.slice(1).map(line => {
    const values = line.split(',').map(v => v.trim().replace(/^"|"$/g, ''));
    return Object.fromEntries(headers.map((h, i) => [h, values[i] ?? '']));
  });
  return { headers, rows };
}

// ─── Default Column Auto-mapping ─────────────────────────────────────────────
const autoMappingRules: Record<string, keyof Device | 'model' | 'firmwareStatus'> = {
  'name': 'name',
  'device name': 'name',
  'model (unifi)': 'model',
  'model': 'model',
  'mac address': 'mac',
  'mac': 'mac',
  'ip address': 'ip',
  'ip': 'ip',
  'status': 'firmwareStatus',
  'type': 'type',
  'location': 'location',
};

function autoMapColumn(header: string): keyof Device | 'model' | 'firmwareStatus' | '' {
  const normalized = header.toLowerCase().trim();
  return autoMappingRules[normalized] ?? '';
}

// ─── Validation ───────────────────────────────────────────────────────────────
function normalizeMac(mac: string): string {
  return mac.toLowerCase().replace(/[:-]/g, ':').replace(/(.{2})(?!$)/g, '$1:').replace(/:{2,}/g, ':');
}

function isValidMac(mac: string): boolean {
  return /^([0-9a-f]{2}:){5}[0-9a-f]{2}$/i.test(mac);
}

function validateRows(rawRows: Record<string, string>[]): ImportRow[] {
  const existingMacs = new Set(deviceStore.devices.map((d: Device) => d.mac.toLowerCase()));
  const seenMacs = new Set<string>();

  return rawRows.map((raw, idx) => {
    const mapped: Partial<Device> = {};

    // Apply column mappings
    for (const m of columnMappings.value) {
      if (!m.targetField || !raw[m.sourceColumn]) continue;
      const val = raw[m.sourceColumn].trim();
      if (m.targetField === 'mac') {
        mapped.mac = normalizeMac(val);
      } else if (m.targetField === 'model' || m.targetField === 'firmwareStatus') {
        (mapped as any)[m.targetField] = val;
      } else {
        (mapped as any)[m.targetField] = val;
      }
    }

    // Validation checks
    if (!mapped.name) {
      return { rowIndex: idx, raw, mapped, status: 'missing_field' as const, reason: 'Missing device name' };
    }
    if (!mapped.mac) {
      return { rowIndex: idx, raw, mapped, status: 'missing_field' as const, reason: 'Missing MAC address' };
    }
    if (!isValidMac(mapped.mac)) {
      return { rowIndex: idx, raw, mapped, status: 'invalid_mac' as const, reason: `Invalid MAC format: ${mapped.mac}` };
    }
    if (existingMacs.has(mapped.mac.toLowerCase()) || seenMacs.has(mapped.mac.toLowerCase())) {
      return { rowIndex: idx, raw, mapped, status: 'duplicate_mac' as const, reason: `Duplicate MAC: ${mapped.mac}` };
    }

    seenMacs.add(mapped.mac.toLowerCase());
    return { rowIndex: idx, raw, mapped, status: 'valid' as const };
  });
}

// ─── Step Transitions ────────────────────────────────────────────────────────
async function proceedFromUpload() {
  if (!selectedFile.value) return;
  const { headers } = await parseCSV(selectedFile.value);
  columnMappings.value = headers.map(h => ({
    sourceColumn: h,
    targetField: autoMapColumn(h) as any
  }));
  step.value = 1;
}

async function proceedFromMapping() {
  if (!selectedFile.value) return;
  const { rows } = await parseCSV(selectedFile.value);
  parsedRows.value = validateRows(rows);
  step.value = 2;
}

async function executeImport() {
  const results: ImportSummary['results'] = [];
  let imported = 0;
  let skipped = 0;
  let failed = 0;

  for (const row of parsedRows.value) {
    if (row.status === 'valid') {
      try {
        await deviceStore.addDevice({
          name: row.mapped.name!,
          type: (row.mapped.type as any) || 'Access Point',
          ip: row.mapped.ip || '0.0.0.0',
          mac: row.mapped.mac!,
          addressingMode: 'DHCP',
          location: row.mapped.location || '',
        });
        results.push({ rowIndex: row.rowIndex, status: 'imported' });
        imported++;
      } catch {
        results.push({ rowIndex: row.rowIndex, status: 'failed', reason: 'API error' });
        failed++;
      }
    } else if (row.status === 'duplicate_mac') {
      results.push({ rowIndex: row.rowIndex, status: 'skipped', reason: row.reason });
      skipped++;
    } else {
      results.push({ rowIndex: row.rowIndex, status: 'failed', reason: row.reason });
      failed++;
    }
  }

  importSummary.value = { imported, skipped, failed, results };
  step.value = 3;
}

// ─── Helpers ─────────────────────────────────────────────────────────────────
function rowStatusClass(status: string) {
  switch (status) {
    case 'valid': return 'bg-[#34D399]/15 text-[#34D399] border-[#34D399]/30';
    case 'duplicate_mac': return 'bg-amber-500/15 text-amber-400 border-amber-500/30';
    default: return 'bg-[#F16565]/15 text-[#F16565] border-[#F16565]/30';
  }
}

function rowStatusLabel(status: string) {
  switch (status) {
    case 'valid': return '✓ Valid';
    case 'duplicate_mac': return 'Duplicate MAC';
    case 'missing_field': return 'Missing Field';
    case 'invalid_mac': return 'Bad MAC';
    default: return 'Error';
  }
}

function downloadErrorLog() {
  const failed = importSummary.value.results.filter(r => r.status === 'failed' || r.status === 'skipped');
  const csv = ['Row,Status,Reason', ...failed.map(r => `${r.rowIndex + 1},${r.status},"${r.reason}"`)].join('\n');
  const blob = new Blob([csv], { type: 'text/csv' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'import-errors.csv';
  a.click();
  URL.revokeObjectURL(url);
}

function handleClose() {
  emit('close');
  // Reset after close animation
  setTimeout(() => {
    step.value = 0;
    selectedFile.value = null;
    parsedRows.value = [];
    columnMappings.value = [];
    importSummary.value = { imported: 0, skipped: 0, failed: 0, results: [] };
  }, 300);
}

// Reset on open
watch(() => props.isOpen, (val) => {
  if (val) {
    step.value = 0;
    selectedFile.value = null;
    parsedRows.value = [];
  }
});
</script>

<style scoped>
.modal-fade-enter-active, .modal-fade-leave-active {
  transition: opacity 0.2s ease;
}
.modal-fade-enter-from, .modal-fade-leave-to {
  opacity: 0;
}
.modal-fade-enter-active .bg-\[\#151517\], .modal-fade-leave-active .bg-\[\#151517\] {
  transition: transform 0.2s ease;
}
.modal-fade-enter-from .bg-\[\#151517\], .modal-fade-leave-to .bg-\[\#151517\] {
  transform: scale(0.97) translateY(8px);
}
</style>
