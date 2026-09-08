ALTER TABLE public_monitors
    ADD COLUMN IF NOT EXISTS keyword TEXT,
    ADD COLUMN IF NOT EXISTS json_path TEXT,
    ADD COLUMN IF NOT EXISTS expected_value TEXT;

ALTER TABLE public_monitors
    DROP CONSTRAINT IF EXISTS public_monitors_monitor_type_check;
ALTER TABLE public_monitors
    ADD CONSTRAINT public_monitors_monitor_type_check
    CHECK (monitor_type IN ('http', 'http_keyword', 'http_json', 'tcp', 'ping', 'dns'));
