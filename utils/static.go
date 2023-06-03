package utils

import (
	"GoBot/db/symbol"
	"context"
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"math"
	"strconv"
	"strings"
	"time"
)

// 四捨五入 最後保留小數點後面位數
// e,g, 12345.7671
// precision=2, 12345.77
// precision=1, 12345.8
func FloorFloat(n float64, precision float64) float64 {
	scale := math.Pow(10., precision)
	return math.Floor(scale*n) / scale
}

func FloorInt(n float64, digit float64) float64 {
	scale := math.Pow(10., -digit)
	return math.Floor(scale*n) / scale
}

type DBClient interface {
	GetSymbolQuery(symbol string)
	GetActivityDateRange(symbol string, Enddate string, NumDay int) []time.Time
	GetNMinHisData(Symbol string, beginDateYYYY_dd_mm interface{}, EndDateYYYY_dd_mm interface{}, hour int32, minute int32, interval interface{}) (Data []*TradingData)
	Get10MinHisData(Symbol string, beginDateYYYY_dd_mm interface{}, EndDateYYYY_dd_mm interface{}) (Data []*TradingData)
	GetDB() *sql.DB
	GetConn() *symbol.Queries
	CloseSession()
}

var Loc, _ = time.LoadLocation("EST")

type SymbolStatic struct {
	Symbols     []string
	RedisCLi    *redis.Client
	AnyDBClient DBClient
}

type Unit struct {
	Vol      int64
	EventNum int32
	InsVol   int64
	extend   bool
	date     []time.Time
	dateNum  int
	Pwave    decimal.Decimal
	Nwave    decimal.Decimal
}

// 統計資料酷中 所有時間 各價格 交易次數 (具冪等性 重複執行 不影響 若有新資料 會用上次資料持續更新)
func (self *SymbolStatic) StaticAllTradingTimes(Date2_YYYY_mm_dd string) {
	// test redis "2023-03-07T00:00:00-05:00"
	for _, each_symbol := range self.Symbols {

		ctx := context.Background()

		Date2_YYYY_mm_dd_time, err := time.ParseInLocation("2006-01-02", Date2_YYYY_mm_dd, Loc)
		if err != nil {
			logrus.Infoln(err)
		}

		// 假設該天非開市日期 往前查找最新的開市日期
		for i := 0; i <= 100; i++ {
			self.AnyDBClient.GetSymbolQuery(each_symbol)
			open, err := self.AnyDBClient.GetConn().HisDateMarketOpen(ctx, Date2_YYYY_mm_dd_time)

			if err != nil {
				logrus.Infoln(err)
			}
			if open == true {
				Date2_YYYY_mm_dd = Date2_YYYY_mm_dd_time.Format("2006-01-02")
				break
			}
			Date2_YYYY_mm_dd_time = Date2_YYYY_mm_dd_time.Add(-time.Hour * 24)

			if i == 100 {
				panic("can't find market opening day")
			}

		}

		Date1_YYYY_mm_dd := "1900-01-01"
		lastDate, err := self.RedisCLi.Get(ctx, each_symbol+"_AllTradingTimesUpdate").Time()
		fmt.Println(lastDate)
		if err != nil {
			logrus.Infoln(err)
		}
		if lastDate.IsZero() == false {
			Date1_YYYY_mm_dd = lastDate.Add(time.Hour * 24).Format("2006-01-02")
		}

		self.AnyDBClient.GetSymbolQuery(each_symbol)

		res, err := self.AnyDBClient.GetConn().HisPeriodData(ctx, symbol.HisPeriodDataParams{Date: Date1_YYYY_mm_dd, Date_2: Date2_YYYY_mm_dd})
		fmt.Println(Date1_YYYY_mm_dd, Date2_YYYY_mm_dd)
		if err != nil {
			fmt.Println(err)
		}

		HashKey := each_symbol + "_AllTradingTimes"
		fmt.Println(HashKey)
		temp := make(map[string]int64)
		for index, val := range res {

			val_float, _ := strconv.ParseFloat(val.Price, 4)

			key_1f := fmt.Sprintf("%.1f", FloorFloat(val_float, 1)) + "_1f" + "-"
			key_1d := fmt.Sprintf("%d", int(FloorFloat(val_float, 0))) + "_1d" + "-"
			key_2d := fmt.Sprintf("%d", int(FloorInt(val_float, 1))) + "_2d" + "-"

			temp[key_1f] += val.Volume
			temp[key_1d] += val.Volume
			temp[key_2d] += val.Volume

			if index == len(res)-1 {
				res1f := self.RedisCLi.HGet(ctx, HashKey, key_1f)
				if res1f == nil {
					self.RedisCLi.HSet(ctx, HashKey, key_1f, 0)
				}
				res1d := self.RedisCLi.HGet(ctx, HashKey, key_1d)
				if res1d == nil {
					self.RedisCLi.HSet(ctx, HashKey, key_1d, 0)
				}
				res2d := self.RedisCLi.HGet(ctx, HashKey, key_2d)
				if res2d == nil {
					self.RedisCLi.HSet(ctx, HashKey, key_2d, 0)
				}

				for key, val := range temp {
					self.RedisCLi.HGet(ctx, key, "")
					self.RedisCLi.HIncrBy(ctx, HashKey, key, val)
				}

			}

		}

		latestUpdate, _ := time.ParseInLocation("2006-01-02", Date2_YYYY_mm_dd, Loc)
		fmt.Println(each_symbol + "_AllTradingTimesUpdate")
		self.RedisCLi.Set(ctx, each_symbol+"_AllTradingTimesUpdate", latestUpdate, 0)

	}

}

