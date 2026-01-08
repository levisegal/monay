-- +goose Up
-- Drop the old unique constraint that doesn't include description
-- This caused same-day sells of different tax lots (same qty/amount) to be treated as duplicates
alter table monay.transactions
    drop constraint if exists transactions_account_id_security_id_transaction_type_transa_key;

-- Add new unique constraint that includes description to distinguish tax lot sales
alter table monay.transactions
    add constraint transactions_unique_key
    unique nulls not distinct (account_id, security_id, transaction_type, transaction_date, quantity_micros, amount_micros, description);

-- +goose Down
alter table monay.transactions
    drop constraint if exists transactions_unique_key;

alter table monay.transactions
    add constraint transactions_account_id_security_id_transaction_type_transa_key
    unique nulls not distinct (account_id, security_id, transaction_type, transaction_date, quantity_micros, amount_micros);

