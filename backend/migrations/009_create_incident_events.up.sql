CREATE TABLE IF NOT EXISTS incident_events (
    id VARCHAR(100) PRIMARY KEY,
    incident_id VARCHAR(100) NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    channel VARCHAR(50),
    detail TEXT NOT NULL,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_incident_events_incident_id ON incident_events(incident_id);
