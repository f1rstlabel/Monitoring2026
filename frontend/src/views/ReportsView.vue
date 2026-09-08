<template>
  <div class="min-w-0 space-y-6">
    <!-- Header -->
    <div class="flex flex-col gap-4 border-b border-subtle pb-4 lg:flex-row lg:items-center lg:justify-between">
      <div>
        <h1 class="text-xl font-extrabold text-text-main tracking-tight">{{ reportScope === 'PUBLIC_MONITOR' ? (publicReportTab === 'summary' ? 'Public Monitoring Summary' : 'Public Monitoring Incident Reports') : 'Availability & SLA Reports' }}</h1>
        <p class="text-xs text-text-secondary mt-1">{{ reportScope === 'PUBLIC_MONITOR' ? (publicReportTab === 'summary' ? 'Aggregate availability and response across all public endpoints.' : 'Outage reports by monitor with access to timeline and notification history.') : 'Network performance analysis, device uptime metrics, MTTR, and downtime distribution' }}</p>
      </div>

      <div class="flex flex-wrap items-center gap-2 lg:justify-end">
        <!-- Period Toggle -->
        <div class="flex items-center bg-card border border-subtle rounded-lg p-0.5">
          <button
            @click="setPeriod('daily')"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-all cursor-pointer"
            :class="reportStore.period === 'daily' ? 'bg-subtle text-text-main shadow-sm' : 'text-text-secondary hover:text-text-main'"
          >
            Today
          </button>
          <button
            @click="setPeriod('weekly')"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-all cursor-pointer"
            :class="reportStore.period === 'weekly' ? 'bg-subtle text-text-main shadow-sm' : 'text-text-secondary hover:text-text-main'"
          >
            Last 7 Days
          </button>
          <button
            @click="setPeriod('monthly')"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-all cursor-pointer"
            :class="reportStore.period === 'monthly' ? 'bg-subtle text-text-main shadow-sm' : 'text-text-secondary hover:text-text-main'"
          >
            Last 30 Days
          </button>
          <button
            @click="setPeriod('custom')"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-all cursor-pointer"
            :class="reportStore.period === 'custom' ? 'bg-subtle text-text-main shadow-sm' : 'text-text-secondary hover:text-text-main'"
          >
            Custom Range
          </button>
        </div>

        <!-- Export Excel / CSV Dropdown -->
        <div v-if="authStore.hasPermission('reports.export')" class="relative">
          <button
            @click="showExportDropdown = !showExportDropdown"
            class="px-3.5 py-1.5 rounded-lg border border-subtle bg-surface hover:bg-hover text-emerald-400 font-semibold text-xs transition-all flex items-center gap-1.5 shadow-sm"
            title="Export as Excel / CSV Spreadsheet"
          >
            <FileSpreadsheet class="w-3.5 h-3.5 text-emerald-400" />
            <span>Export {{ reportScope === 'PUBLIC_MONITOR' ? (publicReportTab === 'summary' ? 'Monitor Summary' : 'Incident Reports') : 'Data' }}</span>
            <ChevronDown class="w-3 h-3 text-emerald-500/70 ml-0.5" />
          </button>

          <!-- Dropdown Options -->
          <div
            v-if="showExportDropdown"
            @click="showExportDropdown = false"
            class="absolute right-0 mt-1.5 w-52 bg-card border border-subtle rounded-xl shadow-2xl py-1.5 z-50 animate-in fade-in slide-in-from-top-1 duration-150"
          >
            <div class="px-3 py-1.5 border-b border-subtle/60 font-mono text-[10px] uppercase font-bold text-text-secondary">
              Select Export Format
            </div>
            <button
              @click="exportReport('xls')"
              class="w-full text-left px-3.5 py-2 text-xs font-mono text-text-main hover:bg-emerald-500/10 hover:text-emerald-400 flex items-center gap-2.5 transition-colors"
            >
              <FileSpreadsheet class="w-4 h-4 text-emerald-400 shrink-0" />
              <div>
                <div class="font-bold">Excel Spreadsheet (.xls)</div>
                <div class="text-[10px] text-text-secondary font-sans">Structured Excel workbook</div>
              </div>
            </button>
            <button
              @click="exportReport('csv')"
              class="w-full text-left px-3.5 py-2 text-xs font-mono text-text-main hover:bg-emerald-500/10 hover:text-emerald-400 flex items-center gap-2.5 transition-colors"
            >
              <FileText class="w-4 h-4 text-sky-400 shrink-0" />
              <div>
                <div class="font-bold">CSV File (.csv)</div>
                <div class="text-[10px] text-text-secondary font-sans">Standard UTF-8 CSV document</div>
              </div>
            </button>
          </div>
        </div>

        <!-- Export PDF -->
        <button
          v-if="authStore.hasPermission('reports.export')"
          @click="handlePrintPDF()"
          class="px-3.5 py-1.5 rounded-lg border border-subtle bg-surface hover:bg-hover text-sky-400 font-semibold text-xs transition-all flex items-center gap-1.5"
          title="Export SLA Audit as PDF"
        >
          <Printer class="w-3.5 h-3.5 text-sky-400" />
          Export PDF
        </button>
      </div>
    </div>

    <nav class="flex flex-wrap gap-1 border-b border-subtle pb-1" aria-label="Report source">
      <button v-for="scope in reportScopes" :key="scope.value" type="button" class="source-tab" :class="reportScope === scope.value ? 'source-tab-active' : ''" @click="reportScope = scope.value">{{ scope.label }}</button>
    </nav>

    <!-- Custom Date Range Bar & Type/Location Filters -->
    <div class="flex flex-wrap items-center gap-3">
      <!-- Custom Date Inputs -->
      <div v-if="reportStore.period === 'custom'" class="flex items-center gap-2 bg-surface border border-subtle rounded-lg px-3 py-1.5 text-xs text-text-secondary">
        <Calendar class="w-3.5 h-3.5 text-brand-periwinkle" />
        <span class="text-[10px] uppercase font-mono text-text-secondary">From:</span>
        <input
          v-model="reportStore.startDate"
          type="date"
          @change="reportStore.fetchReport()"
          class="bg-card border border-subtle rounded px-2 py-1 text-xs text-text-main font-mono focus:outline-none focus:border-brand-periwinkle"
        />
        <span class="text-[10px] uppercase font-mono text-text-secondary">To:</span>
        <input
          v-model="reportStore.endDate"
          type="date"
          @change="reportStore.fetchReport()"
          class="bg-card border border-subtle rounded px-2 py-1 text-xs text-text-main font-mono focus:outline-none focus:border-brand-periwinkle"
        />
      </div>

      <select
        v-if="reportScope !== 'PUBLIC_MONITOR'"
        v-model="reportStore.deviceTypeFilter"
        class="bg-card border border-subtle rounded-lg px-3 py-2 text-xs text-text-main focus:outline-none focus:border-brand-periwinkle font-mono"
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
        v-if="reportScope !== 'PUBLIC_MONITOR'"
        v-model="reportStore.locationFilter"
        class="bg-card border border-subtle rounded-lg px-3 py-2 text-xs text-text-main focus:outline-none focus:border-brand-periwinkle font-mono"
      >
        <option v-for="loc in reportStore.uniqueLocations" :key="loc" :value="loc">
          {{ loc === 'All' ? 'All Locations' : loc }}
        </option>
      </select>

    </div>

    <section v-if="reportScope === 'PUBLIC_MONITOR'" class="public-report-layout">
      <div class="public-report-main space-y-5">
      <div class="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-brand-periwinkle/25 bg-brand-periwinkle/5 p-4">
        <div><p class="text-[10px] font-mono uppercase tracking-wider text-brand-periwinkle">Report scope</p><h2 class="mt-1 text-sm font-bold text-text-main">Public monitoring reports</h2><p class="mt-1 text-[10px] text-text-secondary">Each row represents one outage. Open the detail for its timeline, notifications, and printable report.</p></div>
        <router-link to="/public-monitoring" class="text-xs font-semibold text-brand-periwinkle hover:underline">Open monitors</router-link>
      </div>
      <div class="public-report-tabs" role="tablist" aria-label="Public monitoring report views">
        <button type="button" role="tab" :aria-selected="publicReportTab === 'incidents'" class="public-report-tab" :class="publicReportTab === 'incidents' ? 'public-report-tab-active' : ''" @click="publicReportTab = 'incidents'">Incident Reports</button>
        <button type="button" role="tab" :aria-selected="publicReportTab === 'summary'" class="public-report-tab" :class="publicReportTab === 'summary' ? 'public-report-tab-active' : ''" @click="publicReportTab = 'summary'">Monitor Summary</button>
      </div>
      <div class="grid grid-cols-2 gap-3 xl:grid-cols-4">
        <div class="public-report-card"><span>Total checks</span><strong>{{ publicReport?.totalChecks ?? '—' }}</strong><small>{{ reportPeriodLabel }}</small></div>
        <div class="public-report-card"><span>Uptime</span><strong class="text-status-up">{{ publicReport ? `${publicReport.uptimePercent.toFixed(2)}%` : '—' }}</strong><small>Availability</small></div>
        <div class="public-report-card"><span>Incidents</span><strong class="text-status-down">{{ publicReport?.incidentCount ?? '—' }}</strong><small>{{ publicReport?.downtimeMinutes ?? 0 }} minutes downtime</small></div>
        <div class="public-report-card"><span>Avg response</span><strong>{{ publicReport ? `${publicReport.avgLatencyMs.toFixed(0)}ms` : '—' }}</strong><small>UP checks only</small></div>
      </div>
      <div v-if="publicReportTab === 'incidents' && publicIncidentsLoading" class="space-y-3"><SkeletonTable :rows="6" :cols="7" /></div>
      <div v-else-if="publicReportTab === 'incidents'" class="space-y-3">
        <div class="responsive-table-wrap responsive-public-incidents-table overflow-x-auto rounded-xl border border-subtle bg-surface">
        <table class="w-full text-left text-xs text-text-secondary"><thead class="border-b border-subtle bg-card font-mono text-[10px] uppercase text-text-muted"><tr><th class="px-4 py-3">Incident</th><th class="px-4 py-3">Monitor</th><th class="px-4 py-3">Status</th><th class="px-4 py-3">Started</th><th class="px-4 py-3">Duration</th><th class="px-4 py-3">Reason</th><th class="px-4 py-3">Action</th></tr></thead><tbody class="divide-y divide-subtle"><tr v-for="incident in publicIncidents" :key="incident.id" class="hover:bg-card"><td class="px-4 py-3 font-mono font-bold text-brand-periwinkle">{{ formatPublicIncidentId(incident.id) }}</td><td class="px-4 py-3"><strong class="block text-text-main">{{ incident.monitorName }}</strong><span class="text-[10px] font-mono text-text-muted">{{ incident.targetUrl }}</span></td><td class="px-4 py-3" :class="incident.status === 'ACTIVE' ? 'text-status-down' : 'text-status-up'">{{ incident.status }}</td><td class="px-4 py-3 whitespace-nowrap text-[10px] font-mono">{{ formatReportDate(incident.startedAt) }}</td><td class="px-4 py-3 whitespace-nowrap text-[10px] font-mono">{{ formatIncidentDuration(incident) }}</td><td class="max-w-[280px] truncate px-4 py-3 text-[10px]">{{ incident.lastError || incident.firstError || incident.resolutionReason || 'Recovered' }}</td><td class="px-4 py-3"><router-link :to="`/incidents/${incident.id}?source=PUBLIC_MONITOR`" class="font-semibold text-brand-periwinkle hover:underline">View report</router-link></td></tr><tr v-if="!publicIncidents.length"><td colspan="7" class="py-12 text-center text-text-muted">No public incidents found for this period.</td></tr></tbody></table>
        </div>
        <PaginationControl v-model:current-page="publicIncidentPage" v-model:page-size="publicIncidentPageSize" :total="publicIncidentTotal" />
      </div>
      <div v-if="publicReportTab === 'summary'" class="public-summary-layout">
        <div class="responsive-table-wrap responsive-public-summary-table overflow-x-auto rounded-xl border border-subtle bg-surface">
          <div class="flex items-center justify-between border-b border-subtle px-5 py-4"><div><h3 class="text-xs font-bold text-text-main">Monitor summary</h3><p class="mt-1 text-[10px] text-text-muted">Per-monitor availability and response metrics for {{ reportPeriodLabel.toLowerCase() }}.</p></div><router-link to="/public-monitoring" class="text-[10px] font-semibold text-brand-periwinkle hover:underline">Open monitors</router-link></div>
          <table class="w-full min-w-[980px] text-left text-xs text-text-secondary"><thead class="border-b border-subtle bg-card font-mono text-[10px] uppercase text-text-muted"><tr><th class="px-5 py-3">Monitor</th><th class="px-5 py-3">Type</th><th class="px-5 py-3">Target</th><th class="px-5 py-3">Group</th><th class="px-5 py-3">Status</th><th class="px-5 py-3">Uptime</th><th class="px-5 py-3">Checks</th><th class="px-5 py-3">Avg response</th><th class="px-5 py-3">Incidents</th></tr></thead><tbody class="divide-y divide-subtle"><tr v-for="summary in paginatedPublicSummaries" :key="summary.monitorId" class="cursor-pointer transition-colors hover:bg-card" @click="router.push({ path: '/public-monitoring', query: { monitorId: summary.monitorId } })"><td class="px-5 py-3"><strong class="block text-text-main">{{ summary.monitorName }}</strong><span class="text-[10px] font-mono text-text-muted">{{ summary.lastChecked ? formatReportDate(summary.lastChecked) : 'Never checked' }}</span></td><td class="px-5 py-3 font-mono text-[10px]">{{ publicMonitorTypeLabel(summary.monitorType) }}</td><td class="max-w-[260px] truncate px-5 py-3 font-mono text-[10px]" :title="summary.targetUrl">{{ summary.targetUrl || 'Host-based check' }}</td><td class="px-5 py-3 text-[10px]">{{ summary.groupName || 'Ungrouped' }}</td><td class="px-5 py-3 font-mono font-bold" :class="summary.status === 'UP' ? 'text-status-up' : summary.status === 'DOWN' ? 'text-status-down' : 'text-amber-400'">{{ summary.status }}</td><td class="px-5 py-3 font-mono text-status-up">{{ summary.uptimePercent.toFixed(2) }}%</td><td class="px-5 py-3 font-mono">{{ summary.totalChecks }} <span class="text-[10px] text-text-muted">({{ summary.upChecks }} UP / {{ summary.downChecks }} DOWN)</span></td><td class="px-5 py-3 font-mono">{{ summary.avgLatencyMs.toFixed(0) }}ms</td><td class="px-5 py-3 font-mono" :class="summary.incidentCount ? 'text-status-down' : 'text-text-secondary'">{{ summary.incidentCount }}</td></tr><tr v-if="!publicSummaries.length"><td colspan="9" class="py-12 text-center text-text-muted">No monitor summary data is available for this period.</td></tr></tbody></table>
        </div>
        <PaginationControl v-model:current-page="publicSummaryPage" v-model:page-size="publicSummaryPageSize" :total="publicSummaries.length" />
      </div>
      </div>
      <aside class="public-report-sidecard" aria-label="Public monitoring report context">
        <div class="public-sidecard-header">
          <div>
            <p class="public-sidecard-eyebrow">Report context</p>
            <h2 class="public-sidecard-title">{{ publicReportTab === 'incidents' ? 'Incident health' : 'Monitor coverage' }}</h2>
          </div>
          <span class="public-sidecard-live"><span class="public-sidecard-live-dot" /> Live data</span>
        </div>
        <div class="public-sidecard-period">{{ reportPeriodLabel }}</div>
        <div v-if="publicReportTab === 'incidents'" class="public-sidecard-stats">
          <div class="public-side-stat"><span>Active incidents</span><strong class="text-status-down">{{ publicActiveIncidentCount }}</strong></div>
          <div class="public-side-stat"><span>Resolved incidents</span><strong class="text-status-up">{{ publicResolvedIncidentCount }}</strong></div>
          <div class="public-side-stat"><span>Total downtime</span><strong>{{ publicReport?.downtimeMinutes ?? 0 }}<small> min</small></strong></div>
          <div class="public-sidecard-callout" :class="publicActiveIncidentCount > 0 ? 'public-sidecard-callout-danger' : 'public-sidecard-callout-ok'">
            <span class="public-sidecard-callout-label">Current signal</span>
            <strong>{{ publicActiveIncidentCount > 0 ? 'Attention required' : 'No active outage' }}</strong>
            <span>{{ publicActiveIncidentCount > 0 ? 'Review the active timeline and notification delivery.' : 'All recorded public incidents are resolved.' }}</span>
          </div>
          <div class="public-sidecard-detail"><span>Most affected monitor</span><strong>{{ publicTopIncidentMonitor }}</strong></div>
        </div>
        <div v-else class="public-sidecard-stats">
          <div class="public-side-stat"><span>Monitors covered</span><strong>{{ publicMonitorCount }}</strong></div>
          <div class="public-side-stat"><span>Operational</span><strong class="text-status-up">{{ publicOperationalCount }}</strong></div>
          <div class="public-side-stat"><span>Down / paused</span><strong :class="publicDownMonitorCount > 0 ? 'text-status-down' : 'text-text-main'">{{ publicDownMonitorCount }} / {{ publicPausedMonitorCount }}</strong></div>
          <div class="public-sidecard-callout public-sidecard-callout-neutral">
            <span class="public-sidecard-callout-label">Coverage signal</span>
            <strong>{{ publicReport ? `${publicReport.uptimePercent.toFixed(2)}% uptime` : 'Waiting for data' }}</strong>
            <span>{{ publicReport ? `${publicReport.totalChecks} checks in ${reportPeriodLabel.toLowerCase()}.` : 'Monitor summary will appear after the report loads.' }}</span>
          </div>
          <div class="public-sidecard-detail"><span>Average response</span><strong>{{ publicReport ? `${publicReport.avgLatencyMs.toFixed(0)} ms` : '—' }}</strong></div>
        </div>
        <router-link to="/public-monitoring" class="public-sidecard-link">Open public monitoring <span>→</span></router-link>
      </aside>
    </section>

    <template v-else>

    <!-- Overview Stat Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-4 gap-4">
      <template v-if="reportStore.isLoading">
        <SkeletonCard v-for="i in 4" :key="i" />
      </template>
      <template v-else>
        <div class="bg-surface border border-subtle rounded-xl p-5 space-y-1">
          <p class="text-[10px] font-mono text-text-secondary uppercase tracking-wider">Avg SLA Uptime</p>
          <p class="text-2xl font-extrabold text-status-up font-mono">{{ reportStore.filteredAvgSlaUptime.toFixed(2) }}%</p>
          <p class="text-[11px] text-text-muted font-mono">Target: 99.50%</p>
        </div>

        <div class="bg-surface border border-subtle rounded-xl p-5 space-y-1">
          <p class="text-[10px] font-mono text-text-secondary uppercase tracking-wider">Total Outage Events</p>
          <p class="text-2xl font-extrabold text-amber-400 font-mono">{{ reportStore.filteredTotalOutageEvents }}</p>
          <p class="text-[11px] text-text-muted font-mono">Avg MTTR: {{ reportStore.avgMttrMinutes.toFixed(0) }}m</p>
        </div>

        <div class="bg-surface border border-subtle rounded-xl p-5 space-y-1">
          <p class="text-[10px] font-mono text-text-secondary uppercase tracking-wider">Alert Delivery</p>
          <p class="text-2xl font-extrabold text-brand-periwinkle font-mono">{{ reportStore.alertDeliveryRate }}%</p>
          <p class="text-[11px] text-text-muted font-mono">WA + Telegram</p>
        </div>

        <div class="bg-surface border border-status-warning/30 rounded-xl p-5 space-y-1">
          <p class="text-[10px] font-mono text-amber-400/80 uppercase tracking-wider">Recurring Issues</p>
          <p class="text-2xl font-extrabold text-amber-400 font-mono">{{ reportStore.filteredFlapDevices.length }}</p>
          <p class="text-[11px] text-text-muted font-mono">≥5 downs / 7 days</p>
        </div>
      </template>
    </div>
    <!-- Report Section Tabs Selector -->
    <div class="flex border-b border-subtle gap-6 text-sm font-mono pb-0.5">
      <button
        @click="activeTab = 'downtime'"
        class="pb-3 border-b-2 font-semibold transition-all relative text-xs"
        :class="activeTab === 'downtime' ? 'border-brand-periwinkle text-text-main' : 'border-transparent text-text-muted hover:text-text-secondary'"
      >
        Downtime by Device
      </button>
      <button
        @click="activeTab = 'recurring'"
        class="pb-3 border-b-2 font-semibold transition-all relative text-xs"
        :class="activeTab === 'recurring' ? 'border-amber-500 text-text-main' : 'border-transparent text-text-muted hover:text-text-secondary'"
      >
        Recurring Issues
      </button>
      <button
        @click="activeTab = 'active_incidents'"
        class="pb-3 border-b-2 font-semibold transition-all relative text-xs"
        :class="activeTab === 'active_incidents' ? 'border-red-500 text-text-main' : 'border-transparent text-text-muted hover:text-text-secondary'"
      >
        Active Incidents
      </button>
    </div>

    <!-- Tab 1: Downtime by Device -->
    <div v-if="activeTab === 'downtime'" class="grid grid-cols-1 xl:grid-cols-3 gap-6 items-start">
      <!-- Downtime Table (2/3 width) -->
      <div class="xl:col-span-2 space-y-4">
        <SkeletonTable v-if="reportStore.isLoading" :rows="7" :cols="5" />
        <div v-else class="responsive-table-wrap bg-surface border border-subtle rounded-xl overflow-hidden">
          <div class="px-5 py-3.5 border-b border-subtle flex items-center justify-between">
            <h3 class="text-xs font-bold uppercase tracking-wider text-text-secondary font-mono">Downtime by Device</h3>
            <span class="text-[10px] text-text-muted font-mono">Sorted by most-down first</span>
          </div>
          <table class="responsive-data-table w-full text-left text-xs text-text-secondary">
            <thead class="bg-card font-mono text-[10px] uppercase text-text-muted border-b border-subtle">
              <tr>
                <th class="py-3 px-4">Device</th>
                <th class="py-3 px-4">Location</th>
                <th class="py-3 px-4 text-center">Down Count</th>
                <th class="py-3 px-4">Total Downtime</th>
                <th class="py-3 px-4">Last Down</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-subtle">
              <tr v-if="reportStore.isLoading">
                <td colspan="5" class="p-0 border-0">
                  <SkeletonTable :rows="5" :cols="5" />
                </td>
              </tr>
              <template v-else-if="paginatedDowntimeRows.length > 0">
                <tr
                  v-for="(row, idx) in paginatedDowntimeRows"
                  :key="row.deviceId"
                class="hover:bg-card transition-colors"
                :class="{ 'border-l-2 border-l-[#F5A65B]': row.downCount >= 5 }"
              >
                <td data-label="Device" class="py-3 px-4">
                  <div class="flex items-center gap-2">
                    <span
                      class="text-[10px] font-mono font-bold w-5 h-5 rounded-full flex items-center justify-center"
                      :class="idx === 0 ? 'bg-status-down/20 text-status-down' : idx === 1 ? 'bg-amber-500/20 text-amber-400' : 'bg-subtle text-text-secondary'"
                    >
                      {{ (downtimePage - 1) * downtimePageSize + idx + 1 }}
                    </span>
                    <div>
                      <p class="font-bold text-text-main">{{ row.deviceName }}</p>
                      <p class="text-[10px] font-mono text-text-muted">{{ row.deviceType }}</p>
                    </div>
                  </div>
                </td>
                <td data-label="Location" class="py-3 px-4 text-text-secondary max-w-[150px] truncate text-[11px]">{{ row.location }}</td>
                <td data-label="Down Count" class="py-3 px-4 text-center">
                  <span
                    class="text-sm font-extrabold font-mono"
                    :class="row.downCount >= 6 ? 'text-status-down' : row.downCount >= 3 ? 'text-amber-400' : 'text-text-secondary'"
                  >
                    {{ row.downCount }}
                  </span>
                </td>
                <td data-label="Total Downtime" class="py-3 px-4 font-mono font-semibold text-status-down">
                  {{ reportStore.formatDowntime(row.totalDowntimeMinutes) }}
                </td>
                <td data-label="Last Down" class="py-3 px-4 font-mono text-text-secondary text-[11px]">{{ row.lastDown }}</td>
              </tr>
              </template>
              <tr v-else-if="reportStore.filteredRows.length === 0">
                <td colspan="5" class="py-10 text-center text-text-muted text-xs">
                  No downtime events for selected filters
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <PaginationControl
          v-model:current-page="downtimePage"
          v-model:page-size="downtimePageSize"
          :total="reportStore.filteredRows.length"
        />
      </div>

      <!-- Bar Chart (1/3 width) -->
      <div class="xl:col-span-1">
        <template v-if="reportStore.isLoading">
          <div class="bg-surface border border-subtle rounded-xl p-5 space-y-3">
            <Skeleton height="0.75rem" width="60%" />
            <Skeleton v-for="i in 7" :key="i" height="1.25rem" :width="`${90 - i * 8}%`" />
          </div>
        </template>
        <DowntimeBarChart
          v-else
          :data="reportStore.filteredRows"
          :period="reportStore.period"
        />
      </div>
    </div>

    <!-- Tab 2: Recurring Issues -->
    <div v-if="activeTab === 'recurring'" class="grid grid-cols-1 xl:grid-cols-3 gap-6 items-start">
      <!-- Recurring Table (2/3 width) -->
      <div class="xl:col-span-2 space-y-4">
        <SkeletonTable v-if="reportStore.isLoading" :rows="2" :cols="5" />
        <template v-else>
          <RecurringIssuesTable :devices="paginatedFlapDevices" />
          <PaginationControl
            v-model:current-page="recurringPage"
            v-model:page-size="recurringPageSize"
            :total="reportStore.filteredFlapDevices.length"
          />
        </template>
      </div>

      <!-- Recurring Flap Chart (1/3 width) -->
      <div class="xl:col-span-1">
        <template v-if="reportStore.isLoading">
          <div class="bg-surface border border-subtle rounded-xl p-5 space-y-3">
            <Skeleton height="0.75rem" width="60%" />
            <Skeleton v-for="i in 5" :key="i" height="1.25rem" :width="`${85 - i * 10}%`" />
          </div>
        </template>
        <RecurringBarChart
          v-else
          :devices="reportStore.filteredFlapDevices"
        />
      </div>
    </div>

    <!-- Tab 3: Active Incidents -->
    <div v-if="activeTab === 'active_incidents'" class="grid grid-cols-1 xl:grid-cols-3 gap-6 items-start">
      <!-- Active Incidents Table (2/3 width) -->
      <div class="xl:col-span-2 space-y-4">
        <SkeletonTable v-if="reportStore.isLoading" :rows="4" :cols="8" />
        <div v-else class="bg-surface border border-subtle rounded-xl overflow-hidden">
          <div class="px-5 py-3.5 border-b border-subtle flex items-center justify-between">
            <h3 class="text-xs font-bold uppercase tracking-wider text-text-secondary font-mono">Active Incidents Report Table</h3>
            <span class="text-[10px] text-text-muted font-mono">Real-time incident ticket queue</span>
          </div>
          <div class="responsive-table-wrap overflow-x-auto">
            <table class="responsive-data-table w-full text-left text-xs text-text-secondary">
              <thead class="bg-card font-mono text-[10px] uppercase text-text-muted border-b border-subtle">
                <tr>
                  <th class="py-3 px-4">Ticket ID</th>
                  <th class="py-3 px-4">Device Name</th>
                  <th class="py-3 px-4">Type</th>
                  <th class="py-3 px-4">IP Address</th>
                  <th class="py-3 px-4">Duration</th>
                  <th class="py-3 px-4 text-center">Affected</th>
                  <th class="py-3 px-4">Status</th>
                  <th class="py-3 px-4 text-right">Action</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-subtle">
                <tr v-if="incidentStore.isLoading || reportStore.isLoading">
                  <td colspan="8" class="p-0 border-0">
                    <SkeletonTable :rows="5" :cols="8" />
                  </td>
                </tr>
                <template v-else-if="paginatedIncidents.length > 0">
                  <tr
                    v-for="inc in paginatedIncidents"
                    :key="inc.id"
                    class="hover:bg-card transition-colors"
                  >
                    <td data-label="Ticket ID" class="py-3 px-4 font-mono font-bold text-brand-periwinkle">{{ inc.id }}</td>
                    <td data-label="Device Name" class="py-3 px-4">
                      <div>
                        <p class="font-bold text-text-main">{{ inc.deviceName }}</p>
                        <p class="text-[10px] text-text-muted font-mono">{{ inc.location }}</p>
                      </div>
                    </td>
                    <td data-label="Type" class="py-3 px-4 text-text-secondary">{{ inc.deviceType }}</td>
                    <td data-label="IP Address" class="py-3 px-4 font-mono text-text-secondary">{{ inc.deviceIp }}</td>
                    <td data-label="Duration" class="py-3 px-4 font-mono font-semibold" :class="inc.status === 'ACTIVE' ? 'text-red-400' : 'text-text-secondary'">
                      {{ formatLiveDuration(inc.startedAt, inc.status, inc.duration) }}
                    </td>
                    <td data-label="Affected" class="py-3 px-4 text-center font-mono">
                      <span class="px-2 py-0.5 rounded text-[10px] bg-red-500/10 text-red-400 border border-red-500/20 font-bold">
                        {{ inc.affectedDevicesCount }}
                      </span>
                    </td>
                    <td data-label="Status" class="py-3 px-4">
                      <span
                        class="px-2 py-0.5 rounded text-[10px] font-mono font-semibold"
                        :class="inc.status === 'ACTIVE' ? 'bg-status-down/10 text-status-down border border-status-down/20' : 'bg-status-up/10 text-status-up border border-status-up/20'"
                      >
                        {{ inc.status }}
                      </span>
                    </td>
                    <td data-label="Action" class="py-3 px-4 text-right">
                      <router-link
                        :to="`/incidents/${inc.id}`"
                        class="px-2.5 py-1 rounded-lg bg-subtle hover:bg-hover text-xs font-semibold text-text-main transition-colors inline-block"
                      >
                        View Details
                      </router-link>
                    </td>
                  </tr>
                </template>
                <tr v-else-if="filteredIncidents.length === 0">
                  <td colspan="8" class="py-10 text-center text-text-muted text-xs">
                    No active or recent incidents reported
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <PaginationControl
          v-model:current-page="incidentsPage"
          v-model:page-size="incidentsPageSize"
          :total="filteredIncidents.length"
        />
      </div>

      <!-- Active Incidents Chart (1/3 width) -->
      <div class="xl:col-span-1">
        <template v-if="reportStore.isLoading">
          <div class="bg-surface border border-subtle rounded-xl p-5 space-y-3">
            <Skeleton height="0.75rem" width="60%" />
            <Skeleton v-for="i in 4" :key="i" height="1.25rem" :width="`${85 - i * 10}%`" />
          </div>
        </template>
        <ActiveIncidentsChart
          v-else
          :incidents="filteredIncidents"
        />
      </div>
    </div>

    </template>

    <PrintableSLAAudit
      v-if="isPrintRendered && reportScope !== 'PUBLIC_MONITOR'"
      :period="reportStore.period"
      :active-tab="activeTab"
      :avg-sla-uptime="reportStore.filteredAvgSlaUptime"
      :total-outage-events="reportStore.filteredTotalOutageEvents"
      :avg-mttr-minutes="reportStore.avgMttrMinutes"
      :alert-delivery-rate="reportStore.alertDeliveryRate"
      :rows="reportStore.filteredRows"
      :flap-devices="reportStore.filteredFlapDevices"
      :incidents="filteredIncidents"
    />
    <PrintablePublicMonitorReport
      v-if="isPrintRendered && reportScope === 'PUBLIC_MONITOR'"
      :tab="publicReportTab"
      :period-label="reportPeriodLabel"
      :report="publicReport"
      :incidents="publicPrintIncidents"
      :summaries="publicSummaries"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useReportStore } from '../stores/reportStore';
