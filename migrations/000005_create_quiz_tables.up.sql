-- migrations/000005_create_quiz_tables.up.sql
CREATE TABLE quizzes (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    institution_id   UUID NOT NULL,
    created_by       UUID NOT NULL,
    title            VARCHAR(255) NOT NULL,
    description      TEXT,
    duration_minutes INTEGER NOT NULL,
    total_marks      INTEGER NOT NULL DEFAULT 0,
    is_published     BOOLEAN NOT NULL DEFAULT false,
    deleted_at       TIMESTAMPTZ,
    CONSTRAINT fk_quizzes_institution FOREIGN KEY (institution_id) REFERENCES institutions(id),
    CONSTRAINT fk_quizzes_creator FOREIGN KEY (created_by) REFERENCES users(id)
);
CREATE INDEX idx_quizzes_institution_id ON quizzes(institution_id);
CREATE INDEX idx_quizzes_created_by ON quizzes(created_by);
CREATE INDEX idx_quizzes_deleted_at ON quizzes(deleted_at);

CREATE TABLE questions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    quiz_id     UUID NOT NULL,
    text        TEXT NOT NULL,
    marks       INTEGER NOT NULL,
    order_index INTEGER NOT NULL,
    CONSTRAINT fk_questions_quiz FOREIGN KEY (quiz_id) REFERENCES quizzes(id)
);
CREATE INDEX idx_questions_quiz_id ON questions(quiz_id);

CREATE TABLE options (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    question_id UUID NOT NULL,
    text        VARCHAR(255) NOT NULL,
    is_correct  BOOLEAN NOT NULL DEFAULT false,
    order_index INTEGER NOT NULL,
    CONSTRAINT fk_options_question FOREIGN KEY (question_id) REFERENCES questions(id)
);
CREATE INDEX idx_options_question_id ON options(question_id);
