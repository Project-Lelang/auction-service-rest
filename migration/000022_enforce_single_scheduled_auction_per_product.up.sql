-- MySQL permits multiple NULL values in a unique index. This functional index
-- therefore enforces uniqueness only while an auction is SCHEDULED and still
-- allows historical auctions for the same product.
CREATE UNIQUE INDEX uq_auctions_scheduled_product
    ON auctions ((CASE WHEN status = 'SCHEDULED' THEN product_id ELSE NULL END));
