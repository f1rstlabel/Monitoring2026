ALTER TABLE incidents DROP CONSTRAINT IF EXISTS incidents_device_id_fkey;
ALTER TABLE incidents ADD CONSTRAINT incidents_device_id_fkey FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE;
