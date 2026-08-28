import { ref, watchEffect } from 'vue';

const isDark = ref(false);

export function useTheme() {
  const initTheme = () => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem('theme');
      if (stored === 'dark') {
        isDark.value = true;
      } else if (stored === 'light') {
        isDark.value = false;
      } else {
        isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches;
      }
    }
  };

  watchEffect(() => {
    if (typeof window !== 'undefined') {
      if (isDark.value) {
        document.documentElement.classList.add('dark');
        localStorage.setItem('theme', 'dark');
      } else {
        document.documentElement.classList.remove('dark');
        localStorage.setItem('theme', 'light');
      }
    }
  });

  const toggleTheme = () => {
    isDark.value = !isDark.value;
  };

  return {
    isDark,
    initTheme,
    toggleTheme
  };
}
