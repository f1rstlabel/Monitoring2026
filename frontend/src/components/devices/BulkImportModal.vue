<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div
        v-if="isOpen"
        class="fixed inset-0 bg-black/75 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        @click.self="handleClose"
      >
        <div class="bg-surface border border-subtle rounded-2xl w-full max-w-3xl shadow-2xl shadow-black/50 flex flex-col max-h-[90vh] overflow-hidden">

          <!-- Modal Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-subtle shrink-0 bg-card">
            <h3 class="text-base font-bold text-text-main flex items-center gap-2">
              <FileSpreadsheet class="w-5 h-5 text-brand-periwinkle" />
              Bulk Import Devices
            </h3>
            <button @click="handleClose" class="p-1 rounded-lg text-text-secondary hover:text-text-main hover:bg-subtle transition-colors cursor-pointer">
              <X class="w-5 h-5" />
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
                    ? 'bg-status-up text-black'
                    : step === i
                      ? 'bg-brand-periwinkle text-white'
                      : 'bg-subtle text-text-muted'"
                >
                  <Check v-if="step > i" class="w-3 h-3" />
                  <span v-else>{{ i + 1 }}</span>
                </div>
                <span
                  class="text-[10px] font-mono uppercase tracking-wider transition-colors"
                  :class="step === i ? 'text-text-main' : 'text-text-muted'"
                >{{ s }}</span>
              </div>
              <div v-if="i < steps.length - 1" class="mx-3 flex-1 h-px bg-subtle w-8" />
            </div>
          </div>

          <!-- Step Content (scrollable) -->
          <div class="flex-1 overflow-y-auto px-6 py-4">

            <!-- ── Step 0: File Upload ──────────────────────────────────── -->
            <div v-if="step === 0" class="space-y-4">
              <div
                class="border-2 border-dashed rounded-xl p-10 flex flex-col items-center gap-3 transition-colors cursor-pointer"
                :class="isDragOver
                  ? 'border-brand-periwinkle bg-brand-periwinkle/5'
                  : 'border-subtle hover:border-brand-periwinkle/50 hover:bg-card'"
                @dragover.prevent="isDragOver = true"
                @dragleave="isDragOver = false"
                @drop.prevent="handleDrop"
                @click="fileInputRef?.click()"
              >
                <div class="w-14 h-14 rounded-xl bg-brand-periwinkle/10 border border-brand-periwinkle/20 flex items-center justify-center">
                  <Upload class="w-6 h-6 text-brand-periwinkle" />
                </div>
                <div class="text-center">
                  <p class="text-sm font-semibold text-text-main">Drop your file here or click to browse</p>
                  <p class="text-xs text-text-muted mt-1">Supports <span class="font-mono text-text-secondary">.csv</span>, <span class="font-mono text-text-secondary">.xlsx</span>, and <span class="font-mono text-text-secondary">.xls</span> — max 10MB</p>
                </div>
                <input ref="fileInputRef" type="file" accept=".csv,.xlsx,.xls,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" class="hidden" @change="handleFileSelect" />
              </div>

              <!-- Format hint & Template Downloads -->
              <div class="bg-main border border-subtle rounded-xl p-4 space-y-3">
                <div class="flex items-center justify-between flex-wrap gap-2">
                  <p class="text-[10px] font-mono uppercase tracking-widest text-text-muted">Template Format (.CSV / .XLSX / .XLS)</p>
                  <div class="flex items-center gap-2">
                    <button type="button" @click="downloadExcelTemplate" class="px-2.5 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 hover:bg-emerald-500/20 text-xs flex items-center gap-1.5 font-semibold transition-colors cursor-pointer">
                      <FileSpreadsheet class="w-3.5 h-3.5" />
                      Download Excel (.xlsx)
                    </button>
                    <button type="button" @click="downloadCSVTemplate" class="px-2.5 py-1.5 rounded-lg bg-brand-periwinkle/10 border border-brand-periwinkle/30 text-brand-periwinkle hover:bg-brand-periwinkle/20 text-xs flex items-center gap-1.5 font-semibold transition-colors cursor-pointer">
                      <Download class="w-3.5 h-3.5" />
                      Download CSV (.csv)
                    </button>
                  </div>
                </div>
                <pre class="text-[10px] font-mono text-text-secondary leading-5 overflow-x-auto whitespace-pre">Device Name,Device Type,Addressing Mode,IP Address,MAC Address,Location,Rack,SNMP Polling Enabled,SNMP Community,SNMP Port,Custom Failure Threshold Override,Custom Failure Threshold Value
