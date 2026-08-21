-- 017_add_bulk_devices_permission.down.sql
DELETE FROM role_permissions WHERE feature_key = 'devices.bulk';
