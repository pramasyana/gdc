-- Idempotency keys used to deduplicate retried POST /tasks requests.

CREATE TABLE IF NOT EXISTS idempotency_keys (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    idempotency_key CHAR(36)        NOT NULL,
    user_id         BIGINT UNSIGNED NOT NULL,
    request_method  VARCHAR(10)     NOT NULL,
    request_path    VARCHAR(255)    NOT NULL,
    request_hash    CHAR(64)        NOT NULL,
    state           ENUM('processing','completed') NOT NULL DEFAULT 'processing',
    response_status INT             NULL,
    response_body   MEDIUMTEXT      NULL,
    created_at      DATETIME(3)     NOT NULL,
    expires_at      DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_idempotency_keys_key (idempotency_key),
    KEY idx_idempotency_keys_expires (expires_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
