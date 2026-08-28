<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div 
        v-if="isOpen" 
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm"
        @click.self="$emit('close')"
      >
        <div 
          class="bg-surface border border-subtle rounded-xl shadow-2xl w-full overflow-hidden flex flex-col max-h-[90vh]"
          :class="maxWidth || 'max-w-2xl'"
        >
          <!-- Modal Header -->
          <div class="px-6 py-4 border-b border-subtle flex items-center justify-between bg-card">
            <h3 class="text-base font-bold text-text-main flex items-center gap-2">
              <slot name="icon" />
              {{ title }}
            </h3>
            <button
              @click="$emit('close')"
              class="p-1 rounded-lg text-text-secondary hover:text-text-main hover:bg-subtle transition-colors"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Modal Body -->
          <div class="p-6 overflow-y-auto space-y-4">
            <slot />
          </div>

          <!-- Modal Footer -->
          <div v-if="$slots.footer" class="px-6 py-4 border-t border-subtle bg-card flex items-center justify-end gap-3">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { X } from 'lucide-vue-next';

defineProps<{
  isOpen: boolean;
  title: string;
  maxWidth?: string;
}>();

defineEmits(['close']);
</script>
