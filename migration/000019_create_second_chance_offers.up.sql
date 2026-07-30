CREATE TABLE IF NOT EXISTS second_chance_offers (
    id         BIGINT      NOT NULL AUTO_INCREMENT,
    auction_id BIGINT      NOT NULL,
    seller_id  BIGINT      NOT NULL,
    buyer_id   BIGINT      NOT NULL,
    bid_id     BIGINT      NOT NULL,
    status     VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    expired_at DATETIME    NULL,
    created_at DATETIME    NOT NULL,
    updated_at DATETIME    NOT NULL,
    PRIMARY KEY (id),
    INDEX idx_second_chance_offers_buyer_status_expired (buyer_id, status, expired_at),
    INDEX idx_second_chance_offers_auction_status (auction_id, status),
    CONSTRAINT fk_second_chance_offers_auction FOREIGN KEY (auction_id) REFERENCES auctions(id),
    CONSTRAINT fk_second_chance_offers_seller FOREIGN KEY (seller_id) REFERENCES users(id),
    CONSTRAINT fk_second_chance_offers_buyer FOREIGN KEY (buyer_id) REFERENCES users(id),
    CONSTRAINT fk_second_chance_offers_bid FOREIGN KEY (bid_id) REFERENCES auction_bids(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
