CREATE TABLE IF NOT EXISTS payments (
    id                BIGINT                                                               NOT NULL AUTO_INCREMENT,
    auction_id        BIGINT                                                               NOT NULL,
    user_id           BIGINT                                                               NOT NULL,
    payment_method_id BIGINT                                                               NULL,
    amount            DECIMAL(15,2)                                                        NOT NULL,
    status            VARCHAR(50)                                                          NOT NULL DEFAULT 'WAITING_FOR_PAYMENT',
    snap_url          VARCHAR(500)                                                         NULL,
    snap_token        VARCHAR(255)                                                         NULL,
    expired_at        DATETIME                                                             NULL,
    created_at        DATETIME                                                             NOT NULL,
    updated_at        DATETIME                                                             NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_payments_auction        FOREIGN KEY (auction_id)        REFERENCES auctions(id),
    CONSTRAINT fk_payments_user           FOREIGN KEY (user_id)           REFERENCES users(id),
    CONSTRAINT fk_payments_method         FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
