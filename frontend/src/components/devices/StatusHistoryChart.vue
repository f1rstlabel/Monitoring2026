<template>
  <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4 shadow-xl">
    <!-- Header with Metric, Presentation Mode & Range Selector -->
    <div class="flex flex-wrap items-center justify-between gap-3 border-b border-[#26262A] pb-3">
      <div class="flex items-center gap-3">
        <div>
          <h3 class="text-xs font-bold text-gray-200 uppercase font-mono tracking-wider">Historical Reachability &amp; Metrics</h3>
          <p class="text-[11px] text-gray-500 mt-0.5">Realtime WebSocket Stream (Powered by ApexCharts)</p>
        </div>
        <!-- Realtime Live Stream Badge -->
        <div class="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-[10px] font-mono text-emerald-400">
          <span class="w-1.5 h-1.5 rounded-full bg-[#3ECF8E] pulsing-dot-green"></span>
          <span class="font-semibold uppercase tracking-wider">Realtime Live</span>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <!-- Metric Type Selector (Status, CPU, Memory) -->
        <div class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5 text-xs font-mono">
          <button
            v-for="mode in (['status', 'cpu', 'memory'] as const)"
            :key="mode"
            @click="switchMetric(mode)"
            class="px-2.5 py-1 rounded transition-colors uppercase font-semibold text-[11px]"
            :class="activeMetric === mode ? 'bg-[#7B96F5] text-white' : 'text-gray-400 hover:text-white'"
          >
            {{ mode }}
          </button>
        </div>

        <!-- Presentation View Mode Toggle -->
        <div class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5 text-xs font-mono">
          <!-- Line/Area — always available -->
          <button
            @click="viewMode = 'line'"
            title="Line/Area Time Series Chart"
            class="px-2.5 py-1 rounded transition-colors text-[11px] flex items-center gap-1"
            :class="viewMode === 'line' ? 'bg-[#26262A] text-white font-bold' : 'text-gray-400 hover:text-white'"
          >
            <Activity class="w-3.5 h-3.5" />
            <span>Line/Area</span>
          </button>

          <!-- Donut — only for Status metric -->
          <button
            v-if="activeMetric === 'status'"
            @click="viewMode = 'donut'"
            title="Up vs Down Proportion (selected range)"
            class="px-2.5 py-1 rounded transition-colors text-[11px] flex items-center gap-1"
            :class="viewMode === 'donut' ? 'bg-[#26262A] text-white font-bold' : 'text-gray-400 hover:text-white'"
          >
            <PieChart class="w-3.5 h-3.5" />
            <span>Up/Down</span>
          </button>

          <!-- RadialBar Gauge — only for CPU / Memory -->
          <button
            v-if="activeMetric !== 'status'"
            @click="viewMode = 'gauge'"
            title="Current Snapshot — Radial Gauge"
            class="px-2.5 py-1 rounded transition-colors text-[11px] flex items-center gap-1"
            :class="viewMode === 'gauge' ? 'bg-[#26262A] text-white font-bold' : 'text-gray-400 hover:text-white'"
          >
            <Gauge class="w-3.5 h-3.5" />
            <span>Gauge</span>
          </button>
        </div>

        <!-- Snapshot badge — shown when gauge is active -->
        <div
          v-if="viewMode === 'gauge'"
          class="flex items-center gap-1 px-2 py-0.5 rounded-full bg-amber-500/10 border border-amber-500/30 text-[10px] font-mono text-amber-400"
          title="Gauge shows the current live value — not a historical range"
        >
          <span class="text-[9px] uppercase font-bold tracking-wider">Snapshot · Current Value</span>
        </div>

        <!-- Time Range Selector — hidden when gauge (snapshot) -->
        <div v-if="viewMode !== 'gauge'" class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5 text-xs font-mono">
          <button
            v-for="range in (['24h', '7d', '30d', 'custom'] as const)"
            :key="range"
            @click="selectRange(range)"
            class="px-2 py-1 rounded transition-colors text-[11px]"
            :class="activeRange === range ? 'bg-[#26262A] text-white font-bold' : 'text-gray-400 hover:text-white'"
          >
            {{ range }}
          </button>
        </div>
      </div>
    </div>

    <!-- Custom Date Range Picker -->
    <div v-if="activeRange === 'custom' && viewMode !== 'gauge'" class="flex items-center gap-3 bg-[#18181B] border border-[#26262A] rounded-lg p-2.5 text-xs font-mono">
      <div class="flex items-center gap-2">
        <span class="text-gray-400">From:</span>
        <input type="date" v-model="customFrom" @change="fetchAndRender" class="bg-[#0A0A0B] border border-[#26262A] rounded px-2 py-1 text-white text-xs" />
      </div>
      <div class="flex items-center gap-2">
        <span class="text-gray-400">To:</span>
        <input type="date" v-model="customTo" @change="fetchAndRender" class="bg-[#0A0A0B] border border-[#26262A] rounded px-2 py-1 text-white text-xs" />
      </div>
      <button @click="fetchAndRender" class="px-3 py-1 bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-bold rounded text-xs">
        Apply Range
      </button>
    </div>

    <!-- ApexCharts Container — full width, no centering flex wrapper -->
    <div class="relative w-full" :class="viewMode === 'gauge' || viewMode === 'donut' ? 'h-72' : 'h-64'">
      <apexchart
        v-if="!isEmpty"
        :key="`${viewMode}-${activeMetric}-${activeRange}`"
        width="100%"
        :height="viewMode === 'gauge' || viewMode === 'donut' ? 280 : 250"
        :type="apexChartType"
        :options="apexOptions"
        :series="apexSeries"
      />
      <div v-else class="absolute inset-0 flex items-center justify-center text-xs font-mono text-gray-500 bg-[#151517]/80">
        No probe metric data available for selected range
      </div>
    </div>

    <!-- Donut range-aware legend -->
    <div v-if="viewMode === 'donut' && !isEmpty" class="flex items-center justify-center gap-6 text-xs font-mono text-gray-400 pt-1">
      <span class="flex items-center gap-1.5">
        <span class="w-2.5 h-2.5 rounded-full bg-[#3ECF8E] inline-block"></span>
        UP (reachable) — {{ upPct }}%
      </span>
      <span class="flex items-center gap-1.5">
        <span class="w-2.5 h-2.5 rounded-full bg-[#F16565] inline-block"></span>
        DOWN (0 ms) — {{ downPct }}%
      </span>
      <span class="text-[10px] text-gray-600">Range: {{ activeRange !== 'custom' ? activeRange : `${customFrom} → ${customTo}` }}</span>
    </div>

    <!-- Line range-aware legend -->
    <div v-if="activeMetric === 'status' && viewMode === 'line' && !isEmpty" class="flex items-center justify-center gap-6 text-[10px] font-mono text-gray-500 pt-1 border-t border-[#26262A]/40 mt-1">
      <span class="flex items-center gap-1.5">
        <span class="w-3.5 h-0.5 bg-[#3ECF8E] inline-block"></span>
        Green Line = Latency while UP
      </span>
      <span class="flex items-center gap-1.5">
        <span class="w-3 h-3 bg-[#F16565]/20 border border-[#F16565]/35 inline-block rounded-sm"></span>
        Red Band = Downtime Period
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import VueApexCharts from 'vue3-apexcharts';
import api from '../../api/client';
import { wsClient } from '../../ws/websocket';
import { Activity, PieChart, Gauge } from 'lucide-vue-next';

