-- Seed default admin user
-- Credentials:
-- Email: admin@example.com
-- Password: Admin123!
-- Role: admin

INSERT INTO users (
    id,
    first_name,
    last_name,
    phone_number,
    staff_id,
    email,
    password_hash,
    status,
    role,
    created_at,
    updated_at
) VALUES (
    'usr-admin-00000001',
    'System',
    'Admin',
    '+14155550100',
    'STF-ADMIN-001',
    'admin@example.com',
    '$2a$10$0vXD2zCnO6IxbIyNhXWVDu1clkxhJvVIS2yNLBraoALv.V3RVZjNa',
    'active',
    'admin',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (email) DO UPDATE
SET
    role = 'admin',
    status = 'active',
    updated_at = CURRENT_TIMESTAMP;