// 統計某個時間點前 N天開市日 價格區間 交易次數 輸出至 redis, (執行前會刪除舊版本)
func (self *SymbolStatic) StaticNDayTradingTimes(Date2_YYYY_mm_dd string, DayNum int) {

	for _, each_symbol := range self.Symbols {

		ctx := context.Background()

		Date2_YYYY_mm_dd_time, err := time.ParseInLocation("2006-01-02", Date2_YYYY_mm_dd, Loc)
		if err != nil {
			logrus.Infoln(err)
		}

		// 假設該天非開市日期 往前查找最新的開市日期
		for i := 0; i <= 100; i++ {
			self.AnyDBClient.GetSymbolQuery(each_symbol)
			open, err := self.AnyDBClient.GetConn().HisDateMarketOpen(ctx, Date2_YYYY_mm_dd_time)

			if err != nil {
				logrus.Infoln(err)
			}
			if open == true {
				Date2_YYYY_mm_dd = Date2_YYYY_mm_dd_time.Format("2006-01-02")
				break
			}
			Date2_YYYY_mm_dd_time = Date2_YYYY_mm_dd_time.Add(-time.Hour * 24)

			if i == 100 {
				panic("can't find market opening day")
			}

		}

		MarketDays := self.AnyDBClient.GetActivityDateRange(each_symbol, Date2_YYYY_mm_dd, DayNum)
		Date1_YYYY_mm_dd := MarketDays[len(MarketDays)-1]
		// 查找最前面的開市日期

		Date2_YYYY_mm_dd_time = Date2_YYYY_mm_dd_time.Add(time.Hour * 24)
		if err != nil {
			logrus.Infoln(err)
		}

		if err != nil {
			logrus.Infoln(err)
		}

		res, err := self.AnyDBClient.GetConn().HisPeriodData(ctx, symbol.HisPeriodDataParams{Date: Date1_YYYY_mm_dd, Date_2: Date2_YYYY_mm_dd})
		fmt.Println(Date1_YYYY_mm_dd, Date2_YYYY_mm_dd)
		if err != nil {
			fmt.Println(err)
		}

		HashKey := each_symbol + "_" + fmt.Sprintf("%v", DayNum) + "D" + "TradingTimes"
		success, _ := self.RedisCLi.Del(ctx, HashKey).Result()
		if success == 1 {
			logrus.Infoln("delete previous version hash key:", HashKey)
		}
		fmt.Println(HashKey)
		temp := make(map[string]int64)
		for index, val := range res {

			val_float, _ := strconv.ParseFloat(val.Price, 4)

			key_1f := fmt.Sprintf("%.1f", FloorFloat(val_float, 1)) + "_1f" + "-"
			key_1d := fmt.Sprintf("%d", int(FloorFloat(val_float, 0))) + "_1d" + "-"
			key_2d := fmt.Sprintf("%d", int(FloorInt(val_float, 1))) + "_2d" + "-"

			temp[key_1f] += val.Volume
			temp[key_1d] += val.Volume
			temp[key_2d] += val.Volume

			if index == len(res)-1 {
				res1f := self.RedisCLi.HGet(ctx, HashKey, key_1f)
				if res1f == nil {
					self.RedisCLi.HSet(ctx, HashKey, key_1f, 0)
				}
				res1d := self.RedisCLi.HGet(ctx, HashKey, key_1d)
				if res1d == nil {
					self.RedisCLi.HSet(ctx, HashKey, key_1d, 0)
				}
				res2d := self.RedisCLi.HGet(ctx, HashKey, key_2d)
				if res2d == nil {
					self.RedisCLi.HSet(ctx, HashKey, key_2d, 0)
				}

				for key, val := range temp {
					self.RedisCLi.HGet(ctx, key, "")
					self.RedisCLi.HIncrBy(ctx, HashKey, key, val)
				}

			}

		}
		EndofUpdate := Date1_YYYY_mm_dd.Format("2006.01.02")
		latestUpdate, _ := time.ParseInLocation("2006-01-02", Date2_YYYY_mm_dd, Loc)
		fmt.Println(each_symbol + "_" + fmt.Sprintf("%v", DayNum) + "D" + "TradingTimesUpdate")
		self.RedisCLi.Set(ctx, each_symbol+"_"+fmt.Sprintf("%v", DayNum)+"D"+"TradingTimesUpdate", latestUpdate, 0)
		fmt.Println(each_symbol + "_" + fmt.Sprintf("%v", DayNum) + "D" + "TradingTimesUpdateEndOf")
		self.RedisCLi.Set(ctx, each_symbol+"_"+fmt.Sprintf("%v", DayNum)+"D"+"TradingTimesUpdateEndOf", EndofUpdate, 0)

	}

}

