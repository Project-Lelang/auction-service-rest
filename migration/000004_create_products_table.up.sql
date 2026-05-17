CREATE TABLE IF NOT EXISTS products (
    id              CHAR(36)        NOT NULL PRIMARY KEY,
    user_id         CHAR(36)        NOT NULL,
    name            VARCHAR(255)    NOT NULL,
    description     TEXT            NULL,
    `condition`     VARCHAR(20)     NOT NULL DEFAULT 'NEW',
    cover_image_path VARCHAR(512)   NULL,
    image_paths      JSON           NULL,
    status          VARCHAR(50)     NOT NULL DEFAULT 'DRAFT',
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_products_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
