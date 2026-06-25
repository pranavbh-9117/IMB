-- migrations/000006_create_attempt_tables.up.sql
CREATE TABLE quiz_attempts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    institution_id UUID NOT NULL,
    quiz_id        UUID NOT NULL,
    student_id     UUID NOT NULL,
    started_at     TIMESTAMPTZ NOT NULL,
    submitted_at   TIMESTAMPTZ,
    score          INTEGER NOT NULL DEFAULT 0,
    total_marks    INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT fk_quiz_attempts_institution FOREIGN KEY (institution_id) REFERENCES institutions(id),
    CONSTRAINT fk_quiz_attempts_quiz FOREIGN KEY (quiz_id) REFERENCES quizzes(id),
    CONSTRAINT fk_quiz_attempts_student FOREIGN KEY (student_id) REFERENCES users(id)
);
CREATE INDEX idx_quiz_attempts_institution_id ON quiz_attempts(institution_id);
CREATE INDEX idx_quiz_attempts_quiz_id ON quiz_attempts(quiz_id);
CREATE INDEX idx_quiz_attempts_student_id ON quiz_attempts(student_id);

CREATE TABLE quiz_answers (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    attempt_id         UUID NOT NULL,
    question_id        UUID NOT NULL,
    selected_option_id UUID,
    CONSTRAINT fk_quiz_answers_attempt FOREIGN KEY (attempt_id) REFERENCES quiz_attempts(id),
    CONSTRAINT fk_quiz_answers_question FOREIGN KEY (question_id) REFERENCES questions(id),
    CONSTRAINT fk_quiz_answers_selected_option FOREIGN KEY (selected_option_id) REFERENCES options(id)
);
CREATE INDEX idx_quiz_answers_attempt_id ON quiz_answers(attempt_id);
CREATE INDEX idx_quiz_answers_question_id ON quiz_answers(question_id);
