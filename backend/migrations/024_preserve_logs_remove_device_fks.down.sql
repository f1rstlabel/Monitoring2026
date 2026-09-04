-- 024_preserve_logs_remove_device_fks.down.sql
-- Restore SET NULL foreign keys if rolled back

ALTER TABLE incidents DROP CONSTRAINT IF EXISTS incidents_device_id_fkey;
ALTER TABLE incidents ADD CONSTRAINT incidents_device_id_fkey
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE SET NULL;

ALTER TABLE device_metrics DROP CONSTRAINT IF EXISTS device_metrics_device_id_fkey;
ALTER TABLE device_metrics ADD CONSTRAINT device_metrics_device_id_fkey
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE SET NULL;

ALTER TABLE device_status_log DROP CONSTRAINT IF EXISTS device_status_log_device_id_fkey;
ALTER TABLE device_status_log ADD CONSTRAINT device_status_log_device_id_fkey
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE SET NULL;

ALTER TABLE ip_change_log DROP CONSTRAINT IF EXISTS ip_change_log_device_id_fkey;
ALTER TABLE ip_change_log ADD CONSTRAINT ip_change_log_device_id_fkey
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE SET NULL;
