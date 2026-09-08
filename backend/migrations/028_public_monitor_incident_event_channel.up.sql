ALTER TABLE public_monitor_incident_events
    ADD COLUMN IF NOT EXISTS channel VARCHAR(50) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_public_monitor_incident_events_channel
    ON public_monitor_incident_events (channel);
