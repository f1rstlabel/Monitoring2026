-- Public HTTP/HTTPS monitoring is intentionally separate from network devices.
CREATE TABLE IF NOT EXISTS public_monitors (
    id                   VARCHAR(64) PRIMARY KEY,
    name                 VARCHAR(255) NOT NULL,
    monitor_type         VARCHAR(20) NOT NULL DEFAULT 'http'
                         CHECK (monitor_type IN ('http')),
    target_url           TEXT NOT NULL,
    group_name           VARCHAR(120) NOT NULL DEFAULT 'Ungrouped',
    interval_seconds     INTEGER NOT NULL DEFAULT 60
                         CHECK (interval_seconds BETWEEN 15 AND 86400),
    timeout_seconds      INTEGER NOT NULL DEFAULT 10
                         CHECK (timeout_seconds BETWEEN 1 AND 120),
    retry_count          INTEGER NOT NULL DEFAULT 0
                         CHECK (retry_count BETWEEN 0 AND 5),
    enabled              BOOLEAN NOT NULL DEFAULT true,
    notify_on_failure    BOOLEAN NOT NULL DEFAULT false,
    status               VARCHAR(20) NOT NULL DEFAULT 'PAUSED'
                         CHECK (status IN ('UP', 'DOWN', 'PAUSED')),
    uptime_24h           DOUBLE PRECISION NOT NULL DEFAULT 100.0,
    uptime_30d           DOUBLE PRECISION NOT NULL DEFAULT 100.0,
    last_status_code     INTEGER NOT NULL DEFAULT 0,
    last_latency_ms      INTEGER NOT NULL DEFAULT 0,
    last_error           TEXT NOT NULL DEFAULT '',
    consecutive_failures INTEGER NOT NULL DEFAULT 0,
    last_checked         TIMESTAMP WITH TIME ZONE,
    created_by_user_id   VARCHAR(64),
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_public_monitors_status ON public_monitors (status);
CREATE INDEX IF NOT EXISTS idx_public_monitors_group ON public_monitors (group_name);
CREATE INDEX IF NOT EXISTS idx_public_monitors_enabled ON public_monitors (enabled, last_checked);

CREATE TABLE IF NOT EXISTS public_monitor_checks (
    id             VARCHAR(64) PRIMARY KEY,
    monitor_id     VARCHAR(64) NOT NULL REFERENCES public_monitors(id) ON DELETE CASCADE,
    checked_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status         VARCHAR(20) NOT NULL CHECK (status IN ('UP', 'DOWN')),
    status_code    INTEGER NOT NULL DEFAULT 0,
    latency_ms     INTEGER NOT NULL DEFAULT 0,
    error_message  TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_public_monitor_checks_monitor_time
    ON public_monitor_checks (monitor_id, checked_at DESC);

CREATE TABLE IF NOT EXISTS public_monitor_events (
    id          VARCHAR(64) PRIMARY KEY,
    monitor_id  VARCHAR(64) NOT NULL REFERENCES public_monitors(id) ON DELETE CASCADE,
    event_type  VARCHAR(40) NOT NULL,
    status      VARCHAR(20) NOT NULL,
    message     TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_public_monitor_events_monitor_time
    ON public_monitor_events (monitor_id, occurred_at DESC);

-- View access is enabled by default for the two non-admin roles.
INSERT INTO role_permissions (role, feature_key, enabled) VALUES
('pimpinan', 'public_monitoring.view', true),
('pimpinan', 'public_monitoring.create', false),
('pimpinan', 'public_monitoring.edit', false),
('pimpinan', 'public_monitoring.delete', false),
('anggota', 'public_monitoring.view', true),
('anggota', 'public_monitoring.create', true),
('anggota', 'public_monitoring.edit', true),
('anggota', 'public_monitoring.delete', false)
ON CONFLICT (role, feature_key) DO NOTHING;
