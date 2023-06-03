-- name: HisDateMarketOpen :one
select count(*) > 500
from "2023"
where tradingtime between $1 and date ($1)+1
  and formt = False;


-- name: HisPeriodData :many
select tradingtime,
       price,
       volume,
       eventnum,
       insrate,
       formt,
       EXTRACT(YEAR FROM tradingtime)   as "YY",
       EXTRACT(MONTH FROM tradingtime)  as "mm",
       EXTRACT(DAY FROM tradingtime)    as "dd",
       EXTRACT(HOUR FROM tradingtime)   as "h",
       EXTRACT(minute FROM tradingtime) as "m",
       EXTRACT(second FROM tradingtime) as "s"
from "2023"
where tradingtime between date ($1)
  and date ($2)+1
order by "YY" desc, "mm" desc, "dd" desc, "h" asc, "m" asc, "s" asc;


-- name: HisPeriodDataBackTesting :many
select tradingtime,
       price,
       volume,
       eventnum,
       insrate,
       formt,
       EXTRACT(YEAR FROM tradingtime)   as "YY",
       EXTRACT(MONTH FROM tradingtime)  as "mm",
       EXTRACT(DAY FROM tradingtime)    as "dd",
       EXTRACT(HOUR FROM tradingtime)   as "h",
       EXTRACT(minute FROM tradingtime) as "m",
       EXTRACT(second FROM tradingtime) as "s"
from "2023"
where tradingtime between date ($1)
  and date ($2)+1
order by "YY" asc, "mm" asc, "dd" asc, "h" asc, "m" asc, "s" asc;


-- name: HisMinData :many
select tradingtime,
       price,
       volume,
       eventnum,
       insrate,
       formt,
       EXTRACT(YEAR FROM tradingtime)   as "YY",
       EXTRACT(MONTH FROM tradingtime)  as "mm",
       EXTRACT(DAY FROM tradingtime)    as "dd",
       EXTRACT(HOUR FROM tradingtime)   as "h",
       EXTRACT(minute FROM tradingtime) as "m",
       EXTRACT(second FROM tradingtime) as "s"
from "2023"
where (tradingtime between date ($1)
  and date ($2)+1)
  and (EXTRACT (HOUR FROM tradingtime))= cast ($3 AS INTEGER)
  and (EXTRACT (Minute FROM tradingtime) between cast ($4 AS INTEGER)
  and cast ($4+$5 AS INTEGER))
order by "YY" asc, "mm" asc, "dd" asc, "h" asc, "m" asc, "s" asc;



-- name: His10MinData :many
select tradingtime,
       price,
       volume,
       eventnum,
       insrate,
       formt,
       EXTRACT(YEAR FROM tradingtime)   as "YY",
       EXTRACT(MONTH FROM tradingtime)  as "mm",
       EXTRACT(DAY FROM tradingtime)    as "dd",
       EXTRACT(HOUR FROM tradingtime)   as "h",
       EXTRACT(minute FROM tradingtime) as "m",
       EXTRACT(second FROM tradingtime) as "s"
from "2023"
where (tradingtime between date ($1)
  and date ($2)+1)
  and (EXTRACT (Minute FROM tradingtime)%10=0) and (EXTRACT (second FROM tradingtime) < 1) and formt = false
order by "YY" asc, "mm" asc, "dd" asc, "h" asc, "m" asc, "s" asc;
