<template>
  <div class="space-y-6 max-w-6xl">
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-[#26262A] pb-4">
      <div>
        <h1 class="text-xl font-extrabold text-white tracking-tight">System Configuration</h1>
        <p class="text-xs text-gray-400 mt-1">Configure polling engine, alert gateways, failure thresholds, and NOC access control</p>
      </div>

      <button
        @click="handleSaveConfig"
        class="px-4 py-2 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 transition-all flex items-center gap-2"
      >
        <Save class="w-4 h-4" />
        {{ isSaveSuccess ? 'Configuration Saved!' : 'Save Changes' }}
      </button>
    </div>

    <!-- Section 1: Notification Channels -->
    <div class="space-y-3">
      <h2 class="text-xs font-bold uppercase tracking-wider text-gray-400 font-mono flex items-center gap-2">
        <Send class="w-4 h-4 text-[#7B96F5]" />
        Notification Channels
      </h2>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <!-- WhatsApp API Card -->
        <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4 shadow-xl">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
                <MessageSquare class="w-5 h-5" />
              </div>
              <div>
                <h3 class="text-sm font-bold text-white">WhatsApp API Gateway</h3>
                <p class="text-xs font-mono text-gray-400 mt-0.5">{{ whatsAppNumber || 'Gateway Configured' }}</p>
              </div>
            </div>
            <span
              class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold border"
              :class="isWhatsAppConnected ? 'bg-[#3ECF8E]/15 text-[#3ECF8E] border-[#3ECF8E]/30' : 'bg-amber-500/15 text-amber-400 border-amber-500/30'"
            >
              <span class="w-1.5 h-1.5 rounded-full" :class="isWhatsAppConnected ? 'bg-[#3ECF8E] pulsing-dot-green' : 'bg-amber-400'"></span>
              {{ isWhatsAppConnected ? 'CONNECTED' : 'DISCONNECTED' }}
            </span>
          </div>

          <div v-if="waTestMessage" class="p-2.5 rounded bg-[#18181B] border border-[#26262A] text-xs font-mono" :class="waTestSuccess ? 'text-[#34D399]' : 'text-[#F16565]'">
            {{ waTestMessage }}
          </div>

          <div class="flex items-center justify-between pt-3 border-t border-[#26262A] text-xs">
            <button
              @click="handleWATest"
              :disabled="!isWhatsAppConnected || isWATesting"
              class="h-9 px-3.5 rounded-lg border border-[#26262A] bg-[#18181B] hover:bg-[#26262A] text-gray-200 text-xs font-semibold flex items-center gap-2 transition-all disabled:opacity-40"
            >
              <Send class="w-4 h-4 text-[#7B96F5]" />
              <span>{{ isWATesting ? 'Sending...' : 'Send Test Notification' }}</span>
            </button>
            <div class="flex items-center gap-2.5">
              <button
                v-if="isWhatsAppConnected"
                @click="handleWADisconnect"
                class="h-9 px-3.5 rounded-lg border border-red-500/30 bg-red-500/10 hover:bg-red-500/20 text-red-400 text-xs font-semibold flex items-center gap-2 transition-all"
                title="Disconnect WhatsApp Gateway"
              >
                <LogOut class="w-4 h-4 text-red-400" />
                <span>Disconnect</span>
              </button>
              <button
                @click="isWATargetModalOpen = true"
                class="h-9 px-3.5 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-semibold flex items-center gap-2 shadow-sm shadow-[#7B96F5]/20 transition-all"
              >
                <Target class="w-4 h-4" />
                <span>Configure Targets</span>
              </button>
              <button
                @click="isWAModalOpen = true"
                class="h-9 px-3.5 rounded-lg bg-[#18181B] border border-[#26262A] hover:bg-[#26262A] text-gray-200 text-xs font-semibold flex items-center gap-2 transition-all"
              >
                <QrCode class="w-4 h-4 text-gray-400" />
                <span>QR Reconnect</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Telegram Bot Card -->
        <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4 shadow-xl">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-sky-500/15 border border-sky-500/30 flex items-center justify-center text-sky-400">
                <Send class="w-5 h-5" />
              </div>
              <div>
                <h3 class="text-sm font-bold text-white">Telegram Bot</h3>
                <p class="text-xs font-mono text-sky-400 mt-0.5">{{ telegramHandle }}</p>
              </div>
            </div>
            <span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold bg-[#3ECF8E]/15 text-[#3ECF8E] border border-[#3ECF8E]/30">
              <span class="w-1.5 h-1.5 rounded-full bg-[#3ECF8E]"></span>
              CONNECTED
            </span>
          </div>

          <div v-if="tgTestMessage" class="p-2.5 rounded bg-[#18181B] border border-[#26262A] text-xs font-mono" :class="tgTestSuccess ? 'text-[#34D399]' : 'text-[#F16565]'">
            {{ tgTestMessage }}
          </div>

          <div class="flex items-center justify-between pt-3 border-t border-[#26262A] text-xs">
            <button
              @click="handleTGTest"
              :disabled="isTGTesting"
              class="h-9 px-3.5 rounded-lg border border-[#26262A] bg-[#18181B] hover:bg-[#26262A] text-gray-200 text-xs font-semibold flex items-center gap-2 transition-all disabled:opacity-40"
            >
              <Send class="w-4 h-4 text-sky-400" />
              <span>{{ isTGTesting ? 'Sending...' : 'Send Test Notification' }}</span>
            </button>
            <button
              @click="isTGModalOpen = true"
              class="h-9 px-3.5 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-semibold flex items-center gap-2 shadow-sm shadow-[#7B96F5]/20 transition-all"
            >
              <Sliders class="w-4 h-4" />
              <span>Configure Targets</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Section 2: Engine Polling & Rate-Limit Configuration -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Engine Polling Card -->
      <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-5 shadow-xl">
        <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
          <h2 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono flex items-center gap-2">
            <Activity class="w-4 h-4 text-[#7B96F5]" />
            Engine Polling Settings
          </h2>

          <button
            type="button"
            @click="handleManualRefresh"
            :disabled="isRefreshing"
            class="px-2.5 py-1 rounded-lg bg-[#18181B] border border-[#26262A] hover:border-[#7B96F5] text-[#7B96F5] hover:text-[#95ABF7] text-xs font-mono font-semibold transition-all flex items-center gap-1.5 disabled:opacity-50"
            title="Force immediate ICMP poll cycle"
          >
            <RefreshCw class="w-3.5 h-3.5" :class="isRefreshing ? 'animate-spin' : ''" />
            <span>{{ isRefreshing ? 'Polling...' : 'Refresh Now' }}</span>
          </button>
        </div>

        <div class="space-y-4 text-xs">
          <!-- Polling Interval (Slider 1s–60s) -->
          <div class="space-y-2">
            <div class="flex items-center justify-between font-mono">
              <label class="uppercase text-[10px] text-gray-400 font-semibold">ICMP Polling Interval</label>
              <span class="text-[#7B96F5] font-bold text-sm">{{ settingStore.settings.polling.intervalSeconds }}s</span>
            </div>
            <input
              type="range"
              min="1"
              max="60"
              v-model.number="settingStore.settings.polling.intervalSeconds"
              class="w-full h-1.5 bg-[#18181B] border border-[#26262A] rounded-lg appearance-none cursor-pointer accent-[#7B96F5]"
            />
            <p class="text-[10px] text-gray-500 font-mono">Frequency of ICMP ping probes executed across device inventory</p>
          </div>

          <!-- Concurrency Batch Size -->
          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Concurrency Batch Size</label>
            <input
              type="number"
              min="5"
              max="200"
              v-model.number="settingStore.settings.polling.concurrencyBatchSize"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 font-mono text-xs focus:outline-none focus:border-[#7B96F5]"
            />
            <p class="text-[10px] text-gray-500 font-mono">Parallel probe workers dispatched per cycle batch</p>
          </div>

          <!-- Default Debounce Time -->
          <div class="space-y-1.5 border-t border-[#26262A] pt-3">
            <div class="flex items-center justify-between font-mono">
              <label class="uppercase text-[10px] text-gray-400 font-semibold">Default Failure Debounce Time</label>
              <span class="text-amber-400 font-bold text-xs">
                {{ (settingStore.settings.polling.intervalSeconds || 15) * 3 }}s (3 checks)
              </span>
            </div>
            <p class="text-[10px] text-gray-400 font-mono">
              Minimum 60 seconds recommended to prevent notification flapping.
            </p>
          </div>

          <!-- Flapping Reuse Window -->
          <div class="space-y-1.5 border-t border-[#26262A] pt-3">
            <div class="flex items-center justify-between font-mono">
              <label class="uppercase text-[10px] text-gray-400 font-semibold">Flap Detection Reuse Window</label>
              <span class="text-amber-400 font-bold text-sm">
                {{ settingStore.settings.polling.flapReuseWindowMinutes || 10 }}m
              </span>
            </div>
            <input
              type="range"
              min="1"
              max="60"
              v-model.number="settingStore.settings.polling.flapReuseWindowMinutes"
              class="w-full h-1.5 bg-[#18181B] border border-[#26262A] rounded-lg appearance-none cursor-pointer accent-[#7B96F5]"
            />
            <p class="text-[10px] text-gray-500 font-mono">
              Window duration in minutes to reopen existing incident tickets for flapping devices rather than creating new ones
            </p>
          </div>
        </div>
      </div>

      <!-- Rate-Limit Configuration Card -->
      <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-5 shadow-xl">
        <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
          <h2 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono flex items-center gap-2">
            <Sliders class="w-4 h-4 text-[#7B96F5]" />
            Rate-Limit Configuration
          </h2>
          <span class="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-[#3ECF8E]/10 text-[#3ECF8E] border border-[#3ECF8E]/30">
            ASYNQ QUEUE ACTIVE
          </span>
        </div>

        <div class="space-y-4 text-xs">
          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">
              Max Notifications Per Minute (Queue Spacing)
            </label>
            <div class="flex items-center gap-2">
              <input
                type="number"
                min="1"
                max="300"
                v-model.number="settingStore.settings.rateLimitMaxMsgPerMin"
                class="flex-1 bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 font-mono text-xs focus:outline-none focus:border-[#7B96F5]"
              />
              <span class="text-xs font-mono text-gray-400">msgs/min</span>
            </div>
            <p class="text-[10px] text-gray-500 font-mono">
              Enforces minimum delay spacing between dispatched alert messages to prevent rate-limiting or carrier bans.
            </p>
          </div>
        </div>
      </div>

      <!-- Threshold Defaults Table Card -->
      <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4 shadow-xl col-span-1 md:col-span-2">
        <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
          <div>
            <h2 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono flex items-center gap-2">
              <Sliders class="w-4 h-4 text-[#7B96F5]" />
              Device Type Failure Threshold Defaults
            </h2>
            <p class="text-[11px] text-gray-500 mt-0.5">
              Default consecutive ICMP failure check count before declaring a device DOWN per category
            </p>
          </div>
          <button
            type="button"
            @click="handleSaveThresholds"
            :disabled="isSavingThresholds"
            class="px-3 py-1.5 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-semibold flex items-center gap-1.5 shadow-sm transition-all disabled:opacity-50"
          >
            <Save class="w-3.5 h-3.5" />
            <span>{{ isSavingThresholds ? 'Saving...' : 'Save Thresholds' }}</span>
          </button>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs text-gray-300">
            <thead class="bg-[#18181B] font-mono text-[10px] uppercase text-gray-400">
              <tr>
                <th class="py-2.5 px-4">Device Category</th>
                <th class="py-2.5 px-4">Consecutive ICMP Failures</th>
                <th class="py-2.5 px-4">Effective Debounce Duration</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[#26262A]">
              <tr v-for="t in settingStore.settings.thresholds" :key="t.type" class="hover:bg-[#18181B]">
                <td class="py-2.5 px-4 font-bold text-white font-mono">{{ t.type }}</td>
                <td class="py-2.5 px-4">
                  <div class="flex items-center gap-2">
                    <input
                      type="number"
                      min="1"
                      max="10"
                      v-model.number="t.consecutiveFailures"
                      class="w-24 bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-1.5 text-gray-200 font-mono text-xs focus:outline-none focus:border-[#7B96F5]"
                    />
                    <span class="text-[11px] font-mono text-gray-500">failed checks</span>
                  </div>
                </td>
                <td class="py-2.5 px-4 font-mono text-amber-400 text-xs font-semibold">
                  {{ (t.consecutiveFailures || 3) * (settingStore.settings.polling.intervalSeconds || 15) }} seconds
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Section 3: Users & Roles Management -->
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4 shadow-xl">
      <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
        <h2 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono flex items-center gap-2">
          <Users class="w-4 h-4 text-[#7B96F5]" />
          Users & Roles Management
        </h2>

        <button
          @click="isAddUserModalOpen = true"
          class="px-3 py-1.5 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs flex items-center gap-1.5 shadow-md shadow-[#7B96F5]/20"
        >
          <UserPlus class="w-4 h-4" />
          Add User
        </button>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs text-gray-300">
          <thead class="bg-[#18181B] font-mono text-[10px] uppercase text-gray-400">
            <tr>
              <th class="py-3 px-4">User</th>
              <th class="py-3 px-4">Role</th>
              <th class="py-3 px-4">Status</th>
              <th class="py-3 px-4">Last Active</th>
              <th class="py-3 px-4 text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#26262A]">
            <tr v-for="usr in settingStore.users" :key="usr.id" class="hover:bg-[#18181B]">
              <td class="py-3 px-4 flex items-center gap-3">
                <img :src="usr.avatarUrl" class="w-7 h-7 rounded-full object-cover border border-[#26262A]" />
                <div>
                  <h4 class="font-bold text-white">{{ usr.name }}</h4>
                  <p class="text-[10px] font-mono text-gray-500">{{ usr.email }}</p>
                </div>
              </td>
              <td class="py-3 px-4">
                <span class="px-2 py-0.5 rounded font-mono text-[10px] font-bold uppercase bg-[#7B96F5]/15 text-[#7B96F5] border border-[#7B96F5]/30">
                  {{ usr.role }}
                </span>
              </td>
              <td class="py-3 px-4">
                <span class="inline-flex items-center gap-1.5 text-xs text-[#3ECF8E]">
                  <span class="w-1.5 h-1.5 rounded-full bg-[#3ECF8E]"></span>
                  {{ usr.status }}
                </span>
              </td>
              <td class="py-3 px-4 font-mono text-gray-500 text-[11px]">{{ usr.lastActive }}</td>
              <td class="py-3 px-4 text-right">
                <button
                  @click="openEditUser(usr)"
                  class="px-2.5 py-1 rounded bg-[#18181B] border border-[#26262A] hover:border-[#7B96F5] text-[#7B96F5] hover:text-[#95ABF7] text-[11px] font-mono transition-colors"
                >
                  Edit / Reset Password
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Section 3: User Activity & Session Logs (Paginated & Realtime WS Broadcast) -->
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4 shadow-xl">
      <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
        <div>
          <h2 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono flex items-center gap-2">
            <Activity class="w-4 h-4 text-[#7B96F5]" />
            User Activity & Audit Logs
          </h2>
          <p class="text-[11px] text-gray-500 mt-0.5">Real-time log of logins, logouts, administrative edits, and role changes</p>
        </div>
        <button
          @click="fetchUserLogs"
          class="px-2.5 py-1 rounded bg-[#18181B] border border-[#26262A] hover:border-[#7B96F5] text-gray-300 text-xs font-mono flex items-center gap-1.5"
        >
          <RefreshCw class="w-3.5 h-3.5" :class="isLogsLoading ? 'animate-spin' : ''" />
          Refresh Logs
        </button>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs text-gray-300">
          <thead class="bg-[#18181B] font-mono text-[10px] uppercase text-gray-500">
            <tr>
              <th class="py-2.5 px-3">User</th>
              <th class="py-2.5 px-3">Action</th>
              <th class="py-2.5 px-3">Detail</th>
              <th class="py-2.5 px-3">IP Address</th>
              <th class="py-2.5 px-3 text-right">Occurred At</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#26262A]">
            <tr v-if="userLogs.length === 0">
              <td colspan="5" class="py-4 text-center text-gray-500 font-mono text-xs">No activity logs recorded yet</td>
            </tr>
            <tr v-for="log in userLogs" :key="log.id" class="hover:bg-[#18181B]">
              <td class="py-2.5 px-3 font-semibold text-white">{{ log.userName || log.userId }}</td>
              <td class="py-2.5 px-3">
                <span
                  class="px-2 py-0.5 rounded text-[10px] font-mono font-bold uppercase"
                  :class="[
                    log.action === 'login' ? 'bg-[#3ECF8E]/15 text-[#3ECF8E] border border-[#3ECF8E]/30' :
                    log.action === 'logout' ? 'bg-amber-500/15 text-amber-400 border border-amber-500/30' :
                    'bg-[#7B96F5]/15 text-[#7B96F5] border border-[#7B96F5]/30'
                  ]"
                >
                  {{ log.action }}
                </span>
              </td>
              <td class="py-2.5 px-3 font-mono text-gray-300 text-[11px]">{{ log.detail }}</td>
              <td class="py-2.5 px-3 font-mono text-gray-400 text-[10px]">{{ log.ipAddress || '127.0.0.1' }}</td>
              <td class="py-2.5 px-3 text-right font-mono text-gray-500 text-[10px]">
                {{ new Date(log.occurredAt || log.timestamp || Date.now()).toLocaleString('id-ID', { hour12: false }) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <PaginationControl
        v-model:current-page="logsPage"
        v-model:page-size="logsPageSize"
        :total="logsTotal"
        @change="fetchUserLogs"
      />
    </div>

    <!-- Add User Modal (Direct Creation with Password) -->
    <Modal :is-open="isAddUserModalOpen" title="Add New User" @close="isAddUserModalOpen = false">
      <template #default>
        <form @submit.prevent="handleAddUser" class="space-y-4 text-xs">
          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-gray-400">Full Name *</label>
            <input
              v-model="addUserForm.name"
              type="text"
              required
              placeholder="e.g. Ahmad Hidayat"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
            />
          </div>

          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-gray-400">Email Address *</label>
            <input
              v-model="addUserForm.email"
              type="email"
              required
              placeholder="e.g. ahmad@jabarprov.go.id"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
            />
          </div>

          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-gray-400">Initial Password *</label>
            <input
              v-model="addUserForm.password"
              type="password"
              required
              placeholder="••••••••"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
            />
          </div>

          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-gray-400">Role *</label>
            <select
              v-model="addUserForm.role"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5] font-mono"
            >
              <option value="admin">ADMIN (Full Unrestricted Access)</option>
              <option value="anggota">NOC STAFF (Operator)</option>
              <option value="pimpinan">PIMPINAN (Executive)</option>
            </select>
          </div>
        </form>
      </template>

      <template #footer>
        <button
          @click="isAddUserModalOpen = false"
          class="px-4 py-2 rounded-lg border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs"
        >
          Cancel
        </button>
        <button
          @click="handleAddUser"
          class="px-5 py-2 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20"
        >
          Create User
        </button>
      </template>
    </Modal>

    <!-- Modals -->
    <WhatsAppTargetModal :is-open="isWATargetModalOpen" @close="isWATargetModalOpen = false" />
    <WhatsAppQRModal :is-open="isWAModalOpen" @close="isWAModalOpen = false" @connected="onWAConnected" />
    <UserEditModal :is-open="isUserEditModalOpen" :user="selectedUser" @close="isUserEditModalOpen = false" @saved="settingStore.fetchUsers()" />
    <TelegramConfigModal :is-open="isTGModalOpen" :bot-token="telegramToken" :chat-id="telegramChatId" @close="isTGModalOpen = false" @saved="onTGSaved" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue';
import { useSettingStore } from '../stores/settingStore';
import Modal from '../components/common/Modal.vue';
import PaginationControl from '../components/common/PaginationControl.vue';
import WhatsAppQRModal from '../components/settings/WhatsAppQRModal.vue';
import TelegramConfigModal from '../components/settings/TelegramConfigModal.vue';
import WhatsAppTargetModal from '../components/settings/WhatsAppTargetModal.vue';
import UserEditModal from '../components/users/UserEditModal.vue';
import type { User, UserRole } from '../types';
import {
  Save,
  Send,
  MessageSquare,
  QrCode,
  Activity,
  Sliders,
  Users,
  UserPlus,
  RefreshCw,
  Target,
  LogOut
} from 'lucide-vue-next';
import api from '../api/client';
import { wsClient } from '../ws/websocket';

const settingStore = useSettingStore();
const isSaveSuccess = ref(false);

const isAddUserModalOpen = ref(false);
const isWATargetModalOpen = ref(false);
const isWAModalOpen = ref(false);
const isTGModalOpen = ref(false);
const isUserEditModalOpen = ref(false);
const selectedUser = ref<User | null>(null);

function openEditUser(usr: User) {
  selectedUser.value = usr;
  isUserEditModalOpen.value = true;
}

const isWhatsAppConnected = ref(false);
const whatsAppNumber = ref('');
const telegramToken = ref('');
const telegramChatId = ref('');
const telegramHandle = computed(() => telegramChatId.value || 'Not Configured');

const isWATesting = ref(false);
const waTestSuccess = ref(true);
const waTestMessage = ref('');

const isTGTesting = ref(false);
const tgTestSuccess = ref(true);
const tgTestMessage = ref('');

const isRefreshing = ref(false);
const isSavingThresholds = ref(false);

async function handleSaveThresholds() {
  isSavingThresholds.value = true;
  try {
    await api.put('/settings/thresholds', {
      thresholds: settingStore.settings.thresholds
    });
  } catch (e) {
    // fallback
  } finally {
    isSavingThresholds.value = false;
  }
}

async function handleManualRefresh() {
  isRefreshing.value = true;
  try {
    await api.post('/monitoring/refresh-now');
  } catch (e) {
    // fallback
  } finally {
    setTimeout(() => { isRefreshing.value = false; }, 600);
  }
}

const addUserForm = reactive({
  name: '',
  email: '',
  password: '',
  role: 'anggota' as UserRole
});

const userLogs = ref<any[]>([]);
const isLogsLoading = ref(false);
const logsPage = ref(1);
const logsPageSize = ref(10);
const logsTotal = ref(0);

async function fetchUserLogs() {
  isLogsLoading.value = true;
  try {
    const res = await api.get('/user-logs', {
      params: { page: logsPage.value, page_size: logsPageSize.value }
    });
    if (res.data && Array.isArray(res.data.items)) {
      userLogs.value = res.data.items;
      logsTotal.value = res.data.total || res.data.items.length;
    } else if (Array.isArray(res.data)) {
      userLogs.value = res.data;
      logsTotal.value = res.data.length;
    }
  } catch (e) {
    // fallback
  } finally {
    isLogsLoading.value = false;
  }
}

let unsubscribeWS: (() => void) | null = null;

onMounted(() => {
  settingStore.fetchSettings();
  settingStore.fetchUsers();
  fetchIntegrationsConfig();
  fetchUserLogs();

  wsClient.connect();
  unsubscribeWS = wsClient.subscribe((msg: any) => {
    if (msg.type === 'USER_LOG_CREATED') {
      userLogs.value.unshift({
        id: 'log-' + Date.now(),
        userName: msg.title || 'User Activity',
        action: 'event',
        detail: msg.description || 'Activity logged',
        ipAddress: '127.0.0.1',
        occurredAt: msg.timestamp || new Date().toISOString()
      });
      if (userLogs.value.length > logsPageSize.value) {
        userLogs.value.pop();
      }
      logsTotal.value++;
    }
  });
});

onUnmounted(() => {
  if (unsubscribeWS) {
    unsubscribeWS();
    unsubscribeWS = null;
  }
});

async function fetchIntegrationsConfig() {
  try {
    const tgRes = await api.get('/integrations/telegram/config');
    if (tgRes.data.botToken) telegramToken.value = tgRes.data.botToken;
    if (tgRes.data.chatId) telegramChatId.value = tgRes.data.chatId;
  } catch (e) {
    // fallback
  }

  try {
    const waRes = await api.get('/integrations/whatsapp/status');
    if (waRes.data.status === 'connected') {
      isWhatsAppConnected.value = true;
      whatsAppNumber.value = waRes.data.linkedNumber || '';
    } else {
      isWhatsAppConnected.value = false;
      whatsAppNumber.value = '';
    }
  } catch (e) {
    // fallback
  }
}

async function handleSaveConfig() {
  await settingStore.saveSettings();
  isSaveSuccess.value = true;
  setTimeout(() => {
    isSaveSuccess.value = false;
  }, 2500);
}

function onWAConnected(num: string) {
  isWhatsAppConnected.value = true;
  whatsAppNumber.value = num || '+62 812-9000-8888';
}

async function handleWADisconnect() {
  try {
    await api.post('/integrations/whatsapp/disconnect');
  } catch (e) {
    // fallback
  }
  isWhatsAppConnected.value = false;
}

async function handleWATest() {
  isWATesting.value = true;
  waTestMessage.value = '';
  try {
    await api.post('/integrations/whatsapp/test');
    waTestSuccess.value = true;
    waTestMessage.value = '✓ Test notification delivered to configured targets';
  } catch (e: any) {
    waTestSuccess.value = false;
    waTestMessage.value = '✗ Test failed: ' + (e.response?.data?.error || 'Gateway offline');
  } finally {
    isWATesting.value = false;
  }
}

function onTGSaved(cfg: { botToken: string; chatId: string }) {
  telegramToken.value = cfg.botToken;
  telegramChatId.value = cfg.chatId;
}

async function handleTGTest() {
  isTGTesting.value = true;
  tgTestMessage.value = '';
  try {
    await api.post('/integrations/telegram/test', {
      botToken: telegramToken.value,
      chatId: telegramChatId.value
    });
    tgTestSuccess.value = true;
    tgTestMessage.value = '✓ Test broadcast delivered to Telegram Channel ' + telegramChatId.value;
  } catch (e: any) {
    tgTestSuccess.value = false;
    tgTestMessage.value = '✗ Test failed: ' + (e.response?.data?.error || 'API Error');
  } finally {
    isTGTesting.value = false;
  }
}

async function handleAddUser() {
  if (!addUserForm.name || !addUserForm.email || !addUserForm.password) return;
  try {
    await api.post('/users', {
      name: addUserForm.name,
      email: addUserForm.email,
      role: addUserForm.role,
      password: addUserForm.password
    });
    await settingStore.fetchUsers();
    addUserForm.name = '';
    addUserForm.email = '';
    addUserForm.password = '';
    isAddUserModalOpen.value = false;
  } catch (e: any) {
    alert(e.response?.data?.error || 'Failed to create user');
  }
}
</script>
