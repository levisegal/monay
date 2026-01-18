-- name: DeleteLotsByAccount :exec
delete from lot_dispositions
where lot_id in (select id from lots where account_id = @account_id);

-- name: DeleteLotsForAccount :exec
delete from lots
where account_id = @account_id;

-- name: CreateLot :one
insert into lots (
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
from lots
where id = @id;

-- name: ListLotsByAccountAndSecurity :many
select *
from lots
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
from lots l
join securities s on s.id = l.security_id
where l.account_id = @account_id
order by l.acquired_date asc;

-- name: UpdateLotRemaining :exec
update lots
set remaining_micros = @remaining_micros
where id = @id;

-- name: CreateLotDisposition :one
insert into lot_dispositions (
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
from lot_dispositions d
join lots l on l.id = d.lot_id
where d.sell_transaction_id = @sell_transaction_id;

-- name: ListDispositionsByYear :many
select
    d.*,
    l.acquired_date,
    l.security_id,
    s.symbol,
    s.name as security_name
from lot_dispositions d
join lots l on l.id = d.lot_id
join securities s on s.id = l.security_id
where
    strftime('%Y', d.disposed_date) = @year
order by d.disposed_date asc;

-- name: SumRealizedGainsByYear :one
select
    coalesce(sum(case when holding_period = 'short_term' then realized_gain_micros else 0 end), 0) as short_term_gains,
    coalesce(sum(case when holding_period = 'long_term' then realized_gain_micros else 0 end), 0) as long_term_gains,
    coalesce(sum(realized_gain_micros), 0) as total_gains
from lot_dispositions
where strftime('%Y', disposed_date) = @year;

-- name: SumRemainingBySymbol :many
select
    s.symbol,
    s.id as security_id,
    coalesce(sum(l.remaining_micros), 0) as remaining_micros
from securities s
left join lots l on l.security_id = s.id and l.account_id = @account_id
where s.id in (
    select distinct security_id
    from transactions
    where account_id = @account_id and security_id is not null
)
group by s.symbol, s.id;

-- name: ListHoldingsByAccount :many
select
    s.symbol,
    s.name as security_name,
    sum(l.remaining_micros) as quantity_micros,
    sum(
        case when l.remaining_micros > 0
        then cast(cast(l.cost_basis_micros as real) / cast(l.quantity_micros as real) * cast(l.remaining_micros as real) as integer)
        else 0 end
    ) as cost_basis_micros,
    min(l.acquired_date) as earliest_acquired
from lots l
join securities s on s.id = l.security_id
where l.account_id = @account_id and l.remaining_micros > 0
group by s.symbol, s.name
order by s.symbol;

-- name: ListAllHoldings :many
select
    a.institution_name as broker,
    a.name as account_name,
    s.symbol,
    s.name as security_name,
    sum(l.remaining_micros) as quantity_micros,
    sum(
        case when l.remaining_micros > 0
        then cast(cast(l.cost_basis_micros as real) / cast(l.quantity_micros as real) * cast(l.remaining_micros as real) as integer)
        else 0 end
    ) as cost_basis_micros,
    min(l.acquired_date) as earliest_acquired
from lots l
join securities s on s.id = l.security_id
join accounts a on a.id = l.account_id
where l.remaining_micros > 0
group by a.institution_name, a.name, s.symbol, s.name
order by cost_basis_micros desc;

-- name: ListPositions :many
select
    s.symbol,
    s.name as security_name,
    count(distinct a.id) as account_count,
    sum(l.remaining_micros) as quantity_micros,
    sum(
        case when l.remaining_micros > 0
        then cast(cast(l.cost_basis_micros as real) / cast(l.quantity_micros as real) * cast(l.remaining_micros as real) as integer)
        else 0 end
    ) as cost_basis_micros,
    min(l.acquired_date) as earliest_acquired
from lots l
join securities s on s.id = l.security_id
join accounts a on a.id = l.account_id
where l.remaining_micros > 0
group by s.symbol, s.name
order by cost_basis_micros desc;
