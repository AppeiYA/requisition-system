-- Links a user to their assigned department. Rows are only created when an
-- invitation is accepted (i.e. when the user actually registers).
CREATE TABLE IF NOT EXISTS department_memberships (
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    department_id VARCHAR(36) NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
    job_title VARCHAR(60) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, department_id)
);

CREATE INDEX IF NOT EXISTS idx_department_memberships_department_id ON department_memberships(department_id);
