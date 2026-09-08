<template>
  <Teleport to="body">
    <section id="public-monitor-report" class="public-monitor-print-page hidden bg-white p-0 font-sans text-slate-900 print:block">
      <header class="public-print-header">
        <div>
          <p class="public-print-overline">SANOC · PUBLIC MONITORING</p>
          <h1>{{ tab === 'incidents' ? 'Public Monitoring Incident Reports' : 'Public Monitor Summary' }}</h1>
          <p class="public-print-subtitle">{{ periodLabel }} · Generated {{ printedAt }}</p>
        </div>
        <div class="public-print-badge">{{ tab === 'incidents' ? 'INCIDENT REPORTS' : 'MONITOR SUMMARY' }}</div>
      </header>

      <div class="public-print-metrics">
        <div><span>Total checks</span><strong>{{ report?.totalChecks ?? 0 }}</strong></div>
        <div><span>Uptime</span><strong>{{ report ? `${report.uptimePercent.toFixed(2)}%` : '—' }}</strong></div>
        <div><span>Incidents</span><strong>{{ report?.incidentCount ?? 0 }}</strong></div>
        <div><span>Average response</span><strong>{{ report ? `${report.avgLatencyMs.toFixed(0)} ms` : '—' }}</strong></div>
      </div>

      <table v-if="tab === 'incidents'" class="public-print-table">
        <colgroup><col class="col-incident"><col class="col-monitor"><col class="col-target"><col class="col-status"><col class="col-started"><col class="col-duration"><col class="col-reason"></colgroup>
        <thead><tr><th>Incident</th><th>Monitor</th><th>Target</th><th>Status</th><th>Started</th><th>Duration</th><th>Reason</th></tr></thead>
        <tbody>
          <tr v-for="incident in incidents" :key="incident.id">
            <td class="public-print-mono">{{ formatIncidentId(incident.id) }}</td>
            <td>{{ incident.monitorName }}</td>
            <td>{{ incident.targetUrl || 'Host-based check' }}</td>
            <td>{{ incident.status }}</td>
            <td>{{ formatDate(incident.startedAt) }}</td>
            <td>{{ formatDuration(incident) }}</td>
            <td>{{ incident.lastError || incident.firstError || incident.resolutionReason || 'Recovered' }}</td>
          </tr>
          <tr v-if="!incidents.length"><td colspan="7" class="public-print-empty">No public incidents found for this period.</td></tr>
        </tbody>
      </table>

      <table v-else class="public-print-table">
        <colgroup><col class="col-monitor"><col class="col-type"><col class="col-target"><col class="col-group"><col class="col-status"><col class="col-uptime"><col class="col-checks"><col class="col-response"><col class="col-incidents"></colgroup>
        <thead><tr><th>Monitor</th><th>Type</th><th>Target</th><th>Group</th><th>Status</th><th>Uptime</th><th>Checks</th><th>Average response</th><th>Incidents</th></tr></thead>
        <tbody>
          <tr v-for="summary in summaries" :key="summary.monitorId">
            <td>{{ summary.monitorName }}</td>
            <td>{{ publicMonitorTypeLabel(summary.monitorType) }}</td>
            <td>{{ summary.targetUrl || 'Host-based check' }}</td>
            <td>{{ summary.groupName || 'Ungrouped' }}</td>
            <td>{{ summary.status }}</td>
            <td>{{ summary.uptimePercent.toFixed(2) }}%</td>
            <td>{{ summary.totalChecks }}</td>
            <td>{{ summary.avgLatencyMs.toFixed(0) }} ms</td>
            <td>{{ summary.incidentCount }}</td>
          </tr>
          <tr v-if="!summaries.length"><td colspan="9" class="public-print-empty">No monitor summary data is available for this period.</td></tr>
        </tbody>
      </table>
    </section>
  </Teleport>
</template>

<script setup lang="ts">
import type { PublicMonitorIncident, PublicMonitorReport, PublicMonitorSummary } from '../../types';
import { publicMonitorTypeLabel } from '../../utils/publicMonitorLabels';

