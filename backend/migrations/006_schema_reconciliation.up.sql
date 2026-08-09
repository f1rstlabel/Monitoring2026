-- 006_schema_reconciliation.up.sql
-- Reconcile database schema with corrected ERD specifications

-- 1. Drop legacy tables no longer needed
DROP TABLE IF EXISTS device_dependencies CASCADE;
DROP TABLE IF EXISTS incident_notes CASCADE;

-- 2. Create user_logs table for user activity, sessions, and audit trail
CREATE TABLE IF NOT EXISTS user_logs (
    id          VARCHAR(64) PRIMARY KEY,
    user_id     VARCHAR(64) REFERENCES users(id) ON DELETE SET NULL,
    action      VARCHAR(50) NOT NULL, -- 'login' | 'logout' | 'action'
    detail      TEXT,
    ip_address  VARCHAR(50),
    user_agent  TEXT,
    session_id  VARCHAR(128),
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_logs_occurred ON user_logs (occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_logs_user ON user_logs (user_id);
