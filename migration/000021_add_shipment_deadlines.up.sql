ALTER TABLE shipments
    ADD COLUMN buyer_address_deadline_at DATETIME NULL AFTER delivery_proof_image_path,
    ADD COLUMN ship_deadline_at DATETIME NULL AFTER buyer_address_deadline_at,
    ADD COLUMN receive_deadline_at DATETIME NULL AFTER ship_deadline_at,
    ADD COLUMN delivered_at DATETIME NULL AFTER receive_deadline_at,
    ADD COLUMN auto_received_at DATETIME NULL AFTER received_at,
    ADD COLUMN buyer_address_failed_at DATETIME NULL AFTER auto_received_at,
    ADD COLUMN seller_failed_at DATETIME NULL AFTER buyer_address_failed_at,
    ADD INDEX idx_shipments_buyer_address_deadline (buyer_address_deadline_at),
    ADD INDEX idx_shipments_ship_deadline (ship_deadline_at),
    ADD INDEX idx_shipments_delivered (delivered_at),
    ADD INDEX idx_shipments_receive_deadline (receive_deadline_at);