AP Biro Umum,Access Point,Static IP,10.20.1.18,00:1A:2B:3C:4D:5E,Gedung Sate,Rack A,FALSE,public,161,FALSE,3
AP Core Switch,Switch,DHCP Reservation,,00:1A:2B:3C:4D:5F,Gedung Sate,Rack B,TRUE,public,161,TRUE,5</pre>
              </div>

              <!-- Selected file info -->
              <div v-if="selectedFile" class="flex items-center gap-3 bg-card border border-status-up/30 rounded-xl p-3">
                <FileCheck class="w-4 h-4 text-status-up shrink-0" />
                <div class="flex-1 min-w-0">
                  <p class="text-xs font-semibold text-text-main truncate">{{ selectedFile.name }}</p>
                  <p class="text-[10px] text-text-muted font-mono">{{ (selectedFile.size / 1024).toFixed(1) }} KB</p>
                </div>
                <button @click.stop="clearFile" class="text-text-secondary hover:text-text-main">
                  <X class="w-4 h-4" />
                </button>
              </div>
            </div>

            <!-- ── Step 1: Column Mapping ────────────────────────────────── -->
            <div v-else-if="step === 1" class="space-y-4">
              <p class="text-xs text-text-secondary">
                Map each column from your file to the corresponding SANOC field.
                Unrecognized columns can be skipped.
              </p>
              <div class="space-y-2">
                <div
                  v-for="(mapping, i) in columnMappings"
                  :key="i"
                  class="flex items-center gap-4 bg-card border border-subtle rounded-xl px-4 py-3"
                >
                  <div class="flex-1 min-w-0">
                    <p class="text-[10px] font-mono text-text-muted uppercase tracking-wider mb-0.5">Source Column</p>
                    <p class="text-xs font-semibold text-text-main truncate">{{ mapping.sourceColumn }}</p>
                  </div>
                  <ArrowRight class="w-4 h-4 text-text-muted shrink-0" />
                  <div class="flex-1">
                    <p class="text-[10px] font-mono text-text-muted uppercase tracking-wider mb-0.5">Map To</p>
                    <select
                      v-model="mapping.targetField"
                      class="w-full bg-main border border-subtle rounded-lg px-2 py-1.5 text-xs text-text-main focus:outline-none focus:border-brand-periwinkle font-mono"
                    >
                      <option value="">— Skip this column —</option>
                      <option value="name">Device Name</option>
                      <option value="type">Device Type</option>
                      <option value="addressingMode">Addressing Mode</option>
                      <option value="ip">IP Address</option>
                      <option value="mac">MAC Address</option>
                      <option value="location">Location</option>
                      <option value="rack">Rack</option>
                      <option value="snmpEnabled">SNMP Polling Enabled</option>
                      <option value="snmpCommunity">SNMP Community</option>
                      <option value="snmpPort">SNMP Port</option>
                      <option value="useCustomThreshold">Custom Failure Threshold Override</option>
                      <option value="customFailureThreshold">Custom Failure Threshold Value</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>

            <!-- ── Step 2: Preview & Validate ──────────────────────────── -->
            <div v-else-if="step === 2" class="space-y-3">
              <div class="flex items-center justify-between">
                <p class="text-xs text-text-secondary">
                  <span class="text-text-main font-semibold">{{ parsedRows.length }}</span> rows parsed.
                  Review validation results before importing.
                </p>
                <div class="flex items-center gap-2 text-[10px] font-mono">
                  <span class="px-2 py-0.5 rounded-full bg-status-up/15 text-status-up border border-status-up/30">
                    {{ validCount }} valid
                  </span>
                  <span class="px-2 py-0.5 rounded-full bg-amber-500/15 text-amber-400 border border-amber-500/30">
                    {{ dupCount }} duplicate
                  </span>
                  <span class="px-2 py-0.5 rounded-full bg-status-down/15 text-status-down border border-status-down/30">
                    {{ errorCount }} error
                  </span>
                </div>
              </div>

              <div class="bg-card border border-subtle rounded-xl overflow-hidden">
                <table class="w-full text-xs text-text-secondary">
                  <thead class="bg-main border-b border-subtle font-mono text-[10px] uppercase text-text-muted">
                    <tr>
                      <th class="py-2.5 px-3">#</th>
                      <th class="py-2.5 px-3">Name</th>
                      <th class="py-2.5 px-3">MAC Address</th>
                      <th class="py-2.5 px-3">IP Address</th>
                      <th class="py-2.5 px-3">Model</th>
                      <th class="py-2.5 px-3">Status</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-subtle">
                    <tr v-for="row in parsedRows" :key="row.rowIndex" class="hover:bg-card">
                      <td class="py-2 px-3 font-mono text-text-muted">{{ row.rowIndex + 1 }}</td>
                      <td class="py-2 px-3 font-semibold text-text-main">{{ row.mapped.name || '—' }}</td>
                      <td class="py-2 px-3 font-mono text-text-secondary text-[11px]">{{ row.mapped.mac || '—' }}</td>
                      <td class="py-2 px-3 font-mono text-text-secondary">{{ row.mapped.ip || '—' }}</td>
                      <td class="py-2 px-3 text-text-secondary">{{ row.mapped.model || '—' }}</td>
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
                    ? 'bg-status-up/15 border-status-up/40'
                    : 'bg-amber-500/15 border-amber-500/40'"
                >
                  <CheckCircle v-if="importSummary.failed === 0" class="w-7 h-7 text-status-up" />
                  <AlertCircle v-else class="w-7 h-7 text-amber-400" />
                </div>
                <p class="text-base font-bold text-text-main">Import Complete</p>
                <p class="text-xs text-text-secondary mt-1">Results for your CSV import</p>
              </div>

              <div class="grid grid-cols-3 gap-3">
                <div class="bg-main border border-status-up/20 rounded-xl p-4 text-center">
                  <p class="text-2xl font-extrabold text-status-up font-mono">{{ importSummary.imported }}</p>
                  <p class="text-[10px] text-text-secondary uppercase font-mono mt-1">Imported</p>
                </div>
                <div class="bg-main border border-amber-500/20 rounded-xl p-4 text-center">
                  <p class="text-2xl font-extrabold text-amber-400 font-mono">{{ importSummary.skipped }}</p>
                  <p class="text-[10px] text-text-secondary uppercase font-mono mt-1">Skipped</p>
                </div>
                <div class="bg-main border border-status-down/20 rounded-xl p-4 text-center">
                  <p class="text-2xl font-extrabold text-status-down font-mono">{{ importSummary.failed }}</p>
                  <p class="text-[10px] text-text-secondary uppercase font-mono mt-1">Failed</p>
                </div>
              </div>

              <!-- Location Processing Summary -->
              <div v-if="newlyCreatedLocations.length > 0 || matchedLocations.length > 0" class="bg-card border border-subtle rounded-xl p-4 space-y-3">
                <p class="text-[10px] font-mono uppercase text-text-muted">Location Processing Summary</p>
                <div v-if="newlyCreatedLocations.length > 0" class="space-y-1">
                  <p class="text-xs font-semibold text-emerald-400">Newly Created Locations ({{ newlyCreatedLocations.length }}):</p>
                  <p class="text-[11px] text-text-secondary pl-2 leading-relaxed">{{ newlyCreatedLocations.join(', ') }}</p>
                </div>
                <div v-if="matchedLocations.length > 0" class="space-y-1">
                  <p class="text-xs font-semibold text-sky-400">Matched Existing Locations ({{ matchedLocations.length }}):</p>
                  <p class="text-[11px] text-text-secondary pl-2 leading-relaxed">{{ matchedLocations.join(', ') }}</p>
                </div>
              </div>

              <div v-if="importSummary.failed > 0" class="bg-card border border-subtle rounded-xl p-4 space-y-2">
                <p class="text-[10px] font-mono uppercase text-text-muted">Failed rows</p>
                <div v-for="r in importSummary.results.filter((x: any) => x.status === 'failed')" :key="r.rowIndex" class="flex items-center gap-3">
                  <span class="text-[10px] font-mono text-text-muted">Row {{ r.rowIndex + 1 }}</span>
                  <span class="text-xs text-status-down">{{ r.reason }}</span>
                </div>
                <button @click="downloadErrorLog" class="mt-2 text-xs text-brand-periwinkle hover:underline flex items-center gap-1 bg-transparent border-0 cursor-pointer">
                  <Download class="w-3.5 h-3.5" />
                  Download error log
                </button>
              </div>
            </div>
          </div>

          <!-- Footer Actions -->
          <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-subtle shrink-0 bg-card">
            <button
              v-if="step > 0 && step < 3"
              @click="step--"
              class="px-4 py-2 rounded-lg border border-subtle text-text-secondary hover:text-text-main hover:bg-subtle transition-colors text-xs font-medium flex items-center gap-2 cursor-pointer"
            >
              <ChevronLeft class="w-4 h-4" />
              Back
            </button>

            <!-- Step 0: Next requires file -->
            <button
              v-if="step === 0"
              :disabled="!selectedFile"
              @click="proceedFromUpload"
              class="px-5 py-2 rounded-lg bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-semibold text-xs shadow-md shadow-brand-periwinkle/20 transition-all flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            >
              Next: Map Columns
              <ChevronRight class="w-4 h-4" />
            </button>

            <!-- Step 1: Next confirms mapping -->
            <button
              v-else-if="step === 1"
              @click="proceedFromMapping"
              class="px-5 py-2 rounded-lg bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-semibold text-xs shadow-md shadow-brand-periwinkle/20 transition-all flex items-center gap-2 cursor-pointer"
            >
              Next: Preview
              <ChevronRight class="w-4 h-4" />
            </button>

            <!-- Step 2: Confirm import -->
            <button
              v-else-if="step === 2"
              :disabled="validCount === 0"
              @click="executeImport"
              class="px-5 py-2 rounded-lg bg-status-up hover:bg-emerald-400 text-text-main font-semibold text-xs shadow-md shadow-status-up/20 transition-all flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            >
              <Upload class="w-4 h-4" />
              Import {{ validCount }} Device{{ validCount !== 1 ? 's' : '' }}
            </button>

            <!-- Step 3: Done -->
            <button
              v-else-if="step === 3"
              @click="handleClose"
              class="px-5 py-2 rounded-lg bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-semibold text-xs shadow-md shadow-brand-periwinkle/20 transition-all cursor-pointer"
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
import * as XLSX from 'xlsx';
import { useDeviceStore } from '../../stores/deviceStore';
import { useToastStore } from '../../stores/toastStore';
import { locationsApi } from '../../api';
import type { ImportRow, ImportResult, ImportSummary, ColumnMapping, Device } from '../../types';
import {
  FileSpreadsheet, X, Upload, Check, ArrowRight, FileCheck,
  ChevronLeft, ChevronRight, CheckCircle, AlertCircle, Download
} from 'lucide-vue-next';

