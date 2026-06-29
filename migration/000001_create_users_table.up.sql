CREATE TABLE IF NOT EXISTS users (
    id                          BIGINT          NOT NULL AUTO_INCREMENT PRIMARY KEY,
    fullname                    VARCHAR(255)    NOT NULL,
    email                       VARCHAR(255)    NOT NULL UNIQUE,
    nik                         VARCHAR(50)     DEFAULT NULL,
    birth                       DATE            NOT NULL,
    gender                      VARCHAR(10)     DEFAULT NULL,
    bank_account_number         VARCHAR(50)     DEFAULT NULL,
    bank_account_name           VARCHAR(100)    DEFAULT NULL,
    bank_name                   VARCHAR(100)    DEFAULT NULL,
    identity_image_path         VARCHAR(255)    DEFAULT NULL,
    selfie_identity_image_path  VARCHAR(255)    DEFAULT NULL,
    balance                     DECIMAL(15,2)   NOT NULL DEFAULT 0.00,
    is_verified                 TINYINT(1)      NOT NULL DEFAULT 0,
    is_deleted                  TINYINT(1)      NOT NULL DEFAULT 0,
    password                    VARCHAR(255)    NOT NULL,
    created_at                  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
