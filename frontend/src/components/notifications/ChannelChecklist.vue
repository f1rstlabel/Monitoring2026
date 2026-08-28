<template>
  <div class="space-y-1.5 font-mono text-xs">
    <div v-for="item in channelStatuses" :key="item.channel" class="flex items-center gap-2">
      <!-- Status Icon -->
      <span v-if="item.status === 'delivered'" class="text-emerald-400 font-bold flex items-center gap-1">
        <CheckCircle2 class="w-3.5 h-3.5" />
        <span>{{ item.channelName }} &mdash; Delivered</span>
        <span v-if="item.timestamp" class="text-[10px] text-text-muted">({{ item.timestamp }})</span>
      </span>

      <span v-else-if="item.status === 'failed'" class="text-red-400 font-bold flex items-center gap-1">
        <XCircle class="w-3.5 h-3.5" />
        <span>{{ item.channelName }} &mdash; Failed</span>
        <span v-if="item.error" class="text-[10px] text-red-400/80">({{ item.error }})</span>
        <span class="text-[10px] text-amber-400 font-semibold">&rarr; Fallback</span>
      </span>

      <span v-else class="text-text-muted flex items-center gap-1">
        <SkipForward class="w-3.5 h-3.5 text-text-muted" />
        <span>{{ item.channelName }} &mdash; Skipped</span>
        <span class="text-[10px] text-text-muted">({{ item.note || 'Prior channel delivered' }})</span>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { CheckCircle2, XCircle, SkipForward } from 'lucide-vue-next';

interface TimelineEvent {
  channel?: string;
  eventType?: string;
  detail?: string;
  status?: string;
  timestamp?: string;
  error?: string;
}

const props = defineProps<{
  events?: TimelineEvent[];
  logRow?: { channel: string; status: string; channelIcon?: string; timestamp?: string };
}>();

interface ChannelStatusItem {
  channel: string;
  channelName: string;
  status: 'delivered' | 'failed' | 'skipped';
  timestamp?: string;
  error?: string;
  note?: string;
}

const channelStatuses = computed<ChannelStatusItem[]>(() => {
  const result: ChannelStatusItem[] = [];

  if (props.logRow) {
    const isWA = props.logRow.channel.toLowerCase().includes('whatsapp');
    const isSuccess = props.logRow.status.toLowerCase() === 'delivered' || props.logRow.status.toLowerCase() === 'sent';

    if (isWA) {
      if (isSuccess) {
        result.push({
          channel: 'whatsapp',
          channelName: 'WhatsApp',
          status: 'delivered',
          timestamp: props.logRow.timestamp
        });
        result.push({
          channel: 'telegram',
          channelName: 'Telegram',
          status: 'skipped',
          note: 'WhatsApp delivered successfully'
        });
      } else {
        result.push({
          channel: 'whatsapp',
          channelName: 'WhatsApp',
          status: 'failed',
          error: props.logRow.channelIcon || 'Failed to deliver',
          timestamp: props.logRow.timestamp
        });
        result.push({
          channel: 'telegram',
          channelName: 'Telegram',
          status: 'delivered',
          timestamp: props.logRow.timestamp
        });
      }
    } else {
      result.push({
        channel: 'whatsapp',
        channelName: 'WhatsApp',
        status: 'failed',
        error: 'WhatsApp sidecar unreachable / failed',
        timestamp: props.logRow.timestamp
      });
      result.push({
        channel: 'telegram',
        channelName: 'Telegram',
        status: isSuccess ? 'delivered' : 'failed',
        timestamp: props.logRow.timestamp
      });
    }
    return result;
  }

  // Parse events timeline if available
  let waDelivered = false;
  let waFailed = false;
  let waTime = '';
  let waErr = '';

  let tgDelivered = false;
  let tgFailed = false;
  let tgTime = '';
  let tgErr = '';

  if (props.events) {
    for (const evt of props.events) {
      const ch = (evt.channel || '').toLowerCase();
      const type = (evt.eventType || '').toLowerCase();
      if (ch.includes('whatsapp')) {
        if (type.includes('delivered')) {
          waDelivered = true;
          waTime = evt.timestamp || '';
        } else if (type.includes('failed')) {
          waFailed = true;
          waErr = evt.detail || 'Failed';
          waTime = evt.timestamp || '';
        }
      } else if (ch.includes('telegram')) {
        if (type.includes('delivered')) {
          tgDelivered = true;
          tgTime = evt.timestamp || '';
        } else if (type.includes('failed')) {
          tgFailed = true;
          tgErr = evt.detail || 'Failed';
          tgTime = evt.timestamp || '';
        }
      }
    }
  }

  if (waDelivered) {
    result.push({ channel: 'whatsapp', channelName: 'WhatsApp', status: 'delivered', timestamp: waTime });
    result.push({ channel: 'telegram', channelName: 'Telegram', status: 'skipped', note: 'WhatsApp delivered successfully' });
  } else if (waFailed) {
    result.push({ channel: 'whatsapp', channelName: 'WhatsApp', status: 'failed', error: waErr, timestamp: waTime });
    if (tgDelivered) {
      result.push({ channel: 'telegram', channelName: 'Telegram', status: 'delivered', timestamp: tgTime });
    } else if (tgFailed) {
      result.push({ channel: 'telegram', channelName: 'Telegram', status: 'failed', error: tgErr, timestamp: tgTime });
    } else {
      result.push({ channel: 'telegram', channelName: 'Telegram', status: 'delivered', timestamp: waTime });
    }
  } else {
    // Default fallback display
    result.push({ channel: 'whatsapp', channelName: 'WhatsApp', status: 'delivered' });
    result.push({ channel: 'telegram', channelName: 'Telegram', status: 'skipped', note: 'WhatsApp delivered successfully' });
  }

  return result;
});
</script>
