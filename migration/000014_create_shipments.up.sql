CREATE TABLE shipments (
    id                        CHAR(36)     NOT NULL,
    auction_bid_id            CHAR(36)     NOT NULL,
    user_id                   CHAR(36)     NOT NULL,
    tracking_number           VARCHAR(100) NULL,
    courier_code              VARCHAR(50)  NULL,
    delivery_proof_image_path VARCHAR(500) NULL,
    shipped_at                DATETIME     NULL,
    received_at               DATETIME     NULL,
    created_at                DATETIME     NOT NULL,
    updated_at                DATETIME     NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_shipments_bid  FOREIGN KEY (auction_bid_id) REFERENCES auction_bids(id),
    CONSTRAINT fk_shipments_user FOREIGN KEY (user_id)        REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
