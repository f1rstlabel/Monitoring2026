-- Public monitor type registry and first-class groups.
-- Existing group_name and target_url columns remain for backward compatibility.
CREATE TABLE IF NOT EXISTS public_monitor_groups (
    id            VARCHAR(64) PRIMARY KEY,
    name          VARCHAR(120) NOT NULL UNIQUE,
    description   TEXT NOT NULL DEFAULT '',
    display_order INTEGER NOT NULL DEFAULT 0,
    enabled       BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO public_monitor_groups (id, name, description, display_order, enabled)
VALUES ('grp-ungrouped', 'Ungrouped', 'Monitor yang belum dimasukkan ke group lain.', 0, true)
ON CONFLICT (name) DO NOTHING;

ALTER TABLE public_monitors
    ADD COLUMN IF NOT EXISTS group_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS target_host VARCHAR(255),
    ADD COLUMN IF NOT EXISTS target_port INTEGER,
    ADD COLUMN IF NOT EXISTS dns_record_type VARCHAR(10);

ALTER TABLE public_monitors
    DROP CONSTRAINT IF EXISTS public_monitors_monitor_type_check;
ALTER TABLE public_monitors
    ADD CONSTRAINT public_monitors_monitor_type_check
    CHECK (monitor_type IN ('http', 'tcp', 'ping', 'dns'));

ALTER TABLE public_monitors
    DROP CONSTRAINT IF EXISTS public_monitors_target_port_check;
ALTER TABLE public_monitors
    ADD CONSTRAINT public_monitors_target_port_check
    CHECK (target_port IS NULL OR target_port BETWEEN 1 AND 65535);

ALTER TABLE public_monitors
    DROP CONSTRAINT IF EXISTS fk_public_monitors_group;
ALTER TABLE public_monitors
    ADD CONSTRAINT fk_public_monitors_group
    FOREIGN KEY (group_id) REFERENCES public_monitor_groups(id) ON DELETE SET NULL;

INSERT INTO public_monitor_groups (id, name, description, display_order, enabled)
SELECT 'grp-' || substr(md5(lower(trim(group_name))), 1, 16), trim(group_name), '', 0, true
FROM public_monitors
WHERE trim(COALESCE(group_name, '')) <> ''
GROUP BY trim(group_name)
ON CONFLICT (name) DO NOTHING;

UPDATE public_monitors m
SET group_id = g.id
FROM public_monitor_groups g
WHERE m.group_id IS NULL AND lower(trim(COALESCE(m.group_name, 'Ungrouped'))) = lower(g.name);

UPDATE public_monitors
SET group_id = 'grp-ungrouped'
WHERE group_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_public_monitor_groups_order
    ON public_monitor_groups (display_order, name);
CREATE INDEX IF NOT EXISTS idx_public_monitors_group_id
    ON public_monitors (group_id);

INSERT INTO role_permissions (role, feature_key, enabled) VALUES
('pimpinan', 'public_monitoring.groups.view', true),
('pimpinan', 'public_monitoring.groups.manage', false),
('anggota', 'public_monitoring.groups.view', true),
('anggota', 'public_monitoring.groups.manage', true)
ON CONFLICT (role, feature_key) DO NOTHING;
