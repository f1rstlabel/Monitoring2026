<template>
  <div class="chart-shell">
    <div class="chart-toolbar">
      <div class="chart-mode-switcher" role="group" aria-label="Chart view mode">
        <button v-for="mode in viewModes" :key="mode.value" type="button" class="chart-mode-button" :class="viewMode === mode.value ? 'chart-mode-active' : ''" :aria-pressed="viewMode === mode.value" :title="mode.label" @click="viewMode = mode.value">
          <component :is="mode.icon" class="h-3.5 w-3.5 shrink-0" /><span class="chart-mode-label">{{ mode.label }}</span>
        </button>
      </div>
      <div class="chart-live-state"><span class="live-dot" :class="livePoints.length ? 'live-dot-active' : ''" />{{ livePoints.length ? 'Live checks included' : 'Historical data' }}</div>
    </div>

    <div v-if="!hasData" class="chart-empty">
      <Activity class="mb-2 h-8 w-8 text-text-muted" />
      <p class="text-xs font-semibold text-text-secondary">No chart data available</p>
      <p class="mt-1 text-[10px] text-text-muted">Data will appear after the endpoint completes several health checks.</p>
    </div>
    <apexchart v-else :key="viewMode" :type="chartType" height="280" :options="options" :series="series" />

    <div v-if="viewMode === 'donut' && hasData" class="chart-legend">
      <span><i class="legend-dot bg-status-up" />UP checks <strong>{{ statusCounts.up }}</strong></span>
      <span><i class="legend-dot bg-status-down" />DOWN checks <strong>{{ statusCounts.down }}</strong></span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import VueApexCharts from 'vue3-apexcharts';
import { Activity, BarChart3, Gauge, LineChart, PieChart, TrendingUp } from 'lucide-vue-next';
import { useThemeStore } from '../../stores/themeStore';
import type { PublicMonitorReport } from '../../types';

const apexchart = VueApexCharts;

interface LivePoint {
  timestamp: string;
  status: 'UP' | 'DOWN';
  latencyMs: number;
}

const props = withDefaults(defineProps<{ report: PublicMonitorReport | null; livePoints?: LivePoint[] }>(), { livePoints: () => [] });
const themeStore = useThemeStore();

function themeColor(variable: string, fallback: string): string {
  if (typeof window === 'undefined') return fallback;
  return getComputedStyle(document.documentElement).getPropertyValue(variable).trim() || fallback;
}
type ViewMode = 'area' | 'stepline' | 'bar' | 'donut' | 'gauge';
const viewMode = ref<ViewMode>('area');
const viewModes = [
  { value: 'area' as const, label: 'Area', icon: TrendingUp },
  { value: 'stepline' as const, label: 'Stepline', icon: LineChart },
  { value: 'bar' as const, label: 'Bar', icon: BarChart3 },
  { value: 'donut' as const, label: 'Donut', icon: PieChart },
  { value: 'gauge' as const, label: 'Gauge', icon: Gauge }
];

const trendPoints = computed(() => {
  const historical = (props.report?.points || []).map(point => ({ timestamp: point.bucket, uptimePercent: point.uptimePercent, avgLatencyMs: point.avgLatencyMs }));
  const live = props.livePoints.map(point => ({ timestamp: point.timestamp, uptimePercent: point.status === 'UP' ? 100 : 0, avgLatencyMs: point.latencyMs }));
  return [...historical, ...live].sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
});

