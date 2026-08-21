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
          v-if="authStore.canBulkManageDevices"
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
          <template v-if="paginatedFlatDevices.length > 0">
            <tr
              v-for="device in paginatedFlatDevices"
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
                    @click.stop="openDiagnosticsForDevice(device)"
                    class="p-1.5 rounded-lg text-gray-400 hover:text-[#3ECF8E] hover:bg-[#3ECF8E]/10 cursor-pointer transition-colors"
                    title="Run Ping / Diagnostics"
                  >
                    <Terminal class="w-3.5 h-3.5" />
                  </button>
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
      class="fixed bottom-6 left-1/2 -translate-x-1/2 bg-[#151517] border border-[#7B96F5]/50 shadow-2xl rounded-2xl px-5 py-3.5 flex items-center gap-3.5 z-40 animate-fadeIn"
    >
      <div class="flex items-center gap-2 border-r border-[#26262A] pr-3.5">
        <span class="w-6 h-6 rounded-full bg-[#7B96F5] text-white font-mono text-xs font-bold flex items-center justify-center shadow-sm shadow-[#7B96F5]/30">
          {{ selectedDeviceIds.length }}
        </span>
        <span class="text-xs font-bold text-white font-mono hidden sm:inline">Perangkat Dipilih</span>
      </div>

      <div class="flex items-center gap-2">
        <button
          type="button"
          @click="isBulkDrawerOpen = true"
          class="px-4 py-2 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-[#7B96F5]/25 transition-all cursor-pointer font-mono"
        >
          <Sliders class="w-3.5 h-3.5" />
          <span>Kelola Massal (Bulk Edit) &rarr;</span>
        </button>

        <button
          type="button"
          @click="triggerBulkPoll"
          :disabled="isExecutingBulk"
          class="px-3 py-2 rounded-xl bg-[#3ECF8E]/15 hover:bg-[#3ECF8E]/25 border border-[#3ECF8E]/30 text-[#3ECF8E] text-xs font-bold flex items-center gap-1.5 transition-all disabled:opacity-40 cursor-pointer font-mono"
          title="Instant ICMP Ping"
        >
          <RefreshCw class="w-3.5 h-3.5" :class="isExecutingBulk && bulkActionType === 'poll' ? 'animate-spin' : ''" />
          <span class="hidden md:inline">Ping Massal</span>
        </button>

        <button
          v-if="authStore.canDeleteDevice"
          type="button"
          @click="isBulkDeleteConfirmModalOpen = true"
          :disabled="isExecutingBulk"
          class="px-3 py-2 rounded-xl bg-red-500/15 hover:bg-red-500/25 border border-red-500/30 text-red-400 text-xs font-bold flex items-center gap-1.5 transition-all disabled:opacity-40 cursor-pointer font-mono"
          title="Hapus Massal"
        >
          <Trash2 class="w-3.5 h-3.5" />
          <span class="hidden md:inline">Hapus</span>
        </button>
      </div>

      <button
        type="button"
        @click="selectedDeviceIds = []"
        class="text-gray-400 hover:text-white p-1 rounded-lg border-l border-[#26262A] pl-3 text-xs font-mono cursor-pointer"
        title="Deselect All"
      >
        Batal Pilih
      </button>
    </div>

    <!-- Slide-Over Right Drawer for Bulk Configuration -->
    <div v-if="isBulkDrawerOpen" class="fixed inset-0 z-50 overflow-hidden animate-fadeIn">
      <!-- Backdrop -->
      <div class="absolute inset-0 bg-black/75 backdrop-blur-xs transition-opacity" @click="handleCloseBulkDrawer"></div>

      <div class="fixed inset-y-0 right-0 max-w-full flex pl-10">
        <div class="w-screen max-w-lg bg-[#151517] border-l border-[#26262A] shadow-2xl p-6 flex flex-col justify-between overflow-y-auto">
          <!-- Drawer Header & Content -->
          <div class="space-y-5">
            <!-- Header -->
            <div class="flex items-center justify-between border-b border-[#26262A] pb-4">
              <div class="flex items-center gap-2.5">
                <div class="w-9 h-9 rounded-xl bg-[#7B96F5]/15 border border-[#7B96F5]/30 flex items-center justify-center text-[#7B96F5]">
                  <Sliders class="w-5 h-5" />
                </div>
                <div>
                  <h2 class="text-sm font-bold text-white font-mono uppercase tracking-wide">BULK CONFIGURATION</h2>
                  <p class="text-xs text-gray-400 font-mono">{{ selectedDeviceIds.length }} perangkat dipilih untuk diubah massal</p>
                </div>
              </div>
              <button @click="handleCloseBulkDrawer" class="text-gray-400 hover:text-white p-1.5 rounded-lg hover:bg-[#26262A] cursor-pointer transition-colors">
                <X class="w-5 h-5" />
              </button>
            </div>

            <!-- Selected Devices Chips Preview with Quick Deselect Chip -->
            <div class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-[10px] font-mono uppercase text-gray-400 font-bold">Daftar Perangkat Terpilih:</span>
                <span class="text-[10px] font-mono text-[#7B96F5]">{{ selectedDevicesList.length }} items</span>
              </div>
              <div class="flex flex-wrap gap-1.5 max-h-24 overflow-y-auto p-2.5 bg-[#18181B] border border-[#26262A] rounded-xl">
                <span
                  v-for="d in selectedDevicesList.slice(0, 10)"
                  :key="d.id"
                  class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-lg bg-[#26262A] border border-[#333338] text-[10px] font-mono text-gray-200 group"
                >
                  <span class="truncate max-w-[120px]">{{ d.name }}</span>
                  <button
                    type="button"
                    @click="removeSelectedDevice(d.id)"
                    class="text-gray-400 hover:text-red-400 cursor-pointer opacity-70 group-hover:opacity-100"
                    title="Keluarkan dari batch"
                  >
                    <X class="w-3 h-3" />
                  </button>
                </span>
                <span v-if="selectedDevicesList.length > 10" class="px-2 py-0.5 rounded-lg bg-[#7B96F5]/20 text-[#7B96F5] text-[10px] font-mono font-bold">
                  +{{ selectedDevicesList.length - 10 }} lainnya
                </span>
              </div>
            </div>

            <!-- Validation Error Banner -->
            <div v-if="bulkFormError" class="p-3 bg-red-500/10 border border-red-500/30 rounded-xl flex items-start gap-2.5 text-red-400 text-xs font-mono">
              <AlertTriangle class="w-4 h-4 shrink-0 mt-0.5" />
              <span>{{ bulkFormError }}</span>
            </div>

            <!-- SECTION 1: Kategori & Mode Pengalamatan -->
            <div class="p-4 bg-[#18181B] border border-[#26262A] rounded-2xl space-y-3.5 shadow-sm">
              <div class="flex items-center justify-between">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="bulkForm.enableType"
                    class="w-4 h-4 rounded bg-[#151517] border-[#333338] text-[#7B96F5] accent-[#7B96F5] cursor-pointer"
                  />
                  <span class="text-xs font-bold font-mono text-white flex items-center gap-1.5">
                    <Layers class="w-3.5 h-3.5 text-[#7B96F5]" />
                    Ubah Tipe / Kategori
                  </span>
                </label>
                <span v-if="bulkForm.enableType" class="text-[9px] font-mono uppercase bg-[#7B96F5]/20 text-[#7B96F5] px-1.5 py-0.5 rounded font-bold">Aktif</span>
              </div>

              <div v-if="bulkForm.enableType" class="pl-6 space-y-2 pt-1 border-t border-[#26262A]/60">
                <select
                  v-model="bulkForm.type"
                  class="w-full bg-[#151517] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 text-xs font-mono focus:outline-none focus:border-[#7B96F5]"
                >
                  <option value="Access Point">Access Point</option>
                  <option value="Switch">Switch</option>
                  <option value="Router">Router</option>
                  <option value="SmartPower">SmartPower</option>
                  <option value="CCTV">CCTV</option>
                  <option value="NVR">NVR</option>
                </select>
              </div>

              <div class="border-t border-[#26262A]/60 pt-3">
                <div class="flex items-center justify-between">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      v-model="bulkForm.enableAddressingMode"
                      class="w-4 h-4 rounded bg-[#151517] border-[#333338] text-[#7B96F5] accent-[#7B96F5] cursor-pointer"
                    />
                    <span class="text-xs font-bold font-mono text-white flex items-center gap-1.5">
                      <Radio class="w-3.5 h-3.5 text-[#7B96F5]" />
                      Ubah Mode Pengalamatan
                    </span>
                  </label>
                  <span v-if="bulkForm.enableAddressingMode" class="text-[9px] font-mono uppercase bg-[#7B96F5]/20 text-[#7B96F5] px-1.5 py-0.5 rounded font-bold">Aktif</span>
                </div>

                <div v-if="bulkForm.enableAddressingMode" class="pl-6 pt-2 flex items-center gap-6 text-xs text-gray-300">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input type="radio" v-model="bulkForm.addressingMode" value="Static" class="accent-[#7B96F5]" />
                    <span>Static IP</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input type="radio" v-model="bulkForm.addressingMode" value="DHCP" class="accent-[#7B96F5]" />
                    <span>DHCP Reservation</span>
                  </label>
                </div>
              </div>
            </div>

            <!-- SECTION 2: Penempatan Lokasi & Rak -->
            <div class="p-4 bg-[#18181B] border border-[#26262A] rounded-2xl space-y-3.5 shadow-sm">
              <div class="flex items-center justify-between">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="bulkForm.enableLocation"
                    class="w-4 h-4 rounded bg-[#151517] border-[#333338] text-[#7B96F5] accent-[#7B96F5] cursor-pointer"
                  />
                  <span class="text-xs font-bold font-mono text-white flex items-center gap-1.5">
                    <MapPin class="w-3.5 h-3.5 text-[#7B96F5]" />
                    Relokasi Lokasi / Site
                  </span>
                </label>
                <span v-if="bulkForm.enableLocation" class="text-[9px] font-mono uppercase bg-[#7B96F5]/20 text-[#7B96F5] px-1.5 py-0.5 rounded font-bold">Aktif</span>
              </div>

              <div v-if="bulkForm.enableLocation" class="pl-6 space-y-2 pt-1 border-t border-[#26262A]/60">
                <LocationCombobox
                  v-model="bulkForm.location"
                  v-model:locationId="bulkForm.locationId"
                />
              </div>

              <div class="border-t border-[#26262A]/60 pt-3">
                <div class="flex items-center justify-between">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      v-model="bulkForm.enableRack"
                      class="w-4 h-4 rounded bg-[#151517] border-[#333338] text-[#7B96F5] accent-[#7B96F5] cursor-pointer"
                    />
                    <span class="text-xs font-bold font-mono text-white flex items-center gap-1.5">
                      <Server class="w-3.5 h-3.5 text-[#7B96F5]" />
                      Ubah Posisi Rak (Rack)
                    </span>
                  </label>
                  <span v-if="bulkForm.enableRack" class="text-[9px] font-mono uppercase bg-[#7B96F5]/20 text-[#7B96F5] px-1.5 py-0.5 rounded font-bold">Aktif</span>
                </div>

                <div v-if="bulkForm.enableRack" class="pl-6 pt-2">
                  <input
                    v-model="bulkForm.rack"
                    type="text"
                    placeholder="e.g. Rack B-04 (atau kosongkan jika tanpa rak)"
                    class="w-full bg-[#151517] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 text-xs font-mono focus:outline-none focus:border-[#7B96F5]"
                  />
                </div>
              </div>
            </div>

            <!-- SECTION 3: Konfigurasi SNMP Polling -->
            <div class="p-4 bg-[#18181B] border border-[#26262A] rounded-2xl space-y-3.5 shadow-sm">
              <div class="flex items-center justify-between">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="bulkForm.enableSNMP"
                    class="w-4 h-4 rounded bg-[#151517] border-[#333338] text-[#7B96F5] accent-[#7B96F5] cursor-pointer"
                  />
                  <span class="text-xs font-bold font-mono text-white flex items-center gap-1.5">
                    <Cpu class="w-3.5 h-3.5 text-[#7B96F5]" />
                    Konfigurasi SNMP Polling
                  </span>
                </label>
                <span v-if="bulkForm.enableSNMP" class="text-[9px] font-mono uppercase bg-[#7B96F5]/20 text-[#7B96F5] px-1.5 py-0.5 rounded font-bold">Aktif</span>
              </div>

              <div v-if="bulkForm.enableSNMP" class="pl-6 space-y-3 pt-1 border-t border-[#26262A]/60">
                <div class="flex items-center justify-between pt-1">
                  <span class="text-xs text-gray-300 font-mono">SNMP Query Status</span>
                  <button
                    type="button"
                    @click="bulkForm.snmpEnabled = !bulkForm.snmpEnabled"
                    class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                    :class="bulkForm.snmpEnabled ? 'bg-[#7B96F5]' : 'bg-[#26262A]'"
                  >
                    <span
                      class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                      :class="bulkForm.snmpEnabled ? 'translate-x-4' : 'translate-x-0'"
                    />
                  </button>
                </div>

                <div v-if="bulkForm.snmpEnabled" class="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
                  <div class="space-y-1">
                    <label class="block font-mono uppercase text-[10px] text-gray-400">Community String</label>
                    <input
                      v-model="bulkForm.snmpCommunity"
                      type="text"
                      placeholder="public"
                      class="w-full bg-[#151517] border border-[#26262A] rounded-xl px-3 py-1.5 text-xs text-white font-mono focus:outline-none focus:border-[#7B96F5]"
                    />
                  </div>
                  <div class="space-y-1">
                    <label class="block font-mono uppercase text-[10px] text-gray-400">Port (Default 161)</label>
                    <input
                      v-model.number="bulkForm.snmpPort"
                      type="number"
                      min="1"
                      max="65535"
                      class="w-full bg-[#151517] border border-[#26262A] rounded-xl px-3 py-1.5 text-xs text-white font-mono focus:outline-none focus:border-[#7B96F5]"
                    />
                  </div>
                </div>
              </div>
            </div>

            <!-- SECTION 4: Toleransi Alarm / Failure Threshold Override -->
            <div class="p-4 bg-[#18181B] border border-[#26262A] rounded-2xl space-y-3.5 shadow-sm">
              <div class="flex items-center justify-between">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="bulkForm.enableThreshold"
                    class="w-4 h-4 rounded bg-[#151517] border-[#333338] text-[#7B96F5] accent-[#7B96F5] cursor-pointer"
                  />
                  <span class="text-xs font-bold font-mono text-white flex items-center gap-1.5">
                    <Sliders class="w-3.5 h-3.5 text-[#7B96F5]" />
                    Ambang Batas Alarm (Failure Threshold)
                  </span>
                </label>
                <span v-if="bulkForm.enableThreshold" class="text-[9px] font-mono uppercase bg-[#7B96F5]/20 text-[#7B96F5] px-1.5 py-0.5 rounded font-bold">Aktif</span>
              </div>

              <div v-if="bulkForm.enableThreshold" class="pl-6 space-y-3 pt-1 border-t border-[#26262A]/60">
                <div class="flex items-center justify-between pt-1">
                  <span class="text-xs text-gray-300 font-mono">Custom Threshold Override</span>
                  <button
                    type="button"
                    @click="bulkForm.useCustomThreshold = !bulkForm.useCustomThreshold"
                    class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                    :class="bulkForm.useCustomThreshold ? 'bg-[#7B96F5]' : 'bg-[#26262A]'"
                  >
                    <span
                      class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                      :class="bulkForm.useCustomThreshold ? 'translate-x-4' : 'translate-x-0'"
                    />
                  </button>
                </div>

                <div v-if="bulkForm.useCustomThreshold" class="space-y-1 pt-1">
                  <label class="block font-mono uppercase text-[10px] text-gray-400">Jumlah Kegagalan Probe Sebelum Alert (1–10)</label>
                  <div class="flex items-center gap-3">
                    <input
                      v-model.number="bulkForm.failureThreshold"
                      type="number"
                      min="1"
                      max="10"
                      class="w-24 bg-[#151517] border border-[#26262A] rounded-xl px-3 py-1.5 text-xs text-white font-mono text-center focus:outline-none focus:border-[#7B96F5]"
                    />
                    <span class="text-[11px] font-mono text-gray-400">kali gagal berturut-turut</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- SECTION 5: Quick Actions (Instant Poll & Bulk Delete) -->
            <div class="p-4 bg-[#18181B] border border-[#26262A] rounded-2xl space-y-3 shadow-sm">
              <h3 class="text-xs font-bold text-gray-400 font-mono uppercase tracking-wider">Aksi Cepat Lainnya</h3>
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
                <button
                  type="button"
                  @click="triggerBulkPoll"
                  :disabled="isExecutingBulk"
                  class="w-full py-2.5 px-3 rounded-xl bg-[#3ECF8E]/10 hover:bg-[#3ECF8E]/20 border border-[#3ECF8E]/30 text-[#3ECF8E] text-xs font-bold font-mono transition-all disabled:opacity-40 flex items-center justify-center gap-1.5 cursor-pointer"
                >
                  <RefreshCw class="w-3.5 h-3.5" :class="isExecutingBulk && bulkActionType === 'poll' ? 'animate-spin' : ''" />
                  <span>{{ isExecutingBulk && bulkActionType === 'poll' ? 'Pinging...' : 'Instant Ping All' }}</span>
                </button>

                <button
                  v-if="authStore.canDeleteDevice"
                  type="button"
                  @click="isBulkDeleteConfirmModalOpen = true"
                  :disabled="isExecutingBulk"
                  class="w-full py-2.5 px-3 rounded-xl bg-red-500/10 hover:bg-red-500/20 border border-red-500/30 text-red-400 text-xs font-bold font-mono transition-all disabled:opacity-40 flex items-center justify-center gap-1.5 cursor-pointer"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                  <span>Hapus Massal</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Sticky Drawer Footer -->
          <div class="pt-5 border-t border-[#26262A] bg-[#151517] sticky bottom-0 space-y-3 mt-6">
            <div class="flex items-center justify-between text-[11px] font-mono text-gray-400">
              <span>{{ activeBulkFieldCount }} properti aktif</span>
              <span class="text-[#7B96F5] font-bold">{{ selectedDeviceIds.length }} perangkat terpilih</span>
            </div>

            <div class="flex items-center gap-2.5">
              <button
                type="button"
                @click="resetBulkForm"
                class="px-4 py-2.5 rounded-xl bg-[#18181B] border border-[#26262A] hover:bg-[#26262A] text-gray-300 text-xs font-mono transition-colors cursor-pointer"
              >
                Reset
              </button>
              <button
                type="button"
                @click="handleAttemptApplyBulk"
                :disabled="activeBulkFieldCount === 0 || isExecutingBulk"
                class="flex-1 py-2.5 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-bold font-mono shadow-md shadow-[#7B96F5]/20 transition-all disabled:opacity-40 flex items-center justify-center gap-1.5 cursor-pointer"
              >
                <Check class="w-4 h-4" />
                <span>{{ isExecutingBulk && bulkActionType === 'update' ? 'Menerapkan...' : 'Terapkan Perubahan Massal' }}</span>
              </button>
            </div>
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
      @saved="loadDevices(); isFormModalOpen = false"
    />

    <!-- Diagnostic Terminal Modal -->
    <DiagnosticTerminalModal
      :is-open="isTerminalModalOpen"
      :initial-target="selectedDiagnosticTarget"
      @close="isTerminalModalOpen = false"
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

    <!-- Bulk Update Confirmation Modal -->
    <Modal :is-open="isBulkConfirmModalOpen" title="Konfirmasi Perubahan Massal" @close="isBulkConfirmModalOpen = false">
      <template #default>
        <div class="space-y-3.5 text-xs font-mono">
          <div class="p-4 bg-[#7B96F5]/10 border border-[#7B96F5]/30 rounded-2xl text-gray-200 space-y-2.5">
            <h4 class="font-bold text-white text-xs flex items-center gap-1.5">
              <Info class="w-4 h-4 text-[#7B96F5]" />
              Ringkasan Perubahan Konfigurasi
            </h4>
            <p class="text-[11px] text-gray-300">
              Anda akan menerapkan konfigurasi berikut secara massal ke <strong class="text-white font-bold">{{ selectedDeviceIds.length }} perangkat terpilih</strong>:
            </p>
            <ul class="list-disc list-inside space-y-1.5 text-[11px] text-gray-200 pt-1 border-t border-[#26262A]/80">
              <li v-if="bulkForm.enableType">
                Tipe / Kategori &rarr; <span class="text-[#7B96F5] font-bold">{{ bulkForm.type }}</span>
              </li>
              <li v-if="bulkForm.enableAddressingMode">
                Mode Pengalamatan &rarr; <span class="text-[#7B96F5] font-bold">{{ bulkForm.addressingMode }}</span>
              </li>
              <li v-if="bulkForm.enableLocation">
                Relokasi Lokasi &rarr; <span class="text-[#7B96F5] font-bold">{{ bulkForm.location }}</span>
              </li>
              <li v-if="bulkForm.enableRack">
                Posisi Rak (Rack) &rarr; <span class="text-[#7B96F5] font-bold">{{ bulkForm.rack || '(Dikosongkan / Reset)' }}</span>
              </li>
              <li v-if="bulkForm.enableSNMP">
                SNMP Polling &rarr; <span class="text-[#7B96F5] font-bold">{{ bulkForm.snmpEnabled ? `Aktif (Community: ${bulkForm.snmpCommunity}, Port: ${bulkForm.snmpPort})` : 'Nonaktif' }}</span>
              </li>
              <li v-if="bulkForm.enableThreshold">
                Threshold Override &rarr; <span class="text-[#7B96F5] font-bold">{{ bulkForm.useCustomThreshold ? `Aktif (${bulkForm.failureThreshold} checks)` : 'Nonaktif (Gunakan Standar Sistem)' }}</span>
              </li>
            </ul>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex items-center justify-end gap-3 w-full">
          <button
            type="button"
            @click="isBulkConfirmModalOpen = false"
            class="px-4 py-2 rounded-xl border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-mono cursor-pointer"
          >
            Batal
          </button>
          <button
            type="button"
            @click="executeBulkUpdate"
            :disabled="isExecutingBulk"
            class="px-5 py-2 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 font-mono disabled:opacity-50 flex items-center gap-1.5 cursor-pointer"
          >
            <Check class="w-3.5 h-3.5" />
            <span>{{ isExecutingBulk ? 'Menyimpan...' : 'Ya, Terapkan Sekarang' }}</span>
          </button>
        </div>
      </template>
    </Modal>

    <!-- Bulk Delete Confirmation Modal -->
    <Modal :is-open="isBulkDeleteConfirmModalOpen" title="Konfirmasi Hapus Massal" @close="isBulkDeleteConfirmModalOpen = false">
      <template #default>
        <div class="space-y-4 text-xs font-mono">
          <div class="p-4 bg-red-500/10 border border-red-500/30 rounded-2xl flex items-start gap-3 text-red-400">
            <AlertTriangle class="w-5 h-5 shrink-0 mt-0.5" />
            <div class="space-y-1.5">
              <h4 class="font-bold text-white text-xs">Tindakan Bersifat Permanen</h4>
              <p class="text-[11px] text-gray-300 leading-relaxed">
                Apakah Anda yakin ingin menghapus permanen <strong class="text-white font-bold">{{ selectedDeviceIds.length }} perangkat terpilih</strong> dari sistem inventaris? Seluruh riwayat probe, uptime, dan log status perangkat tersebut akan dihapus.
              </p>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex items-center justify-end gap-3 w-full">
          <button
            type="button"
            @click="isBulkDeleteConfirmModalOpen = false"
            class="px-4 py-2 rounded-xl border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-mono cursor-pointer"
          >
            Batal
          </button>
          <button
            type="button"
            @click="executeBulkDelete"
            :disabled="isExecutingBulk"
            class="px-5 py-2 rounded-xl bg-red-500 hover:bg-red-600 text-white font-semibold text-xs shadow-md shadow-red-500/20 font-mono disabled:opacity-50 flex items-center gap-1.5 cursor-pointer"
          >
            <Trash2 class="w-3.5 h-3.5" />
            <span>{{ isExecutingBulk ? 'Menghapus...' : `Hapus ${selectedDeviceIds.length} Perangkat` }}</span>
          </button>
        </div>
      </template>
    </Modal>

    <!-- Bulk Drawer Discard Unsaved Changes Modal -->
    <Modal :is-open="isBulkDiscardModalOpen" title="Batalkan Perubahan Bulk" @close="isBulkDiscardModalOpen = false">
      <template #default>
        <div class="p-4 bg-amber-500/10 border border-amber-500/30 rounded-2xl flex items-start gap-3 text-amber-400 text-xs font-mono">
          <AlertTriangle class="w-5 h-5 shrink-0 mt-0.5" />
          <div class="space-y-1">
            <h4 class="font-bold text-white text-xs">Perubahan Belum Diterapkan</h4>
            <p class="text-gray-300 text-[11px] leading-relaxed">
              Anda telah mengaktifkan konfigurasi pada panel Kelola Massal namun belum menyimpannya. Apakah Anda ingin membatalkan perubahan dan menutup panel?
            </p>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex items-center justify-end gap-3 w-full">
          <button
            type="button"
            @click="isBulkDiscardModalOpen = false"
            class="px-4 py-2 rounded-xl border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-mono cursor-pointer"
          >
            Tetap Edit
          </button>
          <button
            type="button"
            @click="confirmDiscardAndCloseBulkDrawer"
            class="px-5 py-2 rounded-xl bg-red-500 hover:bg-red-600 text-white font-semibold text-xs shadow-md shadow-red-500/20 font-mono cursor-pointer"
          >
            Batalkan &amp; Tutup
          </button>
        </div>
      </template>
    </Modal>

    <!-- Unified Feedback Modal -->
    <Modal :is-open="showFeedbackModal" :title="feedbackTitle" @close="showFeedbackModal = false">
      <template #default>
        <div class="space-y-3 text-xs font-mono">
          <div
            class="p-4 rounded-2xl border flex items-start gap-3"
            :class="feedbackSuccess ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300' : 'bg-red-500/10 border-red-500/30 text-red-300'"
          >
            <CheckCircle2 v-if="feedbackSuccess" class="w-5 h-5 shrink-0 text-emerald-400 mt-0.5" />
            <AlertTriangle v-else class="w-5 h-5 shrink-0 text-red-400 mt-0.5" />
            <div class="space-y-1">
              <h4 class="font-bold text-white text-sm">{{ feedbackTitle }}</h4>
              <p class="text-xs opacity-90 leading-relaxed">{{ feedbackMessage }}</p>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <button
          @click="showFeedbackModal = false"
          class="px-5 py-2 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 cursor-pointer font-mono"
        >
          OK
        </button>
      </template>
    </Modal>

    <!-- Bulk Import Modal -->
    <BulkImportModal :is-open="isImportModalOpen" @close="isImportModalOpen = false; loadDevices(); fetchAllLocations();" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useDeviceStore } from '../stores/deviceStore';
