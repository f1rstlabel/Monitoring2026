CREATE INDEX IF NOT EXISTS idx_device_status_log_incident ON device_status_log (incident_id);
CREATE INDEX IF NOT EXISTS idx_devices_location ON devices (location_id);
CREATE INDEX IF NOT EXISTS idx_devices_created_by ON devices (created_by_user_id);