import { useIncidentStore } from '../stores/incidentStore';
import { useAuthStore } from '../stores/authStore';
import { publicMonitoringApi } from '../api';
import DowntimeBarChart from '../components/reports/DowntimeBarChart.vue';
import RecurringBarChart from '../components/reports/RecurringBarChart.vue';
import ActiveIncidentsChart from '../components/reports/ActiveIncidentsChart.vue';
import RecurringIssuesTable from '../components/reports/RecurringIssuesTable.vue';
import PrintableSLAAudit from '../components/reports/PrintableSLAAudit.vue';
import PrintablePublicMonitorReport from '../components/reports/PrintablePublicMonitorReport.vue';
import PaginationControl from '../components/common/PaginationControl.vue';
import SkeletonCard from '../components/common/SkeletonCard.vue';
import SkeletonTable from '../components/common/SkeletonTable.vue';
import Skeleton from '../components/common/Skeleton.vue';
import { FileText, Printer, Calendar, ChevronDown, FileSpreadsheet } from 'lucide-vue-next';
import { downloadCSV, downloadXLS } from '../utils/exportUtils';
import { publicMonitorTypeLabel } from '../utils/publicMonitorLabels';
import type { PublicMonitorIncident, PublicMonitorReport } from '../types';

const reportStore = useReportStore();
const incidentStore = useIncidentStore();
const authStore = useAuthStore();
const route = useRoute();
const router = useRouter();