import { useAuthStore } from '../stores/authStore';
import type { Device, DeviceType, LocationItem, BulkDeviceUpdates } from '../types';
import { devicesApi, locationsApi } from '../api';
import api from '../api/client';
import DeviceFormModal from '../components/devices/DeviceFormModal.vue';
import LocationCombobox from '../components/devices/LocationCombobox.vue';
import Modal from '../components/common/Modal.vue';
import StatusPill from '../components/common/StatusPill.vue';
import SkeletonTable from '../components/common/SkeletonTable.vue';
import Skeleton from '../components/common/Skeleton.vue';
import BulkImportModal from '../components/devices/BulkImportModal.vue';
import PaginationControl from '../components/common/PaginationControl.vue';
import DiagnosticTerminalModal from '../components/diagnostics/DiagnosticTerminalModal.vue';
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
  Sliders,
  AlertTriangle,
  RefreshCw,
  Terminal,
  Check,
  CheckCircle2,
  Info,
  Layers,
  Radio,
  Server,
  Cpu
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
const isTerminalModalOpen = ref(false);
const selectedDiagnosticTarget = ref('');
const editTarget = ref<Device | null>(null);
const deleteTarget = ref<Device | null>(null);
const currentPage = ref(1);
const pageSize = ref(10);

function openDiagnosticsForDevice(device: Device) {
  selectedDiagnosticTarget.value = device.ip || device.mac || '';
  isTerminalModalOpen.value = true;
}

