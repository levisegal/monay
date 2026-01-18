-- name: GetSecurity :one
select *
from securities
where id = @id;

-- name: GetSecurityBySymbol :one
select *
from securities
where symbol = @symbol;

-- name: ListSecurities :many
select *
from securities
order by symbol;

-- name: UpsertSecurity :one
insert into securities (
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
    name = coalesce(excluded.name, securities.name),
    security_type = coalesce(excluded.security_type, securities.security_type),
    cusip = coalesce(excluded.cusip, securities.cusip),
    updated_at = datetime('now')
returning *;
