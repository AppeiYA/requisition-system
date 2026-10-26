-- Seed default item categories used by requisitions.

INSERT INTO item_categories (
    id,
    name,
    description,
    status,
    created_at,
    updated_at
) VALUES
    (
        'cat-it-00000001',
        'IT & Internet',
        'Computers, networking, internet Services and accessories',
        'active',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        'cat-eq-00000002',
        'Equipment',
        'Tools, machinery, Furniture and other physical assets.',
        'active',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        'cat-os-00000003',
        'Office supplies',
        'Papers, stationery, printers, ink and everyday office needs',
        'active',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        'cat-sw-00000004',
        'Software',
        'Applications, Licenses, Subscriptions and digital tools.',
        'active',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        'cat-mt-00000005',
        'Maintenance',
        'Repairs, Servicing and upkeep of equipment and facilities.',
        'active',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        'cat-tf-00000006',
        'Transportation & Fuel',
        'Fuel, Vehicle hire, Logistics and travel related expenses.',
        'active',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    )
ON CONFLICT (LOWER(name)) DO UPDATE
SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = 'active',
    updated_at = CURRENT_TIMESTAMP;
