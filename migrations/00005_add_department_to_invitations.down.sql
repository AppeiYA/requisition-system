DROP INDEX IF EXISTS idx_invitations_department_id;

ALTER TABLE invitations
    DROP COLUMN IF EXISTS job_title,
    DROP COLUMN IF EXISTS department_id;
