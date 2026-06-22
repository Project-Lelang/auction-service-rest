CREATE TABLE IF NOT EXISTS role_requests (
    id          BIGINT          NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id     BIGINT          NOT NULL,
    status      VARCHAR(100)    NOT NULL DEFAULT 'REQUESTED', -- REQUESTED, APPROVED, REJECTED
    role        VARCHAR(100)    NOT NULL,                     -- BIDDER, SELLER
    message     VARCHAR(255)    NULL,                         -- Alasan reject dari admin
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_role_requests_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;