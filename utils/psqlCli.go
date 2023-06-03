package utils

import (
	"GoBot/db/symbol"
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // 連接postgres 需import這個
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"runtime"
	"time"
)

// db info
const (
	HOST     = "localhost"
	USERNAME = "postgres"
	PASSWD   = "example"
	TIMEZONE = "EST5EDT"
)

// 處理各標的資料庫交互
type PsqlClient struct {
	DB   *sql.DB         // 非真正連線 但會確認該連線session是可以被使用的
	Conn *symbol.Queries // 真正的連線
}

func (self *PsqlClient) GetSymbolQuery(DBNAME string) {
	dbinfo := fmt.Sprintf("port=5432 host=%v user=%v password=%v dbname=%v sslmode=disable timezone=%v", HOST, USERNAME, PASSWD, DBNAME, TIMEZONE)
	db, err := sql.Open("postgres", dbinfo)
	db.SetMaxOpenConns(runtime.NumCPU() * 10)
	db.SetMaxIdleConns(runtime.NumCPU() * 5)
	if err != nil {
		fmt.Println(err)
	}
	self.DB = db
	self.Conn = symbol.New(self.DB)
}

// 取得某標的某時煎點 前n日 開市日期
func (self *PsqlClient) GetActivityDateRange(Symbol string, EndDate_YYYY_mm_dd string, NumDay int) (Date_YYYY_mm_dd []time.Time) {

	ctx := context.Background()
	if self.Conn == nil {
		self.GetSymbolQuery(Symbol)
	}
	count := 0
	dates := make([]time.Time, 0)
	enddate, err := time.Parse("2006-01-02", EndDate_YYYY_mm_dd)
	if err != nil {
		logrus.Warningln(err)
		return
	}

	for i := 0; i < NumDay*2; i++ {
		date := time.Date(enddate.Year(), enddate.Month(), enddate.Day(), 0, 0, 0, 0, Loc).Add(time.Hour * -24 * time.Duration(i))
		res, err := self.Conn.HisDateMarketOpen(ctx, date)
		if err != nil {
			fmt.Println(err)
		}

		if res == true {
			count += 1
			dates = append(dates, date)
			if count == NumDay {
				break
			}
		}

	}
	logrus.Infof("Get latest %v market days before %v", NumDay, EndDate_YYYY_mm_dd)
	MarketDate := make([]time.Time, 0)
	for _, val := range dates {
		MarketDate = append(MarketDate, val)
		logrus.Infoln(val)
	}

	return MarketDate
}

// 取得某標的  某個日期  之前 N個開市交易日 歷史資料 , 排序為 最新 > 最舊 (單日 由小至大)
func (self *PsqlClient) GetHisData(Symbol string, EndDate string, NumDay int) []*TradingData {
	ctx := context.Background()
	if self.Conn == nil {
		self.GetSymbolQuery(Symbol)
	}
	marketDates := self.GetActivityDateRange(Symbol, EndDate, NumDay)
	beginDate := marketDates[len(marketDates)-1].Format("2006-01-02")
	data, err := self.Conn.HisPeriodData(ctx, symbol.HisPeriodDataParams{Date: beginDate, Date_2: EndDate})
	defer self.CloseSession()
	if err != nil {
		logrus.Warningln(err)
	}
	HisData := make([]*TradingData, 0)
	for _, val := range data {
		price, err := decimal.NewFromString(val.Price)
		if err != nil {
			logrus.Warningln(err)
		}
		vol := decimal.NewFromInt(val.Volume)
		insrate := int32(val.Insrate)
		eventNum := int32(val.Eventnum)
		unitData := &TradingData{Symbol: Symbol, Price: price, Volume: vol, InsRate: insrate, EventNum: eventNum, TradingTime: val.Tradingtime, FormT: val.Formt}
		HisData = append(HisData, unitData)
	}

	return HisData
}

