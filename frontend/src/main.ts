import { createApp } from 'vue';
import { createPinia } from 'pinia';
import VueApexCharts from 'vue3-apexcharts';
import router from './router';
import App from './App.vue';
import './styles/index.css';
import '@fontsource/jetbrains-mono';
import '@fontsource/inter';

const app = createApp(App);

app.use(createPinia());
app.use(router);
app.use(VueApexCharts as any);

// Wait until the initial route is resolved so /login never briefly renders the dashboard.
router.isReady().then(() => {
  app.mount('#app');
});
