-- +goose Up
create table monay.cash_transactions (
    id text primary key,
    account_id text not null references monay.accounts (id) on delete cascade,
    transaction_id text references monay.transactions (id) on delete cascade,
    transaction_date date not null,
    cash_type text not null,
    amount_micros bigint not null,
    security_id text references monay.securities (id) on delete set null,
    description text,
    created_at timestamptz not null default now()
);

create index cash_transactions_account_id_idx on monay.cash_transactions (account_id);
create index cash_transactions_date_idx on monay.cash_transactions (transaction_date);
create index cash_transactions_type_idx on monay.cash_transactions (cash_type);

-- +goose Down
drop table if exists monay.cash_transactions;

