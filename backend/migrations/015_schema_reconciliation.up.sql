-- 015_schema_reconciliation.up.sql
-- Reconciliation of database schema with revised audit ERD

-- 1. Add created_by_user_id column to devices table if not present
ALTER TABLE devices ADD COLUMN IF NOT EXISTS created_by_user_id VARCHAR(64) REFERENCES users(id) ON DELETE SET NULL;

-- 2. Ensure incident_events table matches exact audit ERD structure
CREATE TABLE IF NOT EXISTS incident_events (
    event_id VARCHAR(64) PRIMARY KEY,
    incident_id VARCHAR(64) NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    channel VARCHAR(20),
    detail TEXT,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Rename column id -> event_id if existing table used 'id'
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'incident_events' AND column_name = 'id'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'incident_events' AND column_name = 'event_id'
    ) THEN
        ALTER TABLE incident_events RENAME COLUMN id TO event_id;
    END IF;
END $$;

-- Add index on incident_id for fast timeline lookups
CREATE INDEX IF NOT EXISTS idx_incident_events_incident ON incident_events(incident_id);

-- 3. Ensure integration_settings generic key-value JSONB table & default seed keys exist
CREATE TABLE IF NOT EXISTS integration_settings (
    key VARCHAR(64) PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO integration_settings (key, value) VALUES
('telegram', '{"bot_token": "", "chat_id": ""}'::jsonb),
('whatsapp', '{"status": "disconnected", "linked_number": ""}'::jsonb),
('system_settings', '{"rateLimitMaxMsgPerMin": 60, "polling": {"intervalSeconds": 15, "concurrencyBatchSize": 50, "debounceSeconds": 30}, "thresholds": [{"type": "Access Point", "consecutiveFailures": 3}, {"type": "Switch", "consecutiveFailures": 2}, {"type": "Router", "consecutiveFailures": 2}, {"type": "SmartPower", "consecutiveFailures": 4}, {"type": "CCTV", "consecutiveFailures": 5}, {"type": "NVR", "consecutiveFailures": 3}]}'::jsonb)
ON CONFLICT (key) DO NOTHING;

-- 4. Clean up removed incidents.resolve permission from role_permissions matrix
DELETE FROM role_permissions WHERE feature_key = 'incidents.resolve';
