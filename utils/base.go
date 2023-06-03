package utils

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"math"
	"strconv"
	"time"
)

type Indicator struct {
	Data []*TradingData
}

type JudgeReferData struct {
	SQ         *SQ
	IntervalSQ *HisIntervalSQ
	ReferOrder *UnitLimitOrder
}

// OrderKind: Normal,OTA ....
type ApplyOrder struct {
	Strategy     string
	OrderContent map[string]*UnitLimitOrder

	// Normal or OTA
	OrderKind      string
	DefaultContent *UnitLimitOrder
}

// 均價
func (self *Indicator) GetAvgPrice() decimal.Decimal {
	size := len(self.Data)
	sum := decimal.Zero
	for _, val := range self.Data {
		sum = sum.Add((*val).Price)
	}

	if size == 0 {
		size = 1
	}

	return sum.Div(decimal.NewFromInt(int64(size))).Round(4)
}

// 價量 加權平均價格
func (self *Indicator) GetWtAvgPrice() decimal.Decimal {

	volsum := decimal.Zero
	pvSum := decimal.Zero
	for i := 0; i < len(self.Data); i++ {
		pv := self.Data[i].Price.Mul(self.Data[i].Volume)
		pvSum = pvSum.Add(pv)
		volsum = volsum.Add(self.Data[i].Volume)
	}

	if volsum.Equal(decimal.Zero) {
		volsum = decimal.NewFromInt(1)
	}

	return pvSum.Div(volsum).Round(4)

}

// 取得該單位時間內 交易量總合
func (self *Indicator) GetUnitVolume() int {

	volsum := decimal.Zero
	for i := 0; i < len(self.Data); i++ {
		volsum = volsum.Add(self.Data[i].Volume)
	}

	return int(volsum.IntPart())
}

// 相對強弱指標  一段時間內 漲幅與跌幅絕對值相加 分之 漲幅 . 可理解成一段時間的正負波動分佈
// -1 表示 計算之後 沒有動量 例如價格持平 或是 只有一個資料 無法計算差價
func (self *Indicator) GetRSI() float64 {
	Pspread := decimal.Zero
	Nspread := decimal.Zero

	for index, _ := range self.Data {
		if index == 0 {
			continue
		}
		s := self.Data[index].Price.Sub(self.Data[index-1].Price)
		if s.Cmp(decimal.NewFromInt(0)) == 1 || s.Cmp(decimal.NewFromInt(0)) == 0 {
			Pspread = Pspread.Add(s)
		} else {
			Nspread = Nspread.Add(s)
		}
	}
	spread := Pspread.Sub(Nspread)

	if spread.Equal(decimal.Zero) {
		return -1.0
	}
	RSI, _ := (Pspread).Div(spread).Mul(decimal.NewFromInt(100)).Round(1).Float64()
	return RSI

}

type Candle struct {
	Open  *TradingData
	Close *TradingData
	High  *TradingData
	Low   *TradingData
}

func (self *Indicator) GetCandles() *Candle {
	openVal := self.Data[0]
	closeVal := self.Data[len(self.Data)-1]

	var max *TradingData
	var min *TradingData
	for index, val := range self.Data {

		if index == 0 {
			max = val
			min = val
		}

		if max.Price.LessThan(val.Price) {
			max = val
		} else if min.Price.GreaterThan(val.Price) {
			min = val
		}
	}
	//fmt.Println("open=", openVal)
	//fmt.Println("close=", closeVal)
	//fmt.Println("high=", max)
	//fmt.Println("low=", min)

	return &Candle{Open: openVal, Close: closeVal, High: max, Low: min}

}

type CMP struct {
}

