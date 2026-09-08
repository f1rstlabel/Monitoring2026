<template>
  <div class="device-chart-shell">
    <!-- Header with Metric Tabs, Presentation Views & Range Selector -->
    <div class="device-chart-header">
      <div class="flex items-center gap-3">
        <div>
          <h3 class="text-xs font-bold text-text-main uppercase font-mono tracking-wider flex items-center gap-2">
            <Activity class="w-4 h-4 text-brand-periwinkle" />
            Availability and response
          </h3>
          <p class="text-[11px] text-text-secondary font-mono mt-0.5">Uptime and latency for the selected device</p>
        </div>
        <!-- Realtime Live Stream Badge -->
        <div class="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-[10px] font-mono text-emerald-400">
          <span class="w-1.5 h-1.5 rounded-full bg-status-up pulsing-dot-green"></span>
          <span class="font-semibold uppercase tracking-wider">Live Poller</span>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <!-- Metric Type Selector (Latency, CPU, Memory, All) -->
        <div class="flex items-center bg-card border border-subtle rounded-lg p-0.5 text-xs font-mono">
          <button
            v-for="mode in metricOptions"
            :key="mode.id"
            @click="switchMetric(mode.id)"
            class="px-2.5 py-1 rounded transition-all uppercase font-semibold text-[11px] flex items-center gap-1 cursor-pointer"
            :class="activeMetric === mode.id ? 'bg-brand-periwinkle text-white shadow-sm' : 'text-text-secondary hover:text-text-main'"
          >
            <component :is="mode.icon" class="w-3 h-3" />
            <span>{{ mode.label }}</span>
          </button>
        </div>

        <!-- Presentation View Mode Toggle (Area, Step, Bar, Donut, Gauge) -->
        <div class="chart-mode-switcher" role="group" aria-label="Chart view mode">
          <!-- Area Chart -->
          <button
            @click="viewMode = 'area'"
            title="Area Time-Series Chart"
            class="chart-mode-button"
            :class="viewMode === 'area' ? 'chart-mode-active' : ''"
          >
            <Activity class="w-3.5 h-3.5" />
            <span class="chart-mode-label">Area</span>
          </button>

          <!-- Stepline / Line Chart -->
          <button
            @click="viewMode = 'stepline'"
            title="Stepline Chart"
            class="chart-mode-button"
            :class="viewMode === 'stepline' ? 'chart-mode-active' : ''"
          >
            <TrendingUp class="w-3.5 h-3.5" />
            <span class="chart-mode-label">Stepline</span>
          </button>

          <!-- Bar / Column Chart -->
          <button
            @click="viewMode = 'bar'"
            title="Bar / Column Chart"
            class="chart-mode-button"
            :class="viewMode === 'bar' ? 'chart-mode-active' : ''"
          >
            <BarChart3 class="w-3.5 h-3.5" />
            <span class="chart-mode-label">Bar</span>
          </button>

          <!-- Donut Breakdown -->
          <button
            @click="viewMode = 'donut'"
            title="Proportion Breakdown (Selected Range)"
            class="chart-mode-button"
            :class="viewMode === 'donut' ? 'chart-mode-active' : ''"
          >
            <PieChart class="w-3.5 h-3.5" />
            <span class="chart-mode-label">Donut</span>
          </button>

          <!-- RadialBar Gauge (Snapshot) -->
          <button
            @click="viewMode = 'gauge'"
            title="Current Snapshot Gauge"
            class="chart-mode-button"
            :class="viewMode === 'gauge' ? 'chart-mode-active' : ''"
          >
            <Gauge class="w-3.5 h-3.5" />
            <span class="chart-mode-label">Gauge</span>
          </button>
        </div>

        <!-- Time Range Selector -->
        <div v-if="viewMode !== 'gauge'" class="flex items-center bg-card border border-subtle rounded-lg p-0.5 text-xs font-mono">
          <button
            v-for="range in (['1h', '24h', '7d', '30d', 'custom'] as const)"
            :key="range"
            @click="selectRange(range)"
            class="px-2 py-1 rounded transition-colors text-[11px] cursor-pointer"
            :class="activeRange === range ? 'bg-subtle text-text-main font-bold' : 'text-text-secondary hover:text-text-main'"
          >
            {{ range }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="isDeviceDown" class="device-down-banner" role="status">
      <div class="device-down-banner-main">
        <div class="device-down-icon"><Zap class="w-4 h-4" /></div>
        <div>
          <span class="device-down-eyebrow">Current state</span>
          <strong>Device is down</strong>
          <p>No response was recorded on the latest poll. Latency is unavailable until the device responds again.</p>
        </div>
      </div>
      <div class="device-down-meta">
        <span>Last sample</span>
        <strong>{{ latestRecordedAt }}</strong>
      </div>
    </div>

    <!-- Custom Date Range Picker -->
    <div v-if="activeRange === 'custom' && viewMode !== 'gauge'" class="flex flex-wrap items-center gap-3 bg-card border border-subtle rounded-lg p-2.5 text-xs font-mono">
      <div class="flex items-center gap-2">
        <span class="text-text-secondary">From:</span>
        <input type="date" v-model="customFrom" @change="fetchAndRender" class="bg-main border border-subtle rounded px-2 py-1 text-text-main text-xs" />
      </div>
      <div class="flex items-center gap-2">
        <span class="text-text-secondary">To:</span>
        <input type="date" v-model="customTo" @change="fetchAndRender" class="bg-main border border-subtle rounded px-2 py-1 text-text-main text-xs" />
      </div>
      <button @click="fetchAndRender" class="px-3 py-1 bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-bold rounded text-xs cursor-pointer">
        Apply Range
      </button>
    </div>

    <!-- Availability KPI strip -->
    <div v-if="!isEmpty && activeMetric === 'status'" class="metric-strip">
      <div class="metric-card">
        <span>Current latency</span>
        <strong class="text-brand-periwinkle">
          <template v-if="latestStatusLatency !== null">{{ latestStatusLatency.toFixed(1) }}<small>ms</small></template>
          <template v-else>Unavailable</template>
        </strong>
      </div>
      <div class="metric-card">
        <span>Uptime</span>
        <strong class="text-status-up">{{ uptimeValue.toFixed(2) }}<small>%</small></strong>
      </div>
      <div class="metric-card">
        <span>Average latency</span>
        <strong>{{ avgMetricVal.toFixed(1) }}<small>ms</small></strong>
      </div>
      <div class="metric-card">
        <span>Down checks</span>
        <strong :class="downCount > 0 ? 'text-status-down' : 'text-status-up'">{{ downCount }}</strong>
      </div>
    </div>

    <!-- Resource KPI strip -->
    <div v-if="!isEmpty && activeMetric !== 'all' && activeMetric !== 'status'" class="metric-strip">
      <div class="bg-card border border-subtle rounded-lg px-3 py-2">
        <span class="text-[10px] uppercase font-mono text-text-secondary block font-semibold">Current Live</span>
        <span class="text-sm font-bold font-mono text-text-main flex items-center gap-1.5 mt-0.5">
          <span class="w-2 h-2 rounded-full inline-block" :style="{ backgroundColor: accentColor }"></span>
          {{ latestValue.toFixed(1) }} {{ metricUnit }}
        </span>
      </div>
      <div class="bg-card border border-subtle rounded-lg px-3 py-2">
        <span class="text-[10px] uppercase font-mono text-text-secondary block font-semibold">Average</span>
        <span class="text-sm font-bold font-mono text-text-main mt-0.5 block">
          {{ avgMetricVal.toFixed(1) }} {{ metricUnit }}
        </span>
      </div>
      <div class="bg-card border border-subtle rounded-lg px-3 py-2">
        <span class="text-[10px] uppercase font-mono text-text-secondary block font-semibold">Peak (Max)</span>
        <span class="text-sm font-bold font-mono text-red-400 mt-0.5 block">
          {{ maxMetricVal.toFixed(1) }} {{ metricUnit }}
        </span>
      </div>
      <div class="bg-card border border-subtle rounded-lg px-3 py-2">
        <span class="text-[10px] uppercase font-mono text-text-secondary block font-semibold">Minimum</span>
        <span class="text-sm font-bold font-mono text-status-up mt-0.5 block">
          {{ minMetricVal.toFixed(1) }} {{ metricUnit }}
        </span>
      </div>
    </div>

    <!-- Combined Mode Telemetry KPI Strip -->
    <div v-if="!isEmpty && activeMetric === 'all'" class="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-1">
      <div class="bg-card border border-subtle rounded-lg p-3 flex items-center justify-between">
        <div>
          <span class="text-[10px] uppercase font-mono text-text-secondary block font-semibold flex items-center gap-1">
            <Zap class="w-3 h-3 text-status-up" />
            ICMP Latency
          </span>
          <span class="text-sm font-bold font-mono text-status-up mt-0.5 block">
            {{ latestLatency.toFixed(1) }} ms <span class="text-[10px] text-text-muted font-normal">(Avg: {{ avgLatency.toFixed(1) }} ms)</span>
          </span>
        </div>
      </div>
      <div class="bg-card border border-subtle rounded-lg p-3 flex items-center justify-between">
        <div>
          <span class="text-[10px] uppercase font-mono text-text-secondary block font-semibold flex items-center gap-1">
            <Cpu class="w-3 h-3 text-brand-periwinkle" />
            SNMP CPU Load
          </span>
          <span class="text-sm font-bold font-mono text-brand-periwinkle mt-0.5 block">
            {{ latestCpu.toFixed(1) }}% <span class="text-[10px] text-text-muted font-normal">(Avg: {{ avgCpu.toFixed(1) }}%)</span>
          </span>
        </div>
      </div>
      <div class="bg-card border border-subtle rounded-lg p-3 flex items-center justify-between">
        <div>
          <span class="text-[10px] uppercase font-mono text-text-secondary block font-semibold flex items-center gap-1">
            <Server class="w-3 h-3 text-status-warning" />
            SNMP RAM Usage
          </span>
          <span class="text-sm font-bold font-mono text-status-warning mt-0.5 block">
            {{ latestMem.toFixed(1) }}% <span class="text-[10px] text-text-muted font-normal">(Avg: {{ avgMem.toFixed(1) }}%)</span>
          </span>
        </div>
      </div>
    </div>

    <!-- ApexCharts Graph Container -->
    <div class="device-chart-canvas" :class="viewMode === 'gauge' || viewMode === 'donut' ? 'device-chart-canvas-tall' : ''">
      <apexchart
        v-if="!isEmpty"
        :key="`${viewMode}-${activeMetric}-${activeRange}`"
        width="100%"
        :height="viewMode === 'gauge' || viewMode === 'donut' ? 280 : 250"
        :type="apexChartType"
        :options="apexOptions"
        :series="apexSeries"
      />

      <!-- Clean No-Dummy-Data Empty State -->
      <div v-else class="absolute inset-0 flex flex-col items-center justify-center p-6 text-center bg-surface border border-subtle/60 rounded-xl space-y-2">
        <div class="w-10 h-10 rounded-full bg-card border border-subtle flex items-center justify-center text-text-muted">
          <component :is="activeMetricIcon" class="w-5 h-5" />
        </div>
        <h4 class="text-xs font-mono font-bold text-text-secondary uppercase tracking-wider">
          No Realtime {{ activeMetricLabel }} Recorded
        </h4>
        <p class="text-[11px] font-mono text-text-muted max-w-sm leading-relaxed">
          {{ emptyStateDescription }}
        </p>
      </div>
    </div>

    <!-- Donut Range-Aware Legend -->
    <div v-if="viewMode === 'donut' && !isEmpty" class="flex flex-wrap items-center justify-center gap-6 text-xs font-mono text-text-secondary pt-1">
      <template v-if="activeMetric === 'all'">
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-status-up inline-block"></span>
          Avg Ping: {{ avgLatency.toFixed(1) }} ms
        </span>
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-brand-periwinkle inline-block"></span>
          Avg CPU: {{ avgCpu.toFixed(1) }}%
        </span>
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-status-warning inline-block"></span>
          Avg RAM: {{ avgMem.toFixed(1) }}%
        </span>
      </template>
      <template v-else-if="activeMetric === 'status'">
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-status-up inline-block"></span>
          UP (Reachable) — {{ upPct }}%
        </span>
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-status-down inline-block"></span>
          DOWN (0 ms) — {{ downPct }}%
        </span>
      </template>
      <template v-else>
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-status-up inline-block"></span>
          Normal &lt;50% ({{ lowResourceCount }})
        </span>
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-status-warning inline-block"></span>
          Moderate 50-80% ({{ medResourceCount }})
        </span>
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-status-down inline-block"></span>
          High &gt;80% ({{ highResourceCount }})
        </span>
      </template>
      <span class="text-[10px] text-text-muted">Range: {{ activeRange !== 'custom' ? activeRange : `${customFrom} → ${customTo}` }}</span>
    </div>

    <!-- Line/Area Legend -->
    <div v-if="activeMetric === 'status' && (viewMode === 'area' || viewMode === 'stepline') && !isEmpty" class="flex items-center justify-center gap-6 text-[10px] font-mono text-text-secondary pt-1 border-t border-subtle/40 mt-1">
      <span class="flex items-center gap-1.5 text-status-up font-semibold">
        <span class="w-3.5 h-1 bg-status-up inline-block rounded"></span>
        Uptime (%)
      </span>
      <span class="flex items-center gap-1.5 text-brand-periwinkle font-semibold">
        <span class="w-3.5 h-1 bg-brand-periwinkle inline-block rounded"></span>
        Latency (ms)
      </span>
      <span v-if="downPeriods.length > 0" class="flex items-center gap-1.5 text-red-400 font-semibold">
        <span class="w-3 h-3 bg-status-down/20 border border-status-down/50 inline-block rounded-sm"></span>
        Red Band = Downtime Period
      </span>
    </div>

    <!-- Combined Mode Area/Step Legend -->
    <div v-if="activeMetric === 'all' && (viewMode === 'area' || viewMode === 'stepline' || viewMode === 'bar') && !isEmpty" class="flex items-center justify-center gap-6 text-[10px] font-mono text-text-secondary pt-1 border-t border-subtle/40 mt-1">
      <span class="flex items-center gap-1.5 text-status-up font-semibold">
        <span class="w-3 h-3 rounded-full bg-status-up/20 border border-status-up inline-block"></span>
        Ping Latency (ms)
      </span>
      <span class="flex items-center gap-1.5 text-brand-periwinkle font-semibold">
        <span class="w-3 h-3 rounded-full bg-brand-periwinkle/20 border border-brand-periwinkle inline-block"></span>
        CPU Load (%)
      </span>
      <span class="flex items-center gap-1.5 text-status-warning font-semibold">
        <span class="w-3 h-3 rounded-full bg-status-warning/20 border border-status-warning inline-block"></span>
        RAM Usage (%)
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import VueApexCharts from 'vue3-apexcharts';
import api from '../../api/client';
import { wsClient } from '../../ws/websocket';
import { useThemeStore } from '../../stores/themeStore';
import {
  Activity,
  BarChart3,
  PieChart,
  Gauge,
  TrendingUp,
  Zap,
  Cpu,
  Server,
  Layers
} from 'lucide-vue-next';

const apexchart = VueApexCharts;

const props = defineProps<{
  deviceId?: string;
  deviceStatus?: string;
}>();

const themeStore = useThemeStore();

function themeColor(variable: string, fallback: string): string {
  if (typeof window === 'undefined') return fallback;
  return getComputedStyle(document.documentElement).getPropertyValue(variable).trim() || fallback;
}

type MetricMode = 'status' | 'cpu' | 'memory' | 'all';
type ViewMode = 'area' | 'bar' | 'stepline' | 'gauge' | 'donut';
type RangeMode = '1h' | '24h' | '7d' | '30d' | 'custom';

const activeMetric = ref<MetricMode>('status');
const viewMode = ref<ViewMode>('area');
const activeRange = ref<RangeMode>('24h');

const customFrom = ref('');
const customTo = ref('');
const isEmpty = ref(false);

const rawLatencyMetrics = ref<{ value: number; recordedAt: string }[]>([]);
const rawCpuMetrics = ref<{ value: number; recordedAt: string }[]>([]);
const rawMemMetrics = ref<{ value: number; recordedAt: string }[]>([]);

const latestRecordedAt = computed(() => {
  const latest = rawLatencyMetrics.value[rawLatencyMetrics.value.length - 1]?.recordedAt;
  return latest ? new Date(latest).toLocaleString('en-GB', { dateStyle: 'short', timeStyle: 'short' }) : 'Unavailable';
});
const isDeviceDown = computed(() => {
  if (props.deviceStatus === 'DOWN') return true;
  const latest = rawLatencyMetrics.value[rawLatencyMetrics.value.length - 1];
  return Boolean(latest && latest.value === 0);
});

const MAX_RENDER_POINTS = 180;

interface AvailabilityBucket {
  recordedAt: string;
  uptimePercent: number;
  latencyMs: number | null;
  totalChecks: number;
  upChecks: number;
  downChecks: number;
}

function compactMetrics(items: { value: number; recordedAt: string }[], maxPoints = MAX_RENDER_POINTS, preserveZero = false) {
  if (items.length <= maxPoints) return items;

  const bucketSize = Math.ceil(items.length / maxPoints);
  const compacted: { value: number; recordedAt: string }[] = [];

  for (let start = 0; start < items.length; start += bucketSize) {
    const bucket = items.slice(start, Math.min(start + bucketSize, items.length));
    const downPoint = preserveZero ? bucket.find(item => item.value === 0) : undefined;

    if (downPoint) {
      compacted.push(downPoint);
      continue;
    }

    const average = bucket.reduce((sum, item) => sum + item.value, 0) / bucket.length;
    compacted.push({
      value: average,
      recordedAt: bucket[Math.floor(bucket.length / 2)].recordedAt
    });
  }

  return compacted;
}

const renderLatencyMetrics = computed(() => compactMetrics(rawLatencyMetrics.value, MAX_RENDER_POINTS, true));
const renderCpuMetrics = computed(() => compactMetrics(rawCpuMetrics.value));
const renderMemMetrics = computed(() => compactMetrics(rawMemMetrics.value));
const chartLatencyMetrics = computed(() => viewMode.value === 'bar' ? compactMetrics(rawLatencyMetrics.value, 48, true) : renderLatencyMetrics.value);
const chartCpuMetrics = computed(() => viewMode.value === 'bar' ? compactMetrics(rawCpuMetrics.value, 48) : renderCpuMetrics.value);
const chartMemMetrics = computed(() => viewMode.value === 'bar' ? compactMetrics(rawMemMetrics.value, 48) : renderMemMetrics.value);

function availabilityBucketSize() {
  if (activeRange.value === '1h') return 5 * 60 * 1000;
  if (activeRange.value === '24h') return 60 * 60 * 1000;
  if (activeRange.value === '7d') return 6 * 60 * 60 * 1000;
  if (activeRange.value === '30d') return 24 * 60 * 60 * 1000;

  const from = customFrom.value ? new Date(`${customFrom.value}T00:00:00`).getTime() : Date.now() - 7 * 24 * 60 * 60 * 1000;
  const to = customTo.value ? new Date(`${customTo.value}T23:59:59`).getTime() : Date.now();
  const rangeMs = Math.max(1, to - from);
  return rangeMs <= 24 * 60 * 60 * 1000 ? 5 * 60 * 1000 : rangeMs <= 7 * 24 * 60 * 60 * 1000 ? 6 * 60 * 60 * 1000 : 24 * 60 * 60 * 1000;
}

const availabilityBuckets = computed<AvailabilityBucket[]>(() => {
  const bucketSize = availabilityBucketSize();
  const buckets = new Map<number, { totalChecks: number; upChecks: number; downChecks: number; latencyTotal: number; latencyChecks: number }>();

  for (const item of rawLatencyMetrics.value) {
    const timestamp = new Date(item.recordedAt).getTime();
    if (!Number.isFinite(timestamp)) continue;
    const bucketStart = Math.floor(timestamp / bucketSize) * bucketSize;
    const current = buckets.get(bucketStart) || { totalChecks: 0, upChecks: 0, downChecks: 0, latencyTotal: 0, latencyChecks: 0 };
    current.totalChecks += 1;
    if (item.value > 0) {
      current.upChecks += 1;
      current.latencyTotal += item.value;
      current.latencyChecks += 1;
    } else {
      current.downChecks += 1;
    }
    buckets.set(bucketStart, current);
  }

  return [...buckets.entries()].sort(([a], [b]) => a - b).map(([bucketStart, bucket]) => ({
    recordedAt: new Date(bucketStart).toISOString(),
    uptimePercent: bucket.totalChecks ? Number(((bucket.upChecks / bucket.totalChecks) * 100).toFixed(2)) : 0,
    latencyMs: bucket.latencyChecks ? Number((bucket.latencyTotal / bucket.latencyChecks).toFixed(2)) : null,
    totalChecks: bucket.totalChecks,
    upChecks: bucket.upChecks,
    downChecks: bucket.downChecks
  }))
});

const metricOptions = [
  { id: 'status' as MetricMode, label: 'Uptime + latency', icon: Zap },
  { id: 'cpu' as MetricMode, label: 'CPU (SNMP)', icon: Cpu },
  { id: 'memory' as MetricMode, label: 'RAM (SNMP)', icon: Server },
  { id: 'all' as MetricMode, label: 'Combined', icon: Layers }
];

const activeMetricLabel = computed(() => {
  if (activeMetric.value === 'status') return 'Uptime and latency';
  if (activeMetric.value === 'cpu') return 'SNMP CPU Load';
  if (activeMetric.value === 'memory') return 'SNMP Memory Usage';
  return 'Telemetry Metrics';
});

const activeMetricIcon = computed(() => {
  if (activeMetric.value === 'status') return Zap;
  if (activeMetric.value === 'cpu') return Cpu;
  if (activeMetric.value === 'memory') return Server;
  return Layers;
});

const metricUnit = computed(() => {
  if (activeMetric.value === 'status') return 'ms';
  return '%';
});

const emptyStateDescription = computed(() => {
  if (activeMetric.value === 'status') {
    return 'No ICMP ping probes recorded yet. Probes will register automatically on the next polling cycle.';
  }
  return `No real-time SNMP ${activeMetric.value.toUpperCase()} data recorded yet. Ensure SNMP v2c is active on port 161 with community 'public'.`;
});

function switchMetric(mode: MetricMode) {
  activeMetric.value = mode;
  fetchAndRender();
}

function selectRange(range: RangeMode) {
  activeRange.value = range;
  if (range === 'custom') {
    const today = new Date();
    const sevenDaysAgo = new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000);
    customFrom.value = sevenDaysAgo.toISOString().split('T')[0];
    customTo.value = today.toISOString().split('T')[0];
  }
  fetchAndRender();
}

