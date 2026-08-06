<template>
  <!-- Matches exact stat-card dimensions to prevent layout shift -->
  <div class="bg-[#151517] border border-[#26262A] rounded-xl p-5 space-y-3" aria-hidden="true">
    <!-- Icon + label row -->
    <div class="flex items-center justify-between">
      <Skeleton width="7rem" height="0.75rem" />
      <Skeleton width="2rem" height="2rem" customClass="rounded-lg" />
    </div>
    <!-- Value -->
    <Skeleton width="4rem" height="2rem" customClass="rounded-lg" />
    <!-- Subtitle -->
    <Skeleton width="60%" height="0.65rem" />
  </div>
</template>

<script setup lang="ts">
import Skeleton from './Skeleton.vue';
</script>
