ALTER TABLE user_addresses
    ADD COLUMN latitude DECIMAL(10, 7) NULL AFTER biteship_area_id,
    ADD COLUMN longitude DECIMAL(10, 7) NULL AFTER latitude;
