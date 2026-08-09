-- Add snmp_sys_name column to devices table
ALTER TABLE devices ADD COLUMN IF NOT EXISTS snmp_sys_name VARCHAR(255) DEFAULT '';
