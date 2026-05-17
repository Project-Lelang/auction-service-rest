CREATE TABLE IF NOT EXISTS product_status_histories (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    product_id  CHAR(36) NOT NULL,
    status      VARCHAR(50)         NOT NULL,
    message     TEXT        NULL,
    created_at  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_psh_product FOREIGN KEY (product_id) REFERENCES products(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