// Bulk Mode & Comprehensive Form State
const isBulkMode = ref(false);
const isBulkDrawerOpen = ref(false);
const selectedDeviceIds = ref<string[]>([]);
const isExecutingBulk = ref(false);
const bulkActionType = ref<'update' | 'poll' | 'delete' | ''>('');
const bulkFormError = ref('');

const isBulkConfirmModalOpen = ref(false);
const isBulkDeleteConfirmModalOpen = ref(false);
const isBulkDiscardModalOpen = ref(false);

// Unified Feedback Modal State
const showFeedbackModal = ref(false);
const feedbackTitle = ref('');
const feedbackMessage = ref('');
const feedbackSuccess = ref(true);

function triggerFeedback(title: string, message: string, success = true) {
  feedbackTitle.value = title;
  feedbackMessage.value = message;
  feedbackSuccess.value = success;
  showFeedbackModal.value = true;
}

const bulkForm = reactive({
  enableType: false,
  type: 'Access Point' as DeviceType,

  enableAddressingMode: false,
  addressingMode: 'Static' as 'Static' | 'DHCP',

  enableLocation: false,
  location: '',
  locationId: '',

  enableRack: false,
  rack: '',

  enableSNMP: false,
  snmpEnabled: false,
  snmpCommunity: 'public',
  snmpPort: 161,

  enableThreshold: false,
  useCustomThreshold: false,
  failureThreshold: 3
});

