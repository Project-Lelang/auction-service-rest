CREATE TABLE IF NOT EXISTS products (
    id              BIGINT          NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT          NOT NULL,
    validator_user_id BIGINT        NULL,
    name            VARCHAR(255)    NOT NULL,
    description     TEXT            NULL,
    `condition`     VARCHAR(20)     NOT NULL DEFAULT 'NEW',
    cover_image_path VARCHAR(512)   NULL,
    image_paths      JSON           NULL,
    status          VARCHAR(50)     NOT NULL DEFAULT 'DRAFT',
    weight_gram INT NOT NULL DEFAULT 1000,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_products_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_products_validator_user FOREIGN KEY (validator_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