// 據算兩組平均變化量 New compare to B e.g. New 100 , Old 80   if compare the value ,  return 20 , New 80 Old 100 , return -20
func (self CMP) GetGradient(New []*TradingData, Old []*TradingData) *TradingData {
	newSize := len(New)
	oldSize := len(Old)

	newData := new(TradingData)
	oldData := new(TradingData)
	avgData := new(TradingData)

	for _, val := range New {
		newData.Volume = newData.Volume.Add(val.Volume)
		newData.Price = newData.Price.Add(val.Price)
		newData.EventNum = newData.EventNum + val.EventNum
	}

	for _, val := range Old {
		oldData.Volume = oldData.Volume.Add(val.Volume)
		oldData.Price = oldData.Price.Add(val.Price)
		oldData.EventNum = oldData.EventNum + val.EventNum
	}

	if newSize == 0 {
		newSize = 1
	}
	if oldSize == 0 {
		newSize = 1
	}

	newData.Volume = newData.Volume.Div(decimal.NewFromInt(int64(newSize)))
	newData.Price = newData.Price.Div(decimal.NewFromInt(int64(newSize)))
	newData.EventNum = newData.EventNum / int32(newSize)

	oldData.Volume = oldData.Volume.Div(decimal.NewFromInt(int64(oldSize)))
	oldData.Price = oldData.Price.Div(decimal.NewFromInt(int64(oldSize)))
	oldData.EventNum = oldData.EventNum / int32(oldSize)

	avgData.Volume = newData.Volume.Sub(oldData.Volume)
	avgData.Price = newData.Price.Sub(oldData.Price)
	avgData.EventNum = newData.EventNum - oldData.EventNum
	return avgData

}

type SQ struct {
	Data              []*TradingData
	LatestData        *TradingData
	TimeAllowSec      time.Duration
	RedisCli          *redis.Client
	AllReferVolume    map[string]*int // redis 預先撈出的所有最小單位 對應時間 參考資料
	AllReferEventNum  map[string]*int
	UnitReferVolume   *int // 本次時間 利用 預先撈出最小資料計算的結果
	UnitReferEventNum *int
	UnitReferPWave    decimal.Decimal
	UnitReferNWave    decimal.Decimal
	AllReferNminWave  map[string]string
	Refer3DBuffer     decimal.Decimal
	Refer1DBuffer     decimal.Decimal
	BufferDate        string
}

func (self *SQ) init() {

	self.Data = make([]*TradingData, 0)
	self.AllReferVolume = make(map[string]*int)
	self.AllReferEventNum = make(map[string]*int)
	self.AllReferNminWave = make(map[string]string)
	if self.TimeAllowSec == 0 {
		self.TimeAllowSec = 5000 * time.Millisecond
	}

}

// 相依refer data, 需先跑一次統計相對應時間 交易量
// self,Data 必須為有序的 由小至大 且 new必須大於self.Data所有數據
func (self *SQ) Append(new *TradingData) {

	if self.Data == nil || new.TradingTime.Before(self.Data[0].TradingTime) {
		self.init()
	}

	self.Data = append(self.Data, new)

	self.LatestData = new

	new_temp := []*TradingData{}

	for index, val := range self.Data {
		if val.TradingTime.Add(self.TimeAllowSec).After(new.TradingTime) { // old time plus time duration  larger than new is True
			self.Data = append(new_temp, self.Data[index:]...)
			break
		}
	}

	//if redis exists, get reference data
	if self.RedisCli != nil {

		// 參考歷史各時段交易量
		//if len(self.AllReferEventNum) == 0 && len(self.AllReferEventNum) == 0 {
		//	self.GetReferDataFromRedis()
		//}
		//self.getReferData()

		//self.getReferWave() // 參考前幾日平均波動 這個為非對時段平均波動 而是前幾天所有每個單位波動總平均
	}

}

func (self *SQ) GetDiff() decimal.Decimal {

	if len(self.Data) < 2 {
		return decimal.Zero
	}

	return self.LatestData.Price.Sub(self.Data[0].Price)

}