// 價位壓力排名
type BufferRank struct {
	Symbol       string
	Price        []string
	TradingTimes []int
}

// DayNum: any number or 0 e.g. 2,3,5,10,0=All , digit: 1d 2d 1f
// 統計 所有已產生的交易次數統計 由大至小進行排序
func (self *SymbolStatic) StaticNDayTradingTimesRank(Symbol string, DayNum int, digit string) *BufferRank {
	var timeRange string
	if DayNum == 0 {
		timeRange = "All"
	} else {
		timeRange = fmt.Sprintf("%vD", DayNum)
	}

	ctx := context.Background()
	hkey := fmt.Sprintf("%v_%vTradingTimes", Symbol, timeRange)
	cmds, err := self.RedisCLi.HKeys(ctx, hkey).Result()
	if err != nil {
		logrus.Infoln(err)
	}
	keys := make([]string, 0)
	volumes := make([]int, 0)
	for _, field_key := range cmds {
		exists := strings.Contains(field_key, digit)
		if exists {
			keys = append(keys, field_key)
		}
	}

	vols, err := self.RedisCLi.HMGet(ctx, hkey, keys...).Result()

	if err != nil {
		logrus.Infoln(err)
	}

	for _, val := range vols {
		intVal, err := strconv.Atoi(val.(string))
		if err != nil {
			logrus.Infoln(err)
		}
		volumes = append(volumes, intVal)
	}

	size := len(vols)

	for i := 0; i < size; i++ {

		for j := 0; j < size-i-1; j++ {

			if volumes[j] <= volumes[j+1] {
				volumes[j], volumes[j+1] = volumes[j+1], volumes[j]
				keys[j], keys[j+1] = keys[j+1], keys[j]
			}

		}

	}

	//fmt.Println(keys)
	//fmt.Println(volumes)

	if keys != nil && volumes != nil {
		return &BufferRank{Symbol: Symbol, Price: keys, TradingTimes: volumes}
	}

	return nil

}

