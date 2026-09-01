import { defineStore } from 'pinia';
import { ref } from 'vue';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface ToastItem {
  id: string;
  title: string;
  message: string;
  type: ToastType;
  duration: number;
}

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<ToastItem[]>([]);

  function showToast(
    title: string,
    message: string,
    type: ToastType = 'success',
    duration = 4000
  ) {
    const id = 'toast-' + Date.now() + '-' + Math.random().toString(36).substring(2, 7);
    const item: ToastItem = { id, title, message, type, duration };
    toasts.value.push(item);

    if (duration > 0) {
      setTimeout(() => {
        removeToast(id);
      }, duration);
    }
    return id;
  }

  function removeToast(id: string) {
    toasts.value = toasts.value.filter((t) => t.id !== id);
  }

  function success(title: string, message: string, duration = 4000) {
    return showToast(title, message, 'success', duration);
  }

  function error(title: string, message: string, duration = 5000) {
    return showToast(title, message, 'error', duration);
  }

  function info(title: string, message: string, duration = 4000) {
    return showToast(title, message, 'info', duration);
  }

  function warning(title: string, message: string, duration = 4500) {
    return showToast(title, message, 'warning', duration);
  }

  return {
    toasts,
    showToast,
    removeToast,
    success,
    error,
    info,
    warning
  };
});
