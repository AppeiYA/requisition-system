-- Seed default departments
-- Codes are max 4 letters and are normalized (uppercased) in the domain before insert.

INSERT INTO departments (
    id,
    name,
    description,
    code,
    status,
    created_at,
    updated_at
) VALUES
    (
        'dpt-eng-00000001',
        'Engineering',
        'Builds and maintains the requisition system platform.',
        'ENG',
        'active',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        'dpt-des-00000002',
        'Design',
        'Owns product, interaction and visual design.',
        'DES',
        'active',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        'dpt-adm-00000003',
        'Administrator',
        'Manages system administration and day-to-day operations.',
        'ADM',
        'active',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    )
ON CONFLICT (UPPER(code)) DO UPDATE
SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = 'active',
    updated_at = CURRENT_TIMESTAMP;
