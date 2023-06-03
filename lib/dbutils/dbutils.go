package dbutils

import (
	"GoBot/utils"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"reflect"
	"time"
)

type PsqlPool struct {
	SymbolList     []string
	PgxConnPoolMap map[string]*pgxpool.Pool
	InitStatus     bool
	HOST           string
}

func (self *PsqlPool) Init() {

	if !self.InitStatus {
		fmt.Println("initialize...")
		self.createDatabase()
		self.createTable()
		self.PgxConnPoolMap = make(map[string]*pgxpool.Pool)
		for _, symbol := range self.SymbolList {
			db_url := fmt.Sprintf("postgres://postgres:example@%v:5432/%v", self.HOST, symbol)
			dbpool, err := pgxpool.New(context.Background(), db_url)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
				os.Exit(1)
			}

			self.PgxConnPoolMap[symbol] = dbpool
		}
	}

	self.InitStatus = true
	fmt.Println("initialization complete...")

}

func (self *PsqlPool) ClosePool() {
	for _, symbol := range self.SymbolList {
		self.PgxConnPoolMap[symbol].Close()
	}
	defer log.Println("close all pool")

}

func (self *PsqlPool) createDatabase() {

	url := fmt.Sprintf("postgres://postgres:example@%v:5432/%v", self.HOST, "postgres")
	ctx := context.Background()
	conn, _ := pgx.Connect(ctx, url)
	defer conn.Close(ctx)

	for _, symbol := range self.SymbolList {
		SQLDB := fmt.Sprintf("CREATE DATABASE \"%v\"", symbol)
		tag, err := conn.Exec(ctx, SQLDB)

		if err != nil {
			log.Println(err)
		} else {
			log.Println(tag)
		}
	}
}

func (self *PsqlPool) createTable() {

	for _, symbol := range self.SymbolList {
		url := fmt.Sprintf("postgres://postgres:example@%v:5432/%v", self.HOST, symbol)
		ctx := context.Background()
		conn, _ := pgx.Connect(ctx, url)
		year := time.Now().Year()
		SQLTable := fmt.Sprintf("CREATE TABLE \"%v\" (\n  TradingTime timestamptz NOT NULL,\n  Price decimal NOT NULL,\n  volume bigint NOT NULL,\n  EventNum integer NOT NULL,\n  InsRate integer NOT NULL,\n  NextRate integer NOT NULL,\n  FormT boolean NOT NULL,\n  created_at timestamptz DEFAULT (now())\n);", year)
		tag, err := conn.Exec(ctx, SQLTable)
		if err != nil {
			log.Println(err)
		} else {
			log.Println(tag)
			SQLTable = fmt.Sprintf("CREATE INDEX ON \"%v\" (TradingTime);", year)
			tag, err = conn.Exec(ctx, SQLTable)
			if err != nil {
				log.Println(err)
			} else {
				log.Println(tag)
			}
		}
		conn.Close(ctx)
	}

}

func (self *PsqlPool) InsertData(Data utils.TradingData) {
	ctx := context.Background()
	conn, err := self.PgxConnPoolMap[Data.Symbol].Acquire(ctx)
	if err != nil {
		log.Fatalln("can't get conn from pool")
	}
	defer conn.Release()
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	SQL := fmt.Sprintf("insert into \"%v\" (tradingtime,price,volume,eventNum,insrate,nextrate,formt) values ($1,$2,$3,$4,$5,$6,$7)", time.Now().Year())
	res, err := tx.Exec(ctx, SQL, Data.TradingTime, Data.Price, Data.Volume, Data.EventNum, Data.InsRate, Data.NextRate, Data.FormT)
	if err != nil {
		log.Println(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			conn.Release()
		} else {
			tx.Commit(ctx)
			log.Printf("[insert into a row] insert into row: %d", res.RowsAffected())
			conn.Release()
		}
	}()

}

// SCHEMA 綁定 ptr of  struct of slice
func (self *PsqlPool) QueryRows(Conn *pgxpool.Conn, SQL string, SCHEMA interface{}) {

	rtype := reflect.TypeOf(SCHEMA).Elem() // 進來為 *[]struct .elem  = []struct type
	rval := reflect.ValueOf(SCHEMA).Elem() // 進來為 *[]struct .elem  = []struct val
	element_rtype := rtype.Elem()          // 取 struct
	if rtype.Kind() != reflect.Ptr && element_rtype.Kind() != reflect.Struct {
		log.Fatalln("SCHEMA must be a pointer of structure")
	}

	//NumField := rtype.NumField()
	rows, err := Conn.Query(context.Background(), SQL)
	// 取得 iter( single rows )

	defer rows.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	// get db

	template_struct := reflect.New(element_rtype).Elem()
	// 利用 struct type, 建立 struct 的 ptr, 取預設指向一個有預設值的struct, e.g. new(int)返回 *int, 指向0
	// 取得該ptr 上的值 , itr db撈出來的值 依序賦予該 ptr上的值
	for rows.Next() {
		val, _ := rows.Values()
		for index, v := range val {

			template_struct.Field(index).Set(reflect.ValueOf(v))

		}
		rval.Set(reflect.Append(rval, template_struct)) // 將 struct 賦予  []struct
	}

}
