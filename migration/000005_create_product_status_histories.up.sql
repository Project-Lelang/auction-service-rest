CREATE TABLE IF NOT EXISTS product_status_histories (
    id          BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    product_id  BIGINT NOT NULL,
    validator_user_id BIGINT NULL,
    status      VARCHAR(50)         NOT NULL,
    message     TEXT        NULL,
    created_at  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_psh_product FOREIGN KEY (product_id) REFERENCES products(id),
    CONSTRAINT fk_psh_validator_user FOREIGN KEY (validator_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
