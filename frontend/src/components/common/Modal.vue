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
          class="bg-[#151517] border border-[#26262A] rounded-xl shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col max-h-[90vh]"
        >
          <!-- Modal Header -->
          <div class="px-6 py-4 border-b border-[#26262A] flex items-center justify-between bg-[#18181B]">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <slot name="icon" />
              {{ title }}
            </h3>
            <button
              @click="$emit('close')"
              class="p-1 rounded-lg text-gray-400 hover:text-gray-200 hover:bg-[#26262A] transition-colors"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Modal Body -->
          <div class="p-6 overflow-y-auto space-y-4">
            <slot />
          </div>

          <!-- Modal Footer -->
          <div v-if="$slots.footer" class="px-6 py-4 border-t border-[#26262A] bg-[#18181B] flex items-center justify-end gap-3">
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
}>();

defineEmits(['close']);
</script>