async function fetchMetricsForType(mType: string): Promise<{ value: number; recordedAt: string }[]> {
  if (!props.deviceId) return [];
  try {
    const res = await api.get(`/devices/${props.deviceId}/metrics`, {
      params: {
        type: mType,
        range: activeRange.value,
        from: customFrom.value,
        to: customTo.value
      }
    });

    let dataItems: any[] = [];
    if (Array.isArray(res.data)) {
      dataItems = res.data;
    } else if (res.data && Array.isArray(res.data.items)) {
      dataItems = res.data.items;
    } else if (res.data && Array.isArray(res.data.data)) {
      dataItems = res.data.data;
    }

    return dataItems
      .map((m: any) => ({
        value: Number(m.value) || 0,
        recordedAt: m.recordedAt
      }))
      .sort((a, b) => new Date(a.recordedAt).getTime() - new Date(b.recordedAt).getTime());
  } catch (e) {
    return [];
  }
}

async function fetchAndRender() {
  if (!props.deviceId) return;

  if (activeMetric.value === 'all') {
    const [lat, cpu, mem] = await Promise.all([
      fetchMetricsForType('latency'),
      fetchMetricsForType('cpu'),
      fetchMetricsForType('memory')
    ]);
    rawLatencyMetrics.value = lat;
    rawCpuMetrics.value = cpu;
    rawMemMetrics.value = mem;
    isEmpty.value = lat.length === 0 && cpu.length === 0 && mem.length === 0;
  } else if (activeMetric.value === 'status') {
    const lat = await fetchMetricsForType('latency');
    rawLatencyMetrics.value = lat;
    isEmpty.value = lat.length === 0;
  } else if (activeMetric.value === 'cpu') {
    const cpu = await fetchMetricsForType('cpu');
    rawCpuMetrics.value = cpu;
    isEmpty.value = cpu.length === 0;
  } else if (activeMetric.value === 'memory') {
    const mem = await fetchMetricsForType('memory');
    rawMemMetrics.value = mem;
    isEmpty.value = mem.length === 0;
  }
}

