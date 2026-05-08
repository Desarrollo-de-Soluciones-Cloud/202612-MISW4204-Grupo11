ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS assignment_id INT;

ALTER TABLE tasks
ALTER COLUMN time_invested SET DEFAULT 0;

UPDATE tasks
SET time_invested = 0
WHERE time_invested IS NULL;

ALTER TABLE tasks
ALTER COLUMN time_invested SET NOT NULL;

SELECT id, title, assignment_id
FROM tasks
WHERE assignment_id IS NULL;

UPDATE tasks
SET assignment_id = 1
WHERE assignment_id IS NULL;

ALTER TABLE tasks
ALTER COLUMN assignment_id SET NOT NULL;