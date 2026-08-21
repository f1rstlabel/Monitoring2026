<template>
  <Modal :is-open="isOpen" :title="mode === 'add' ? 'Add New Device' : 'Edit Device Configuration'" @close="handleAttemptClose">
    <template #icon>
      <PlusCircle v-if="mode === 'add'" class="w-5 h-5 text-[#7B96F5]" />
      <Pencil v-else class="w-5 h-5 text-[#7B96F5]" />
    </template>

    <form @submit.prevent="handleSubmit" class="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
      <!-- Device Name -->
      <div class="space-y-1.5">
        <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium">Device Name *</label>
        <input
          v-model="form.name"
          type="text"
          required
          placeholder="e.g. AP Biro Umum Lt 2 Conference"
          class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
        />
      </div>

      <!-- Device Type Dropdown -->
      <div class="space-y-1.5">
        <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium">Device Type *</label>
        <select
          v-model="form.type"
          required
          class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
        >
          <option value="Access Point">Access Point</option>
          <option value="Switch">Switch</option>
          <option value="Router">Router</option>
          <option value="SmartPower">SmartPower</option>
          <option value="CCTV">CCTV</option>
          <option value="NVR">NVR</option>
        </select>
      </div>

      <!-- Addressing Mode (Radio) — full width row -->
      <div class="space-y-1.5 col-span-1 md:col-span-2">
        <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium">Addressing Mode *</label>
        <div class="flex items-center gap-6 pt-1">
          <label class="flex items-center gap-2 cursor-pointer text-gray-300">
            <input type="radio" v-model="form.addressingMode" value="Static" class="accent-[#7B96F5]" />
            <span>Static IP</span>
          </label>
          <label class="flex items-center gap-2 cursor-pointer text-gray-300">
            <input type="radio" v-model="form.addressingMode" value="DHCP" class="accent-[#7B96F5]" />
            <span>DHCP Reservation</span>
          </label>
        </div>
        <p v-if="form.addressingMode === 'DHCP'" class="text-[10px] text-amber-400">
          DHCP mode: MAC address is required. IP will be resolved automatically from Core Switch ARP.
        </p>
      </div>

      <!-- IP Address — required only for Static -->
      <div class="space-y-1.5">
        <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium flex items-center justify-between">
          <span>IP Address {{ form.addressingMode === 'Static' ? '*' : '(auto-resolved)' }}</span>
          <span v-if="isDetecting && form.addressingMode === 'Static'" class="text-[#7B96F5] text-[10px] font-normal lowercase flex items-center gap-1">
            <RefreshCw class="w-2.5 h-2.5 animate-spin" /> scanning arp...
          </span>
        </label>
        <div class="relative">
          <input
            v-model="form.ip"
            type="text"
            :required="form.addressingMode === 'Static'"
            :placeholder="form.addressingMode === 'Static' ? 'e.g. 10.20.1.18' : 'Auto-detected from MAC'"
            class="w-full bg-[#18181B] border border-[#26262A] rounded-lg pl-3 pr-9 py-2 font-mono text-gray-200 focus:outline-none focus:border-[#7B96F5]"
          />
          <div class="absolute right-2.5 top-1/2 -translate-y-1/2 flex items-center">
            <button
              v-if="form.ip"
              type="button"
              :disabled="isDetecting"
              @click="autoDetectManual"
              class="p-1 text-gray-400 hover:text-[#7B96F5] transition-colors rounded cursor-pointer disabled:opacity-50"
              title="Click to scan MAC address from ARP"
            >
              <RefreshCw class="w-3.5 h-3.5" :class="isDetecting ? 'animate-spin text-[#7B96F5]' : ''" />
            </button>
          </div>
        </div>
        <p v-if="detectSuccess && form.addressingMode === 'DHCP'" class="text-[10px] text-emerald-400 font-mono flex items-center gap-1">
          <CheckCircle2 class="w-3 h-3 shrink-0" /> {{ detectSuccess }}
        </p>
        <p v-if="detectError && form.addressingMode === 'DHCP'" class="text-[10px] text-amber-400 font-mono flex items-center gap-1">
          <AlertTriangle class="w-3 h-3 shrink-0" /> {{ detectError }}
        </p>
      </div>

      <!-- MAC Address — required for DHCP -->
      <div class="space-y-1.5">
        <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium flex items-center justify-between">
          <span>MAC Address {{ form.addressingMode === 'DHCP' ? '*' : '' }}</span>
          <span v-if="isDetecting && form.addressingMode === 'DHCP'" class="text-[#7B96F5] text-[10px] font-normal lowercase flex items-center gap-1">
            <RefreshCw class="w-2.5 h-2.5 animate-spin" /> scanning core switch...
          </span>
        </label>
        <div class="relative">
          <input
            v-model="form.mac"
            type="text"
            :required="form.addressingMode === 'DHCP'"
            placeholder="e.g. 00:1A:2B:3C:4D:5E"
            class="w-full bg-[#18181B] border border-[#26262A] rounded-lg pl-3 pr-9 py-2 font-mono text-gray-200 focus:outline-none focus:border-[#7B96F5]"
          />
          <div class="absolute right-2.5 top-1/2 -translate-y-1/2 flex items-center">
            <button
              v-if="form.mac"
              type="button"
              :disabled="isDetecting"
              @click="autoDetectManual"
              class="p-1 text-gray-400 hover:text-[#7B96F5] transition-colors rounded cursor-pointer disabled:opacity-50"
              title="Click to scan IP address from Core Switch ARP"
            >
              <RefreshCw class="w-3.5 h-3.5" :class="isDetecting ? 'animate-spin text-[#7B96F5]' : ''" />
            </button>
          </div>
        </div>
        <p v-if="detectSuccess && form.addressingMode === 'Static'" class="text-[10px] text-emerald-400 font-mono flex items-center gap-1">
          <CheckCircle2 class="w-3 h-3 shrink-0" /> {{ detectSuccess }}
        </p>
        <p v-if="detectError && form.addressingMode === 'Static'" class="text-[10px] text-amber-400 font-mono flex items-center gap-1">
          <AlertTriangle class="w-3 h-3 shrink-0" /> {{ detectError }}
        </p>
      </div>

      <!-- Location & Rack -->
      <div class="space-y-1.5">
        <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium">Location & Rack</label>
        <div class="grid grid-cols-2 gap-2">
          <LocationCombobox
            v-model="form.location"
            v-model:locationId="form.locationId"
          />
          <input
            v-model="form.rack"
            type="text"
            placeholder="Rack B-04"
            class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
          />
        </div>
      </div>

      <!-- SNMP Polling Configuration -->
      <div class="col-span-1 md:col-span-2 border-t border-[#26262A] pt-4 mt-2 space-y-3">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-white">SNMP Polling</p>
            <p class="text-[10px] text-gray-400">Query sysUpTime, ifOperStatus, sysName via SNMP v2c (falls back to ICMP)</p>
          </div>
          <button
            type="button"
            @click="form.snmpEnabled = !form.snmpEnabled"
            class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
            :class="form.snmpEnabled ? 'bg-[#7B96F5]' : 'bg-[#26262A]'"
          >
            <span
              class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
              :class="form.snmpEnabled ? 'translate-x-4' : 'translate-x-0'"
            />
          </button>
        </div>

        <div v-if="form.snmpEnabled" class="grid grid-cols-1 md:grid-cols-3 gap-3 bg-[#18181B] border border-[#26262A] rounded-lg p-3">
          <div>
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium mb-1">Community String</label>
            <input
              v-model="form.snmpCommunity"
              type="password"
              placeholder="public"
              class="w-full bg-[#0A0A0B] border border-[#26262A] rounded px-2.5 py-1.5 text-xs text-white focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
          <div>
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium mb-1">Port</label>
            <input
              v-model.number="form.snmpPort"
              type="number"
              min="1"
              max="65535"
              placeholder="161"
              class="w-full bg-[#0A0A0B] border border-[#26262A] rounded px-2.5 py-1.5 text-xs text-white font-mono focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
          <div>
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium mb-1">IfIndex (Optional)</label>
            <input
              v-model.number="form.snmpIfIndex"
              type="number"
              min="1"
              placeholder="1"
              class="w-full bg-[#0A0A0B] border border-[#26262A] rounded px-2.5 py-1.5 text-xs text-white font-mono focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
        </div>
      </div>



      <!-- Custom Failure Threshold Override Section -->
      <div class="col-span-1 md:col-span-2 border-t border-[#26262A] pt-4 space-y-3">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-white">Custom Failure Threshold Override</p>
            <p class="text-[10px] text-gray-400">Override default consecutive ICMP failure count before marking DOWN</p>
          </div>
          <button
            type="button"
            @click="form.useCustomThreshold = !form.useCustomThreshold"
            class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
            :class="form.useCustomThreshold ? 'bg-[#7B96F5]' : 'bg-[#26262A]'"
          >
            <span
              class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
              :class="form.useCustomThreshold ? 'translate-x-4' : 'translate-x-0'"
            />
          </button>
        </div>

        <div v-if="form.useCustomThreshold" class="flex items-center gap-3 bg-[#18181B] border border-[#26262A] rounded-lg p-3">
          <Sliders class="w-4 h-4 text-[#7B96F5] shrink-0" />
          <div class="flex-1">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium mb-1">Consecutive Fails Threshold</label>
            <input
              v-model.number="form.failureThreshold"
              type="number"
              min="1"
              max="10"
              required
              class="w-24 bg-[#0A0A0B] border border-[#26262A] rounded px-2 py-1 text-xs text-white font-mono text-center focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
          <span class="text-[10px] text-gray-500 font-mono">1 = Immediate, 10 = Max Debounce</span>
        </div>
      </div>

      <!-- Validation error banner -->
      <div v-if="formError" class="col-span-1 md:col-span-2 bg-[#F16565]/10 border border-[#F16565]/30 rounded-lg px-3 py-2 text-[11px] text-[#F16565]">
        {{ formError }}
      </div>
    </form>

    <template #footer>
      <div class="flex items-center justify-end gap-3 w-full">
        <button
          type="button"
          @click="handleAttemptClose"
          class="px-4 py-2 rounded-lg border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-medium transition-colors cursor-pointer"
        >
          Cancel
        </button>
        <button
          type="button"
          :disabled="isSubmitting"
          @click="handleSubmit"
          class="px-5 py-2 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 transition-all flex items-center gap-2 disabled:opacity-50 cursor-pointer"
        >
          <Save class="w-4 h-4" />
          {{ isSubmitting ? 'Saving...' : (mode === 'add' ? 'Save Device' : 'Save Changes') }}
        </button>
      </div>
    </template>
  </Modal>

  <!-- Discard Confirmation Modal -->
  <Modal :is-open="showDiscardConfirm" title="Discard Unsaved Changes?" @close="showDiscardConfirm = false">
    <template #default>
      <div class="p-4 bg-amber-500/10 border border-amber-500/30 rounded-xl flex items-start gap-3 text-amber-400 text-xs">
        <AlertTriangle class="w-5 h-5 shrink-0 mt-0.5" />
        <div class="space-y-1">
          <h4 class="font-bold text-white text-sm">Unsaved Changes</h4>
          <p class="text-gray-300 leading-relaxed">
            You have unsaved changes in this device configuration. Are you sure you want to discard them and revert to original values?
          </p>
        </div>
      </div>
    </template>
    <template #footer>
      <div class="flex items-center justify-end gap-3 w-full">
        <button
          type="button"
          @click="showDiscardConfirm = false"
          class="px-4 py-2 rounded-xl border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-mono cursor-pointer"
        >
          Keep Editing
        </button>
        <button
          type="button"
          @click="confirmDiscardAndClose"
          class="px-5 py-2 rounded-xl bg-red-500 hover:bg-red-600 text-white font-semibold text-xs shadow-md shadow-red-500/20 font-mono cursor-pointer"
        >
          Discard &amp; Revert
        </button>
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue';
import Modal from '../common/Modal.vue';
import LocationCombobox from './LocationCombobox.vue';
import { PlusCircle, Pencil, Sliders, Save, RefreshCw, CheckCircle2, AlertTriangle } from 'lucide-vue-next';
import type { Device, DeviceType, AddressingMode } from '../../types';
import { useDeviceStore } from '../../stores/deviceStore';
import api from '../../api/client';

