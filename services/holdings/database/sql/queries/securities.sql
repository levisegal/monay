-- name: GetSecurity :one
select *
from monay.securities
where id = @id;

-- name: GetSecurityBySymbol :one
select *
from monay.securities
where symbol = @symbol;

-- name: ListSecurities :many
select *
from monay.securities
order by symbol;

-- name: UpsertSecurity :one
insert into monay.securities (
    id,
    symbol,
    name,
    security_type,
    cusip
) values (
    @id,
    @symbol,
    @name,
    @security_type,
    @cusip
)
on conflict (symbol) do update set
    name = coalesce(excluded.name, monay.securities.name),
    security_type = coalesce(excluded.security_type, monay.securities.security_type),
    cusip = coalesce(excluded.cusip, monay.securities.cusip),
    updated_at = now()
returning *;