const props = defineProps<{ isOpen: boolean }>();
const emit = defineEmits(['close']);

const deviceStore = useDeviceStore();
const toastStore = useToastStore();

// ─── State ───────────────────────────────────────────────────────────────────
const step = ref(0);
const steps = ['Upload File', 'Map Columns', 'Preview', 'Result'];
const isDragOver = ref(false);
const selectedFile = ref<File | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);
const columnMappings = ref<ColumnMapping[]>([]);
const parsedRows = ref<ImportRow[]>([]);
const importSummary = ref<ImportSummary>({ imported: 0, skipped: 0, failed: 0, results: [] });
const newlyCreatedLocations = ref<string[]>([]);
const matchedLocations = ref<string[]>([]);

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
  if (!file.name.match(/\.(csv|xlsx|xls)$/i)) {
    toastStore.error('Format File Tidak Didukung', 'Hanya file format .csv, .xlsx, dan .xls yang didukung.');
    return;
  }
  selectedFile.value = file;
}

function clearFile() {
  selectedFile.value = null;
  if (fileInputRef.value) fileInputRef.value.value = '';
}

// ─── File Parsing (Supports CSV, XLSX, and XLS) ───────────────────────────────
async function parseFile(file: File): Promise<{ headers: string[]; rows: Record<string, string>[] }> {
  try {
    const arrayBuffer = await file.arrayBuffer();
    const workbook = XLSX.read(arrayBuffer, { type: 'array' });
    
    if (!workbook.SheetNames || workbook.SheetNames.length === 0) {
      throw new Error('File tidak memiliki sheet data yang valid.');
    }
    
    const firstSheetName = workbook.SheetNames[0];
    const worksheet = workbook.Sheets[firstSheetName];
    const rawData = XLSX.utils.sheet_to_json<any[]>(worksheet, { header: 1, defval: '' });
    
    if (!rawData || rawData.length === 0) {
      return { headers: [], rows: [] };
    }
    
    // Find header row (first non-empty row)
    let headerRowIdx = 0;
    while (headerRowIdx < rawData.length && (!rawData[headerRowIdx] || rawData[headerRowIdx].every(c => String(c || '').trim() === ''))) {
      headerRowIdx++;
    }
    
    if (headerRowIdx >= rawData.length) {
      return { headers: [], rows: [] };
    }
    
    const rawHeaders = rawData[headerRowIdx] as any[];
    const headers: string[] = [];
    const headerIndices: number[] = [];
    
    rawHeaders.forEach((h, idx) => {
      const cleanH = String(h || '').trim();
      if (cleanH) {
        headers.push(cleanH);
        headerIndices.push(idx);
      }
    });
    
    const rows: Record<string, string>[] = [];
    for (let i = headerRowIdx + 1; i < rawData.length; i++) {
      const rowArr = rawData[i] as any[];
      if (!rowArr || rowArr.every(c => String(c || '').trim() === '')) {
        continue; // Skip empty row
      }
      
      const rowObj: Record<string, string> = {};
      headers.forEach((h, hIdx) => {
        const colIdx = headerIndices[hIdx];
        rowObj[h] = String(rowArr[colIdx] ?? '').trim();
      });
      rows.push(rowObj);
    }
    
    return { headers, rows };
  } catch (err: any) {
    console.error('Error parsing spreadsheet file:', err);
    toastStore.error('Failed to Read File', err.message || 'Invalid spreadsheet file format');
    return { headers: [], rows: [] };
  }
}

