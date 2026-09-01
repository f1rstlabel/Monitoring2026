-- 021_add_source_to_ip_change_log.up.sql
ALTER TABLE ip_change_log ADD COLUMN IF NOT EXISTS source VARCHAR(32) DEFAULT 'L3_ARP';
