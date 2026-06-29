ALTER TABLE payments
    ADD COLUMN code VARCHAR(50) NULL AFTER id;

-- Existing Midtrans transactions used the numeric payment ID as order_id.
-- Preserve that value so notifications for in-flight legacy payments still work.
UPDATE payments
SET code = CAST(id AS CHAR)
WHERE code IS NULL;

ALTER TABLE payments
    MODIFY COLUMN code VARCHAR(50) NOT NULL,
    ADD UNIQUE KEY uq_payments_code (code);
