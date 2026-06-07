CREATE TABLE IF NOT EXISTS auction_winners (
    id             CHAR(36)                                                   NOT NULL,
    auction_id     CHAR(36)                                                   NOT NULL,
    auction_bid_id CHAR(36)                                                   NOT NULL,
    status         VARCHAR(50)                                                NOT NULL DEFAULT 'WAITING_FOR_PAYMENT',
    created_at     DATETIME                                                   NOT NULL,
    updated_at     DATETIME                                                   NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_auction_winners_auction FOREIGN KEY (auction_id)     REFERENCES auctions(id),
    CONSTRAINT fk_auction_winners_bid     FOREIGN KEY (auction_bid_id) REFERENCES auction_bids(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