const activeBulkFieldCount = computed(() => {
  let count = 0;
  if (bulkForm.enableType) count++;
  if (bulkForm.enableAddressingMode) count++;
  if (bulkForm.enableLocation) count++;
  if (bulkForm.enableRack) count++;
  if (bulkForm.enableSNMP) count++;
  if (bulkForm.enableThreshold) count++;
  return count;
});

function isBulkFormDirty(): boolean {
  return activeBulkFieldCount.value > 0;
}

function resetBulkForm() {
  bulkForm.enableType = false;
  bulkForm.type = 'Access Point';
  bulkForm.enableAddressingMode = false;
  bulkForm.addressingMode = 'Static';
  bulkForm.enableLocation = false;
  bulkForm.location = '';
  bulkForm.locationId = '';
  bulkForm.enableRack = false;
  bulkForm.rack = '';
  bulkForm.enableSNMP = false;
  bulkForm.snmpEnabled = false;
  bulkForm.snmpCommunity = 'public';
  bulkForm.snmpPort = 161;
  bulkForm.enableThreshold = false;
  bulkForm.useCustomThreshold = false;
  bulkForm.failureThreshold = 3;
  bulkFormError.value = '';
}

function handleCloseBulkDrawer() {
  if (isBulkFormDirty()) {
    isBulkDiscardModalOpen.value = true;
  } else {
    isBulkDrawerOpen.value = false;
  }
}

