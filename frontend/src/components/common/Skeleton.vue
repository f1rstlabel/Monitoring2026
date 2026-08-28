<template>
  <div
    :class="[
      'relative overflow-hidden rounded-[10px] bg-card border border-subtle',
      customClass
    ]"
    :style="{ width: width, height: height }"
    aria-hidden="true"
  >
    <!-- Diagonal shimmer sweep -->
    <div class="skeleton-shimmer" />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  width?: string;
  height?: string;
  customClass?: string;
}>(), {
  width: '100%',
  height: '1rem',
  customClass: ''
});
</script>

<style scoped>
.skeleton-shimmer {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    105deg,
    transparent 30%,
    rgba(123, 150, 245, 0.06) 50%,
    transparent 70%
  );
  background-size: 200% 100%;
  animation: shimmer 1.6s ease-in-out infinite;
}

@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