// ─── Default Column Auto-mapping ─────────────────────────────────────────────
const autoMappingRules: Record<string, keyof Device | ''> = {
  'device name': 'name',
  'nama perangkat': 'name',
  'nama': 'name',
  'device type': 'type',
  'tipe perangkat': 'type',
  'tipe': 'type',
  'addressing mode': 'addressingMode',
  'mode ip': 'addressingMode',
  'ip address': 'ip',
  'ip': 'ip',
  'mac address': 'mac',
  'mac': 'mac',
  'location': 'location',
  'lokasi': 'location',
  'rack': 'rack',
  'rak': 'rack',
  'snmp polling enabled': 'snmpEnabled',
  'snmp enabled': 'snmpEnabled',
  'snmp community': 'snmpCommunity',
  'snmp port': 'snmpPort',
  'custom failure threshold override': 'useCustomThreshold',
  'custom failure threshold value': 'customFailureThreshold'
};

function autoMapColumn(header: string): keyof Device | '' {
  const normalized = header.toLowerCase().trim();
  return autoMappingRules[normalized] ?? '';
}

// ─── Validation ───────────────────────────────────────────────────────────────
function isValidIp(ip: string): boolean {
  return /^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/.test(ip);
}

function isValidMacFormat(rawMac: string): boolean {
  const clean = rawMac.trim();
  const colonHyphenRegex = /^([0-9a-fA-F]{2}[:\-]){5}[0-9a-fA-F]{2}$/;
  const dotRegex = /^[0-9a-fA-F]{4}\.[0-9a-fA-F]{4}\.[0-9a-fA-F]{4}$/;
  const rawHexRegex = /^[0-9a-fA-F]{12}$/;
  
  return colonHyphenRegex.test(clean) || dotRegex.test(clean) || rawHexRegex.test(clean);
}