defineProps<{
  tab: 'incidents' | 'summary';
  periodLabel: string;
  report: PublicMonitorReport | null;
  incidents: PublicMonitorIncident[];
  summaries: PublicMonitorSummary[];
}>();

const printedAt = new Date().toLocaleString('en-GB', { dateStyle: 'medium', timeStyle: 'short' });

function formatIncidentId(id: string) {
  return id.toLowerCase().startsWith('pinc-') ? id.toUpperCase() : id;
}

function formatDate(value: string) {
  return new Date(value).toLocaleString('en-GB', { dateStyle: 'short', timeStyle: 'short' });
}

function formatDuration(incident: PublicMonitorIncident) {
  if (incident.status === 'ACTIVE') return 'Ongoing';
  const seconds = Math.max(0, incident.durationSeconds || 0);
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return [days ? `${days}d` : '', hours ? `${hours}h` : '', minutes ? `${minutes}m` : '0m'].filter(Boolean).join(' ');
}
</script>

<style>
#public-monitor-report { width: 100%; color: #0f172a; }
.public-monitor-print-page { page: public-monitor; }
@page public-monitor { size: A4 landscape; margin: 10mm; }
.public-print-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 24px; border-bottom: 2px solid #0f172a; padding-bottom: 16px; margin-bottom: 18px; }
.public-print-overline { margin: 0 0 4px; font: 700 10px/1.2 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .12em; color: #475569; }
.public-print-header h1 { margin: 0; font-size: 21px; line-height: 1.2; font-weight: 800; }
.public-print-subtitle { margin: 6px 0 0; color: #64748b; font-size: 11px; }
.public-print-badge { border: 1px solid #cbd5e1; border-radius: 5px; background: #f8fafc; padding: 8px 10px; font: 700 9px/1.2 ui-monospace, SFMono-Regular, Menlo, monospace; color: #334155; white-space: nowrap; }
.public-print-metrics { display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px; margin-bottom: 18px; }
.public-print-metrics > div { border: 1px solid #cbd5e1; border-radius: 5px; padding: 10px; background: #f8fafc; }
.public-print-metrics span { display: block; font: 700 9px/1.2 ui-monospace, SFMono-Regular, Menlo, monospace; color: #64748b; text-transform: uppercase; }
.public-print-metrics strong { display: block; margin-top: 5px; font: 800 16px/1.2 ui-monospace, SFMono-Regular, Menlo, monospace; }
.public-print-table { width: 100%; border-collapse: collapse; font-size: 9px; table-layout: fixed; }
.public-print-table .col-incident { width: 15%; }
.public-print-table .col-monitor { width: 12%; }
.public-print-table .col-target { width: 20%; }
.public-print-table .col-status { width: 9%; }
.public-print-table .col-started { width: 12%; }
.public-print-table .col-duration { width: 10%; }
.public-print-table .col-reason { width: 22%; }
.public-print-table .col-type { width: 10%; }
.public-print-table .col-group { width: 11%; }
.public-print-table .col-uptime { width: 9%; }
.public-print-table .col-checks { width: 8%; }
.public-print-table .col-response { width: 12%; }
.public-print-table .col-incidents { width: 9%; }
.public-print-table th, .public-print-table td { border: 1px solid #cbd5e1; padding: 7px 6px; text-align: left; vertical-align: top; overflow-wrap: anywhere; }
.public-print-table th { background: #e2e8f0; color: #334155; font: 700 8px/1.2 ui-monospace, SFMono-Regular, Menlo, monospace; text-transform: uppercase; }
.public-print-mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-weight: 700; }
.public-print-empty { padding: 24px !important; text-align: center !important; color: #64748b; }
@media print {
  #public-monitor-report { display: block !important; visibility: visible !important; position: static !important; width: 100% !important; max-width: 100% !important; min-height: calc(100vh - 20mm); background: #ffffff !important; }
  .public-print-table { page-break-inside: auto; }
  .public-print-table tr { page-break-inside: avoid; }
}
</style>