const currentActiveList = computed(() => {
  if (activeMetric.value === 'status') return rawLatencyMetrics.value;
  if (activeMetric.value === 'cpu') return rawCpuMetrics.value;
  if (activeMetric.value === 'memory') return rawMemMetrics.value;
  return rawLatencyMetrics.value;
});

const latestValue = computed(() => {
  const list = currentActiveList.value;
  if (list.length === 0) return 0;
  return list[list.length - 1].value;
});

const latestStatusLatency = computed(() => {
  if (rawLatencyMetrics.value.length === 0) return null;
  const latest = rawLatencyMetrics.value[rawLatencyMetrics.value.length - 1];
  return latest.value > 0 ? latest.value : null;
});

const statisticalList = computed(() => {
  if (activeMetric.value === 'status') {
    return currentActiveList.value.filter(item => item.value > 0);
  }
  return currentActiveList.value;
});

const uptimeValue = computed(() => {
  const total = upCount.value + downCount.value;
  return total === 0 ? 0 : (upCount.value / total) * 100;
});

const avgMetricVal = computed(() => {
  const list = statisticalList.value;
  if (list.length === 0) return 0;
  const sum = list.reduce((acc, m) => acc + m.value, 0);
  return sum / list.length;
});

const maxMetricVal = computed(() => {
  const list = statisticalList.value;
  if (list.length === 0) return 0;
  return Math.max(...list.map(m => m.value));
});