const props = defineProps<{
  isOpen: boolean;
  mode: 'add' | 'edit';
  device?: Device | null;
}>();

const emit = defineEmits(['close', 'saved']);
const deviceStore = useDeviceStore();

const isSubmitting = ref(false);
const isDetecting = ref(false);
const detectError = ref('');
const detectSuccess = ref('');
const formError = ref('');
const showDiscardConfirm = ref(false);

const form = reactive({
  name: '',
  type: 'Access Point' as DeviceType,
  ip: '',
  mac: '',
  addressingMode: 'Static' as AddressingMode,
  locationId: '',
  location: '',
  rack: '',
  snmpEnabled: false,
  snmpCommunity: 'public',
  snmpPort: 161,
  snmpIfIndex: undefined as number | undefined,
  useCustomThreshold: false,
  failureThreshold: 3
});

// Snapshot of initial values to track dirty changes and allow clean rollback
const initialSnapshot = ref<any>(null);

function buildSnapshot(dev?: Device | null, mode: 'add' | 'edit' = 'add') {
  if (mode === 'edit' && dev) {
    return {
      name: dev.name || '',
      type: dev.type || 'Access Point',
      ip: dev.ip || '',
      mac: dev.mac || '',
      addressingMode: dev.addressingMode || 'Static',
      locationId: dev.locationId || '',
      location: dev.location || '',
      rack: dev.rack || '',
      snmpEnabled: dev.snmpEnabled || false,
      snmpCommunity: dev.snmpCommunity || 'public',
      snmpPort: dev.snmpPort || 161,
      snmpIfIndex: dev.snmpIfIndex,
      useCustomThreshold: dev.useCustomThreshold ?? (dev.customFailureThreshold !== undefined && dev.customFailureThreshold !== null),
      failureThreshold: dev.customFailureThreshold ?? dev.failureThreshold ?? 3
    };
  }
  return {
    name: '',
    type: 'Access Point' as DeviceType,
    ip: '',
    mac: '',
    addressingMode: 'Static' as AddressingMode,
    locationId: '',
    location: '',
    rack: '',
    snmpEnabled: false,
    snmpCommunity: 'public',
    snmpPort: 161,
    snmpIfIndex: undefined,
    useCustomThreshold: false,
    failureThreshold: 3
  };
}

