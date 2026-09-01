-- Add dedicated indexes on all Foreign Keys to guarantee fast CASCADE and set-based deletes
CREATE INDEX IF NOT EXISTS idx_incident_notes_incident ON incident_notes (incident_id);
CREATE INDEX IF NOT EXISTS idx_device_metrics_device_id ON device_metrics (device_id);
CREATE INDEX IF NOT EXISTS idx_device_status_log_device_id ON device_status_log (device_id);
CREATE INDEX IF NOT EXISTS idx_device_dependencies_child ON device_dependencies (child_id);
CREATE INDEX IF NOT EXISTS idx_device_dependencies_parent ON device_dependencies (parent_id);
CREATE INDEX IF NOT EXISTS idx_ip_change_log_device_id ON ip_change_log (device_id);

-- Ensure all foreign keys have ON DELETE CASCADE
ALTER TABLE incidents DROP CONSTRAINT IF EXISTS incidents_device_id_fkey;
ALTER TABLE incidents ADD CONSTRAINT incidents_device_id_fkey FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE;

ALTER TABLE device_metrics DROP CONSTRAINT IF EXISTS device_metrics_device_id_fkey;
ALTER TABLE device_metrics ADD CONSTRAINT device_metrics_device_id_fkey FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE;

ALTER TABLE device_status_log DROP CONSTRAINT IF EXISTS device_status_log_device_id_fkey;
ALTER TABLE device_status_log ADD CONSTRAINT device_status_log_device_id_fkey FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE;

ALTER TABLE ip_change_log DROP CONSTRAINT IF EXISTS ip_change_log_device_id_fkey;
ALTER TABLE ip_change_log ADD CONSTRAINT ip_change_log_device_id_fkey FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE;
