CREATE TABLE IF NOT EXISTS products (
    id              CHAR(36)        NOT NULL PRIMARY KEY,
    user_id         CHAR(36)        NOT NULL,
    name            VARCHAR(255)    NOT NULL,
    description     TEXT            NULL,
    `condition`     ENUM('NEW','PRELOVED') NOT NULL DEFAULT 'NEW',
    cover_image_url VARCHAR(512)    NULL,
    image_urls      JSON            NULL,
    status          ENUM('DRAFT','REQUEST','VERIFIED','REJECTED','ON_BIDS','COMPLETED') NOT NULL DEFAULT 'DRAFT',
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_products_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