const isDirty = computed(() => {
  if (!initialSnapshot.value) return false;
  const s = initialSnapshot.value;
  return (
    form.name !== s.name ||
    form.type !== s.type ||
    form.ip !== s.ip ||
    form.mac !== s.mac ||
    form.addressingMode !== s.addressingMode ||
    form.locationId !== s.locationId ||
    form.location !== s.location ||
    form.rack !== s.rack ||
    form.snmpEnabled !== s.snmpEnabled ||
    form.snmpCommunity !== s.snmpCommunity ||
    form.snmpPort !== s.snmpPort ||
    form.snmpIfIndex !== s.snmpIfIndex ||
    form.useCustomThreshold !== s.useCustomThreshold ||
    form.failureThreshold !== s.failureThreshold
  );
});

function handleAttemptClose() {
  if (isDirty.value) {
    showDiscardConfirm.value = true;
  } else {
    confirmDiscardAndClose();
  }
}

function confirmDiscardAndClose() {
  if (initialSnapshot.value) {
    Object.assign(form, JSON.parse(JSON.stringify(initialSnapshot.value)));
  }
  showDiscardConfirm.value = false;
  detectSuccess.value = '';
  detectError.value = '';
  formError.value = '';
  emit('close');
}

let debounceTimer: any = null;
const ipv4Regex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
const macRegex = /^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$/;