const minMetricVal = computed(() => {
  const list = statisticalList.value;
  if (list.length === 0) return 0;
  return Math.min(...list.map(m => m.value));
});

// Multi-metric helpers for Combined mode
const latestLatency = computed(() => rawLatencyMetrics.value.length ? rawLatencyMetrics.value[rawLatencyMetrics.value.length - 1].value : 0);
const latestCpu = computed(() => rawCpuMetrics.value.length ? rawCpuMetrics.value[rawCpuMetrics.value.length - 1].value : 0);
const latestMem = computed(() => rawMemMetrics.value.length ? rawMemMetrics.value[rawMemMetrics.value.length - 1].value : 0);

const avgLatency = computed(() => {
  if (!rawLatencyMetrics.value.length) return 0;
  return rawLatencyMetrics.value.reduce((acc, m) => acc + m.value, 0) / rawLatencyMetrics.value.length;
});
const avgCpu = computed(() => {
  if (!rawCpuMetrics.value.length) return 0;
  return rawCpuMetrics.value.reduce((acc, m) => acc + m.value, 0) / rawCpuMetrics.value.length;
});
const avgMem = computed(() => {
  if (!rawMemMetrics.value.length) return 0;
  return rawMemMetrics.value.reduce((acc, m) => acc + m.value, 0) / rawMemMetrics.value.length;
});

