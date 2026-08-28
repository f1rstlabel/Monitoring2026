<template>
  <div class="bg-surface border border-subtle rounded-xl overflow-hidden" aria-hidden="true">
    <!-- Header row -->
    <div class="bg-card border-b border-subtle flex items-center gap-4 px-4 py-3">
      <Skeleton v-for="i in cols" :key="i" height="0.625rem" :width="colWidths[(i - 1) % colWidths.length]" />
    </div>
    <!-- Data rows -->
    <div class="divide-y divide-subtle">
      <div
        v-for="row in rows"
        :key="row"
        class="flex items-center gap-4 px-4 py-3.5"
      >
        <!-- First col wider (name) -->
        <Skeleton height="0.75rem" width="30%" />
        <Skeleton height="0.625rem" :width="colWidths[1 % colWidths.length]" />
        <Skeleton height="0.625rem" :width="colWidths[2 % colWidths.length]" />
        <Skeleton v-if="cols > 3" height="0.625rem" :width="colWidths[3 % colWidths.length]" />
        <Skeleton v-if="cols > 4" height="0.625rem" :width="colWidths[4 % colWidths.length]" />
        <!-- Status pill-shaped skeleton -->
        <Skeleton v-if="cols > 5" height="1.3rem" width="3rem" customClass="rounded-full ml-auto" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Skeleton from './Skeleton.vue';

withDefaults(defineProps<{
  rows?: number;
  cols?: number;
}>(), {
  rows: 8,
  cols: 6
});

const colWidths = ['20%', '12%', '14%', '16%', '10%', '8%'];
</script>
