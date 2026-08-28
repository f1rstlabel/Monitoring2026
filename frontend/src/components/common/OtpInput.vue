<template>
  <div class="flex items-center justify-center gap-2 sm:gap-3" @paste="handlePaste">
    <input
      v-for="(_, idx) in length"
      :key="idx"
      :ref="(el) => setInputRef(el, idx)"
      type="text"
      inputmode="numeric"
      pattern="[0-9]*"
      maxlength="1"
      :value="digits[idx] || ''"
      :disabled="disabled"
      @input="handleInput(idx, $event)"
      @keydown="handleKeyDown(idx, $event)"
      @focus="handleFocus(idx)"
      class="w-10 h-12 sm:w-12 sm:h-14 bg-card border rounded-xl text-center text-xl font-bold font-mono text-text-main transition-all focus:outline-none select-none"
      :class="[
        error
          ? 'border-red-500/80 text-red-400 bg-red-500/5 focus:border-red-500 focus:ring-2 focus:ring-red-500/20'
          : digits[idx]
          ? 'border-brand-periwinkle bg-brand-periwinkle/10 text-white shadow-sm shadow-brand-periwinkle/20'
          : 'border-subtle text-text-secondary focus:border-brand-periwinkle focus:ring-2 focus:ring-brand-periwinkle/20 hover:border-brand-periwinkle/50',
        disabled ? 'opacity-40 cursor-not-allowed' : ''
      ]"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue';

const props = withDefaults(
  defineProps<{
    modelValue?: string;
    length?: number;
    disabled?: boolean;
    error?: boolean;
    autoFocus?: boolean;
  }>(),
  {
    modelValue: '',
    length: 6,
    disabled: false,
    error: false,
    autoFocus: true
  }
);

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void;
  (e: 'complete', value: string): void;
}>();

const digits = ref<string[]>(Array(props.length).fill(''));
const inputRefs = ref<HTMLInputElement[]>([]);

function setInputRef(el: any, idx: number) {
  if (el) {
    inputRefs.value[idx] = el as HTMLInputElement;
  }
}

watch(
  () => props.modelValue,
  (newVal) => {
    const val = newVal || '';
    const arr = Array(props.length).fill('');
    for (let i = 0; i < props.length && i < val.length; i++) {
      arr[i] = val[i];
    }
    digits.value = arr;
  },
  { immediate: true }
);

function updateModel() {
  const fullVal = digits.value.join('');
  emit('update:modelValue', fullVal);
  if (fullVal.length === props.length && !digits.value.includes('')) {
    emit('complete', fullVal);
  }
}

function handleInput(idx: number, event: Event) {
  const target = event.target as HTMLInputElement;
  const rawVal = target.value;
  const numOnly = rawVal.replace(/\D/g, '');

  if (!numOnly) {
    digits.value[idx] = '';
    target.value = '';
    updateModel();
    return;
  }

  // Set the digit
  digits.value[idx] = numOnly.slice(-1);
  target.value = digits.value[idx];
  updateModel();

  // Advance to next input
  if (idx < props.length - 1) {
    nextTick(() => {
      inputRefs.value[idx + 1]?.focus();
      inputRefs.value[idx + 1]?.select();
    });
  }
}

function handleKeyDown(idx: number, event: KeyboardEvent) {
  if (event.key === 'Backspace') {
    if (!digits.value[idx] && idx > 0) {
      event.preventDefault();
      digits.value[idx - 1] = '';
      updateModel();
      nextTick(() => {
        inputRefs.value[idx - 1]?.focus();
      });
    } else {
      digits.value[idx] = '';
      updateModel();
    }
  } else if (event.key === 'ArrowLeft' && idx > 0) {
    event.preventDefault();
    inputRefs.value[idx - 1]?.focus();
  } else if (event.key === 'ArrowRight' && idx < props.length - 1) {
    event.preventDefault();
    inputRefs.value[idx + 1]?.focus();
  }
}

function handlePaste(event: ClipboardEvent) {
  event.preventDefault();
  const pasteData = event.clipboardData?.getData('text') || '';
  const cleanNum = pasteData.replace(/\D/g, '').slice(0, props.length);

  if (!cleanNum) return;

  const newDigits = Array(props.length).fill('');
  for (let i = 0; i < props.length; i++) {
    if (i < cleanNum.length) {
      newDigits[i] = cleanNum[i];
    }
  }
  digits.value = newDigits;
  updateModel();

  const nextFocus = Math.min(cleanNum.length, props.length - 1);
  nextTick(() => {
    inputRefs.value[nextFocus]?.focus();
  });
}

function handleFocus(idx: number) {
  nextTick(() => {
    inputRefs.value[idx]?.select();
  });
}

onMounted(() => {
  if (props.autoFocus && inputRefs.value[0]) {
    inputRefs.value[0].focus();
  }
});
</script>
