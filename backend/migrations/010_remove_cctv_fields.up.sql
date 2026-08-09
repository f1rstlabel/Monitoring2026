-- 010_remove_cctv_fields.up.sql
-- Remove CCTV stream and snapshot fields from devices table

ALTER TABLE devices DROP COLUMN IF EXISTS stream_url;
ALTER TABLE devices DROP COLUMN IF EXISTS snapshot_url;
ALTER TABLE devices DROP COLUMN IF EXISTS stream_username;
ALTER TABLE devices DROP COLUMN IF EXISTS stream_password;
