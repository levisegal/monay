-- name: GetPosition :one
select *
from positions
where id = @id;

-- name: ListAllPositions :many
select
    p.*,
    s.symbol,
    s.name as security_name,
    a.name as account_name
from positions p
join securities s on s.id = p.security_id
join accounts a on a.id = p.account_id
order by a.name, s.symbol;

-- name: ListPositionsByAccount :many
select
    p.*,
    s.symbol,
    s.name as security_name
from positions p
join securities s on s.id = p.security_id
where p.account_id = @account_id
order by s.symbol;

-- name: ListPositionsByAccountAndDate :many
select
    p.*,
    s.symbol,
    s.name as security_name
from positions p
join securities s on s.id = p.security_id
where p.account_id = @account_id and p.as_of_date = @as_of_date
order by s.symbol;

-- name: UpsertPosition :one
insert into positions (
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
    updated_at = datetime('now')
returning *;

-- name: DeletePosition :exec
delete from positions
where id = @id;