// 垂直統計
// 計算一段日期 每日某段每五秒內 平均交易量 寫入redis key:value ,<symbo_name>_mmddHHMMSS
// 統計範例為 e.g. 093001 093002 ... ~ 093004.999  都會統計在 093000 中
func (self *SymbolStatic) StaticInfo(Date1_YYYY_mm_dd string, Date2_YYYY_mm_dd string) {

	ctx := context.Background()
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, Loc)
	end := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, Loc)

	for _, s := range self.Symbols {

		staticData := make(map[string]*Unit)
		interval := 0
		all_key := make([]string, 0)
		self.AnyDBClient.GetSymbolQuery(s)

		for {

			moment := start.Add(time.Duration(time.Second) * 5 * time.Duration(interval))
			if moment.After(end) {
				break
			}

			key := s + "_" + moment.Format("150405")
			staticData[key] = &Unit{}
			staticData[key].date = make([]time.Time, 0)
			all_key = append(all_key, key)
			interval += 1

		}

		res, err := self.AnyDBClient.GetConn().HisPeriodData(ctx, symbol.HisPeriodDataParams{Date: Date1_YYYY_mm_dd, Date_2: Date2_YYYY_mm_dd})

		if err != nil {
			fmt.Println(err)
		}

		for _, val := range res {
			secTokey := fmt.Sprintf("%02d", (val.Tradingtime.Second()/5)*5)

			HHMMSS := val.Tradingtime.Format("1504") + secTokey
			key := s + "_" + HHMMSS

			staticData[key].Vol += val.Volume
			staticData[key].EventNum += val.Eventnum
			staticData[key].InsVol += (val.Volume * int64(val.Insrate/100))
			staticData[key].extend = val.Formt
			staticData[key].date = append(staticData[key].date, time.Date(val.Tradingtime.Year(), val.Tradingtime.Month(), val.Tradingtime.Day(), 0, 0, 0, 0, Loc))
		}

		for _, key := range all_key {
			if staticData[key].Vol >= 0 {
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
				self.RedisCLi.HSet(ctx, key, "Vol", staticData[key].Vol, "EventNum", staticData[key].EventNum, "InsVol", staticData[key].InsVol, "Extend", staticData[key].extend)

			}
			update_start := s + "_StaticInfoBegin"
			update_end := s + "_StaticInfoEnd"
			self.RedisCLi.Set(ctx, update_start, Date1_YYYY_mm_dd, 0)
			self.RedisCLi.Set(ctx, update_end, Date2_YYYY_mm_dd, 0)
		}

	}

}

type UnitWave struct {
	Pwave decimal.Decimal
	Nwave decimal.Decimal
}

