CREATE TABLE IF NOT EXISTS user_strikes (
    id            BIGINT       NOT NULL AUTO_INCREMENT,
    bidder_id     BIGINT       NOT NULL,
    auction_id    BIGINT       NOT NULL,
    seller_id     BIGINT       NOT NULL,
    strike_reason VARCHAR(50)  NOT NULL,
    status        VARCHAR(50)  NOT NULL DEFAULT 'ACTIVE',
    expired_at    DATETIME     NULL,
    created_at    DATETIME     NOT NULL,
    updated_at    DATETIME     NOT NULL,
    PRIMARY KEY (id),
    INDEX idx_user_strikes_bidder_status_expired (bidder_id, status, expired_at),
    INDEX idx_user_strikes_auction (auction_id),
    CONSTRAINT fk_user_strikes_bidder FOREIGN KEY (bidder_id) REFERENCES users(id),
    CONSTRAINT fk_user_strikes_auction FOREIGN KEY (auction_id) REFERENCES auctions(id),
    CONSTRAINT fk_user_strikes_seller FOREIGN KEY (seller_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
