-- migrations/000008_create_quiz_materials.up.sql
CREATE TABLE quiz_materials (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    quiz_id           UUID NOT NULL,
    uploaded_by       UUID NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    stored_filename   VARCHAR(255) NOT NULL,
    storage_path      TEXT NOT NULL,
    content_type      VARCHAR(100) NOT NULL,
    file_size         BIGINT NOT NULL,
    CONSTRAINT fk_quiz_materials_quiz FOREIGN KEY (quiz_id) REFERENCES quizzes(id) ON DELETE CASCADE,
    CONSTRAINT fk_quiz_materials_uploader FOREIGN KEY (uploaded_by) REFERENCES users(id)
);
CREATE INDEX idx_quiz_materials_quiz_id ON quiz_materials(quiz_id);
CREATE INDEX idx_quiz_materials_uploaded_by ON quiz_materials(uploaded_by);