const apexchart = VueApexCharts;

const props = defineProps<{
  deviceId?: string;
}>();

type MetricMode = 'status' | 'cpu' | 'memory';
type ViewMode = 'line' | 'gauge' | 'donut';
type RangeMode = '24h' | '7d' | '30d' | 'custom';

const activeMetric = ref<MetricMode>('status');
const viewMode = ref<ViewMode>('line');
const activeRange = ref<RangeMode>('24h');

const customFrom = ref('');
const customTo = ref('');
const isEmpty = ref(false);

const rawMetrics = ref<{ value: number; recordedAt: string }[]>([]);

/** Switch metric and auto-correct viewMode to something valid for the new metric */
function switchMetric(mode: MetricMode) {
  activeMetric.value = mode;
  // donut is only for status; gauge is only for cpu/memory
  if (mode === 'status' && viewMode.value === 'gauge') viewMode.value = 'line';
  if (mode !== 'status' && viewMode.value === 'donut') viewMode.value = 'line';
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

async function fetchAndRender() {
  if (!props.deviceId) return;

  const mType = activeMetric.value === 'status' ? 'latency' : activeMetric.value;

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

    rawMetrics.value = dataItems.map((m: any) => ({
      value: Number(m.value) || 0,
      recordedAt: m.recordedAt
    })).sort((a, b) => new Date(a.recordedAt).getTime() - new Date(b.recordedAt).getTime());

    isEmpty.value = rawMetrics.value.length === 0;
  } catch (e) {
    rawMetrics.value = [];
    isEmpty.value = true;
  }
}

