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
let isRestarting = false;

// ─── Baileys Connection ──────────────────────────────────────────────────────
async function connectToWhatsApp() {
  if (isRestarting) return;
  isRestarting = true;

  try {
    if (!fs.existsSync(AUTH_DIR)) {
      fs.mkdirSync(AUTH_DIR, { recursive: true });
    }

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
      printQRInTerminal: false, // We handle QR ourselves
      logger: pino({ level: 'silent' }),
      browser: ['GovMonitor IT', 'Chrome', '126.0.0'],
      generateHighQualityLinkPreview: false,
      syncFullHistory: false
    });

    sock.ev.on('creds.update', saveCreds);

    sock.ev.on('connection.update', (update) => {
      const { connection, lastDisconnect, qr } = update;

      if (qr) {
        // This is the ONLY valid source of a QR code — exact unmodified string from Baileys
        qrCode = qr;
        connectionStatus = 'pending';
        console.log(`[Baileys Sidecar] New QR code generated (len=${qr.length}). Waiting for phone scan...`);
        // Log first 30 chars so Go backend can compare byte-for-byte
        console.log(`[Baileys Sidecar] QR prefix: ${qr.substring(0, 30)}...`);
      }

      if (connection === 'close') {
        const statusCode = lastDisconnect?.error?.output?.statusCode;
        const isLoggedOut = statusCode === DisconnectReason.loggedOut;
        const isRestartRequired = statusCode === DisconnectReason.restartRequired; // 405

        console.log(`[Baileys Sidecar] Connection closed. statusCode=${statusCode} loggedOut=${isLoggedOut} restartRequired=${isRestartRequired}`);

        if (isLoggedOut) {
          // User explicitly logged out — clear auth and stay disconnected
          connectionStatus = 'disconnected';
          qrCode = '';
          linkedNumber = '';
          clearAuthDir();
          console.log('[Baileys Sidecar] Logged out — auth cleared. Use /connect to start fresh session.');
        } else {
          // restartRequired (405 after QR scan) or any other close — reconnect automatically
          // This is EXPECTED after a QR scan: WhatsApp closes then immediately reopens with saved creds
          connectionStatus = 'pending';
          qrCode = '';
          console.log('[Baileys Sidecar] Auto-reconnecting (restartRequired or transient close)...');
          isRestarting = false;
          setTimeout(connectToWhatsApp, 1500);
        }
      }

      if (connection === 'open') {
        // THIS is the ONLY valid moment to mark status as "connected"
        connectionStatus = 'connected';
        qrCode = '';
        const userJid = sock?.user?.id || '';
        linkedNumber = userJid.split(':')[0] || userJid.split('@')[0] || userJid;
        console.log(`[Baileys Sidecar] ✓ WhatsApp Connected! connection=open, user=${sock?.user?.id}, linkedNumber=${linkedNumber}`);
        isRestarting = false;
      }
    });

    isRestarting = false;
  } catch (err) {
    console.error('[Baileys Sidecar] Fatal connection error:', err);
    connectionStatus = 'disconnected';
    isRestarting = false;
  }
}

function clearAuthDir() {
  if (fs.existsSync(AUTH_DIR)) {
    try {
      fs.rmSync(AUTH_DIR, { recursive: true, force: true });
      console.log('[Baileys Sidecar] Auth directory cleared:', AUTH_DIR);
    } catch (err) {
      console.error('[Baileys Sidecar] Failed to clear auth dir:', err);
    }
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

// GET /qr — returns current QR string (exact Baileys string, unmodified)
app.get('/qr', (req, res) => {
  res.json({
    status: connectionStatus,
    qr: qrCode,  // exact string from Baileys connection.update — never re-encoded here
    linkedNumber
  });
});

// POST /connect — initiates or verifies connection
app.post('/connect', async (req, res) => {
  if (connectionStatus === 'connected') {
    return res.json({ status: 'connected', linkedNumber });
  }

  // Reset and start fresh session
  qrCode = '';
  connectionStatus = 'pending';
  isRestarting = false;

  // If already has saved auth, it will reconnect without showing QR
  connectToWhatsApp();

  res.json({
    status: 'pending',
    message: 'Baileys WhatsApp connection initiated. Poll /qr for QR code or /status for updates.'
  });
});

// POST /disconnect — logout and clear session
app.post('/disconnect', async (req, res) => {
  try {
    if (sock) {
      await sock.logout().catch(() => {});
      sock = null;
    }
  } catch (e) {
    console.error('[Baileys Sidecar] Error during logout:', e.message);
  }

  connectionStatus = 'disconnected';
  qrCode = '';
  linkedNumber = '';
  isRestarting = false;
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

  // Wait for connection to be ready
  let lastErr = null;
  const maxWait = 10;
  for (let attempt = 1; attempt <= maxWait; attempt++) {
    if (connectionStatus === 'connected' && sock) break;

    let hasSavedAuth = false;
    try {
      hasSavedAuth = fs.existsSync(AUTH_DIR) && fs.readdirSync(AUTH_DIR).length > 0;
    } catch (e) {}

    if (connectionStatus === 'disconnected' && hasSavedAuth && !isRestarting) {
      console.log(`[Baileys Sidecar] Saved session exists but socket is disconnected. Triggering reconnect...`);
      connectToWhatsApp();
    }

    if (attempt < maxWait) {
      console.log(`[Baileys Sidecar] Session status '${connectionStatus}' (attempt ${attempt}/${maxWait}). Waiting 1s for socket ready...`);
      await new Promise(resolve => setTimeout(resolve, 1000));
      continue;
    }
    return res.status(503).json({
      error: `WhatsApp session not connected (status: ${connectionStatus}). Please scan QR code in Settings → Integrations.`
    });
  }

  // Send message directly to target JID
  try {
    await sock.sendMessage(jid, { text: message });
    console.log(`[Baileys Sidecar] ✓ Message delivered to ${jid}`);
    return res.json({ success: true, recipient: jid });
  } catch (err) {
    lastErr = err;
    const errMsg = err.message || 'unknown error';
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
      const friendlyErr = `Failed to send to group ${groupName}: "forbidden". Ensure +${linkedNum} is added to the group and has message permissions (make admin if group is admin-only).`;
      console.error(`[Baileys Sidecar] ✗ ${friendlyErr}`);
      return res.status(403).json({ error: friendlyErr });
    }

    console.error(`[Baileys Sidecar] ✗ Failed to send to ${jid}: ${errMsg}`);
    return res.status(500).json({ error: errMsg || 'Failed to send message' });
  }
});

// ─── Start Server ────────────────────────────────────────────────────────────
app.listen(PORT, () => {
  console.log(`\x1b[32m[GovMonitor IT Baileys Sidecar]\x1b[0m Listening on http://localhost:${PORT}`);
  console.log(`[Baileys Sidecar] Auth session path: ${AUTH_DIR}`);
  console.log(`[Baileys Sidecar] Internal token auth: ${getInternalToken() ? 'ENABLED' : 'DISABLED (dev mode)'}`);

  // Auto-start connection on boot (will reconnect with saved creds if auth exists)
  connectToWhatsApp();
});
