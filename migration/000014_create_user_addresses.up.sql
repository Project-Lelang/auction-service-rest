CREATE TABLE IF NOT EXISTS user_addresses (
    id             BIGINT       NOT NULL AUTO_INCREMENT,
    user_id        BIGINT       NOT NULL,
    label          VARCHAR(100) NOT NULL,
    recipient_name VARCHAR(255) NOT NULL,
    phone          VARCHAR(20)  NOT NULL,
    city_id VARCHAR(100) NOT NULL DEFAULT '',
    city_name      VARCHAR(100) NOT NULL,
    province_name  VARCHAR(100) NOT NULL,
    address        TEXT         NOT NULL,
    postal_code    VARCHAR(10)  NOT NULL,
    biteship_area_id VARCHAR(100) NOT NULL DEFAULT '',
    is_default     TINYINT(1)   NOT NULL DEFAULT 0,
    created_at     DATETIME     NOT NULL,
    updated_at     DATETIME     NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_user_addresses_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
