<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-[#26262A] pb-5">
      <div>
        <div class="flex items-center gap-2">
          <h1 class="text-xl font-extrabold text-white tracking-tight">System Settings & Administration</h1>
          <span class="px-2 py-0.5 rounded-full text-[10px] font-mono font-bold uppercase bg-[#7B96F5]/10 text-[#7B96F5] border border-[#7B96F5]/25">
            {{ currentRoleName }}
          </span>
        </div>
        <p class="text-xs text-gray-400 mt-1">
          Configure notification gateways, polling engine parameters, data retention, and user role management
        </p>
      </div>
    </div>

    <!-- Skeleton Loading State -->
    <div v-if="settingStore.isLoading && !isInitialLoaded" class="grid grid-cols-1 lg:grid-cols-4 gap-6">
      <div class="lg:col-span-1 bg-[#151517] border border-[#26262A] rounded-2xl p-4 space-y-3">
        <Skeleton width="40%" height="1rem" />
        <div v-for="i in 6" :key="i" class="p-3 bg-[#18181B] border border-[#26262A] rounded-xl space-y-2">
          <Skeleton width="60%" height="0.85rem" />
          <Skeleton width="80%" height="0.65rem" />
        </div>
      </div>
      <div class="lg:col-span-3 bg-[#151517] border border-[#26262A] rounded-2xl p-6 space-y-6">
        <Skeleton width="30%" height="1.5rem" />
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div v-for="i in 4" :key="i" class="p-5 bg-[#18181B] border border-[#26262A] rounded-xl space-y-3">
            <Skeleton width="50%" height="1rem" />
            <Skeleton width="100%" height="2rem" />
          </div>
        </div>
      </div>
    </div>

    <!-- Access Restricted (403) Screen when user has 0 settings permissions -->
    <div v-else-if="availableTabs.length === 0" class="p-12 text-center bg-[#151517] border border-[#26262A] rounded-2xl space-y-4 max-w-lg mx-auto shadow-2xl animate-fadeIn">
      <div class="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-red-500/10 border border-red-500/30 text-red-400">
        <ShieldAlert class="w-7 h-7" />
      </div>
      <h2 class="text-base font-bold text-white font-mono">Access Restricted (403)</h2>
      <p class="text-xs text-gray-400 leading-relaxed font-mono">
        Your current role (<strong class="text-[#7B96F5]">{{ currentRoleName }}</strong>) does not have permission to view or modify any system configuration categories.
      </p>
      <div class="pt-3">
        <router-link
          to="/dashboard"
          class="px-5 py-2.5 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs font-mono inline-flex items-center gap-2 shadow-lg shadow-[#7B96F5]/20 cursor-pointer"
        >
          Back to Dashboard
        </router-link>
      </div>
    </div>

    <!-- Main Sub-Sidebar & Content Layout -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-4 gap-6 items-start">
      <!-- Sub-Sidebar Navigation (Left Column) -->
      <aside class="lg:col-span-1 bg-[#151517] border border-[#26262A] rounded-2xl p-3.5 space-y-1.5 shadow-xl sticky top-4">
        <div class="px-3 py-2 border-b border-[#26262A]/60 mb-2">
          <span class="text-[10px] font-mono font-bold uppercase tracking-wider text-gray-400">Settings Categories</span>
        </div>

        <nav class="space-y-1" aria-label="Settings Categories">
          <button
            v-for="tab in availableTabs"
            :key="tab.id"
            @click="switchTab(tab.id)"
            class="w-full text-left p-3 rounded-xl transition-all flex items-start gap-3 group relative cursor-pointer"
            :class="[
              activeTab === tab.id
                ? 'bg-[#7B96F5]/10 border border-[#7B96F5]/30 text-white shadow-sm shadow-[#7B96F5]/10'
                : 'border border-transparent text-gray-400 hover:text-gray-200 hover:bg-[#18181B] hover:border-[#26262A]'
            ]"
          >
            <component
              :is="tab.icon"
              class="w-4 h-4 shrink-0 mt-0.5 transition-colors"
              :class="activeTab === tab.id ? 'text-[#7B96F5]' : 'text-gray-400 group-hover:text-gray-200'"
            />
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between">
                <span class="text-xs font-bold font-mono tracking-tight truncate" :class="activeTab === tab.id ? 'text-white' : 'text-gray-300'">
                  {{ tab.label }}
                </span>
                <span
                  v-if="tab.badge"
                  class="text-[9px] font-mono px-1.5 py-0.5 rounded font-bold uppercase"
                  :class="tab.badgeClass || 'bg-[#26262A] text-gray-400'"
                >
                  {{ tab.badge }}
                </span>
              </div>
              <p class="text-[10px] font-sans text-gray-400 truncate mt-0.5">{{ tab.description }}</p>
            </div>

            <!-- Active Indicator Dot -->
            <span
              v-if="activeTab === tab.id"
              class="w-1.5 h-1.5 rounded-full bg-[#7B96F5] absolute right-2.5 top-1/2 -translate-y-1/2 shadow-sm shadow-[#7B96F5]"
            ></span>
          </button>
        </nav>
      </aside>

      <!-- Content Area (Right Column) -->
      <main class="lg:col-span-3 space-y-6 min-w-0">
        <!-- ══════════════════════════════════════════════════════════════════════════
             TAB 1: Notification Channels & Rate Limits
             ══════════════════════════════════════════════════════════════════════════ -->
        <div v-if="activeTab === 'notifications'" class="space-y-6 animate-fadeIn">
          <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-white font-mono flex items-center gap-2">
                <Send class="w-4 h-4 text-[#7B96F5]" />
                NOTIFICATION GATEWAYS &amp; RATE LIMITS
              </h2>
              <p class="text-xs text-gray-400 mt-0.5">Configure primary and fallback alert channels and queue rate-limiting.</p>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- WhatsApp API Card -->
            <div class="bg-[#151517] border border-[#26262A] rounded-2xl p-5 space-y-4 shadow-xl">
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

              <div v-if="waTestMessage" class="p-2.5 rounded-lg bg-[#18181B] border border-[#26262A] text-xs font-mono" :class="waTestSuccess ? 'text-[#34D399]' : 'text-[#F16565]'">
                {{ waTestMessage }}
              </div>

              <div class="pt-3 border-t border-[#26262A] space-y-2">
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <button
                    @click="handleWATest"
                    :disabled="!isWhatsAppConnected || isWATesting"
                    class="h-9 px-3 rounded-lg border border-[#26262A] bg-[#18181B] hover:bg-[#26262A] text-gray-200 text-xs font-semibold flex items-center justify-center gap-1.5 transition-all disabled:opacity-40 whitespace-nowrap cursor-pointer"
                  >
                    <Send class="w-3.5 h-3.5 text-[#7B96F5]" />
                    <span>{{ isWATesting ? 'Sending...' : 'Send Test Notification' }}</span>
                  </button>
                  <button
                    @click="isWATargetModalOpen = true"
                    class="h-9 px-3 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-semibold flex items-center justify-center gap-1.5 shadow-sm shadow-[#7B96F5]/20 transition-all whitespace-nowrap cursor-pointer"
                  >
                    <Target class="w-3.5 h-3.5" />
                    <span>Configure Targets</span>
                  </button>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <button
                    @click="isWAModalOpen = true"
                    class="h-9 px-3 rounded-lg bg-[#18181B] border border-[#26262A] hover:bg-[#26262A] text-gray-200 text-xs font-semibold flex items-center justify-center gap-1.5 transition-all whitespace-nowrap cursor-pointer"
                  >
                    <QrCode class="w-3.5 h-3.5 text-gray-400" />
                    <span>QR Reconnect</span>
                  </button>
                  <button
                    v-if="isWhatsAppConnected"
                    @click="handleWADisconnect"
                    class="h-9 px-3 rounded-lg border border-red-500/30 bg-red-500/10 hover:bg-red-500/20 text-red-400 text-xs font-semibold flex items-center justify-center gap-1.5 transition-all whitespace-nowrap cursor-pointer"
                  >
                    <LogOut class="w-3.5 h-3.5 text-red-400" />
                    <span>Disconnect</span>
                  </button>
                </div>
              </div>
            </div>

            <!-- Telegram Bot Card -->
            <div class="bg-[#151517] border border-[#26262A] rounded-2xl p-5 space-y-4 shadow-xl">
              <div class="flex items-start justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-xl bg-sky-500/15 border border-sky-500/30 flex items-center justify-center text-sky-400">
                    <Send class="w-5 h-5" />
                  </div>
                  <div>
                    <h3 class="text-sm font-bold text-white">Telegram Bot Gateway</h3>
                    <p class="text-xs font-mono text-sky-400 mt-0.5">{{ telegramHandle }}</p>
                  </div>
                </div>
                <span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold bg-[#3ECF8E]/15 text-[#3ECF8E] border border-[#3ECF8E]/30">
                  <span class="w-1.5 h-1.5 rounded-full bg-[#3ECF8E]"></span>
                  ACTIVE FALLBACK
                </span>
              </div>

              <div v-if="tgTestMessage" class="p-2.5 rounded-lg bg-[#18181B] border border-[#26262A] text-xs font-mono" :class="tgTestSuccess ? 'text-[#34D399]' : 'text-[#F16565]'">
                {{ tgTestMessage }}
              </div>

              <div class="pt-3 border-t border-[#26262A]">
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <button
                    @click="handleTGTest"
                    :disabled="isTGTesting"
                    class="h-9 px-3 rounded-lg border border-[#26262A] bg-[#18181B] hover:bg-[#26262A] text-gray-200 text-xs font-semibold flex items-center justify-center gap-1.5 transition-all disabled:opacity-40 whitespace-nowrap cursor-pointer"
                  >
                    <Send class="w-3.5 h-3.5 text-sky-400" />
                    <span>{{ isTGTesting ? 'Sending...' : 'Send Test Notification' }}</span>
                  </button>
                  <button
                    @click="isTGModalOpen = true"
                    class="h-9 px-3 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-semibold flex items-center justify-center gap-1.5 shadow-sm shadow-[#7B96F5]/20 transition-all whitespace-nowrap cursor-pointer"
                  >
                    <Sliders class="w-3.5 h-3.5" />
                    <span>Configure Channel</span>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Rate Limit Card -->
          <div class="bg-[#151517] border border-[#26262A] rounded-2xl p-5 space-y-4 shadow-xl">
            <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
              <h3 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono flex items-center gap-2">
                <Sliders class="w-4 h-4 text-[#7B96F5]" />
                Asynq Redis Queue Rate-Limit Spacing
              </h3>
              <button
                type="button"
                @click="handleSaveRateLimit"
                :disabled="isSavingRateLimit"
                class="px-3.5 py-1.5 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-semibold flex items-center gap-1.5 shadow-md shadow-[#7B96F5]/20 transition-all disabled:opacity-50 cursor-pointer"
              >
                <Save class="w-3.5 h-3.5" />
                <span>{{ isSavingRateLimit ? 'Saving...' : 'Save Queue Settings' }}</span>
              </button>
            </div>

            <div class="space-y-3 text-xs max-w-md">
              <div class="space-y-1.5">
                <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">
                  Max Dispatched Notifications Per Minute
                </label>
                <div class="flex items-center gap-2">
                  <input
                    type="number"
                    min="1"
                    max="300"
                    v-model.number="settingStore.settings.rateLimitMaxMsgPerMin"
                    class="w-32 bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 font-mono text-xs focus:outline-none focus:border-[#7B96F5]"
                  />
                  <span class="text-xs font-mono text-gray-400">messages / min</span>
                </div>
                <p class="text-[10px] text-gray-500 font-mono">
                  Enforces minimum delay spacing between dispatched alert messages to prevent rate-limiting or carrier bans.
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- ══════════════════════════════════════════════════════════════════════════
             TAB 2: Engine Polling & Failure Thresholds
             ══════════════════════════════════════════════════════════════════════════ -->
        <div v-else-if="activeTab === 'polling'" class="space-y-6 animate-fadeIn">
          <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-white font-mono flex items-center gap-2">
                <Activity class="w-4 h-4 text-[#7B96F5]" />
                ENGINE POLLING &amp; FAILURE THRESHOLDS
              </h2>
              <p class="text-xs text-gray-400 mt-0.5">Control ICMP probe frequencies, concurrency workers, debouncing, and anti-flap reuse.</p>
            </div>
            <div class="flex items-center gap-2">
              <button
                type="button"
                @click="handleManualRefresh"
                :disabled="isRefreshing"
                class="px-3 py-1.5 rounded-xl bg-[#18181B] border border-[#26262A] hover:border-[#7B96F5] text-[#7B96F5] hover:text-[#95ABF7] text-xs font-mono font-semibold transition-all flex items-center gap-1.5 disabled:opacity-50 cursor-pointer"
                title="Force immediate ICMP poll cycle"
              >
                <RefreshCw class="w-3.5 h-3.5" :class="isRefreshing ? 'animate-spin' : ''" />
                <span>{{ isRefreshing ? 'Polling...' : 'Trigger Poll Now' }}</span>
              </button>
              <button
                type="button"
                @click="handleSaveEngineAndThresholds"
                :disabled="isSavingEngine"
                class="px-3.5 py-1.5 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-semibold flex items-center gap-1.5 shadow-md shadow-[#7B96F5]/20 transition-all disabled:opacity-50 cursor-pointer"
              >
                <Save class="w-3.5 h-3.5" />
                <span>{{ isSavingEngine ? 'Saving...' : 'Save Engine & Thresholds' }}</span>
              </button>
            </div>
          </div>

          <!-- Engine Polling Settings Card -->
          <div class="bg-[#151517] border border-[#26262A] rounded-2xl p-5 space-y-5 shadow-xl">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6 text-xs">
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
              <div class="space-y-2">
                <div class="flex items-center justify-between font-mono">
                  <label class="uppercase text-[10px] text-gray-400 font-semibold">Concurrency Batch Size</label>
                  <span class="text-white font-bold font-mono">{{ settingStore.settings.polling.concurrencyBatchSize }} workers</span>
                </div>
                <input
                  type="number"
                  min="5"
                  max="200"
                  v-model.number="settingStore.settings.polling.concurrencyBatchSize"
                  class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 font-mono text-xs focus:outline-none focus:border-[#7B96F5]"
                />
                <p class="text-[10px] text-gray-500 font-mono">Parallel probe workers dispatched per cycle batch</p>
              </div>

              <!-- Flapping Reuse Window -->
              <div class="space-y-2 border-t border-[#26262A] pt-4 md:col-span-2">
                <div class="flex items-center justify-between font-mono">
                  <label class="uppercase text-[10px] text-gray-400 font-semibold">Flap Detection Reuse Window</label>
                  <span class="text-amber-400 font-bold text-sm">
                    {{ settingStore.settings.polling.flapReuseWindowMinutes || 10 }} minutes
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
                  Window duration in minutes to reopen existing incident tickets for flapping devices rather than creating duplicate incidents.
                </p>
              </div>
            </div>
          </div>

          <!-- Device Type Failure Threshold Defaults Table Card -->
          <div class="bg-[#151517] border border-[#26262A] rounded-2xl p-5 space-y-4 shadow-xl">
            <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
              <div>
                <h3 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono flex items-center gap-2">
                  <Sliders class="w-4 h-4 text-[#7B96F5]" />
                  Device Category Failure Thresholds
                </h3>
                <p class="text-[11px] text-gray-500 mt-0.5">
                  Consecutive ICMP check confirmations required before flipping state (DOWN &harr; UP)
                </p>
              </div>
            </div>

            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs text-gray-300">
                <thead class="bg-[#18181B] font-mono text-[10px] uppercase text-gray-400">
                  <tr>
                    <th class="py-2.5 px-4">Device Category</th>
                    <th class="py-2.5 px-4">Consecutive ICMP Checks</th>
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
                          class="w-20 bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-1.5 text-gray-200 font-mono text-xs focus:outline-none focus:border-[#7B96F5]"
                        />
                        <span class="text-[11px] font-mono text-gray-500">checks</span>
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

        <!-- ══════════════════════════════════════════════════════════════════════════
             TAB 3: Core Router / Switch SNMP Settings
             ══════════════════════════════════════════════════════════════════════════ -->
        <div v-else-if="activeTab === 'network'" class="space-y-6 animate-fadeIn">
          <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-white font-mono flex items-center gap-2">
                <Network class="w-4 h-4 text-[#7B96F5]" />
                CORE ROUTER / SWITCH SNMP DISCOVERY
              </h2>
              <p class="text-xs text-gray-400 mt-0.5">Cross-subnet L3 ARP table querying for automatic IP-to-MAC resolution.</p>
            </div>
            <button
              type="button"
              @click="handleSaveCoreSwitch"
              :disabled="isSavingCoreSwitch"
              class="px-3.5 py-1.5 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-semibold flex items-center gap-1.5 shadow-md shadow-[#7B96F5]/20 transition-all disabled:opacity-50 cursor-pointer"
            >
              <Save class="w-3.5 h-3.5" />
              <span>{{ isSavingCoreSwitch ? 'Saving...' : 'Save SNMP Target' }}</span>
            </button>
          </div>

          <div class="bg-[#151517] border border-[#26262A] rounded-2xl p-5 space-y-4 shadow-xl">
            <div v-if="settingStore.settings.coreSwitch" class="space-y-4 text-xs">
              <div class="grid grid-cols-1 sm:grid-cols-4 gap-4">
                <div class="space-y-1.5 sm:col-span-1">
                  <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Core Switch IP</label>
                  <input
                    type="text"
                    v-model="settingStore.settings.coreSwitch.ip"
                    placeholder="e.g. 10.10.1.1"
                    class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 font-mono text-xs focus:outline-none focus:border-[#7B96F5]"
                  />
                </div>
                <div class="space-y-1.5 sm:col-span-1">
                  <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Community String</label>
                  <input
                    type="text"
                    v-model="settingStore.settings.coreSwitch.community"
                    placeholder="public"
                    class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 font-mono text-xs focus:outline-none focus:border-[#7B96F5]"
                  />
                </div>
                <div class="space-y-1.5 sm:col-span-1">
                  <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Port</label>
                  <input
                    type="number"
                    v-model.number="settingStore.settings.coreSwitch.port"
                    placeholder="161"
                    class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 font-mono text-xs focus:outline-none focus:border-[#7B96F5]"
                  />
                </div>
                <div class="space-y-1.5 sm:col-span-1">
                  <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">SNMP Version</label>
                  <select
                    v-model="settingStore.settings.coreSwitch.version"
                    class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 font-mono text-xs focus:outline-none focus:border-[#7B96F5]"
                  >
                    <option value="v2c">v2c</option>
                    <option value="v1">v1</option>
                  </select>
                </div>
              </div>

              <div class="p-3.5 rounded-xl bg-[#7B96F5]/10 border border-[#7B96F5]/25 text-[11px] font-mono text-gray-200 flex items-start gap-3">
                <AlertCircle class="w-4 h-4 text-[#7B96F5] shrink-0 mt-0.5" />
                <div class="space-y-1 leading-relaxed">
                  <p class="font-bold text-[#7B96F5]">Layer-3 Cross-Subnet MAC Resolution:</p>
                  <p class="text-gray-300">
                    Host OS <code class="bg-black/40 px-1 py-0.5 rounded text-gray-200">arp -a</code> cannot inspect MAC addresses across VLAN subnets. Auto Detect queries this Core Router / Switch via SNMP OID <code class="bg-black/40 px-1 py-0.5 rounded text-[#7B96F5]">ipNetToMediaPhysAddress (.1.3.6.1.2.1.4.22.1.2)</code> to accurately correlate cross-subnet IP addresses to device MAC addresses.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- ══════════════════════════════════════════════════════════════════════════
             TAB 4: Incident Retention & Archiving Policy
             ══════════════════════════════════════════════════════════════════════════ -->
        <div v-else-if="activeTab === 'retention'" class="space-y-6 animate-fadeIn">
          <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-white font-mono flex items-center gap-2">
                <Archive class="w-4 h-4 text-[#7B96F5]" />
                INCIDENT RETENTION &amp; ARCHIVING POLICY
              </h2>
              <p class="text-xs text-gray-400 mt-0.5">Automated database archiving and housekeeping schedules for resolved tickets.</p>
            </div>
            <button
              type="button"
              @click="handleSaveRetention"
              :disabled="isSavingRetention"
              class="px-3.5 py-1.5 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white text-xs font-semibold flex items-center gap-1.5 shadow-md shadow-[#7B96F5]/20 transition-all disabled:opacity-50 cursor-pointer"
            >
              <Save class="w-3.5 h-3.5" />
              <span>{{ isSavingRetention ? 'Saving...' : 'Save Retention Policy' }}</span>
            </button>
          </div>

          <div class="bg-[#151517] border border-[#26262A] rounded-2xl p-5 space-y-4 shadow-xl">
            <div class="space-y-4 text-xs max-w-lg">
              <div class="space-y-1.5">
                <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">
                  Auto-Archive Resolved Incidents Older Than
                </label>
                <div class="flex items-center gap-2">
                  <input
                    type="number"
                    min="7"
                    max="365"
                    v-model.number="settingStore.settings.retentionDays"
                    class="w-32 bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 font-mono text-xs focus:outline-none focus:border-[#7B96F5]"
                  />
                  <span class="text-xs font-mono text-gray-400">days</span>
                </div>
                <p class="text-[10px] text-gray-500 font-mono">
                  Resolved tickets older than this threshold are safely transferred to <code class="text-gray-300">incidents_archive</code>. Active/Open incidents are never purged.
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- ══════════════════════════════════════════════════════════════════════════
             TAB 5: Location Management
             ══════════════════════════════════════════════════════════════════════════ -->
        <div v-else-if="activeTab === 'locations'" class="space-y-6 animate-fadeIn">
          <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-white font-mono flex items-center gap-2">
                <MapPin class="w-4 h-4 text-[#7B96F5]" />
                LOCATION &amp; SITE MANAGEMENT
              </h2>
              <p class="text-xs text-gray-400 mt-0.5">Manage installation sites, buildings, floors, and server rooms.</p>
            </div>
            <button
              @click="openAddLocationModal"
              class="px-3.5 py-1.5 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs flex items-center gap-1.5 shadow-md shadow-[#7B96F5]/20 cursor-pointer"
            >
              <Plus class="w-4 h-4" />
              Add Location
            </button>
          </div>

          <div class="bg-[#151517] border border-[#26262A] rounded-2xl p-5 space-y-4 shadow-xl">
            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs text-gray-300">
                <thead class="bg-[#18181B] font-mono text-[10px] uppercase text-gray-400">
                  <tr>
                    <th class="py-3 px-4">Location Name</th>
                    <th class="py-3 px-4">Description</th>
                    <th class="py-3 px-4">Assigned Devices</th>
                    <th class="py-3 px-4 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-[#26262A]">
                  <tr v-if="isLocationsLoading">
                    <td colspan="4" class="p-0 border-0"><SkeletonTable :rows="3" :cols="4" /></td>
                  </tr>
                  <tr v-else-if="locationsList.length === 0">
                    <td colspan="4" class="py-4 text-center text-gray-500 font-mono text-xs">No locations registered</td>
                  </tr>
                  <tr v-else v-for="loc in locationsList" :key="loc.id" class="hover:bg-[#18181B]">
                    <td class="py-3 px-4 font-bold text-white font-mono flex items-center gap-2">
                      <MapPin class="w-3.5 h-3.5 text-[#7B96F5]" />
                      <span>{{ loc.name }}</span>
                    </td>
                    <td class="py-3 px-4 text-gray-400 font-mono">{{ loc.description || '-' }}</td>
                    <td class="py-3 px-4 font-mono">
                      <span class="px-2 py-0.5 rounded text-[10px] font-bold" :class="loc.deviceCount ? 'bg-[#7B96F5]/10 text-[#7B96F5]' : 'bg-gray-500/10 text-gray-400'">
                        {{ loc.deviceCount || 0 }} devices
                      </span>
                    </td>
                    <td class="py-3 px-4 text-right">
                      <div class="flex items-center justify-end gap-2">
                        <button
                          @click="openEditLocationModal(loc)"
                          class="px-2.5 py-1 rounded-lg bg-[#18181B] border border-[#26262A] hover:border-[#7B96F5] text-[#7B96F5] hover:text-[#95ABF7] text-[11px] font-mono transition-colors cursor-pointer"
                        >
                          Edit
                        </button>
                        <button
                          @click="confirmDeleteLocation(loc)"
                          class="px-2.5 py-1 rounded-lg bg-red-500/10 border border-red-500/30 hover:bg-red-500/20 text-red-400 hover:text-red-300 text-[11px] font-mono transition-colors flex items-center gap-1 cursor-pointer"
                        >
                          <Trash2 class="w-3 h-3" />
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- ══════════════════════════════════════════════════════════════════════════
             TAB 6: User Accounts & Roles Management
             ══════════════════════════════════════════════════════════════════════════ -->
        <div v-else-if="activeTab === 'users'" class="space-y-6 animate-fadeIn">
          <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-white font-mono flex items-center gap-2">
                <Users class="w-4 h-4 text-[#7B96F5]" />
                USERS &amp; ROLES MANAGEMENT
              </h2>
              <p class="text-xs text-gray-400 mt-0.5">Manage operator accounts, privileges, and system roles.</p>
            </div>
            <button
              @click="openAddUserModal"
              class="px-3.5 py-1.5 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs flex items-center gap-1.5 shadow-md shadow-[#7B96F5]/20 cursor-pointer"
            >
              <UserPlus class="w-4 h-4" />
              Add User
            </button>
          </div>

          <div class="bg-[#151517] border border-[#26262A] rounded-2xl p-5 space-y-4 shadow-xl">
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
                  <tr v-if="settingStore.isLoading">
                    <td colspan="5" class="p-0 border-0"><SkeletonTable :rows="3" :cols="5" /></td>
                  </tr>
                  <tr v-else-if="paginatedUsers.length === 0">
                    <td colspan="5" class="py-4 text-center text-gray-500 font-mono text-xs">No user accounts registered</td>
                  </tr>
                  <tr v-else v-for="usr in paginatedUsers" :key="usr.id" class="hover:bg-[#18181B]">
                    <td class="py-3 px-4 flex items-center gap-3">
                      <!-- User Profile Picture with Initials Fallback -->
                      <img
                        v-if="usr.avatarUrl"
                        :src="usr.avatarUrl"
                        class="w-8 h-8 rounded-full object-cover border border-[#26262A]"
                        alt="Avatar"
                      />
                      <div
                        v-else
                        class="w-8 h-8 rounded-full bg-[#7B96F5]/15 border border-[#7B96F5]/30 flex items-center justify-center font-bold text-[#7B96F5] font-mono text-xs"
                      >
                        {{ (usr.name || usr.username || 'U').charAt(0).toUpperCase() }}
                      </div>
                      <div>
                        <h4 class="font-bold text-white flex items-center gap-2">
                          <span>{{ usr.name }}</span>
                          <span v-if="usr.username" class="text-[10px] font-mono text-[#7B96F5]">@{{ usr.username }}</span>
                        </h4>
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
                      <div class="flex items-center justify-end gap-2">
                        <button
                          @click="openEditUser(usr)"
                          class="px-2.5 py-1 rounded-lg bg-[#18181B] border border-[#26262A] hover:border-[#7B96F5] text-[#7B96F5] hover:text-[#95ABF7] text-[11px] font-mono transition-colors cursor-pointer"
                        >
                          Edit / Reset Password
                        </button>
                        <button
                          v-if="usr.id !== authStore.user?.id && usr.email !== authStore.user?.email"
                          @click="confirmDeleteUser(usr)"
                          class="px-2.5 py-1 rounded-lg bg-red-500/10 border border-red-500/30 hover:bg-red-500/20 text-red-400 hover:text-red-300 text-[11px] font-mono transition-colors flex items-center gap-1 cursor-pointer"
                          title="Delete User Account"
                        >
                          <Trash2 class="w-3 h-3" />
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <PaginationControl
              v-model:current-page="usersPage"
              v-model:page-size="usersPageSize"
              :total="usersTotal"
            />
          </div>
        </div>

        <!-- ══════════════════════════════════════════════════════════════════════════
             TAB 7: Audit Logs & RBAC Matrix
             ══════════════════════════════════════════════════════════════════════════ -->
        <div v-else-if="activeTab === 'audit'" class="space-y-6 animate-fadeIn">
          <!-- Role Permission Matrix for Admins -->
          <PermissionMatrix v-if="authStore.user.role === 'admin'" />

          <!-- User Activity & Session Logs -->
          <div class="bg-[#151517] border border-[#26262A] rounded-2xl p-5 space-y-4 shadow-xl">
            <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
              <div>
                <h3 class="text-xs font-bold uppercase tracking-wider text-gray-200 font-mono flex items-center gap-2">
                  <Activity class="w-4 h-4 text-[#7B96F5]" />
                  User Activity &amp; Audit Logs
                </h3>
                <p class="text-[11px] text-gray-500 mt-0.5">Real-time audit record of authentication, configuration changes, and sessions.</p>
              </div>
              <button
                @click="fetchUserLogs"
                class="px-2.5 py-1 rounded-lg bg-[#18181B] border border-[#26262A] hover:border-[#7B96F5] text-gray-300 text-xs font-mono flex items-center gap-1.5 cursor-pointer"
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
                  <tr v-if="isLogsLoading || settingStore.isLoading">
                    <td colspan="5" class="p-0 border-0"><SkeletonTable :rows="4" :cols="5" /></td>
                  </tr>
                  <tr v-else-if="userLogs.length === 0">
                    <td colspan="5" class="py-4 text-center text-gray-500 font-mono text-xs">No activity logs recorded yet</td>
                  </tr>
                  <tr v-else v-for="log in userLogs" :key="log.id" class="hover:bg-[#18181B]">
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
        </div>

        <!-- Fallback if unauthorized tab is somehow targeted -->
        <div v-else class="p-8 text-center bg-[#151517] border border-[#26262A] rounded-2xl space-y-3">
          <ShieldAlert class="w-8 h-8 text-amber-400 mx-auto" />
          <h3 class="text-sm font-bold text-white font-mono">Category Unavailable</h3>
          <p class="text-xs text-gray-400">You do not have permission to view this settings category.</p>
        </div>
      </main>
    </div>

    <!-- ══════════════════════════════════════════════════════════════════════════
         MODALS & NOTIFICATIONS
         ══════════════════════════════════════════════════════════════════════════ -->

    <!-- Save Success / Feedback Popup Modal -->
    <Modal :is-open="showFeedbackModal" :title="feedbackTitle" @close="showFeedbackModal = false">
      <template #default>
        <div class="space-y-3 text-xs font-mono">
          <div
            class="p-4 rounded-xl border flex items-start gap-3"
            :class="feedbackSuccess ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300' : 'bg-red-500/10 border-red-500/30 text-red-300'"
          >
            <Check v-if="feedbackSuccess" class="w-5 h-5 shrink-0 text-emerald-400 mt-0.5" />
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
          class="px-5 py-2 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 cursor-pointer"
        >
          OK
        </button>
      </template>
    </Modal>

    <!-- Location Modal (Add / Edit) -->
    <Modal :is-open="isLocationModalOpen" :title="isLocationEdit ? 'Edit Location' : 'Add New Location'" @close="isLocationModalOpen = false">
      <template #default>
        <form @submit.prevent="handleSaveLocation" class="space-y-4 text-xs">
          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-gray-400">Location Name *</label>
            <input
              v-model="locationForm.name"
              type="text"
              required
              placeholder="e.g. Gedung Sate - Lt 2"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-gray-400">Description</label>
            <input
              v-model="locationForm.description"
              type="text"
              placeholder="e.g. Server Room Main Switch Rack"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
        </form>
      </template>

      <template #footer>
        <button
          @click="isLocationModalOpen = false"
          class="px-4 py-2 rounded-xl border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs cursor-pointer"
        >
          Cancel
        </button>
        <button
          @click="handleSaveLocation"
          :disabled="isSavingLocation"
          class="px-5 py-2 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 disabled:opacity-50 cursor-pointer"
        >
          {{ isSavingLocation ? 'Saving...' : isLocationEdit ? 'Update Location' : 'Create Location' }}
        </button>
      </template>
    </Modal>

    <!-- Delete Location Modal -->
    <Modal :is-open="isDeleteLocationModalOpen" title="Confirm Delete Location" @close="isDeleteLocationModalOpen = false">
      <template #default>
        <div class="space-y-4 text-xs">
          <div v-if="locationToDelete?.deviceCount && locationToDelete.deviceCount > 0" class="p-3 bg-amber-500/10 border border-amber-500/30 rounded-xl text-amber-300 space-y-2">
            <div class="flex items-center gap-2 font-bold font-mono">
              <AlertTriangle class="w-4 h-4 text-amber-400 shrink-0" />
              <span>Deletion Blocked</span>
            </div>
            <p class="text-[11px] font-mono">
              <strong>{{ locationToDelete.deviceCount }} devices</strong> are currently assigned to location <strong class="text-white">{{ locationToDelete.name }}</strong>. Please reassign or delete those devices first before deleting this location.
            </p>
          </div>
          <div v-else class="p-3 bg-red-500/10 border border-red-500/30 rounded-xl flex items-start gap-3 text-red-400">
            <AlertTriangle class="w-5 h-5 shrink-0 mt-0.5" />
            <div>
              <h4 class="font-bold font-mono">Confirm Deletion</h4>
              <p class="mt-1 text-[11px] text-gray-300">
                Are you sure you want to delete location <strong class="text-white font-mono">{{ locationToDelete?.name }}</strong>?
              </p>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <button
          @click="isDeleteLocationModalOpen = false"
          class="px-4 py-2 rounded-xl border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-mono cursor-pointer"
        >
          Cancel
        </button>
        <button
          v-if="!locationToDelete?.deviceCount || locationToDelete.deviceCount === 0"
          @click="handleDeleteLocation"
          :disabled="isDeletingLocation"
          class="px-5 py-2 rounded-xl bg-red-500 hover:bg-red-600 text-white font-semibold text-xs shadow-md shadow-red-500/20 font-mono disabled:opacity-50 flex items-center gap-1.5 cursor-pointer"
        >
          <Trash2 class="w-3.5 h-3.5" />
          <span>{{ isDeletingLocation ? 'Deleting...' : 'Delete Location' }}</span>
        </button>
      </template>
    </Modal>

    <!-- Add User Modal with Real Email & OTP Verification -->
    <Modal :is-open="isAddUserModalOpen" :title="addUserStep === 1 ? 'Add New User (Real Email Required)' : 'Verify User Email OTP'" @close="isAddUserModalOpen = false">
      <template #default>
        <!-- Step Progress Indicator -->
        <div class="flex items-center gap-2 mb-4 p-2.5 rounded-xl bg-[#18181B] border border-[#26262A]">
          <div class="flex items-center gap-1.5 flex-1">
            <span
              class="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold font-mono"
              :class="addUserStep === 1 ? 'bg-[#7B96F5] text-white' : 'bg-[#3ECF8E] text-black'"
            >
              {{ addUserStep === 1 ? '1' : '✓' }}
            </span>
            <span class="text-[11px] font-mono" :class="addUserStep === 1 ? 'text-white font-bold' : 'text-gray-400'">
              1. User Details
            </span>
          </div>
          <div class="w-4 h-px bg-[#26262A]"></div>
          <div class="flex items-center gap-1.5 flex-1">
            <span
              class="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold font-mono"
              :class="addUserStep === 2 ? 'bg-[#7B96F5] text-white' : 'bg-[#26262A] text-gray-400'"
            >
              2
            </span>
            <span class="text-[11px] font-mono" :class="addUserStep === 2 ? 'text-white font-bold' : 'text-gray-500'">
              2. OTP Verification
            </span>
          </div>
        </div>

        <!-- STEP 1: User Information Form -->
        <div v-if="addUserStep === 1" class="space-y-3.5 text-xs">
          <div class="p-3 bg-[#7B96F5]/10 border border-[#7B96F5]/25 rounded-xl flex items-start gap-2.5 text-[#7B96F5]">
            <ShieldAlert class="w-4 h-4 shrink-0 mt-0.5" />
            <p class="text-[11px] leading-relaxed">
              Must use an active, <strong>real email address</strong> (e.g., <code>operator@jabarprov.go.id</code> or <code>user@gmail.com</code>). The system will send a 6-digit OTP code to verify the account.
            </p>
          </div>

          <div class="space-y-1">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Full Name *</label>
            <input
              v-model="addUserForm.name"
              type="text"
              required
              placeholder="e.g. Ahmad Hidayat"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 text-xs focus:outline-none focus:border-[#7B96F5]"
            />
          </div>

          <div class="space-y-1">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Real Email Address *</label>
            <input
              v-model="addUserForm.email"
              type="email"
              required
              placeholder="e.g. ahmad@diskominfo.jabarprov.go.id"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 text-xs focus:outline-none focus:border-[#7B96F5]"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Username</label>
              <input
                v-model="addUserForm.username"
                type="text"
                placeholder="Auto from email"
                class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 text-xs focus:outline-none focus:border-[#7B96F5]"
              />
            </div>

            <div class="space-y-1">
              <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Access Role *</label>
              <select
                v-model="addUserForm.role"
                class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 text-xs focus:outline-none focus:border-[#7B96F5]"
              >
                <option value="anggota">SANOC Member</option>
                <option value="pimpinan">Leadership</option>
                <option value="admin">Admin</option>
              </select>
            </div>
          </div>

          <div class="space-y-1">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold">Initial Password * (Min 8 Characters)</label>
            <input
              v-model="addUserForm.password"
              type="password"
              required
              placeholder="At least 8 characters"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-xl px-3 py-2 text-gray-200 text-xs focus:outline-none focus:border-[#7B96F5]"
            />
          </div>
        </div>

        <!-- STEP 2: OTP Verification -->
        <div v-else class="space-y-4 text-xs">
          <div class="p-3.5 bg-[#18181B] border border-[#26262A] rounded-xl space-y-2 text-center">
            <div class="w-10 h-10 rounded-full bg-[#7B96F5]/10 text-[#7B96F5] border border-[#7B96F5]/30 mx-auto flex items-center justify-center">
              <Send class="w-5 h-5" />
            </div>
            <h4 class="font-bold text-white text-sm">Verify User Email</h4>
            <p class="text-gray-400 text-xs leading-relaxed">
              A 6-digit verification code has been sent to:<br />
              <strong class="text-[#7B96F5] font-mono text-xs">{{ addUserForm.email }}</strong>
            </p>
          </div>

          <div class="space-y-3 py-1">
            <label class="block font-mono uppercase text-[10px] text-gray-400 font-semibold text-center">
              Enter 6-Digit Verification Code (1 Digit Per Box)
            </label>
            <OtpInput
              v-model="addUserOTP"
              :length="6"
              :auto-focus="true"
              @complete="handleConfirmAddUser"
            />
          </div>

          <div class="flex items-center justify-between text-xs text-gray-400 pt-1">
            <button
              type="button"
              @click="addUserStep = 1"
              class="text-gray-400 hover:text-white transition-colors cursor-pointer"
            >
              &larr; Change Details / Email
            </button>
            <button
              type="button"
              @click="handleSendAddUserOTP"
              :disabled="addUserCountdown > 0 || isSendingAddUserOTP"
              class="text-[#7B96F5] hover:text-[#95ABF7] font-semibold disabled:opacity-50 disabled:cursor-not-allowed transition-colors cursor-pointer"
            >
              <span v-if="addUserCountdown > 0">Resend in ({{ addUserCountdown }}s)</span>
              <span v-else>{{ isSendingAddUserOTP ? 'Sending...' : 'Resend OTP' }}</span>
            </button>
          </div>
        </div>
      </template>

      <template #footer>
        <button
          @click="isAddUserModalOpen = false"
          class="px-4 py-2 rounded-xl border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs cursor-pointer"
        >
          Cancel
        </button>

        <button
          v-if="addUserStep === 1"
          @click="handleSendAddUserOTP"
          :disabled="isSendingAddUserOTP"
          class="px-5 py-2 rounded-xl bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-md shadow-[#7B96F5]/20 disabled:opacity-50 flex items-center gap-1.5 cursor-pointer"
        >
          <Send class="w-3.5 h-3.5" :class="isSendingAddUserOTP ? 'animate-spin' : ''" />
          <span>{{ isSendingAddUserOTP ? 'Sending OTP...' : 'Next: Send Verification Code' }}</span>
        </button>

        <button
          v-else
          @click="handleConfirmAddUser"
          :disabled="isCreatingUser || !addUserOTP || addUserOTP.length !== 6"
          class="px-5 py-2 rounded-xl bg-[#3ECF8E] hover:bg-[#34B77B] text-black font-bold text-xs shadow-md shadow-[#3ECF8E]/20 disabled:opacity-50 flex items-center gap-1.5 cursor-pointer"
        >
          <Check class="w-3.5 h-3.5" :class="isCreatingUser ? 'animate-spin' : ''" />
          <span>{{ isCreatingUser ? 'Registering...' : 'Verify & Create Account' }}</span>
        </button>
      </template>
    </Modal>

    <!-- Delete User Modal -->
    <Modal :is-open="isDeleteModalOpen" title="Confirm Delete User" @close="isDeleteModalOpen = false">
      <template #default>
        <div class="space-y-4 text-xs">
          <div class="p-3 bg-red-500/10 border border-red-500/30 rounded-xl flex items-start gap-3 text-red-400">
            <AlertTriangle class="w-5 h-5 shrink-0 mt-0.5" />
            <div>
              <h4 class="font-bold font-mono">This Action Cannot Be Undone</h4>
              <p class="mt-1 text-[11px] text-gray-300">
                Are you sure you want to permanently delete user account <strong class="text-white font-mono">{{ userToDelete?.name }}</strong> ({{ userToDelete?.email }})?
              </p>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <button
          @click="isDeleteModalOpen = false"
          class="px-4 py-2 rounded-xl border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs font-mono cursor-pointer"
        >
          Cancel
        </button>
        <button
          @click="handleDeleteUser"
          :disabled="isDeletingUser"
          class="px-5 py-2 rounded-xl bg-red-500 hover:bg-red-600 text-white font-semibold text-xs shadow-md shadow-red-500/20 font-mono disabled:opacity-50 flex items-center gap-1.5 cursor-pointer"
        >
          <Trash2 class="w-3.5 h-3.5" />
          <span>{{ isDeletingUser ? 'Deleting...' : 'Delete User' }}</span>
        </button>
      </template>
    </Modal>

    <!-- Other Modals -->
    <WhatsAppTargetModal :is-open="isWATargetModalOpen" @close="isWATargetModalOpen = false" />
    <WhatsAppQRModal :is-open="isWAModalOpen" @close="isWAModalOpen = false" @connected="onWAConnected" />
    <UserEditModal :is-open="isUserEditModalOpen" :user="selectedUser" @close="isUserEditModalOpen = false" @saved="settingStore.fetchUsers()" />
    <TelegramConfigModal :is-open="isTGModalOpen" :bot-token="telegramToken" :chat-id="telegramChatId" @close="isTGModalOpen = false" @saved="onTGSaved" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useSettingStore } from '../stores/settingStore';
import { useAuthStore } from '../stores/authStore';
import Modal from '../components/common/Modal.vue';
import OtpInput from '../components/common/OtpInput.vue';
import Skeleton from '../components/common/Skeleton.vue';
import SkeletonTable from '../components/common/SkeletonTable.vue';
import PaginationControl from '../components/common/PaginationControl.vue';
import WhatsAppQRModal from '../components/settings/WhatsAppQRModal.vue';
import TelegramConfigModal from '../components/settings/TelegramConfigModal.vue';
import WhatsAppTargetModal from '../components/settings/WhatsAppTargetModal.vue';
import UserEditModal from '../components/users/UserEditModal.vue';
import PermissionMatrix from '../components/settings/PermissionMatrix.vue';
import type { User, UserRole, LocationItem } from '../types';
import {
  Send,
  MessageSquare,
  QrCode,
  Activity,
  Sliders,
  Users,
  UserPlus,
  RefreshCw,
  Target,
  LogOut,
  Trash2,
  AlertTriangle,
  AlertCircle,
  Network,
  Archive,
  MapPin,
  Plus,
  Save,
  Check,
  ShieldAlert
} from 'lucide-vue-next';
import api from '../api/client';
import { locationsApi, authApi, usersApi } from '../api';
import { wsClient } from '../ws/websocket';

const route = useRoute();
const router = useRouter();
const settingStore = useSettingStore();
const authStore = useAuthStore();

// Core UI States
const isInitialLoaded = ref(false);
const activeTab = ref<string>('notifications');

// Save Feedback Modal States
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

// Notifications state
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
const isSavingRateLimit = ref(false);
const isSavingEngine = ref(false);
const isSavingCoreSwitch = ref(false);
const isSavingRetention = ref(false);

// Modals state
const isAddUserModalOpen = ref(false);
const isWATargetModalOpen = ref(false);
const isWAModalOpen = ref(false);
const isTGModalOpen = ref(false);
const isUserEditModalOpen = ref(false);
const selectedUser = ref<User | null>(null);

const isDeleteModalOpen = ref(false);
const userToDelete = ref<User | null>(null);
const isDeletingUser = ref(false);

// Add User & OTP states
const addUserStep = ref<1 | 2>(1);
const addUserForm = reactive({
  username: '',
  name: '',
  email: '',
  role: 'anggota' as UserRole,
  password: ''
});
const addUserOTP = ref('');
const isSendingAddUserOTP = ref(false);
const isCreatingUser = ref(false);
const addUserCountdown = ref(0);
let addUserTimer: any = null;

function startAddUserCountdown() {
  addUserCountdown.value = 60;
  if (addUserTimer) clearInterval(addUserTimer);
  addUserTimer = setInterval(() => {
    if (addUserCountdown.value > 0) {
      addUserCountdown.value--;
    } else {
      clearInterval(addUserTimer);
      addUserTimer = null;
    }
  }, 1000);
}

function openAddUserModal() {
  addUserStep.value = 1;
  addUserForm.username = '';
  addUserForm.name = '';
  addUserForm.email = '';
  addUserForm.role = 'anggota';
  addUserForm.password = '';
  addUserOTP.value = '';
  isAddUserModalOpen.value = true;
}

function validateRealEmail(email: string): boolean {
  const trimmed = email.trim().toLowerCase();
  const re = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
  if (!re.test(trimmed)) return false;
  const parts = trimmed.split('@');
  if (parts.length !== 2) return false;
  const domain = parts[1];
  const blocked = ['test.com', 'example.com', 'test.test', 'local.com', 'localhost', 'dummy.com', 'temp.com', 'fake.com'];
  return !blocked.some(b => domain === b || domain.endsWith('.' + b));
}

async function handleSendAddUserOTP() {
  if (!addUserForm.name || !addUserForm.email || !addUserForm.password) {
    triggerFeedback('Form Belum Lengkap', 'Nama, email, dan kata sandi wajib diisi.', false);
    return;
  }
  if (!validateRealEmail(addUserForm.email)) {
    triggerFeedback('Email Tidak Valid', 'Wajib menggunakan alamat email real yang valid (contoh: operator@jabarprov.go.id atau user@gmail.com).', false);
    return;
  }
  if (addUserForm.password.length < 8) {
    triggerFeedback('Password Kurang Kuat', 'Kata sandi minimal 8 karakter.', false);
    return;
  }

  isSendingAddUserOTP.value = true;
  try {
    const res = await authApi.sendVerificationOTP(addUserForm.email.trim().toLowerCase());
    addUserStep.value = 2;
    startAddUserCountdown();
    triggerFeedback('Kode OTP Terkirim', res.message || `Kode verifikasi 6-digit telah dikirim ke ${addUserForm.email}`, true);
  } catch (e: any) {
    triggerFeedback('Gagal Mengirim OTP', e.response?.data?.error || 'Terjadi kesalahan sistem saat mengirim kode OTP.', false);
  } finally {
    isSendingAddUserOTP.value = false;
  }
}

async function handleConfirmAddUser() {
  if (!addUserOTP.value || addUserOTP.value.trim().length !== 6) {
    triggerFeedback('Kode OTP Tidak Lengkap', 'Silakan masukkan 6 digit kode OTP verifikasi email.', false);
    return;
  }

  isCreatingUser.value = true;
  try {
    await usersApi.createUser({
      username: addUserForm.username.trim() || addUserForm.email.split('@')[0],
      name: addUserForm.name.trim(),
      email: addUserForm.email.trim().toLowerCase(),
      role: addUserForm.role,
      password: addUserForm.password,
      verificationCode: addUserOTP.value.trim()
    });

    await settingStore.fetchUsers();
    isAddUserModalOpen.value = false;
    if (addUserTimer) clearInterval(addUserTimer);
    triggerFeedback('Akun Berhasil Dibuat', `Pengguna ${addUserForm.name} (${addUserForm.email}) berhasil didaftarkan dan terverifikasi.`, true);
  } catch (e: any) {
    triggerFeedback('Pendaftaran Gagal', e.response?.data?.error || 'Gagal memverifikasi OTP atau membuat akun.', false);
  } finally {
    isCreatingUser.value = false;
  }
}

const currentRoleName = computed(() => {
  const r = authStore.user?.role || 'anggota';
  if (r === 'admin') return 'ADMIN';
  if (r === 'pimpinan') return 'PIMPINAN';
  return 'ANGGOTA SANOC';
});

// Category Tabs Definition with Permissions (Strict 1-to-1 mapping with Role Permissions)
const allTabs = computed(() => [
  {
    id: 'notifications',
    label: 'Notification Gateway',
    description: 'WhatsApp, Telegram & Rate Limits',
    icon: Send,
    badge: isWhatsAppConnected.value ? 'WA Connected' : 'WA Offline',
    badgeClass: isWhatsAppConnected.value ? 'bg-[#3ECF8E]/15 text-[#3ECF8E]' : 'bg-amber-500/15 text-amber-400',
    permission: () => authStore.user?.role === 'admin' || authStore.hasPermission('settings.notifications')
  },
  {
    id: 'polling',
    label: 'Engine & Thresholds',
    description: 'ICMP interval, flapping & debounce',
    icon: Activity,
    badge: undefined,
    badgeClass: undefined,
    permission: () => authStore.user?.role === 'admin' || authStore.hasPermission('settings.polling')
  },
  {
    id: 'network',
    label: 'Core Switch & SNMP',
    description: 'Cross-subnet L3 ARP & router target',
    icon: Network,
    badge: undefined,
    badgeClass: undefined,
    permission: () => authStore.user?.role === 'admin' || authStore.hasPermission('settings.network')
  },
  {
    id: 'retention',
    label: 'Retention & Cleanup',
    description: 'Incident retention & DB cleanup',
    icon: Archive,
    badge: undefined,
    badgeClass: undefined,
    permission: () => authStore.user?.role === 'admin' || authStore.hasPermission('settings.retention')
  },
  {
    id: 'locations',
    label: 'Location Management',
    description: 'Manage node locations & racks',
    icon: MapPin,
    badge: undefined,
    badgeClass: undefined,
    permission: () => authStore.user?.role === 'admin' || authStore.hasPermission('settings.locations')
  },
  {
    id: 'users',
    label: 'Users & Roles',
    description: 'Credentials, accounts & privileges',
    icon: Users,
    badge: undefined,
    badgeClass: undefined,
    permission: () => authStore.user?.role === 'admin' || authStore.hasPermission('settings.users')
  },
  {
    id: 'audit',
    label: 'System Audit Log',
    description: 'Activity audit & permission matrix',
    icon: Activity,
    badge: undefined,
    badgeClass: undefined,
    permission: () => authStore.user?.role === 'admin' || authStore.hasPermission('settings.audit')
  }
]);

const availableTabs = computed(() => {
  return allTabs.value.filter((t) => t.permission());
});

function switchTab(tabId: string) {
  activeTab.value = tabId;
  router.replace({ query: { ...route.query, tab: tabId } });
}

// Sync active tab with route query and permissions
watch(
  [availableTabs, () => route.query.tab],
  ([tabs, queryTab]) => {
    if (!tabs || tabs.length === 0) {
      activeTab.value = '';
      return;
    }
    const requested = (queryTab as string) || activeTab.value;
    const match = tabs.find((t) => t.id === requested);
    if (match) {
      activeTab.value = match.id;
    } else {
      activeTab.value = tabs[0].id;
    }
  },
  { immediate: true }
);

function openEditUser(usr: User) {
  selectedUser.value = usr;
  isUserEditModalOpen.value = true;
}

function confirmDeleteUser(usr: User) {
  userToDelete.value = usr;
  isDeleteModalOpen.value = true;
}

async function handleDeleteUser() {
  if (!userToDelete.value) return;
  isDeletingUser.value = true;
  try {
    await settingStore.deleteUser(userToDelete.value.id);
    await settingStore.fetchUsers();
    isDeleteModalOpen.value = false;
    userToDelete.value = null;
    triggerFeedback('User Account Deleted', 'Akun pengguna berhasil dihapus dari sistem.', true);
  } catch (e: any) {
    triggerFeedback('Gagal Menghapus Pengguna', e.response?.data?.error || 'Terjadi kesalahan sistem.', false);
  } finally {
    isDeletingUser.value = false;
  }
}

// Per-category Save Handlers
async function handleSaveRateLimit() {
  isSavingRateLimit.value = true;
  try {
    await settingStore.saveSettings();
    triggerFeedback('Queue Settings Saved', `Asynq Redis rate-limit updated: ${settingStore.settings.rateLimitMaxMsgPerMin} messages/minute.`);
  } catch (e: any) {
    triggerFeedback('Save Failed', e.response?.data?.error || 'Failed to save rate limit settings', false);
  } finally {
    isSavingRateLimit.value = false;
  }
}

async function handleSaveEngineAndThresholds() {
  isSavingEngine.value = true;
  try {
    await Promise.all([
      settingStore.saveSettings(),
      api.put('/settings/thresholds', {
        thresholds: settingStore.settings.thresholds
      })
    ]);
    triggerFeedback(
      'Engine & Thresholds Saved',
      `Polling interval (${settingStore.settings.polling.intervalSeconds}s), concurrency (${settingStore.settings.polling.concurrencyBatchSize}), and device failure thresholds updated successfully!`
    );
  } catch (e: any) {
    triggerFeedback('Save Failed', e.response?.data?.error || 'Failed to save polling engine and thresholds', false);
  } finally {
    isSavingEngine.value = false;
  }
}

async function handleSaveCoreSwitch() {
  isSavingCoreSwitch.value = true;
  try {
    await settingStore.saveSettings();
    triggerFeedback('Core Switch SNMP Saved', `Target SNMP Core Switch (${settingStore.settings.coreSwitch?.ip || 'Configured'}) saved for cross-subnet L3 ARP resolution.`);
  } catch (e: any) {
    triggerFeedback('Save Failed', e.response?.data?.error || 'Failed to save Core Switch SNMP settings', false);
  } finally {
    isSavingCoreSwitch.value = false;
  }
}

async function handleSaveRetention() {
  isSavingRetention.value = true;
  try {
    await settingStore.saveSettings();
    triggerFeedback('Retention Policy Saved', `Auto-archive policy set to ${settingStore.settings.retentionDays} days.`);
  } catch (e: any) {
    triggerFeedback('Save Failed', e.response?.data?.error || 'Failed to save retention policy', false);
  } finally {
    isSavingRetention.value = false;
  }
}

async function handleManualRefresh() {
  isRefreshing.value = true;
  try {
    await api.post('/monitoring/refresh-now');
    triggerFeedback('Polling Cycle Dispatched', 'ICMP probe cycle was immediately triggered across all registered devices.');
  } catch (e: any) {
    triggerFeedback('Poll Failed', e.response?.data?.error || 'Failed to trigger poll', false);
  } finally {
    setTimeout(() => {
      isRefreshing.value = false;
    }, 600);
  }
}

const usersPage = ref(1);
const usersPageSize = ref(5);
const usersTotal = computed(() => settingStore.users.length);
const paginatedUsers = computed(() => {
  const start = (usersPage.value - 1) * usersPageSize.value;
  return settingStore.users.slice(start, start + usersPageSize.value);
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

function onWAConnected(num: string) {
  isWhatsAppConnected.value = true;
  whatsAppNumber.value = num || '+62 812-9000-8888';
  triggerFeedback('WhatsApp Connected', 'WhatsApp API gateway connected successfully!');
}

async function handleWADisconnect() {
  try {
    await api.post('/integrations/whatsapp/disconnect');
  } catch (e) {
    // fallback
  }
  isWhatsAppConnected.value = false;
  triggerFeedback('WhatsApp Disconnected', 'WhatsApp gateway session has been terminated.');
}

async function handleWATest() {
  isWATesting.value = true;
  waTestMessage.value = '';
  try {
    await api.post('/integrations/whatsapp/test');
    waTestSuccess.value = true;
    waTestMessage.value = '✓ Test notification delivered to configured targets';
    triggerFeedback('WhatsApp Test Sent', 'Test notification was delivered to configured on-call recipients.');
  } catch (e: any) {
    waTestSuccess.value = false;
    waTestMessage.value = '✗ Test failed: ' + (e.response?.data?.error || 'Gateway offline');
    triggerFeedback('WhatsApp Test Failed', e.response?.data?.error || 'Gateway offline', false);
  } finally {
    isWATesting.value = false;
  }
}

function onTGSaved(cfg: { botToken: string; chatId: string }) {
  telegramToken.value = cfg.botToken;
  telegramChatId.value = cfg.chatId;
  triggerFeedback('Telegram Bot Configured', `Telegram channel target saved: ${cfg.chatId}`);
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
    triggerFeedback('Telegram Test Sent', 'Test broadcast was delivered to Telegram Channel ' + telegramChatId.value);
  } catch (e: any) {
    tgTestSuccess.value = false;
    tgTestMessage.value = '✗ Test failed: ' + (e.response?.data?.error || 'API Error');
    triggerFeedback('Telegram Test Failed', e.response?.data?.error || 'API Error', false);
  } finally {
    isTGTesting.value = false;
  }
}

// Location Management State
const locationsList = ref<LocationItem[]>([]);
const isLocationsLoading = ref(false);
const isLocationModalOpen = ref(false);
const isLocationEdit = ref(false);
const editingLocationId = ref('');
const isSavingLocation = ref(false);
const locationForm = reactive({ name: '', description: '' });

const isDeleteLocationModalOpen = ref(false);
const locationToDelete = ref<LocationItem | null>(null);
const isDeletingLocation = ref(false);

async function fetchLocations() {
  isLocationsLoading.value = true;
  try {
    const res = await locationsApi.getLocations();
    if (res) locationsList.value = res;
  } catch (e) {
    // fallback
  } finally {
    isLocationsLoading.value = false;
  }
}

function openAddLocationModal() {
  isLocationEdit.value = false;
  editingLocationId.value = '';
  locationForm.name = '';
  locationForm.description = '';
  isLocationModalOpen.value = true;
}

function openEditLocationModal(loc: LocationItem) {
  isLocationEdit.value = true;
  editingLocationId.value = loc.id;
  locationForm.name = loc.name;
  locationForm.description = loc.description || '';
  isLocationModalOpen.value = true;
}

async function handleSaveLocation() {
  if (!locationForm.name.trim()) return;
  isSavingLocation.value = true;
  try {
    if (isLocationEdit.value && editingLocationId.value) {
      await locationsApi.updateLocation(editingLocationId.value, locationForm.name, locationForm.description);
      triggerFeedback('Location Updated', `Lokasi "${locationForm.name}" berhasil diperbarui.`);
    } else {
      await locationsApi.createLocation(locationForm.name, locationForm.description);
      triggerFeedback('Location Created', `Lokasi "${locationForm.name}" berhasil ditambahkan.`);
    }
    await fetchLocations();
    isLocationModalOpen.value = false;
  } catch (e: any) {
    triggerFeedback('Location Error', e.response?.data?.error || 'Failed to save location', false);
  } finally {
    isSavingLocation.value = false;
  }
}

function confirmDeleteLocation(loc: LocationItem) {
  locationToDelete.value = loc;
  isDeleteLocationModalOpen.value = true;
}

async function handleDeleteLocation() {
  if (!locationToDelete.value) return;
  isDeletingLocation.value = true;
  try {
    await locationsApi.deleteLocation(locationToDelete.value.id);
    await fetchLocations();
    isDeleteLocationModalOpen.value = false;
    triggerFeedback('Location Deleted', `Lokasi "${locationToDelete.value.name}" berhasil dihapus.`);
  } catch (e: any) {
    triggerFeedback('Location Delete Failed', e.response?.data?.error || 'Failed to delete location', false);
  } finally {
    isDeletingLocation.value = false;
  }
}

onMounted(async () => {
  try {
    await Promise.allSettled([
      settingStore.fetchSettings(),
      settingStore.fetchUsers(),
      fetchIntegrationsConfig(),
      fetchUserLogs(),
      fetchLocations()
    ]);
  } catch (e) {
    // fallback
  } finally {
    isInitialLoaded.value = true;
    settingStore.isLoading = false;
  }

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
</script>
