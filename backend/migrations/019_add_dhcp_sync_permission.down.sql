-- 019_add_dhcp_sync_permission.down.sql
DELETE FROM role_permissions WHERE feature_key = 'settings.dhcp_sync';