function confirmDiscardAndCloseBulkDrawer() {
  resetBulkForm();
  isBulkDiscardModalOpen.value = false;
  isBulkDrawerOpen.value = false;
}

function removeSelectedDevice(id: string) {
  selectedDeviceIds.value = selectedDeviceIds.value.filter((dId) => dId !== id);
  if (selectedDeviceIds.value.length === 0) {
    isBulkDrawerOpen.value = false;
    resetBulkForm();
  }
}

function toggleBulkMode() {
  if (!authStore.canBulkManageDevices) {
    isBulkMode.value = false;
    isBulkDrawerOpen.value = false;
    selectedDeviceIds.value = [];
    resetBulkForm();
    return;
  }
  isBulkMode.value = !isBulkMode.value;
  if (!isBulkMode.value) {
    selectedDeviceIds.value = [];
    isBulkDrawerOpen.value = false;
    resetBulkForm();
  }
}

watch(() => authStore.canBulkManageDevices, (can) => {
  if (!can && isBulkMode.value) {
    isBulkMode.value = false;
    isBulkDrawerOpen.value = false;
    selectedDeviceIds.value = [];
    resetBulkForm();
  }
});

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

const paginatedFlatDevices = computed(() => {
  if (deviceStore.devices.length <= pageSize.value) {
    return deviceStore.devices;
  }
  const start = (currentPage.value - 1) * pageSize.value;
  return deviceStore.devices.slice(start, start + pageSize.value);
});

