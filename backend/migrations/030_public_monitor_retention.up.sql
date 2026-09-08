-- Retain public-monitor records for audit instead of physically deleting them.
ALTER TABLE public_monitors
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS deleted_by_user_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS deletion_reason TEXT NOT NULL DEFAULT '';

ALTER TABLE public_monitor_incidents
    ADD COLUMN IF NOT EXISTS resolution_reason VARCHAR(80) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_public_monitors_deleted_at
    ON public_monitors (deleted_at, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_public_monitor_incidents_resolution_reason
    ON public_monitor_incidents (resolution_reason, resolved_at DESC);
