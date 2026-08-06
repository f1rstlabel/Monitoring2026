<template>
  <Modal :is-open="isOpen" :title="mode === 'add' ? 'Add New Device' : 'Edit Device Configuration'" @close="$emit('close')">
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
          DHCP mode: MAC address is required. IP will be resolved automatically by the poller.
        </p>
      </div>

      <!-- IP Address — required only for Static -->
      <div class="space-y-1.5">
        <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium">
          IP Address {{ form.addressingMode === 'Static' ? '*' : '(optional — last seen)' }}
        </label>
        <div class="flex gap-2">
          <input
            v-model="form.ip"
            type="text"
            :required="form.addressingMode === 'Static'"
            :placeholder="form.addressingMode === 'Static' ? 'e.g. 10.20.1.18' : 'Auto-filled by poller or Auto Detect'"
            class="flex-1 bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 font-mono text-gray-200 focus:outline-none focus:border-[#7B96F5]"
          />
          <button
            v-if="form.addressingMode === 'DHCP' && form.mac"
            type="button"
            :disabled="isDetecting"
            @click="autoDetect"
            class="px-2.5 py-2 rounded-lg bg-[#7B96F5]/15 border border-[#7B96F5]/30 text-[#7B96F5] hover:bg-[#7B96F5]/25 transition-colors flex items-center gap-1.5 text-[11px] font-medium disabled:opacity-50 shrink-0"
            title="Resolve current IP from ARP table for this MAC"
          >
            <RefreshCw class="w-3.5 h-3.5" :class="isDetecting ? 'animate-spin' : ''" />
            {{ isDetecting ? '...' : 'Detect' }}
          </button>
        </div>
        <p v-if="detectError" class="text-[10px] text-amber-400">{{ detectError }}</p>
      </div>

      <!-- MAC Address — required for DHCP -->
      <div class="space-y-1.5">
        <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium">
          MAC Address {{ form.addressingMode === 'DHCP' ? '*' : '' }}
        </label>
        <div class="flex gap-2">
          <input
            v-model="form.mac"
            type="text"
            :required="form.addressingMode === 'DHCP'"
            placeholder="e.g. 00:1A:2B:3C:4D:5E"
            class="flex-1 bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 font-mono text-gray-200 focus:outline-none focus:border-[#7B96F5]"
          />
          <button
            v-if="form.addressingMode === 'Static' && form.ip"
            type="button"
            :disabled="isDetecting"
            @click="autoDetect"
            class="px-2.5 py-2 rounded-lg bg-[#7B96F5]/15 border border-[#7B96F5]/30 text-[#7B96F5] hover:bg-[#7B96F5]/25 transition-colors flex items-center gap-1.5 text-[11px] font-medium disabled:opacity-50 shrink-0"
            title="Resolve MAC from ARP table for this IP"
          >
            <RefreshCw class="w-3.5 h-3.5" :class="isDetecting ? 'animate-spin' : ''" />
            {{ isDetecting ? '...' : 'Detect' }}
          </button>
        </div>
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

      <!-- CCTV / NVR Camera Stream Configuration -->
      <div v-if="form.type === 'CCTV' || form.type === 'NVR'" class="col-span-1 md:col-span-2 border-t border-[#26262A] pt-4 mt-2 space-y-3">
        <div>
          <p class="text-xs font-semibold text-white">Live Stream & Snapshot Configuration</p>
          <p class="text-[10px] text-gray-400">Configure connection details for live camera viewing inside NOC dashboard</p>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3 bg-[#18181B] border border-[#26262A] rounded-lg p-3">
          <div>
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium mb-1">RTSP Stream URL</label>
            <input
              v-model="form.streamUrl"
              type="text"
              placeholder="e.g. rtsp://10.20.10.5:554/stream1"
              class="w-full bg-[#0A0A0B] border border-[#26262A] rounded px-2.5 py-1.5 text-xs text-white focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
          <div>
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium mb-1">Snapshot Image URL</label>
            <input
              v-model="form.snapshotUrl"
              type="text"
              placeholder="e.g. http://10.20.10.5/snapshot.jpg"
              class="w-full bg-[#0A0A0B] border border-[#26262A] rounded px-2.5 py-1.5 text-xs text-white focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
          <div>
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium mb-1">Stream Username</label>
            <input
              v-model="form.streamUsername"
              type="text"
              placeholder="admin"
              class="w-full bg-[#0A0A0B] border border-[#26262A] rounded px-2.5 py-1.5 text-xs text-white focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
          <div>
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-medium mb-1">Stream Password</label>
            <input
              v-model="form.streamPassword"
              type="password"
              placeholder="••••••••"
              class="w-full bg-[#0A0A0B] border border-[#26262A] rounded px-2.5 py-1.5 text-xs text-white focus:outline-none focus:border-[#7B96F5]"
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
          @click="$emit('close')"
          class="px-4 py-2 rounded-lg border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-medium transition-colors"
        >
          Cancel
        </button>
        <button
          type="button"
          :disabled="isSubmitting"
          @click="handleSubmit"
          class="px-5 py-2 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 transition-all flex items-center gap-2 disabled:opacity-50"
        >
          <Save class="w-4 h-4" />
          {{ isSubmitting ? 'Saving...' : (mode === 'add' ? 'Save Device' : 'Save Changes') }}
        </button>
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import Modal from '../common/Modal.vue';
import LocationCombobox from './LocationCombobox.vue';
import { PlusCircle, Pencil, Sliders, Save, RefreshCw } from 'lucide-vue-next';
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
const formError = ref('');

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
  failureThreshold: 3,
  streamUrl: '',
  snapshotUrl: '',
  streamUsername: '',
  streamPassword: ''
});

