<template>
  <div
    class="fixed top-5 right-5 z-[99999] flex flex-col gap-2.5 max-w-sm w-full pointer-events-none px-4 sm:px-0"
    aria-live="polite"
  >
    <TransitionGroup
      enter-active-class="transform ease-out duration-300 transition"
      enter-from-class="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-4 scale-95"
      enter-to-class="translate-y-0 opacity-100 sm:translate-x-0 scale-100"
      leave-active-class="transition ease-in duration-200"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95 translate-x-4"
    >
      <div
        v-for="item in toastStore.toasts"
        :key="item.id"
        class="pointer-events-auto w-full rounded-2xl p-4 border shadow-2xl backdrop-blur-xl flex items-start gap-3.5 transition-all font-mono text-xs"
        :class="getToastClasses(item.type)"
      >
        <!-- Icon -->
        <div class="shrink-0 mt-0.5">
          <CheckCircle2 v-if="item.type === 'success'" class="w-4 h-4 text-emerald-400" />
          <AlertTriangle v-else-if="item.type === 'error'" class="w-4 h-4 text-red-400" />
          <AlertCircle v-else-if="item.type === 'warning'" class="w-4 h-4 text-amber-400" />
          <Info v-else class="w-4 h-4 text-brand-periwinkle" />
        </div>

        <!-- Content -->
        <div class="flex-1 min-w-0 space-y-1">
          <h4 class="font-bold text-text-main text-xs tracking-tight">
            {{ item.title }}
          </h4>
          <p class="text-[11px] text-text-secondary leading-relaxed font-sans break-words">
            {{ item.message }}
          </p>
        </div>

        <!-- Close Button -->
        <button
          type="button"
          @click="toastStore.removeToast(item.id)"
          class="shrink-0 text-text-muted hover:text-text-main p-1 rounded-lg hover:bg-white/5 transition-colors cursor-pointer"
          title="Tutup Notifikasi"
        >
          <X class="w-3.5 h-3.5" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { useToastStore, type ToastType } from '../../stores/toastStore';
import { CheckCircle2, AlertTriangle, AlertCircle, Info, X } from 'lucide-vue-next';

const toastStore = useToastStore();

function getToastClasses(type: ToastType) {
  switch (type) {
    case 'success':
      return 'bg-surface/95 border-emerald-500/30 shadow-emerald-500/10 text-emerald-300';
    case 'error':
      return 'bg-surface/95 border-red-500/30 shadow-red-500/10 text-red-300';
    case 'warning':
      return 'bg-surface/95 border-amber-500/30 shadow-amber-500/10 text-amber-300';
    case 'info':
    default:
      return 'bg-surface/95 border-brand-periwinkle/30 shadow-brand-periwinkle/10 text-brand-periwinkle';
  }
}
</script>
