ALTER TABLE user_addresses
    ADD COLUMN biteship_area_id VARCHAR(100) NOT NULL DEFAULT '' AFTER postal_code;