watch(
  () => props.device,
  (newDev) => {
    if (props.mode === 'edit' && newDev) {
      form.name = newDev.name || '';
      form.type = newDev.type || 'Access Point';
      form.ip = newDev.ip || '';
      form.mac = newDev.mac || '';
      form.addressingMode = newDev.addressingMode || 'Static';
      form.locationId = newDev.locationId || '';
      form.location = newDev.location || '';
      form.rack = newDev.rack || '';
      form.snmpEnabled = newDev.snmpEnabled || false;
      form.snmpCommunity = newDev.snmpCommunity || 'public';
      form.snmpPort = newDev.snmpPort || 161;
      form.snmpIfIndex = newDev.snmpIfIndex;
      form.useCustomThreshold = newDev.useCustomThreshold ?? (newDev.customFailureThreshold !== undefined && newDev.customFailureThreshold !== null);
      form.failureThreshold = newDev.customFailureThreshold ?? newDev.failureThreshold ?? 3;
      form.streamUrl = newDev.streamUrl || '';
      form.snapshotUrl = newDev.snapshotUrl || '';
      form.streamUsername = newDev.streamUsername || '';
      form.streamPassword = newDev.streamPassword || '';
    } else if (props.mode === 'add') {
      form.name = '';
      form.type = 'Access Point';
      form.ip = '';
      form.mac = '';
      form.addressingMode = 'Static';
      form.locationId = '';
      form.location = '';
      form.rack = '';
      form.snmpEnabled = false;
      form.snmpCommunity = 'public';
      form.snmpPort = 161;
      form.snmpIfIndex = undefined;
      form.useCustomThreshold = false;
      form.failureThreshold = 3;
      form.streamUrl = '';
      form.snapshotUrl = '';
      form.streamUsername = '';
      form.streamPassword = '';
    }
  },
  { immediate: true }
);

async function autoDetect() {
  detectError.value = '';
  isDetecting.value = true;
  try {
    const params = form.addressingMode === 'Static' ? { ip: form.ip } : { mac: form.mac };
    const res = await api.get('/devices/auto-detect', { params });
    if (res.data.found) {
      if (form.addressingMode === 'Static') {
        form.mac = res.data.mac;
      } else {
        form.ip = res.data.ip;
      }
    } else {
      detectError.value = res.data.message || 'Device not found in ARP table.';
    }
  } catch (e: any) {
    detectError.value = e.response?.data?.message || 'Failed to detect device from ARP table.';
  } finally {
    isDetecting.value = false;
  }
}

async function handleSubmit() {
  formError.value = '';

  if (!form.name.trim()) {
    formError.value = 'Device name is required';
    return;
  }
  if (form.addressingMode === 'Static' && !form.ip) {
    formError.value = 'IP address is required for Static addressing mode';
    return;
  }
  if (form.addressingMode === 'DHCP' && !form.mac) {
    formError.value = 'MAC address is required for DHCP addressing mode';
    return;
  }

  isSubmitting.value = true;

  try {
    const payload = {
      name: form.name,
      type: form.type,
      ip: form.ip,
      mac: form.mac,
      addressingMode: form.addressingMode,
      location: form.location,
      rack: form.rack,
      snmpEnabled: form.snmpEnabled,
      snmpCommunity: form.snmpCommunity,
      snmpPort: form.snmpPort,
      snmpIfIndex: form.snmpIfIndex,
      useCustomThreshold: form.useCustomThreshold,
      customFailureThreshold: form.useCustomThreshold ? form.failureThreshold : null,
      failureThreshold: form.useCustomThreshold ? form.failureThreshold : 3,
      streamUrl: form.streamUrl,
      snapshotUrl: form.snapshotUrl,
      streamUsername: form.streamUsername,
      streamPassword: form.streamPassword
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
