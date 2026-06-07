CREATE TABLE IF NOT EXISTS payment_methods (
    id         CHAR(36)                                             NOT NULL,
    code       VARCHAR(50)                                          NOT NULL,
    type       VARCHAR(50)                                          NOT NULL,
    name       VARCHAR(100)                                         NOT NULL,
    is_active  TINYINT(1)                                           NOT NULL DEFAULT 1,
    created_at DATETIME                                             NOT NULL,
    updated_at DATETIME                                             NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_payment_methods_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
