-- Add rich SNMP system fields to devices table
ALTER TABLE devices ADD COLUMN IF NOT EXISTS snmp_sys_descr TEXT DEFAULT '';
ALTER TABLE devices ADD COLUMN IF NOT EXISTS snmp_sys_uptime VARCHAR(100) DEFAULT '';
ALTER TABLE devices ADD COLUMN IF NOT EXISTS snmp_sys_contact VARCHAR(255) DEFAULT '';
ALTER TABLE devices ADD COLUMN IF NOT EXISTS snmp_sys_location VARCHAR(255) DEFAULT '';
