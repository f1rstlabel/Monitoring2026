import dotenv from 'dotenv';
import path from 'path';
import { fileURLToPath } from 'url';
import express from 'express';
import cors from 'cors';
import makeWASocket, {
  useMultiFileAuthState,
  DisconnectReason,
  fetchLatestBaileysVersion
} from '@whiskeysockets/baileys';
import pino from 'pino';
import fs from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Load .env from sidecar dir and parent backend dir
dotenv.config();
dotenv.config({ path: path.resolve(__dirname, '../.env') });
dotenv.config({ path: path.resolve(__dirname, './.env') });

// ─── Process-Level Crash Protection ──────────────────────────────────────────
process.on('uncaughtException', (err) => {
  console.error('[Baileys Sidecar] Uncaught Exception caught (prevented crash):', err?.message || err);
});

process.on('unhandledRejection', (reason) => {
  console.error('[Baileys Sidecar] Unhandled Rejection caught (prevented crash):', reason?.message || reason);
});

const app = express();
app.use(cors());
app.use(express.json());

const PORT = process.env.PORT || 3001;
const AUTH_DIR = path.resolve(process.env.WHATSAPP_SESSION_PATH || './auth_sessions/whatsapp');

function getInternalToken() {
  return (process.env.INTERNAL_TOKEN || process.env.WA_SIDECAR_TOKEN || '').trim();
}

// ─── Internal Token Auth Middleware ─────────────────────────────────────────
function requireToken(req, res, next) {
  const token = getInternalToken();
  if (!token) return next(); // no token configured → open (dev mode)

  const header = req.headers['x-internal-token'] || req.headers['authorization'];
  const raw = Array.isArray(header) ? header[0] : (header || '');
  const provided = raw.trim().replace(/^Bearer\s+/i, '').trim();

  if (provided !== token) {
    console.warn(`[Baileys Sidecar] 401 Unauthorized: token mismatch for ${req.method} ${req.path}`);
    return res.status(401).json({ error: 'Unauthorized: invalid internal token' });
  }
  return next();
}

// Skip auth for health (so Go backend can heartbeat without token)
app.get('/health', (req, res) => {
  let hasSavedAuth = false;
  try {
    hasSavedAuth = fs.existsSync(AUTH_DIR) && fs.readdirSync(AUTH_DIR).length > 0;
  } catch (e) {}

  const isHealthy = connectionStatus === 'connected' || (connectionStatus === 'pending' && hasSavedAuth);
  res.status(isHealthy ? 200 : 503).json({ status: connectionStatus, healthy: isHealthy, hasSavedAuth });
});

// Apply auth to all other routes
app.use(requireToken);

// ─── Connection State ────────────────────────────────────────────────────────
let sock = null;
let qrCode = '';
let connectionStatus = 'disconnected'; // 'disconnected' | 'pending' | 'connected'
let linkedNumber = '';
let isConnecting = false;
let reconnectTimer = null;
let reconnectAttempts = 0;

function cleanupExistingSocket() {
  if (sock) {
    try {
      sock.ev.removeAllListeners();
      sock.end(undefined);
    } catch (e) {
      // ignore cleanup errors
    }
    sock = null;
  }
}

function clearAuthDir() {
  cleanupExistingSocket();
  if (fs.existsSync(AUTH_DIR)) {
    try {
      const files = fs.readdirSync(AUTH_DIR);
      for (const file of files) {
        const curPath = path.join(AUTH_DIR, file);
        try {
          if (fs.lstatSync(curPath).isDirectory()) {
            fs.rmSync(curPath, { recursive: true, force: true });
          } else {
            fs.unlinkSync(curPath);
          }
        } catch (e) {
          // ignore single-file permission locks on Windows
        }
      }
      try {
        fs.rmSync(AUTH_DIR, { recursive: true, force: true });
      } catch (e) {}
      console.log('[Baileys Sidecar] Auth directory cleared successfully:', AUTH_DIR);
    } catch (err) {
      console.error('[Baileys Sidecar] Failed to clear auth dir:', err?.message || err);
    }
  }
}