// Donut calculations
const upCount = computed(() => rawLatencyMetrics.value.filter(m => m.value > 0).length);
const downCount = computed(() => rawLatencyMetrics.value.filter(m => m.value === 0).length);
const upPct = computed(() => {
  const total = upCount.value + downCount.value;
  return total === 0 ? 0 : Math.round((upCount.value / total) * 100);
});
const downPct = computed(() => 100 - upPct.value);

const lowResourceCount = computed(() => currentActiveList.value.filter(m => m.value < 50).length);
const medResourceCount = computed(() => currentActiveList.value.filter(m => m.value >= 50 && m.value <= 80).length);
const highResourceCount = computed(() => currentActiveList.value.filter(m => m.value > 80).length);

// Lightweight Contiguous DOWN period blocks
const downPeriods = computed(() => {
  if (activeMetric.value !== 'status') return [];

  const periods: { x: number; x2: number }[] = [];
  let inDown = false;
  let startX = 0;

  for (let i = 0; i < rawLatencyMetrics.value.length; i++) {
    const item = rawLatencyMetrics.value[i];
    const t = new Date(item.recordedAt).getTime();

    if (item.value === 0) {
      if (!inDown) {
        inDown = true;
        startX = t;
      }
    } else {
      if (inDown) {
        inDown = false;
        if (t > startX) {
          periods.push({ x: startX, x2: t });
        }
      }
    }
  }

  if (inDown && rawLatencyMetrics.value.length > 0) {
    const lastT = new Date(rawLatencyMetrics.value[rawLatencyMetrics.value.length - 1].recordedAt).getTime();
    if (lastT > startX) {
      periods.push({ x: startX, x2: lastT });
    }
  }

  return periods.length > 80 ? periods.slice(-80) : periods;
});