function normalizeMac(rawMac: string): string {
  const clean = rawMac.replace(/[^0-9a-fA-F]/g, '').toLowerCase();
  if (clean.length !== 12) return rawMac;
  const parts = [];
  for (let i = 0; i < 6; i++) {
    parts.push(clean.slice(i * 2, i * 2 + 2));
  }
  return parts.join(':');
}

function validateRows(rawRows: Record<string, string>[]): ImportRow[] {
  const existingMacs = new Set(deviceStore.devices.map((d: Device) => d.mac.toLowerCase()));
  const seenMacs = new Set<string>();
  const validTypes = new Set(['Access Point', 'Switch', 'SmartPower', 'Router', 'CCTV', 'NVR']);

  return rawRows.map((raw, idx) => {
    const mapped: Partial<Device> = {};

    // Apply column mappings
    for (const m of columnMappings.value) {
      if (!m.targetField || raw[m.sourceColumn] === undefined) continue;
      const val = raw[m.sourceColumn].trim();
      
      if (m.targetField === 'snmpEnabled' || m.targetField === 'useCustomThreshold') {
        const lower = val.toLowerCase();
        (mapped as any)[m.targetField] = (lower === 'true' || lower === 'yes' || lower === '1');
      } else if (m.targetField === 'snmpPort' || m.targetField === 'customFailureThreshold') {
        const num = parseInt(val, 10);
        (mapped as any)[m.targetField] = isNaN(num) ? undefined : num;
      } else {
        (mapped as any)[m.targetField] = val;
      }
    }

    // Name check
    if (!mapped.name) {
      return { rowIndex: idx, raw, mapped, status: 'missing_field', reason: 'Missing device name' };
    }
    
    // Type check
    if (!mapped.type) {
      return { rowIndex: idx, raw, mapped, status: 'missing_field', reason: 'Missing device type' };
    }
    if (!validTypes.has(mapped.type)) {
      return { rowIndex: idx, raw, mapped, status: 'failed', reason: `Unrecognized Device Type: '${mapped.type}'` };
    }

    // Addressing mode check
    const rawMode = (mapped.addressingMode as string || '').trim();
    if (!rawMode) {
      return { rowIndex: idx, raw, mapped, status: 'missing_field', reason: 'Missing addressing mode' };
    }
    
    let normalizedMode: 'Static' | 'DHCP';
    if (rawMode.toLowerCase() === 'static ip' || rawMode.toLowerCase() === 'static') {
      normalizedMode = 'Static';
    } else if (rawMode.toLowerCase() === 'dhcp reservation' || rawMode.toLowerCase() === 'dhcp') {
      normalizedMode = 'DHCP';
    } else {
      return { rowIndex: idx, raw, mapped, status: 'failed', reason: `Invalid addressing mode: '${rawMode}'` };
    }
    mapped.addressingMode = normalizedMode;

    // IP Check
    const rawIp = (mapped.ip || '').trim();
    if (normalizedMode === 'Static' && !rawIp) {
      return { rowIndex: idx, raw, mapped, status: 'missing_field', reason: 'IP Address is required for Static IP mode' };
    }
    if (rawIp && !isValidIp(rawIp)) {
      return { rowIndex: idx, raw, mapped, status: 'failed', reason: `Invalid IP address format: '${rawIp}'` };
    }
    mapped.ip = rawIp;

    // MAC Check
    const rawMac = (mapped.mac || '').trim();
    if (!rawMac) {
      return { rowIndex: idx, raw, mapped, status: 'missing_field', reason: 'Missing MAC address' };
    }
    if (!isValidMacFormat(rawMac)) {
      return { rowIndex: idx, raw, mapped, status: 'invalid_mac', reason: `Invalid MAC format: '${rawMac}'` };
    }
    
    const normalizedMac = normalizeMac(rawMac);
    mapped.mac = normalizedMac;

    if (existingMacs.has(normalizedMac.toLowerCase()) || seenMacs.has(normalizedMac.toLowerCase())) {
      return { rowIndex: idx, raw, mapped, status: 'duplicate_mac', reason: `Duplicate MAC: '${normalizedMac}'` };
    }
    seenMacs.add(normalizedMac.toLowerCase());

    // SNMP Check
    if (mapped.snmpEnabled) {
      if (mapped.snmpPort !== undefined && (mapped.snmpPort < 1 || mapped.snmpPort > 65535)) {
        return { rowIndex: idx, raw, mapped, status: 'failed', reason: `Invalid SNMP port: ${mapped.snmpPort}` };
      }
    }

    // Threshold Check
    if (mapped.useCustomThreshold) {
      if (mapped.customFailureThreshold === undefined || mapped.customFailureThreshold === null || mapped.customFailureThreshold < 1 || mapped.customFailureThreshold > 10) {
        return { rowIndex: idx, raw, mapped, status: 'failed', reason: `Invalid fails threshold: ${mapped.customFailureThreshold || ''} (must be 1-10)` };
      }
    }

    return { rowIndex: idx, raw, mapped, status: 'valid' };
  });
}

