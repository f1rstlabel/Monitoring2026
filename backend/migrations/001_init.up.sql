-- GovMonitor IT — PostgreSQL Schema Migration

CREATE TABLE IF NOT EXISTS users (
    id          VARCHAR(64) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    email       VARCHAR(255) NOT NULL UNIQUE,
    password    VARCHAR(255) NOT NULL, -- bcrypt hash
    role        VARCHAR(50) NOT NULL DEFAULT 'anggota'
                CHECK (role IN ('superadmin', 'pimpinan', 'anggota')),
    status      VARCHAR(50) NOT NULL DEFAULT 'Active'
                CHECK (status IN ('Active', 'Inactive')),
    avatar_url  TEXT,
    last_active VARCHAR(100),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed Default Superadmin user (password: admin123)
-- bcrypt hash for "admin123": $2a$10$8R3ySnqo4Ry9xsCCbu1CZu69ccq4t8UdrYMQcextliQh57qCo6bhC
INSERT INTO users (id, name, email, password, role, avatar_url, last_active) VALUES
('u-001', 'Budi Santoso (NOC Admin)', 'admin.noc@jabarprov.go.id', '$2a$10$8R3ySnqo4Ry9xsCCbu1CZu69ccq4t8UdrYMQcextliQh57qCo6bhC', 'superadmin', 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256', 'Active Now'),
('u-002', 'Sari Dewi (Diskominfo)', 'sari.dewi@jabarprov.go.id', '$2a$10$8R3ySnqo4Ry9xsCCbu1CZu69ccq4t8UdrYMQcextliQh57qCo6bhC', 'pimpinan', 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=256', '2h ago'),
('u-003', 'Rian Pratama (Operator)', 'rian.pratama@jabarprov.go.id', '$2a$10$8R3ySnqo4Ry9xsCCbu1CZu69ccq4t8UdrYMQcextliQh57qCo6bhC', 'anggota', 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=256', '30m ago')
ON CONFLICT (email) DO NOTHING;

CREATE TABLE IF NOT EXISTS devices (
    id                  VARCHAR(64) PRIMARY KEY,
    name                VARCHAR(255) NOT NULL,
    type                VARCHAR(50) NOT NULL
                        CHECK (type IN ('Access Point','Switch','Router','SmartPower','CCTV','NVR')),
    mac_address         VARCHAR(50) NOT NULL UNIQUE,
    last_known_ip       VARCHAR(50),
    addressing_mode     VARCHAR(20) NOT NULL DEFAULT 'DHCP'
                        CHECK (addressing_mode IN ('Static', 'DHCP')),
    model               VARCHAR(100),
    firmware_status     VARCHAR(100),
    status              VARCHAR(20) NOT NULL DEFAULT 'UP'
                        CHECK (status IN ('UP', 'DOWN')),
    parent_device_id    VARCHAR(64) REFERENCES devices(id) ON DELETE SET NULL,
    location            VARCHAR(255),
    rack                VARCHAR(100),
    failure_threshold   INTEGER NOT NULL DEFAULT 3,
    uptime_30d          DOUBLE PRECISION DEFAULT 100.0,
    down_count_7d       INTEGER DEFAULT 0,
    down_count_30d      INTEGER DEFAULT 0,
    checked_seconds_ago INTEGER DEFAULT 0,
    last_checked        VARCHAR(100),
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_devices_mac ON devices (LOWER(mac_address));
CREATE INDEX IF NOT EXISTS idx_devices_type ON devices (type);
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices (status);

-- Seed initial inventory
INSERT INTO devices (id, name, type, mac_address, last_known_ip, status, addressing_mode, location, rack, failure_threshold, uptime_30d) VALUES
('dev-1', 'Router Core Gedung Sate Utama', 'Router', '00:1a:2b:3c:4d:5e', '10.20.0.1', 'UP', 'Static', 'Server Room Lt. 1', 'Rack A-01', 1, 99.9),
('dev-2', 'Distribution Switch Core-01', 'Switch', '00:1a:2b:3c:4d:5f', '10.20.0.2', 'UP', 'Static', 'Rack A1 - Main Server Room', 'Rack A-02', 2, 99.8),
('dev-3', 'Distribution Switch Core-02', 'Switch', '00:1a:2b:3c:4d:60', '10.20.0.3', 'DOWN', 'Static', 'Rack A2 - Main Server Room', 'Rack A-03', 2, 94.2),
('dev-4', 'AP Biro Umum Lt 1 North', 'Access Point', '00:1a:2b:3c:4d:61', '10.20.1.12', 'UP', 'Static', 'Biro Umum - Hallway North', 'Wall Mount', 3, 99.7),
('dev-5', 'AP Biro Umum Lt 2 Conference', 'Access Point', 'f4:8e:38:1a:0b:44', '10.20.1.18', 'DOWN', 'DHCP', 'Ruang Rapat Sekretariat Lt 2', 'Rack B-04', 3, 99.8)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS device_dependencies (
    parent_id VARCHAR(64) NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    child_id  VARCHAR(64) NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    PRIMARY KEY (parent_id, child_id)
);

CREATE TABLE IF NOT EXISTS incidents (
    id                      VARCHAR(64) PRIMARY KEY,
    device_id               VARCHAR(64) NOT NULL REFERENCES devices(id),
    status                  VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                            CHECK (status IN ('ACTIVE', 'RESOLVED')),
    packet_loss             INTEGER DEFAULT 100,
    latency_ms              INTEGER DEFAULT 0,
    affected_devices_count  INTEGER DEFAULT 1,
    dependencies_count      INTEGER DEFAULT 0,
    started_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at             TIMESTAMP WITH TIME ZONE,
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_incidents_device ON incidents (device_id);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents (status);

-- Seed initial incident
INSERT INTO incidents (id, device_id, status, packet_loss, latency_ms, affected_devices_count, dependencies_count) VALUES
('INC-2026-089', 'dev-5', 'ACTIVE', 100, 0, 14, 3)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS incident_notes (
    id          VARCHAR(64) PRIMARY KEY,
    incident_id VARCHAR(64) NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    author_id   VARCHAR(64) REFERENCES users(id),
    body        TEXT NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS device_status_log (
    id          VARCHAR(64) PRIMARY KEY,
    device_id   VARCHAR(64) NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    status      VARCHAR(20) NOT NULL CHECK (status IN ('UP', 'DOWN')),
    incident_id VARCHAR(64) REFERENCES incidents(id) ON DELETE SET NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_status_log_device ON device_status_log (device_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_status_log_status ON device_status_log (status, created_at DESC);

CREATE TABLE IF NOT EXISTS ip_change_log (
    id          VARCHAR(64) PRIMARY KEY,
    device_id   VARCHAR(64) NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    old_ip      VARCHAR(50) NOT NULL,
    new_ip      VARCHAR(50) NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ip_change_device ON ip_change_log (device_id, created_at DESC);

CREATE TABLE IF NOT EXISTS notification_log (
    id          VARCHAR(64) PRIMARY KEY,
    incident_id VARCHAR(64) REFERENCES incidents(id) ON DELETE SET NULL,
    channel     VARCHAR(50) NOT NULL,
    recipient   VARCHAR(255) NOT NULL,
    status      VARCHAR(50) NOT NULL DEFAULT 'Pending'
                CHECK (status IN ('Pending', 'Sent', 'Delivered', 'Failed')),
    error_msg   TEXT,
    sent_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notif_log_incident ON notification_log (incident_id, sent_at DESC);

CREATE TABLE IF NOT EXISTS integration_settings (
    key        VARCHAR(64) PRIMARY KEY,
    value      JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO integration_settings (key, value) VALUES
('telegram',        '{"bot_token": "", "chat_id": ""}'::jsonb),
('whatsapp',        '{"status": "disconnected", "linked_number": ""}'::jsonb),
('system_settings', '{"rateLimitMaxMsgPerMin": 60, "polling": {"intervalSeconds": 15, "concurrencyBatchSize": 50}, "thresholds": [{"type": "Access Point", "consecutiveFailures": 3}, {"type": "Switch", "consecutiveFailures": 2}, {"type": "Router", "consecutiveFailures": 2}, {"type": "SmartPower", "consecutiveFailures": 4}, {"type": "CCTV", "consecutiveFailures": 5}, {"type": "NVR", "consecutiveFailures": 3}]}'::jsonb)
ON CONFLICT (key) DO NOTHING;
