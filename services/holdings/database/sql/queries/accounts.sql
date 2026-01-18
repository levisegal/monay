-- name: GetAccount :one
select *
from accounts
where id = @id;

-- name: GetAccountByName :one
select *
from accounts
where name = @name;

-- name: GetAccountByExternalNumber :one
select *
from accounts
where
    institution_name = @institution_name
    and external_account_number = @external_account_number;

-- name: ListAccounts :many
select *
from accounts
order by name;

-- name: CreateAccount :one
insert into accounts (
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
update accounts
set
    name = coalesce(nullif(@name, ''), name),
    institution_name = coalesce(nullif(@institution_name, ''), institution_name),
    external_account_number = coalesce(@external_account_number, external_account_number),
    account_type = coalesce(nullif(@account_type, ''), account_type),
    updated_at = datetime('now')
where id = @id
returning *;

-- name: DeleteAccount :exec
delete from accounts
where id = @id;