const showExportDropdown = ref(false);
const activeTab = ref<'downtime' | 'recurring' | 'active_incidents'>('downtime');
const reportScope = ref<'DEVICE' | 'PUBLIC_MONITOR'>('DEVICE');
const reportScopes = [{ value: 'DEVICE' as const, label: 'DEVICE MONITORING' }, { value: 'PUBLIC_MONITOR' as const, label: 'PUBLIC MONITORING' }];
const publicIncidents = ref<PublicMonitorIncident[]>([]);
const publicPrintIncidents = ref<PublicMonitorIncident[]>([]);
const publicIncidentsLoading = ref(false);
const publicReport = ref<PublicMonitorReport | null>(null);
const publicReportTab = ref<'incidents' | 'summary'>('incidents');
const publicIncidentPage = ref(1);
const publicIncidentPageSize = ref(10);
const publicIncidentTotal = ref(0);
const publicActiveIncidentTotal = ref(0);
const publicResolvedIncidentTotal = ref(0);
const publicSummaryPage = ref(1);
const publicSummaryPageSize = ref(10);
const reportPeriodLabel = computed(() => reportStore.period === 'daily' ? 'Last 24 hours' : reportStore.period === 'weekly' ? 'Last 7 days' : reportStore.period === 'custom' ? 'Selected period' : 'Last 30 days');
const publicSummaries = computed(() => publicReport.value?.monitorSummaries || []);
const paginatedPublicSummaries = computed(() => {
  const start = (publicSummaryPage.value - 1) * publicSummaryPageSize.value;
  return publicSummaries.value.slice(start, start + publicSummaryPageSize.value);
});
const publicMonitorCount = computed(() => publicSummaries.value.length);
const publicOperationalCount = computed(() => publicSummaries.value.filter(item => item.status === 'UP').length);
const publicDownMonitorCount = computed(() => publicSummaries.value.filter(item => item.status === 'DOWN').length);
const publicPausedMonitorCount = computed(() => publicSummaries.value.filter(item => item.status === 'PAUSED' || !item.enabled).length);
const publicActiveIncidentCount = computed(() => publicActiveIncidentTotal.value);
const publicResolvedIncidentCount = computed(() => publicResolvedIncidentTotal.value);
const publicTopIncidentMonitor = computed(() => {
  const top = [...publicSummaries.value].sort((a, b) => b.incidentCount - a.incidentCount)[0];
  return top && top.incidentCount > 0 ? `${top.monitorName} (${top.incidentCount})` : 'No incidents recorded';
});

