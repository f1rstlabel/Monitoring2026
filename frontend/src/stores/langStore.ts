import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { translations } from '../i18n/translations';

export type Language = 'id' | 'en';

export const useLangStore = defineStore('lang', () => {
  const saved = (localStorage.getItem('sanoc_help_lang') as Language) || 'id';
  const currentLang = ref<Language>(saved === 'en' ? 'en' : 'id');

  function setLang(lang: Language) {
    currentLang.value = lang;
    localStorage.setItem('sanoc_help_lang', lang);
  }

  function toggleLang() {
    setLang(currentLang.value === 'id' ? 'en' : 'id');
  }

  const isIndonesian = computed(() => currentLang.value === 'id');
  const isEnglish = computed(() => currentLang.value === 'en');

  // Khusus Pusat Bantuan
  const t = computed(() => translations[currentLang.value]);

  return {
    currentLang,
    isIndonesian,
    isEnglish,
    t,
    setLang,
    toggleLang
  };
});
