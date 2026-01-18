-- name: CreateCashTransaction :exec
insert into cash_transactions (
    id,
    account_id,
    transaction_id,
    transaction_date,
    cash_type,
    amount_micros,
    security_id,
    description
) values (
    @id,
    @account_id,
    @transaction_id,
    @transaction_date,
    @cash_type,
    @amount_micros,
    @security_id,
    @description
)
on conflict do nothing;

-- name: GetCashBalance :one
select coalesce(sum(amount_micros), 0) as balance_micros
from cash_transactions
where account_id = @account_id;

-- name: GetCashBalanceAsOfDate :one
select coalesce(sum(amount_micros), 0) as balance_micros
from cash_transactions
where
    account_id = @account_id
    and transaction_date <= @as_of_date;

-- name: ListCashTransactions :many
select
    ct.*,
    s.symbol,
    s.name as security_name
from cash_transactions ct
left join securities s on s.id = ct.security_id
where ct.account_id = @account_id
order by ct.transaction_date desc, ct.created_at desc;

-- name: ListCashTransactionsByDateRange :many
select
    ct.*,
    s.symbol,
    s.name as security_name
from cash_transactions ct
left join securities s on s.id = ct.security_id
where
    ct.account_id = @account_id
    and ct.transaction_date >= @start_date
    and ct.transaction_date <= @end_date
order by ct.transaction_date desc, ct.created_at desc;

-- name: GetOpeningCashBalance :one
select *
from cash_transactions
where
    account_id = @account_id
    and cash_type = 'opening'
limit 1;

-- name: DeleteCashTransactionsByAccount :exec
delete from cash_transactions
where account_id = @account_id;

-- name: DeleteCashTransactionsByTransactionId :exec
delete from cash_transactions
where transaction_id = @transaction_id;

-- name: DeleteNonOpeningCashTransactionsByAccount :exec
delete from cash_transactions
where
    account_id = @account_id
    and cash_type != 'opening';
