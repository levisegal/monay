-- Holdings Service SQLite Schema
-- Tables created on startup, no migrations needed

create table if not exists accounts (
    id text primary key,
    user_id text,
    name text not null,
    institution_name text not null,
    external_account_number text,
    account_type text not null,
    created_at text not null default (datetime('now')),
    updated_at text not null default (datetime('now')),
    unique (institution_name, external_account_number)
);

create index if not exists accounts_institution_name_idx on accounts (institution_name);
create index if not exists accounts_user_id_idx on accounts (user_id);

create table if not exists securities (
    id text primary key,
    symbol text not null unique,
    name text,
    security_type text,
    cusip text,
    created_at text not null default (datetime('now')),
    updated_at text not null default (datetime('now'))
);

create index if not exists securities_cusip_idx on securities (cusip) where cusip is not null;

create table if not exists positions (
    id text primary key,
    account_id text not null references accounts (id) on delete cascade,
    security_id text not null references securities (id) on delete cascade,
    quantity_micros integer not null,
    cost_basis_micros integer,
    market_value_micros integer,
    as_of_date text not null,
    created_at text not null default (datetime('now')),
    updated_at text not null default (datetime('now')),
    unique (account_id, security_id, as_of_date)
);

create index if not exists positions_account_id_idx on positions (account_id);
create index if not exists positions_as_of_date_idx on positions (as_of_date);

create table if not exists transactions (
    id text primary key,
    account_id text not null references accounts (id) on delete cascade,
    security_id text references securities (id) on delete cascade,
    transaction_type text not null,
    transaction_date text not null,
    quantity_micros integer,
    price_micros integer,
    amount_micros integer not null,
    fees_micros integer,
    description text,
    created_at text not null default (datetime('now')),
    unique (account_id, security_id, transaction_type, transaction_date, quantity_micros, amount_micros, description)
);

create index if not exists transactions_account_id_idx on transactions (account_id);
create index if not exists transactions_date_idx on transactions (transaction_date);
create index if not exists transactions_type_idx on transactions (transaction_type);

create table if not exists lots (
    id text primary key,
    account_id text not null references accounts (id) on delete cascade,
    security_id text not null references securities (id) on delete cascade,
    transaction_id text not null references transactions (id) on delete cascade,
    acquired_date text not null,
    quantity_micros integer not null,
    remaining_micros integer not null,
    cost_basis_micros integer not null,
    created_at text not null default (datetime('now'))
);

create index if not exists lots_account_id_idx on lots (account_id);
create index if not exists lots_security_id_idx on lots (security_id);
create index if not exists lots_acquired_date_idx on lots (acquired_date);

create table if not exists lot_dispositions (
    id text primary key,
    lot_id text not null references lots (id) on delete cascade,
    sell_transaction_id text not null references transactions (id) on delete cascade,
    disposed_date text not null,
    quantity_micros integer not null,
    cost_basis_micros integer not null,
    proceeds_micros integer not null,
    realized_gain_micros integer not null,
    holding_period text not null,
    created_at text not null default (datetime('now'))
);

create index if not exists lot_dispositions_lot_id_idx on lot_dispositions (lot_id);
create index if not exists lot_dispositions_sell_transaction_id_idx on lot_dispositions (sell_transaction_id);
create index if not exists lot_dispositions_disposed_date_idx on lot_dispositions (disposed_date);

create table if not exists cash_transactions (
    id text primary key,
    account_id text not null references accounts (id) on delete cascade,
    transaction_id text references transactions (id) on delete cascade,
    transaction_date text not null,
    cash_type text not null,
    amount_micros integer not null,
    security_id text references securities (id) on delete set null,
    description text,
    created_at text not null default (datetime('now'))
);

create index if not exists cash_transactions_account_id_idx on cash_transactions (account_id);
create index if not exists cash_transactions_date_idx on cash_transactions (transaction_date);
create index if not exists cash_transactions_type_idx on cash_transactions (cash_type);
