CREATE TABLE IF NOT EXISTS users (
    id                  CHAR(36)        NOT NULL PRIMARY KEY,
    fullname            VARCHAR(255)    NOT NULL,
    phone               VARCHAR(20)     NOT NULL UNIQUE,
    nik                 VARCHAR(50)     DEFAULT NULL,
    birth               DATE            NOT NULL,
    gender              VARCHAR(10)     DEFAULT NULL,
    bank_account_number VARCHAR(50)     DEFAULT NULL,
    is_verified         TINYINT(1)      NOT NULL DEFAULT 0,
    is_deleted          TINYINT(1)      NOT NULL DEFAULT 0,
    password            VARCHAR(255)    NOT NULL,
    created_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