const latestValue = computed(() => {
  if (rawMetrics.value.length === 0) return 0;
  return rawMetrics.value[rawMetrics.value.length - 1].value;
});

// Donut computation: count reachable (latency > 0) vs unreachable (latency === 0) data points
const upCount = computed(() => rawMetrics.value.filter(m => m.value > 0).length);
const downCount = computed(() => rawMetrics.value.filter(m => m.value === 0).length);
const upPct = computed(() => {
  const total = upCount.value + downCount.value;
  return total === 0 ? 0 : Math.round((upCount.value / total) * 100);
});
const downPct = computed(() => 100 - upPct.value);

const downPeriods = computed(() => {
  if (activeMetric.value !== 'status') return [];

  const periods: { x: number; x2: number }[] = [];
  let inDownPeriod = false;
  let startX = 0;

  for (let i = 0; i < rawMetrics.value.length; i++) {
    const item = rawMetrics.value[i];
    const t = new Date(item.recordedAt).getTime();

    if (item.value === 0) {
      if (!inDownPeriod) {
        inDownPeriod = true;
        startX = t;
      }
    } else {
      if (inDownPeriod) {
        inDownPeriod = false;
        periods.push({ x: startX, x2: t });
      }
    }
  }

  if (inDownPeriod && rawMetrics.value.length > 0) {
    const lastT = new Date(rawMetrics.value[rawMetrics.value.length - 1].recordedAt).getTime();
    periods.push({ x: startX, x2: lastT });
  }

  return periods;
});

const apexChartType = computed((): 'area' | 'radialBar' | 'donut' => {
  if (viewMode.value === 'gauge') return 'radialBar';
  if (viewMode.value === 'donut') return 'donut';
  return 'area';
});

const accentColor = computed(() => {
  if (activeMetric.value === 'cpu') return '#7B96F5';
  if (activeMetric.value === 'memory') return '#F59E0B';
  return '#3ECF8E';
});

const apexSeries = computed((): any => {
  if (viewMode.value === 'gauge') {
    const val = latestValue.value;
    return [Math.min(100, Math.max(0, Math.round(val)))];
  }

  if (viewMode.value === 'donut') {
    // [UP count, DOWN count] for donut slices — range-aware (rawMetrics is fetched per range)
    const total = upCount.value + downCount.value;
    if (total === 0) return [1, 0]; // avoid empty chart
    return [upCount.value, downCount.value];
  }

  // line/area
  const seriesData = rawMetrics.value.map(item => ({
    x: new Date(item.recordedAt).getTime(),
    y: (activeMetric.value === 'status' && item.value === 0) ? null : item.value
  }));
  const label = activeMetric.value === 'status' ? 'Latency (ms)' : activeMetric.value === 'cpu' ? 'CPU Utilization (%)' : 'Memory Usage (%)';
  return [{ name: label, data: seriesData }];
});

