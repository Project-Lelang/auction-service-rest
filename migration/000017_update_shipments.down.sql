ALTER TABLE shipments
    DROP FOREIGN KEY fk_shipments_buyer_addr,
    DROP FOREIGN KEY fk_shipments_seller_addr,
    DROP COLUMN buyer_address_id,
    DROP COLUMN seller_address_id,
    DROP COLUMN buyer_address_snapshot,
    DROP COLUMN seller_address_snapshot,
    DROP COLUMN service_code,
    DROP COLUMN shipping_cost,
    DROP COLUMN estimated_costs;
