-- migrations/000010_create_daily_institute_statistics.up.sql

CREATE TABLE IF NOT EXISTS daily_institute_statistics (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    institution_id         UUID NOT NULL,
    report_date            DATE NOT NULL,
    total_quiz_attempts    INTEGER NOT NULL DEFAULT 0,
    unique_students_tested INTEGER NOT NULL DEFAULT 0,
    leaves_approved        INTEGER NOT NULL DEFAULT 0,
    leaves_rejected        INTEGER NOT NULL DEFAULT 0,
    leaves_pending         INTEGER NOT NULL DEFAULT 0,
    top_students           JSONB,
    faculty_leave_stats    JSONB,
    CONSTRAINT fk_daily_institute_statistics_institution FOREIGN KEY (institution_id) REFERENCES institutions(id) ON DELETE CASCADE,
    CONSTRAINT uq_daily_institute_statistics_report UNIQUE (institution_id, report_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_institute_statistics_institution_id ON daily_institute_statistics(institution_id);
CREATE INDEX IF NOT EXISTS idx_daily_institute_statistics_report_date ON daily_institute_statistics(report_date);
