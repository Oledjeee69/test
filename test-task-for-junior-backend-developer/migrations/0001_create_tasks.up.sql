CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	recurrence JSONB NULL,
	source_task_id BIGINT NULL REFERENCES tasks (id) ON DELETE CASCADE,
	scheduled_for DATE NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_source_task_id_scheduled_for ON tasks (source_task_id, scheduled_for) WHERE source_task_id IS NOT NULL AND scheduled_for IS NOT NULL;