// 取出歷史開場波動 , 返回wave 為波動率百分比
func (self *SQ) GetReferOpenGap(Symbol string) (lastPrice, pWave decimal.Decimal, nWave decimal.Decimal) { //

	val := self.RedisCli.HMGet(context.Background(), Symbol+"_OpenGap", "lastPrice", "pWave", "nWave")

	lastPrice, err := decimal.NewFromString(val.Val()[0].(string))

	if err != nil {
		logrus.Warningln(err)
	}

	pWave, err = decimal.NewFromString(val.Val()[1].(string))

	if err != nil {
		logrus.Warningln(err)
	}

	nWave, err = decimal.NewFromString(val.Val()[2].(string))

	if err != nil {
		logrus.Warningln(err)
	}

	return lastPrice, pWave, nWave

}

// return 3D && 10D avg TradingBuffer
func (self *SQ) GetReferBuffer(Date string) (Buffer3DAVG decimal.Decimal, Buffer1DAVG decimal.Decimal) {
	if len(self.Data) < 1 {
		return decimal.Zero, decimal.Zero
	}

	symbol := self.Data[0].Symbol

	if self.BufferDate == "" {
		self.BufferDate, _ = self.RedisCli.Get(context.Background(), symbol+"_1DAvgBufferDate").Result()
		buffer3D, _ := self.RedisCli.Get(context.Background(), symbol+"_3DAvgBuffer").Result()
		self.Refer3DBuffer, _ = decimal.NewFromString(buffer3D)
		buffer1D, _ := self.RedisCli.Get(context.Background(), symbol+"_1DAvgBuffer").Result()
		self.Refer1DBuffer, _ = decimal.NewFromString(buffer1D)
		self.BufferDate, _ = self.RedisCli.Get(context.Background(), symbol+"_1DAvgBufferDate").Result()
	}

	if Date != self.BufferDate {
		s := []string{symbol}
		psqlCli := &PsqlClient{}
		static := SymbolStatic{Symbols: s, AnyDBClient: psqlCli, RedisCLi: self.RedisCli}
		defer static.AnyDBClient.CloseSession()
		static.StaticNDayBuffer(s, Date, 3, "1f")

		static.StaticNDayBuffer(s, Date, 1, "1f")

		buffer3D, _ := self.RedisCli.Get(context.Background(), symbol+"_3DAvgBuffer").Result()
		self.Refer3DBuffer, _ = decimal.NewFromString(buffer3D)
		buffer5D, _ := self.RedisCli.Get(context.Background(), symbol+"_1DAvgBuffer").Result()
		self.Refer1DBuffer, _ = decimal.NewFromString(buffer5D)

		self.BufferDate, _ = self.RedisCli.Get(context.Background(), symbol+"_1DAvgBufferDate").Result()

	}

	return self.Refer3DBuffer, self.Refer1DBuffer

}

// 一次將所有redis參考統計資料 (各時段交易量 交易筆數) 取出
func (self *SQ) GetReferDataFromRedis() {
	ctx := context.Background()

	symobl := self.Data[0].Symbol
	baseTime := time.Date(self.Data[0].TradingTime.Year(), self.Data[0].TradingTime.Month(), self.Data[0].TradingTime.Day(), 4, 0, 0, 0, self.Data[0].TradingTime.Location())
	pipeline := self.RedisCli.Pipeline()
	for i := 0; i < 11520; i++ {
		timediff := i * 5
		t := baseTime.Add(time.Duration(timediff) * time.Second)
		key := symobl + "_" + t.Format("150405")
		pipeline.HMGet(ctx, key, "Vol", "EventNum")

	}

	Cmds, err := pipeline.Exec(ctx)

	if err != nil {
		logrus.Infoln(err)
	}

	for _, cmd := range Cmds {
		cmd, ok := cmd.(*redis.SliceCmd)
		if !ok {
			logrus.Infoln(err)
		}

		args := cmd.Args()
		key, ok := args[1].(string)

		if !ok {
			logrus.Infoln(err)
		}

		val, err := cmd.Result()

		if err != nil {
			logrus.Infoln(err)
		}

		vol, _ := strconv.ParseInt(val[0].(string), 10, 64)
		eventNum, _ := strconv.ParseInt(val[1].(string), 10, 64)

		self.AllReferVolume[key] = new(int)
		newVol := int(vol)
		self.AllReferVolume[key] = &newVol

		self.AllReferEventNum[key] = new(int)
		newEventNum := int(eventNum)
		self.AllReferEventNum[key] = &newEventNum

	}

}

