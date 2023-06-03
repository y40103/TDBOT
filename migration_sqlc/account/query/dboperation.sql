-- name: CreateNewAccount :one
insert into accountinfo (account_id,balance) values ($1,$2)
on conflict do nothing
RETURNING *;

-- name: DeleteAccount :exec

delete FROM accountinfo where account_id=$1
RETURNING *;

-- name: get_version :one
select version from accountinfo where account_id=$1;

-- name: AddAccountBalance :one
update accountinfo set balance = balance + $3,version=version+1 where account_id=$1 and version=$2
RETURNING *;

-- name: QueryBalance :one
SELECT balance FROM accountinfo
where account_id = $1;

-- name: UpdateHoldSymbolStatus :one
insert into hold_symbol (account_id,symbol,quantity)
values ($1,$2,$3)
on conflict (symbol) do update
set quantity = hold_symbol.quantity + $3
RETURNING *;


-- name: CreateHisBuy :one
insert into his_buy (task_id,account_id,symbol,buy_price,quantity,"trigger")
values ($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name: CreateHisSell :one
insert into his_sell (task_id,account_id,symbol,sell_price,quantity,income,"trigger")
values ($1,$2,$3,$4,$5,$6,$7)
RETURNING *;

-- name: GetTaskBuyPrice :one
select buy_price
from his_buy
where task_id=$1;

-- name: CreateNewTask :one
insert into processing_task (task_id,account_id,symbol)
values ($1,$2,$3)
RETURNING *;

-- name: CloseTask :exec
DELETE FROM processing_task
WHERE processing_task.task_id = $1 and
(
 select sum(quantity)=
    (select quantity from his_buy where his_buy.task_id=$1) as close_task
from his_sell
where his_sell.task_id=$1)
;

-- name: CloseSymbol :exec
Delete FROM hold_symbol
where quantity=0;


-- name: GetAccountTaskStatus :many
select * from processing_task
where account_id=$1;




