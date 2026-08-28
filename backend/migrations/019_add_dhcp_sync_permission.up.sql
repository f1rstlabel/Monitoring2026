-- 019_add_dhcp_sync_permission.up.sql
INSERT INTO role_permissions (role, feature_key, enabled) VALUES 
('pimpinan', 'settings.dhcp_sync', false),
('anggota', 'settings.dhcp_sync', false)
ON CONFLICT (role, feature_key) DO NOTHING;