// 從預先取出的參考資料(交易量,交易事件數量) 取得該時段相對應 歷史交易量與交易事件數量
func (self *SQ) getReferData() {

	if len(self.Data) > 0 {
		keys := make([]string, 0)
		// 若為15秒為一次間隔資料 ,會用單位資料計算出(redis單位統計資料為5秒一次)
		// slice > 093001 093002 - 093015.99 , 此時會找  093000 093005 093010 的資料統計 (093000 5秒資料也是用 e.g. 093001 093002 ... ~ 093004.999 合併出來的)
		// e.g. 093000 > 會找 093000 093005 093010 三筆資料 合併 , 這邊是找宣告 09300 093005 093010 存放位置

		symobl := self.Data[0].Symbol
		sec1 := math.Round((float64(self.Data[0].TradingTime.Second()) / 5)) * 5
		baseTime := time.Date(self.Data[0].TradingTime.Year(), self.Data[0].TradingTime.Month(), self.Data[0].TradingTime.Day(), self.Data[0].TradingTime.Hour(), self.Data[0].TradingTime.Minute(), int(sec1), 0, self.Data[0].TradingTime.Location())

		for i := 0; i < int(self.TimeAllowSec.Seconds()/5); i++ {
			timediff := i * 5
			t := baseTime.Add(time.Duration(timediff) * time.Second)
			key := symobl + "_" + t.Format("150405")
			keys = append(keys, key)

		}

		sumVol := new(int)
		sumEventNum := new(int)
		for _, key := range keys {
			if _, ok := self.AllReferVolume[key]; ok {
				vol := *self.AllReferVolume[key]
				*sumVol += vol
			}
			if _, ok := self.AllReferEventNum[key]; ok {
				eventNum := *self.AllReferEventNum[key]
				*sumEventNum += eventNum
			}
		}
		self.UnitReferVolume = sumVol
		self.UnitReferEventNum = sumEventNum

	}

}

func (self *SQ) getReferWave() {
	if self.UnitReferNWave.Equal(decimal.Zero) && self.UnitReferPWave.Equal(decimal.Zero) {
		ctx := context.Background()
		key := self.Data[0].Symbol + "_Wave"
		cmds, _ := self.RedisCli.HMGet(ctx, key, "pwave", "nwave").Result()
		self.UnitReferPWave, _ = decimal.NewFromString(cmds[0].(string))
		self.UnitReferNWave, _ = decimal.NewFromString(cmds[1].(string))
	}
}

func (self *SQ) GetAllReferNMinWaveFromRedis() {

	Symbol := self.Data[0].Symbol
	hkey := Symbol + "_" + "NminWave"
	res, err := self.RedisCli.HGetAll(context.Background(), hkey).Result()
	if err != nil {
		logrus.Warningln(err)
	}
	self.AllReferNminWave = res
}

