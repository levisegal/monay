-- +goose Up
create table monay.lots (
    id text primary key,
    account_id text not null references monay.accounts (id) on delete cascade,
    security_id text not null references monay.securities (id) on delete cascade,
    transaction_id text not null references monay.transactions (id) on delete cascade,
    acquired_date date not null,
    quantity_micros bigint not null,
    remaining_micros bigint not null,
    cost_basis_micros bigint not null,
    created_at timestamptz not null default now()
);

create index lots_account_id_idx on monay.lots (account_id);
create index lots_security_id_idx on monay.lots (security_id);
create index lots_acquired_date_idx on monay.lots (acquired_date);

create table monay.lot_dispositions (
    id text primary key,
    lot_id text not null references monay.lots (id) on delete cascade,
    sell_transaction_id text not null references monay.transactions (id) on delete cascade,
    disposed_date date not null,
    quantity_micros bigint not null,
    cost_basis_micros bigint not null,
    proceeds_micros bigint not null,
    realized_gain_micros bigint not null,
    holding_period text not null,
    created_at timestamptz not null default now()
);

create index lot_dispositions_lot_id_idx on monay.lot_dispositions (lot_id);
create index lot_dispositions_sell_transaction_id_idx on monay.lot_dispositions (sell_transaction_id);
create index lot_dispositions_disposed_date_idx on monay.lot_dispositions (disposed_date);

-- +goose Down
drop table if exists monay.lot_dispositions;
drop table if exists monay.lots;


