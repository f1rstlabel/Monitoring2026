-- 021_add_source_to_ip_change_log.down.sql
ALTER TABLE ip_change_log DROP COLUMN IF EXISTS source;
