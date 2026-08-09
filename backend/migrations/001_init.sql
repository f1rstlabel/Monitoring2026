-- GovMonitor IT — Complete Schema Migration
-- Run with: sqlite3 govmonitor.db < migrations/001_init.sql

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- ─── Users & Roles ───────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS users (
    id          TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
    name        TEXT NOT NULL,
    email       TEXT NOT NULL UNIQUE,
    password    TEXT NOT NULL,           -- bcrypt hash
    role        TEXT NOT NULL DEFAULT 'anggota'
                CHECK (role IN ('superadmin', 'pimpinan', 'anggota')),
    status      TEXT NOT NULL DEFAULT 'Active'
                CHECK (status IN ('Active', 'Inactive')),
    avatar_url  TEXT,
    last_active TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Default superadmin (password: "admin" — change before production)
INSERT OR IGNORE INTO users (id, name, email, password, role, avatar_url, last_active) VALUES
('u-001', 'Budi Santoso', 'admin.noc@jabarprov.go.id',
 '$2a$12$placeholder_bcrypt_hash_admin', 'superadmin',
 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256',
 'Active Now'),
('u-002', 'Sari Dewi', 'sari.dewi@jabarprov.go.id',
 '$2a$12$placeholder_bcrypt_hash_pimpinan', 'pimpinan',
 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=256',
 '2h ago'),
('u-003', 'Rian Pratama', 'rian.pratama@jabarprov.go.id',
 '$2a$12$placeholder_bcrypt_hash_anggota', 'anggota',
 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=256',
 '30m ago');

-- ─── Devices ─────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS devices (
    id                  TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
    name                TEXT NOT NULL,
    type                TEXT NOT NULL
                        CHECK (type IN ('Access Point','Switch','Router','SmartPower','CCTV','NVR')),
    -- MAC is the stable identity key for DHCP devices — must be unique
    mac_address         TEXT NOT NULL UNIQUE,
    -- last_known_ip is updated by DHCP resolution; may change frequently
    last_known_ip       TEXT,
    addressing_mode     TEXT NOT NULL DEFAULT 'DHCP'
                        CHECK (addressing_mode IN ('Static', 'DHCP')),
    model               TEXT,           -- e.g. "U6 IW", "U6 Enterprise"
    firmware_status     TEXT,           -- "Up to date" from UniFi — NOT monitoring status
    status              TEXT NOT NULL DEFAULT 'UP'
                        CHECK (status IN ('UP', 'DOWN')),
    parent_device_id    TEXT REFERENCES devices(id) ON DELETE SET NULL,
    location            TEXT,
    rack                TEXT,
    failure_threshold   INTEGER NOT NULL DEFAULT 3,
    uptime_30d          REAL DEFAULT 100.0,
    down_count_7d       INTEGER DEFAULT 0,
    down_count_30d      INTEGER DEFAULT 0,
    checked_seconds_ago INTEGER DEFAULT 0,
    last_checked        TEXT,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for DHCP MAC resolution lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_mac ON devices (lower(mac_address));
CREATE INDEX IF NOT EXISTS idx_devices_type ON devices (type);
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices (status);

-- ─── Device Dependencies ─────────────────────────────────────────────────────
-- Maps child devices to their logical parent (for cascade incident grouping)

CREATE TABLE IF NOT EXISTS device_dependencies (
    parent_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    child_id  TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    PRIMARY KEY (parent_id, child_id)
);

-- ─── Incidents ───────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS incidents (
    id                      TEXT PRIMARY KEY DEFAULT ('INC-' || strftime('%Y', 'now') || '-' || substr(lower(hex(randomblob(4))), 1, 6)),
    device_id               TEXT NOT NULL REFERENCES devices(id),
    status                  TEXT NOT NULL DEFAULT 'ACTIVE'
                            CHECK (status IN ('ACTIVE', 'RESOLVED')),
    packet_loss             INTEGER DEFAULT 100,
    latency_ms              INTEGER DEFAULT 0,
    affected_devices_count  INTEGER DEFAULT 1,
    dependencies_count      INTEGER DEFAULT 0,
    started_at              DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at             DATETIME,
    created_at              DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_incidents_device ON incidents (device_id);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents (status);

-- ─── Incident Notes ───────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS incident_notes (
    id          TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
    incident_id TEXT NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    author_id   TEXT REFERENCES users(id),
    body        TEXT NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ─── Device Status Log ────────────────────────────────────────────────────────
-- Every UP/DOWN transition is recorded here for SLA reporting and flap detection.

CREATE TABLE IF NOT EXISTS device_status_log (
    id          TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
    device_id   TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    status      TEXT NOT NULL CHECK (status IN ('UP', 'DOWN')),
    incident_id TEXT REFERENCES incidents(id) ON DELETE SET NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_status_log_device ON device_status_log (device_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_status_log_status ON device_status_log (status, created_at DESC);

-- ─── IP Change Log ────────────────────────────────────────────────────────────
-- Audit trail for DHCP device IP rotations.

CREATE TABLE IF NOT EXISTS ip_change_log (
    id          TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
    device_id   TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    old_ip      TEXT NOT NULL,
    new_ip      TEXT NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ip_change_device ON ip_change_log (device_id, created_at DESC);

-- ─── Notification Log ─────────────────────────────────────────────────────────
-- Records every dispatch attempt (WhatsApp + Telegram) for the Incident Detail UI.

CREATE TABLE IF NOT EXISTS notification_log (
    id          TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
    incident_id TEXT REFERENCES incidents(id) ON DELETE SET NULL,
    channel     TEXT NOT NULL,          -- 'WhatsApp' | 'Telegram'
    recipient   TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'Pending'
                CHECK (status IN ('Pending', 'Sent', 'Delivered', 'Failed')),
    error_msg   TEXT,
    sent_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notif_log_incident ON notification_log (incident_id, sent_at DESC);
