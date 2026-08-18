<template>
  <div class="space-y-6">
    <!-- Header Title & Action -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-xl font-extrabold text-white tracking-tight">Device Management</h1>
        <p class="text-xs text-gray-400 mt-1">
          Active inventory and monitoring parameters for all {{ deviceStore.summary.totalDevices }} infrastructure nodes
        </p>
      </div>

      <!-- Action Buttons — role-gated via v-if -->
      <div class="flex items-center gap-2.5">
        <!-- Bulk Mode Toggle -->
        <button
          v-if="authStore.canEditDevice"
          @click="toggleBulkMode"
          class="px-3.5 py-2 rounded-lg border text-xs font-semibold transition-all flex items-center gap-2 cursor-pointer"
          :class="isBulkMode ? 'bg-[#7B96F5]/15 border-[#7B96F5]/40 text-[#7B96F5] shadow-sm shadow-[#7B96F5]/20' : 'border-[#26262A] bg-[#151517] hover:bg-[#1E1E22] text-gray-300 hover:text-white'"
        >
          <Sliders class="w-4 h-4" />
          <span>{{ isBulkMode ? 'Exit Bulk Mode' : 'Bulk Operations' }}</span>
        </button>

        <!-- Import: admin only -->
        <button
          v-if="authStore.canImportDevices"
          @click="isImportModalOpen = true"
          class="px-3.5 py-2 rounded-lg border border-[#26262A] bg-[#151517] hover:bg-[#1E1E22] text-gray-300 hover:text-white font-semibold text-xs transition-all flex items-center gap-2 cursor-pointer"
        >
          <Upload class="w-4 h-4" />
          Import CSV / Excel
        </button>
        <!-- Add: admin + anggota -->
        <button
          v-if="authStore.canAddDevice"
          @click="openAddDevice()"
          class="px-4 py-2 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 transition-all flex items-center gap-2 cursor-pointer"
        >
          <Plus class="w-4 h-4" />
          Add Device
        </button>
      </div>
    </div>

    <!-- Filter Bar & View Toggle -->
    <div class="bg-[#151517] border border-[#26262A] rounded-lg p-4 flex flex-wrap items-center justify-between gap-4">
      <div class="flex flex-wrap items-center gap-3 flex-1">
        <div class="relative flex-1 min-w-[240px]">
          <Search class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            v-model="deviceStore.searchQuery"
            type="text"
            placeholder="Search by name, IP, or MAC address..."
            class="w-full bg-[#18181B] border border-[#26262A] rounded-lg pl-9 pr-4 py-2 text-xs text-gray-200 focus:outline-none focus:border-[#7B96F5] placeholder-gray-600 font-mono"
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
          class="text-xs text-gray-400 hover:text-gray-200 flex items-center gap-1 transition-colors cursor-pointer"
        >
          <X class="w-3.5 h-3.5" />
          Clear Filters
        </button>
      </div>

      <!-- View Toggle Buttons (Grouped by Location vs Flat List) -->
      <div class="flex items-center gap-2">
        <div class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5">
          <button
            @click="viewMode = 'grouped'"
            class="px-3 py-1.5 rounded-md text-xs font-semibold flex items-center gap-1.5 transition-all cursor-pointer"
            :class="viewMode === 'grouped' ? 'bg-[#26262A] text-white shadow-sm' : 'text-gray-400 hover:text-gray-200'"
          >
            <MapPin class="w-3.5 h-3.5" />
            Group by Location
          </button>
          <button
            @click="viewMode = 'flat'"
            class="px-3 py-1.5 rounded-md text-xs font-semibold flex items-center gap-1.5 transition-all cursor-pointer"
            :class="viewMode === 'flat' ? 'bg-[#26262A] text-white shadow-sm' : 'text-gray-400 hover:text-gray-200'"
          >
            <List class="w-3.5 h-3.5" />
            Flat Table List
          </button>
        </div>

        <span class="text-xs font-mono text-gray-400 border-l border-[#26262A] pl-3 ml-1">
          Showing {{ deviceStore.filteredDevices.length }} of {{ deviceStore.summary.totalDevices }} devices
        </span>
      </div>
    </div>

    <!-- Skeleton while loading -->
    <template v-if="deviceStore.isLoading">
      <div v-if="viewMode === 'grouped'" class="space-y-4">
        <div v-for="g in 3" :key="g" class="bg-[#151517] border border-[#26262A] rounded-lg p-4 space-y-3">
          <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
            <Skeleton width="30%" height="1.1rem" />
            <Skeleton width="15%" height="0.8rem" />
          </div>
          <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
            <div v-for="i in 3" :key="i" class="p-3 bg-[#18181B] border border-[#26262A] rounded-md space-y-2">
              <Skeleton width="60%" height="0.8rem" />
              <Skeleton width="40%" height="0.65rem" />
            </div>
          </div>
        </div>
      </div>
      <SkeletonTable v-else :rows="10" :cols="isBulkMode ? 9 : 8" />
    </template>

    <!-- Grouped View by Location (Drill-down Accordion Grid) -->
    <div v-else-if="viewMode === 'grouped'" class="space-y-4">
      <div v-if="paginatedGroupedByLocation.length === 0" class="bg-[#151517] border border-[#26262A] rounded-lg p-12 text-center">
        <MapPin class="w-8 h-8 text-gray-600 mx-auto mb-2" />
        <p class="text-sm font-semibold text-gray-300">No location groups match your filters</p>
        <p class="text-xs text-gray-500 mt-1">Try clearing active search or device type filters</p>
      </div>

      <div
        v-for="group in paginatedGroupedByLocation"
        :key="group.locationName"
        class="bg-[#151517] border rounded-lg overflow-hidden shadow-lg transition-all"
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
                <span>{{ group.locationName }}</span>
                <div v-if="authStore.canEditDevice && group.locationName !== 'Unassigned'" class="flex items-center gap-1 ml-1" @click.stop>
                  <button
                    @click.stop="openEditLocationModal(group.locationName)"
                    class="p-1 rounded text-gray-400 hover:text-[#7B96F5] hover:bg-[#7B96F5]/10 transition-colors cursor-pointer"
                    title="Rename Location"
                  >
                    <Pencil class="w-3 h-3" />
                  </button>
                  <button
                    v-if="authStore.canDeleteDevice"
                    @click.stop="confirmDeleteLocationGroup(group.locationName, group.devices.length)"
                    class="p-1 rounded text-gray-400 hover:text-[#F16565] hover:bg-[#F16565]/10 transition-colors cursor-pointer"
                    title="Delete Location"
                  >
                    <Trash2 class="w-3 h-3" />
                  </button>
                </div>
              </h3>
              <p class="text-[11px] font-mono text-gray-400">
                {{ group.devices.length }} Nodes Monitored &bull; {{ group.upCount }} UP, {{ group.downCount }} DOWN
              </p>
            </div>
          </div>

          <div class="flex items-center gap-3" @click.stop>
            <!-- Select Group Checkbox (Visible only in Bulk Mode) -->
            <button
              v-if="isBulkMode"
              @click.stop="toggleSelectGroup(group.devices)"
              class="px-2 py-1 rounded border border-[#26262A] bg-[#151517] hover:border-[#7B96F5] text-[11px] font-mono text-gray-300 transition-colors flex items-center gap-1.5 cursor-pointer"
            >
              <input
                type="checkbox"
                :checked="isGroupSelected(group.devices)"
                class="rounded border-gray-700 bg-gray-900 text-[#7B96F5] focus:ring-0 cursor-pointer"
                @click.stop="toggleSelectGroup(group.devices)"
              />
              <span>Select Group</span>
            </button>

            <!-- Status Pill Summary -->
            <span
              class="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold uppercase border"
              :class="group.downCount > 0 ? 'bg-red-500/15 text-red-400 border-red-500/30' : 'bg-[#3ECF8E]/15 text-[#3ECF8E] border-[#3ECF8E]/30'"
            >
              {{ group.downCount > 0 ? `${group.downCount} OUTAGE(S)` : 'OPERATIONAL' }}
            </span>

            <ChevronDown
              @click="toggleGroupExpand(group.locationName)"
              class="w-4 h-4 text-gray-400 transition-transform duration-200 cursor-pointer"
              :class="expandedGroups[group.locationName] !== false ? 'rotate-180' : ''"
            />
          </div>
        </div>

        <!-- Group Devices Table / Grid (Drill Down) -->
        <div v-if="expandedGroups[group.locationName] !== false" class="overflow-x-auto">
          <table class="w-full text-left text-xs text-gray-300">
            <thead class="bg-[#151517] font-mono text-[10px] uppercase text-gray-500 border-b border-[#26262A]">
              <tr>
                <th v-if="isBulkMode" class="py-3 px-4 w-10">
                  <input
                    type="checkbox"
                    :checked="isGroupSelected(group.devices)"
                    @change="toggleSelectGroup(group.devices)"
                    class="rounded border-gray-700 bg-gray-900 text-[#7B96F5] focus:ring-0 cursor-pointer"
                  />
                </th>
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
                :class="{
                  'border-l-2 border-l-[#F16565] bg-[#F16565]/5': device.status === 'DOWN',
                  'bg-[#7B96F5]/5': selectedDeviceIds.includes(device.id)
                }"
                @click="router.push(`/devices/${device.id}`)"
              >
                <td v-if="isBulkMode" class="py-3 px-4" @click.stop>
                  <input
                    type="checkbox"
                    :value="device.id"
                    v-model="selectedDeviceIds"
                    class="rounded border-gray-700 bg-gray-900 text-[#7B96F5] focus:ring-0 cursor-pointer"
                  />
                </td>
                <td class="py-3 px-4 font-bold text-white group-hover:text-[#7B96F5]">
                  <div class="flex items-center gap-2">
                    <span>{{ device.name }}</span>
                    <span v-if="device.snmpEnabled" class="px-1.5 py-0.5 rounded text-[9px] font-mono font-bold bg-[#3ECF8E]/10 text-[#3ECF8E] border border-[#3ECF8E]/30 flex items-center gap-1" title="SNMP Polling Active">
                      <span class="w-1 h-1 rounded-full bg-[#3ECF8E] animate-ping"></span>
                      SNMP
                    </span>
                  </div>
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
                      class="p-1.5 rounded-lg text-gray-400 hover:text-[#7B96F5] hover:bg-[#7B96F5]/10 cursor-pointer"
                      title="Edit Device"
                    >
                      <Pencil class="w-3.5 h-3.5" />
                    </button>
                    <button
                      v-if="authStore.canDeleteDevice"
                      @click.stop="confirmDelete(device)"
                      class="p-1.5 rounded-lg text-gray-400 hover:text-[#F16565] hover:bg-[#F16565]/10 cursor-pointer"
                      title="Delete Device"
                    >
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                    <button
                      @click.stop="router.push(`/devices/${device.id}`)"
                      class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-[#26262A] cursor-pointer"
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
    <div v-else class="bg-[#151517] border border-[#26262A] rounded-lg overflow-hidden shadow-xl">
      <table class="w-full text-left text-xs text-gray-300">
        <thead class="bg-[#18181B] border-b border-[#26262A] font-mono text-[10px] uppercase text-gray-400 sticky top-0">
          <tr>
            <th v-if="isBulkMode" class="py-3.5 px-4 w-10">
              <input
                type="checkbox"
                :checked="isAllVisibleSelected"
                @change="toggleSelectAllVisible"
                class="rounded border-gray-700 bg-gray-900 text-[#7B96F5] focus:ring-0 cursor-pointer"
              />
            </th>
            <th class="py-3.5 px-4">Device Name</th>
            <th class="py-3.5 px-4">Type</th>
            <th class="py-3.5 px-4">IP Address</th>
            <th class="py-3.5 px-4">MAC Address</th>
            <th class="py-3.5 px-4">Location / Site</th>
            <th class="py-3.5 px-4">IP Mode</th>
            <th class="py-3.5 px-4">Status</th>
            <th class="py-3.5 px-4 text-right">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#26262A]">
          <template v-if="deviceStore.devices.length > 0">
            <tr
              v-for="device in deviceStore.devices"
              :key="device.id"
              class="hover:bg-[#18181B] transition-colors group cursor-pointer"
              :class="{
                'border-l-2 border-l-[#F16565] bg-[#F16565]/5': device.status === 'DOWN',
                'bg-[#7B96F5]/5': selectedDeviceIds.includes(device.id)
              }"
              @click="router.push(`/devices/${device.id}`)"
            >
              <td v-if="isBulkMode" class="py-3 px-4" @click.stop>
                <input
                  type="checkbox"
                  :value="device.id"
                  v-model="selectedDeviceIds"
                  class="rounded border-gray-700 bg-gray-900 text-[#7B96F5] focus:ring-0 cursor-pointer"
                />
              </td>
              <td class="py-3 px-4">
                <div class="flex items-center gap-2">
                  <span class="font-bold text-white group-hover:text-[#7B96F5] transition-colors">
                    {{ device.name }}
                  </span>
                  <span v-if="device.snmpEnabled" class="px-1.5 py-0.5 rounded text-[9px] font-mono font-bold bg-[#3ECF8E]/10 text-[#3ECF8E] border border-[#3ECF8E]/30 flex items-center gap-1" title="SNMP Polling Active">
                    <span class="w-1 h-1 rounded-full bg-[#3ECF8E] animate-ping"></span>
                    SNMP
                  </span>
                </div>
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
                  {{ device.addressingMode || 'STATIC' }}
                </span>
              </td>
              <td class="py-3 px-4">
                <StatusPill :status="device.status" />
              </td>
              <td class="py-3 px-4 text-right" @click.stop>
                <div class="flex items-center justify-end gap-1">
                  <button
                    v-if="authStore.canEditDevice"
                    @click.stop="openEditDevice(device)"
                    class="p-1.5 rounded-lg text-gray-400 hover:text-[#7B96F5] hover:bg-[#7B96F5]/10 cursor-pointer"
                    title="Edit Device"
                  >
                    <Pencil class="w-3.5 h-3.5" />
                  </button>
                  <button
                    v-if="authStore.canDeleteDevice"
                    @click.stop="confirmDelete(device)"
                    class="p-1.5 rounded-lg text-gray-400 hover:text-[#F16565] hover:bg-[#F16565]/10 cursor-pointer"
                    title="Delete Device"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                  <button
                    @click.stop="router.push(`/devices/${device.id}`)"
                    class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-[#26262A] cursor-pointer"
                  >
                    <ChevronRight class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </template>
          <tr v-else>
            <td :colspan="isBulkMode ? 9 : 8" class="py-12 text-center text-gray-500 font-mono text-xs">
              No devices matching filter criteria
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination Control -->
    <PaginationControl
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :total="paginatedTotal"
    />

    <!-- Floating Bulk Actions Bar (Visible when items selected in Bulk Mode) -->
    <div
      v-if="isBulkMode && selectedDeviceIds.length > 0"
      class="fixed bottom-6 left-1/2 -translate-x-1/2 bg-[#151517] border border-[#7B96F5]/50 shadow-2xl rounded-2xl px-5 py-3.5 flex items-center gap-4 z-40 animate-fadeIn"
    >
      <div class="flex items-center gap-2 border-r border-[#26262A] pr-4">
        <span class="w-6 h-6 rounded-full bg-[#7B96F5] text-white font-mono text-xs font-bold flex items-center justify-center shadow-sm shadow-[#7B96F5]/30">
          {{ selectedDeviceIds.length }}
        </span>
        <span class="text-xs font-bold text-white font-mono">Devices Selected</span>
      </div>

      <div class="flex items-center gap-2">
        <button
          @click="isBulkDrawerOpen = true"
          class="px-4 py-2 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-[#7B96F5]/25 transition-all cursor-pointer"
        >
          <Sliders class="w-3.5 h-3.5" />
          <span>Open Bulk Actions Drawer &rarr;</span>
        </button>
      </div>

      <button
        @click="selectedDeviceIds = []"
        class="text-gray-400 hover:text-white p-1 rounded-lg border-l border-[#26262A] pl-3 text-xs font-mono cursor-pointer"
        title="Deselect All"
      >
        Deselect All
      </button>
    </div>

    <!-- Slide-Over Right Drawer for Bulk Configuration -->
    <div v-if="isBulkDrawerOpen" class="fixed inset-0 z-50 overflow-hidden animate-fadeIn">
      <!-- Backdrop -->
      <div class="absolute inset-0 bg-black/70 backdrop-blur-xs transition-opacity" @click="isBulkDrawerOpen = false"></div>

      <div class="fixed inset-y-0 right-0 max-w-full flex pl-10">
        <div class="w-screen max-w-md bg-[#151517] border-l border-[#26262A] shadow-2xl p-6 flex flex-col justify-between overflow-y-auto">
          <!-- Drawer Header & Content -->
          <div class="space-y-5">
            <div class="flex items-center justify-between border-b border-[#26262A] pb-4">
              <div class="flex items-center gap-2.5">
                <div class="w-9 h-9 rounded-xl bg-[#7B96F5]/15 border border-[#7B96F5]/30 flex items-center justify-center text-[#7B96F5]">
                  <Sliders class="w-5 h-5" />
                </div>
                <div>
                  <h2 class="text-sm font-bold text-white font-mono">BULK CONFIGURATION</h2>
                  <p class="text-xs text-gray-400 font-mono">{{ selectedDeviceIds.length }} devices selected</p>
                </div>
              </div>
              <button @click="isBulkDrawerOpen = false" class="text-gray-400 hover:text-white p-1 rounded-lg cursor-pointer">
                <X class="w-5 h-5" />
              </button>
            </div>

            <!-- Selected Devices Chips Preview -->
            <div class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-[10px] font-mono uppercase text-gray-400 font-bold">Selected Devices:</span>
                <span class="text-[10px] font-mono text-[#7B96F5]">{{ selectedDevicesList.length }} items</span>
              </div>
              <div class="flex flex-wrap gap-1.5 max-h-28 overflow-y-auto p-2.5 bg-[#18181B] border border-[#26262A] rounded-xl">
                <span
                  v-for="d in selectedDevicesList.slice(0, 8)"
                  :key="d.id"
                  class="px-2 py-0.5 rounded-lg bg-[#26262A] border border-[#333338] text-[10px] font-mono text-gray-200"
                >
                  {{ d.name }}
                </span>
                <span v-if="selectedDevicesList.length > 8" class="px-2 py-0.5 rounded-lg bg-[#7B96F5]/20 text-[#7B96F5] text-[10px] font-mono font-bold">
                  +{{ selectedDevicesList.length - 8 }} more
                </span>
              </div>
            </div>

            <!-- Form Section 1: Pindah Lokasi Massal -->
            <div class="p-4 bg-[#18181B] border border-[#26262A] rounded-2xl space-y-3">
              <div class="flex items-center justify-between">
                <h3 class="text-xs font-bold text-white font-mono flex items-center gap-1.5">
                  <MapPin class="w-3.5 h-3.5 text-[#7B96F5]" />
                  Relocate Location / Site
                </h3>
              </div>
              <select
                v-model="bulkLocationValue"
                class="w-full bg-[#151517] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 text-xs font-mono focus:outline-none focus:border-[#7B96F5]"
              >
                <option value="" disabled>Select Target Location</option>
                <option v-for="loc in availableLocations" :key="loc" :value="loc">{{ loc }}</option>
              </select>
              <button
                @click="applyBulkLocation"
                :disabled="!bulkLocationValue || isExecutingBulk"
                class="w-full py-2 rounded-xl bg-[#7B96F5]/15 hover:bg-[#7B96F5]/25 border border-[#7B96F5]/30 text-[#7B96F5] text-xs font-bold font-mono transition-colors disabled:opacity-40 flex items-center justify-center gap-1.5 cursor-pointer"
              >
                <span>{{ isExecutingBulk && bulkActionType === 'location' ? 'Saving...' : 'Apply Location Changes' }}</span>
              </button>
            </div>

            <!-- Form Section 2: Ubah Kategori / Tipe Perangkat -->
            <div class="p-4 bg-[#18181B] border border-[#26262A] rounded-2xl space-y-3">
              <div class="flex items-center justify-between">
                <h3 class="text-xs font-bold text-white font-mono flex items-center gap-1.5">
                  <Sliders class="w-3.5 h-3.5 text-[#7B96F5]" />
                  Change Device Type / Category
                </h3>
              </div>
              <select
                v-model="bulkTypeValue"
                class="w-full bg-[#151517] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 text-xs font-mono focus:outline-none focus:border-[#7B96F5]"
              >
                <option value="Access Point">Access Point</option>
                <option value="Switch">Switch</option>
                <option value="Router">Router</option>
                <option value="SmartPower">SmartPower</option>
                <option value="CCTV">CCTV</option>
                <option value="NVR">NVR</option>
              </select>
              <button
                @click="applyBulkType"
                :disabled="isExecutingBulk"
                class="w-full py-2 rounded-xl bg-[#7B96F5]/15 hover:bg-[#7B96F5]/25 border border-[#7B96F5]/30 text-[#7B96F5] text-xs font-bold font-mono transition-colors disabled:opacity-40 flex items-center justify-center gap-1.5 cursor-pointer"
              >
                <span>{{ isExecutingBulk && bulkActionType === 'type' ? 'Saving...' : 'Apply New Type' }}</span>
              </button>
            </div>

            <!-- Form Section 3: Force Immediate ICMP Poll -->
            <div class="p-4 bg-[#18181B] border border-[#26262A] rounded-2xl space-y-3">
              <div class="flex items-center justify-between">
                <h3 class="text-xs font-bold text-white font-mono flex items-center gap-1.5">
                  <Activity class="w-3.5 h-3.5 text-[#3ECF8E]" />
                  Trigger Instant Polling
                </h3>
              </div>
              <p class="text-[11px] text-gray-400 font-mono">
                Send instant ICMP ping probes to {{ selectedDeviceIds.length }} selected devices.
              </p>
              <button
                @click="triggerBulkPoll"
                :disabled="isExecutingBulk"
                class="w-full py-2 rounded-xl bg-[#3ECF8E]/15 hover:bg-[#3ECF8E]/25 border border-[#3ECF8E]/30 text-[#3ECF8E] text-xs font-bold font-mono transition-colors disabled:opacity-40 flex items-center justify-center gap-1.5 cursor-pointer"
              >
                <RefreshCw class="w-3.5 h-3.5" :class="isExecutingBulk && bulkActionType === 'poll' ? 'animate-spin' : ''" />
                <span>{{ isExecutingBulk && bulkActionType === 'poll' ? 'Pinging...' : 'Poll Selected Devices' }}</span>
              </button>
            </div>

            <!-- Form Section 4: Hapus Massal -->
            <div v-if="authStore.canDeleteDevice" class="p-4 bg-red-500/10 border border-red-500/30 rounded-2xl space-y-3">
              <div class="flex items-center justify-between text-red-400">
                <h3 class="text-xs font-bold font-mono flex items-center gap-1.5">
                  <Trash2 class="w-3.5 h-3.5" />
                  Bulk Delete Devices
                </h3>
              </div>
              <p class="text-[11px] text-gray-300 font-mono">
                Permanently remove {{ selectedDeviceIds.length }} devices from system inventory.
              </p>
              <button
                @click="confirmBulkDelete"
                :disabled="isExecutingBulk"
                class="w-full py-2 rounded-xl bg-red-500 hover:bg-red-600 text-white text-xs font-bold font-mono shadow-md shadow-red-500/20 transition-colors disabled:opacity-40 flex items-center justify-center gap-1.5 cursor-pointer"
              >
                <Trash2 class="w-3.5 h-3.5" />
                <span>{{ isExecutingBulk && bulkActionType === 'delete' ? 'Deleting...' : `Delete ${selectedDeviceIds.length} Devices` }}</span>
              </button>
            </div>
          </div>

          <!-- Drawer Footer -->
          <div class="pt-5 border-t border-[#26262A] flex items-center justify-between mt-6">
            <button
              @click="selectedDeviceIds = []"
              class="px-4 py-2 rounded-xl bg-[#18181B] border border-[#26262A] hover:bg-[#26262A] text-gray-300 text-xs font-mono cursor-pointer"
            >
              Deselect All
            </button>
            <button
              @click="isBulkDrawerOpen = false"
              class="px-5 py-2 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-bold font-mono shadow-md shadow-[#7B96F5]/20 cursor-pointer"
            >
              Close Panel
            </button>
          </div>
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

    <!-- Single Device Delete Modal -->
    <Modal :is-open="deleteTarget !== null" title="Confirm Delete Device" @close="deleteTarget = null">
      <template #default>
        <div class="space-y-4 text-xs">
          <div class="p-3 bg-red-500/10 border border-red-500/30 rounded-xl flex items-start gap-3 text-red-400">
            <AlertTriangle class="w-5 h-5 shrink-0 mt-0.5" />
            <div>
              <h4 class="font-bold font-mono">Tindakan Ini Tidak Dapat Dibatalkan</h4>
              <p class="mt-1 text-[11px] text-gray-300">
                Apakah Anda yakin ingin menghapus perangkat <strong class="text-white font-mono">{{ deleteTarget?.name }}</strong> (IP: {{ deleteTarget?.ip }})?
              </p>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <button
          @click="deleteTarget = null"
          class="px-4 py-2 rounded-xl border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-mono cursor-pointer"
        >
          Cancel
        </button>
        <button
          @click="executeDelete"
          class="px-5 py-2 rounded-xl bg-red-500 hover:bg-red-600 text-white font-semibold text-xs shadow-md shadow-red-500/20 font-mono flex items-center gap-1.5 cursor-pointer"
        >
          <Trash2 class="w-3.5 h-3.5" />
          <span>Delete Device</span>
        </button>
      </template>
    </Modal>

    <!-- Location Edit / Rename Modal in DevicesView -->
    <Modal :is-open="isLocEditModalOpen" title="Rename Location" @close="isLocEditModalOpen = false">
      <template #default>
        <div class="space-y-4 text-xs">
          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-gray-400">Location Name *</label>
            <input
              v-model="locEditName"
              type="text"
              required
              class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
            />
          </div>
          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-gray-400">Description</label>
            <input
              v-model="locEditDescription"
              type="text"
              placeholder="e.g. Server Room Main Switch Rack"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
        </div>
      </template>
      <template #footer>
        <button
          @click="isLocEditModalOpen = false"
          class="px-4 py-2 rounded-xl border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-mono cursor-pointer"
        >
          Cancel
        </button>
        <button
          @click="handleSaveLocEdit"
          :disabled="isSavingLocEdit"
          class="px-5 py-2 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 font-mono disabled:opacity-50 cursor-pointer"
        >
          {{ isSavingLocEdit ? 'Saving...' : 'Rename Location' }}
        </button>
      </template>
    </Modal>

    <!-- Location Delete Modal in DevicesView -->
    <Modal :is-open="isLocDeleteModalOpen" title="Confirm Delete Location" @close="isLocDeleteModalOpen = false">
      <template #default>
        <div class="space-y-4 text-xs">
          <div v-if="locDeleteCount > 0" class="p-3 bg-amber-500/10 border border-amber-500/30 rounded-xl text-amber-300 space-y-2">
            <div class="flex items-center gap-2 font-bold font-mono">
              <AlertTriangle class="w-4 h-4 text-amber-400 shrink-0" />
              <span>Deletion Blocked</span>
            </div>
            <p class="text-[11px] font-mono">
              <strong>{{ locDeleteCount }} devices</strong> are currently assigned to location <strong class="text-white font-semibold">{{ locDeleteName }}</strong>. Please reassign or move those devices before deleting this location.
            </p>
          </div>
          <div v-else class="p-3 bg-red-500/10 border border-red-500/30 rounded-xl flex items-start gap-3 text-red-400">
            <AlertTriangle class="w-5 h-5 shrink-0 mt-0.5" />
            <div>
              <h4 class="font-bold font-mono">Confirm Deletion</h4>
              <p class="mt-1 text-[11px] text-gray-300">
                Are you sure you want to delete empty location <strong class="text-white font-mono">{{ locDeleteName }}</strong>?
              </p>
            </div>
          </div>
        </div>
      </template>
      <template #footer>
        <button
          @click="isLocDeleteModalOpen = false"
          class="px-4 py-2 rounded-xl border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-mono cursor-pointer"
        >
          Cancel
        </button>
        <button
          v-if="locDeleteCount === 0"
          @click="handleConfirmLocDelete"
          :disabled="isDeletingLoc"
          class="px-5 py-2 rounded-xl bg-red-500 hover:bg-red-600 text-white font-semibold text-xs shadow-md font-mono disabled:opacity-50 flex items-center gap-1.5 cursor-pointer"
        >
          <Trash2 class="w-3.5 h-3.5" />
          <span>{{ isDeletingLoc ? 'Deleting...' : 'Delete Location' }}</span>
        </button>
      </template>
    </Modal>

    <!-- Bulk Import Modal -->
    <BulkImportModal :is-open="isImportModalOpen" @close="isImportModalOpen = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useDeviceStore } from '../stores/deviceStore';
import { useAuthStore } from '../stores/authStore';
import type { Device, DeviceType, LocationItem } from '../types';
import { devicesApi, locationsApi } from '../api';
import api from '../api/client';
import DeviceFormModal from '../components/devices/DeviceFormModal.vue';
import Modal from '../components/common/Modal.vue';
import StatusPill from '../components/common/StatusPill.vue';
import SkeletonTable from '../components/common/SkeletonTable.vue';
import Skeleton from '../components/common/Skeleton.vue';
import BulkImportModal from '../components/devices/BulkImportModal.vue';
import PaginationControl from '../components/common/PaginationControl.vue';
import {
  Plus,
  Search,
  ChevronRight,
  ChevronDown,
  Upload,
  Pencil,
  Trash2,
  X,
  MapPin,
  List,
  Activity,
  Sliders,
  AlertTriangle,
  RefreshCw
} from 'lucide-vue-next';

const router = useRouter();
const route = useRoute();
const deviceStore = useDeviceStore();
const authStore = useAuthStore();

const viewMode = ref<'grouped' | 'flat'>('flat');
const expandedGroups = reactive<Record<string, boolean>>({});
const isFormModalOpen = ref(false);
const formModalMode = ref<'add' | 'edit'>('add');
const isImportModalOpen = ref(false);
const editTarget = ref<Device | null>(null);
const deleteTarget = ref<Device | null>(null);
const currentPage = ref(1);
const pageSize = ref(10);

// Bulk Mode State
const isBulkMode = ref(false);
const isBulkDrawerOpen = ref(false);
const selectedDeviceIds = ref<string[]>([]);
const isExecutingBulk = ref(false);
const bulkActionType = ref<'location' | 'type' | 'poll' | 'delete' | ''>('');
const bulkLocationValue = ref('');
const bulkTypeValue = ref<DeviceType>('Access Point');

function toggleBulkMode() {
  isBulkMode.value = !isBulkMode.value;
  if (!isBulkMode.value) {
    selectedDeviceIds.value = [];
    isBulkDrawerOpen.value = false;
  }
}

const selectedDevicesList = computed(() => {
  return deviceStore.devices.filter((d) => selectedDeviceIds.value.includes(d.id));
});

// Location Management in Group Header State
const isLocEditModalOpen = ref(false);
const locEditId = ref('');
const locEditName = ref('');
const locEditDescription = ref('');
const isSavingLocEdit = ref(false);

const isLocDeleteModalOpen = ref(false);
const locDeleteId = ref('');
const locDeleteName = ref('');
const locDeleteCount = ref(0);
const isDeletingLoc = ref(false);

const allLocations = ref<LocationItem[]>([]);

async function fetchAllLocations() {
  try {
    const locs = await locationsApi.getLocations();
    if (locs) allLocations.value = locs;
  } catch (e) {}
}

async function openEditLocationModal(locName: string) {
  await fetchAllLocations();
  const match = allLocations.value.find((l) => l.name.toLowerCase() === locName.toLowerCase());
  locEditId.value = match?.id || '';
  locEditName.value = locName;
  locEditDescription.value = match?.description || '';
  isLocEditModalOpen.value = true;
}

async function handleSaveLocEdit() {
  if (!locEditName.value.trim()) return;
  isSavingLocEdit.value = true;
  try {
    if (locEditId.value) {
      await locationsApi.updateLocation(locEditId.value, locEditName.value, locEditDescription.value);
    } else {
      await locationsApi.createLocation(locEditName.value, locEditDescription.value);
    }
    isLocEditModalOpen.value = false;
    await fetchAllLocations();
    loadDevices();
  } catch (e: any) {
    alert(e.response?.data?.error || 'Failed to rename location');
  } finally {
    isSavingLocEdit.value = false;
  }
}

async function confirmDeleteLocationGroup(locName: string, count: number) {
  await fetchAllLocations();
  const match = allLocations.value.find((l) => l.name.toLowerCase() === locName.toLowerCase());
  locDeleteId.value = match?.id || '';
  locDeleteName.value = locName;
  locDeleteCount.value = count;
  isLocDeleteModalOpen.value = true;
}

async function handleConfirmLocDelete() {
  if (!locDeleteId.value && !locDeleteName.value) return;
  isDeletingLoc.value = true;
  try {
    if (locDeleteId.value) {
      await locationsApi.deleteLocation(locDeleteId.value);
    }
    isLocDeleteModalOpen.value = false;
    await fetchAllLocations();
    loadDevices();
  } catch (e: any) {
    alert(e.response?.data?.error || 'Failed to delete location');
  } finally {
    isDeletingLoc.value = false;
  }
}

const availableLocations = computed(() => {
  const set = new Set<string>();
  for (const loc of allLocations.value) {
    if (loc.name) set.add(loc.name);
  }
  for (const d of deviceStore.devices) {
    if (d.location) set.add(d.location);
  }
  return Array.from(set);
});

function isGroupSelected(groupDevices: Device[]) {
  if (!groupDevices || groupDevices.length === 0) return false;
  return groupDevices.every((d) => selectedDeviceIds.value.includes(d.id));
}

function toggleSelectGroup(groupDevices: Device[]) {
  if (isGroupSelected(groupDevices)) {
    const ids = groupDevices.map((d) => d.id);
    selectedDeviceIds.value = selectedDeviceIds.value.filter((id) => !ids.includes(id));
  } else {
    for (const d of groupDevices) {
      if (!selectedDeviceIds.value.includes(d.id)) {
        selectedDeviceIds.value.push(d.id);
      }
    }
  }
}

const visibleDevices = computed(() => {
  if (viewMode.value === 'flat') {
    return deviceStore.devices;
  }
  return deviceStore.filteredDevices;
});

const isAllVisibleSelected = computed(() => {
  if (visibleDevices.value.length === 0) return false;
  return visibleDevices.value.every((d) => selectedDeviceIds.value.includes(d.id));
});

function toggleSelectAllVisible() {
  if (isAllVisibleSelected.value) {
    const ids = visibleDevices.value.map((d) => d.id);
    selectedDeviceIds.value = selectedDeviceIds.value.filter((id) => !ids.includes(id));
  } else {
    for (const d of visibleDevices.value) {
      if (!selectedDeviceIds.value.includes(d.id)) {
        selectedDeviceIds.value.push(d.id);
      }
    }
  }
}

// Bulk Actions Handlers
async function applyBulkLocation() {
  if (!bulkLocationValue.value || selectedDeviceIds.value.length === 0) return;
  isExecutingBulk.value = true;
  bulkActionType.value = 'location';
  try {
    await devicesApi.bulkAction({
      action: 'update',
      deviceIds: selectedDeviceIds.value,
      updates: { location: bulkLocationValue.value }
    });
    alert(`Lokasi untuk ${selectedDeviceIds.value.length} perangkat berhasil diubah ke "${bulkLocationValue.value}".`);
    selectedDeviceIds.value = [];
    isBulkDrawerOpen.value = false;
    loadDevices();
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal memindahkan lokasi perangkat');
  } finally {
    isExecutingBulk.value = false;
    bulkActionType.value = '';
  }
}

async function applyBulkType() {
  if (!bulkTypeValue.value || selectedDeviceIds.value.length === 0) return;
  isExecutingBulk.value = true;
  bulkActionType.value = 'type';
  try {
    await devicesApi.bulkAction({
      action: 'update',
      deviceIds: selectedDeviceIds.value,
      updates: { type: bulkTypeValue.value }
    });
    alert(`Tipe untuk ${selectedDeviceIds.value.length} perangkat berhasil diubah ke "${bulkTypeValue.value}".`);
    selectedDeviceIds.value = [];
    isBulkDrawerOpen.value = false;
    loadDevices();
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal mengubah tipe perangkat');
  } finally {
    isExecutingBulk.value = false;
    bulkActionType.value = '';
  }
}

async function triggerBulkPoll() {
  if (selectedDeviceIds.value.length === 0) return;
  isExecutingBulk.value = true;
  bulkActionType.value = 'poll';
  try {
    await api.post('/monitoring/refresh-now');
    alert(`ICMP Probe scan berhasil dikirimkan ke ${selectedDeviceIds.value.length} perangkat.`);
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal melakukan polling');
  } finally {
    setTimeout(() => {
      isExecutingBulk.value = false;
      bulkActionType.value = '';
    }, 600);
  }
}

async function confirmBulkDelete() {
  if (selectedDeviceIds.value.length === 0) return;
  if (!confirm(`Apakah Anda yakin ingin menghapus permanen ${selectedDeviceIds.value.length} perangkat terpilih?`)) {
    return;
  }
  isExecutingBulk.value = true;
  bulkActionType.value = 'delete';
  try {
    await devicesApi.bulkAction({
      action: 'delete',
      deviceIds: selectedDeviceIds.value
    });
    alert(`${selectedDeviceIds.value.length} perangkat berhasil dihapus dari inventaris.`);
    selectedDeviceIds.value = [];
    isBulkDrawerOpen.value = false;
    loadDevices();
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal menghapus perangkat');
  } finally {
    isExecutingBulk.value = false;
    bulkActionType.value = '';
  }
}

const paginatedTotal = computed(() => {
  if (viewMode.value === 'flat') {
    return deviceStore.totalCount;
  }
  return groupedByLocation.value.length;
});

const paginatedGroupedByLocation = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  return groupedByLocation.value.slice(start, start + pageSize.value);
});