// 取得歷史某時間點 後 N分鐘 資料
func (self *PsqlClient) GetNMinHisData(Symbol string, beginDateYYYY_dd_mm interface{}, EndDateYYYY_dd_mm interface{}, hour int32, minute int32, interval interface{}) (Data []*TradingData) {

	if self.Conn == nil {
		self.GetSymbolQuery(Symbol)
	}

	res, err := self.Conn.HisMinData(context.Background(), symbol.HisMinDataParams{beginDateYYYY_dd_mm, EndDateYYYY_dd_mm, hour, minute, interval})
	defer self.CloseSession()
	if err != nil {
		logrus.Warningln("fail to get GetNMinHisData from db:", err.Error())
	}

	Data = make([]*TradingData, 0)

	for _, t := range res {
		price, _ := decimal.NewFromString(t.Price)
		vol := decimal.NewFromInt(t.Volume)
		unitData := TradingData{Symbol: Symbol, Price: price, Volume: vol, TradingTime: t.Tradingtime, EventNum: t.Eventnum, InsRate: t.Insrate, FormT: t.Formt}
		Data = append(Data, &unitData)
	}

	return Data
}

// 取得過去一段時間 10分鐘整點歷史資料 只會回傳開市資料
func (self *PsqlClient) Get10MinHisData(Symbol string, beginDateYYYY_dd_mm interface{}, EndDateYYYY_dd_mm interface{}) (Data []*TradingData) {

	if self.Conn == nil {
		self.GetSymbolQuery(Symbol)
	}

	res, err := self.Conn.His10MinData(context.Background(), symbol.His10MinDataParams{beginDateYYYY_dd_mm, EndDateYYYY_dd_mm})
	defer self.CloseSession()
	if err != nil {
		logrus.Warningln("fail to get GetNMinHisData from db:", err.Error())
	}

	Data = make([]*TradingData, 0)

	for _, t := range res {
		price, _ := decimal.NewFromString(t.Price)
		vol := decimal.NewFromInt(t.Volume)
		unitData := TradingData{Symbol: Symbol, Price: price, Volume: vol, TradingTime: t.Tradingtime, EventNum: t.Eventnum, InsRate: t.Insrate, FormT: t.Formt}
		Data = append(Data, &unitData)
	}

	return Data
}

// 取得某標的  某個日期  之前 N個開市交易日 歷史資料 , 排序為 最舊至最新  (單日 由小至大)
func (self *PsqlClient) GetHisDataBackTest(Symbol string, EndDate string, NumDay int) []*TradingData {
	ctx := context.Background()
	if self.Conn == nil {
		self.GetSymbolQuery(Symbol)
	}
	marketDates := self.GetActivityDateRange(Symbol, EndDate, NumDay)
	beginDate := marketDates[len(marketDates)-1].Format("2006-01-02")
	data, err := self.Conn.HisPeriodDataBackTesting(ctx, symbol.HisPeriodDataBackTestingParams{Date: beginDate, Date_2: EndDate})
	if err != nil {
		logrus.Warningln(err)
	}
	defer self.CloseSession()
	HisData := make([]*TradingData, 0)
	for _, val := range data {
		price, err := decimal.NewFromString(val.Price)
		if err != nil {
			logrus.Warningln(err)
		}
		vol := decimal.NewFromInt(val.Volume)
		insrate := int32(val.Insrate)
		eventNum := int32(val.Eventnum)
		unitData := &TradingData{Symbol: Symbol, Price: price, Volume: vol, InsRate: insrate, EventNum: eventNum, TradingTime: val.Tradingtime, FormT: val.Formt}
		HisData = append(HisData, unitData)
	}

	return HisData
}

func (self *PsqlClient) GetConn() *symbol.Queries {
	return self.Conn
}

func (self *PsqlClient) GetDB() *sql.DB {
	return self.DB
}

// 關閉此次連線
func (self *PsqlClient) CloseSession() {
	if self.DB != nil {
		err := self.DB.Close()
		if err != nil {
			fmt.Println(err)
		}

		logrus.Infoln("close session...")
	}
}
