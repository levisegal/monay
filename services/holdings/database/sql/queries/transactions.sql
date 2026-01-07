-- name: GetTransaction :one
select *
from monay.transactions
where id = @id;

-- name: ListTransactionsByAccount :many
select
    t.*,
    s.symbol,
    s.name as security_name
from monay.transactions t
left join monay.securities s on s.id = t.security_id
where t.account_id = @account_id
order by t.transaction_date desc, t.created_at desc;

-- name: ListTransactionsByAccountAndDateRange :many
select
    t.*,
    s.symbol,
    s.name as security_name
from monay.transactions t
left join monay.securities s on s.id = t.security_id
where
    t.account_id = @account_id
    and t.transaction_date >= @start_date
    and t.transaction_date <= @end_date
order by t.transaction_date desc, t.created_at desc;

-- name: CreateTransaction :exec
insert into monay.transactions (
    id,
    account_id,
    security_id,
    transaction_type,
    transaction_date,
    quantity_micros,
    price_micros,
    amount_micros,
    fees_micros,
    description
) values (
    @id,
    @account_id,
    @security_id,
    @transaction_type,
    @transaction_date,
    @quantity_micros,
    @price_micros,
    @amount_micros,
    @fees_micros,
    @description
)
on conflict do nothing;

-- name: DeleteTransaction :exec
delete from monay.transactions
where id = @id;

-- name: DeleteTransactionsByAccount :exec
delete from monay.transactions
where account_id = @account_id;
