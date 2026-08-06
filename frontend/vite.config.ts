import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

const isOfflineError = (err: any) => {
  const code = err?.code;
  const msg = err?.message || '';
  return (
    code === 'ECONNREFUSED' ||
    code === 'ECONNRESET' ||
    code === 'ECONNABORTED' ||
    code === 'EPIPE' ||
    code === 'ENOTFOUND' ||
    err?.name === 'AggregateError' ||
    msg.includes('ECONNREFUSED') ||
    msg.includes('ECONNRESET') ||
    msg.includes('ECONNABORTED')
  );
};

const setupSilentProxy = (proxy: any) => {
  const originalEmit = proxy.emit;
  proxy.emit = function (event: string, ...args: any[]) {
    if (event === 'error') {
      const err = args[0];
      if (isOfflineError(err)) {
        return false; // Silence ECONNREFUSED when backend is offline
      }
    }
    return originalEmit.apply(this, [event, ...args]);
  };
};

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
        configure: setupSilentProxy
      },
      '/ws': {
        target: 'ws://127.0.0.1:8080',
        ws: true,
        configure: setupSilentProxy
      }
    }
  }
})


