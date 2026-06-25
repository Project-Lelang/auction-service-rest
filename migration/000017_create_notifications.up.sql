CREATE TABLE IF NOT EXISTS notifications (
    id                        BIGINT       NOT NULL AUTO_INCREMENT,
    user_id                  BIGINT       NOT NULL,
    title                      VARCHAR(255) NOT NULL,
    body                       TEXT         NOT NULL,
    type                       VARCHAR(50)  NOT NULL,
    reference_id               BIGINT       NULL,
    is_read                    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at                DATETIME     NOT NULL,
    updated_at                DATETIME     NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
