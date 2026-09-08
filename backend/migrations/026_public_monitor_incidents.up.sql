-- Incident lifecycle and delivery history for public monitors.
-- Kept separate from incidents/notification_log so network-device history is untouched.
CREATE TABLE IF NOT EXISTS public_monitor_incidents (
    id                    VARCHAR(64) PRIMARY KEY,
    monitor_id            VARCHAR(64) NOT NULL REFERENCES public_monitors(id) ON DELETE CASCADE,
    status                VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                          CHECK (status IN ('ACTIVE', 'RESOLVED')),
    started_at            TIMESTAMP WITH TIME ZONE NOT NULL,
    resolved_at           TIMESTAMP WITH TIME ZONE,
    duration_seconds      BIGINT NOT NULL DEFAULT 0,
    initial_status_code   INTEGER NOT NULL DEFAULT 0,
    final_status_code     INTEGER NOT NULL DEFAULT 0,
    first_error           TEXT NOT NULL DEFAULT '',
    last_error            TEXT NOT NULL DEFAULT '',
    created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_public_monitor_active_incident
    ON public_monitor_incidents (monitor_id) WHERE status = 'ACTIVE';
CREATE INDEX IF NOT EXISTS idx_public_monitor_incidents_monitor_time
    ON public_monitor_incidents (monitor_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_public_monitor_incidents_status
    ON public_monitor_incidents (status, started_at DESC);

CREATE TABLE IF NOT EXISTS public_monitor_incident_events (
    id           VARCHAR(64) PRIMARY KEY,
    incident_id  VARCHAR(64) NOT NULL REFERENCES public_monitor_incidents(id) ON DELETE CASCADE,
    event_type   VARCHAR(40) NOT NULL,
    detail       TEXT NOT NULL DEFAULT '',
    occurred_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_public_monitor_incident_events_incident_time
    ON public_monitor_incident_events (incident_id, occurred_at DESC);

CREATE TABLE IF NOT EXISTS public_monitor_notification_logs (
    id           VARCHAR(64) PRIMARY KEY,
    incident_id  VARCHAR(64) NOT NULL REFERENCES public_monitor_incidents(id) ON DELETE CASCADE,
    channel      VARCHAR(50) NOT NULL,
    recipient    VARCHAR(255) NOT NULL DEFAULT '',
    status       VARCHAR(50) NOT NULL DEFAULT 'Pending',
    error_msg    TEXT NOT NULL DEFAULT '',
    sent_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_public_monitor_notification_logs_incident_time
    ON public_monitor_notification_logs (incident_id, sent_at DESC);