const downtimePage = ref(1);
const downtimePageSize = ref(10);

const recurringPage = ref(1);
const recurringPageSize = ref(10);

const incidentsPage = ref(1);
const incidentsPageSize = ref(10);

const paginatedDowntimeRows = computed(() => {
  const start = (downtimePage.value - 1) * downtimePageSize.value;
  return reportStore.filteredRows.slice(start, start + downtimePageSize.value);
});

const paginatedFlapDevices = computed(() => {
  const start = (recurringPage.value - 1) * recurringPageSize.value;
  return reportStore.filteredFlapDevices.slice(start, start + recurringPageSize.value);
});

const filteredIncidents = computed(() => {
  let filtered = incidentStore.incidents;

  // Filter by date range based on reportStore.period
  const now = new Date().getTime();
  let fromTime = 0;
  if (reportStore.period === 'daily') fromTime = now - 24 * 3600 * 1000;
  else if (reportStore.period === 'weekly') fromTime = now - 7 * 24 * 3600 * 1000;
  else if (reportStore.period === 'monthly') fromTime = now - 30 * 24 * 3600 * 1000;
  else if (reportStore.period === 'custom' && reportStore.startDate) {
    fromTime = new Date(reportStore.startDate).getTime();
  }
  let toTime = now;
  if (reportStore.period === 'custom' && reportStore.endDate) {
    const d = new Date(reportStore.endDate);
    d.setHours(23, 59, 59, 999);
    toTime = d.getTime();
  }
  
  if (fromTime > 0) {
    filtered = filtered.filter(inc => {
      const incTime = new Date(inc.startedAt || inc.startTime).getTime();
      return incTime >= fromTime && incTime <= toTime;
    });
  }

  if (reportStore.locationFilter !== 'All') {
    filtered = filtered.filter(inc => inc.location && inc.location.includes(reportStore.locationFilter));
  }
  if (reportStore.deviceTypeFilter !== 'All') {
    filtered = filtered.filter(inc => inc.deviceType === reportStore.deviceTypeFilter);
  }
  // Always filter only ACTIVE incidents for this tab
  filtered = filtered.filter(inc => inc.status === 'ACTIVE');
  return filtered;
});

