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

-- name: ListCashEquivalentsByAccount :many
select distinct s.symbol
from transactions t
join securities s on t.security_id = s.id
where t.account_id = @account_id
    and s.cash_equivalent = 1;

-- name: UpsertSecurity :one
insert into securities (
    id,
    symbol,
    name,
    security_type,
    cusip,
    cash_equivalent
) values (
    @id,
    @symbol,
    @name,
    @security_type,
    @cusip,
    @cash_equivalent
)
on conflict (symbol) do update set
    name = coalesce(excluded.name, securities.name),
    security_type = coalesce(excluded.security_type, securities.security_type),
    cusip = coalesce(excluded.cusip, securities.cusip),
    cash_equivalent = max(excluded.cash_equivalent, securities.cash_equivalent),
    updated_at = datetime('now')
returning *;