const apexChartType = computed((): 'area' | 'bar' | 'radialBar' | 'donut' => {
  if (viewMode.value === 'gauge') return 'radialBar';
  if (viewMode.value === 'donut') return 'donut';
  if (viewMode.value === 'bar') return 'bar';
  return 'area';
});

const accentColor = computed(() => {
  themeStore.currentTheme;
  if (activeMetric.value === 'cpu' || activeMetric.value === 'all') return themeColor('--accent-blue', '#657BE8');
  if (activeMetric.value === 'memory') return themeColor('--status-warning', '#B46A13');
  return themeColor('--status-up', '#12805C');
});

const apexSeries = computed((): any => {
  if (viewMode.value === 'gauge') {
    if (activeMetric.value === 'all') {
      return [
        Math.min(100, Math.max(0, Math.round(latestLatency.value))),
        Math.min(100, Math.max(0, Math.round(latestCpu.value))),
        Math.min(100, Math.max(0, Math.round(latestMem.value)))
      ];
    }
    const val = activeMetric.value === 'status' ? uptimeValue.value : latestValue.value;
    return [Math.min(100, Math.max(0, Math.round(val)))];
  }

  if (viewMode.value === 'donut') {
    if (activeMetric.value === 'all') {
      return [
        Math.max(1, Math.round(avgLatency.value)),
        Math.max(1, Math.round(avgCpu.value)),
        Math.max(1, Math.round(avgMem.value))
      ];
    }
    if (activeMetric.value === 'status') {
      const total = upCount.value + downCount.value;
      if (total === 0) return [1, 0];
      return [upCount.value, downCount.value];
    } else {
      const total = currentActiveList.value.length;
      if (total === 0) return [1, 0, 0];
      return [lowResourceCount.value, medResourceCount.value, highResourceCount.value];
    }
  }

  if (activeMetric.value === 'all') {
    return [
      {
        name: 'Ping Latency (ms)',
        data: chartLatencyMetrics.value.map(i => ({ x: new Date(i.recordedAt).getTime(), y: i.value }))
      },
      {
        name: 'CPU Load (%)',
        data: chartCpuMetrics.value.map(i => ({ x: new Date(i.recordedAt).getTime(), y: i.value }))
      },
      {
        name: 'RAM Usage (%)',
        data: chartMemMetrics.value.map(i => ({ x: new Date(i.recordedAt).getTime(), y: i.value }))
      }
    ];
  }

  if (activeMetric.value === 'status') {
    return [
      {
        name: 'Uptime (%)',
        data: availabilityBuckets.value.map(bucket => ({
          x: new Date(bucket.recordedAt).getTime(),
          y: bucket.uptimePercent
        }))
      },
      {
        name: 'Latency (ms)',
        data: availabilityBuckets.value.map(bucket => ({
          x: new Date(bucket.recordedAt).getTime(),
          // A fully failed bucket has no response time.
          y: bucket.latencyMs
        }))
      }
    ];
  }

  const seriesData = compactMetrics(currentActiveList.value).map(item => ({
    x: new Date(item.recordedAt).getTime(),
    y: item.value
  }));
  const label = activeMetric.value === 'cpu' ? 'CPU Load (%)' : 'RAM Usage (%)';
  return [{ name: label, data: seriesData }];
});

