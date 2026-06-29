ALTER TABLE payments
    DROP INDEX uq_payments_code,
    DROP COLUMN code;
