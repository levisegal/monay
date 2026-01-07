-- +goose Up
create table monay.accounts (
    id text primary key,
    user_id text,
    name text not null,
    institution_name text not null,
    external_account_number text,
    account_type text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (institution_name, external_account_number)
);

create index accounts_institution_name_idx on monay.accounts (institution_name);
create index accounts_user_id_idx on monay.accounts (user_id);

-- +goose Down
drop table if exists monay.accounts;