const paginatedIncidents = computed(() => {
  const start = (incidentsPage.value - 1) * incidentsPageSize.value;
  return filteredIncidents.value.slice(start, start + incidentsPageSize.value);
});

// Reset page numbers to 1 when filters or tabs change
watch(
  () => [
    activeTab.value,
    reportStore.period,
    reportStore.deviceTypeFilter,
    reportStore.locationFilter
  ],
  () => {
    downtimePage.value = 1;
    recurringPage.value = 1;
    incidentsPage.value = 1;
  }
);

const now = ref(Date.now());
let durationInterval: any = null;
const isPrintRendered = ref(false);

function setPeriod(p: 'daily' | 'weekly' | 'monthly' | 'custom') {
  reportStore.period = p;
  if (reportScope.value === 'PUBLIC_MONITOR') {
    const shouldFetchImmediately = publicIncidentPage.value === 1;
    publicIncidentPage.value = 1;
    publicSummaryPage.value = 1;
    if (shouldFetchImmediately) fetchPublicIncidents();
  }
  else {
    reportStore.fetchReport();
    // Refetch all incidents so we can filter them by date locally.
    incidentStore.fetchIncidents({ status: 'ACTIVE', source: 'DEVICE' });
  }
}

function getPublicReportParams(): { period: string; startDate?: string; endDate?: string } {
  const params: { period: string; startDate?: string; endDate?: string } = { period: reportStore.period };
  if (reportStore.period === 'custom') {
    params.startDate = reportStore.startDate;
    params.endDate = reportStore.endDate;
  }
  return params;
}