// ─── Baileys Connection ──────────────────────────────────────────────────────
async function connectToWhatsApp() {
  if (isConnecting) {
    console.log('[Baileys Sidecar] Connection attempt already in progress, skipping duplicate call.');
    return;
  }
  isConnecting = true;

  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }

  try {
    if (!fs.existsSync(AUTH_DIR)) {
      fs.mkdirSync(AUTH_DIR, { recursive: true });
    }

    cleanupExistingSocket();

    const { state, saveCreds } = await useMultiFileAuthState(AUTH_DIR);

    let version;
    try {
      const fetched = await fetchLatestBaileysVersion();
      version = fetched.version;
      console.log(`[Baileys Sidecar] Using WhatsApp WA version: ${version.join('.')}`);
    } catch (e) {
      version = [2, 3000, 1015901307];
      console.warn('[Baileys Sidecar] Could not fetch latest WA version, using fallback:', version.join('.'));
    }

    sock = makeWASocket({
      version,
      auth: state,
      printQRInTerminal: false,
      logger: pino({ level: 'silent' }),
      browser: ['SANOC Monitoring', 'Chrome', '126.0.0'],
      generateHighQualityLinkPreview: false,
      syncFullHistory: false,
      markOnlineOnConnect: false,
      keepAliveIntervalMs: 25000,
      connectTimeoutMs: 60000,
      defaultQueryTimeoutMs: 60000,
      retryRequestDelayMs: 500,
      maxMsgRetryCount: 5
    });

    sock.ev.on('creds.update', async () => {
      try {
        await saveCreds();
      } catch (err) {
        console.error('[Baileys Sidecar] Failed to save creds:', err?.message || err);
      }
    });

    sock.ev.on('connection.update', (update) => {
      const { connection, lastDisconnect, qr } = update;

      if (qr) {
        qrCode = qr;
        connectionStatus = 'pending';
        console.log(`[Baileys Sidecar] New QR code ready (length: ${qr.length}). Waiting for WhatsApp scan...`);
      }

      if (connection === 'close') {
        const statusCode = lastDisconnect?.error?.output?.statusCode;
        const errMsg = lastDisconnect?.error?.message || '';
        const isLoggedOut = statusCode === DisconnectReason.loggedOut; // 401
        const isRestartRequired = statusCode === DisconnectReason.restartRequired; // 515 or 405

        console.log(`[Baileys Sidecar] Connection closed. statusCode=${statusCode} (${errMsg}) loggedOut=${isLoggedOut} restartRequired=${isRestartRequired}`);

        if (isLoggedOut) {
          connectionStatus = 'disconnected';
          qrCode = '';
          linkedNumber = '';
          reconnectAttempts = 0;
          clearAuthDir();
          console.log('[Baileys Sidecar] Session logged out. Auth directory cleared.');
          isConnecting = false;
        } else {
          connectionStatus = 'pending';
          isConnecting = false;

          let hasSavedAuth = false;
          try {
            hasSavedAuth = fs.existsSync(AUTH_DIR) && fs.readdirSync(AUTH_DIR).length > 0;
          } catch (e) {}

          // Exponential backoff reconnect delay (min 1.5s, max 8s)
          reconnectAttempts++;
          const delay = Math.min(1500 * Math.pow(1.3, Math.min(reconnectAttempts, 6)), 8000);
          console.log(`[Baileys Sidecar] Auto-reconnecting in ${(delay / 1000).toFixed(1)}s (attempt #${reconnectAttempts}, hasAuth=${hasSavedAuth})...`);

          reconnectTimer = setTimeout(() => {
            connectToWhatsApp();
          }, delay);
        }
      }

      if (connection === 'open') {
        connectionStatus = 'connected';
        qrCode = '';
        reconnectAttempts = 0;
        const userJid = sock?.user?.id || '';
        linkedNumber = userJid.split(':')[0] || userJid.split('@')[0] || userJid;
        console.log(`[Baileys Sidecar] ✓ WhatsApp Connected! user=${userJid}, linkedNumber=${linkedNumber}`);
        isConnecting = false;
      }
    });

    isConnecting = false;
  } catch (err) {
    console.error('[Baileys Sidecar] Fatal socket initialization error:', err?.message || err);
    connectionStatus = 'disconnected';
    isConnecting = false;

    // Retry after 4s
    reconnectTimer = setTimeout(() => {
      connectToWhatsApp();
    }, 4000);
  }
}

// ─── REST Endpoints ──────────────────────────────────────────────────────────

app.get('/status', (req, res) => {
  res.json({
    status: connectionStatus,
    linkedNumber,
    qr: qrCode
  });
});

// GET /qr — returns current QR string
app.get('/qr', (req, res) => {
  res.json({
    status: connectionStatus,
    qr: qrCode,
    linkedNumber
  });
});

// POST /connect — initiates or verifies connection
app.post('/connect', async (req, res) => {
  if (connectionStatus === 'connected') {
    return res.json({ status: 'connected', linkedNumber });
  }

  if (connectionStatus === 'pending' && sock && qrCode) {
    return res.json({
      status: 'pending',
      qr: qrCode,
      message: 'Connection already in progress.'
    });
  }

  qrCode = '';
  connectionStatus = 'pending';
  reconnectAttempts = 0;
  isConnecting = false;
  cleanupExistingSocket();

  connectToWhatsApp();

  res.json({
    status: 'pending',
    message: 'Baileys WhatsApp connection initiated.'
  });
});

