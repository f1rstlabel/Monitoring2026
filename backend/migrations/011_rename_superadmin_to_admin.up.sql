-- 011_rename_superadmin_to_admin.up.sql
-- Update users role check constraint and migrate existing superadmin rows to admin

-- Update users table check constraint if present
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_role_check'
    ) THEN
        ALTER TABLE users DROP CONSTRAINT users_role_check;
    END IF;
END $$;

-- Update existing superadmin roles in users table
UPDATE users SET role = 'admin' WHERE role = 'superadmin';

-- Add new constraint for roles
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'pimpinan', 'anggota'));

-- Update role_permissions table if any superadmin rows exist
UPDATE role_permissions SET role = 'admin' WHERE role = 'superadmin';