async function fetchPublicIncidents() {
  publicIncidentsLoading.value = true;
  try {
    const params = getPublicReportParams();
    const [response, report, activeResponse, resolvedResponse] = await Promise.all([
      publicMonitoringApi.getIncidents({ ...params, page: publicIncidentPage.value, page_size: publicIncidentPageSize.value }),
      publicMonitoringApi.getReport(params),
      publicMonitoringApi.getIncidents({ ...params, status: 'ACTIVE', page: 1, page_size: 1 }),
      publicMonitoringApi.getIncidents({ ...params, status: 'RESOLVED', page: 1, page_size: 1 })
    ]);
    publicIncidents.value = response.items || response.data || [];
    publicIncidentTotal.value = response.total || 0;
    publicActiveIncidentTotal.value = activeResponse.total || 0;
    publicResolvedIncidentTotal.value = resolvedResponse.total || 0;
    publicReport.value = report;
  } catch (err) {
    console.error('Failed to fetch public monitoring incidents:', err);
    publicIncidents.value = [];
    publicIncidentTotal.value = 0;
    publicActiveIncidentTotal.value = 0;
    publicResolvedIncidentTotal.value = 0;
    publicReport.value = null;
  } finally {
    publicIncidentsLoading.value = false;
  }
}

async function fetchAllPublicIncidents() {
  const params = getPublicReportParams();
  const first = await publicMonitoringApi.getIncidents({ ...params, page: 1, page_size: 100 });
  const all = [...(first.items || first.data || [])] as PublicMonitorIncident[];
  const totalPages = first.totalPages || 1;
  for (let page = 2; page <= totalPages; page += 1) {
    const response = await publicMonitoringApi.getIncidents({ ...params, page, page_size: 100 });
    all.push(...(response.items || response.data || []));
  }
  return all;
}