// POST /disconnect — logout and clear session
app.post('/disconnect', async (req, res) => {
  try {
    if (sock) {
      await sock.logout().catch(() => {});
    }
  } catch (e) {
    console.error('[Baileys Sidecar] Error during logout:', e?.message || e);
  }

  cleanupExistingSocket();

  connectionStatus = 'disconnected';
  qrCode = '';
  linkedNumber = '';
  reconnectAttempts = 0;
  clearAuthDir();

  res.json({ status: 'disconnected', message: 'WhatsApp session logged out and auth cleared' });
});

// POST /send — send a WhatsApp message
app.post('/send', async (req, res) => {
  const { recipient, message, target } = req.body;
  const num = recipient || target;

  if (!num || !message) {
    return res.status(400).json({ error: 'recipient and message are required' });
  }

  // Normalize recipient / JID
  const rawNum = String(num).trim();
  let jid = '';
  const isGroup = rawNum.endsWith('@g.us') || rawNum.includes('@g.us');

  if (rawNum.endsWith('@g.us')) {
    jid = rawNum;
  } else if (rawNum.endsWith('@s.whatsapp.net')) {
    jid = rawNum;
  } else if (rawNum.includes('@g.us')) {
    const cleanGid = rawNum.split('@')[0].replace(/[^0-9-]/g, '');
    jid = `${cleanGid}@g.us`;
  } else {
    // Individual phone number
    let formattedNum = rawNum.replace(/[^0-9]/g, '');
    if (formattedNum.startsWith('0')) {
      formattedNum = '62' + formattedNum.slice(1);
    }
    jid = `${formattedNum}@s.whatsapp.net`;
  }

  // Wait for connection to be ready (up to 8s)
  const maxWait = 8;
  for (let attempt = 1; attempt <= maxWait; attempt++) {
    if (connectionStatus === 'connected' && sock) break;

    let hasSavedAuth = false;
    try {
      hasSavedAuth = fs.existsSync(AUTH_DIR) && fs.readdirSync(AUTH_DIR).length > 0;
    } catch (e) {}

    if (connectionStatus === 'disconnected' && hasSavedAuth && !isConnecting) {
      console.log(`[Baileys Sidecar] Session is disconnected with saved auth. Triggering auto-reconnect...`);
      connectToWhatsApp();
    }

    if (attempt < maxWait) {
      await new Promise((resolve) => setTimeout(resolve, 1000));
      continue;
    }
    return res.status(503).json({
      error: `WhatsApp session not connected (status: ${connectionStatus}). Please verify QR connection in Settings.`
    });
  }

  // Send message directly to target JID
  try {
    await sock.sendMessage(jid, { text: message });
    console.log(`[Baileys Sidecar] ✓ Message delivered to ${jid}`);
    return res.json({ success: true, recipient: jid });
  } catch (err) {
    const errMsg = err?.message || 'unknown error';
    const isForbidden = errMsg.toLowerCase().includes('forbidden') || errMsg.toLowerCase().includes('not-authorized');

    if (isForbidden && isGroup) {
      let groupName = jid;
      try {
        const meta = await sock.groupMetadata(jid);
        if (meta && meta.subject) {
          groupName = `"${meta.subject}" (${jid})`;
        }
      } catch (e) {}

      const linkedNum = linkedNumber || sock?.user?.id?.split(':')[0]?.split('@')[0] || '?';
      const friendlyErr = `Failed to send to group ${groupName}: "forbidden". Ensure +${linkedNum} is added to the group and has message permissions.`;
      console.error(`[Baileys Sidecar] ✗ ${friendlyErr}`);
      return res.status(403).json({ error: friendlyErr });
    }

    console.error(`[Baileys Sidecar] ✗ Failed to send to ${jid}: ${errMsg}`);
    return res.status(500).json({ error: errMsg || 'Failed to send message' });
  }
});

// ─── Start Server ────────────────────────────────────────────────────────────
app.listen(PORT, () => {
  console.log(`\x1b[32m[SANOC Baileys Sidecar]\x1b[0m Listening on http://localhost:${PORT}`);
  console.log(`[Baileys Sidecar] Auth session path: ${AUTH_DIR}`);
  console.log(`[Baileys Sidecar] Internal token auth: ${getInternalToken() ? 'ENABLED' : 'DISABLED (dev mode)'}`);

  // Auto-start connection on boot
  connectToWhatsApp();
});