// 為百分比寫至redis
func (self *SymbolStatic) StaticWave(timeSec int, Date1_YYYY_mm_dd string, Date2_YYYY_mm_dd string) {

	ctx := context.Background()

	for _, s := range self.Symbols {
		self.AnyDBClient.GetSymbolQuery(s)
		defer self.AnyDBClient.CloseSession()
		res, err := self.AnyDBClient.GetConn().HisPeriodData(ctx, symbol.HisPeriodDataParams{Date: Date1_YYYY_mm_dd, Date_2: Date2_YYYY_mm_dd})

		if err != nil {
			fmt.Println(err)
		}

		sq := SQ{TimeAllowSec: time.Second * time.Duration(timeSec)}
		unitWave := &UnitWave{}
		pcount, ncount := decimal.Zero, decimal.Zero
		for _, val := range res {
			price, _ := decimal.NewFromString(val.Price)
			vol := decimal.NewFromInt(val.Volume)
			eventNum := val.Eventnum

			data := TradingData{Symbol: s, Price: price, Volume: vol, EventNum: eventNum, TradingTime: val.Tradingtime}
			sq.Append(&data)
			ind5 := Indicator{sq.Data}

			candle := ind5.GetCandles()

			pWave := candle.High.Price.Sub(candle.Open.Price)
			nWave := candle.Open.Price.Sub(candle.Low.Price)
			if !pWave.Equal(decimal.Zero) {
				unitWave.Pwave = unitWave.Pwave.Add(pWave)
				pcount = pcount.Add(decimal.NewFromInt(1))
			}
			if !nWave.Equal(decimal.Zero) {
				unitWave.Nwave = unitWave.Nwave.Add(nWave)
				ncount = ncount.Add(decimal.NewFromInt(1))
			}

		}
		sq.RedisCli = GetRedis("0")
		staticPwave := unitWave.Pwave.Div(pcount)
		staticNwave := unitWave.Nwave.Div(ncount)
		referPrice, _ := decimal.NewFromString(res[0].Price)

		fmt.Printf("%v - %v\n", Date1_YYYY_mm_dd, Date2_YYYY_mm_dd)
		fmt.Println("正向波動", staticPwave, "負向波動", staticNwave)
		fmt.Println(pcount, ncount)
		fmt.Println("正向百分比波動率", (staticPwave.Div(referPrice))) // 參考最後一天第一個價格
		fmt.Println("負向百分比波動率", (staticNwave.Div(referPrice)))

		key := s + "_Wave"
		pwave, _ := (staticPwave.Div(referPrice)).Float64()
		pwave += 1.0
		nwave, _ := (staticNwave.Div(referPrice)).Float64()
		nwave += 1.0
		self.RedisCLi.Del(ctx, key)
		self.RedisCLi.HSet(ctx, key, "pwave", pwave, "nwave", nwave)

		update_start := s + "_WaveBegin"
		update_end := s + "_WaveEnd"
		self.RedisCLi.Set(ctx, update_start, Date1_YYYY_mm_dd, 0)
		self.RedisCLi.Set(ctx, update_end, Date2_YYYY_mm_dd, 0)

	}

}

type WaveNMin struct {
	Pwave    decimal.Decimal
	Nwave    decimal.Decimal
	P        decimal.Decimal // Pwave不為0的總統計數量
	N        decimal.Decimal // Nwave不為0的總統計數量
	AvgPWave decimal.Decimal // 波動幅度  為波動數值 非百分比 , 最後sq取得後 會再轉成百分比 再存為屬性
	AvgNWave decimal.Decimal // 波動幅度  為波動數值 非百分比
}

// 開始統計結果
func (self *WaveNMin) CalAvgWave() {

	if self.P.Equal(decimal.Zero) {
		self.AvgPWave = decimal.Zero
	} else {

		// 結果類似為0.0000XXXX
		self.AvgPWave = self.Pwave.Div(self.P)

	}

	if self.N.Equal(decimal.Zero) {
		self.AvgNWave = decimal.Zero
	} else {

		self.AvgNWave = self.Nwave.Div(self.N)

	}

}

