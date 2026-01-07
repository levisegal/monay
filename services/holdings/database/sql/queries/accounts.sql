-- name: GetAccount :one
select *
from monay.accounts
where id = @id;

-- name: GetAccountByName :one
select *
from monay.accounts
where name = @name;

-- name: GetAccountByExternalNumber :one
select *
from monay.accounts
where
    institution_name = @institution_name
    and external_account_number = @external_account_number;

-- name: ListAccounts :many
select *
from monay.accounts
order by name;

-- name: CreateAccount :one
insert into monay.accounts (
    id,
    user_id,
    name,
    institution_name,
    external_account_number,
    account_type
) values (
    @id,
    @user_id,
    @name,
    @institution_name,
    @external_account_number,
    @account_type
)
returning *;

-- name: UpdateAccount :one
update monay.accounts
set
    name = coalesce(@name, name),
    institution_name = coalesce(@institution_name, institution_name),
    external_account_number = coalesce(@external_account_number, external_account_number),
    account_type = coalesce(@account_type, account_type),
    updated_at = now()
where id = @id
returning *;

-- name: DeleteAccount :exec
delete from monay.accounts
where id = @id;
