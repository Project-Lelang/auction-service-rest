CREATE TABLE IF NOT EXISTS withdrawal_requests (
    id                BIGINT         NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id           BIGINT         NOT NULL,
    validator_user_id BIGINT         DEFAULT NULL,
    amount            DECIMAL(15, 2) NOT NULL,
    status            VARCHAR(20)    NOT NULL DEFAULT 'REQUESTED',
    created_at        DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_withdrawal_requests_user           FOREIGN KEY (user_id)           REFERENCES users (id),
    CONSTRAINT fk_withdrawal_requests_validator_user FOREIGN KEY (validator_user_id) REFERENCES users (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
