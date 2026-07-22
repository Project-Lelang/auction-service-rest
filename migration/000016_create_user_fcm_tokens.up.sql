CREATE TABLE IF NOT EXISTS user_fcm_tokens (
    id                        BIGINT       NOT NULL AUTO_INCREMENT,
    user_id                   BIGINT       NOT NULL,
    fcm_token                 VARCHAR(255) NOT NULL,
    created_at                DATETIME     NOT NULL,
    updated_at                DATETIME     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_user_fcm_tokens_fcm_token (fcm_token),
    CONSTRAINT fk_user_fcm_tokens_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
