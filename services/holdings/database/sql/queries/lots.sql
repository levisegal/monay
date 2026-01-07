-- name: DeleteLotsByAccount :exec
DELETE FROM monay.lot_dispositions
WHERE lot_id IN (SELECT id FROM monay.lots WHERE account_id = @account_id);

-- name: DeleteLotsForAccount :exec
DELETE FROM monay.lots
WHERE account_id = @account_id;

-- name: CreateLot :one
insert into monay.lots (
    id,
    account_id,
    security_id,
    transaction_id,
    acquired_date,
    quantity_micros,
    remaining_micros,
    cost_basis_micros
) values (
    @id,
    @account_id,
    @security_id,
    @transaction_id,
    @acquired_date,
    @quantity_micros,
    @remaining_micros,
    @cost_basis_micros
)
returning *;

-- name: GetLot :one
select *
from monay.lots
where id = @id;

-- name: ListLotsByAccountAndSecurity :many
select *
from monay.lots
where
    account_id = @account_id
    and security_id = @security_id
    and remaining_micros > 0
order by acquired_date asc;

-- name: ListLotsByAccount :many
select
    l.*,
    s.symbol,
    s.name as security_name
from monay.lots l
join monay.securities s on s.id = l.security_id
where l.account_id = @account_id
order by l.acquired_date asc;

-- name: UpdateLotRemaining :exec
update monay.lots
set remaining_micros = @remaining_micros
where id = @id;

-- name: CreateLotDisposition :one
insert into monay.lot_dispositions (
    id,
    lot_id,
    sell_transaction_id,
    disposed_date,
    quantity_micros,
    cost_basis_micros,
    proceeds_micros,
    realized_gain_micros,
    holding_period
) values (
    @id,
    @lot_id,
    @sell_transaction_id,
    @disposed_date,
    @quantity_micros,
    @cost_basis_micros,
    @proceeds_micros,
    @realized_gain_micros,
    @holding_period
)
returning *;

-- name: ListDispositionsBySellTransaction :many
select
    d.*,
    l.acquired_date,
    l.security_id
from monay.lot_dispositions d
join monay.lots l on l.id = d.lot_id
where d.sell_transaction_id = @sell_transaction_id;

-- name: ListDispositionsByYear :many
select
    d.*,
    l.acquired_date,
    l.security_id,
    s.symbol,
    s.name as security_name
from monay.lot_dispositions d
join monay.lots l on l.id = d.lot_id
join monay.securities s on s.id = l.security_id
where
    extract(year from d.disposed_date) = @year
order by d.disposed_date asc;

-- name: SumRealizedGainsByYear :one
select
    coalesce(sum(case when holding_period = 'short_term' then realized_gain_micros else 0 end), 0)::bigint as short_term_gains,
    coalesce(sum(case when holding_period = 'long_term' then realized_gain_micros else 0 end), 0)::bigint as long_term_gains,
    coalesce(sum(realized_gain_micros), 0)::bigint as total_gains
from monay.lot_dispositions
where extract(year from disposed_date) = @year;

