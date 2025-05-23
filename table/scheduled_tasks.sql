CREATE TABLE scheduled_tasks (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(255) UNIQUE NOT NULL,
    org_id INTEGER REFERENCES organizations(id) NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    interval BIGINT NOT NULL,
    task_type VARCHAR(255) NOT NULL,
    task_data JSONB DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);