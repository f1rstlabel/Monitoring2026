import { defineStore } from 'pinia';
import { ref } from 'vue';

export type Theme = 'dark';

export const useThemeStore = defineStore('theme', () => {
  const currentTheme = ref<Theme>('dark');

  function initTheme() {
    currentTheme.value = 'dark';
    document.documentElement.setAttribute('data-theme', 'dark');
    document.documentElement.classList.add('dark');
  }

  return {
    currentTheme,
    initTheme
  };
});
