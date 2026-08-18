<template>
  <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4 shadow-xl">
    <!-- Header with Metric Tabs, Presentation Views & Range Selector -->
    <div class="flex flex-wrap items-center justify-between gap-3 border-b border-[#26262A] pb-3">
      <div class="flex items-center gap-3">
        <div>
          <h3 class="text-xs font-bold text-gray-200 uppercase font-mono tracking-wider flex items-center gap-2">
            <Activity class="w-4 h-4 text-[#7B96F5]" />
            Telemetry &amp; Performance Analytics
          </h3>
          <p class="text-[11px] text-gray-400 font-mono mt-0.5">Real-time ICMP &amp; SNMP Telemetry Time-Series</p>
        </div>
        <!-- Realtime Live Stream Badge -->
        <div class="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-[10px] font-mono text-emerald-400">
          <span class="w-1.5 h-1.5 rounded-full bg-[#3ECF8E] pulsing-dot-green"></span>
          <span class="font-semibold uppercase tracking-wider">Live Poller</span>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <!-- Metric Type Selector (Latency, CPU, Memory, All) -->
        <div class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5 text-xs font-mono">
          <button
            v-for="mode in metricOptions"
            :key="mode.id"
            @click="switchMetric(mode.id)"
            class="px-2.5 py-1 rounded transition-all uppercase font-semibold text-[11px] flex items-center gap-1 cursor-pointer"
            :class="activeMetric === mode.id ? 'bg-[#7B96F5] text-white shadow-sm' : 'text-gray-400 hover:text-white'"
          >
            <component :is="mode.icon" class="w-3 h-3" />
            <span>{{ mode.label }}</span>
          </button>
        </div>

        <!-- Presentation View Mode Toggle (Area, Step, Bar, Donut, Gauge) — Available for ALL metrics including Combined! -->
        <div class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5 text-xs font-mono">
          <!-- Area Chart -->
          <button
            @click="viewMode = 'area'"
            title="Area Time-Series Chart"
            class="px-2 py-1 rounded transition-colors text-[11px] flex items-center gap-1 cursor-pointer"
            :class="viewMode === 'area' ? 'bg-[#26262A] text-white font-bold' : 'text-gray-400 hover:text-white'"
          >
            <Activity class="w-3.5 h-3.5" />
            <span class="hidden sm:inline">Area</span>
          </button>

          <!-- Stepline / Line Chart -->
          <button
            @click="viewMode = 'stepline'"
            title="Stepline Chart"
            class="px-2 py-1 rounded transition-colors text-[11px] flex items-center gap-1 cursor-pointer"
            :class="viewMode === 'stepline' ? 'bg-[#26262A] text-white font-bold' : 'text-gray-400 hover:text-white'"
          >
            <TrendingUp class="w-3.5 h-3.5" />
            <span class="hidden sm:inline">Step</span>
          </button>

          <!-- Bar / Column Chart -->
          <button
            @click="viewMode = 'bar'"
            title="Bar / Column Chart"
            class="px-2 py-1 rounded transition-colors text-[11px] flex items-center gap-1 cursor-pointer"
            :class="viewMode === 'bar' ? 'bg-[#26262A] text-white font-bold' : 'text-gray-400 hover:text-white'"
          >
            <BarChart3 class="w-3.5 h-3.5" />
            <span class="hidden sm:inline">Bar</span>
          </button>

          <!-- Donut Breakdown -->
          <button
            @click="viewMode = 'donut'"
            title="Proportion Breakdown (Selected Range)"
            class="px-2 py-1 rounded transition-colors text-[11px] flex items-center gap-1 cursor-pointer"
            :class="viewMode === 'donut' ? 'bg-[#26262A] text-white font-bold' : 'text-gray-400 hover:text-white'"
          >
            <PieChart class="w-3.5 h-3.5" />
            <span class="hidden sm:inline">Donut</span>
          </button>

          <!-- RadialBar Gauge (Snapshot) -->
          <button
            @click="viewMode = 'gauge'"
            title="Current Snapshot Gauge"
            class="px-2 py-1 rounded transition-colors text-[11px] flex items-center gap-1 cursor-pointer"
            :class="viewMode === 'gauge' ? 'bg-[#26262A] text-white font-bold' : 'text-gray-400 hover:text-white'"
          >
            <Gauge class="w-3.5 h-3.5" />
            <span class="hidden sm:inline">Gauge</span>
          </button>
        </div>

        <!-- Time Range Selector -->
        <div v-if="viewMode !== 'gauge'" class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5 text-xs font-mono">
          <button
            v-for="range in (['1h', '24h', '7d', '30d', 'custom'] as const)"
            :key="range"
            @click="selectRange(range)"
            class="px-2 py-1 rounded transition-colors text-[11px] cursor-pointer"
            :class="activeRange === range ? 'bg-[#26262A] text-white font-bold' : 'text-gray-400 hover:text-white'"
          >
            {{ range }}
          </button>
        </div>
      </div>
    </div>

    <!-- Custom Date Range Picker -->
    <div v-if="activeRange === 'custom' && viewMode !== 'gauge'" class="flex flex-wrap items-center gap-3 bg-[#18181B] border border-[#26262A] rounded-lg p-2.5 text-xs font-mono">
      <div class="flex items-center gap-2">
        <span class="text-gray-400">From:</span>
        <input type="date" v-model="customFrom" @change="fetchAndRender" class="bg-[#0A0A0B] border border-[#26262A] rounded px-2 py-1 text-white text-xs" />
      </div>
      <div class="flex items-center gap-2">
        <span class="text-gray-400">To:</span>
        <input type="date" v-model="customTo" @change="fetchAndRender" class="bg-[#0A0A0B] border border-[#26262A] rounded px-2 py-1 text-white text-xs" />
      </div>
      <button @click="fetchAndRender" class="px-3 py-1 bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-bold rounded text-xs cursor-pointer">
        Apply Range
      </button>
    </div>

    <!-- Quick Stats Telemetry Strip (Current, Average, Peak, Min) -->
    <div v-if="!isEmpty && activeMetric !== 'all'" class="grid grid-cols-2 sm:grid-cols-4 gap-2 pt-1">
      <div class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2">
        <span class="text-[10px] uppercase font-mono text-gray-400 block font-semibold">Current Live</span>
        <span class="text-sm font-bold font-mono text-white flex items-center gap-1.5 mt-0.5">
          <span class="w-2 h-2 rounded-full inline-block" :style="{ backgroundColor: accentColor }"></span>
          {{ latestValue.toFixed(1) }} {{ metricUnit }}
        </span>
      </div>
      <div class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2">
        <span class="text-[10px] uppercase font-mono text-gray-400 block font-semibold">Average</span>
        <span class="text-sm font-bold font-mono text-gray-200 mt-0.5 block">
          {{ avgMetricVal.toFixed(1) }} {{ metricUnit }}
        </span>
      </div>
      <div class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2">
        <span class="text-[10px] uppercase font-mono text-gray-400 block font-semibold">Peak (Max)</span>
        <span class="text-sm font-bold font-mono text-red-400 mt-0.5 block">
          {{ maxMetricVal.toFixed(1) }} {{ metricUnit }}
        </span>
      </div>
      <div class="bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2">
        <span class="text-[10px] uppercase font-mono text-gray-400 block font-semibold">Minimum</span>
        <span class="text-sm font-bold font-mono text-[#10B981] mt-0.5 block">
          {{ minMetricVal.toFixed(1) }} {{ metricUnit }}
        </span>
      </div>
    </div>

    <!-- Combined Mode Telemetry KPI Strip -->
    <div v-if="!isEmpty && activeMetric === 'all'" class="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-1">
      <div class="bg-[#18181B] border border-[#26262A] rounded-lg p-3 flex items-center justify-between">
        <div>
          <span class="text-[10px] uppercase font-mono text-gray-400 block font-semibold flex items-center gap-1">
            <Zap class="w-3 h-3 text-[#10B981]" />
            ICMP Latency
          </span>
          <span class="text-sm font-bold font-mono text-[#10B981] mt-0.5 block">
            {{ latestLatency.toFixed(1) }} ms <span class="text-[10px] text-gray-500 font-normal">(Avg: {{ avgLatency.toFixed(1) }} ms)</span>
          </span>
        </div>
      </div>
      <div class="bg-[#18181B] border border-[#26262A] rounded-lg p-3 flex items-center justify-between">
        <div>
          <span class="text-[10px] uppercase font-mono text-gray-400 block font-semibold flex items-center gap-1">
            <Cpu class="w-3 h-3 text-[#7B96F5]" />
            SNMP CPU Load
          </span>
          <span class="text-sm font-bold font-mono text-[#7B96F5] mt-0.5 block">
            {{ latestCpu.toFixed(1) }}% <span class="text-[10px] text-gray-500 font-normal">(Avg: {{ avgCpu.toFixed(1) }}%)</span>
          </span>
        </div>
      </div>
      <div class="bg-[#18181B] border border-[#26262A] rounded-lg p-3 flex items-center justify-between">
        <div>
          <span class="text-[10px] uppercase font-mono text-gray-400 block font-semibold flex items-center gap-1">
            <Server class="w-3 h-3 text-[#F59E0B]" />
            SNMP RAM Usage
          </span>
          <span class="text-sm font-bold font-mono text-[#F59E0B] mt-0.5 block">
            {{ latestMem.toFixed(1) }}% <span class="text-[10px] text-gray-500 font-normal">(Avg: {{ avgMem.toFixed(1) }}%)</span>
          </span>
        </div>
      </div>
    </div>

    <!-- ApexCharts Graph Container -->
    <div class="relative w-full" :class="viewMode === 'gauge' || viewMode === 'donut' ? 'h-72' : 'h-64'">
      <apexchart
        v-if="!isEmpty"
        :key="`${viewMode}-${activeMetric}-${activeRange}-${rawMetricsKey}`"
        width="100%"
        :height="viewMode === 'gauge' || viewMode === 'donut' ? 280 : 250"
        :type="apexChartType"
        :options="apexOptions"
        :series="apexSeries"
      />

      <!-- Clean No-Dummy-Data Empty State -->
      <div v-else class="absolute inset-0 flex flex-col items-center justify-center p-6 text-center bg-[#151517] border border-[#26262A]/60 rounded-xl space-y-2">
        <div class="w-10 h-10 rounded-full bg-[#18181B] border border-[#26262A] flex items-center justify-center text-gray-500">
          <component :is="activeMetricIcon" class="w-5 h-5" />
        </div>
        <h4 class="text-xs font-mono font-bold text-gray-300 uppercase tracking-wider">
          No Realtime {{ activeMetricLabel }} Recorded
        </h4>
        <p class="text-[11px] font-mono text-gray-500 max-w-sm leading-relaxed">
          {{ emptyStateDescription }}
        </p>
      </div>
    </div>

    <!-- Donut Range-Aware Legend -->
    <div v-if="viewMode === 'donut' && !isEmpty" class="flex flex-wrap items-center justify-center gap-6 text-xs font-mono text-gray-400 pt-1">
      <template v-if="activeMetric === 'all'">
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-[#10B981] inline-block"></span>
          Avg Ping: {{ avgLatency.toFixed(1) }} ms
        </span>
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-[#7B96F5] inline-block"></span>
          Avg CPU: {{ avgCpu.toFixed(1) }}%
        </span>
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-[#F59E0B] inline-block"></span>
          Avg RAM: {{ avgMem.toFixed(1) }}%
        </span>
      </template>
      <template v-else-if="activeMetric === 'status'">
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-[#10B981] inline-block"></span>
          UP (Reachable) — {{ upPct }}%
        </span>
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-[#EF4444] inline-block"></span>
          DOWN (0 ms) — {{ downPct }}%
        </span>
      </template>
      <template v-else>
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-[#10B981] inline-block"></span>
          Normal &lt;50% ({{ lowResourceCount }})
        </span>
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-[#F59E0B] inline-block"></span>
          Moderate 50-80% ({{ medResourceCount }})
        </span>
        <span class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-[#EF4444] inline-block"></span>
          High &gt;80% ({{ highResourceCount }})
        </span>
      </template>
      <span class="text-[10px] text-gray-600">Range: {{ activeRange !== 'custom' ? activeRange : `${customFrom} → ${customTo}` }}</span>
    </div>

    <!-- Line/Area Legend -->
    <div v-if="activeMetric === 'status' && (viewMode === 'area' || viewMode === 'stepline') && !isEmpty" class="flex items-center justify-center gap-6 text-[10px] font-mono text-gray-400 pt-1 border-t border-[#26262A]/40 mt-1">
      <span class="flex items-center gap-1.5 text-gray-300 font-semibold">
        <span class="w-3.5 h-1 bg-[#10B981] inline-block rounded"></span>
        Solid Green Line = ICMP Ping Latency
      </span>
      <span v-if="downPeriods.length > 0" class="flex items-center gap-1.5 text-red-400 font-semibold">
        <span class="w-3 h-3 bg-[#F16565]/20 border border-[#F16565]/50 inline-block rounded-sm"></span>
        Red Band = Downtime Period
      </span>
    </div>

    <!-- Combined Mode Area/Step Legend -->
    <div v-if="activeMetric === 'all' && (viewMode === 'area' || viewMode === 'stepline' || viewMode === 'bar') && !isEmpty" class="flex items-center justify-center gap-6 text-[10px] font-mono text-gray-400 pt-1 border-t border-[#26262A]/40 mt-1">
      <span class="flex items-center gap-1.5 text-[#00E396] font-semibold">
        <span class="w-3 h-3 rounded-full bg-[#00E396]/20 border border-[#00E396] inline-block"></span>
        Ping Latency (ms)
      </span>
      <span class="flex items-center gap-1.5 text-[#38BDF8] font-semibold">
        <span class="w-3 h-3 rounded-full bg-[#38BDF8]/20 border border-[#38BDF8] inline-block"></span>
        CPU Load (%)
      </span>
      <span class="flex items-center gap-1.5 text-[#FBBF24] font-semibold">
        <span class="w-3 h-3 rounded-full bg-[#FBBF24]/20 border border-[#FBBF24] inline-block"></span>
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
}>();

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

const rawMetricsKey = computed(() => {
  return `${rawLatencyMetrics.value.length}-${rawCpuMetrics.value.length}-${rawMemMetrics.value.length}`;
});

const metricOptions = [
  { id: 'status' as MetricMode, label: 'Latency (ICMP)', icon: Zap },
  { id: 'cpu' as MetricMode, label: 'CPU (SNMP)', icon: Cpu },
  { id: 'memory' as MetricMode, label: 'RAM (SNMP)', icon: Server },
  { id: 'all' as MetricMode, label: 'Combined', icon: Layers }
];

const activeMetricLabel = computed(() => {
  if (activeMetric.value === 'status') return 'Ping Latency';
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

const avgMetricVal = computed(() => {
  const list = currentActiveList.value;
  if (list.length === 0) return 0;
  const sum = list.reduce((acc, m) => acc + m.value, 0);
  return sum / list.length;
});

const maxMetricVal = computed(() => {
  const list = currentActiveList.value;
  if (list.length === 0) return 0;
  return Math.max(...list.map(m => m.value));
});

const minMetricVal = computed(() => {
  const list = currentActiveList.value;
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

  return periods;
});

const apexChartType = computed((): 'area' | 'bar' | 'radialBar' | 'donut' => {
  if (viewMode.value === 'gauge') return 'radialBar';
  if (viewMode.value === 'donut') return 'donut';
  if (viewMode.value === 'bar') return 'bar';
  return 'area';
});

const accentColor = computed(() => {
  if (activeMetric.value === 'cpu') return '#38BDF8';
  if (activeMetric.value === 'memory') return '#FBBF24';
  if (activeMetric.value === 'all') return '#38BDF8';
  return '#00E396';
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
    const val = latestValue.value;
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
        data: rawLatencyMetrics.value.map(i => ({ x: new Date(i.recordedAt).getTime(), y: i.value }))
      },
      {
        name: 'CPU Load (%)',
        data: rawCpuMetrics.value.map(i => ({ x: new Date(i.recordedAt).getTime(), y: i.value }))
      },
      {
        name: 'RAM Usage (%)',
        data: rawMemMetrics.value.map(i => ({ x: new Date(i.recordedAt).getTime(), y: i.value }))
      }
    ];
  }

  const seriesData = currentActiveList.value.map(item => ({
    x: new Date(item.recordedAt).getTime(),
    y: item.value
  }));
  const label = activeMetric.value === 'status' ? 'Latency (ms)' : activeMetric.value === 'cpu' ? 'CPU Load (%)' : 'RAM Usage (%)';
  return [{ name: label, data: seriesData }];
});

const apexOptions = computed((): any => {
  const metricLabel = activeMetric.value.toUpperCase();

  if (viewMode.value === 'gauge') {
    const isMulti = activeMetric.value === 'all';
    const gaugeColors = isMulti ? ['#10B981', '#7B96F5', '#F59E0B'] : [accentColor.value];
    const gaugeLabels = isMulti ? ['Ping (ms)', 'CPU (%)', 'RAM (%)'] : [metricLabel];

    return {
      chart: {
        type: 'radialBar',
        background: 'transparent',
        foreColor: '#9CA3AF',
        sparkline: { enabled: false }
      },
      plotOptions: {
        radialBar: {
          startAngle: isMulti ? -180 : -135,
          endAngle: isMulti ? 180 : 135,
          hollow: {
            margin: 0,
            size: isMulti ? '40%' : '68%',
            background: '#151517'
          },
          track: {
            background: '#26262A',
            strokeWidth: '100%'
          },
          dataLabels: {
            show: true,
            name: {
              offsetY: -8,
              color: '#9CA3AF',
              fontSize: '11px',
              fontWeight: '600',
              fontFamily: 'JetBrains Mono, monospace'
            },
            value: {
              offsetY: 6,
              color: '#FFFFFF',
              fontSize: isMulti ? '16px' : '24px',
              fontWeight: 'bold',
              fontFamily: 'JetBrains Mono, monospace',
              formatter: (val: number) => `${Math.round(val)}%`
            },
            total: {
              show: isMulti,
              label: 'Avg Load',
              color: '#9CA3AF',
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
      ? ['#10B981', '#7B96F5', '#F59E0B']
      : isStatus
      ? ['#10B981', '#EF4444']
      : ['#10B981', '#F59E0B', '#EF4444'];

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
        foreColor: '#9CA3AF'
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
                color: '#9CA3AF',
                offsetY: -8
              },
              value: {
                show: true,
                fontSize: '20px',
                fontFamily: 'JetBrains Mono, monospace',
                fontWeight: 'bold',
                color: '#FFFFFF',
                offsetY: 4,
                formatter: (val: string) => `${Math.round(Number(val))}%`
              },
              total: {
                show: true,
                showAlways: true,
                label: isMulti ? 'Combined' : isStatus ? 'Uptime' : `Avg ${activeMetric.value.toUpperCase()}`,
                fontSize: '10px',
                fontFamily: 'JetBrains Mono, monospace',
                color: '#9CA3AF',
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
      stroke: { width: 2, colors: ['#151517'] },
      tooltip: {
        theme: 'dark'
      }
    };
  }

  // Area / Bar / Stepline Options
  const isMulti = activeMetric.value === 'all';
  const colors = isMulti ? ['#00E396', '#38BDF8', '#FBBF24'] : [accentColor.value];

  return {
    chart: {
      type: apexChartType.value,
      background: 'transparent',
      toolbar: { show: false },
      zoom: { enabled: true },
      foreColor: '#9CA3AF',
      fontFamily: 'JetBrains Mono, monospace'
    },
    // Lightweight DOWN Period shading (Fast & Smooth)
    annotations: {
      xaxis: downPeriods.value.map(p => ({
        x: p.x,
        x2: p.x2,
        fillColor: '#EF4444',
        opacity: 0.2,
        strokeDashArray: 0,
        borderColor: '#EF4444',
        borderWidth: 1,
        label: {
          borderColor: '#EF4444',
          style: {
            color: '#FFFFFF',
            background: '#EF4444',
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
      curve: viewMode.value === 'stepline' ? 'stepline' : 'smooth',
      width: viewMode.value === 'stepline' ? 3.5 : 3,
      connectNulls: false
    },
    markers: {
      size: viewMode.value === 'bar' ? 0 : (viewMode.value === 'stepline' ? (currentActiveList.value.length < 80 ? 4 : 2) : (currentActiveList.value.length < 50 ? 4 : 0)),
      strokeWidth: 2,
      strokeColors: '#151517',
      hover: { size: 6.5 }
    },
    plotOptions: {
      bar: {
        columnWidth: '55%',
        borderRadius: 2
      }
    },
    fill: {
      type: viewMode.value === 'bar' ? 'solid' : 'gradient',
      gradient: {
        shade: 'dark',
        type: 'vertical',
        shadeIntensity: 0.8,
        opacityFrom: viewMode.value === 'stepline' ? 0.65 : 0.55,
        opacityTo: viewMode.value === 'stepline' ? 0.15 : 0.08,
        stops: [0, 90, 100]
      }
    },
    dataLabels: { enabled: false },
    xaxis: {
      type: 'datetime',
      labels: {
        datetimeUTC: false
      },
      axisBorder: { color: '#26262A' },
      axisTicks: { color: '#26262A' }
    },
    yaxis: isMulti
      ? [
          {
            title: { text: 'Ping (ms)', style: { color: '#00E396', fontSize: '10px' } },
            labels: { formatter: (v: number) => `${v.toFixed(0)} ms`, style: { colors: '#00E396' } }
          },
          {
            opposite: true,
            title: { text: 'SNMP %', style: { color: '#38BDF8', fontSize: '10px' } },
            labels: { formatter: (v: number) => `${v.toFixed(0)}%`, style: { colors: '#38BDF8' } },
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
    grid: { borderColor: '#26262A', strokeDashArray: 3 },
    tooltip: {
      theme: 'dark',
      x: { format: 'dd MMM HH:mm:ss' }
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