// 計算某個時間點後 間隔 interval 分鐘內 平均 波動幅度 (timeSec SQ open close 差距 為一個單位 將分鐘內所有產生的進行平均)
func (self *SymbolStatic) StaticUnitWave(timeSec int, EndDate string, NumDay int, interval int) {
	if self.RedisCLi == nil {
		self.RedisCLi = GetRedis("0")
	}

	pipe := self.RedisCLi.Pipeline()

	for _, s := range self.Symbols {
		self.AnyDBClient.GetSymbolQuery(s)
		dates := self.AnyDBClient.GetActivityDateRange(s, EndDate, NumDay)
		self.AnyDBClient.CloseSession()

		now := time.Now()
		beginDate := dates[len(dates)-1].Format("2006-01-02")
		EndDate = dates[0].Format("2006-01-02")
		hourBegin := 9
		minuteBeing := 30
		referHourMin := time.Date(now.Year(), now.Month(), now.Day(), hourBegin, minuteBeing, 0, 0, Loc)

		for i := 0; i < 5000; i++ { // 6.5*60 = 3900 為一天開市最多分鐘數 防止意外loop , 設定5000
			self.AnyDBClient.GetSymbolQuery(s)
			res := self.AnyDBClient.GetNMinHisData(
				s,
				beginDate,
				EndDate,
				int32(referHourMin.Hour()),
				int32(referHourMin.Minute()),
				interval,
			)
			self.AnyDBClient.CloseSession()

			if len(res) > 0 && res[0].FormT == true && res[0].TradingTime.Hour() >= 15 {
				break
			}

			currentDate := 0

			lastIndex := len(res) - 1

			sq := SQ{TimeAllowSec: time.Duration(timeSec) * time.Second, RedisCli: self.RedisCLi}
			ind := Indicator{}

			NminWave := new(WaveNMin)

			for index, data := range res {

				newDate := data.TradingTime.Day()

				// 初始第一個
				if currentDate == 0 {

					currentDate = newDate

				}

				sq.Append(data)

				ind.Data = sq.Data
				candle := ind.GetCandles()

				// n秒內 波動價差
				unitPWave := candle.High.Price.Sub(candle.Open.Price)
				unitNWave := candle.Low.Price.Sub(candle.Open.Price)

				if unitNWave.LessThan(decimal.Zero) {
					NminWave.N = NminWave.N.Add(decimal.NewFromInt(1))
					NminWave.Nwave = NminWave.Nwave.Add(unitNWave.Abs())
				}

				if unitPWave.GreaterThan(decimal.Zero) {
					NminWave.P = NminWave.P.Add(decimal.NewFromInt(1))
					NminWave.Pwave = NminWave.Pwave.Add(unitPWave.Abs())
				}

				//最後一項
				if index == lastIndex {
					NminWave.CalAvgWave()
					Hkey := s + "_" + "NminWave"
					Pfiled := fmt.Sprintf("%02d%02d", referHourMin.Hour(), referHourMin.Minute()) + "PWave"
					Nfiled := fmt.Sprintf("%02d%02d", referHourMin.Hour(), referHourMin.Minute()) + "NWave"
					res0 := pipe.HSet(context.Background(), Hkey, Pfiled, NminWave.AvgPWave.String())
					logrus.Infoln("PIPLINE ADD: ", res0)
					res1 := pipe.HSet(context.Background(), Hkey, Nfiled, NminWave.AvgNWave.String())
					logrus.Infoln("PIPLINE ADD: ", res1)
				}

			}

			referHourMin = referHourMin.Add(time.Minute * time.Duration(interval))

		}

	}

	_, err := pipe.Exec(context.Background())

	if err != nil {
		logrus.Warningln("fail to execute the StaticUnitWave ")
		panic(err)
	}

	logrus.Infoln("success to execute the StaticUnitWave")

}

