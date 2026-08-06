-- 1. Drop parent_device_id column from devices
ALTER TABLE devices DROP COLUMN IF EXISTS parent_device_id;

-- 2. Delete all legacy mock inventory
DELETE FROM incidents WHERE device_id NOT IN ('dev-local', 'dev-cloudflare', 'dev-google');
DELETE FROM device_status_log WHERE device_id NOT IN ('dev-local', 'dev-cloudflare', 'dev-google');
DELETE FROM ip_change_log WHERE device_id NOT IN ('dev-local', 'dev-cloudflare', 'dev-google');
DELETE FROM notification_log WHERE incident_id NOT IN (SELECT id FROM incidents);
DELETE FROM devices WHERE id NOT IN ('dev-local', 'dev-cloudflare', 'dev-google');

-- 3. Upsert test fixtures
INSERT INTO devices (id, name, type, mac_address, last_known_ip, status, addressing_mode, location, rack, failure_threshold, uptime_30d) VALUES
('dev-local', 'Test - Localhost', 'Router', '00:00:00:00:00:01', '127.0.0.1', 'UP', 'Static', 'Local Host Server', 'Rack L-01', 2, 100.0),
('dev-cloudflare', 'Test - Cloudflare DNS', 'Access Point', '00:00:00:00:00:02', '1.1.1.1', 'UP', 'Static', 'Cloudflare Public DNS', 'Edge 1', 3, 99.9),
('dev-google', 'Test - Google DNS', 'Switch', '00:00:00:00:00:03', '8.8.8.8', 'UP', 'Static', 'Google Public DNS', 'Edge 2', 3, 99.9)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    last_known_ip = EXCLUDED.last_known_ip,
    status = EXCLUDED.status,
    addressing_mode = EXCLUDED.addressing_mode;
