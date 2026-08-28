<template>
  <div v-if="!hasConsented" class="fixed bottom-0 left-0 right-0 z-50 p-4 animate-fade-in">
    <div class="max-w-4xl mx-auto bg-surface border border-subtle shadow-2xl p-5 flex flex-col md:flex-row items-center justify-between gap-4 rounded-xl">
      <div class="flex-1">
        <h3 class="text-text-main font-bold text-sm flex items-center gap-2">
          <Cookie class="w-4 h-4 text-brand-periwinkle" />
          Pengaturan Privasi & Cookie
        </h3>
        <p class="text-text-secondary text-xs mt-1.5 leading-relaxed">
          SANOC menggunakan <b>Cookie Persisten (Persistent Cookie)</b> berjenis <i>HttpOnly</i> yang bertahan selama 24 jam. Fungsinya murni untuk menyimpan status login secara aman agar Anda tidak perlu login ulang saat menutup browser. Kami <b>tidak</b> menggunakan cookie analitik maupun keranjang belanja.
        </p>
      </div>
      <div class="flex items-center gap-3 shrink-0 w-full md:w-auto">
        <button 
          @click="acceptEssentialOnly" 
          class="flex-1 md:flex-none px-4 py-2 border border-subtle bg-card text-text-secondary hover:bg-card hover:text-text-main font-semibold text-xs rounded-lg transition-colors cursor-pointer"
        >
          Hanya Esensial
        </button>
        <button 
          @click="acceptAllCookies" 
          class="flex-1 md:flex-none px-4 py-2 bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-bold text-xs rounded-lg shadow-md shadow-brand-periwinkle/20 transition-colors cursor-pointer"
        >
          Terima Semua
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Cookie } from 'lucide-vue-next';

const hasConsented = ref(true);

onMounted(() => {
  const consent = localStorage.getItem('sanoc_cookie_consent');
  if (!consent) {
    hasConsented.value = false;
  }
});

function acceptAllCookies() {
  localStorage.setItem('sanoc_cookie_consent', 'all');
  hasConsented.value = true;
}

function acceptEssentialOnly() {
  localStorage.setItem('sanoc_cookie_consent', 'essential');
  hasConsented.value = true;
}
</script>