// 取得歷史各時間 平均波動幅度,    返回百分比 0.000XXX %
func (self *SQ) GetUnitReferNMinWave(interval int) (PNMinWave decimal.Decimal, NNMinWave decimal.Decimal) {

	if len(self.AllReferNminWave) == 0 {
		self.GetAllReferNMinWaveFromRedis()
	}

	data := self.LatestData
	//now := time.Now()
	//datatime := time.Date(now.Year(), now.Month(), now.Day(), 9, 32, 0, 0, Loc)
	//data.TradingTime = datatime

	referMin := interval * (data.TradingTime.Minute() / 5)

	referTime := time.Date(data.TradingTime.Year(), data.TradingTime.Month(), data.TradingTime.Day(), data.TradingTime.Hour(), referMin, 0, 0, Loc).Format("1504")

	Pkey := referTime + "PWave"
	Nkey := referTime + "NWave"
	PvalD := decimal.Zero
	NvalD := decimal.Zero

	Pval, ok := self.AllReferNminWave[Pkey]
	if !ok {
		logrus.Infoln("NMinWave NO EXISTS")
	} else {
		PvalD, _ = decimal.NewFromString(Pval)
	}

	Nval, ok := self.AllReferNminWave[Nkey]
	if !ok {
		logrus.Infoln("NMinWave NO EXISTS")
	} else {
		NvalD, _ = decimal.NewFromString(Nval)
	}

	PNMinWave = PvalD.Div(data.Price)
	NNMinWave = NvalD.Div(data.Price)

	return PNMinWave, NNMinWave

}

type UnitChange struct {
	Price      decimal.Decimal
	CreateTime time.Time
}

func (self *SQ) GetPriceChange() *UnitChange {

	if len(self.Data) < 2 {
		return nil
	}

	return &UnitChange{Price: self.LatestData.Price.Sub(self.Data[0].Price), CreateTime: self.LatestData.TradingTime}

}

type HisSQ struct {
	SQSet        []SQ
	TimeAllowSec int
}

func (self *HisSQ) Append(NewSQ SQ) {

	// 初始化
	if self.SQSet == nil {
		self.SQSet = make([]SQ, 0)
	}

	// Size 小於 2, 無法計算
	if len(self.SQSet) < 1 {
		self.SQSet = append(self.SQSet, NewSQ)
		return
	}

	if (self.SQSet[0].Data[0].TradingTime.Add(time.Second * time.Duration(self.TimeAllowSec))).After(NewSQ.Data[len(NewSQ.Data)-1].TradingTime) {

		self.SQSet = append(self.SQSet, NewSQ)

	} else {

		index := 0
		for {

			// 數組有序 從最舊的資料開始找 是否逾期 找到逾期的最新資料 index,
			if (self.SQSet[index].Data[0].TradingTime.Add(time.Second * time.Duration(self.TimeAllowSec))).Before(NewSQ.Data[len(NewSQ.Data)-1].TradingTime) {
				index += 1

				// 若全部逾期 只輸最新的
				if index == len(self.SQSet) {
					self.SQSet = []SQ{NewSQ}
					return
				}

				continue
			}

			// 之後將該index後面的與最新資料合併
			self.SQSet = append(self.SQSet[index:], NewSQ)
			//fmt.Printf("APPEND %+v\n", NewChange)
			break
		}

	}

}

func HttpSuccess(httpStatusCode int) bool {

	if httpStatusCode >= 200 && httpStatusCode < 300 {
		return true
	}

	return false

}

// 保存TimeAllowSec 時間長度的SQ, 且每個SQ間隔至少intervalSec 秒
type HisIntervalSQ struct {
	SQSet        []SQ
	IntervalSec  int
	TimeAllowSec int
}

// 用sq 結尾做比較
func (self *HisIntervalSQ) Append(sq SQ) {

	if len(self.SQSet) == 0 {
		self.SQSet = append(self.SQSet, sq)
		return
	}

	for index, val := range self.SQSet {

		if sq.LatestData.TradingTime.Before(val.LatestData.TradingTime.Add(time.Second * time.Duration(self.TimeAllowSec))) {
			self.SQSet = self.SQSet[index:]
			if sq.LatestData.TradingTime.After(self.SQSet[len(self.SQSet)-1].LatestData.TradingTime.Add(time.Duration(self.IntervalSec) * time.Second)) {
				self.SQSet = append(self.SQSet, sq)
			}
			break
		}

	}

	//for _, val := range self.SQSet {
	//	fmt.Println(val.LatestData.TradingTime)
	//}
	//fmt.Println("----")

}