const hasActiveFilters = computed(
  () =>
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

async function executeDelete() {
  if (!deleteTarget.value) return;
  try {
    await devicesApi.deleteDevice(deleteTarget.value.id);
    deleteTarget.value = null;
    loadDevices();
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal menghapus perangkat');
  }
}

function loadDevices() {
  if (viewMode.value === 'flat') {
    deviceStore.fetchDevices({
      page: currentPage.value,
      pageSize: pageSize.value
    });
  } else {
    deviceStore.fetchDevices();
  }
}

onMounted(() => {
  fetchAllLocations();
  if (route.query.status) {
    deviceStore.selectedStatusFilter = route.query.status as string;
  }
  if (route.query.type) {
    deviceStore.selectedTypeFilter = route.query.type as string;
  }
  loadDevices();
});

watch([currentPage, pageSize], () => {
  loadDevices();
});

watch(
  () => route.query,
  (newQ) => {
    if (newQ.status) {
      deviceStore.selectedStatusFilter = newQ.status as string;
    }
    if (newQ.type) {
      deviceStore.selectedTypeFilter = newQ.type as string;
    }
  }
);

watch(
  () => [
    viewMode.value,
    deviceStore.searchQuery,
    deviceStore.selectedTypeFilter,
    deviceStore.selectedStatusFilter
  ],
  () => {
    currentPage.value = 1;
    loadDevices();
  }
);
</script>