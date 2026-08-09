-- 012_seed_role_permissions.up.sql
-- Ensure default permissions for pimpinan and anggota

-- Delete superadmin rows from role_permissions if any, since admin implicitly passes all permission checks
DELETE FROM role_permissions WHERE role = 'superadmin' OR role = 'admin';

-- Default permissions for pimpinan (read-only across most areas, full reports access)
INSERT INTO role_permissions (role, feature_key, enabled) VALUES
('pimpinan', 'devices.view', true),
('pimpinan', 'devices.create', false),
('pimpinan', 'devices.edit', false),
('pimpinan', 'devices.delete', false),
('pimpinan', 'devices.import', false),
('pimpinan', 'incidents.view', true),
('pimpinan', 'incidents.resolve', false),
('pimpinan', 'reports.view', true),
('pimpinan', 'reports.export', true),
('pimpinan', 'settings.notifications', false),
('pimpinan', 'settings.polling', false),
('pimpinan', 'settings.thresholds', false),
('pimpinan', 'settings.users', false),
('pimpinan', 'settings.permissions', false)
ON CONFLICT (role, feature_key) DO UPDATE SET enabled = EXCLUDED.enabled;

-- Default permissions for anggota (device/incident management, read-only/no settings/users access)
INSERT INTO role_permissions (role, feature_key, enabled) VALUES
('anggota', 'devices.view', true),
('anggota', 'devices.create', true),
('anggota', 'devices.edit', true),
('anggota', 'devices.delete', false),
('anggota', 'devices.import', false),
('anggota', 'incidents.view', true),
('anggota', 'incidents.resolve', true),
('anggota', 'reports.view', true),
('anggota', 'reports.export', true),
('anggota', 'settings.notifications', false),
('anggota', 'settings.polling', false),
('anggota', 'settings.thresholds', false),
('anggota', 'settings.users', false),
('anggota', 'settings.permissions', false)
ON CONFLICT (role, feature_key) DO UPDATE SET enabled = EXCLUDED.enabled;
