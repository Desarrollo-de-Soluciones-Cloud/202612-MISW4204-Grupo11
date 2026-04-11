
CREATE TABLE IF NOT EXISTS assignments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    academic_space_id BIGINT NOT NULL REFERENCES academic_spaces(id),
    professor_id BIGINT NOT NULL REFERENCES users(id),
    role_in_assignment VARCHAR(25) NOT NULL CHECK (
    role_in_assignment IN ('monitor', 'graduate_assistant')
    ),
    contracted_hours_per_week INT NOT NULL CHECK (contracted_hours_per_week > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_assignment_user_space_role UNIQUE (user_id, academic_space_id, role_in_assignment)
);
CREATE INDEX IF NOT EXISTS idx_assignments_user ON assignments(user_id);
CREATE INDEX IF NOT EXISTS idx_assignments_space ON assignments(academic_space_id);
CREATE INDEX IF NOT EXISTS idx_assignments_professor ON assignments(professor_id);
CREATE INDEX IF NOT EXISTS idx_assignments_user_role ON assignments(user_id, role_in_assignment);