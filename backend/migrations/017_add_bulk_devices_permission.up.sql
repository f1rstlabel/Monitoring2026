-- 017_add_bulk_devices_permission.up.sql
-- Add devices.bulk permission to role_permissions

INSERT INTO role_permissions (role, feature_key, enabled) VALUES
('pimpinan', 'devices.bulk', false),
('anggota', 'devices.bulk', false)
ON CONFLICT (role, feature_key) DO NOTHING;
