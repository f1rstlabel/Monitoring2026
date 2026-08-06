-- 1. Add custom threshold columns to devices
ALTER TABLE devices ADD COLUMN IF NOT EXISTS use_custom_threshold BOOLEAN DEFAULT false;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS custom_failure_threshold INTEGER DEFAULT NULL;

-- 2. Create device_metrics table for SNMP CPU and Memory time-series tracking
CREATE TABLE IF NOT EXISTS device_metrics (
    id VARCHAR(64) PRIMARY KEY,
    device_id VARCHAR(64) NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    metric_type VARCHAR(50) NOT NULL, -- 'cpu' | 'memory'
    value DOUBLE PRECISION NOT NULL,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_device_metrics_device_type ON device_metrics (device_id, metric_type, recorded_at DESC);

-- 3. Add user permissions & status columns to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS permissions JSONB DEFAULT '[]'::jsonb;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- 4. Clean out old fake inventory and insert well-known test targets
DELETE FROM incidents WHERE device_id NOT IN ('dev-local', 'dev-cloudflare', 'dev-google');
DELETE FROM device_status_log WHERE device_id NOT IN ('dev-local', 'dev-cloudflare', 'dev-google');
DELETE FROM ip_change_log WHERE device_id NOT IN ('dev-local', 'dev-cloudflare', 'dev-google');
DELETE FROM devices WHERE id NOT IN ('dev-local', 'dev-cloudflare', 'dev-google');

INSERT INTO devices (id, name, type, mac_address, last_known_ip, status, addressing_mode, location, rack, failure_threshold, uptime_30d) VALUES
('dev-local', 'Test - Localhost', 'Router', '00:00:00:00:00:01', '127.0.0.1', 'UP', 'Static', 'Local Host Server', 'Rack L-01', 2, 100.0),
('dev-cloudflare', 'Test - Cloudflare DNS', 'Access Point', '00:00:00:00:00:02', '1.1.1.1', 'UP', 'Static', 'Cloudflare Public DNS', 'Edge 1', 3, 99.9),
('dev-google', 'Test - Google DNS', 'Switch', '00:00:00:00:00:03', '8.8.8.8', 'UP', 'Static', 'Google Public DNS', 'Edge 2', 3, 99.9)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    last_known_ip = EXCLUDED.last_known_ip,
    status = EXCLUDED.status,
    addressing_mode = EXCLUDED.addressing_mode;
