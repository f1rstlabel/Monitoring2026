<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-subtle pb-5">
      <div>
        <div class="flex items-center gap-2">
          <h1 class="text-xl font-extrabold text-text-main tracking-tight">System Settings & Administration</h1>
          <span class="px-2 py-0.5 rounded-full text-[10px] font-mono font-bold uppercase bg-brand-periwinkle/10 text-brand-periwinkle border border-brand-periwinkle/25">
            {{ currentRoleName }}
          </span>
        </div>
        <p class="text-xs text-text-secondary mt-1">
          Configure notification gateways, polling engine parameters, data retention, and user role management
        </p>
      </div>
    </div>

    <!-- Skeleton Loading State -->
    <div v-if="settingStore.isLoading && !isInitialLoaded" class="grid grid-cols-1 lg:grid-cols-4 gap-6">
      <div class="lg:col-span-1 bg-surface border border-subtle rounded-2xl p-4 space-y-3">
        <Skeleton width="40%" height="1rem" />
        <div v-for="i in 6" :key="i" class="p-3 bg-card border border-subtle rounded-xl space-y-2">
          <Skeleton width="60%" height="0.85rem" />
          <Skeleton width="80%" height="0.65rem" />
        </div>
      </div>
      <div class="lg:col-span-3 bg-surface border border-subtle rounded-2xl p-6 space-y-6">
        <Skeleton width="30%" height="1.5rem" />
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div v-for="i in 4" :key="i" class="p-5 bg-card border border-subtle rounded-xl space-y-3">
            <Skeleton width="50%" height="1rem" />
            <Skeleton width="100%" height="2rem" />
          </div>
        </div>
      </div>
    </div>

    <!-- Access Restricted (403) Screen when user has 0 settings permissions -->
    <div v-else-if="availableTabs.length === 0" class="p-12 text-center bg-surface border border-subtle rounded-2xl space-y-4 max-w-lg mx-auto shadow-2xl animate-fadeIn">
      <div class="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-red-500/10 border border-red-500/30 text-red-400">
        <ShieldAlert class="w-7 h-7" />
      </div>
      <h2 class="text-base font-bold text-text-main font-mono">Access Restricted (403)</h2>
      <p class="text-xs text-text-secondary leading-relaxed font-mono">
        Your current role (<strong class="text-brand-periwinkle">{{ currentRoleName }}</strong>) does not have permission to view or modify any system configuration categories.
      </p>
      <div class="pt-3">
        <router-link
          to="/dashboard"
          class="px-5 py-2.5 rounded-xl bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-semibold text-xs font-mono inline-flex items-center gap-2 shadow-lg shadow-brand-periwinkle/20 cursor-pointer"
        >
          Back to Dashboard
        </router-link>
      </div>
    </div>

    <!-- Main Sub-Sidebar & Content Layout -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-4 gap-6 items-start">
      <!-- Sub-Sidebar Navigation (Left Column) -->
      <aside class="lg:col-span-1 bg-surface border border-subtle rounded-2xl p-3.5 space-y-1.5 shadow-xl sticky top-4">
        <div class="px-3 py-2 border-b border-subtle/60 mb-2">
          <span class="text-[10px] font-mono font-bold uppercase tracking-wider text-text-secondary">Settings Categories</span>
        </div>

        <nav class="space-y-1" aria-label="Settings Categories">
          <button
            v-for="tab in availableTabs"
            :key="tab.id"
            @click="switchTab(tab.id)"
            class="w-full text-left p-3 rounded-xl transition-all flex items-start gap-3 group relative cursor-pointer"
            :class="[
              activeTab === tab.id
                ? 'bg-brand-periwinkle/10 border border-brand-periwinkle/30 text-brand-periwinkle shadow-sm shadow-brand-periwinkle/10'
                : 'border border-transparent text-text-secondary hover:text-text-main hover:bg-card hover:border-subtle'
            ]"
          >
            <component
              :is="tab.icon"
              class="w-4 h-4 shrink-0 mt-0.5 transition-colors"
              :class="activeTab === tab.id ? 'text-brand-periwinkle' : 'text-text-secondary group-hover:text-text-main'"
            />
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-1.5 min-w-0">
                  <span class="text-xs font-bold font-mono tracking-tight truncate" :class="activeTab === tab.id ? 'text-text-main' : 'text-text-secondary'">
                    {{ tab.label }}
                  </span>
                  <span
                    v-if="isTabDirty(tab.id)"
                    class="w-2 h-2 rounded-full bg-amber-400 shrink-0 shadow-sm shadow-amber-400/50 animate-pulse"
                    title="Perubahan belum disimpan"
                  ></span>
                </div>
                <span
                  v-if="tab.badge"
                  class="text-[9px] font-mono px-1.5 py-0.5 rounded font-bold uppercase"
                  :class="tab.badgeClass || 'bg-subtle text-text-secondary'"
                >
                  {{ tab.badge }}
                </span>
              </div>
              <p class="text-[10px] font-sans text-text-secondary truncate mt-0.5">{{ tab.description }}</p>
            </div>

            <!-- Active Indicator Dot -->
            <span
              v-if="activeTab === tab.id"
              class="w-1.5 h-1.5 rounded-full bg-brand-periwinkle absolute right-2.5 top-1/2 -translate-y-1/2 shadow-sm shadow-brand-periwinkle"
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
          <div class="flex items-center justify-between border-b border-subtle pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-text-main font-mono flex items-center gap-2">
                <Send class="w-4 h-4 text-brand-periwinkle" />
                NOTIFICATION GATEWAYS &amp; RATE LIMITS
              </h2>
              <p class="text-xs text-text-secondary mt-0.5">Configure primary and fallback alert channels and queue rate-limiting.</p>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- WhatsApp API Card -->
            <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-4 shadow-xl">
              <div class="flex items-start justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-xl bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
                    <MessageSquare class="w-5 h-5" />
                  </div>
                  <div>
                    <h3 class="text-sm font-bold text-text-main">WhatsApp API Gateway</h3>
                    <p class="text-xs font-mono text-text-secondary mt-0.5">{{ whatsAppNumber || 'Gateway Configured' }}</p>
                  </div>
                </div>
                <span
                  class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold border"
                  :class="isWhatsAppConnected ? 'bg-status-up/15 text-status-up border-status-up/30' : 'bg-amber-500/15 text-amber-400 border-amber-500/30'"
                >
                  <span class="w-1.5 h-1.5 rounded-full" :class="isWhatsAppConnected ? 'bg-status-up pulsing-dot-green' : 'bg-amber-400'"></span>
                  {{ isWhatsAppConnected ? 'CONNECTED' : 'DISCONNECTED' }}
                </span>
              </div>

              <div v-if="waTestMessage" class="p-2.5 rounded-lg bg-card border border-subtle text-xs font-mono" :class="waTestSuccess ? 'text-status-up' : 'text-status-down'">
                {{ waTestMessage }}
              </div>

              <div class="pt-3 border-t border-subtle space-y-2">
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <button
                    @click="handleWATest"
                    :disabled="!isWhatsAppConnected || isWATesting"
                    class="h-9 px-3 rounded-lg border border-subtle bg-card hover:bg-subtle text-text-main text-xs font-semibold flex items-center justify-center gap-1.5 transition-all disabled:opacity-40 whitespace-nowrap cursor-pointer"
                  >
                    <Send class="w-3.5 h-3.5 text-brand-periwinkle" />
                    <span>{{ isWATesting ? 'Sending...' : 'Send Test Notification' }}</span>
                  </button>
                  <button
                    @click="isWATargetModalOpen = true"
                    class="h-9 px-3 rounded-lg bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white text-xs font-semibold flex items-center justify-center gap-1.5 shadow-sm shadow-brand-periwinkle/20 transition-all whitespace-nowrap cursor-pointer"
                  >
                    <Target class="w-3.5 h-3.5" />
                    <span>Configure Targets</span>
                  </button>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <button
                    @click="isWAModalOpen = true"
                    class="h-9 px-3 rounded-lg bg-card border border-subtle hover:bg-subtle text-text-main text-xs font-semibold flex items-center justify-center gap-1.5 transition-all whitespace-nowrap cursor-pointer"
                  >
                    <QrCode class="w-3.5 h-3.5 text-text-secondary" />
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
            <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-4 shadow-xl">
              <div class="flex items-start justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-xl bg-sky-500/15 border border-sky-500/30 flex items-center justify-center text-sky-400">
                    <Send class="w-5 h-5" />
                  </div>
                  <div>
                    <h3 class="text-sm font-bold text-text-main">Telegram Bot Gateway</h3>
                    <p class="text-xs font-mono text-sky-400 mt-0.5">{{ telegramHandle }}</p>
                  </div>
                </div>
                <span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold bg-status-up/15 text-status-up border border-status-up/30">
                  <span class="w-1.5 h-1.5 rounded-full bg-status-up"></span>
                  ACTIVE FALLBACK
                </span>
              </div>

              <div v-if="tgTestMessage" class="p-2.5 rounded-lg bg-card border border-subtle text-xs font-mono" :class="tgTestSuccess ? 'text-status-up' : 'text-status-down'">
                {{ tgTestMessage }}
              </div>

              <div class="pt-3 border-t border-subtle">
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <button
                    @click="handleTGTest"
                    :disabled="isTGTesting"
                    class="h-9 px-3 rounded-lg border border-subtle bg-card hover:bg-subtle text-text-main text-xs font-semibold flex items-center justify-center gap-1.5 transition-all disabled:opacity-40 whitespace-nowrap cursor-pointer"
                  >
                    <Send class="w-3.5 h-3.5 text-sky-400" />
                    <span>{{ isTGTesting ? 'Sending...' : 'Send Test Notification' }}</span>
                  </button>
                  <button
                    @click="isTGModalOpen = true"
                    class="h-9 px-3 rounded-lg bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white text-xs font-semibold flex items-center justify-center gap-1.5 shadow-sm shadow-brand-periwinkle/20 transition-all whitespace-nowrap cursor-pointer"
                  >
                    <Sliders class="w-3.5 h-3.5" />
                    <span>Configure Channel</span>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Rate Limit Card -->
          <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-4 shadow-xl">
            <div class="flex items-center justify-between border-b border-subtle pb-3">
              <h3 class="text-xs font-bold uppercase tracking-wider text-text-main font-mono flex items-center gap-2">
                <Sliders class="w-4 h-4 text-brand-periwinkle" />
                Asynq Redis Queue Rate-Limit Spacing
              </h3>
              <div class="flex items-center gap-2">
                <button
                  v-if="isTabDirty('notifications')"
                  type="button"
                  @click="revertTabState('notifications')"
                  class="px-3 py-1.5 rounded-xl border border-subtle text-text-secondary hover:text-text-main text-xs font-medium transition-colors cursor-pointer"
                >
                  Batal / Reset
                </button>
                <button
                  type="button"
                  @click="handleSaveRateLimit"
                  :disabled="isSavingRateLimit"
                  class="px-3.5 py-1.5 rounded-xl bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white text-xs font-semibold flex items-center gap-1.5 shadow-md shadow-brand-periwinkle/20 transition-all disabled:opacity-50 cursor-pointer"
                >
                  <Save class="w-3.5 h-3.5" />
                  <span>{{ isSavingRateLimit ? 'Saving...' : 'Save Queue Settings' }}</span>
                </button>
              </div>
            </div>

            <div class="space-y-3 text-xs max-w-md">
              <div class="space-y-1.5">
                <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">
                  Message Delay (Seconds)
                </label>
                <div class="flex items-center gap-2">
                  <input
                    type="number"
                    min="1"
                    max="300"
                    v-model.number="settingStore.settings.rateLimitMaxMsgPerMin"
                    class="w-32 bg-card border border-subtle rounded-xl px-3 py-2 text-text-main font-mono text-xs focus:outline-none focus:border-brand-periwinkle"
                  />
                  <span class="text-xs font-mono text-text-secondary">seconds / message</span>
                </div>
                <p class="text-[10px] text-text-muted font-mono">
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
          <div class="flex items-center justify-between border-b border-subtle pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-text-main font-mono flex items-center gap-2">
                <Activity class="w-4 h-4 text-brand-periwinkle" />
                ENGINE POLLING &amp; FAILURE THRESHOLDS
              </h2>
              <p class="text-xs text-text-secondary mt-0.5">Control ICMP probe frequencies, concurrency workers, debouncing, and anti-flap reuse.</p>
            </div>
            <div class="flex items-center gap-2">
              <button
                v-if="isTabDirty('polling')"
                type="button"
                @click="revertTabState('polling')"
                class="px-3 py-1.5 rounded-xl border border-subtle text-text-secondary hover:text-text-main text-xs font-medium transition-colors cursor-pointer"
              >
                Batal / Reset
              </button>
              <button
                type="button"
                @click="handleManualRefresh"
                :disabled="isRefreshing"
                class="px-3 py-1.5 rounded-xl bg-card border border-subtle hover:border-brand-periwinkle text-brand-periwinkle hover:text-brand-periwinkle-hover text-xs font-mono font-semibold transition-all flex items-center gap-1.5 disabled:opacity-50 cursor-pointer"
                title="Force immediate ICMP poll cycle"
              >
                <RefreshCw class="w-3.5 h-3.5" :class="isRefreshing ? 'animate-spin' : ''" />
                <span>{{ isRefreshing ? 'Polling...' : 'Trigger Poll Now' }}</span>
              </button>
              <button
                type="button"
                @click="handleSaveEngineAndThresholds"
                :disabled="isSavingEngine"
                class="px-3.5 py-1.5 rounded-xl bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white text-xs font-semibold flex items-center gap-1.5 shadow-md shadow-brand-periwinkle/20 transition-all disabled:opacity-50 cursor-pointer"
              >
                <Save class="w-3.5 h-3.5" />
                <span>{{ isSavingEngine ? 'Saving...' : 'Save Engine & Thresholds' }}</span>
              </button>
            </div>
          </div>

          <!-- Engine Polling Settings Card -->
          <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-5 shadow-xl">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6 text-xs">
              <!-- Polling Interval (Slider 1s–60s) -->
              <div class="space-y-2">
                <div class="flex items-center justify-between font-mono">
                  <label class="uppercase text-[10px] text-text-secondary font-semibold">ICMP Polling Interval</label>
                  <span class="text-brand-periwinkle font-bold text-sm">{{ settingStore.settings.polling.intervalSeconds }}s</span>
                </div>
                <input
                  type="range"
                  min="1"
                  max="60"
                  v-model.number="settingStore.settings.polling.intervalSeconds"
                  class="w-full h-1.5 bg-card border border-subtle rounded-lg appearance-none cursor-pointer"
                />
                <p class="text-[10px] text-text-muted font-mono">Frequency of ICMP ping probes executed across device inventory</p>
              </div>

              <!-- Concurrency Batch Size -->
              <div class="space-y-2">
                <div class="flex items-center justify-between font-mono">
                  <label class="uppercase text-[10px] text-text-secondary font-semibold">Concurrency Batch Size</label>
                  <span class="text-text-main font-bold font-mono">{{ settingStore.settings.polling.concurrencyBatchSize }} workers</span>
                </div>
                <input
                  type="number"
                  min="5"
                  max="200"
                  v-model.number="settingStore.settings.polling.concurrencyBatchSize"
                  class="w-full bg-card border border-subtle rounded-xl px-3 py-2 text-text-main font-mono text-xs focus:outline-none focus:border-brand-periwinkle"
                />
                <p class="text-[10px] text-text-muted font-mono">Parallel probe workers dispatched per cycle batch</p>
              </div>

              <!-- Flapping Reuse Window -->
              <div class="space-y-2 border-t border-subtle pt-4 md:col-span-2">
                <div class="flex items-center justify-between font-mono">
                  <label class="uppercase text-[10px] text-text-secondary font-semibold">Flap Detection Reuse Window</label>
                  <span class="text-amber-400 font-bold text-sm">
                    {{ settingStore.settings.polling.flapReuseWindowMinutes || 10 }} minutes
                  </span>
                </div>
                <input
                  type="range"
                  min="1"
                  max="60"
                  v-model.number="settingStore.settings.polling.flapReuseWindowMinutes"
                  class="w-full h-1.5 bg-card border border-subtle rounded-lg appearance-none cursor-pointer"
                />
                <p class="text-[10px] text-text-muted font-mono">
                  Window duration in minutes to reopen existing incident tickets for flapping devices rather than creating duplicate incidents.
                </p>
              </div>
            </div>
          </div>

          <!-- Device Type Failure Threshold Defaults Table Card -->
          <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-4 shadow-xl">
            <div class="flex items-center justify-between border-b border-subtle pb-3">
              <div>
                <h3 class="text-xs font-bold uppercase tracking-wider text-text-main font-mono flex items-center gap-2">
                  <Sliders class="w-4 h-4 text-brand-periwinkle" />
                  Device Category Failure Thresholds
                </h3>
                <p class="text-[11px] text-text-muted mt-0.5">
                  Consecutive ICMP check confirmations required before flipping state (DOWN &harr; UP)
                </p>
              </div>
            </div>

            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs text-text-secondary">
                <thead class="bg-card font-mono text-[10px] uppercase text-text-secondary">
                  <tr>
                    <th class="py-2.5 px-4">Device Category</th>
                    <th class="py-2.5 px-4">Consecutive ICMP Checks</th>
                    <th class="py-2.5 px-4">Effective Debounce Duration</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-subtle">
                  <tr v-for="t in settingStore.settings.thresholds" :key="t.type" class="hover:bg-card">
                    <td class="py-2.5 px-4 font-bold text-text-main font-mono">{{ t.type }}</td>
                    <td class="py-2.5 px-4">
                      <div class="flex items-center gap-2">
                        <input
                          type="number"
                          min="1"
                          max="10"
                          v-model.number="t.consecutiveFailures"
                          class="w-20 bg-card border border-subtle rounded-lg px-3 py-1.5 text-text-main font-mono text-xs focus:outline-none focus:border-brand-periwinkle"
                        />
                        <span class="text-[11px] font-mono text-text-muted">checks</span>
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
          <div class="flex items-center justify-between border-b border-subtle pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-text-main font-mono flex items-center gap-2">
                <Network class="w-4 h-4 text-brand-periwinkle" />
                CORE ROUTER / SWITCH SNMP DISCOVERY
              </h2>
              <p class="text-xs text-text-secondary mt-0.5">Cross-subnet L3 ARP table querying for automatic IP-to-MAC resolution.</p>
            </div>
            <div class="flex items-center gap-2">
              <button
                v-if="isTabDirty('network')"
                type="button"
                @click="revertTabState('network')"
                class="px-3 py-1.5 rounded-xl border border-subtle text-text-secondary hover:text-text-main text-xs font-medium transition-colors cursor-pointer"
              >
                Batal / Reset
              </button>
              <button
                type="button"
                @click="handleSaveCoreSwitch"
                :disabled="isSavingCoreSwitch"
                class="px-3.5 py-1.5 rounded-xl bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white text-xs font-semibold flex items-center gap-1.5 shadow-md shadow-brand-periwinkle/20 transition-all disabled:opacity-50 cursor-pointer"
              >
                <Save class="w-3.5 h-3.5" />
                <span>{{ isSavingCoreSwitch ? 'Saving...' : 'Save SNMP Target' }}</span>
              </button>
            </div>
          </div>

          <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-4 shadow-xl">
            <div v-if="settingStore.settings.coreSwitch" class="space-y-4 text-xs">
              <div class="grid grid-cols-1 sm:grid-cols-4 gap-4">
                <div class="space-y-1.5 sm:col-span-1">
                  <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">Core Switch IP</label>
                  <input
                    type="text"
                    v-model="settingStore.settings.coreSwitch.ip"
                    placeholder="e.g. 10.10.1.1"
                    class="w-full bg-card border border-subtle rounded-xl px-3 py-2 text-text-main font-mono text-xs focus:outline-none focus:border-brand-periwinkle"
                  />
                </div>
                <div class="space-y-1.5 sm:col-span-1">
                  <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">Community String</label>
                  <input
                    type="text"
                    v-model="settingStore.settings.coreSwitch.community"
                    placeholder="public"
                    class="w-full bg-card border border-subtle rounded-xl px-3 py-2 text-text-main font-mono text-xs focus:outline-none focus:border-brand-periwinkle"
                  />
                </div>
                <div class="space-y-1.5 sm:col-span-1">
                  <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">Port</label>
                  <input
                    type="number"
                    v-model.number="settingStore.settings.coreSwitch.port"
                    placeholder="161"
                    class="w-full bg-card border border-subtle rounded-xl px-3 py-2 text-text-main font-mono text-xs focus:outline-none focus:border-brand-periwinkle"
                  />
                </div>
                <div class="space-y-1.5 sm:col-span-1">
                  <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">SNMP Version</label>
                  <select
                    v-model="settingStore.settings.coreSwitch.version"
                    class="w-full bg-card border border-subtle rounded-xl px-3 py-2 text-text-main font-mono text-xs focus:outline-none focus:border-brand-periwinkle"
                  >
                    <option value="v2c">v2c</option>
                    <option value="v1">v1</option>
                  </select>
                </div>
              </div>

              <div class="p-3.5 rounded-xl bg-brand-periwinkle/10 border border-brand-periwinkle/25 text-[11px] font-mono text-text-main flex items-start gap-3">
                <AlertCircle class="w-4 h-4 text-brand-periwinkle shrink-0 mt-0.5" />
                <div class="space-y-1 leading-relaxed">
                  <p class="font-bold text-brand-periwinkle">Layer-3 Cross-Subnet MAC Resolution:</p>
                  <p class="text-text-secondary">
                    Host OS <code class="bg-card border border-subtle px-1 py-0.5 rounded text-text-main">arp -a</code> cannot inspect MAC addresses across VLAN subnets. Auto Detect queries this Core Router / Switch via SNMP OID <code class="bg-card border border-subtle px-1 py-0.5 rounded text-brand-periwinkle">ipNetToMediaPhysAddress (.1.3.6.1.2.1.4.22.1.2)</code> to accurately correlate cross-subnet IP addresses to device MAC addresses.
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
          <div class="flex items-center justify-between border-b border-subtle pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-text-main font-mono flex items-center gap-2">
                <Archive class="w-4 h-4 text-brand-periwinkle" />
                INCIDENT RETENTION &amp; ARCHIVING POLICY
              </h2>
              <p class="text-xs text-text-secondary mt-0.5">Automated database archiving and housekeeping schedules for resolved tickets.</p>
            </div>
            <div class="flex items-center gap-2">
              <button
                v-if="isTabDirty('retention')"
                type="button"
                @click="revertTabState('retention')"
                class="px-3 py-1.5 rounded-xl border border-subtle text-text-secondary hover:text-text-main text-xs font-medium transition-colors cursor-pointer"
              >
                Batal / Reset
              </button>
              <button
                type="button"
                @click="handleSaveRetention"
                :disabled="isSavingRetention"
                class="px-3.5 py-1.5 rounded-xl bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white text-xs font-semibold flex items-center gap-1.5 shadow-md shadow-brand-periwinkle/20 transition-all disabled:opacity-50 cursor-pointer"
              >
                <Save class="w-3.5 h-3.5" />
                <span>{{ isSavingRetention ? 'Saving...' : 'Save Retention Policy' }}</span>
              </button>
            </div>
          </div>

          <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-4 shadow-xl">
            <div class="space-y-4 text-xs max-w-lg">
              <div class="space-y-1.5">
                <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">
                  Auto-Archive Resolved Incidents Older Than
                </label>
                <div class="flex items-center gap-2">
                  <input
                    type="number"
                    min="7"
                    max="365"
                    v-model.number="settingStore.settings.retentionDays"
                    class="w-32 bg-card border border-subtle rounded-xl px-3 py-2 text-text-main font-mono text-xs focus:outline-none focus:border-brand-periwinkle"
                  />
                  <span class="text-xs font-mono text-text-secondary">days</span>
                </div>
                <p class="text-[10px] text-text-muted font-mono">
                  Resolved tickets older than this threshold are safely transferred to <code class="text-text-secondary">incidents_archive</code>. Active/Open incidents are never purged.
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- ══════════════════════════════════════════════════════════════════════════
             TAB 5: Location Management
             ══════════════════════════════════════════════════════════════════════════ -->
        <div v-else-if="activeTab === 'locations'" class="space-y-6 animate-fadeIn">
          <div class="flex items-center justify-between border-b border-subtle pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-text-main font-mono flex items-center gap-2">
                <MapPin class="w-4 h-4 text-brand-periwinkle" />
                LOCATION &amp; SITE MANAGEMENT
              </h2>
              <p class="text-xs text-text-secondary mt-0.5">Manage installation sites, buildings, floors, and server rooms.</p>
            </div>
            <button
              @click="openAddLocationModal"
              class="px-3.5 py-1.5 rounded-xl bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-semibold text-xs flex items-center gap-1.5 shadow-md shadow-brand-periwinkle/20 cursor-pointer"
            >
              <Plus class="w-4 h-4" />
              Add Location
            </button>
          </div>

          <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-4 shadow-xl">
            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs text-text-secondary">
                <thead class="bg-card font-mono text-[10px] uppercase text-text-secondary">
                  <tr>
                    <th class="py-3 px-4">Location Name</th>
                    <th class="py-3 px-4">Description</th>
                    <th class="py-3 px-4">Assigned Devices</th>
                    <th class="py-3 px-4 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-subtle">
                  <tr v-if="isLocationsLoading">
                    <td colspan="4" class="p-0 border-0"><SkeletonTable :rows="3" :cols="4" /></td>
                  </tr>
                  <tr v-else-if="locationsList.length === 0">
                    <td colspan="4" class="py-4 text-center text-text-muted font-mono text-xs">No locations registered</td>
                  </tr>
                  <tr v-else v-for="loc in locationsList" :key="loc.id" class="hover:bg-card">
                    <td class="py-3 px-4 font-bold text-text-main font-mono flex items-center gap-2">
                      <MapPin class="w-3.5 h-3.5 text-brand-periwinkle" />
                      <span>{{ loc.name }}</span>
                    </td>
                    <td class="py-3 px-4 text-text-secondary font-mono">{{ loc.description || '-' }}</td>
                    <td class="py-3 px-4 font-mono">
                      <span class="px-2 py-0.5 rounded text-[10px] font-bold" :class="loc.deviceCount ? 'bg-brand-periwinkle/10 text-brand-periwinkle' : 'bg-gray-500/10 text-text-secondary'">
                        {{ loc.deviceCount || 0 }} devices
                      </span>
                    </td>
                    <td class="py-3 px-4 text-right">
                      <div class="flex items-center justify-end gap-2">
                        <button
                          @click="openEditLocationModal(loc)"
                          class="px-2.5 py-1 rounded-lg bg-card border border-subtle hover:border-brand-periwinkle text-brand-periwinkle hover:text-brand-periwinkle-hover text-[11px] font-mono transition-colors cursor-pointer"
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
          <div class="flex items-center justify-between border-b border-subtle pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-text-main font-mono flex items-center gap-2">
                <Users class="w-4 h-4 text-brand-periwinkle" />
                USERS &amp; ROLES MANAGEMENT
              </h2>
              <p class="text-xs text-text-secondary mt-0.5">Manage operator accounts, privileges, and system roles.</p>
            </div>
            <button
              @click="openAddUserModal"
              class="px-3.5 py-1.5 rounded-xl bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-semibold text-xs flex items-center gap-1.5 shadow-md shadow-brand-periwinkle/20 cursor-pointer"
            >
              <UserPlus class="w-4 h-4" />
              Add User
            </button>
          </div>

          <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-4 shadow-xl">
            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs text-text-secondary">
                <thead class="bg-card font-mono text-[10px] uppercase text-text-secondary">
                  <tr>
                    <th class="py-3 px-4">User</th>
                    <th class="py-3 px-4">Role</th>
                    <th class="py-3 px-4">Status</th>
                    <th class="py-3 px-4">Last Active</th>
                    <th class="py-3 px-4 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-subtle">
                  <tr v-if="settingStore.isLoading">
                    <td colspan="5" class="p-0 border-0"><SkeletonTable :rows="3" :cols="5" /></td>
                  </tr>
                  <tr v-else-if="paginatedUsers.length === 0">
                    <td colspan="5" class="py-4 text-center text-text-muted font-mono text-xs">No user accounts registered</td>
                  </tr>
                  <tr v-else v-for="usr in paginatedUsers" :key="usr.id" class="hover:bg-card">
                    <td class="py-3 px-4 flex items-center gap-3">
                      <!-- User Profile Picture with Initials Fallback -->
                      <img
                        v-if="usr.avatarUrl"
                        :src="usr.avatarUrl"
                        class="w-8 h-8 rounded-full object-cover border border-subtle"
                        alt="Avatar"
                      />
                      <div
                        v-else
                        class="w-8 h-8 rounded-full bg-brand-periwinkle/15 border border-brand-periwinkle/30 flex items-center justify-center font-bold text-brand-periwinkle font-mono text-xs"
                      >
                        {{ (usr.name || usr.username || 'U').charAt(0).toUpperCase() }}
                      </div>
                      <div>
                        <h4 class="font-bold text-text-main flex items-center gap-2">
                          <span>{{ usr.name }}</span>
                          <span v-if="usr.username" class="text-[10px] font-mono text-brand-periwinkle">@{{ usr.username }}</span>
                        </h4>
                        <p class="text-[10px] font-mono text-text-muted">{{ usr.email }}</p>
                      </div>
                    </td>
                    <td class="py-3 px-4">
                      <span class="px-2 py-0.5 rounded font-mono text-[10px] font-bold uppercase bg-brand-periwinkle/15 text-brand-periwinkle border border-brand-periwinkle/30">
                        {{ usr.role }}
                      </span>
                    </td>
                    <td class="py-3 px-4">
                      <span class="inline-flex items-center gap-1.5 text-xs text-status-up">
                        <span class="w-1.5 h-1.5 rounded-full bg-status-up"></span>
                        {{ usr.status }}
                      </span>
                    </td>
                    <td class="py-3 px-4 font-mono text-text-muted text-[11px]">{{ usr.lastActive }}</td>
                    <td class="py-3 px-4 text-right">
                      <div class="flex items-center justify-end gap-2">
                        <button
                          @click="openEditUser(usr)"
                          class="px-2.5 py-1 rounded-lg bg-card border border-subtle hover:border-brand-periwinkle text-brand-periwinkle hover:text-brand-periwinkle-hover text-[11px] font-mono transition-colors cursor-pointer"
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
          <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-4 shadow-xl">
            <div class="flex items-center justify-between border-b border-subtle pb-3">
              <div>
                <h3 class="text-xs font-bold uppercase tracking-wider text-text-main font-mono flex items-center gap-2">
                  <Activity class="w-4 h-4 text-brand-periwinkle" />
                  User Activity &amp; Audit Logs
                </h3>
                <p class="text-[11px] text-text-muted mt-0.5">Real-time audit record of authentication, configuration changes, and sessions.</p>
              </div>
              <button
                @click="fetchUserLogs"
                class="px-2.5 py-1 rounded-lg bg-card border border-subtle hover:border-brand-periwinkle text-text-secondary text-xs font-mono flex items-center gap-1.5 cursor-pointer"
              >
                <RefreshCw class="w-3.5 h-3.5" :class="isLogsLoading ? 'animate-spin' : ''" />
                Refresh Logs
              </button>
            </div>

            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs text-text-secondary">
                <thead class="bg-card font-mono text-[10px] uppercase text-text-muted">
                  <tr>
                    <th class="py-2.5 px-3">User</th>
                    <th class="py-2.5 px-3">Action</th>
                    <th class="py-2.5 px-3">Detail</th>
                    <th class="py-2.5 px-3">IP Address</th>
                    <th class="py-2.5 px-3 text-right">Occurred At</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-subtle">
                  <tr v-if="isLogsLoading || settingStore.isLoading">
                    <td colspan="5" class="p-0 border-0"><SkeletonTable :rows="4" :cols="5" /></td>
                  </tr>
                  <tr v-else-if="userLogs.length === 0">
                    <td colspan="5" class="py-4 text-center text-text-muted font-mono text-xs">No activity logs recorded yet</td>
                  </tr>
                  <tr v-else v-for="log in userLogs" :key="log.id" class="hover:bg-card">
                    <td class="py-2.5 px-3 font-semibold text-text-main">{{ log.userName || log.userId }}</td>
                    <td class="py-2.5 px-3">
                      <span
                        class="px-2 py-0.5 rounded text-[10px] font-mono font-bold uppercase"
                        :class="[
                          log.action === 'login' ? 'bg-status-up/15 text-status-up border border-status-up/30' :
                          log.action === 'logout' ? 'bg-amber-500/15 text-amber-400 border border-amber-500/30' :
                          'bg-brand-periwinkle/15 text-brand-periwinkle border border-brand-periwinkle/30'
                        ]"
                      >
                        {{ log.action }}
                      </span>
                    </td>
                    <td class="py-2.5 px-3 font-mono text-text-secondary text-[11px]">{{ log.detail }}</td>
                    <td class="py-2.5 px-3 font-mono text-text-secondary text-[10px]">{{ log.ipAddress || '127.0.0.1' }}</td>
                    <td class="py-2.5 px-3 text-right font-mono text-text-muted text-[10px]">
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

        <!-- ══════════════════════════════════════════════════════════════════════════
             DHCP Sync Engine Tab
             ══════════════════════════════════════════════════════════════════════════ -->
        <div v-else-if="activeTab === 'dhcp'" class="space-y-6 animate-fadeIn">
          <div class="bg-surface border border-subtle rounded-2xl p-6 shadow-xl">
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-status-up/10 flex items-center justify-center text-status-up border border-status-up/20">
                  <Network class="w-5 h-5" />
                </div>
                <div>
                  <h3 class="text-sm font-bold text-text-main font-mono">DHCP Sync Engine</h3>
                  <p class="text-xs text-text-secondary">Worker is actively listening to Kea MySQL lease updates</p>
                </div>
              </div>
              <div class="px-3 py-1 bg-status-up/10 text-status-up border border-status-up/20 rounded-lg text-xs font-mono font-bold flex items-center gap-2">
                <span class="w-2 h-2 rounded-full bg-status-up animate-pulse"></span>
                Online
              </div>
            </div>
          </div>

          <!-- Logs Table -->
          <div class="bg-surface border border-subtle rounded-2xl overflow-hidden shadow-xl">
            <div class="px-5 py-4 border-b border-subtle bg-card flex justify-between items-center">
              <h3 class="text-sm font-bold text-text-main font-mono">Recent IP Change Logs</h3>
              <button @click="fetchDHCPLogs(true)" class="p-1.5 hover:bg-subtle rounded text-text-secondary transition-colors">
                <RefreshCw class="w-4 h-4" :class="{'animate-spin text-brand-periwinkle': isFetchingDHCPLogs}" />
              </button>
            </div>
            <div class="overflow-x-auto">
              <table class="w-full text-left text-sm whitespace-nowrap">
                <thead class="bg-card border-b border-subtle text-text-secondary text-xs font-mono uppercase tracking-wider">
                  <tr>
                    <th class="px-5 py-3 font-semibold">Device</th>
                    <th class="px-5 py-3 font-semibold">Old IP</th>
                    <th class="px-5 py-3 font-semibold">New IP</th>
                    <th class="px-5 py-3 font-semibold">Timestamp</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-subtle/50">
                  <tr v-if="dhcpLogs.length === 0">
                    <td colspan="4" class="px-5 py-8 text-center text-text-secondary text-xs italic font-mono">No recent IP changes recorded.</td>
                  </tr>
                  <tr v-for="log in dhcpLogs" :key="log.id" class="hover:bg-subtle/30 transition-colors">
                    <td class="px-5 py-3 font-medium text-text-main">{{ log.deviceName || log.deviceId }}</td>
                    <td class="px-5 py-3 font-mono text-xs text-red-400 line-through decoration-red-400/50">{{ log.oldIp || '-' }}</td>
                    <td class="px-5 py-3 font-mono text-xs text-status-up">{{ log.newIp }}</td>
                    <td class="px-5 py-3 text-xs text-text-secondary">{{ new Date(log.timestamp).toLocaleString() }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
        <!-- ══════════════════════════════════════════════════════════════════════════
             TAB 8: Branding & Appearance (Admin Only)
             ══════════════════════════════════════════════════════════════════════════ -->
        <div v-else-if="activeTab === 'branding'" class="space-y-6 animate-fadeIn">
          <div class="flex items-center justify-between border-b border-subtle pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-text-main font-mono flex items-center gap-2">
                <Palette class="w-4 h-4 text-brand-periwinkle" />
                BRANDING &amp; APPEARANCE SETTINGS
              </h2>
              <p class="text-xs text-text-secondary mt-0.5">Customize web logo, browser tab favicon, and system agency title.</p>
            </div>
            <div class="flex items-center gap-2">
              <button
                v-if="isBrandingDirty"
                type="button"
                @click="resetBrandingForm"
                class="px-3.5 py-2 rounded-xl border border-subtle text-text-secondary hover:text-text-main text-xs font-medium transition-colors cursor-pointer"
              >
                Batal / Reset
              </button>
              <button
                @click="handleSaveBranding"
                :disabled="isSavingBranding"
                class="px-4 py-2 rounded-xl bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-semibold text-xs shadow-md shadow-brand-periwinkle/20 flex items-center gap-2 disabled:opacity-50 cursor-pointer"
              >
                <Save class="w-4 h-4" />
                <span>{{ isSavingBranding ? 'Saving...' : 'Save Branding' }}</span>
              </button>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <!-- App Titles Form -->
            <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-4">
              <h3 class="text-xs font-bold text-text-main font-mono uppercase tracking-wider">System Names &amp; Labels</h3>
              
              <div class="space-y-1.5">
                <label class="block font-mono uppercase text-[10px] text-text-secondary">Application Title / Header</label>
                <input
                  v-model="brandingForm.appTitle"
                  type="text"
                  placeholder="e.g. SANOC"
                  class="w-full bg-card border border-subtle rounded-xl px-3 py-2.5 text-xs text-text-main focus:outline-none focus:border-brand-periwinkle font-mono"
                />
              </div>

              <div class="space-y-1.5">
                <label class="block font-mono uppercase text-[10px] text-text-secondary">Application Subtitle</label>
                <input
                  v-model="brandingForm.appSubtitle"
                  type="text"
                  placeholder="e.g. Jabar Regional SANOC"
                  class="w-full bg-card border border-subtle rounded-xl px-3 py-2.5 text-xs text-text-main focus:outline-none focus:border-brand-periwinkle"
                />
              </div>
            </div>

            <!-- Live Preview Card -->
            <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-4">
              <h3 class="text-xs font-bold text-text-main font-mono uppercase tracking-wider">Live Preview</h3>
              
              <div class="p-4 bg-card border border-subtle rounded-xl space-y-3">
                <span class="text-[10px] font-mono text-text-muted uppercase">Sidebar Header Preview</span>
                <div class="flex items-center gap-3 p-3 bg-surface rounded-xl border border-subtle">
                  <div class="w-10 h-10 rounded-xl bg-card border border-subtle flex items-center justify-center shrink-0 overflow-hidden relative shadow-sm">
                    <img
                      v-if="brandingForm.logoUrl"
                      :src="brandingForm.logoUrl"
                      alt="Logo Preview"
                      class="w-full h-full block"
                      :class="(brandingForm.logoFit || 'cover') === 'cover' ? 'object-cover' : 'object-contain p-1'"
                    />
                    <Activity v-else class="w-5 h-5 text-brand-periwinkle" />
                  </div>
                  <div class="min-w-0 flex-1 flex flex-col justify-center">
                    <div class="flex items-center justify-between gap-1">
                      <h4 class="text-sm font-black text-text-main font-mono leading-none truncate">{{ brandingForm.appTitle || 'SANOC' }}</h4>
                      <span class="text-[9px] font-mono text-brand-periwinkle font-semibold bg-brand-periwinkle/10 px-1 py-0.5 rounded border border-brand-periwinkle/20 shrink-0">v2.6.0</span>
                    </div>
                    <p class="text-[10px] text-text-secondary mt-1 truncate leading-tight font-mono uppercase">{{ brandingForm.appSubtitle || 'Jabar Regional SANOC' }}</p>
                  </div>
                </div>

                <span class="text-[10px] font-mono text-text-muted uppercase mt-2 block">Browser Tab Preview</span>
                <div class="flex items-center gap-2 p-2 bg-card rounded-t-lg border-t border-x border-subtle max-w-xs">
                  <img v-if="brandingForm.faviconUrl" :src="brandingForm.faviconUrl" class="w-4 h-4 object-contain rounded" alt="Favicon Preview" />
                  <div v-else class="w-4 h-4 rounded-full bg-brand-periwinkle/30"></div>
                  <span class="text-xs text-text-main truncate font-sans">{{ brandingForm.appTitle || 'SANOC' }} — Network Control Center</span>
                </div>
              </div>
            </div>

            <!-- Upload Logo Card & Fit Mode -->
            <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-4">
              <div class="flex items-center justify-between">
                <h3 class="text-xs font-bold text-text-main font-mono uppercase tracking-wider">Logo Image</h3>
                <span class="text-[10px] text-text-secondary">PNG, SVG, JPG (Max 2MB)</span>
              </div>
              <div class="flex items-center gap-4">
                <div class="w-16 h-16 rounded-2xl bg-card border border-subtle flex items-center justify-center overflow-hidden shrink-0 relative">
                  <img
                    v-if="brandingForm.logoUrl"
                    :src="brandingForm.logoUrl"
                    class="w-full h-full block"
                    :class="(brandingForm.logoFit || 'cover') === 'cover' ? 'object-cover' : 'object-contain p-2'"
                  />
                  <Image v-else class="w-6 h-6 text-text-muted" />
                </div>
                <div class="space-y-2 flex-1">
                  <input
                    ref="logoInputRef"
                    type="file"
                    accept="image/png,image/svg+xml,image/jpeg,image/webp"
                    class="hidden"
                    @change="handleLogoUpload"
                  />
                  <div class="flex gap-2">
                    <button
                      type="button"
                      @click="logoInputRef?.click()"
                      :disabled="isUploadingLogo"
                      class="px-3 py-1.5 rounded-lg bg-brand-periwinkle/15 border border-brand-periwinkle/30 text-brand-periwinkle hover:bg-brand-periwinkle/25 text-xs font-medium flex items-center gap-1.5 transition-colors cursor-pointer"
                    >
                      <Upload class="w-3.5 h-3.5" />
                      {{ isUploadingLogo ? 'Uploading...' : 'Upload Logo' }}
                    </button>
                    <button
                      v-if="brandingForm.logoUrl"
                      type="button"
                      @click="brandingForm.logoUrl = ''"
                      class="px-2.5 py-1.5 rounded-lg border border-subtle text-text-secondary hover:text-red-400 text-xs transition-colors cursor-pointer"
                    >
                      Reset Image
                    </button>
                  </div>
                  <p class="text-[10px] text-text-muted">Tampil pada sidebar kiri, header navigasi, dan halaman login.</p>
                </div>
              </div>

              <!-- Fit Mode Selector -->
              <div v-if="brandingForm.logoUrl" class="p-3 bg-card border border-subtle rounded-xl flex flex-wrap items-center justify-between gap-3">
                <span class="text-[11px] font-mono text-text-secondary font-medium">Tampilan / Mode Logo:</span>
                <div class="flex gap-2">
                  <button
                    type="button"
                    @click="brandingForm.logoFit = 'cover'"
                    class="px-3 py-1.5 rounded-lg text-xs font-mono transition-colors cursor-pointer flex items-center gap-1.5"
                    :class="brandingForm.logoFit === 'cover' ? 'bg-brand-periwinkle text-white font-bold shadow-md shadow-brand-periwinkle/20' : 'bg-surface border border-subtle text-text-secondary hover:text-text-main'"
                  >
                    <span>Penuh (Cover / Pas Rapat)</span>
                  </button>
                  <button
                    type="button"
                    @click="brandingForm.logoFit = 'contain'"
                    class="px-3 py-1.5 rounded-lg text-xs font-mono transition-colors cursor-pointer flex items-center gap-1.5"
                    :class="brandingForm.logoFit === 'contain' ? 'bg-brand-periwinkle text-white font-bold shadow-md shadow-brand-periwinkle/20' : 'bg-surface border border-subtle text-text-secondary hover:text-text-main'"
                  >
                    <span>Utuh (Contain / Proporsional)</span>
                  </button>
                </div>
              </div>
            </div>

            <!-- Upload Favicon Card -->
            <div class="bg-surface border border-subtle rounded-2xl p-5 space-y-3">
              <div class="flex items-center justify-between">
                <h3 class="text-xs font-bold text-text-main font-mono uppercase tracking-wider">Browser Favicon</h3>
                <span class="text-[10px] text-text-secondary">.ICO, .PNG, .SVG (Max 2MB)</span>
              </div>
              <div class="flex items-center gap-4">
                <div class="w-16 h-16 rounded-2xl bg-card border border-subtle flex items-center justify-center overflow-hidden shrink-0">
                  <img v-if="brandingForm.faviconUrl" :src="brandingForm.faviconUrl" class="w-8 h-8 object-contain" />
                  <Globe v-else class="w-6 h-6 text-text-muted" />
                </div>
                <div class="space-y-2 flex-1">
                  <input
                    ref="faviconInputRef"
                    type="file"
                    accept=".ico,image/x-icon,image/png,image/svg+xml"
                    class="hidden"
                    @change="handleFaviconUpload"
                  />
                  <div class="flex gap-2">
                    <button
                      type="button"
                      @click="faviconInputRef?.click()"
                      :disabled="isUploadingFavicon"
                      class="px-3 py-1.5 rounded-lg bg-brand-periwinkle/15 border border-brand-periwinkle/30 text-brand-periwinkle hover:bg-brand-periwinkle/25 text-xs font-medium flex items-center gap-1.5 transition-colors cursor-pointer"
                    >
                      <Upload class="w-3.5 h-3.5" />
                      {{ isUploadingFavicon ? 'Uploading...' : 'Upload Favicon' }}
                    </button>
                    <button
                      v-if="brandingForm.faviconUrl"
                      type="button"
                      @click="brandingForm.faviconUrl = ''"
                      class="px-2.5 py-1.5 rounded-lg border border-subtle text-text-secondary hover:text-red-400 text-xs transition-colors cursor-pointer"
                    >
                      Reset
                    </button>
                  </div>
                  <p class="text-[10px] text-text-muted">Icon displayed in the browser tab and bookmarks.</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Fallback if unauthorized tab is somehow targeted -->
        <div v-else class="p-8 text-center bg-surface border border-subtle rounded-2xl space-y-3">
          <ShieldAlert class="w-8 h-8 text-amber-400 mx-auto" />
          <h3 class="text-sm font-bold text-text-main font-mono">Category Unavailable</h3>
          <p class="text-xs text-text-secondary">You do not have permission to view this settings category.</p>
        </div>
      </main>
    </div>

    <!-- ══════════════════════════════════════════════════════════════════════════
         MODALS & NOTIFICATIONS
         ══════════════════════════════════════════════════════════════════════════ -->

    <!-- Save Success / Feedback Toast Notification -->
    <transition
      enter-active-class="transition ease-out duration-300"
      enter-from-class="transform translate-y-4 opacity-0"
      enter-to-class="transform translate-y-0 opacity-100"
      leave-active-class="transition ease-in duration-200"
      leave-from-class="transform translate-y-0 opacity-100"
      leave-to-class="transform translate-y-4 opacity-0"
    >
      <div
        v-if="showFeedbackModal"
        class="fixed bottom-6 right-6 z-50 max-w-sm w-full p-4 rounded-xl border shadow-xl flex items-start gap-3 backdrop-blur-md"
        :class="feedbackSuccess ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300' : 'bg-red-500/10 border-red-500/30 text-red-300'"
      >
        <Check v-if="feedbackSuccess" class="w-5 h-5 shrink-0 text-emerald-400 mt-0.5" />
        <AlertTriangle v-else class="w-5 h-5 shrink-0 text-red-400 mt-0.5" />
        <div class="space-y-1 pr-6 flex-1">
          <h4 class="font-bold text-text-main text-sm">{{ feedbackTitle }}</h4>
          <p class="text-xs opacity-90 leading-relaxed text-text-secondary">{{ feedbackMessage }}</p>
        </div>
        <button @click="showFeedbackModal = false" class="absolute top-4 right-4 text-text-muted hover:text-text-main transition-colors">
          <X class="w-4 h-4" />
        </button>
      </div>
    </transition>

    <!-- Tab Switch Discard Confirmation Modal -->
    <Modal :is-open="showTabSwitchConfirmModal" title="Perubahan Belum Disimpan" @close="showTabSwitchConfirmModal = false">
      <template #default>
        <div class="p-4 bg-amber-500/10 border border-amber-500/30 rounded-xl flex items-start gap-3 text-amber-400 text-xs">
          <AlertTriangle class="w-5 h-5 shrink-0 mt-0.5" />
          <div class="space-y-1">
            <h4 class="font-bold text-text-main text-sm">Ada Perubahan yang Belum Disimpan</h4>
            <p class="text-text-secondary leading-relaxed">
              Anda telah mengubah konfigurasi pada tab <span class="font-bold text-amber-400">"{{ currentTabLabel }}"</span> tetapi belum menyimpannya. 
              Apakah Anda ingin membatalkan perubahan tersebut dan berpindah tab?
            </p>
          </div>
        </div>
      </template>
      <template #footer>
        <div class="flex items-center justify-end gap-3 w-full">
          <button
            type="button"
            @click="showTabSwitchConfirmModal = false"
            class="px-4 py-2 rounded-xl border border-subtle text-text-secondary hover:text-text-main text-xs font-mono cursor-pointer"
          >
            Tetap di Tab Ini
          </button>
          <button
            type="button"
            @click="confirmDiscardTabSwitch"
            class="px-5 py-2 rounded-xl bg-red-500 hover:bg-red-600 text-white font-semibold text-xs shadow-md shadow-red-500/20 font-mono cursor-pointer"
          >
            Batalkan &amp; Pindah Tab
          </button>
        </div>
      </template>
    </Modal>

    <!-- Location Modal (Add / Edit) -->
    <Modal :is-open="isLocationModalOpen" :title="isLocationEdit ? 'Edit Location' : 'Add New Location'" @close="closeLocationModal">
      <template #default>
        <form @submit.prevent="handleSaveLocation" class="space-y-4 text-xs">
          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-text-secondary">Location Name *</label>
            <input
              v-model="locationForm.name"
              type="text"
              required
              placeholder="e.g. Gedung Sate - Lt 2"
              class="w-full bg-card border border-subtle rounded-xl px-3 py-2 text-text-main focus:outline-none focus:border-brand-periwinkle"
            />
          </div>
          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-text-secondary">Description</label>
            <input
              v-model="locationForm.description"
              type="text"
              placeholder="e.g. Server Room Main Switch Rack"
              class="w-full bg-card border border-subtle rounded-xl px-3 py-2 text-text-main focus:outline-none focus:border-brand-periwinkle"
            />
          </div>
        </form>
      </template>

      <template #footer>
        <button
          type="button"
          @click="closeLocationModal"
          class="px-4 py-2 rounded-xl border border-subtle text-text-secondary hover:text-text-main text-xs cursor-pointer"
        >
          Cancel
        </button>
        <button
          type="button"
          @click="handleSaveLocation"
          :disabled="isSavingLocation"
          class="px-5 py-2 rounded-xl bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-semibold text-xs shadow-md shadow-brand-periwinkle/20 disabled:opacity-50 cursor-pointer"
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
              <strong>{{ locationToDelete.deviceCount }} devices</strong> are currently assigned to location <strong class="text-text-main">{{ locationToDelete.name }}</strong>. Please reassign or delete those devices first before deleting this location.
            </p>
          </div>
          <div v-else class="p-3 bg-red-500/10 border border-red-500/30 rounded-xl flex items-start gap-3 text-red-400">
            <AlertTriangle class="w-5 h-5 shrink-0 mt-0.5" />
            <div>
              <h4 class="font-bold font-mono">Confirm Deletion</h4>
              <p class="mt-1 text-[11px] text-text-secondary">
                Are you sure you want to delete location <strong class="text-text-main font-mono">{{ locationToDelete?.name }}</strong>?
              </p>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <button
          @click="isDeleteLocationModalOpen = false"
          class="px-4 py-2 rounded-xl border border-subtle text-text-secondary hover:text-text-main text-xs font-mono cursor-pointer"
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
        <div class="flex items-center gap-2 mb-4 p-2.5 rounded-xl bg-card border border-subtle">
          <div class="flex items-center gap-1.5 flex-1">
            <span
              class="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold font-mono"
              :class="addUserStep === 1 ? 'bg-brand-periwinkle text-white' : 'bg-status-up text-black'"
            >
              {{ addUserStep === 1 ? '1' : '✓' }}
            </span>
            <span class="text-[11px] font-mono" :class="addUserStep === 1 ? 'text-text-main font-bold' : 'text-text-secondary'">
              1. User Details
            </span>
          </div>
          <div class="w-4 h-px bg-subtle"></div>
          <div class="flex items-center gap-1.5 flex-1">
            <span
              class="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold font-mono"
              :class="addUserStep === 2 ? 'bg-brand-periwinkle text-white' : 'bg-subtle text-text-secondary'"
            >
              2
            </span>
            <span class="text-[11px] font-mono" :class="addUserStep === 2 ? 'text-text-main font-bold' : 'text-text-muted'">
              2. OTP Verification
            </span>
          </div>
        </div>

        <!-- STEP 1: User Information Form -->
        <div v-if="addUserStep === 1" class="space-y-3.5 text-xs">
          <div class="p-3 bg-brand-periwinkle/10 border border-brand-periwinkle/25 rounded-xl flex items-start gap-2.5 text-brand-periwinkle">
            <ShieldAlert class="w-4 h-4 shrink-0 mt-0.5" />
            <p class="text-[11px] leading-relaxed">
              Must use an active, <strong>real email address</strong> (e.g., <code>operator@jabarprov.go.id</code> or <code>user@gmail.com</code>). The system will send a 6-digit OTP code to verify the account.
            </p>
          </div>

          <div class="space-y-1">
            <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">Full Name *</label>
            <input
              v-model="addUserForm.name"
              type="text"
              required
              placeholder="e.g. Ahmad Hidayat"
              class="w-full bg-card border border-subtle rounded-xl px-3 py-2 text-text-main text-xs focus:outline-none focus:border-brand-periwinkle"
            />
          </div>

          <div class="space-y-1">
            <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">Real Email Address *</label>
            <input
              v-model="addUserForm.email"
              type="email"
              required
              placeholder="e.g. ahmad@diskominfo.jabarprov.go.id"
              class="w-full bg-card border border-subtle rounded-xl px-3 py-2 text-text-main text-xs focus:outline-none focus:border-brand-periwinkle"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">Username</label>
              <input
                v-model="addUserForm.username"
                type="text"
                placeholder="Auto from email"
                class="w-full bg-card border border-subtle rounded-xl px-3 py-2 text-text-main text-xs focus:outline-none focus:border-brand-periwinkle"
              />
            </div>

            <div class="space-y-1">
              <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">Access Role *</label>
              <select
                v-model="addUserForm.role"
                class="w-full bg-card border border-subtle rounded-xl px-3 py-2 text-text-main text-xs focus:outline-none focus:border-brand-periwinkle"
              >
                <option value="anggota">SANOC Member</option>
                <option value="pimpinan">Leadership</option>
                <option value="admin">Admin</option>
              </select>
            </div>
          </div>

          <div class="space-y-1">
            <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold">Initial Password * (Min 8 Characters)</label>
            <input
              v-model="addUserForm.password"
              type="password"
              required
              placeholder="At least 8 characters"
              class="w-full bg-card border border-subtle rounded-xl px-3 py-2 text-text-main text-xs focus:outline-none focus:border-brand-periwinkle"
            />
          </div>
        </div>

        <!-- STEP 2: OTP Verification -->
        <div v-else class="space-y-4 text-xs">
          <div class="p-3.5 bg-card border border-subtle rounded-xl space-y-2 text-center">
            <div class="w-10 h-10 rounded-full bg-brand-periwinkle/10 text-brand-periwinkle border border-brand-periwinkle/30 mx-auto flex items-center justify-center">
              <Send class="w-5 h-5" />
            </div>
            <h4 class="font-bold text-text-main text-sm">Verify User Email</h4>
            <p class="text-text-secondary text-xs leading-relaxed">
              A 6-digit verification code has been sent to:<br />
              <strong class="text-brand-periwinkle font-mono text-xs">{{ addUserForm.email }}</strong>
            </p>
          </div>

          <div class="space-y-3 py-1">
            <label class="block font-mono uppercase text-[10px] text-text-secondary font-semibold text-center">
              Enter 6-Digit Verification Code (1 Digit Per Box)
            </label>
            <OtpInput
              v-model="addUserOTP"
              :length="6"
              :auto-focus="true"
              @complete="handleConfirmAddUser"
            />
          </div>

          <div class="flex items-center justify-between text-xs text-text-secondary pt-1">
            <button
              type="button"
              @click="addUserStep = 1"
              class="text-text-secondary hover:text-text-main transition-colors cursor-pointer"
            >
              &larr; Change Details / Email
            </button>
            <button
              type="button"
              @click="handleSendAddUserOTP"
              :disabled="addUserCountdown > 0 || isSendingAddUserOTP"
              class="text-brand-periwinkle hover:text-brand-periwinkle-hover font-semibold disabled:opacity-50 disabled:cursor-not-allowed transition-colors cursor-pointer"
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
          class="px-4 py-2 rounded-xl border border-subtle text-text-secondary hover:text-text-main text-xs cursor-pointer"
        >
          Cancel
        </button>

        <button
          v-if="addUserStep === 1"
          @click="handleSendAddUserOTP"
          :disabled="isSendingAddUserOTP"
          class="px-5 py-2 rounded-xl bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-semibold text-xs shadow-md shadow-brand-periwinkle/20 disabled:opacity-50 flex items-center gap-1.5 cursor-pointer"
        >
          <Send class="w-3.5 h-3.5" :class="isSendingAddUserOTP ? 'animate-spin' : ''" />
          <span>{{ isSendingAddUserOTP ? 'Sending OTP...' : 'Next: Send Verification Code' }}</span>
        </button>

        <button
          v-else
          @click="handleConfirmAddUser"
          :disabled="isCreatingUser || !addUserOTP || addUserOTP.length !== 6"
          class="px-5 py-2 rounded-xl bg-status-up hover:bg-status-up text-black font-bold text-xs shadow-md shadow-status-up/20 disabled:opacity-50 flex items-center gap-1.5 cursor-pointer"
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
              <p class="mt-1 text-[11px] text-text-secondary">
                Are you sure you want to permanently delete user account <strong class="text-text-main font-mono">{{ userToDelete?.name }}</strong> ({{ userToDelete?.email }})?
              </p>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <button
          @click="isDeleteModalOpen = false"
          class="px-4 py-2 rounded-xl border border-subtle text-text-secondary hover:text-text-main text-xs font-mono cursor-pointer"
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
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router';
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
  ShieldAlert,
  Palette,
  Upload,
  Image,
  Globe,
  X
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
let feedbackTimeout: ReturnType<typeof setTimeout> | null = null;

function triggerFeedback(title: string, message: string, success = true) {
  feedbackTitle.value = title;
  feedbackMessage.value = message;
  feedbackSuccess.value = success;
  showFeedbackModal.value = true;
  
  if (feedbackTimeout) clearTimeout(feedbackTimeout);
  feedbackTimeout = setTimeout(() => {
    showFeedbackModal.value = false;
  }, 4000);
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
    badgeClass: isWhatsAppConnected.value ? 'bg-status-up/15 text-status-up' : 'bg-amber-500/15 text-amber-400',
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
    id: 'dhcp',
    label: 'DHCP Sync Engine',
    description: 'DHCP Integration & IP History',
    icon: Network,
    badge: 'Worker Active',
    badgeClass: 'bg-status-up/15 text-status-up',
    permission: () => authStore.user?.role === 'admin' || authStore.hasPermission('settings.polling')
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
  },
  {
    id: 'branding',
    label: 'Branding & Appearance',
    description: 'Logo, Favicon & App Title',
    icon: Palette,
    badge: 'Admin Only',
    badgeClass: 'bg-brand-periwinkle/15 text-brand-periwinkle',
    permission: () => authStore.user?.role === 'admin'
  }
]);