// 計算某個時間點後 間隔 interval 分鐘內 平均 波動幅度 (timeSec SQ open close 差距 為一個單位 將分鐘內所有產生的進行平均)
func (self *SymbolStatic) StaticOpenUnitWave(timeSec int, EndDate string, NumDay int, interval int) {
	if self.RedisCLi == nil {
		self.RedisCLi = GetRedis("0")
	}

	pipe := self.RedisCLi.Pipeline()

	for _, s := range self.Symbols {

		dates := self.AnyDBClient.GetActivityDateRange(s, EndDate, NumDay)
		self.AnyDBClient.GetSymbolQuery(s)
		defer self.AnyDBClient.CloseSession()
		now := time.Now()
		beginDate := dates[len(dates)-1].Format("2006-01-02")
		EndDate = dates[0].Format("2006-01-02")
		hourBegin := 9
		minuteBeing := 30
		referHourMin := time.Date(now.Year(), now.Month(), now.Day(), hourBegin, minuteBeing, 0, 0, Loc)

		for i := 0; i < 1; i++ { // 6.5*60 = 3900 為一天開市最多分鐘數 防止意外loop , 設定5000
			self.AnyDBClient.GetSymbolQuery(s)
			res := self.AnyDBClient.GetNMinHisData(
				s,
				beginDate,
				EndDate,
				int32(referHourMin.Hour()),
				int32(referHourMin.Minute()),
				interval,
			)
			self.AnyDBClient.CloseSession()
			if len(res) > 0 && res[0].FormT == true && res[0].TradingTime.Hour() >= 15 {
				break
			}

			currentDate := 0

			lastIndex := len(res) - 1

			sq := SQ{TimeAllowSec: time.Duration(timeSec) * time.Second, RedisCli: self.RedisCLi}
			ind := Indicator{}

			NminWave := new(WaveNMin)

			for index, data := range res {

				newDate := data.TradingTime.Day()

				// 初始第一個
				if currentDate == 0 {

					currentDate = newDate

				}

				sq.Append(data)

				ind.Data = sq.Data
				candle := ind.GetCandles()

				// n秒內 波動價差
				unitPWave := candle.High.Price.Sub(candle.Open.Price)
				unitNWave := candle.Low.Price.Sub(candle.Open.Price)

				if unitNWave.LessThan(decimal.Zero) {
					NminWave.N = NminWave.N.Add(decimal.NewFromInt(1))
					NminWave.Nwave = NminWave.Nwave.Add(unitNWave.Abs())
				}

				if unitPWave.GreaterThan(decimal.Zero) {
					NminWave.P = NminWave.P.Add(decimal.NewFromInt(1))
					NminWave.Pwave = NminWave.Pwave.Add(unitPWave.Abs())
				}

				//最後一項
				if index == lastIndex {
					NminWave.CalAvgWave()
					Hkey := s + "_" + "NminWave"
					Pfiled := fmt.Sprintf("%02d%02d", referHourMin.Hour(), referHourMin.Minute()) + "PWave"
					Nfiled := fmt.Sprintf("%02d%02d", referHourMin.Hour(), referHourMin.Minute()) + "NWave"
					res0 := pipe.HSet(context.Background(), Hkey, Pfiled, NminWave.AvgPWave.String())
					logrus.Infoln("PIPLINE ADD: ", res0)
					res1 := pipe.HSet(context.Background(), Hkey, Nfiled, NminWave.AvgNWave.String())
					logrus.Infoln("PIPLINE ADD: ", res1)
				}

			}

			referHourMin = referHourMin.Add(time.Minute * time.Duration(interval))

		}

	}

	_, err := pipe.Exec(context.Background())

	if err != nil {
		logrus.Warningln("fail to execute the StaticUnitWave ")
		panic(err)
	}

	logrus.Infoln("success to execute the StaticUnitWave")

}

