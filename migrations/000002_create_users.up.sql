-- migrations/000002_create_users.up.sql
CREATE TABLE users (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    name                 VARCHAR(255) NOT NULL,
    email                VARCHAR(255) NOT NULL,
    password_hash        TEXT,
    role                 VARCHAR(50) NOT NULL,
    institution_id       UUID,
    google_id            VARCHAR(255),
    is_active            BOOLEAN NOT NULL DEFAULT true,
    must_change_password BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT fk_users_institution FOREIGN KEY (institution_id) REFERENCES institutions(id)
);

CREATE UNIQUE INDEX idx_users_email ON users(email);
