-- +goose Up
create table monay.positions (
    id text primary key,
    account_id text not null references monay.accounts (id) on delete cascade,
    security_id text not null references monay.securities (id) on delete cascade,
    quantity_micros bigint not null,
    cost_basis_micros bigint,
    market_value_micros bigint,
    as_of_date date not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (account_id, security_id, as_of_date)
);

create index positions_account_id_idx on monay.positions (account_id);
create index positions_as_of_date_idx on monay.positions (as_of_date);

create table monay.transactions (
    id text primary key,
    account_id text not null references monay.accounts (id) on delete cascade,
    security_id text references monay.securities (id) on delete cascade,
    transaction_type text not null,
    transaction_date date not null,
    quantity_micros bigint,
    price_micros bigint,
    amount_micros bigint not null,
    fees_micros bigint,
    description text,
    created_at timestamptz not null default now(),
    unique nulls not distinct (account_id, security_id, transaction_type, transaction_date, quantity_micros, amount_micros)
);

create index transactions_account_id_idx on monay.transactions (account_id);
create index transactions_date_idx on monay.transactions (transaction_date);
create index transactions_type_idx on monay.transactions (transaction_type);

-- +goose Down
drop table if exists monay.transactions;
drop table if exists monay.positions;
