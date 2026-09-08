<template>
  <Teleport to="body">
    <section id="printable-public-incidents-list" class="public-monitor-print-page hidden bg-white p-0 font-sans text-slate-900 print:block">
      <header class="public-incident-print-header">
        <div>
          <p class="public-incident-print-overline">SANOC · PUBLIC MONITORING</p>
          <h1>Public Monitoring Incident Reports</h1>
          <p>Incident list generated {{ printedAt }}</p>
        </div>
        <div class="public-incident-print-badge">{{ incidents.length }} INCIDENTS</div>
      </header>

      <table class="public-incident-print-table">
        <colgroup><col class="col-incident"><col class="col-monitor"><col class="col-target"><col class="col-status"><col class="col-started"><col class="col-duration"><col class="col-reason"></colgroup>
        <thead><tr><th>Incident</th><th>Monitor</th><th>Target</th><th>Status</th><th>Started</th><th>Duration</th><th>Reason</th></tr></thead>
        <tbody>
          <tr v-for="incident in incidents" :key="incident.id">
            <td class="public-incident-print-mono">{{ formatIncidentId(incident.id) }}</td>
            <td>{{ incident.deviceName }}</td>
            <td>{{ incident.targetUrl || incident.deviceIp || 'Host-based check' }}</td>
            <td>{{ incident.status }}</td>
            <td>{{ formatDate(incident.startedAt || incident.startTime) }}</td>
            <td>{{ incident.duration || '—' }}</td>
            <td>{{ incident.notes?.[0] || incident.timeline?.[incident.timeline.length - 1]?.description || 'Recovered' }}</td>
          </tr>
          <tr v-if="!incidents.length"><td colspan="7" class="public-incident-print-empty">No public incidents found.</td></tr>
        </tbody>
      </table>
    </section>
  </Teleport>
</template>

<script setup lang="ts">
import type { Incident } from '../../types';

defineProps<{ incidents: Incident[] }>();

const printedAt = new Date().toLocaleString('en-GB', { dateStyle: 'medium', timeStyle: 'short' });

function formatIncidentId(id: string) {
  return id.toLowerCase().startsWith('pinc-') ? id.toUpperCase() : id;
}

function formatDate(value?: string) {
  if (!value) return '—';
  return new Date(value).toLocaleString('en-GB', { dateStyle: 'short', timeStyle: 'short' });
}
</script>

<style>
#printable-public-incidents-list { width: 100%; color: #0f172a; }
.public-monitor-print-page { page: public-monitor; }
@page public-monitor { size: A4 landscape; margin: 10mm; }
.public-incident-print-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 24px; border-bottom: 2px solid #0f172a; padding-bottom: 16px; margin-bottom: 18px; }
.public-incident-print-overline { margin: 0 0 4px; font: 700 10px/1.2 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .12em; color: #475569; }
.public-incident-print-header h1 { margin: 0; font-size: 21px; line-height: 1.2; font-weight: 800; }
.public-incident-print-header p:last-child { margin: 6px 0 0; color: #64748b; font-size: 11px; }
.public-incident-print-badge { border: 1px solid #cbd5e1; border-radius: 5px; background: #f8fafc; padding: 8px 10px; font: 700 9px/1.2 ui-monospace, SFMono-Regular, Menlo, monospace; color: #334155; white-space: nowrap; }
.public-incident-print-table { width: 100%; border-collapse: collapse; table-layout: fixed; font-size: 9px; }
.public-incident-print-table .col-incident { width: 15%; }
.public-incident-print-table .col-monitor { width: 12%; }
.public-incident-print-table .col-target { width: 20%; }
.public-incident-print-table .col-status { width: 9%; }
.public-incident-print-table .col-started { width: 12%; }
.public-incident-print-table .col-duration { width: 10%; }
.public-incident-print-table .col-reason { width: 22%; }
.public-incident-print-table th, .public-incident-print-table td { border: 1px solid #cbd5e1; padding: 7px 6px; text-align: left; vertical-align: top; overflow-wrap: anywhere; }
.public-incident-print-table th { background: #e2e8f0; color: #334155; font: 700 8px/1.2 ui-monospace, SFMono-Regular, Menlo, monospace; text-transform: uppercase; }
.public-incident-print-mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-weight: 700; }
.public-incident-print-empty { padding: 24px !important; text-align: center !important; color: #64748b; }
@media print {
  #printable-public-incidents-list { display: block !important; visibility: visible !important; position: static !important; width: 100% !important; max-width: 100% !important; min-height: calc(100vh - 20mm); background: #ffffff !important; }
  .public-incident-print-table tr { page-break-inside: avoid; }
}
</style>
