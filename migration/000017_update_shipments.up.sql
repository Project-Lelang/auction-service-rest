ALTER TABLE shipments
    ADD COLUMN buyer_address_id        CHAR(36)       NULL,
    ADD COLUMN seller_address_id       CHAR(36)       NULL,
    ADD COLUMN buyer_address_snapshot  JSON           NULL,
    ADD COLUMN seller_address_snapshot JSON           NULL,
    ADD COLUMN service_code            VARCHAR(20)    NULL,
    ADD COLUMN shipping_cost           DECIMAL(15,2)  NULL,
    ADD COLUMN estimated_costs         JSON           NULL,
    ADD CONSTRAINT fk_shipments_buyer_addr  FOREIGN KEY (buyer_address_id)  REFERENCES user_addresses(id),
    ADD CONSTRAINT fk_shipments_seller_addr FOREIGN KEY (seller_address_id) REFERENCES user_addresses(id);
