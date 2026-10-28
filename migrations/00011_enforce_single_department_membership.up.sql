ALTER TABLE department_memberships
    ADD CONSTRAINT uq_department_memberships_user_id UNIQUE (user_id);