// ─── Step Transitions ────────────────────────────────────────────────────────
async function proceedFromUpload() {
  if (!selectedFile.value) return;
  const { headers } = await parseFile(selectedFile.value);
  if (headers.length === 0) {
    toastStore.error('Failed to Read File', 'No header column found in spreadsheet file.');
    return;
  }
  columnMappings.value = headers.map(h => ({
    sourceColumn: h,
    targetField: autoMapColumn(h) as any
  }));
  step.value = 1;
}

async function proceedFromMapping() {
  if (!selectedFile.value) return;
  const { rows } = await parseFile(selectedFile.value);
  if (rows.length === 0) {
    toastStore.error('No Data', 'No device rows found in the file.');
    return;
  }
  parsedRows.value = validateRows(rows);
  step.value = 2;
}

async function executeImport() {
  const results: ImportSummary['results'] = [];
  let imported = 0;
  let skipped = 0;
  let failed = 0;
  
  newlyCreatedLocations.value = [];
  matchedLocations.value = [];

  let existingLocsList: { id: string; name: string }[] = [];
  try {
    existingLocsList = await locationsApi.getLocations();
  } catch (err) {
    console.error('Failed to pre-fetch locations:', err);
  }

  const locationMap = new Map<string, string>();
  for (const loc of existingLocsList) {
    locationMap.set(loc.name.trim().toLowerCase(), loc.id);
  }

  const newLocsSet = new Set<string>();
  const matchedLocsSet = new Set<string>();

  for (const row of parsedRows.value) {
    if (row.status === 'valid') {
      try {
        let locId = '';
        let locName = (row.mapped.location || '').trim();
        
        if (locName) {
          const normLocName = locName.toLowerCase();
          if (locationMap.has(normLocName)) {
            locId = locationMap.get(normLocName)!;
            matchedLocsSet.add(locName);
          } else {
            try {
              const created = await locationsApi.createLocation(locName);
              locId = created.id;
              locationMap.set(normLocName, locId);
              newLocsSet.add(locName);
            } catch (err) {
              console.error(`Failed to create location '${locName}', fallback to database auto-creation:`, err);
              newLocsSet.add(locName);
            }
          }
        }

        await deviceStore.addDevice({
          name: row.mapped.name!,
          type: row.mapped.type!,
          ip: row.mapped.ip || '',
          mac: row.mapped.mac!,
          addressingMode: row.mapped.addressingMode!,
          locationId: locId || undefined,
          location: locName || '',
          rack: row.mapped.rack || '',
          snmpEnabled: row.mapped.snmpEnabled || false,
          snmpCommunity: row.mapped.snmpEnabled ? (row.mapped.snmpCommunity || 'public') : '',
          snmpPort: row.mapped.snmpEnabled ? (row.mapped.snmpPort || 161) : undefined,
          useCustomThreshold: row.mapped.useCustomThreshold || false,
          customFailureThreshold: row.mapped.useCustomThreshold ? (row.mapped.customFailureThreshold || 3) : null,
          failureThreshold: row.mapped.useCustomThreshold ? (row.mapped.customFailureThreshold || 3) : 3
        });
        results.push({ rowIndex: row.rowIndex, status: 'imported' });
        imported++;
      } catch (err: any) {
        results.push({ rowIndex: row.rowIndex, status: 'failed', reason: err.response?.data?.message || 'API error' });
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

  newlyCreatedLocations.value = Array.from(newLocsSet);
  matchedLocations.value = Array.from(matchedLocsSet);

  importSummary.value = { imported, skipped, failed, results };
  step.value = 3;

  if (imported > 0) {
    toastStore.success(
      'Import Successful',
      `Successfully imported ${imported} device(s)${skipped > 0 ? ` (${skipped} skipped)` : ''}${failed > 0 ? ` (${failed} failed)` : ''}.`
    );
  } else if (failed > 0) {
    toastStore.error(
      'Import Failed',
      `Failed to import ${failed} device(s). Please check the error log on screen.`
    );
  } else {
    toastStore.info('No Devices Imported', 'All rows were skipped or already registered.');
  }
}

function downloadExcelTemplate() {
  const data = [
    [
      'Device Name',
      'Device Type',
      'Addressing Mode',
      'IP Address',
      'MAC Address',
      'Location',
      'Rack',
      'SNMP Polling Enabled',
      'SNMP Community',
      'SNMP Port',
      'Custom Failure Threshold Override',
      'Custom Failure Threshold Value'
    ],
    [
      'AP Biro Umum',
      'Access Point',
      'Static IP',
      '10.20.1.18',
      '00:1A:2B:3C:4D:5E',
      'Gedung Sate',
      'Rack A',
      'FALSE',
      'public',
      161,
      'FALSE',
      3
    ],
    [
      'AP Core Switch',
      'Switch',
      'DHCP Reservation',
      '',
      '00:1A:2B:3C:4D:5F',
      'Gedung Sate',
      'Rack B',
      'TRUE',
      'public',
      161,
      'TRUE',
      5
    ]
  ];

  const worksheet = XLSX.utils.aoa_to_sheet(data);
  worksheet['!cols'] = [
    { wch: 18 }, // Device Name
    { wch: 14 }, // Device Type
    { wch: 18 }, // Addressing Mode
    { wch: 16 }, // IP Address
    { wch: 20 }, // MAC Address
    { wch: 16 }, // Location
    { wch: 12 }, // Rack
    { wch: 22 }, // SNMP Polling Enabled
    { wch: 16 }, // SNMP Community
    { wch: 12 }, // SNMP Port
    { wch: 32 }, // Custom Failure Threshold Override
    { wch: 30 }  // Custom Failure Threshold Value
  ];

  const workbook = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(workbook, worksheet, 'Devices');
  XLSX.writeFile(workbook, 'device_import_template.xlsx');
  toastStore.success('Template Excel Diunduh', 'File device_import_template.xlsx berhasil diunduh.');
}

function downloadCSVTemplate() {
  const headers = [
    'Device Name',
    'Device Type',
    'Addressing Mode',
    'IP Address',
    'MAC Address',
    'Location',
    'Rack',
    'SNMP Polling Enabled',
    'SNMP Community',
    'SNMP Port',
    'Custom Failure Threshold Override',
    'Custom Failure Threshold Value'
  ].join(',');
  const row1 = 'AP Biro Umum,Access Point,Static IP,10.20.1.18,00:1A:2B:3C:4D:5E,Gedung Sate,Rack A,FALSE,public,161,FALSE,3';
  const row2 = 'AP Core Switch,Switch,DHCP Reservation,,00:1A:2B:3C:4D:5F,Gedung Sate,Rack B,TRUE,public,161,TRUE,5';
  const csv = [headers, row1, row2].join('\n');
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'device_import_template.csv';
  a.click();
  URL.revokeObjectURL(url);
  toastStore.success('Template CSV Diunduh', 'File device_import_template.csv berhasil diunduh.');
}

// ─── Helpers ─────────────────────────────────────────────────────────────────
function rowStatusClass(status: string) {
  switch (status) {
    case 'valid': return 'bg-status-up/15 text-status-up border-status-up/30';
    case 'duplicate_mac': return 'bg-amber-500/15 text-amber-400 border-amber-500/30';
    default: return 'bg-status-down/15 text-status-down border-status-down/30';
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
  const failed = importSummary.value.results.filter((r: ImportResult) => r.status === 'failed' || r.status === 'skipped');
  const csv = ['Row,Status,Reason', ...failed.map((r: ImportResult) => `${r.rowIndex + 1},${r.status},"${r.reason}"`)].join('\n');
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
  setTimeout(() => {
    step.value = 0;
    selectedFile.value = null;
    parsedRows.value = [];
    columnMappings.value = [];
    importSummary.value = { imported: 0, skipped: 0, failed: 0, results: [] };
    newlyCreatedLocations.value = [];
    matchedLocations.value = [];
  }, 300);
}

// Reset on open
watch(() => props.isOpen, (val) => {
  if (val) {
    step.value = 0;
    selectedFile.value = null;
    parsedRows.value = [];
    newlyCreatedLocations.value = [];
    matchedLocations.value = [];
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
