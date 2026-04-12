-- RF-08: Temporal week management
-- Replace week INT with week_start DATE (the Monday of the week)
-- Add is_late flag for automatic late report detection

ALTER TABLE tasks ADD COLUMN IF NOT EXISTS is_late BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE tasks ADD COLUMN IF NOT EXISTS week_start DATE;

-- Backfill week_start from time_registered.
-- PostgreSQL date_trunc('week', ...) returns the Monday (ISO week start).
UPDATE tasks
SET week_start = date_trunc('week', time_registered)::date
WHERE week_start IS NULL;

ALTER TABLE tasks ALTER COLUMN week_start SET NOT NULL;

ALTER TABLE tasks ADD CONSTRAINT chk_week_start_monday
    CHECK (EXTRACT(ISODOW FROM week_start) = 1);

ALTER TABLE tasks DROP COLUMN IF EXISTS week;

CREATE INDEX IF NOT EXISTS idx_tasks_week_start ON tasks(week_start);