async function exportReport(format: 'csv' | 'xls') {
  if (reportScope.value !== 'PUBLIC_MONITOR') {
    reportStore.exportData(format, activeTab.value, incidentStore.incidents);
    return;
  }
  const incidents = publicReportTab.value === 'incidents' ? await fetchAllPublicIncidents() : [];
  const rows = publicReportTab.value === 'summary'
    ? (publicReport.value?.monitorSummaries || []).map((summary) => [summary.monitorName, publicMonitorTypeLabel(summary.monitorType), summary.targetUrl, summary.groupName || 'Ungrouped', summary.status, `${summary.uptimePercent.toFixed(2)}%`, summary.totalChecks, `${summary.avgLatencyMs.toFixed(0)}ms`, summary.incidentCount])
    : incidents.map((incident) => [formatPublicIncidentId(incident.id), incident.monitorName, incident.targetUrl, incident.status, formatReportDate(incident.startedAt), formatIncidentDuration(incident), incident.lastError || incident.firstError || incident.resolutionReason || 'Recovered']);
  const headers = publicReportTab.value === 'summary' ? ['Monitor', 'Type', 'Target', 'Group', 'Status', 'Uptime', 'Total checks', 'Avg response', 'Incidents'] : ['Incident', 'Monitor', 'Target', 'Status', 'Started', 'Duration', 'Reason'];
  const filename = `sanoc-public-monitor-${publicReportTab.value}-${reportStore.period}-${new Date().toISOString().slice(0, 10)}`;
  if (format === 'csv') downloadCSV(filename, headers, rows);
  else downloadXLS(filename, 'SANOC Public Monitoring Report', headers, rows);
}

function formatReportDate(value: string) {
  return new Date(value).toLocaleString('en-GB', { dateStyle: 'short', timeStyle: 'short' });
}

function formatPublicIncidentId(id: string) {
  return id.toLowerCase().startsWith('pinc-') ? id.toUpperCase() : id;
}

function formatIncidentDuration(incident: PublicMonitorIncident) {
  if (incident.status === 'ACTIVE') return formatLiveDuration(incident.startedAt, incident.status, 'Ongoing');
  const totalSeconds = Math.max(0, incident.durationSeconds || 0);
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  return [days ? `${days}d` : '', hours ? `${hours}h` : '', minutes ? `${minutes}m` : '0m'].filter(Boolean).join(' ');
}

function formatLiveDuration(startedAtStr: string | undefined, status: string, durationStr?: string) {
  if (status === 'RESOLVED') {
    return durationStr || 'Resolved';
  }
  if (!startedAtStr) return durationStr || 'Ongoing';
  const start = new Date(startedAtStr);
  const diffMs = now.value - start.getTime();
  if (diffMs < 0) return '0s (ongoing)';
  
  const totalSecs = Math.floor(diffMs / 1000);
  const days = Math.floor(totalSecs / 86400);
  const hours = Math.floor((totalSecs % 86400) / 3600);
  const mins = Math.floor((totalSecs % 3600) / 60);
  const secs = totalSecs % 60;
  
  const parts = [];
  if (days > 0) parts.push(`${days}d`);
  if (hours > 0) parts.push(`${hours}h`);
  if (mins > 0) parts.push(`${mins}m`);
  if (secs > 0 || parts.length === 0) parts.push(`${secs}s`);
  
  return parts.join(' ') + ' (ongoing)';
}

