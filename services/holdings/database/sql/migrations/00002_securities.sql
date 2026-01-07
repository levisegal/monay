-- +goose Up
create table monay.securities (
    id text primary key,
    symbol text not null unique,
    name text,
    security_type text,
    cusip text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index securities_cusip_idx on monay.securities (cusip) where cusip is not null;

-- +goose Down
drop table if exists monay.securities;