const apexOptions = computed((): any => {
  const metricLabel = activeMetric.value === 'status' ? 'LATENCY' : activeMetric.value.toUpperCase();

  if (viewMode.value === 'gauge') {
    return {
      chart: {
        type: 'radialBar',
        background: 'transparent',
        foreColor: '#9CA3AF',
        sparkline: { enabled: false }
      },
      plotOptions: {
        radialBar: {
          startAngle: -135,
          endAngle: 135,
          hollow: {
            margin: 0,
            size: '68%',
            background: '#151517'
          },
          track: {
            background: '#26262A',
            strokeWidth: '100%'
          },
          dataLabels: {
            show: true,
            name: {
              offsetY: -10,
              color: '#9CA3AF',
              fontSize: '12px',
              fontWeight: '600',
              fontFamily: 'JetBrains Mono, monospace'
            },
            value: {
              offsetY: 8,
              color: '#FFFFFF',
              fontSize: '24px',
              fontWeight: 'bold',
              fontFamily: 'JetBrains Mono, monospace',
              formatter: () => {
                const val = latestValue.value;
                return `${val.toFixed(0)}%`;
              }
            }
          }
        }
      },
      colors: [accentColor.value],
      labels: [metricLabel]
    };
  }

  if (viewMode.value === 'donut') {
    const total = upCount.value + downCount.value;
    return {
      chart: {
        type: 'donut',
        background: 'transparent',
        foreColor: '#9CA3AF'
      },
      colors: ['#3ECF8E', '#F16565'],
      labels: [
        `UP (${upCount.value} checks)`,
        `DOWN (${downCount.value} checks)`
      ],
      legend: {
        show: false // we use our custom legend below the chart
      },
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
                fontSize: '22px',
                fontFamily: 'JetBrains Mono, monospace',
                fontWeight: 'bold',
                color: '#FFFFFF',
                offsetY: 4,
                formatter: (val: string) => `${Math.round(Number(val))}%`
              },
              total: {
                show: true,
                showAlways: true,
                label: 'Uptime',
                fontSize: '11px',
                fontFamily: 'JetBrains Mono, monospace',
                color: '#9CA3AF',
                formatter: () => `${upPct.value}%`
              }
            }
          }
        }
      },
      stroke: { width: 2, colors: ['#151517'] },
      tooltip: {
        theme: 'dark',
        y: {
          formatter: (val: number) => `${val} checks (${total > 0 ? Math.round((val / total) * 100) : 0}%)`
        }
      }
    };
  }

  // line/area
  return {
    chart: {
      type: 'area',
      background: 'transparent',
      toolbar: { show: false },
      zoom: { enabled: true },
      foreColor: '#9CA3AF',
      fontFamily: 'JetBrains Mono, monospace'
    },
    annotations: {
      xaxis: downPeriods.value.map(p => ({
        x: p.x,
        x2: p.x2,
        fillColor: '#F16565',
        opacity: 0.2,
        strokeDashArray: 0,
        borderColor: '#F16565',
        borderWidth: 1,
        label: {
          style: {
            color: '#F16565',
            background: '#151517',
            fontSize: '9px',
            fontFamily: 'JetBrains Mono, monospace',
            fontWeight: 'bold'
          },
          text: 'DOWN'
        }
      }))
    },
    colors: [accentColor.value],
    stroke: { curve: 'smooth', width: 2, connectNulls: false },
    fill: {
      type: 'gradient',
      gradient: {
        shadeIntensity: 1,
        opacityFrom: 0.45,
        opacityTo: 0.05,
        stops: [0, 90, 100]
      }
    },
    dataLabels: { enabled: false },
    xaxis: {
      type: 'datetime',
      axisBorder: { color: '#26262A' },
      axisTicks: { color: '#26262A' }
    },
    yaxis: {
      labels: {
        formatter: (val: number) => activeMetric.value === 'status' ? `${val.toFixed(0)} ms` : `${val.toFixed(0)}%`
      }
    },
    grid: { borderColor: '#26262A' },
    tooltip: {
      theme: 'dark',
      x: { format: 'dd MMM HH:mm' }
    }
  };
});

function handleRealtimeWSMessage(data: any) {
  if (!props.deviceId) return;
  const msgDevId = data.deviceId || data.DeviceID;
  if (msgDevId && msgDevId !== props.deviceId) return;

  const latencyMs = Number(data.latencyMs) || 0;
  rawMetrics.value.push({
    value: latencyMs,
    recordedAt: new Date(data.timestamp || Date.now()).toISOString()
  });

  if (rawMetrics.value.length > 300) {
    rawMetrics.value.shift();
  }
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
