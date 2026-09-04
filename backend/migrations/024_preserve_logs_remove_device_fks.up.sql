-- 024_preserve_logs_remove_device_fks.up.sql
-- Remove foreign key constraints from historical log tables to devices(id).
-- Deleting a device from inventory should ONLY remove the device record from devices table.
-- Historical data (incidents, status logs, metrics, IP changes) must remain 100% intact
-- with their original device_id, device names, and IPs preserved for auditing & reporting.

-- 1. Drop foreign key constraints dynamically from log tables
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN (
        SELECT tc.table_name, tc.constraint_name
        FROM information_schema.table_constraints AS tc
        JOIN information_schema.referential_constraints AS rc
          ON tc.constraint_name = rc.constraint_name
        WHERE tc.constraint_type = 'FOREIGN KEY'
          AND tc.table_name IN ('incidents', 'device_metrics', 'device_status_log', 'ip_change_log')
          AND rc.unique_constraint_name IN (
            SELECT constraint_name FROM information_schema.table_constraints
            WHERE table_name = 'devices' AND constraint_type = 'PRIMARY KEY'
          )
    ) LOOP
        EXECUTE 'ALTER TABLE ' || quote_ident(r.table_name) || ' DROP CONSTRAINT IF EXISTS ' || quote_ident(r.constraint_name);
    END LOOP;
END $$;

-- 2. Add snapshot columns to incidents table so device information is never lost upon deletion
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS device_name VARCHAR(255) DEFAULT '';
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS device_type VARCHAR(50) DEFAULT '';
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS device_ip VARCHAR(50) DEFAULT '';
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS device_mac VARCHAR(50) DEFAULT '';

-- Backfill existing incidents with device details
UPDATE incidents i
SET device_name = COALESCE(d.name, 'Unknown Device'),
    device_type = COALESCE(d.type, 'Other'),
    device_ip = COALESCE(d.last_known_ip, ''),
    device_mac = COALESCE(d.mac_address, '')
FROM devices d
WHERE i.device_id = d.id AND (i.device_name IS NULL OR i.device_name = '' OR i.device_mac IS NULL OR i.device_mac = '');

-- 3. Add snapshot columns to device_status_log table
ALTER TABLE device_status_log ADD COLUMN IF NOT EXISTS device_name VARCHAR(255) DEFAULT '';
ALTER TABLE device_status_log ADD COLUMN IF NOT EXISTS device_type VARCHAR(50) DEFAULT '';

UPDATE device_status_log l
SET device_name = COALESCE(d.name, 'Unknown Device'),
    device_type = COALESCE(d.type, 'Other')
FROM devices d
WHERE l.device_id = d.id AND (l.device_name IS NULL OR l.device_name = '');

-- 4. Add snapshot columns to ip_change_log table
ALTER TABLE ip_change_log ADD COLUMN IF NOT EXISTS device_name VARCHAR(255) DEFAULT '';

UPDATE ip_change_log i
SET device_name = COALESCE(d.name, 'Unknown Device')
FROM devices d
WHERE i.device_id = d.id AND (i.device_name IS NULL OR i.device_name = '');

-- 5. Reconnect any orphaned incidents to currently registered devices with matching IP or MAC
UPDATE incidents i
SET device_id = d.id
FROM devices d
WHERE (i.device_ip = d.last_known_ip OR (i.device_mac != '' AND LOWER(i.device_mac) = LOWER(d.mac_address)))
  AND i.device_id NOT IN (SELECT id FROM devices);