// Debounced auto-detect on IP input (Static Mode) & re-detect when cleared/partial in DHCP mode
watch(
  () => form.ip,
  (newIP) => {
    if (form.addressingMode === 'DHCP') {
      // In DHCP mode, if IP was cleared OR partially deleted while MAC is valid, re-fetch active IP!
      const trimmedIP = (newIP || '').trim();
      const trimmedMAC = (form.mac || '').trim();
      if (!ipv4Regex.test(trimmedIP) && macRegex.test(trimmedMAC)) {
        if (debounceTimer) clearTimeout(debounceTimer);
        debounceTimer = setTimeout(() => {
          autoDetectSilently();
        }, 500);
      }
      return;
    }

    if (form.addressingMode !== 'Static') return;
    if (debounceTimer) clearTimeout(debounceTimer);
    detectSuccess.value = '';
    detectError.value = '';

    const trimmed = (newIP || '').trim();
    if (!ipv4Regex.test(trimmed)) return;

    debounceTimer = setTimeout(() => {
      autoDetectSilently();
    }, 350);
  }
);

// Debounced auto-detect on MAC input (DHCP Mode) & re-detect when cleared/partial in Static mode
watch(
  () => form.mac,
  (newMAC) => {
    if (form.addressingMode === 'Static') {
      // In Static mode, if user clears OR partially deletes MAC while IP is valid, re-detect MAC!
      const trimmedMAC = (newMAC || '').trim();
      const trimmedIP = (form.ip || '').trim();
      if (!macRegex.test(trimmedMAC) && ipv4Regex.test(trimmedIP)) {
        if (debounceTimer) clearTimeout(debounceTimer);
        debounceTimer = setTimeout(() => {
          autoDetectSilently();
        }, 500);
      }
      return;
    }

    if (form.addressingMode !== 'DHCP') return;
    if (debounceTimer) clearTimeout(debounceTimer);
    detectSuccess.value = '';
    detectError.value = '';

    const trimmed = (newMAC || '').trim();
    if (!macRegex.test(trimmed)) return;

    debounceTimer = setTimeout(() => {
      autoDetectSilently();
    }, 350);
  }
);

