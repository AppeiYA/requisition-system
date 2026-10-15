-- The department assignment travels with the invitation and becomes the new
-- member's department membership on acceptance. Both columns are immutable
-- once the invitation is created.
ALTER TABLE invitations
    ADD COLUMN IF NOT EXISTS department_id VARCHAR(36) NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS job_title VARCHAR(60) NOT NULL;

CREATE INDEX IF NOT EXISTS idx_invitations_department_id ON invitations(department_id);
