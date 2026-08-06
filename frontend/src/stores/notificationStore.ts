import { defineStore } from 'pinia';
import api from '../api/client';

export interface AppNotification {
  id: string;
  type: 'INCIDENT_NEW' | 'INCIDENT_RESOLVED' | 'FLAP_ALERT' | 'GATEWAY_DISCONNECTED';
  title: string;
  message: string;
  targetUrl: string;
  isUnread: boolean;
  timestamp: string;
}

export const useNotificationStore = defineStore('notifications', {
  state: () => ({
    notifications: [] as AppNotification[],
    isLoading: false
  }),
  getters: {
    unreadCount: (state) => state.notifications.filter(n => n.isUnread).length
  },
  actions: {
    async fetchNotifications() {
      this.isLoading = true;
      try {
        const res = await api.get('/notifications');
        if (res.data && res.data.notifications) {
          this.notifications = res.data.notifications;
        }
      } catch (e) {
        // keep fallback array
      } finally {
        this.isLoading = false;
      }
    },
    async markAllAsRead() {
      this.notifications.forEach(n => { n.isUnread = false; });
      try {
        await api.patch('/notifications/read-all');
      } catch (e) {
        // ignore
      }
    },
    addRealtimeNotification(item: AppNotification) {
      this.notifications.unshift(item);
    }
  }
});
