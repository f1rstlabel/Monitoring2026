<template>
  <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-4">
    <!-- Header with Range Selector & Realtime Indicator -->
    <div class="flex flex-wrap items-center justify-between gap-3 border-b border-[#26262A] pb-3">
      <div class="flex items-center gap-3">
        <div>
          <h3 class="text-xs font-bold text-gray-200 uppercase font-mono tracking-wider">Historical Reachability & Metrics</h3>
          <p class="text-[11px] text-gray-500 mt-0.5">Realtime WebSocket Canvas Stream (Powered by uPlot)</p>
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
            v-for="mode in ['status', 'cpu', 'memory'] as const"
            :key="mode"
            @click="activeMetric = mode"
            class="px-2.5 py-1 rounded transition-colors uppercase font-semibold text-[11px]"
            :class="activeMetric === mode ? 'bg-[#7B96F5] text-white' : 'text-gray-400 hover:text-white'"
          >
            {{ mode }}
          </button>
        </div>

        <!-- Time Range Selector -->
        <div class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5 text-xs font-mono">
          <button
            v-for="range in ['24h', '7d', '30d', 'custom'] as const"
            :key="range"
            @click="activeRange = range"
            class="px-2 py-1 rounded transition-colors text-[11px]"
            :class="activeRange === range ? 'bg-[#26262A] text-white font-bold' : 'text-gray-400 hover:text-white'"
          >
            {{ range }}
          </button>
        </div>

        <!-- Interactive Zoom Controls -->
        <div class="flex items-center bg-[#18181B] border border-[#26262A] rounded-lg p-0.5 text-xs font-mono gap-1">
          <button
            @click="zoomIn"
            title="Zoom In (+)"
            class="px-2 py-1 rounded bg-[#26262A] text-gray-300 hover:text-white hover:bg-[#3A3A40] transition-colors font-bold text-xs"
          >
            +
          </button>
          <button
            @click="zoomOut"
            title="Zoom Out (-)"
            class="px-2 py-1 rounded bg-[#26262A] text-gray-300 hover:text-white hover:bg-[#3A3A40] transition-colors font-bold text-xs"
          >
            -
          </button>
          <button
            @click="resetZoom"
            title="Reset Zoom"
            class="px-2 py-1 rounded bg-[#26262A] text-[#7B96F5] hover:text-white hover:bg-[#3A3A40] transition-colors font-semibold text-[11px]"
          >
            Reset
          </button>
        </div>
      </div>
    </div>

    <!-- Custom Date Range Picker -->
    <div v-if="activeRange === 'custom'" class="flex items-center gap-3 bg-[#18181B] border border-[#26262A] rounded-lg p-2.5 text-xs font-mono">
      <div class="flex items-center gap-2">
        <span class="text-gray-400">From:</span>
        <input type="date" v-model="customFrom" class="bg-[#0A0A0B] border border-[#26262A] rounded px-2 py-1 text-white text-xs" />
      </div>
      <div class="flex items-center gap-2">
        <span class="text-gray-400">To:</span>
        <input type="date" v-model="customTo" class="bg-[#0A0A0B] border border-[#26262A] rounded px-2 py-1 text-white text-xs" />
      </div>
    </div>

    <!-- uPlot Container -->
    <div class="relative w-full h-64 flex items-center justify-center">
      <div ref="chartRef" class="w-full h-full cursor-crosshair"></div>
      <div v-if="isEmpty" class="absolute inset-0 flex items-center justify-center text-xs font-mono text-gray-500 bg-[#151517]/80">
        No probe metric data available for selected range
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue';
import uPlot from 'uplot';
import 'uplot/dist/uPlot.min.css';
import api from '../../api/client';
import { wsClient } from '../../ws/websocket';

const props = defineProps<{
  deviceId?: string;
  history?: { date: string; upCount: number; downCount: number }[];
}>();

type MetricMode = 'status' | 'cpu' | 'memory';
type RangeMode = '24h' | '7d' | '30d' | 'custom';

const activeMetric = ref<MetricMode>('status');
const activeRange = ref<RangeMode>('24h');
const customFrom = ref('');
const customTo = ref('');

const chartRef = ref<HTMLDivElement | null>(null);
const isEmpty = ref(false);
let uplotInstance: uPlot | null = null;
let currentUPlotData: uPlot.AlignedData = [[], [], []];
let initialXMin: number | null = null;
let initialXMax: number | null = null;
let unsubscribeWS: (() => void) | null = null;

