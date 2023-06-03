// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: dboperation.sql

package symbol

import (
	"context"
	"time"
)

const his10MinData = `-- name: His10MinData :many
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
order by "YY" asc, "mm" asc, "dd" asc, "h" asc, "m" asc, "s" asc
`

type His10MinDataParams struct {
	Date   interface{}
	Date_2 interface{}
}

type His10MinDataRow struct {
	Tradingtime time.Time
	Price       string
	Volume      int64
	Eventnum    int32
	Insrate     int32
	Formt       bool
	YY          float64
	Mm          float64
	Dd          float64
	H           float64
	M           float64
	S           float64
}

func (q *Queries) His10MinData(ctx context.Context, arg His10MinDataParams) ([]His10MinDataRow, error) {
	rows, err := q.db.QueryContext(ctx, his10MinData, arg.Date, arg.Date_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []His10MinDataRow
	for rows.Next() {
		var i His10MinDataRow
		if err := rows.Scan(
			&i.Tradingtime,
			&i.Price,
			&i.Volume,
			&i.Eventnum,
			&i.Insrate,
			&i.Formt,
			&i.YY,
			&i.Mm,
			&i.Dd,
			&i.H,
			&i.M,
			&i.S,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const hisDateMarketOpen = `-- name: HisDateMarketOpen :one
select count(*) > 500
from "2023"
where tradingtime between $1 and date ($1)+1
  and formt = False
`

func (q *Queries) HisDateMarketOpen(ctx context.Context, tradingtime time.Time) (bool, error) {
	row := q.db.QueryRowContext(ctx, hisDateMarketOpen, tradingtime)
	var column_1 bool
	err := row.Scan(&column_1)
	return column_1, err
}

const hisMinData = `-- name: HisMinData :many
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
order by "YY" asc, "mm" asc, "dd" asc, "h" asc, "m" asc, "s" asc
`

type HisMinDataParams struct {
	Date    interface{}
	Date_2  interface{}
	Column3 int32
	Column4 int32
	Column5 interface{}
}

type HisMinDataRow struct {
	Tradingtime time.Time
	Price       string
	Volume      int64
	Eventnum    int32
	Insrate     int32
	Formt       bool
	YY          float64
	Mm          float64
	Dd          float64
	H           float64
	M           float64
	S           float64
}

func (q *Queries) HisMinData(ctx context.Context, arg HisMinDataParams) ([]HisMinDataRow, error) {
	rows, err := q.db.QueryContext(ctx, hisMinData,
		arg.Date,
		arg.Date_2,
		arg.Column3,
		arg.Column4,
		arg.Column5,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []HisMinDataRow
	for rows.Next() {
		var i HisMinDataRow
		if err := rows.Scan(
			&i.Tradingtime,
			&i.Price,
			&i.Volume,
			&i.Eventnum,
			&i.Insrate,
			&i.Formt,
			&i.YY,
			&i.Mm,
			&i.Dd,
			&i.H,
			&i.M,
			&i.S,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const hisPeriodData = `-- name: HisPeriodData :many
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
order by "YY" desc, "mm" desc, "dd" desc, "h" asc, "m" asc, "s" asc
`

type HisPeriodDataParams struct {
	Date   interface{}
	Date_2 interface{}
}

type HisPeriodDataRow struct {
	Tradingtime time.Time
	Price       string
	Volume      int64
	Eventnum    int32
	Insrate     int32
	Formt       bool
	YY          float64
	Mm          float64
	Dd          float64
	H           float64
	M           float64
	S           float64
}

func (q *Queries) HisPeriodData(ctx context.Context, arg HisPeriodDataParams) ([]HisPeriodDataRow, error) {
	rows, err := q.db.QueryContext(ctx, hisPeriodData, arg.Date, arg.Date_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []HisPeriodDataRow
	for rows.Next() {
		var i HisPeriodDataRow
		if err := rows.Scan(
			&i.Tradingtime,
			&i.Price,
			&i.Volume,
			&i.Eventnum,
			&i.Insrate,
			&i.Formt,
			&i.YY,
			&i.Mm,
			&i.Dd,
			&i.H,
			&i.M,
			&i.S,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const hisPeriodDataBackTesting = `-- name: HisPeriodDataBackTesting :many
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
order by "YY" asc, "mm" asc, "dd" asc, "h" asc, "m" asc, "s" asc
`

type HisPeriodDataBackTestingParams struct {
	Date   interface{}
	Date_2 interface{}
}

type HisPeriodDataBackTestingRow struct {
	Tradingtime time.Time
	Price       string
	Volume      int64
	Eventnum    int32
	Insrate     int32
	Formt       bool
	YY          float64
	Mm          float64
	Dd          float64
	H           float64
	M           float64
	S           float64
}

func (q *Queries) HisPeriodDataBackTesting(ctx context.Context, arg HisPeriodDataBackTestingParams) ([]HisPeriodDataBackTestingRow, error) {
	rows, err := q.db.QueryContext(ctx, hisPeriodDataBackTesting, arg.Date, arg.Date_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []HisPeriodDataBackTestingRow
	for rows.Next() {
		var i HisPeriodDataBackTestingRow
		if err := rows.Scan(
			&i.Tradingtime,
			&i.Price,
			&i.Volume,
			&i.Eventnum,
			&i.Insrate,
			&i.Formt,
			&i.YY,
			&i.Mm,
			&i.Dd,
			&i.H,
			&i.M,
			&i.S,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}