CREATE TABLE IF NOT EXISTS role_requests (
    id         CHAR(36)     NOT NULL PRIMARY KEY,
    user_id    CHAR(36)     NOT NULL,
    status     VARCHAR(20)  NOT NULL DEFAULT 'REQUESTED',
    role       VARCHAR(20)  NOT NULL,
    message    TEXT         DEFAULT NULL,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_role_requests_user FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
