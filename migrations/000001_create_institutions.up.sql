-- migrations/000001_create_institutions.up.sql
CREATE TABLE institutions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    name       VARCHAR(255) NOT NULL,
    code       VARCHAR(50) NOT NULL,
    address    TEXT,
    phone      VARCHAR(20),
    email      VARCHAR(255),
    is_active  BOOLEAN NOT NULL DEFAULT true
);

CREATE UNIQUE INDEX idx_institutions_code ON institutions(code);