async function fetchAndRender() {
  if (!props.deviceId) return;

  const mType = activeMetric.value === 'status' ? 'latency' : activeMetric.value;
  let rawData: { value: number; recordedAt: string }[] = [];

  try {
    const res = await api.get(`/devices/${props.deviceId}/metrics`, {
      params: {
        type: mType,
        range: activeRange.value,
        from: customFrom.value,
        to: customTo.value
      }
    });
    if (Array.isArray(res.data)) {
      rawData = res.data.map((m: any) => ({
        value: Number(m.value) || 0,
        recordedAt: m.recordedAt
      }));
    }
  } catch (e) {
    rawData = [];
  }

  // Sort ascending by time (required for uPlot)
  rawData.sort((a, b) => new Date(a.recordedAt).getTime() - new Date(b.recordedAt).getTime());

  if (rawData.length === 0) {
    isEmpty.value = true;
    const nowSec = Math.floor(Date.now() / 1000);
    const times = [nowSec - 86400, nowSec - 43200, nowSec];
    const dummyVals = activeMetric.value === 'status' ? [null, null, null] : [0, 0, 0];
    currentUPlotData = activeMetric.value === 'status'
      ? [times, dummyVals, dummyVals]
      : [times, dummyVals];
    renderUPlot(currentUPlotData);
    return;
  }

  isEmpty.value = false;

  const xTimestamps: number[] = [];
  const series1: (number | null)[] = [];
  const series2: (number | null)[] = [];

  const isStatus = activeMetric.value === 'status';

  // Calculate max latency for red loss area height
  let maxLat = 30;
  for (const item of rawData) {
    if (item.value > maxLat) maxLat = item.value;
  }

  for (const item of rawData) {
    const tSec = Math.floor(new Date(item.recordedAt).getTime() / 1000);
    xTimestamps.push(tSec);

    if (isStatus) {
      if (item.value > 0) {
        series1.push(item.value);
        series2.push(null);
      } else {
        series1.push(null);
        series2.push(maxLat); // Prominent red loss marker height during DOWN period
      }
    } else {
      series1.push(item.value);
    }
  }

  currentUPlotData = isStatus
    ? [xTimestamps, series1, series2]
    : [xTimestamps, series1];

  await nextTick();
  renderUPlot(currentUPlotData);
}

function handleRealtimeWSMessage(data: any) {
  if (!props.deviceId) return;

  const msgDevId = data.deviceId || data.DeviceID;
  if (msgDevId && msgDevId !== props.deviceId) return;

  const isStatus = activeMetric.value === 'status';
  const tSec = Math.floor(new Date(data.timestamp || Date.now()).getTime() / 1000);
  const latencyMs = Number(data.latencyMs) || 0;
  const isDown = data.status === 'DOWN' || (isStatus && latencyMs === 0);

  if (!currentUPlotData[0]) return;

  const times = currentUPlotData[0] as number[];
  const s1 = currentUPlotData[1] as (number | null)[];
  const s2 = isStatus ? (currentUPlotData[2] as (number | null)[]) : null;

  // Append new timestamp & metric
  times.push(tSec);
  if (isStatus) {
    if (!isDown && latencyMs > 0) {
      s1.push(latencyMs);
      s2?.push(null);
    } else {
      s1.push(null);
      s2?.push(40); // Red DOWN loss marker
    }
  } else {
    s1.push(latencyMs);
  }

  // Keep max 300 points for memory safety
  if (times.length > 300) {
    times.shift();
    s1.shift();
    if (s2) s2.shift();
  }

  if (uplotInstance) {
    // Ultra-fast lightweight Canvas update (<1 ms, zero lag)
    uplotInstance.setData(currentUPlotData);
  }
}

function wheelZoomPlugin(): uPlot.Plugin {
  return {
    hooks: {
      ready: (u: uPlot) => {
        const over = u.over;
        over.addEventListener('wheel', (e: WheelEvent) => {
          e.preventDefault();
          const rect = over.getBoundingClientRect();
          const left = e.clientX - rect.left;
          const leftPct = left / rect.width;
          const xMin = u.scales.x.min!;
          const xMax = u.scales.x.max!;
          const range = xMax - xMin;
          const factor = e.deltaY < 0 ? 0.75 : 1.25;
          const newRange = range * factor;
          const newMin = xMin + (range - newRange) * leftPct;
          const newMax = newMin + newRange;
          u.setScale('x', { min: newMin, max: newMax });
        });
      }
    }
  };
}

