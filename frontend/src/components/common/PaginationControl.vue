<template>
  <div v-if="total > 0" class="flex flex-wrap items-center justify-between gap-3 bg-surface border border-subtle rounded-xl px-4 py-3 text-xs font-mono text-text-secondary">
    <div class="flex items-center gap-2">
      <span>Showing <strong class="text-text-main">{{ fromItem }}</strong> to <strong class="text-text-main">{{ toItem }}</strong> of <strong class="text-text-main">{{ total }}</strong> items</span>
    </div>

    <div class="flex items-center gap-3">
      <!-- Page Size Selector -->
      <div class="flex items-center gap-1.5">
        <span class="text-text-muted text-[11px]">Per page:</span>
        <select
          :value="pageSize"
          @change="onPageSizeChange"
          class="bg-card border border-subtle rounded px-2 py-1 text-text-main text-xs focus:outline-none focus:border-brand-periwinkle"
        >
          <option :value="5">5</option>
          <option :value="10">10</option>
          <option :value="25">25</option>
          <option :value="50">50</option>
        </select>
      </div>

      <!-- Page Controls -->
      <div class="flex items-center gap-1">
        <button
          @click="changePage(currentPage - 1)"
          :disabled="currentPage <= 1"
          class="px-2.5 py-1 rounded bg-card border border-subtle text-text-secondary hover:text-text-main hover:bg-subtle disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
        >
          Prev
        </button>

        <button
          v-for="p in visiblePages"
          :key="p"
          @click="changePage(p)"
          class="px-2.5 py-1 rounded border text-xs font-bold transition-colors"
          :class="p === currentPage ? 'bg-brand-periwinkle border-brand-periwinkle text-white' : 'bg-card border-subtle text-text-secondary hover:text-text-main hover:bg-hover'"
        >
          {{ p }}
        </button>

        <button
          @click="changePage(currentPage + 1)"
          :disabled="currentPage >= totalPages"
          class="px-2.5 py-1 rounded bg-card border border-subtle text-text-secondary hover:text-text-main hover:bg-subtle disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
        >
          Next
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = withDefaults(
  defineProps<{
    total: number;
    currentPage: number;
    pageSize: number;
  }>(),
  {
    total: 0,
    currentPage: 1,
    pageSize: 10
  }
);

const emit = defineEmits<{
  (e: 'update:currentPage', page: number): void;
  (e: 'update:pageSize', size: number): void;
  (e: 'change', payload: { page: number; pageSize: number }): void;
}>();

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)));

const fromItem = computed(() => (props.total === 0 ? 0 : (props.currentPage - 1) * props.pageSize + 1));
const toItem = computed(() => Math.min(props.total, props.currentPage * props.pageSize));

const visiblePages = computed(() => {
  const pages: number[] = [];
  const maxButtons = 5;
  let start = Math.max(1, props.currentPage - Math.floor(maxButtons / 2));
  let end = start + maxButtons - 1;

  if (end > totalPages.value) {
    end = totalPages.value;
    start = Math.max(1, end - maxButtons + 1);
  }

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }
  return pages;
});

function changePage(page: number) {
  if (page >= 1 && page <= totalPages.value && page !== props.currentPage) {
    emit('update:currentPage', page);
    emit('change', { page, pageSize: props.pageSize });
  }
}

function onPageSizeChange(e: Event) {
  const val = Number((e.target as HTMLSelectElement).value);
  emit('update:pageSize', val);
  emit('update:currentPage', 1);
  emit('change', { page: 1, pageSize: val });
}
</script>
