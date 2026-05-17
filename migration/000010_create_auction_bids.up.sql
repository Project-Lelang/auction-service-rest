CREATE TABLE IF NOT EXISTS auction_bids (
    id         CHAR(36)       NOT NULL PRIMARY KEY,
    user_id    CHAR(36)       NOT NULL,
    auction_id CHAR(36)       NOT NULL,
    amount     DECIMAL(15, 2) NOT NULL,
    created_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_auction_bids_user    FOREIGN KEY (user_id)    REFERENCES users (id),
    CONSTRAINT fk_auction_bids_auction FOREIGN KEY (auction_id) REFERENCES auctions (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
