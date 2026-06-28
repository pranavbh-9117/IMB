-- migrations/000009_create_quiz_leaderboard.up.sql

ALTER TABLE quiz_attempts ADD COLUMN IF NOT EXISTS percentage NUMERIC(5,2) NOT NULL DEFAULT 0.0;

CREATE TABLE IF NOT EXISTS quiz_leaderboard (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    quiz_id      UUID NOT NULL,
    student_id   UUID NOT NULL,
    attempt_id   UUID NOT NULL,
    score        INTEGER NOT NULL DEFAULT 0,
    total_marks  INTEGER NOT NULL DEFAULT 0,
    percentage   NUMERIC(5,2) NOT NULL DEFAULT 0.0,
    submitted_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_quiz_leaderboard_quiz FOREIGN KEY (quiz_id) REFERENCES quizzes(id) ON DELETE CASCADE,
    CONSTRAINT fk_quiz_leaderboard_student FOREIGN KEY (student_id) REFERENCES users(id),
    CONSTRAINT fk_quiz_leaderboard_attempt FOREIGN KEY (attempt_id) REFERENCES quiz_attempts(id),
    CONSTRAINT uq_quiz_leaderboard_student UNIQUE (quiz_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_quiz_leaderboard_quiz_id ON quiz_leaderboard(quiz_id);
CREATE INDEX IF NOT EXISTS idx_quiz_leaderboard_student_id ON quiz_leaderboard(student_id);
CREATE INDEX IF NOT EXISTS idx_quiz_leaderboard_ranking ON quiz_leaderboard(quiz_id, score DESC, submitted_at ASC, student_id ASC);
