CREATE TABLE user_addresses (
    id             CHAR(36)     NOT NULL,
    user_id        CHAR(36)     NOT NULL,
    label          VARCHAR(100) NOT NULL,
    recipient_name VARCHAR(255) NOT NULL,
    phone          VARCHAR(20)  NOT NULL,
    city_id        VARCHAR(10)  NOT NULL,
    city_name      VARCHAR(100) NOT NULL,
    province_name  VARCHAR(100) NOT NULL,
    address        TEXT         NOT NULL,
    postal_code    VARCHAR(10)  NOT NULL,
    is_default     TINYINT(1)   NOT NULL DEFAULT 0,
    created_at     DATETIME     NOT NULL,
    updated_at     DATETIME     NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_user_addresses_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