const statusCounts = computed(() => {
  const up = (props.report?.upChecks || 0) + props.livePoints.filter(point => point.status === 'UP').length;
  const down = (props.report?.downChecks || 0) + props.livePoints.filter(point => point.status === 'DOWN').length;
  return { up, down, total: up + down };
});
const uptimeValue = computed(() => statusCounts.value.total ? Number(((statusCounts.value.up / statusCounts.value.total) * 100).toFixed(2)) : 100);
const hasData = computed(() => trendPoints.value.length > 0 || statusCounts.value.total > 0);
const chartType = computed(() => viewMode.value === 'donut' ? 'donut' : viewMode.value === 'gauge' ? 'radialBar' : viewMode.value === 'bar' ? 'bar' : 'line');
const labels = computed(() => trendPoints.value.map(point => new Date(point.timestamp).toLocaleString('en-GB', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit', second: props.livePoints.length ? '2-digit' : undefined })));

const series = computed(() => {
  if (viewMode.value === 'donut') return [statusCounts.value.up, statusCounts.value.down];
  if (viewMode.value === 'gauge') return [uptimeValue.value];
  return [
    { name: 'Uptime %', type: chartType.value, data: trendPoints.value.map(point => point.uptimePercent) },
    { name: 'Latency ms', type: chartType.value, data: trendPoints.value.map(point => point.avgLatencyMs) }
  ];
});

const options = computed(() => {
  themeStore.currentTheme;
  const chartText = themeColor('--chart-label', '#64748B');
  const chartGrid = themeColor('--chart-grid', '#D5DDE7');
  const primaryText = themeColor('--text-primary', '#182231');
  const cardBackground = themeColor('--bg-card', '#FFFFFF');
  if (viewMode.value === 'donut') {
    return {
      chart: { type: 'donut', toolbar: { show: false }, foreColor: chartText, background: 'transparent', fontFamily: 'JetBrains Mono, monospace' },
      labels: ['UP checks', 'DOWN checks'], colors: ['#57d68d', '#ef476f'], legend: { show: false },
      dataLabels: { enabled: true, formatter: (value: number) => `${value.toFixed(1)}%` },
      plotOptions: { pie: { donut: { size: '68%', labels: { show: true, total: { show: true, label: 'Checks', color: chartText, formatter: () => String(statusCounts.value.total) } } } } },
      stroke: { width: 2, colors: [cardBackground] }, tooltip: { theme: themeStore.currentTheme }
    };
  }
  if (viewMode.value === 'gauge') {
    return {
      chart: { type: 'radialBar', toolbar: { show: false }, foreColor: chartText, background: 'transparent', fontFamily: 'JetBrains Mono, monospace' },
      colors: ['#57d68d'], labels: ['Uptime'],
      plotOptions: { radialBar: { startAngle: -135, endAngle: 135, hollow: { size: '62%', background: cardBackground }, track: { background: chartGrid, strokeWidth: '100%' }, dataLabels: { name: { color: chartText, fontSize: '11px', offsetY: 22 }, value: { color: primaryText, fontSize: '28px', fontWeight: 800, offsetY: -14, formatter: (value: number) => `${value.toFixed(2)}%` } } } },
      stroke: { lineCap: 'round' }
    };
  }
  return {
    chart: { type: chartType.value, toolbar: { show: false }, foreColor: chartText, background: 'transparent', fontFamily: 'JetBrains Mono, monospace', animations: { enabled: true, easing: 'easeinout', speed: 260, dynamicAnimation: { enabled: true, speed: 260 } } },
    colors: ['#57d68d', '#7b96f5'], stroke: { curve: viewMode.value === 'stepline' ? 'stepline' : 'smooth', width: [2, 2] }, dataLabels: { enabled: false },
    grid: { borderColor: chartGrid, strokeDashArray: 3 }, xaxis: { categories: labels.value, labels: { style: { fontSize: '10px' }, rotate: -35, hideOverlappingLabels: true } },
    yaxis: [{ min: 0, max: 100, tickAmount: 5, labels: { formatter: (value: number) => `${value.toFixed(0)}%` } }, { opposite: true, min: 0, labels: { formatter: (value: number) => `${Math.round(value)}ms` } }],
    tooltip: { shared: true, intersect: false, y: [{ formatter: (value: number) => `${value.toFixed(2)}%` }, { formatter: (value: number) => `${Math.round(value)} ms` }] }, legend: { show: false }
  };
});
</script>

<style scoped>
.chart-shell { @apply min-w-0 rounded-lg border border-subtle/70 bg-card/40 p-3; }
.chart-toolbar { @apply mb-3 flex flex-wrap items-center justify-between gap-2; }
.chart-mode-switcher { @apply flex flex-wrap items-center gap-1 rounded-lg border border-subtle bg-surface p-1; display: flex; align-items: center; flex-wrap: wrap; gap: 4px; }
.chart-mode-button { @apply inline-flex items-center gap-1.5 rounded-md px-2 py-1.5 text-[10px] font-mono text-text-secondary transition-colors hover:bg-hover hover:text-text-main focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-periwinkle/60; display: inline-flex; align-items: center; gap: 6px; min-width: 62px; white-space: nowrap; cursor: pointer; }
.chart-mode-label { display: inline-block; white-space: nowrap; }
.chart-mode-active { @apply bg-subtle font-bold text-text-main; }
.chart-live-state { @apply inline-flex items-center gap-1.5 text-[10px] font-mono text-text-muted; }
.live-dot { @apply h-1.5 w-1.5 rounded-full bg-text-muted; }
.live-dot-active { @apply animate-pulse bg-status-up; }
.chart-empty { @apply flex h-64 flex-col items-center justify-center rounded-lg border border-subtle/70 bg-card px-6 text-center; }
.chart-legend { @apply flex flex-wrap items-center justify-center gap-6 pt-1 text-[10px] font-mono text-text-secondary; }
.chart-legend span { @apply inline-flex items-center gap-1.5; }
.chart-legend strong { @apply text-text-main; }
.legend-dot { width: 7px; height: 7px; border-radius: 999px; display: inline-block; }
</style>
