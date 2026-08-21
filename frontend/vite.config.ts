import { defineConfig, createLogger } from 'vite'
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
    code === 'ETIMEDOUT' ||
    err?.name === 'AggregateError' ||
    msg.includes('ECONNREFUSED') ||
    msg.includes('ECONNRESET') ||
    msg.includes('ECONNABORTED') ||
    msg.includes('EPIPE')
  );
};

const customLogger = createLogger();
const originalError = customLogger.error;
customLogger.error = (msg, options) => {
  if (
    msg.includes('ws proxy socket error') ||
    msg.includes('ws proxy error') ||
    msg.includes('http proxy error') ||
    msg.includes('ECONNABORTED') ||
    msg.includes('ECONNRESET') ||
    msg.includes('ECONNREFUSED') ||
    msg.includes('EPIPE')
  ) {
    return;
  }
  originalError(msg, options);
};

const originalWarn = customLogger.warn;
customLogger.warn = (msg, options) => {
  if (
    msg.includes('ws proxy') ||
    msg.includes('ECONNABORTED') ||
    msg.includes('ECONNRESET')
  ) {
    return;
  }
  originalWarn(msg, options);
};

const setupSilentProxy = (proxy: any) => {
  proxy.on('error', (err: any) => {
    if (isOfflineError(err)) {
      return; // Silence offline/aborted connection errors
    }
  });

  proxy.on('proxyReqWsError', (err: any) => {
    if (isOfflineError(err)) {
      return; // Silence WebSocket disconnect/abort errors
    }
  });

  proxy.on('proxyReqWs', (_proxyReq: any, _req: any, socket: any) => {
    socket.on('error', (err: any) => {
      if (isOfflineError(err)) {
        return;
      }
    });
  });

  const originalEmit = proxy.emit;
  proxy.emit = function (event: string, ...args: any[]) {
    if (event === 'error' || event === 'proxyReqWsError') {
      const err = args[0];
      if (isOfflineError(err)) {
        return false;
      }
    }
    return originalEmit.apply(this, [event, ...args]);
  };
};

// https://vite.dev/config/
export default defineConfig({
  customLogger,
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
        xfwd: true,
        configure: setupSilentProxy
      },
      '/ws': {
        target: 'ws://127.0.0.1:8080',
        ws: true,
        xfwd: true,
        configure: setupSilentProxy
      },
      '/uploads': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
        xfwd: true,
        configure: setupSilentProxy
      }
    }
  }
})