async function handlePrintPDF() {
  if (reportScope.value === 'PUBLIC_MONITOR') {
    await fetchPublicIncidents();
    publicPrintIncidents.value = publicReportTab.value === 'incidents' ? await fetchAllPublicIncidents() : [];
    const originalTitle = document.title;
    const tabSlug = publicReportTab.value === 'summary' ? 'monitor-summary' : 'incident-reports';
    document.title = `sanoc-public-monitor-${tabSlug}-${reportStore.period}-${new Date().toISOString().slice(0, 10)}`;
    try {
      isPrintRendered.value = true;
      await nextTick();
      await new Promise((resolve) => setTimeout(resolve, 50));
      window.print();
    } finally {
      isPrintRendered.value = false;
      publicPrintIncidents.value = [];
      document.title = originalTitle;
    }
    return;
  }
  reportStore.isLoading = true;
  const originalTitle = document.title;
  const dateStr = new Date().toISOString().slice(0, 10);
  const tabSlug = activeTab.value.replace(/_/g, '-');
  document.title = `sanoc-pdf-${tabSlug}-${reportStore.period}-${dateStr}`;

  try {
    await reportStore.fetchReport();
    if (activeTab.value === 'active_incidents') {
      await incidentStore.fetchIncidents({ status: 'ACTIVE', source: 'DEVICE' });
    }
    isPrintRendered.value = true;
    await nextTick();
    await new Promise((resolve) => setTimeout(resolve, 50));
    window.print();
  } catch (err) {
    console.error('Failed to print SLA report:', err);
  } finally {
    reportStore.isLoading = false;
    isPrintRendered.value = false;
    document.title = originalTitle;
  }
}

onMounted(() => {
  const requestedScope = route.query.source as string;
  if (requestedScope === 'DEVICE' || requestedScope === 'PUBLIC_MONITOR') {
    reportScope.value = requestedScope;
  }
  if (reportScope.value === 'PUBLIC_MONITOR') fetchPublicIncidents();
  else {
    reportStore.fetchReport();
    incidentStore.fetchIncidents({ status: 'ACTIVE', source: 'DEVICE' });
  }
  durationInterval = setInterval(() => {
    now.value = Date.now();
  }, 1000);
});

watch(reportScope, (scope) => {
  if (scope === 'PUBLIC_MONITOR') {
    const shouldFetchImmediately = publicIncidentPage.value === 1;
    publicIncidentPage.value = 1;
    publicSummaryPage.value = 1;
    if (shouldFetchImmediately) fetchPublicIncidents();
  }
  else if (scope === 'DEVICE') incidentStore.fetchIncidents({ status: 'ACTIVE', source: 'DEVICE' });
});

watch([publicIncidentPage, publicIncidentPageSize], () => {
  if (reportScope.value === 'PUBLIC_MONITOR') fetchPublicIncidents();
});

watch(() => publicSummaries.value.length, (total) => {
  const maxPage = Math.max(1, Math.ceil(total / publicSummaryPageSize.value));
  if (publicSummaryPage.value > maxPage) publicSummaryPage.value = maxPage;
});

onUnmounted(() => {
  if (durationInterval) clearInterval(durationInterval);
});
</script>

<style scoped>
.source-tab { @apply rounded-lg border border-transparent px-3 py-2 text-[10px] font-mono font-semibold tracking-wide text-text-muted transition-colors hover:bg-surface hover:text-text-main; }
.source-tab-active { @apply border-brand-periwinkle/30 bg-brand-periwinkle/10 text-brand-periwinkle; }
.public-report-tabs { @apply flex gap-1 border-b border-subtle; }
.public-report-tab { @apply border-b-2 border-transparent px-3 py-2.5 text-xs font-semibold text-text-muted transition-colors hover:text-text-main; }
.public-report-tab-active { @apply border-brand-periwinkle text-brand-periwinkle; }
.public-report-card { @apply rounded-xl border border-subtle bg-surface p-4; }
.public-report-card span { @apply block text-[10px] font-mono uppercase tracking-wider text-text-muted; }
.public-report-card strong { @apply mt-2 block text-xl font-extrabold font-mono tabular-nums text-text-main; }
.public-report-card small { @apply mt-1 block text-[10px] text-text-muted; }
.public-report-layout { @apply grid grid-cols-1 items-start gap-5 xl:grid-cols-[minmax(0,1fr)_280px]; }
.public-report-main { @apply min-w-0; }
.public-report-sidecard { @apply sticky top-5 rounded-xl border border-subtle bg-surface p-5; }
.public-sidecard-header { @apply flex items-start justify-between gap-3 border-b border-subtle pb-4; }
.public-sidecard-eyebrow { @apply text-[10px] font-mono uppercase tracking-wider text-brand-periwinkle; }
.public-sidecard-title { @apply mt-1 text-sm font-bold text-text-main; }
.public-sidecard-live { @apply inline-flex items-center gap-1.5 rounded-full border border-status-up/25 bg-status-up/10 px-2 py-1 text-[9px] font-mono font-semibold text-status-up; }
.public-sidecard-live-dot { @apply h-1.5 w-1.5 rounded-full bg-status-up; }
.public-sidecard-period { @apply mt-4 text-[10px] font-mono text-text-muted; }
.public-sidecard-stats { @apply mt-3 space-y-3; }
.public-side-stat { @apply flex items-center justify-between border-b border-subtle/70 pb-3 text-xs; }
.public-side-stat span { @apply text-text-secondary; }
.public-side-stat strong { @apply font-mono text-text-main; }
.public-side-stat small { @apply text-[10px] font-normal text-text-muted; }
.public-sidecard-callout { @apply rounded-lg border p-3; }
.public-sidecard-callout-danger { @apply border-status-down/30 bg-status-down/10; }
.public-sidecard-callout-ok { @apply border-status-up/30 bg-status-up/10; }
.public-sidecard-callout-neutral { @apply border-brand-periwinkle/25 bg-brand-periwinkle/10; }
.public-sidecard-callout-label { @apply block text-[9px] font-mono uppercase tracking-wider text-text-muted; }
.public-sidecard-callout strong { @apply mt-1 block text-xs text-text-main; }
.public-sidecard-callout span:last-child { @apply mt-1 block text-[10px] leading-4 text-text-secondary; }
.public-sidecard-detail { @apply flex items-start justify-between gap-3 text-[10px]; }
.public-sidecard-detail span { @apply text-text-muted; }
.public-sidecard-detail strong { @apply max-w-[150px] text-right font-mono text-text-main; }
.public-sidecard-link { @apply mt-5 flex items-center justify-between border-t border-subtle pt-4 text-[10px] font-semibold text-brand-periwinkle transition-colors hover:text-text-main; }
.public-summary-layout { @apply block; }
</style>
