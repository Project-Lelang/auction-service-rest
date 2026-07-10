CREATE TABLE IF NOT EXISTS auctions (
    id             BIGINT         NOT NULL AUTO_INCREMENT PRIMARY KEY,
    code           VARCHAR(50)    NOT NULL,
    product_id     BIGINT         NOT NULL,
    starting_price DECIMAL(15, 2) NOT NULL,
    start_time     DATETIME       NOT NULL,
    end_time       DATETIME       NOT NULL,
    status         VARCHAR(30)    NOT NULL DEFAULT 'SCHEDULED',
    fee            DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    created_at     DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_auctions_product FOREIGN KEY (product_id) REFERENCES products (id),
    UNIQUE KEY uq_auctions_code (code),
    UNIQUE KEY uq_auctions_scheduled_product ((CASE WHEN status = 'SCHEDULED' THEN product_id ELSE NULL END))
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
