-- Add camera stream and snapshot configuration fields to devices table
ALTER TABLE devices ADD COLUMN IF NOT EXISTS stream_url VARCHAR(255) DEFAULT '';
ALTER TABLE devices ADD COLUMN IF NOT EXISTS snapshot_url VARCHAR(255) DEFAULT '';
ALTER TABLE devices ADD COLUMN IF NOT EXISTS stream_username VARCHAR(100) DEFAULT '';
ALTER TABLE devices ADD COLUMN IF NOT EXISTS stream_password VARCHAR(100) DEFAULT '';
