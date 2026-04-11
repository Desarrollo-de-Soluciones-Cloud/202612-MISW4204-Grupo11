CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    status TEXT NOT NULL,
    week INT NOT NULL,
    time_invested INT NOT NULL,
    assignment_id INT NOT NULL,
    time_registered TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    observations TEXT
);

CREATE TABLE IF NOT EXISTS attachments (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    storage_path TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_attachments_task_id ON attachments(task_id);