-- migrations/000004_create_leave_tables.up.sql
CREATE TABLE leave_balances (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_id        UUID NOT NULL,
    institution_id UUID NOT NULL,
    total_days     INTEGER NOT NULL,
    used_days      INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT fk_leave_balances_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_leave_balances_institution FOREIGN KEY (institution_id) REFERENCES institutions(id)
);

CREATE TABLE leave_requests (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_id        UUID NOT NULL,
    institution_id UUID NOT NULL,
    start_date     TIMESTAMPTZ NOT NULL,
    end_date       TIMESTAMPTZ NOT NULL,
    reason         TEXT,
    status         VARCHAR(50) NOT NULL DEFAULT 'pending',
    reviewed_by    UUID,
    reviewed_at    TIMESTAMPTZ,
    review_note    TEXT,
    CONSTRAINT fk_leave_requests_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_leave_requests_institution FOREIGN KEY (institution_id) REFERENCES institutions(id),
    CONSTRAINT fk_leave_requests_reviewer_user FOREIGN KEY (reviewed_by) REFERENCES users(id)
);