func (self *SymbolStatic) StaticNDayBuffer(Symbol []string, EndDate string, DayNum int, Digit string) {
	psqlCli := &PsqlClient{}
	defer psqlCli.CloseSession()
	redisCli := GetRedis("0")
	static := SymbolStatic{Symbols: Symbol, AnyDBClient: psqlCli, RedisCLi: redisCli}
	//static.StaticAllTradingTimes("2023-04-07")
	// 統計所有交易次數
	static.StaticNDayTradingTimes(EndDate, DayNum)
	for _, symbol := range Symbol {

		//// 統計n天內交易次數
		//
		//static.StaticNDayTradingTimesRank(symbol, 10, "2d")
		//// 針對不同精度單位 排序
		//
		res := static.StaticNDayTradingTimesRank(symbol, DayNum, Digit)

		vol := 0
		Wprice := decimal.Zero
		for index, _ := range res.Price {
			if index > 5 {
				break
			}

			bufferPriceRange := res.Price[index]
			bufferPrice := strings.Split(bufferPriceRange, "_")
			bufferPriceD, _ := decimal.NewFromString(bufferPrice[0])
			Wprice = Wprice.Add(bufferPriceD.Mul(decimal.NewFromInt(int64(res.TradingTimes[index]))))
			vol += res.TradingTimes[index]

		}
		sumvol := decimal.NewFromInt(int64(vol))
		avgBuffer, _ := Wprice.Div(sumvol).RoundFloor(int32(2)).Float64()
		self.RedisCLi.Set(context.Background(), fmt.Sprintf("%v_%vDAvgBuffer", symbol, DayNum), avgBuffer, time.Hour*48)
		self.RedisCLi.Set(context.Background(), fmt.Sprintf("%v_%vDAvgBufferDate", symbol, DayNum), EndDate, time.Hour*48)
		fmt.Println(fmt.Sprintf("%v_%vDAvgBuffer", symbol, DayNum), avgBuffer, EndDate)
		fmt.Println(fmt.Sprintf("%v_%vDAvgBuffer", symbol, DayNum), avgBuffer, EndDate)
	}

}

func (self *SymbolStatic) StaticOpenGap(EndDate string, DayNum int) {

	defer self.AnyDBClient.CloseSession()

	for _, s := range self.Symbols {
		pWave, p := decimal.Zero, decimal.Zero
		nWave, n := decimal.Zero, decimal.Zero
		self.AnyDBClient.GetSymbolQuery(s)
		dates := self.AnyDBClient.GetActivityDateRange(s, EndDate, DayNum)
		res := self.AnyDBClient.Get10MinHisData("U", dates[len(dates)-1], dates[0])
		for index, val := range res {

			if index == 0 {
				continue
			}

			if val.TradingTime.Day() != res[index-1].TradingTime.Day() {
				//fmt.Println(val.TradingTime.Day(), "open", val.Price)
				//fmt.Println(res[index-1].TradingTime.Day(), "close", res[index-1].Price)
				wave := val.Price.Sub(res[index-1].Price)
				fmt.Println(wave.Div(res[index-1].Price).Mul(decimal.NewFromInt(100)), "%")
				fmt.Println(val.TradingTime)
				fmt.Println("------")

				if wave.GreaterThan(decimal.Zero) {
					pWave = pWave.Add(wave.Abs().Div(res[index-1].Price).Mul(decimal.NewFromInt(100)))
					//fmt.Println("UP", wave.Abs().Div(res[index-1].Price).Mul(decimal.NewFromInt(100)))
					p = p.Add(decimal.NewFromInt(1))
				} else if wave.LessThan(decimal.Zero) {
					nWave = nWave.Add(wave.Abs().Div(res[index-1].Price).Mul(decimal.NewFromInt(100)))
					//fmt.Println("DOWN", wave.Abs().Div(res[index-1].Price).Mul(decimal.NewFromInt(100)))
					n = n.Add(decimal.NewFromInt(1))
				}

			}

			if index == len(res)-1 {
				price, _ := val.Price.Float64()
				res := self.RedisCLi.HSet(context.Background(), s+"_OpenGap", "lastPrice", price)
				fmt.Println(res)
			}

		}

		if p.GreaterThan(decimal.Zero) {
			avgWave, _ := pWave.Div(p).Float64()
			res := self.RedisCLi.HSet(context.Background(), s+"_OpenGap", "pWave", avgWave)
			expired := self.RedisCLi.Expire(context.Background(), s+"_OpenGap", time.Hour*12)
			fmt.Println(res, expired)

		}

		if n.GreaterThan(decimal.Zero) {
			avgWave, _ := nWave.Div(n).Float64()
			res := self.RedisCLi.HSet(context.Background(), s+"_OpenGap", "nWave", avgWave)
			expired := self.RedisCLi.Expire(context.Background(), s+"_OpenGap", time.Hour*12)
			fmt.Println(res, expired)
		}

	}

}