const availableTabs = computed(() => {
  return allTabs.value.filter((t) => t.permission());
});

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
    if (activeTab.value === 'dhcp' && dhcpLogs.value.length === 0) {
      fetchDHCPLogs();
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

// Snapshot of initial settings to track dirty changes and allow clean rollback
const savedSettingsSnapshot = ref<any>(null);

function updateSettingsSnapshot() {
  if (settingStore.settings) {
    savedSettingsSnapshot.value = JSON.parse(JSON.stringify(settingStore.settings));
  }
}

function isTabDirty(tabId: string): boolean {
  if (tabId === 'branding') {
    return isBrandingDirty.value;
  }
  if (!savedSettingsSnapshot.value || !settingStore.settings) return false;
  const s = savedSettingsSnapshot.value;
  const curr = settingStore.settings;

  if (tabId === 'notifications') {
    return curr.rateLimitMaxMsgPerMin !== s.rateLimitMaxMsgPerMin;
  }

  if (tabId === 'polling') {
    const pollingChanged =
      curr.polling?.intervalSeconds !== s.polling?.intervalSeconds ||
      curr.polling?.concurrencyBatchSize !== s.polling?.concurrencyBatchSize ||
      curr.polling?.flapReuseWindowMinutes !== s.polling?.flapReuseWindowMinutes;
    const thresholdsChanged =
      JSON.stringify(curr.thresholds) !== JSON.stringify(s.thresholds);
    return pollingChanged || thresholdsChanged;
  }

  if (tabId === 'network') {
    return (
      curr.coreSwitch?.ip !== s.coreSwitch?.ip ||
      curr.coreSwitch?.community !== s.coreSwitch?.community ||
      curr.coreSwitch?.port !== s.coreSwitch?.port ||
      curr.coreSwitch?.version !== s.coreSwitch?.version
    );
  }

  if (tabId === 'retention') {
    return curr.retentionDays !== s.retentionDays;
  }

  return false;
}

function revertTabState(tabId: string) {
  if (tabId === 'branding') {
    resetBrandingForm();
    return;
  }
  if (!savedSettingsSnapshot.value) return;
  const s = savedSettingsSnapshot.value;

  if (tabId === 'notifications') {
    settingStore.settings.rateLimitMaxMsgPerMin = s.rateLimitMaxMsgPerMin;
  } else if (tabId === 'polling') {
    if (s.polling) settingStore.settings.polling = JSON.parse(JSON.stringify(s.polling));
    if (s.thresholds) settingStore.settings.thresholds = JSON.parse(JSON.stringify(s.thresholds));
  } else if (tabId === 'network') {
    if (s.coreSwitch) settingStore.settings.coreSwitch = JSON.parse(JSON.stringify(s.coreSwitch));
  } else if (tabId === 'retention') {
    settingStore.settings.retentionDays = s.retentionDays;
  }
}

const isCurrentTabDirty = computed(() => isTabDirty(activeTab.value));
const isAnyTabDirty = computed(() => ['notifications', 'polling', 'network', 'retention', 'branding'].some(isTabDirty));

const currentTabLabel = computed(() => {
  const t = availableTabs.value.find(tab => tab.id === activeTab.value);
  return t ? t.label : activeTab.value;
});

const pendingTabId = ref<string | null>(null);
const showTabSwitchConfirmModal = ref(false);

function switchTab(targetTabId: string) {
  if (targetTabId === activeTab.value) return;
  if (isCurrentTabDirty.value) {
    pendingTabId.value = targetTabId;
    showTabSwitchConfirmModal.value = true;
  } else {
    activeTab.value = targetTabId;
    router.replace({ query: { ...route.query, tab: targetTabId } });
  }
}

function confirmDiscardTabSwitch() {
  revertTabState(activeTab.value);
  showTabSwitchConfirmModal.value = false;
  if (pendingTabId.value) {
    activeTab.value = pendingTabId.value;
    router.replace({ query: { ...route.query, tab: pendingTabId.value } });
    pendingTabId.value = null;
  }
}

onBeforeRouteLeave((_to, _from, next) => {
  if (isAnyTabDirty.value) {
    const ok = window.confirm(`Anda memiliki perubahan konfigurasi yang belum disimpan pada tab "${currentTabLabel.value}". Apakah Anda yakin ingin membatalkan perubahan dan meninggalkan halaman ini?`);
    if (ok) {
      ['notifications', 'polling', 'network', 'retention', 'branding'].forEach(revertTabState);
      next();
    } else {
      next(false);
    }
  } else {
    next();
  }
});

// Per-category Save Handlers
async function handleSaveRateLimit() {
  isSavingRateLimit.value = true;
  try {
    await settingStore.saveSettings();
    updateSettingsSnapshot();
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
    updateSettingsSnapshot();
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
    updateSettingsSnapshot();
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
    updateSettingsSnapshot();
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

const dhcpLogs = ref<any[]>([]);
const isFetchingDHCPLogs = ref(false);

async function fetchDHCPLogs(manual = false) {
  isFetchingDHCPLogs.value = true;
  try {
    const res = await api.get('/settings/dhcp/logs');
    dhcpLogs.value = Array.isArray(res.data) ? res.data : [];
    
    if (manual === true) {
      triggerFeedback('DHCP Sync', 'Logs refreshed successfully!');
    }
  } catch (e) {
    console.error('Failed to fetch DHCP logs:', e);
    if (manual === true) {
      triggerFeedback('Error', 'Failed to refresh DHCP logs', false);
    }
  } finally {
    isFetchingDHCPLogs.value = false;
  }
}

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

function closeLocationModal() {
  locationForm.name = '';
  locationForm.description = '';
  editingLocationId.value = '';
  isLocationEdit.value = false;
  isLocationModalOpen.value = false;
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
    closeLocationModal();
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

// ─── Branding & Appearance Settings ──────────────────────────────────────────
const isSavingBranding = ref(false);
const isUploadingLogo = ref(false);
const isUploadingFavicon = ref(false);
const logoInputRef = ref<HTMLInputElement | null>(null);
const faviconInputRef = ref<HTMLInputElement | null>(null);

const brandingForm = reactive({
  appTitle: '',
  appSubtitle: '',
  logoUrl: '',
  logoFit: 'cover' as 'cover' | 'contain',
  logoScale: 100,
  faviconUrl: '',
  footerText: ''
});

const isBrandingDirty = computed(() => {
  const b = settingStore.branding;
  if (!b) return false;
  return (
    brandingForm.appTitle !== (b.appTitle || 'SANOC') ||
    brandingForm.appSubtitle !== (b.appSubtitle || 'Jabar Regional SANOC') ||
    brandingForm.logoUrl !== (b.logoUrl || '') ||
    brandingForm.logoFit !== (b.logoFit || 'cover') ||
    brandingForm.faviconUrl !== (b.faviconUrl || '')
  );
});

function resetBrandingForm() {
  const b = settingStore.branding;
  if (b) {
    brandingForm.appTitle = b.appTitle || 'SANOC';
    brandingForm.appSubtitle = b.appSubtitle || 'Jabar Regional SANOC';
    brandingForm.logoUrl = b.logoUrl || '';
    brandingForm.logoFit = b.logoFit || 'cover';
    brandingForm.logoScale = b.logoScale || 100;
    brandingForm.faviconUrl = b.faviconUrl || '';
    brandingForm.footerText = b.footerText || 'SANOC Network Operations Center';
  }
}

watch(
  () => settingStore.branding,
  (b) => {
    if (b) {
      brandingForm.appTitle = b.appTitle || 'SANOC';
      brandingForm.appSubtitle = b.appSubtitle || 'Jabar Regional SANOC';
      brandingForm.logoUrl = b.logoUrl || '';
      brandingForm.logoFit = b.logoFit || 'cover';
      brandingForm.logoScale = b.logoScale || 100;
      brandingForm.faviconUrl = b.faviconUrl || '';
      brandingForm.footerText = b.footerText || 'SANOC Network Operations Center';
    }
  },
  { immediate: true, deep: true }
);

async function handleLogoUpload(e: Event) {
  const target = e.target as HTMLInputElement;
  if (!target.files || target.files.length === 0) return;
  const file = target.files[0];
  isUploadingLogo.value = true;
  try {
    const res = await settingStore.uploadBrandingAsset(file, 'logo');
    if (res && res.url) {
      brandingForm.logoUrl = res.url;
      triggerFeedback('Logo Berhasil Diunggah', 'Logo baru siap diterapkan. Klik Simpan Branding untuk mengaktifkan.', true);
    }
  } catch (err: any) {
    triggerFeedback('Upload Gagal', err.response?.data?.error || 'Gagal mengunggah logo.', false);
  } finally {
    isUploadingLogo.value = false;
    if (logoInputRef.value) logoInputRef.value.value = '';
  }
}

async function handleFaviconUpload(e: Event) {
  const target = e.target as HTMLInputElement;
  if (!target.files || target.files.length === 0) return;
  const file = target.files[0];
  isUploadingFavicon.value = true;
  try {
    const res = await settingStore.uploadBrandingAsset(file, 'favicon');
    if (res && res.url) {
      brandingForm.faviconUrl = res.url;
      triggerFeedback('Favicon Berhasil Diunggah', 'Favicon baru siap diterapkan. Klik Simpan Branding untuk mengaktifkan.', true);
    }
  } catch (err: any) {
    triggerFeedback('Upload Gagal', err.response?.data?.error || 'Gagal mengunggah favicon.', false);
  } finally {
    isUploadingFavicon.value = false;
    if (faviconInputRef.value) faviconInputRef.value.value = '';
  }
}

async function handleSaveBranding() {
  isSavingBranding.value = true;
  try {
    await settingStore.updateBranding({
      appTitle: brandingForm.appTitle.trim() || 'SANOC',
      appSubtitle: brandingForm.appSubtitle.trim(),
      logoUrl: brandingForm.logoUrl,
      logoFit: brandingForm.logoFit,
      logoScale: Number(brandingForm.logoScale) || 100,
      faviconUrl: brandingForm.faviconUrl,
      footerText: brandingForm.footerText.trim()
    });
    triggerFeedback('Branding Berhasil Disimpan', 'Logo, Favicon, dan Nama Sistem telah diperbarui di seluruh aplikasi.', true);
  } catch (err: any) {
    triggerFeedback('Gagal Menyimpan Branding', err.response?.data?.error || 'Terjadi kesalahan sistem saat menyimpan branding.', false);
  } finally {
    isSavingBranding.value = false;
  }
}

onMounted(async () => {
  try {
    await Promise.allSettled([
      settingStore.fetchSettings(),
      settingStore.fetchBranding(),
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
    updateSettingsSnapshot();
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
