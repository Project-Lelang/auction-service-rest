CREATE TABLE IF NOT EXISTS auctions (
    id             CHAR(36)       NOT NULL PRIMARY KEY,
    product_id     CHAR(36)       NOT NULL,
    starting_price DECIMAL(15, 2) NOT NULL,
    start_time     DATETIME       NOT NULL,
    end_time       DATETIME       NOT NULL,
    status         VARCHAR(30)    NOT NULL DEFAULT 'SCHEDULED',
    fee            DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    created_at     DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_auctions_product FOREIGN KEY (product_id) REFERENCES products (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
