-- Tasks and their audit log.

CREATE TABLE IF NOT EXISTS tasks (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_id    BIGINT UNSIGNED NOT NULL,
    assignee_id BIGINT UNSIGNED NULL,
    team_id     BIGINT UNSIGNED NOT NULL,
    title       VARCHAR(255)    NOT NULL,
    description TEXT            NULL,
    status      ENUM('todo','in_progress','done','cancelled') NOT NULL DEFAULT 'todo',
    due_date    DATETIME(3)     NULL,
    created_at  DATETIME(3)     NOT NULL,
    updated_at  DATETIME(3)     NOT NULL,
    deleted_at  DATETIME(3)     NULL,
    PRIMARY KEY (id),
    KEY idx_tasks_owner_status (owner_id, status),
    KEY idx_tasks_assignee_status (assignee_id, status),
    KEY idx_tasks_team (team_id),
    KEY idx_tasks_title (title),
    KEY idx_tasks_deleted_at (deleted_at),
    CONSTRAINT fk_tasks_owner FOREIGN KEY (owner_id)
        REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_tasks_assignee FOREIGN KEY (assignee_id)
        REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT fk_tasks_team FOREIGN KEY (team_id)
        REFERENCES teams (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS task_logs (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    task_id    BIGINT UNSIGNED NOT NULL,
    actor_id   BIGINT UNSIGNED NOT NULL,
    action     VARCHAR(50)     NOT NULL,
    old_value  JSON            NULL,
    new_value  JSON            NULL,
    created_at DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    KEY idx_task_logs_task (task_id),
    CONSTRAINT fk_task_logs_task FOREIGN KEY (task_id)
        REFERENCES tasks (id) ON DELETE CASCADE,
    CONSTRAINT fk_task_logs_actor FOREIGN KEY (actor_id)
        REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
