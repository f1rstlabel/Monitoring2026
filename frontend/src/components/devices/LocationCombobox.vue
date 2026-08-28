<template>
  <div class="relative w-full">
    <div class="relative">
      <input
        type="text"
        v-model="query"
        @focus="isOpen = true"
        @input="handleInput"
        @keydown.enter.prevent="handleEnter"
        placeholder="Ketik kata kunci atau pilih lokasi..."
        class="w-full bg-card border border-subtle rounded-lg pl-3 pr-8 py-2 text-xs text-text-main focus:outline-none focus:border-brand-periwinkle"
      />
      <ChevronDown class="w-4 h-4 text-text-secondary absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none" />
    </div>

    <!-- Dropdown Menu -->
    <div
      v-if="isOpen"
      class="absolute left-0 right-0 mt-1 max-h-48 overflow-y-auto bg-surface border border-subtle rounded-lg shadow-xl z-50 divide-y divide-subtle text-xs"
    >
      <!-- Loading indicator -->
      <div v-if="isLoading" class="p-2 text-center text-text-muted font-mono text-[11px]">
        Mencari data lokasi...
      </div>

      <template v-else>
        <!-- Existing matching locations -->
        <div
          v-for="loc in filteredLocations"
          :key="loc.id"
          @click="selectLocation(loc)"
          class="p-2.5 hover:bg-card cursor-pointer flex items-center justify-between transition-colors"
          :class="modelValue === loc.name ? 'bg-brand-periwinkle/10 text-brand-periwinkle font-semibold' : 'text-text-main'"
        >
          <div class="flex items-center gap-2 truncate">
            <MapPin class="w-3.5 h-3.5 text-brand-periwinkle shrink-0" />
            <span class="truncate">{{ loc.name }}</span>
          </div>
          <Check v-if="modelValue === loc.name" class="w-3.5 h-3.5 text-brand-periwinkle shrink-0" />
        </div>

        <!-- Create new location option when query doesn't match an existing location exactly -->
        <div
          v-if="query.trim() && !exactMatch"
          @click="createNewLocation"
          class="p-2.5 hover:bg-brand-periwinkle/15 cursor-pointer text-brand-periwinkle font-semibold flex items-center gap-2 transition-colors bg-brand-periwinkle/5"
        >
          <Plus class="w-3.5 h-3.5 shrink-0" />
          <span>Buat lokasi baru: "{{ query.trim() }}"</span>
        </div>

        <!-- Empty state when no query and no locations -->
        <div v-if="filteredLocations.length === 0 && !query.trim()" class="p-3 text-center text-text-muted text-[11px]">
          Belum ada data lokasi. Ketik nama lokasi baru di atas.
        </div>
      </template>
    </div>

    <!-- Backdrop for closing dropdown -->
    <div v-if="isOpen" @click="isOpen = false" class="fixed inset-0 z-40 bg-transparent"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { locationsApi } from '../../api';
import { MapPin, ChevronDown, Check, Plus } from 'lucide-vue-next';

interface LocationItem {
  id: string;
  name: string;
  description?: string;
}

const props = defineProps<{
  modelValue: string;
  locationId?: string;
}>();

const emit = defineEmits(['update:modelValue', 'update:locationId', 'selected']);

const query = ref(props.modelValue || '');
const isOpen = ref(false);
const isLoading = ref(false);
const locations = ref<LocationItem[]>([]);

watch(() => props.modelValue, (newVal) => {
  query.value = newVal || '';
});

const filteredLocations = computed(() => {
  const q = query.value.trim().toLowerCase();
  if (!q) return locations.value;
  return locations.value.filter(l => l.name.toLowerCase().includes(q));
});

const exactMatch = computed(() => {
  const q = query.value.trim().toLowerCase();
  return locations.value.some(l => l.name.toLowerCase() === q);
});

async function fetchLocations() {
  isLoading.value = true;
  try {
    const res = await locationsApi.getLocations();
    locations.value = res || [];
  } catch (e) {
    console.error('Failed to fetch locations:', e);
  } finally {
    isLoading.value = false;
  }
}

function handleInput() {
  isOpen.value = true;
  emit('update:modelValue', query.value);
  const match = locations.value.find(l => l.name.toLowerCase() === query.value.trim().toLowerCase());
  if (match) {
    emit('update:locationId', match.id);
  } else {
    emit('update:locationId', '');
  }
}

function handleEnter() {
  const match = locations.value.find(l => l.name.toLowerCase() === query.value.trim().toLowerCase());
  if (match) {
    selectLocation(match);
  } else if (query.value.trim()) {
    createNewLocation();
  }
}

function selectLocation(loc: LocationItem) {
  query.value = loc.name;
  emit('update:modelValue', loc.name);
  emit('update:locationId', loc.id);
  emit('selected', loc);
  isOpen.value = false;
}

async function createNewLocation() {
  const name = query.value.trim();
  if (!name) return;
  isLoading.value = true;
  try {
    const created = await locationsApi.createLocation(name);
    if (created) {
      locations.value.push(created);
      selectLocation(created);
    }
  } catch (e) {
    console.error('Failed to create location:', e);
  } finally {
    isLoading.value = false;
  }
}

onMounted(() => {
  fetchLocations();
});
</script>