// Auto-detect when switching addressing mode
watch(
  () => form.addressingMode,
  (newMode) => {
    detectSuccess.value = '';
    detectError.value = '';
    if (newMode === 'Static') {
      const trimmedIP = (form.ip || '').trim();
      if (ipv4Regex.test(trimmedIP)) {
        autoDetectSilently();
      }
    } else if (newMode === 'DHCP') {
      const trimmedMAC = (form.mac || '').trim();
      if (macRegex.test(trimmedMAC)) {
        autoDetectSilently();
      }
    }
  }
);

async function autoDetectSilently() {
  if (isDetecting.value) return;
  isDetecting.value = true;
  detectError.value = '';
  detectSuccess.value = '';

  try {
    const params = form.addressingMode === 'Static' ? { ip: form.ip.trim() } : { mac: form.mac.trim() };
    const res = await api.get('/devices/auto-detect', { params });
    if (res.data.found) {
      if (form.addressingMode === 'Static' && res.data.mac) {
        form.mac = res.data.mac;
        detectSuccess.value = `✓ MAC auto-detected: ${res.data.mac}`;
      } else if (form.addressingMode === 'DHCP' && res.data.ip) {
        form.ip = res.data.ip;
        detectSuccess.value = `✓ IP auto-detected: ${res.data.ip}`;
      }
    }
  } catch (e: any) {
    // Silent for background typing
  } finally {
    isDetecting.value = false;
  }
}

async function autoDetectManual() {
  detectError.value = '';
  detectSuccess.value = '';
  isDetecting.value = true;
  try {
    const params = form.addressingMode === 'Static' ? { ip: form.ip.trim() } : { mac: form.mac.trim() };
    const res = await api.get('/devices/auto-detect', { params });
    if (res.data.found) {
      if (form.addressingMode === 'Static' && res.data.mac) {
        form.mac = res.data.mac;
        detectSuccess.value = `✓ MAC detected: ${res.data.mac}`;
      } else if (form.addressingMode === 'DHCP' && res.data.ip) {
        form.ip = res.data.ip;
        detectSuccess.value = `✓ IP detected: ${res.data.ip}`;
      } else {
        detectError.value = 'Found ARP record but target field is empty.';
      }
    } else {
      detectError.value = res.data.message || (form.addressingMode === 'Static' ? 'MAC address not found in ARP table.' : 'IP address not found for this MAC.');
    }
  } catch (e: any) {
    detectError.value = e.response?.data?.message || 'Failed to query ARP table.';
  } finally {
    isDetecting.value = false;
  }
}

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      detectSuccess.value = '';
      detectError.value = '';
      formError.value = '';
      showDiscardConfirm.value = false;
      const snap = buildSnapshot(props.device, props.mode);
      initialSnapshot.value = JSON.parse(JSON.stringify(snap));
      Object.assign(form, snap);
    }
  },
  { immediate: true }
);

watch(
  () => props.device,
  (newDev) => {
    if (props.isOpen) {
      const snap = buildSnapshot(newDev, props.mode);
      initialSnapshot.value = JSON.parse(JSON.stringify(snap));
      Object.assign(form, snap);
    }
  }
);

async function handleSubmit() {
  formError.value = '';

  if (!form.name.trim()) {
    formError.value = 'Device name is required';
    return;
  }
  if (form.addressingMode === 'Static' && !form.ip.trim()) {
    formError.value = 'IP address is required for Static addressing mode';
    return;
  }
  if (form.addressingMode === 'DHCP' && !form.mac.trim()) {
    formError.value = 'MAC address is required for DHCP addressing mode';
    return;
  }

  isSubmitting.value = true;

  try {
    const payload = {
      name: form.name.trim(),
      type: form.type,
      ip: form.ip.trim(),
      mac: form.mac.trim(),
      addressingMode: form.addressingMode,
      locationId: form.locationId || undefined,
      location: form.location.trim(),
      rack: form.rack.trim(),
      snmpEnabled: form.snmpEnabled,
      snmpCommunity: form.snmpCommunity,
      snmpPort: form.snmpPort,
      snmpIfIndex: form.snmpIfIndex,
      useCustomThreshold: form.useCustomThreshold,
      customFailureThreshold: form.useCustomThreshold ? form.failureThreshold : null,
      failureThreshold: form.useCustomThreshold ? form.failureThreshold : 3
    };

    if (props.mode === 'add') {
      await deviceStore.addDevice(payload);
    } else if (props.mode === 'edit' && props.device) {
      await deviceStore.updateDevice(props.device.id, payload);
    }

    emit('saved');
    emit('close');
  } catch (e: any) {
    formError.value = e.response?.data?.message || 'Failed to save device configuration.';
  } finally {
    isSubmitting.value = false;
  }
}
</script>
