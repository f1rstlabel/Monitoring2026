import { createRouter, createWebHistory } from 'vue-router';
import LoginView from '../views/LoginView.vue';
import DashboardView from '../views/DashboardView.vue';
import DevicesView from '../views/DevicesView.vue';
import DeviceDetailView from '../views/DeviceDetailView.vue';
import IncidentDetailView from '../views/IncidentDetailView.vue';
import IncidentsView from '../views/IncidentsView.vue';
import SettingsView from '../views/SettingsView.vue';
import ReportsView from '../views/ReportsView.vue';
import UserProfileView from '../views/UserProfileView.vue';
import { useAuthStore } from '../stores/authStore';

const routes = [
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    meta: { guestOnly: true }
  },
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: DashboardView,
    meta: { requiresAuth: true }
  },
  {
    path: '/devices',
    name: 'devices',
    component: DevicesView,
    meta: { requiresAuth: true, featureKey: 'devices.view' }
  },
  {
    path: '/devices/:id',
    alias: '/device/:id',
    name: 'device-detail',
    component: DeviceDetailView,
    meta: { requiresAuth: true, featureKey: 'devices.view' }
  },
  {
    path: '/incidents',
    name: 'incidents',
    component: IncidentsView,
    meta: { requiresAuth: true, featureKey: 'incidents.view' }
  },
  {
    path: '/incidents/:id',
    alias: '/incident/:id',
    name: 'incident-detail',
    component: IncidentDetailView,
    meta: { requiresAuth: true, featureKey: 'incidents.view' }
  },
  {
    path: '/reports',
    name: 'reports',
    component: ReportsView,
    meta: { requiresAuth: true, featureKey: 'reports.view' }
  },
  {
    path: '/settings',
    name: 'settings',
    component: SettingsView,
    meta: { requiresAuth: true }
  },
  {
    path: '/profile',
    name: 'profile',
    component: UserProfileView,
    meta: { requiresAuth: true }
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore();
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login');
    return;
  }
  if (to.meta.guestOnly && authStore.isAuthenticated) {
    next('/dashboard');
    return;
  }
  if (authStore.isAuthenticated && to.meta.featureKey) {
    if (!authStore.hasPermission(to.meta.featureKey as string)) {
      next('/dashboard');
      return;
    }
  }
  next();
});

export default router;
