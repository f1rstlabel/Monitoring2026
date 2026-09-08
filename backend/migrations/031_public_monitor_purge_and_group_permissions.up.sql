-- Configuration removal is an auditable tombstone. Do not hard-delete the
-- monitor row because checks, events, incidents, and reports reference it.
ALTER TABLE public_monitors
    ADD COLUMN IF NOT EXISTS purged_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS purged_by_user_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS purge_reason TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_public_monitors_purged_at
    ON public_monitors (purged_at, updated_at DESC);

-- Keep the legacy groups.manage permission for compatibility, but use the
-- granular permissions below for new routes and the settings matrix.
INSERT INTO role_permissions (role, feature_key, enabled)
VALUES
    ('pimpinan', 'public_monitoring.delete', false),
    ('anggota', 'public_monitoring.delete', false),
    ('pimpinan', 'public_monitoring.groups.create', false),
    ('pimpinan', 'public_monitoring.groups.edit', false),
    ('pimpinan', 'public_monitoring.groups.delete', false),
    ('anggota', 'public_monitoring.groups.create', true),
    ('anggota', 'public_monitoring.groups.edit', true),
    ('anggota', 'public_monitoring.groups.delete', false)
ON CONFLICT (role, feature_key) DO UPDATE
SET enabled = EXCLUDED.enabled, updated_at = CURRENT_TIMESTAMP;