const apexOptions = computed((): any => {
  // Keep ApexCharts in sync with the CSS theme tokens when the user toggles mode.
  themeStore.currentTheme;
  const chartText = themeColor('--chart-label', '#64748B');
  const chartGrid = themeColor('--chart-grid', '#D5DDE7');
  const cardBackground = themeColor('--bg-card', '#FFFFFF');
  const statusUp = themeColor('--status-up', '#12805C');
  const statusDown = themeColor('--status-down', '#C73F4A');
  const warning = themeColor('--status-warning', '#B46A13');
  const accent = themeColor('--accent-blue', '#657BE8');
  const metricLabel = activeMetric.value.toUpperCase();

  if (viewMode.value === 'gauge') {
    const isMulti = activeMetric.value === 'all';
    const gaugeColors = isMulti ? [statusUp, accent, warning] : [accentColor.value];
    const gaugeLabels = isMulti ? ['Ping (ms)', 'CPU (%)', 'RAM (%)'] : [activeMetric.value === 'status' ? 'Uptime' : metricLabel];

    return {
      chart: {
        type: 'radialBar',
        background: 'transparent',
        foreColor: chartText,
        sparkline: { enabled: false }
      },
      plotOptions: {
        radialBar: {
          startAngle: isMulti ? -180 : -135,
          endAngle: isMulti ? 180 : 135,
          hollow: {
            margin: 0,
            size: isMulti ? '40%' : '68%',
            background: cardBackground
          },
          track: {
            background: chartGrid,
            strokeWidth: '100%'
          },
          dataLabels: {
            show: true,
            name: {
              offsetY: -8,
              color: chartText,
              fontSize: '11px',
              fontWeight: '600',
              fontFamily: 'JetBrains Mono, monospace'
            },
            value: {
              offsetY: 6,
              color: themeColor('--text-primary', '#182231'),
              fontSize: isMulti ? '16px' : '24px',
              fontWeight: 'bold',
              fontFamily: 'JetBrains Mono, monospace',
              formatter: (val: number) => `${Math.round(val)}%`
            },
            total: {
              show: isMulti,
              label: 'Avg Load',
              color: chartText,
              fontFamily: 'JetBrains Mono, monospace',
              fontSize: '10px',
              formatter: () => `${Math.round((avgCpu.value + avgMem.value) / 2)}%`
            }
          }
        }
      },
      colors: gaugeColors,
      labels: gaugeLabels
    };
  }

  if (viewMode.value === 'donut') {
    const isMulti = activeMetric.value === 'all';
    const isStatus = activeMetric.value === 'status';

    const colors = isMulti
      ? [statusUp, accent, warning]
      : isStatus
      ? [statusUp, statusDown]
      : [statusUp, warning, statusDown];

    const labels = isMulti
      ? [
          `Ping Latency (${avgLatency.value.toFixed(1)} ms)`,
          `CPU Load (${avgCpu.value.toFixed(1)}%)`,
          `RAM Usage (${avgMem.value.toFixed(1)}%)`
        ]
      : isStatus
      ? [
          `UP (${upCount.value} checks)`,
          `DOWN (${downCount.value} checks)`
        ]
      : [
          `Normal <50% (${lowResourceCount.value})`,
          `Moderate 50-80% (${medResourceCount.value})`,
          `High >80% (${highResourceCount.value})`
        ];

    return {
      chart: {
        type: 'donut',
        background: 'transparent',
        foreColor: chartText
      },
      colors,
      labels,
      legend: { show: false },
      dataLabels: {
        enabled: true,
        formatter: (val: number) => `${Math.round(val)}%`,
        style: {
          fontSize: '13px',
          fontFamily: 'JetBrains Mono, monospace',
          fontWeight: '700'
        },
        dropShadow: { enabled: false }
      },
      plotOptions: {
        pie: {
          donut: {
            size: '60%',
            labels: {
              show: true,
              name: {
                show: true,
                fontSize: '11px',
                fontFamily: 'JetBrains Mono, monospace',
                color: chartText,
                offsetY: -8
              },
              value: {
                show: true,
                fontSize: '20px',
                fontFamily: 'JetBrains Mono, monospace',
                fontWeight: 'bold',
                color: themeColor('--text-primary', '#182231'),
                offsetY: 4,
                formatter: (val: string) => `${Math.round(Number(val))}%`
              },
              total: {
                show: true,
                showAlways: true,
                label: isMulti ? 'Combined' : isStatus ? 'Uptime' : `Avg ${activeMetric.value.toUpperCase()}`,
                fontSize: '10px',
                fontFamily: 'JetBrains Mono, monospace',
                color: chartText,
                formatter: () =>
                  isMulti
                    ? `${avgCpu.value.toFixed(0)}%`
                    : isStatus
                    ? `${upPct.value}%`
                    : `${avgMetricVal.value.toFixed(0)}%`
              }
            }
          }
        }
      },
      stroke: { width: 2, colors: [cardBackground] },
      tooltip: {
        theme: themeStore.currentTheme
      }
    };
  }

  // Area / Bar / Stepline Options
  const isMulti = activeMetric.value === 'all';
  const isStatus = activeMetric.value === 'status';
  const colors = isMulti ? [statusUp, accent, warning] : isStatus ? [statusUp, accent] : [accentColor.value];

  return {
    chart: {
      type: apexChartType.value,
      background: 'transparent',
      toolbar: { show: false },
      zoom: { enabled: true },
      foreColor: chartText,
      fontFamily: 'JetBrains Mono, monospace'
    },
    // Lightweight DOWN Period shading (Fast & Smooth)
    annotations: {
      xaxis: downPeriods.value.map(p => ({
        x: p.x,
        x2: p.x2,
        fillColor: statusDown,
        opacity: 0.2,
        strokeDashArray: 0,
        borderColor: statusDown,
        borderWidth: 1,
        label: {
          borderColor: statusDown,
          style: {
            color: '#FFFFFF',
            background: statusDown,
            fontSize: '9px',
            fontFamily: 'JetBrains Mono, monospace',
            fontWeight: 'bold'
          },
          text: 'DOWN'
        }
      }))
    },
    colors,
    stroke: {
      // Area remains smooth for trend reading; Stepline preserves each poll transition.
      curve: viewMode.value === 'stepline' ? 'stepline' : 'smooth',
      width: isStatus ? [2.5, 2.5] : (viewMode.value === 'stepline' ? 3.5 : 3),
      connectNulls: false
    },
    markers: {
      // Match Public Monitoring: keep the graph clean and reveal a point only on hover.
      size: 0,
      strokeWidth: 0,
      strokeColors: cardBackground,
      hover: { size: 5, sizeOffset: 2 }
    },
    plotOptions: {
      bar: {
        columnWidth: '55%',
        borderRadius: 2
      }
    },
    // Public Monitoring uses a clean line treatment without the heavy fill
    // gradient. Keep only a solid fill for bars.
    fill: { type: 'solid', opacity: viewMode.value === 'bar' ? 0.82 : 0 },
    dataLabels: { enabled: false },
    xaxis: {
      type: 'datetime',
      labels: {
        datetimeUTC: false
      },
      axisBorder: { color: chartGrid },
      axisTicks: { color: chartGrid }
    },
    yaxis: isStatus
      ? [
          {
            seriesName: 'Uptime (%)',
            title: { text: 'Uptime', style: { color: statusUp, fontSize: '10px' } },
            labels: { formatter: (v: number) => `${v.toFixed(0)}%`, style: { colors: statusUp } },
            min: 0,
            max: 100
          },
          {
            seriesName: 'Latency (ms)',
            opposite: true,
            title: { text: 'Latency', style: { color: accent, fontSize: '10px' } },
            labels: { formatter: (v: number) => `${v.toFixed(0)} ms`, style: { colors: accent } },
            min: 0
          }
        ]
      : isMulti
      ? [
          {
            title: { text: 'Ping (ms)', style: { color: statusUp, fontSize: '10px' } },
            labels: { formatter: (v: number) => `${v.toFixed(0)} ms`, style: { colors: statusUp } }
          },
          {
            opposite: true,
            title: { text: 'SNMP %', style: { color: accent, fontSize: '10px' } },
            labels: { formatter: (v: number) => `${v.toFixed(0)}%`, style: { colors: accent } },
            min: 0,
            max: 100
          }
        ]
      : {
          labels: {
            formatter: (val: number) =>
              activeMetric.value === 'status' ? `${val.toFixed(0)} ms` : `${val.toFixed(0)}%`
          },
          min: activeMetric.value === 'status' ? undefined : 0,
          max: activeMetric.value === 'status' ? undefined : 100
        },
    grid: { borderColor: chartGrid, strokeDashArray: 3 },
    tooltip: {
      theme: themeStore.currentTheme,
      x: { format: 'dd MMM HH:mm:ss' },
      y: isStatus
        ? [
            { formatter: (value: number | null) => value === null ? 'Unavailable (DOWN)' : `${value.toFixed(0)}%` },
            { formatter: (value: number | null) => value === null ? 'Unavailable (DOWN)' : `${value.toFixed(1)} ms` }
          ]
        : undefined
    }
  };
});

