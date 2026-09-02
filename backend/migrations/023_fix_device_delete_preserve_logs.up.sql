-- 023_fix_device_delete_preserve_logs.up.sql
-- Fix: device deletion should ONLY delete the device row, NOT cascade-delete logs/history.
-- Change FKs from ON DELETE CASCADE to ON DELETE SET NULL and make device_id nullable.
-- Related logs (incidents, metrics, status_log, ip_change_log) are preserved after device deletion.
-- They can be purged separately via dedicated maintenance endpoints when storage is full.

-- incidents.device_id: SET NULL (preserve incident even if device deleted)
ALTER TABLE incidents DROP CONSTRAINT IF EXISTS incidents_device_id_fkey;
ALTER TABLE incidents ALTER COLUMN device_id DROP NOT NULL;
ALTER TABLE incidents ADD CONSTRAINT incidents_device_id_fkey
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE SET NULL;

-- device_metrics.device_id: SET NULL
ALTER TABLE device_metrics DROP CONSTRAINT IF EXISTS device_metrics_device_id_fkey;
ALTER TABLE device_metrics ALTER COLUMN device_id DROP NOT NULL;
ALTER TABLE device_metrics ADD CONSTRAINT device_metrics_device_id_fkey
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_device_metrics_device_id ON device_metrics(device_id);

-- device_status_log.device_id: SET NULL
ALTER TABLE device_status_log DROP CONSTRAINT IF EXISTS device_status_log_device_id_fkey;
ALTER TABLE device_status_log ALTER COLUMN device_id DROP NOT NULL;
ALTER TABLE device_status_log ADD CONSTRAINT device_status_log_device_id_fkey
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_device_status_log_device_id ON device_status_log(device_id);

-- ip_change_log.device_id: SET NULL
ALTER TABLE ip_change_log DROP CONSTRAINT IF EXISTS ip_change_log_device_id_fkey;
ALTER TABLE ip_change_log ALTER COLUMN device_id DROP NOT NULL;
ALTER TABLE ip_change_log ADD CONSTRAINT ip_change_log_device_id_fkey
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_ip_change_log_device_id ON ip_change_log(device_id);

-- device_dependencies: table was dropped in 006 - only handle if it exists (legacy)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'device_dependencies') THEN
    CREATE INDEX IF NOT EXISTS idx_device_dependencies_parent ON device_dependencies(parent_id);
    CREATE INDEX IF NOT EXISTS idx_device_dependencies_child ON device_dependencies(child_id);
  END IF;
END $$;

-- notification_log and incident_events/incident_notes are linked via incident_id, not device_id
-- They are already ON DELETE SET NULL (notification_log) or CASCADE (incident_events) via incidents.
-- When incident is preserved (SET NULL device), its events remain intact.