function renderUPlot(data: uPlot.AlignedData) {
  if (!chartRef.value) return;

  if (uplotInstance) {
    uplotInstance.destroy();
    uplotInstance = null;
  }

  const containerWidth = chartRef.value.getBoundingClientRect().width || 600;
  const isStatus = activeMetric.value === 'status';

  if (data[0] && data[0].length > 0) {
    initialXMin = data[0][0];
    initialXMax = data[0][data[0].length - 1];
  }

  const opts: uPlot.Options = {
    width: containerWidth,
    height: 240,
    title: '',
    tzDate: (ts) => new Date(ts * 1000),
    plugins: [wheelZoomPlugin()],
    series: [
      {
        label: 'Time',
        value: (_: uPlot, rawValue: number | null) =>
          rawValue ? new Date(rawValue * 1000).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' }) + ' WIB' : '--'
      },
      ...(isStatus
        ? [
            {
              label: 'Latency (UP ms)',
              stroke: '#3ECF8E',
              fill: 'rgba(62, 207, 142, 0.15)',
              width: 2,
              spanGaps: false,
              value: (_: uPlot, rawValue: number | null) => (rawValue != null ? `${rawValue.toFixed(0)} ms` : '--')
            },
            {
              label: '🔴 DOWN (100% Packet Loss)',
              stroke: '#F16565',
              fill: 'rgba(241, 101, 101, 0.65)',
              width: 3,
              spanGaps: false,
              points: {
                show: true,
                size: 8,
                fill: '#F16565',
                stroke: '#FFFFFF',
                width: 1.5
              },
              value: (_: uPlot, rawValue: number | null) => (rawValue != null ? '🔴 DOWN (100% Packet Loss)' : '--')
            }
          ]
        : [
            {
              label: activeMetric.value === 'cpu' ? 'CPU Utilization' : 'Memory Usage',
              stroke: activeMetric.value === 'cpu' ? '#7B96F5' : '#F59E0B',
              fill: activeMetric.value === 'cpu' ? 'rgba(123, 150, 245, 0.15)' : 'rgba(245, 158, 11, 0.15)',
              width: 2,
              spanGaps: false,
              value: (_: uPlot, rawValue: number | null) => (rawValue != null ? `${rawValue.toFixed(1)}%` : '--')
            }
          ])
    ],
    axes: [
      {
        stroke: '#9CA3AF',
        grid: { stroke: '#26262A', width: 1 },
        ticks: { stroke: '#26262A', width: 1 },
        font: '10px JetBrains Mono, monospace',
        values: (_: uPlot, ticks: number[]) =>
          ticks.map((t) => new Date(t * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }))
      },
      {
        stroke: '#9CA3AF',
        grid: { stroke: '#26262A', width: 1 },
        ticks: { stroke: '#26262A', width: 1 },
        font: '10px JetBrains Mono, monospace',
        values: (_: uPlot, ticks: number[]) => ticks.map((v) => (isStatus ? `${v.toFixed(0)} ms` : `${v.toFixed(0)}%`))
      }
    ],
    cursor: {
      drag: { setScale: true, x: true, y: false },
      focus: { prox: 30 },
      points: { size: 6, fill: '#FFFFFF' }
    },
    legend: {
      show: true
    }
  };

  uplotInstance = new uPlot(opts, data, chartRef.value);
}

function zoomIn() {
  if (!uplotInstance) return;
  const min = uplotInstance.scales.x.min!;
  const max = uplotInstance.scales.x.max!;
  const range = max - min;
  const center = min + range / 2;
  const newRange = range * 0.6;
  uplotInstance.setScale('x', { min: center - newRange / 2, max: center + newRange / 2 });
}

function zoomOut() {
  if (!uplotInstance) return;
  const min = uplotInstance.scales.x.min!;
  const max = uplotInstance.scales.x.max!;
  const range = max - min;
  const center = min + range / 2;
  const newRange = range * 1.5;
  uplotInstance.setScale('x', { min: center - newRange / 2, max: center + newRange / 2 });
}

function resetZoom() {
  if (uplotInstance && initialXMin != null && initialXMax != null) {
    uplotInstance.setScale('x', { min: initialXMin, max: initialXMax });
  }
}

function handleResize() {
  if (uplotInstance && chartRef.value) {
    const width = chartRef.value.getBoundingClientRect().width;
    if (width > 0) {
      uplotInstance.setSize({ width, height: 240 });
    }
  }
}

watch([activeMetric, activeRange, customFrom, customTo], fetchAndRender);

onMounted(() => {
  fetchAndRender();
  window.addEventListener('resize', handleResize);
  wsClient.connect();
  unsubscribeWS = wsClient.subscribe((data: any) => {
    if (data.type === 'LIVE_FEED' || data.type === 'STATUS_CHANGE') {
      handleRealtimeWSMessage(data);
    }
  });
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
  if (unsubscribeWS) {
    unsubscribeWS();
    unsubscribeWS = null;
  }
  if (uplotInstance) {
    uplotInstance.destroy();
    uplotInstance = null;
  }
});
</script>

<style>
.uplot {
  font-family: 'JetBrains Mono', monospace !important;
}
.uplot .u-legend {
  color: #9ca3af !important;
  font-size: 11px !important;
  font-family: 'JetBrains Mono', monospace !important;
  padding: 0 0 8px 0 !important;
}
.uplot .u-legend .u-series th {
  color: #d1d5db !important;
}
.uplot .u-select {
  background: rgba(123, 150, 245, 0.2) !important;
  border: 1px solid rgba(123, 150, 245, 0.6) !important;
}
</style>