function handleRealtimeWSMessage(data: any) {
  if (!props.deviceId) return;
  const msgDevId = data.deviceId || data.DeviceID;
  if (msgDevId && msgDevId !== props.deviceId) return;

  const nowIso = new Date(data.timestamp || Date.now()).toISOString();

  if (data.latencyMs !== undefined) {
    rawLatencyMetrics.value.push({
      value: Number(data.latencyMs) || 0,
      recordedAt: nowIso
    });
    if (rawLatencyMetrics.value.length > 300) rawLatencyMetrics.value.shift();
  }

  if (data.cpu !== undefined && data.cpu !== null) {
    rawCpuMetrics.value.push({
      value: Number(data.cpu) || 0,
      recordedAt: nowIso
    });
    if (rawCpuMetrics.value.length > 300) rawCpuMetrics.value.shift();
  }

  if (data.memory !== undefined && data.memory !== null) {
    rawMemMetrics.value.push({
      value: Number(data.memory) || 0,
      recordedAt: nowIso
    });
    if (rawMemMetrics.value.length > 300) rawMemMetrics.value.shift();
  }

  isEmpty.value = currentActiveList.value.length === 0;
}

let unsubscribeWS: (() => void) | null = null;

watch([activeRange], () => {
  fetchAndRender();
});

onMounted(() => {
  fetchAndRender();
  wsClient.connect();
  unsubscribeWS = wsClient.subscribe((data: any) => {
    if (data.type === 'LIVE_FEED' || data.type === 'STATUS_CHANGE') {
      handleRealtimeWSMessage(data);
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

<style scoped>
.device-chart-shell {
  @apply min-w-0 rounded-xl border border-subtle bg-surface p-5;
}

.device-chart-header {
  @apply flex flex-wrap items-center justify-between gap-3 border-b border-subtle pb-3;
}

.chart-mode-switcher {
  @apply flex flex-wrap items-center gap-1 rounded-lg border border-subtle bg-card p-1;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
}

.chart-mode-button {
  @apply inline-flex items-center gap-1.5 rounded-md px-2 py-1.5 text-[10px] font-mono text-text-secondary transition-colors hover:bg-hover hover:text-text-main focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-periwinkle/60;
  min-width: 62px;
  white-space: nowrap;
  cursor: pointer;
}

.chart-mode-label {
  display: inline-block;
  white-space: nowrap;
}

.chart-mode-active {
  @apply bg-subtle font-bold text-text-main;
}

.metric-strip {
  @apply grid grid-cols-2 gap-2 pt-1 sm:grid-cols-4;
}

.metric-card {
  @apply rounded-lg border border-subtle bg-card px-3 py-2;
}

.metric-card span {
  @apply block text-[10px] font-mono font-semibold uppercase text-text-secondary;
}

.metric-card strong {
  @apply mt-1 block text-sm font-mono font-bold text-text-main;
  font-variant-numeric: tabular-nums;
}

.metric-card small {
  @apply ml-0.5 text-[10px] font-normal text-text-muted;
}

.device-chart-canvas {
  @apply relative min-h-[250px] w-full;
}

.device-chart-canvas-tall {
  @apply min-h-[280px];
}
.device-down-banner { @apply flex flex-wrap items-center justify-between gap-4 rounded-xl border border-status-down/35 bg-status-down/10 px-4 py-3; }
.device-down-banner-main { @apply flex min-w-0 items-center gap-3; }
.device-down-icon { @apply flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-status-down/40 bg-status-down/15 text-status-down; }
.device-down-eyebrow { @apply block text-[9px] font-mono uppercase tracking-wider text-status-down/80; }
.device-down-banner strong { @apply mt-0.5 block text-sm font-bold text-text-main; }
.device-down-banner p { @apply mt-1 max-w-xl text-[10px] leading-4 text-text-secondary; }
.device-down-meta { @apply shrink-0 border-l border-status-down/25 pl-4 text-right; }
.device-down-meta span { @apply block text-[9px] font-mono uppercase tracking-wider text-text-muted; }
.device-down-meta strong { @apply mt-1 block text-[10px] font-mono text-text-main; }
</style>