const visibleDevices = computed(() => {
  if (viewMode.value === 'flat') {
    return paginatedFlatDevices.value;
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

// ─── Bulk Actions Handlers & Validations ──────────────────────────────────────────

function validateBulkForm(): boolean {
  bulkFormError.value = '';
  if (activeBulkFieldCount.value === 0) {
    bulkFormError.value = 'Silakan aktifkan minimal satu opsi konfigurasi untuk diterapkan.';
    return false;
  }
  if (bulkForm.enableLocation && !bulkForm.location.trim()) {
    bulkFormError.value = 'Lokasi target harus dipilih jika opsi Relokasi Lokasi diaktifkan.';
    return false;
  }
  if (bulkForm.enableSNMP && bulkForm.snmpEnabled) {
    if (!bulkForm.snmpPort || bulkForm.snmpPort < 1 || bulkForm.snmpPort > 65535) {
      bulkFormError.value = 'Port SNMP harus berupa angka antara 1 dan 65535.';
      return false;
    }
  }
  if (bulkForm.enableThreshold && bulkForm.useCustomThreshold) {
    if (!bulkForm.failureThreshold || bulkForm.failureThreshold < 1 || bulkForm.failureThreshold > 10) {
      bulkFormError.value = 'Ambang batas kegagalan probe harus berupa angka antara 1 dan 10.';
      return false;
    }
  }
  return true;
}

function handleAttemptApplyBulk() {
  if (!validateBulkForm()) return;
  isBulkConfirmModalOpen.value = true;
}

async function executeBulkUpdate() {
  if (selectedDeviceIds.value.length === 0 || !validateBulkForm()) return;
  isExecutingBulk.value = true;
  bulkActionType.value = 'update';

  const updates: BulkDeviceUpdates = {};
  if (bulkForm.enableType) updates.type = bulkForm.type;
  if (bulkForm.enableAddressingMode) updates.addressingMode = bulkForm.addressingMode;
  if (bulkForm.enableLocation) {
    updates.location = bulkForm.location.trim();
    if (bulkForm.locationId) updates.locationId = bulkForm.locationId;
  }
  if (bulkForm.enableRack) {
    updates.rack = bulkForm.rack.trim();
  }
  if (bulkForm.enableSNMP) {
    updates.snmpEnabled = bulkForm.snmpEnabled;
    if (bulkForm.snmpEnabled) {
      updates.snmpCommunity = bulkForm.snmpCommunity.trim();
      updates.snmpPort = Number(bulkForm.snmpPort) || 161;
    }
  }
  if (bulkForm.enableThreshold) {
    updates.useCustomThreshold = bulkForm.useCustomThreshold;
    if (bulkForm.useCustomThreshold) {
      updates.customFailureThreshold = Number(bulkForm.failureThreshold) || 3;
      updates.failureThreshold = Number(bulkForm.failureThreshold) || 3;
    }
  }

  try {
    const res = await devicesApi.bulkAction({
      action: 'update',
      deviceIds: selectedDeviceIds.value,
      updates
    });
    isBulkConfirmModalOpen.value = false;
    isBulkDrawerOpen.value = false;
    triggerFeedback(
      'Perubahan Massal Berhasil',
      `Berhasil memperbarui konfigurasi untuk ${res.updatedCount || selectedDeviceIds.value.length} perangkat terpilih.`,
      true
    );
    selectedDeviceIds.value = [];
    resetBulkForm();
    await loadDevices();
    await fetchAllLocations();
  } catch (e: any) {
    triggerFeedback(
      'Gagal Menyimpan Perubahan Massal',
      e.response?.data?.error || 'Terjadi kesalahan sistem saat memperbarui data perangkat.',
      false
    );
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
    triggerFeedback(
      'ICMP Probe Terkirim',
      `Permintaan probe instan berhasil dikirimkan ke ${selectedDeviceIds.value.length} perangkat terpilih. Data status akan terupdate secara real-time.`,
      true
    );
  } catch (e: any) {
    triggerFeedback(
      'Gagal Melakukan Polling',
      e.response?.data?.error || 'Gagal mengirim sinyal polling ke daemon monitoring.',
      false
    );
  } finally {
    setTimeout(() => {
      isExecutingBulk.value = false;
      bulkActionType.value = '';
    }, 500);
  }
}

async function executeBulkDelete() {
  if (selectedDeviceIds.value.length === 0) return;
  isExecutingBulk.value = true;
  bulkActionType.value = 'delete';
  const targetCount = selectedDeviceIds.value.length;

  try {
    await devicesApi.bulkAction({
      action: 'delete',
      deviceIds: selectedDeviceIds.value
    });
    isBulkDeleteConfirmModalOpen.value = false;
    isBulkDrawerOpen.value = false;
    triggerFeedback(
      'Perangkat Berhasil Dihapus',
      `Sebanyak ${targetCount} perangkat telah dihapus permanen dari sistem inventaris.`,
      true
    );
    selectedDeviceIds.value = [];
    resetBulkForm();
    await loadDevices();
    await fetchAllLocations();
  } catch (e: any) {
    triggerFeedback(
      'Gagal Menghapus Perangkat',
      e.response?.data?.error || 'Terjadi kesalahan saat menghapus data perangkat massal.',
      false
    );
  } finally {
    isExecutingBulk.value = false;
    bulkActionType.value = '';
  }
}

const paginatedTotal = computed(() => {
  if (viewMode.value === 'flat') {
    return deviceStore.totalCount || deviceStore.devices.length;
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
    await loadDevices();
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal menghapus perangkat');
  }
}

async function loadDevices() {
  if (viewMode.value === 'flat') {
    await deviceStore.fetchDevices({
      page: currentPage.value,
      pageSize: pageSize.value
    });
  } else {
    await deviceStore.fetchDevices();
  }
}

onMounted(() => {
  fetchAllLocations();
  deviceStore.selectedStatusFilter = (route.query.status as string) || 'All';
  deviceStore.selectedTypeFilter = (route.query.type as string) || 'All';
  if (route.query.search) {
    deviceStore.searchQuery = route.query.search as string;
  }
  loadDevices();
});

watch([currentPage, pageSize], () => {
  loadDevices();
});

watch(
  () => route.query,
  (newQ) => {
    deviceStore.selectedStatusFilter = (newQ.status as string) || 'All';
    deviceStore.selectedTypeFilter = (newQ.type as string) || 'All';
    if (newQ.search !== undefined) {
      deviceStore.searchQuery = (newQ.search as string) || '';
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