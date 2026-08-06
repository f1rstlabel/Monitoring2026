<template>
  <div class="space-y-6">
    <!-- Header Title & Action -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-extrabold text-white tracking-tight">Device Management</h1>
        <p class="text-xs text-gray-400 mt-1">
          Inventory and active monitoring parameters for all
          <span class="font-mono text-[#7B96F5]">{{ deviceStore.summary.totalDevices }}</span>
          infrastructure nodes
        </p>
      </div>

      <!-- Action Buttons — role-gated via v-if -->
      <div class="flex items-center gap-2.5">
        <!-- Import: superadmin only -->
        <button
          v-if="authStore.canImportDevices"
          @click="isImportModalOpen = true"
          class="px-3.5 py-2 rounded-lg border border-[#26262A] bg-[#151517] hover:bg-[#1E1E22] text-gray-300 hover:text-white font-semibold text-xs transition-all flex items-center gap-2"
        >
          <Upload class="w-4 h-4" />
          Import CSV / Excel
        </button>
        <!-- Add: superadmin + anggota -->
        <button
          v-if="authStore.canAddDevice"
          @click="openAddDevice()"
          class="px-4 py-2 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 transition-all flex items-center gap-2"
        >
          <Plus class="w-4 h-4" />
          Add New Device
        </button>
      </div>
    </div>

    <!-- Filter Bar & View Toggle -->
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-4 flex flex-wrap items-center justify-between gap-4">
      <div class="flex flex-wrap items-center gap-3 flex-1">
        <div class="relative flex-1 min-w-[240px]">
          <Search class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            v-model="deviceStore.searchQuery"
            type="text"
            placeholder="Search by name, IP, or MAC address..."
            class="w-full bg-[#18181B] border border-[#26262A] rounded-lg pl-9 pr-4 py-2 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5] placeholder-gray-600"
          />
        </div>

        <select
          v-model="deviceStore.selectedTypeFilter"
          class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
        >
          <option value="All">All Device Types</option>
          <option value="Access Point">Access Point</option>
          <option value="Switch">Switch</option>
          <option value="Router">Router</option>
          <option value="SmartPower">SmartPower</option>
          <option value="CCTV">CCTV</option>
          <option value="NVR">NVR</option>
        </select>

        <select
          v-model="deviceStore.selectedStatusFilter"
          class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
        >
          <option value="All">All Statuses</option>
          <option value="UP">UP Only</option>
          <option value="DOWN">DOWN Only</option>
        </select>

        <button
          v-if="hasActiveFilters"
          @click="clearFilters"
          class="text-xs text-gray-400 hover:text-gray-200 flex items-center gap-1 transition-colors"
        >
          <X class="w-3.5 h-3.5" />
          Clear
        </button>
      </div>

      <!-- View Toggle Buttons (Grouped by Location vs Flat List) -->
      <div class="flex items-center gap-2">
        <div class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5">
          <button
            @click="viewMode = 'grouped'"
            class="px-3 py-1.5 rounded-md text-xs font-semibold flex items-center gap-1.5 transition-all"
            :class="viewMode === 'grouped' ? 'bg-[#26262A] text-white shadow-sm' : 'text-gray-400 hover:text-gray-200'"
          >
            <MapPin class="w-3.5 h-3.5" />
            Grouped by Location
          </button>
          <button
            @click="viewMode = 'flat'"
            class="px-3 py-1.5 rounded-md text-xs font-semibold flex items-center gap-1.5 transition-all"
            :class="viewMode === 'flat' ? 'bg-[#26262A] text-white shadow-sm' : 'text-gray-400 hover:text-gray-200'"
          >
            <List class="w-3.5 h-3.5" />
            Flat List
          </button>
        </div>

        <span class="text-xs font-mono text-gray-400 border-l border-[#26262A] pl-3 ml-1">
          Showing <span class="text-gray-200 font-semibold">{{ deviceStore.filteredDevices.length }}</span> of <span class="text-gray-200 font-semibold">{{ deviceStore.summary.totalDevices }}</span>
        </span>
      </div>
    </div>

    <!-- Skeleton while loading -->
    <SkeletonTable v-if="deviceStore.isLoading" :rows="10" :cols="8" />

    <!-- Grouped View by Location (Drill-down Accordion Grid) -->
    <div v-else-if="viewMode === 'grouped'" class="space-y-4">
      <div v-if="groupedByLocation.length === 0" class="bg-[#151517] border border-[#26262A] rounded-xl p-12 text-center">
        <MapPin class="w-8 h-8 text-gray-600 mx-auto mb-2" />
        <p class="text-sm font-semibold text-gray-300">No location groups match your filters</p>
        <p class="text-xs text-gray-500 mt-1">Try clearing active search or device type filters</p>
      </div>

      <div
        v-for="group in groupedByLocation"
        :key="group.locationName"
        class="bg-[#151517] border rounded-xl overflow-hidden shadow-lg transition-all"
        :class="group.downCount > 0 ? 'border-red-500/30' : 'border-[#26262A]'"
      >
        <!-- Group Header Bar -->
        <div
          @click="toggleGroupExpand(group.locationName)"
          class="p-4 bg-[#18181B] border-b border-[#26262A] flex items-center justify-between cursor-pointer hover:bg-[#1E1E22] transition-colors"
        >
          <div class="flex items-center gap-3">
            <div
              class="w-8 h-8 rounded-lg flex items-center justify-center border"
              :class="group.downCount > 0 ? 'bg-red-500/10 border-red-500/30 text-red-400' : 'bg-[#7B96F5]/10 border-[#7B96F5]/30 text-[#7B96F5]'"
            >
              <MapPin class="w-4 h-4" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-white flex items-center gap-2">
                {{ group.locationName }}
              </h3>
              <p class="text-[11px] font-mono text-gray-400">
                {{ group.devices.length }} Nodes Monitored &bull; {{ group.upCount }} UP, {{ group.downCount }} DOWN
              </p>
            </div>
          </div>

          <div class="flex items-center gap-3">
            <!-- Status Pill Summary -->
            <span
              class="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold uppercase border"
              :class="group.downCount > 0 ? 'bg-red-500/15 text-red-400 border-red-500/30' : 'bg-[#3ECF8E]/15 text-[#3ECF8E] border-[#3ECF8E]/30'"
            >
              {{ group.downCount > 0 ? `${group.downCount} OUTAGE(S)` : 'OPERATIONAL' }}
            </span>

            <ChevronDown
              class="w-4 h-4 text-gray-400 transition-transform duration-200"
              :class="expandedGroups[group.locationName] !== false ? 'rotate-180' : ''"
            />
          </div>
        </div>

        <!-- Group Devices Table / Grid (Drill Down) -->
        <div v-if="expandedGroups[group.locationName] !== false" class="overflow-x-auto">
          <table class="w-full text-left text-xs text-gray-300">
            <thead class="bg-[#151517] font-mono text-[10px] uppercase text-gray-500 border-b border-[#26262A]">
              <tr>
                <th class="py-3 px-4">Device Name</th>
                <th class="py-3 px-4">Type</th>
                <th class="py-3 px-4">IP Address</th>
                <th class="py-3 px-4">MAC Address</th>
                <th class="py-3 px-4">Rack</th>
                <th class="py-3 px-4">Status</th>
                <th class="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[#26262A]">
              <tr
                v-for="device in group.devices"
                :key="device.id"
                class="hover:bg-[#18181B] transition-colors group cursor-pointer"
                :class="{ 'border-l-2 border-l-[#F16565] bg-[#F16565]/5': device.status === 'DOWN' }"
                @click="router.push(`/devices/${device.id}`)"
              >
                <td class="py-3 px-4 font-bold text-white group-hover:text-[#7B96F5]">
                  {{ device.name }}
                </td>
                <td class="py-3 px-4 font-mono text-gray-400">{{ device.type }}</td>
                <td class="py-3 px-4 font-mono text-gray-200">{{ device.ip }}</td>
                <td class="py-3 px-4 font-mono text-gray-500 text-[11px]">{{ device.mac }}</td>
                <td class="py-3 px-4 font-mono text-gray-400 text-[11px]">{{ device.rack || '—' }}</td>
                <td class="py-3 px-4"><StatusPill :status="device.status" /></td>
                <td class="py-3 px-4 text-right" @click.stop>
                  <div class="flex items-center justify-end gap-1">
                    <button
                      v-if="authStore.canEditDevice"
                      @click.stop="openEditDevice(device)"
                      class="p-1.5 rounded-lg text-gray-400 hover:text-[#7B96F5] hover:bg-[#7B96F5]/10"
                      title="Edit Device"
                    >
                      <Pencil class="w-3.5 h-3.5" />
                    </button>
                    <button
                      v-if="authStore.canDeleteDevice"
                      @click.stop="confirmDelete(device)"
                      class="p-1.5 rounded-lg text-gray-400 hover:text-[#F16565] hover:bg-[#F16565]/10"
                      title="Delete Device"
                    >
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                    <button
                      @click.stop="router.push(`/devices/${device.id}`)"
                      class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-[#26262A]"
                    >
                      <ChevronRight class="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Flat List Table View -->
    <div v-else class="bg-[#151517] border border-[#26262A] rounded-xl overflow-hidden shadow-xl">
      <table class="w-full text-left text-xs text-gray-300">
        <thead class="bg-[#18181B] border-b border-[#26262A] font-mono text-[10px] uppercase text-gray-400 sticky top-0">
          <tr>
            <th class="py-3.5 px-4">Device Name</th>
            <th class="py-3.5 px-4">Type</th>
            <th class="py-3.5 px-4">IP Address</th>
            <th class="py-3.5 px-4">MAC Address</th>
            <th class="py-3.5 px-4">Location</th>
            <th class="py-3.5 px-4">Addr Mode</th>
            <th class="py-3.5 px-4">Status</th>
            <th class="py-3.5 px-4 text-right">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#26262A]">
          <template v-if="deviceStore.filteredDevices.length > 0">
            <tr
              v-for="device in deviceStore.filteredDevices.slice(0, pageSize)"
              :key="device.id"
              class="hover:bg-[#18181B] transition-colors group cursor-pointer"
              :class="{
                'border-l-2 border-l-[#F16565] bg-[#F16565]/5': device.status === 'DOWN'
              }"
              @click="router.push(`/devices/${device.id}`)"
            >
              <td class="py-3 px-4">
                <span class="font-bold text-white group-hover:text-[#7B96F5] transition-colors">
                  {{ device.name }}
                </span>
              </td>
              <td class="py-3 px-4">
                <span class="font-mono text-gray-400">{{ device.type }}</span>
              </td>
              <td class="py-3 px-4">
                <span class="font-mono text-gray-200">{{ device.ip }}</span>
              </td>
              <td class="py-3 px-4">
                <span class="font-mono text-gray-500 text-[11px]">{{ device.mac }}</span>
              </td>
              <td class="py-3 px-4">
                <span class="text-gray-300 font-medium truncate max-w-[150px] inline-block">{{ device.location || 'Unassigned' }}</span>
              </td>
              <td class="py-3 px-4">
                <span
                  class="px-2 py-0.5 rounded-full text-[10px] font-mono font-semibold border"
                  :class="device.addressingMode === 'DHCP'
                    ? 'bg-amber-500/10 text-amber-400 border-amber-500/30'
                    : 'bg-[#7B96F5]/10 text-[#7B96F5] border-[#7B96F5]/30'"
                >
                  {{ device.addressingMode }}
                </span>
              </td>
              <td class="py-3 px-4">
                <StatusPill :status="device.status" />
              </td>
              <td class="py-3 px-4 text-right" @click.stop>
                <div class="flex items-center justify-end gap-1">
                  <!-- Edit: superadmin + anggota -->
                  <button
                    v-if="authStore.canEditDevice"
                    @click.stop="openEditDevice(device)"
                    class="p-1.5 rounded-lg text-gray-400 hover:text-[#7B96F5] hover:bg-[#7B96F5]/10 transition-colors"
                    title="Edit Device"
                  >
                    <Pencil class="w-3.5 h-3.5" />
                  </button>
                  <!-- Delete: superadmin only -->
                  <button
                    v-if="authStore.canDeleteDevice"
                    @click.stop="confirmDelete(device)"
                    class="p-1.5 rounded-lg text-gray-400 hover:text-[#F16565] hover:bg-[#F16565]/10 transition-colors"
                    title="Delete Device"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                  <button
                    @click.stop="router.push(`/devices/${device.id}`)"
                    class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-[#26262A] transition-colors"
                    title="View Device Detail"
                  >
                    <ChevronRight class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </template>
          <tr v-else>
            <td colspan="8" class="py-16 text-center">
              <div class="flex flex-col items-center gap-3">
                <Search class="w-8 h-8 text-gray-600" />
                <p class="text-gray-500 font-medium">No devices match your filters</p>
                <button @click="clearFilters" class="text-xs text-[#7B96F5] hover:underline">Clear filters</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination footer -->
      <div
        v-if="deviceStore.filteredDevices.length > pageSize"
        class="flex items-center justify-between px-4 py-3 border-t border-[#26262A] bg-[#18181B]"
      >
        <span class="text-xs text-gray-500 font-mono">
          Showing {{ pageSize }} of {{ deviceStore.filteredDevices.length }} matching devices
        </span>
        <button
          @click="pageSize = Math.min(pageSize + 50, deviceStore.filteredDevices.length)"
          class="text-xs text-[#7B96F5] hover:underline font-semibold"
        >
          Load 50 more
        </button>
      </div>
    </div>

    <!-- Delete Confirm Dialog -->
    <div
      v-if="deleteTarget"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm z-50 flex items-center justify-center p-4"
      @click.self="deleteTarget = null"
    >
      <div class="bg-[#151517] border border-[#F16565]/40 rounded-xl p-6 w-full max-w-sm space-y-4 shadow-2xl">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-full bg-[#F16565]/15 border border-[#F16565]/30 flex items-center justify-center">
            <Trash2 class="w-5 h-5 text-[#F16565]" />
          </div>
          <div>
            <p class="text-sm font-bold text-white">Delete Device?</p>
            <p class="text-xs text-gray-400 mt-0.5">This cannot be undone.</p>
          </div>
        </div>
        <p class="text-xs text-gray-300">
          You are about to remove
          <span class="text-white font-semibold">{{ deleteTarget?.name }}</span>
          from monitoring inventory. All associated status history will be permanently deleted.
        </p>
        <div class="flex gap-2 justify-end">
          <button
            @click="deleteTarget = null"
            class="px-4 py-2 rounded-lg border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-medium transition-colors"
          >
            Cancel
          </button>
          <button
            @click="executeDelete"
            class="px-4 py-2 rounded-lg bg-[#F16565] hover:bg-red-400 text-white text-xs font-semibold transition-colors flex items-center gap-1.5"
          >
            <Trash2 class="w-3.5 h-3.5" />
            Delete Device
          </button>
        </div>
      </div>
    </div>

    <!-- Device Form Modal -->
    <DeviceFormModal
      :is-open="isFormModalOpen"
      :mode="formModalMode"
      :device="editTarget"
      @close="isFormModalOpen = false"
      @saved="deviceStore.fetchDevices(); isFormModalOpen = false"
    />

    <!-- Bulk Import Modal -->
    <BulkImportModal :is-open="isImportModalOpen" @close="isImportModalOpen = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useDeviceStore } from '../stores/deviceStore';
import { useAuthStore } from '../stores/authStore';
import type { Device } from '../types';
import DeviceFormModal from '../components/devices/DeviceFormModal.vue';
import StatusPill from '../components/common/StatusPill.vue';
import SkeletonTable from '../components/common/SkeletonTable.vue';
import BulkImportModal from '../components/devices/BulkImportModal.vue';
import { Plus, Search, ChevronRight, ChevronDown, Upload, Pencil, Trash2, X, MapPin, List } from 'lucide-vue-next';

const router = useRouter();
const deviceStore = useDeviceStore();
const authStore = useAuthStore();

const viewMode = ref<'grouped' | 'flat'>('grouped');
const expandedGroups = reactive<Record<string, boolean>>({});
const isFormModalOpen = ref(false);
const formModalMode = ref<'add' | 'edit'>('add');
const isImportModalOpen = ref(false);
const editTarget = ref<Device | null>(null);
const deleteTarget = ref<Device | null>(null);
const pageSize = ref(50);

const hasActiveFilters = computed(() =>
  deviceStore.searchQuery !== '' ||
  deviceStore.selectedTypeFilter !== 'All' ||
  deviceStore.selectedStatusFilter !== 'All'
);

const groupedByLocation = computed(() => {
  const map: Record<string, { locationName: string; devices: Device[]; upCount: number; downCount: number }> = {};
  for (const d of deviceStore.filteredDevices) {
    const loc = (d.location || 'Unassigned').trim();
    if (!map[loc]) {
      map[loc] = { locationName: loc, devices: [], upCount: 0, downCount: 0 };
    }
    map[loc].devices.push(d);
    if (d.status === 'UP') map[loc].upCount++;
    else map[loc].downCount++;
  }
  return Object.values(map).sort((a, b) => b.downCount - a.downCount || a.locationName.localeCompare(b.locationName));
});

function toggleGroupExpand(locName: string) {
  if (expandedGroups[locName] === undefined) {
    expandedGroups[locName] = false;
  } else {
    expandedGroups[locName] = !expandedGroups[locName];
  }
}

function clearFilters() {
  deviceStore.searchQuery = '';
  deviceStore.selectedTypeFilter = 'All';
  deviceStore.selectedStatusFilter = 'All';
}

function openAddDevice() {
  editTarget.value = null;
  formModalMode.value = 'add';
  isFormModalOpen.value = true;
}

function openEditDevice(device: Device) {
  editTarget.value = device;
  formModalMode.value = 'edit';
  isFormModalOpen.value = true;
}

function confirmDelete(device: Device) {
  deleteTarget.value = device;
}

function executeDelete() {
  if (!deleteTarget.value) return;
  const idx = deviceStore.devices.findIndex((d: Device) => d.id === deleteTarget.value!.id);
  if (idx !== -1) {
    const removed = deviceStore.devices.splice(idx, 1)[0];
    deviceStore.summary.totalDevices--;
    if (removed.status === 'UP') deviceStore.summary.devicesUp--;
    else deviceStore.summary.devicesDown--;
  }
  deleteTarget.value = null;
}

onMounted(() => {
  deviceStore.fetchDevices();
});
</script>