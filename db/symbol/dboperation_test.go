package symbol

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"strconv"
	"testing"
	"time"
)

const (
	USERNAME = "postgres"
	PASSWD   = "example"
	TIMEZONE = "EST5EDT"
)

var Loc, _ = time.LoadLocation("EST")

func GetQuery(DBNAME string) *Queries {
	dbinfo := fmt.Sprintf("port=5432 host=localhost user=%v password=%v dbname=%v sslmode=disable timezone=%v", USERNAME, PASSWD, DBNAME, TIMEZONE)
	db, err := sql.Open("postgres", dbinfo)
	// open db connection pool

	db.SetMaxOpenConns(30)

	if err != nil {
		fmt.Println(err)
	}
	query := New(db)
	return query
}

func GetTxQuery(DBNAME string) (*Queries, *sql.Tx) {
	dbinfo := fmt.Sprintf("port=5432 host=localhost user=%v password=%v dbname=%v sslmode=disable", USERNAME, PASSWD, DBNAME)
	db, err := sql.Open("postgres", dbinfo)
	// open db connection pool

	db.SetMaxOpenConns(30)

	if err != nil {
		fmt.Println(err)
	}
	query := New(db)

	tx, err := db.Begin()

	if err != nil {
		fmt.Println(err)
	}

	return query, tx
}

func TestMarketOpen(t *testing.T) {
	ctx := context.Background()
	query := GetQuery("U")
	count := 0
	dates := make([]time.Time, 0)
	for i := 0; i < 100; i++ {
		date := time.Date(2023, 3, 7, 0, 0, 0, 0, Loc).Add(time.Hour * -24 * time.Duration(i))
		res, err := query.HisDateMarketOpen(ctx, date)
		if err != nil {
			fmt.Println(err)
		}

		if res == true {
			count += 1
			dates = append(dates, date)
			if count == 10 {
				break
			}
		}

	}
	fmt.Println("latest 10 market days")
	for _, val := range dates {
		fmt.Println(val)
	}

}

type Unit struct {
	Vol      int64
	EventNum int32
	InsVol   int64
	extend   bool
	date     []time.Time
	dateNum  int
}

func GetRedis() *redis.Client {

	opt, err := redis.ParseURL("redis://:@localhost:6379/1")
	if err != nil {
		panic(err)
	}

	return redis.NewClient(opt)
}

func TestGetPeriodData(t *testing.T) {
	ctx := context.Background()
	redisCli := GetRedis()
	query := GetQuery("U")
	symbol := "U"
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, Loc)
	end := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, Loc)
	staticData := make(map[string]*Unit)
	interval := 0
	all_key := make([]string, 0)
	for {

		moment := start.Add(time.Duration(time.Second) * 5 * time.Duration(interval))
		if moment.After(end) {
			break
		}

		//fmt.Println(moment)
		key := symbol + "_" + moment.Format("150405")
		//fmt.Println(key)
		staticData[key] = &Unit{}
		staticData[key].date = make([]time.Time, 0)
		all_key = append(all_key, key)
		interval += 1

	}
	//fmt.Println(interval)
	//fmt.Println(staticData)
	res, err := query.HisPeriodData(ctx, HisPeriodDataParams{Date: "2023-02-22", Date_2: "2023-03-07"})

	if err != nil {
		fmt.Println(err)
	}

	for _, val := range res {

		secTokey := strconv.Itoa((val.Tradingtime.Second() / 5) * 5)
		if secTokey == "0" {
			secTokey = "00"
		} else if secTokey == "5" {
			secTokey = "05"
		}
		HHMMSS := val.Tradingtime.Format("1504") + secTokey
		key := symbol + "_" + HHMMSS
		staticData[key].Vol += val.Volume
		staticData[key].EventNum += val.Eventnum
		staticData[key].InsVol += (val.Volume * int64(val.Insrate/100))
		staticData[key].extend = val.Formt
		staticData[key].date = append(staticData[key].date, time.Date(val.Tradingtime.Year(), val.Tradingtime.Month(), val.Tradingtime.Day(), 0, 0, 0, 0, Loc))
		if val.Tradingtime.Hour() == 18 && val.Tradingtime.Minute() >= 29 && val.Tradingtime.Minute() <= 30 {
			//fmt.Println(val.Tradingtime)
			//fmt.Println(val.Volume)
			//fmt.Println(val.Eventnum)
			//fmt.Println(val.Price)
			//fmt.Println(val.Insrate)
			//fmt.Println("----")
		}
	}

	for _, key := range all_key {
		if staticData[key].Vol >= 0 {
			//fmt.Println(key)
			//fmt.Println(staticData[key].date)
			static := make(map[time.Time]bool)
			for _, val := range staticData[key].date {
				static[val] = true
			}
			if len(static) > 0 {
				staticData[key].dateNum = len(static)
				staticData[key].Vol = staticData[key].Vol / int64(staticData[key].dateNum)
				staticData[key].EventNum = staticData[key].EventNum / int32(staticData[key].dateNum)
				staticData[key].InsVol = staticData[key].InsVol / int64(staticData[key].dateNum)
			}
			if staticData[key].Vol == 0 {
				staticData[key].extend = true
			}
			redisCli.HSet(ctx, key, "Vol", staticData[key].Vol, "EvenNum", staticData[key].EventNum, "InsVol", staticData[key].InsVol, "Extend", staticData[key].extend)
			vol, _ := redisCli.HGet(ctx, key, "Vol").Float64()
			ex, _ := redisCli.HGet(ctx, key, "Extend").Bool()
			fmt.Println(key, vol, ex)
		}

	}

}
