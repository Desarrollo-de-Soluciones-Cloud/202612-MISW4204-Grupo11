-- RF-14: AI-generated weekly PDF reports

CREATE TABLE IF NOT EXISTS reports (
    id            BIGSERIAL    PRIMARY KEY,
    professor_id  BIGINT       NOT NULL REFERENCES users(id),
    assignment_id BIGINT       NOT NULL REFERENCES assignments(id),
    user_name     TEXT         NOT NULL,
    user_email    TEXT         NOT NULL,
    role          VARCHAR(25)  NOT NULL,
    week_start    DATE         NOT NULL,
    file_path     TEXT         NOT NULL,
    ai_summary    TEXT         NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_report_week_start_monday
        CHECK (EXTRACT(ISODOW FROM week_start) = 1)
);

CREATE INDEX IF NOT EXISTS idx_reports_professor ON reports(professor_id);
CREATE INDEX IF NOT EXISTS idx_reports_professor_week ON reports(professor_id, week_start);
CREATE UNIQUE INDEX IF NOT EXISTS uq_report_assignment_week ON reports(assignment_id, week_start);
