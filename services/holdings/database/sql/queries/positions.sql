-- name: GetPosition :one
select *
from monay.positions
where id = @id;

-- name: ListPositionsByAccount :many
select
    p.*,
    s.symbol,
    s.name as security_name
from monay.positions p
join monay.securities s on s.id = p.security_id
where p.account_id = @account_id
order by s.symbol;

-- name: ListPositionsByAccountAndDate :many
select
    p.*,
    s.symbol,
    s.name as security_name
from monay.positions p
join monay.securities s on s.id = p.security_id
where p.account_id = @account_id and p.as_of_date = @as_of_date
order by s.symbol;

-- name: UpsertPosition :one
insert into monay.positions (
    id,
    account_id,
    security_id,
    quantity_micros,
    cost_basis_micros,
    market_value_micros,
    as_of_date
) values (
    @id,
    @account_id,
    @security_id,
    @quantity_micros,
    @cost_basis_micros,
    @market_value_micros,
    @as_of_date
)
on conflict (account_id, security_id, as_of_date) do update set
    quantity_micros = excluded.quantity_micros,
    cost_basis_micros = excluded.cost_basis_micros,
    market_value_micros = excluded.market_value_micros,
    updated_at = now()
returning *;

-- name: DeletePosition :exec
delete from monay.positions
where id = @id;
