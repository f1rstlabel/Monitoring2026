-- 007_locations_and_permissions.up.sql
-- Create locations and role_permissions tables, with migration for existing device locations

-- 1. Create locations table
CREATE TABLE IF NOT EXISTS locations (
    id          VARCHAR(64) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Add location_id column to devices table if not exists
ALTER TABLE devices ADD COLUMN IF NOT EXISTS location_id VARCHAR(64) REFERENCES locations(id) ON DELETE SET NULL;

-- 3. Populate locations table dynamically from distinct location strings in devices table
INSERT INTO locations (id, name, description)
SELECT 'loc-' || MD5(TRIM(location)), TRIM(location), 'Location from registered devices'
FROM devices
WHERE location IS NOT NULL AND TRIM(location) != ''
ON CONFLICT (name) DO NOTHING;

-- 4. Update devices.location_id by matching devices.location = locations.name
UPDATE devices d
SET location_id = l.id
FROM locations l
WHERE LOWER(TRIM(d.location)) = LOWER(TRIM(l.name)) AND (d.location_id IS NULL OR d.location_id = '');

-- 5. Delete unused dummy locations that are not referenced in devices
DELETE FROM locations
WHERE id NOT IN (SELECT DISTINCT location_id FROM devices WHERE location_id IS NOT NULL AND location_id != '')
  AND name NOT IN (SELECT DISTINCT location FROM devices WHERE location IS NOT NULL AND location != '');

-- 3. Create role_permissions table for fine-grained per-role access control
CREATE TABLE IF NOT EXISTS role_permissions (
    role        VARCHAR(50) NOT NULL,
    feature_key VARCHAR(100) NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role, feature_key)
);

-- Seed default permissions for pimpinan
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
ON CONFLICT (role, feature_key) DO NOTHING;

-- Seed default permissions for anggota
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
ON CONFLICT (role, feature_key) DO NOTHING;
