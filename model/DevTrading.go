package model

import (
	"GoBot/utils"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type TradingTest struct {
	Symbol          string
	Budget          decimal.Decimal
	Quantity        int
	SQ              *utils.SQ
	IntervalSQ      *utils.HisIntervalSQ
	TradingLock     bool // 目前 用於一些交易意外 鎖住下單功能 or 蓄意停止交易能 , 	// 需滿滿足 //	GetID() string, GetOrderKind() string, OpenJudge(sq *utils.SQ) *utils.ApplyOrder, CloseJudge(sq *utils.SQ) *utils.ApplyOrder,
	DayTrade        bool
	OrderStatusTest *OrderStatusTest
	// 與正式版最大差異為 , 正式版API操作 與 狀態分開更新 (狀態控制主要為 redis 因此當初設計分開,希望程式崩潰 依然還能紀錄狀態)
	// 測試環境上 模擬API掛單 與 狀態控制 共用 , 狀態直接用 struct Field 控制
	TradingDate time.Time // 用來紀錄交易日 若偵測到新資料為不同日 則重制交易
}

// 重製預設值
func (self *TradingTest) Init() {
	self.Symbol = ""
	self.Budget = decimal.Zero
	self.Quantity = 0
	self.SQ = new(utils.SQ)
	self.IntervalSQ = new(utils.HisIntervalSQ)
	self.TradingLock = false
	self.OrderStatusTest = new(OrderStatusTest)
	// 初始化 orderStatus
	if self.DayTrade == false {
		logrus.Infof("coear OrderStatus")
		self.OrderStatusTest.ClearOrderStatus()
	}

}

// SQ存在 Qantity 尚未有數值 進行數量評估
func (self *TradingTest) GetQuantity() {

	if self.Quantity == 0 && len(self.SQ.Data) > 0 {

		self.Quantity = int(self.Budget.Div(self.SQ.Data[0].Price).IntPart())

		logrus.Infof("Set %v Quantity is %v", self.Symbol, self.Quantity)

	}

}

func (self *TradingTest) InspectTradingTime(closeAfterTime time.Time, NowTime time.Time) {

	if NowTime.After(closeAfterTime) {

		self.TradingLock = true
	}

}

func (self *TradingTest) InspectOrderOperationLimit() {

	if self.OrderStatusTest.Put_POST_DELETE.GreaterThan(decimal.NewFromInt(375)) {

		self.TradingLock = true

	}

}

func (self *TradingTest) OpenJudge(Strategy StrategyInterface) *utils.ApplyOrder {
	referData := utils.JudgeReferData{SQ: self.SQ, IntervalSQ: self.IntervalSQ}
	applyOrder := Strategy.GetOpenJudgement(&referData)
	if self.TradingLock == true { // 用於通訊意外 鎖住交易功能
		applyOrder = nil
	}

	return applyOrder

}

func (self *TradingTest) CloseJudge(Strategy StrategyInterface) *utils.ApplyOrder {
	open := self.OrderStatusTest.OpenOrder
	referData := utils.JudgeReferData{SQ: self.SQ, ReferOrder: open, IntervalSQ: self.IntervalSQ}
	applyOrder := Strategy.GetCloseJudgement(&referData)

	if self.TradingLock == true { // 用於通訊意外 鎖住交易功能
		applyOrder = nil
	}

	return applyOrder
}

func (self *TradingTest) OrderProcess(applyOrder *utils.ApplyOrder) *utils.ApplyOrder {

	if applyOrder.OrderKind == "OTA" {

		logrus.Infof("Open: Price:%v Type:%v ", applyOrder.OrderContent["OPEN"].Price, applyOrder.OrderContent["OPEN"].OrderType)
		applyOrder.OrderContent["OPEN"].Quantity = self.Quantity
		//applyOrder.OrderContent["OPEN"].CreateTime = self.SQ.LatestData.TradingTime
		// 加上workingOrder createtime

		applyOrder.OrderContent["CLOSE"].Quantity = self.Quantity
		logrus.Infof("CLOSE: Price:%v Type:%v ", applyOrder.OrderContent["CLOSE"].Price, applyOrder.OrderContent["CLOSE"].OrderType)
	} else if applyOrder.OrderKind == "NORMAL" || applyOrder.OrderKind == "Normal" {

		applyOrder.DefaultContent.Quantity = self.Quantity

		if applyOrder.DefaultContent.OrderType == 10 {
			logrus.Infof("OPEN: Price:%v Type:%v ", applyOrder.DefaultContent.Price, applyOrder.DefaultContent.OrderType)
		} else if applyOrder.DefaultContent.OrderType == -10 {
			logrus.Infof("CLOSE: Price:%v Type:%v ", applyOrder.DefaultContent.Price, applyOrder.DefaultContent.OrderType)
		}

		if applyOrder.DefaultContent.OrderType == 20 {
			logrus.Infof("OPEN: Price:%v Type:%v ", applyOrder.DefaultContent.Price, applyOrder.DefaultContent.OrderType)
		} else if applyOrder.DefaultContent.OrderType == -20 {
			logrus.Infof("CLOSE: Price:%v Type:%v ", applyOrder.DefaultContent.Price, applyOrder.DefaultContent.OrderType)
		}

		//applyOrder.DefaultContent.CreateTime = self.SQ.LatestData.TradingTime
		// 加上workingOrder createtime
	}

	return applyOrder

}

type BackTestPerformance struct {
	Symbol            string
	DayEvent          decimal.Decimal
	WinDay            decimal.Decimal
	AllDayPerformance map[string]decimal.Decimal
	WinRate           decimal.Decimal
	DayEarnsSum       decimal.Decimal // 每日實際收入 總和
	DayBudget         decimal.Decimal // 單日預算
	TotalEvent        decimal.Decimal
	TotalWin          decimal.Decimal
	Put_POST_DELETE   decimal.Decimal
	TotalDay          decimal.Decimal
	Date              string
}

// earns 當日實際收入
func (self *BackTestPerformance) Append(Date string, earns decimal.Decimal) {
	if self.Date == "" {
		self.Date = Date
	}

	self.AllDayPerformance[Date] = decimal.NewFromInt(100).Mul(earns.Div(self.DayBudget)) //百分比表示績效
	self.DayEvent = self.DayEvent.Add(decimal.NewFromInt(1))
	self.DayEarnsSum = self.DayEarnsSum.Add(earns)
	if earns.GreaterThanOrEqual(decimal.Zero) {
		self.WinDay = self.WinDay.Add(decimal.NewFromInt(1))
	}

	self.WinRate = decimal.NewFromInt(100).Mul(self.WinDay.Div(self.DayEvent))
}

func (self *BackTestPerformance) Report() {

	PeriodPerformance := decimal.Zero
	fmt.Printf("%v REPORT#################\n", self.Symbol)
	for date, p := range self.AllDayPerformance {

		fmt.Printf("Date: %v Perforomace: %v Ratio: %v %%\n", date, p.Mul(self.DayBudget).Div(decimal.NewFromInt(100)), p)

		PeriodPerformance = PeriodPerformance.Add(p)
	}

	PeriodPerformance = PeriodPerformance.Div(decimal.NewFromInt(int64(len(self.AllDayPerformance))))

	fmt.Println("")
	fmt.Printf("POST/PUT/DELETE: %v / DAY\n", self.Put_POST_DELETE.DivRound(self.TotalDay, int32(2)))
	fmt.Printf("Period Win Rate: %v / %v = %v %%\n", self.WinDay, self.DayEvent, self.WinRate)
	if !self.TotalEvent.Equal(decimal.Zero) {
		fmt.Printf("ALl Trade WinRate: %v / %v = %v %%\n", self.TotalWin, self.TotalEvent, decimal.NewFromInt(100).Mul(self.TotalWin.Div(self.TotalEvent)))
	}
	fmt.Printf("Average Period Performance: %v %%\n", PeriodPerformance)
	fmt.Printf("Actual earns: %v / %v = %v %%\n", self.DayEarnsSum, self.DayBudget, decimal.NewFromInt(100).Mul(self.DayEarnsSum.Div(self.DayBudget)))
}

type OrderStatusTest struct {
	Symbol          string
	OpenOrder       *utils.UnitLimitOrder
	CloseOrder      *utils.UnitLimitOrder
	OTA             bool
	Stage           int
	Earns           []decimal.Decimal
	Win             decimal.Decimal
	Event           decimal.Decimal
	SumEarns        decimal.Decimal
	TimeOutSec      int // 逾期設定 秒 , N 秒後 order expired
	Put_POST_DELETE decimal.Decimal
}

func (self *OrderStatusTest) CreateOrder(applyOrder *utils.ApplyOrder) {

	if self.OpenOrder == nil {
		self.OpenOrder = new(utils.UnitLimitOrder)
	}

	if self.CloseOrder == nil {
		self.CloseOrder = new(utils.UnitLimitOrder)
	}

	if applyOrder.OrderKind == "OTA" {
		self.Put_POST_DELETE = self.Put_POST_DELETE.Add(decimal.NewFromInt(2))
		self.OTA = true
		self.Symbol = applyOrder.OrderContent["OPEN"].Symbol

		self.OpenOrder.Symbol = applyOrder.OrderContent["OPEN"].Symbol
		self.OpenOrder.Price = applyOrder.OrderContent["OPEN"].Price
		self.OpenOrder.Quantity = applyOrder.OrderContent["OPEN"].Quantity
		self.OpenOrder.OrderType = applyOrder.OrderContent["OPEN"].OrderType
		self.OpenOrder.CreateTime = applyOrder.OrderContent["OPEN"].CreateTime
		self.OpenOrder.Status = "WORKING"

		self.CloseOrder.Symbol = applyOrder.OrderContent["CLOSE"].Symbol
		self.CloseOrder.Price = applyOrder.OrderContent["CLOSE"].Price
		self.CloseOrder.Quantity = applyOrder.OrderContent["CLOSE"].Quantity
		self.CloseOrder.OrderType = applyOrder.OrderContent["CLOSE"].OrderType
		self.CloseOrder.CreateTime = applyOrder.OrderContent["CLOSE"].CreateTime
		self.CloseOrder.Status = "WAITING"

		self.Stage = 1

		logrus.Infoln("Apply Full OTA Order")
		logrus.Infof("put/post/delete: %v", self.Put_POST_DELETE)

	} else if applyOrder.OrderKind == "NORMAL" || applyOrder.OrderKind == "Normal" {
		self.Put_POST_DELETE = self.Put_POST_DELETE.Add(decimal.NewFromInt(1))
		self.OTA = false

		// 全新
		if self.Stage == 0 {

			self.Symbol = applyOrder.DefaultContent.Symbol
			self.OpenOrder.Symbol = applyOrder.DefaultContent.Symbol
			self.OpenOrder.Price = applyOrder.DefaultContent.Price
			self.OpenOrder.Quantity = applyOrder.DefaultContent.Quantity
			self.OpenOrder.OrderType = applyOrder.DefaultContent.OrderType
			self.OpenOrder.Status = "WORKING"

			self.Stage = 1

			// 若為working 紀錄時間 需用來判斷time out
			self.OpenOrder.CreateTime = applyOrder.DefaultContent.CreateTime
			logrus.Infoln("create Normal Open Order")
			logrus.Infof("put/post/delete: %v", self.Put_POST_DELETE)
		}

		// 已經FILLED 尚未開新的

		if self.Stage == 2 && self.OpenOrder.Status == "FILLED" {

			self.Symbol = applyOrder.DefaultContent.Symbol
			self.CloseOrder.Symbol = applyOrder.DefaultContent.Symbol
			self.CloseOrder.Price = applyOrder.DefaultContent.Price
			self.CloseOrder.Quantity = applyOrder.DefaultContent.Quantity
			self.CloseOrder.OrderType = applyOrder.DefaultContent.OrderType
			self.CloseOrder.Status = "WORKING"

			// 若為working 紀錄時間 需用來判斷time out
			self.CloseOrder.CreateTime = applyOrder.DefaultContent.CreateTime
			logrus.Infof("create Normal Close Order %+v", self.CloseOrder)
			logrus.Infof("put/post/delete: %v", self.Put_POST_DELETE)
		}

	}

}

func (self *OrderStatusTest) ClearOrderStatus() {
	self.Symbol = ""
	self.OpenOrder = new(utils.UnitLimitOrder)
	self.CloseOrder = new(utils.UnitLimitOrder)
	self.OTA = false
	self.Stage = 0
	self.Put_POST_DELETE = self.Put_POST_DELETE.Add(decimal.NewFromInt(1))
	logrus.Infof("put/post/delete: %v", self.Put_POST_DELETE)
}

// 存在 返回 working order, 不存在 nil
func (self *OrderStatusTest) GetWorkingOrder() (workingOrder *utils.UnitLimitOrder) {

	if self.Stage == 1 {

		return self.OpenOrder

	} else if self.Stage == 2 {

		// 若stage = 2,  closeOrder 已經存在
		if self.CloseOrder.Status != "" {

			return self.CloseOrder

		}

		// 不存在返回 nil

	}

	// 若為 0 也返回 nil

	return nil

}

// 更換close order , 使用前 需注意需先更新 create time
func (self *OrderStatusTest) ReplaceOpenOrder(ReplaceOrder *utils.UnitLimitOrder) {

	if self.Stage == 1 {
		self.Put_POST_DELETE = self.Put_POST_DELETE.Add(decimal.NewFromInt(1))
		logrus.Infof("put/post/delete: %v", self.Put_POST_DELETE)
		self.OpenOrder.Symbol = ReplaceOrder.Symbol
		self.OpenOrder.Price = ReplaceOrder.Price
		self.OpenOrder.Quantity = ReplaceOrder.Quantity
		self.OpenOrder.OrderType = ReplaceOrder.OrderType
		self.OpenOrder.Status = "WORKING"

		// 若為working 紀錄時間 需用來判斷time out
		self.OpenOrder.CreateTime = ReplaceOrder.CreateTime
		self.CloseOrder.CreateTime = ReplaceOrder.CreateTime

		return
	}

	// 只有stage 2 才可以更換 closeOrder
	panic("NO IN STAGE 2")

}

// 更換close order , 使用前 需注意需先更新 create time
func (self *OrderStatusTest) ReplaceCloseOrder(ReplaceOrder *utils.UnitLimitOrder) {

	if self.Stage == 2 {
		self.Put_POST_DELETE = self.Put_POST_DELETE.Add(decimal.NewFromInt(1))
		logrus.Infof("put/post/delete: %v", self.Put_POST_DELETE)
		self.CloseOrder.Symbol = ReplaceOrder.Symbol
		self.CloseOrder.Price = ReplaceOrder.Price
		self.CloseOrder.Quantity = ReplaceOrder.Quantity
		self.CloseOrder.OrderType = ReplaceOrder.OrderType
		self.CloseOrder.Status = "WORKING"

		// 若為working 紀錄時間 需用來判斷time out
		self.CloseOrder.CreateTime = ReplaceOrder.CreateTime

		return
	}

	// 只有stage 2 才可以更換 closeOrder
	panic("NO IN STAGE 2")

}

// 更新order掛單 成交狀態
// FILLED 後清除所有狀態 從後續處理 (這邊為了模擬正式環境)
func (self *OrderStatusTest) UpdateOrderStatus(NewData *utils.TradingData) {

	if self.Stage == 0 {
		return
	}

	if self.Stage == 1 {

		orderPrice, _ := decimal.NewFromString(self.OpenOrder.Price)

		// 成交價差
		fix := NewData.Price.Mul(decimal.NewFromFloat(0.0001))
		if fix.LessThan(decimal.NewFromFloat(0.01)) {
			fix = decimal.NewFromFloat(0.01)
		}
		// long
		// OpenOrder 新進價格小於掛單 成交 你想買便宜貨... 要有人賣更低價 才會成交
		if self.OpenOrder.OrderType == 10 && NewData.Price.LessThanOrEqual(orderPrice.Sub(fix)) && NewData.Volume.GreaterThanOrEqual(decimal.NewFromInt(100)) && NewData.TradingTime.After(self.OpenOrder.CreateTime.Add(time.Millisecond*200)) { // 近來價格需比order多0.01

			// 防止奇異點價格 小成交量 價差偏離
			if NewData.Price.Sub(orderPrice).Div(orderPrice).Mul(decimal.NewFromInt(100)).Abs().GreaterThan(decimal.NewFromFloat32(0.5)) {
				return
			}

			logrus.Infof("FILLED OPEN ORDER (%v, new: %v) at %v", self.OpenOrder.Price, NewData.Price, NewData.TradingTime)

			self.OpenOrder.Status = "FILLED"
			self.Stage = 2 // openOrder filled 後 stage狀態改變
			// 若為OTA 開啟ClOSE
			if self.OTA == true {
				self.CloseOrder.Status = "WORKING"

				// closeOrder 開啟 更新時間
				//self.CloseOrder.CreateTime = NewData.TradingTime
			}

			// short
			// OpenOrder 新進價格小於掛單 成交 你想賣... 要有人買更高 才會成交
		} else if self.OpenOrder.OrderType == 20 && NewData.Price.GreaterThanOrEqual(orderPrice.Add(fix)) && NewData.Volume.GreaterThanOrEqual(decimal.NewFromInt(100)) && NewData.TradingTime.After(self.OpenOrder.CreateTime.Add(time.Millisecond*200)) { // 近來價格需比order多0.01

			// 防止奇異點價格 小成交量 價差偏離
			if NewData.Price.Sub(orderPrice).Div(orderPrice).Mul(decimal.NewFromInt(100)).Abs().GreaterThan(decimal.NewFromFloat32(0.5)) {
				return
			}

			logrus.Infof("FILLED OPEN ORDER (%v, new: %v) at %v", self.OpenOrder.Price, NewData.Price, NewData.TradingTime)

			self.OpenOrder.Status = "FILLED"
			self.Stage = 2 // openOrder filled 後 stage狀態改變
			// 若為OTA 開啟ClOSE
			if self.OTA == true {
				self.CloseOrder.Status = "WORKING"

				// closeOrder 開啟 更新時間
				//self.CloseOrder.CreateTime = NewData.TradingTime
			}
		}

		// CloseOrder存在 && CloseOrder 新進價格大於掛單 成交
	} else if self.Stage == 2 && self.CloseOrder.Symbol != "" {

		// 成交價差
		fix := NewData.Price.Mul(decimal.NewFromFloat(0.0001))
		if fix.LessThan(decimal.NewFromFloat(0.01)) {
			fix = decimal.NewFromFloat(0.01)
		}

		orderPrice, _ := decimal.NewFromString(self.CloseOrder.Price)

		// 做多 新進價格小於 於掛單 成交 ,  你想賣高價... 要有人想當盤子買高價 才會成交
		if self.CloseOrder.OrderType == -10 && NewData.Price.GreaterThanOrEqual(orderPrice.Add(fix)) && NewData.Volume.GreaterThanOrEqual(decimal.NewFromInt(100)) && NewData.TradingTime.After(self.CloseOrder.CreateTime.Add(time.Millisecond*200)) {

			// 防止奇異點價格 小成交量 價差偏離
			if NewData.Price.Sub(orderPrice).Div(orderPrice).Mul(decimal.NewFromInt(100)).Abs().GreaterThan(decimal.NewFromFloat32(0.5)) {
				return
			}

			logrus.Infof("FILLED CLOSE ORDER (%v, new: %v) at %v", self.CloseOrder.Price, NewData.Price, NewData.TradingTime)

			self.CloseOrder.Status = "FILLED"
			self.Stage = 0 // closeOrder filled 後 stage狀態改變
			open_, _ := decimal.NewFromString(self.OpenOrder.Price)
			close_, _ := decimal.NewFromString(self.CloseOrder.Price)
			num := decimal.NewFromInt(int64(self.CloseOrder.Quantity))

			earns := (close_.Sub(open_)).Mul(num)
			logrus.Infof("(%v - %v) X %v = %v ", close_, open_, num, earns)
			self.Earns = append(self.Earns, earns)
			self.SumEarns = self.SumEarns.Add(earns)

			// 統計勝率
			self.Event = self.Event.Add(decimal.NewFromInt(1))
			if earns.GreaterThan(decimal.Zero) {
				self.Win = self.Win.Add(decimal.NewFromInt(1))
			}

		}

		// 做空 新進價格小於掛單 成交 , 你想買便宜獲... 要有人賣更低價 才會成交
		if self.CloseOrder.OrderType == -20 && NewData.Price.LessThanOrEqual(orderPrice.Sub(fix)) && NewData.Volume.GreaterThanOrEqual(decimal.NewFromInt(100)) && NewData.TradingTime.After(self.CloseOrder.CreateTime.Add(time.Millisecond*200)) {

			// 防止奇異點價格 小成交量 價差偏離
			if NewData.Price.Sub(orderPrice).Div(orderPrice).Mul(decimal.NewFromInt(100)).Abs().GreaterThan(decimal.NewFromFloat32(0.5)) {
				return
			}

			logrus.Infof("FILLED CLOSE ORDER (%v, new: %v) at %v", self.CloseOrder.Price, NewData.Price, NewData.TradingTime)

			self.CloseOrder.Status = "FILLED"
			self.Stage = 0 // closeOrder filled 後 stage狀態改變
			open_, _ := decimal.NewFromString(self.OpenOrder.Price)
			close_, _ := decimal.NewFromString(self.CloseOrder.Price)
			num := decimal.NewFromInt(int64(self.CloseOrder.Quantity)).Mul(decimal.NewFromFloat(-1))

			earns := (close_.Sub(open_)).Mul(num)
			logrus.Infof("(%v - %v) X %v = %v ", close_, open_, num, earns)
			self.Earns = append(self.Earns, earns)
			self.SumEarns = self.SumEarns.Add(earns)

			// 統計勝率
			self.Event = self.Event.Add(decimal.NewFromInt(1))
			if earns.GreaterThan(decimal.Zero) {
				self.Win = self.Win.Add(decimal.NewFromInt(1))
			}

		}

	}

}

// 過期返回 過期 order, 沒過期返回 nil
func (self *OrderStatusTest) InspectTimeOut(NowTime time.Time) (expiredOrder *utils.UnitLimitOrder) {

	// timeOutSec == 0 , 表示無逾期時間
	if self.TimeOutSec == 0 {
		logrus.Infof("No EPXIRED TIME")
		return nil
	}

	workingOrder := self.GetWorkingOrder()

	if workingOrder != nil {

		// open
		if workingOrder.CreateTime.Add(time.Duration(self.TimeOutSec) * time.Second).Before(NowTime) {

			logrus.Infof("Expired Order %+v, Stage %v", workingOrder, self.Stage)
			logrus.Infof("expired Time at %v", NowTime)

			return workingOrder

		}

	}

	return nil

}

// data從channel 給
func DevTrading(MyStrategy StrategyInterface, PerformanceOutput chan *BackTestPerformance, SQAllowSec int, OrderExpired int, Symbol string, Budget decimal.Decimal, DataChan chan *utils.TradingData, wg *sync.WaitGroup) {
	defer wg.Done()

	MyTrading := &TradingTest{}

	OrderExpiredSec := OrderExpired

	staticPerformance := BackTestPerformance{Symbol: Symbol, DayBudget: Budget}
	staticPerformance.AllDayPerformance = make(map[string]decimal.Decimal)

	temp := decimal.Zero

	for {

		data, ok := <-DataChan

		// 確認該channel 是否還存在
		if ok == false {
			break
		}

		// 資料近來為不同交易日 重新初始化
		if !MyTrading.DayTrade && (MyTrading.TradingDate.IsZero() || MyTrading.TradingDate.Day() != data.TradingTime.Day()) {

			if !MyTrading.TradingDate.IsZero() {
				// 到這個階段 data已經為隔日資料 我們這邊目前是存著上一次的績效, 需使用sq最後一次的日期 (之後就會被初始化 才是新日期的開始)
				staticPerformance.TotalDay = staticPerformance.TotalDay.Add(decimal.NewFromInt(1))
				staticPerformance.Put_POST_DELETE = staticPerformance.Put_POST_DELETE.Add(MyTrading.OrderStatusTest.Put_POST_DELETE)
				staticPerformance.TotalWin = staticPerformance.TotalWin.Add(MyTrading.OrderStatusTest.Win)
				staticPerformance.TotalEvent = staticPerformance.TotalEvent.Add(MyTrading.OrderStatusTest.Event)
				staticPerformance.Append(MyTrading.SQ.LatestData.TradingTime.Format("2006-01-02"), (MyTrading.OrderStatusTest.SumEarns))
			}

			MyTrading.Init()
			MyTrading.Symbol = Symbol
			MyTrading.Budget = Budget
			MyTrading.DayTrade = MyStrategy.GetDayTrade()
			MyTrading.SQ = &utils.SQ{RedisCli: utils.GetRedis("0"), TimeAllowSec: time.Second * time.Duration(SQAllowSec)}
			MyTrading.IntervalSQ = &utils.HisIntervalSQ{SQSet: make([]utils.SQ, 0), TimeAllowSec: 900, IntervalSec: 180}
			MyTrading.TradingDate = data.TradingTime
			MyTrading.OrderStatusTest.TimeOutSec = OrderExpiredSec

		}

		MyTrading.DayTrade = MyStrategy.GetDayTrade()
		MyTrading.TradingLock = false
		// 日內交易設定 收盤尾端直接鎖住
		today := data.TradingTime

		// 1200 or 1300 // 11, 30, 0, 0
		MyTrading.InspectTradingTime(time.Date(today.Year(), today.Month(), today.Day(), 11, 30, 0, 0, utils.Loc), data.TradingTime)
		MyTrading.InspectOrderOperationLimit()

		// 盤前不交易
		if data.FormT == true {
			continue
		}
		// 這邊跳過前三十分鐘 // 9, 30, 0, 0,

		if data.TradingTime.In(utils.Loc).Before(time.Date(today.Year(), today.Month(), today.Day(), 9, 30, 0, 0, utils.Loc)) {
			continue
		}

		// 防止小量 價格奇異點
		if temp.IsZero() {
			temp = data.Price
		} else if (temp.Sub(data.Price).Div(data.Price).Mul(decimal.NewFromInt(100))).Abs().GreaterThan(decimal.NewFromFloat(0.3)) && data.Volume.LessThan(decimal.NewFromInt(1000)) {
			//fmt.Println("===")
			//fmt.Println(data.TradingTime, data.Volume)
			//fmt.Println("last", temp, "new", data.Price)
			continue
		} else {
			temp = data.Price
		}

		MyTrading.SQ.Append(data)
		MyTrading.IntervalSQ.Append(*MyTrading.SQ)

		MyTrading.GetQuantity()
		// 此次交易 購買數量初始化

		MyTrading.OrderStatusTest.UpdateOrderStatus(data)
		// 檢查是否有order正在處理
		if MyTrading.OrderStatusTest.OpenOrder.Symbol != "" {

			// 查看是否有正在等待成交的order
			workingOrder := MyTrading.OrderStatusTest.GetWorkingOrder()

			if workingOrder != nil {

				// 查看否有等待成交的order 逾期
				expiredOrder := MyTrading.OrderStatusTest.InspectTimeOut(data.TradingTime)

				if expiredOrder != nil {

					newPrice, _ := data.Price.Float64()
					newPrice = utils.FloorFloat(newPrice, 2) // 小數點下第二為 四捨五入 , api order 只能最多兩位

					// 將過期order 時間 金額 更新 準備等等替換
					expiredOrder.CreateTime = data.TradingTime
					expiredOrder.Price = decimal.NewFromFloat(newPrice).String()

					if MyTrading.OrderStatusTest.OTA == true {

						if MyTrading.OrderStatusTest.Stage == 1 {

							referApply := MyTrading.OpenJudge(MyStrategy)
							if referApply == nil {
								logrus.Infoln("not the time to replace OpenOrder")
								MyTrading.OrderStatusTest.ClearOrderStatus()

								continue
							}
							referApply.OrderContent["OPEN"].Quantity = MyTrading.Quantity
							MyTrading.OrderStatusTest.ReplaceOpenOrder(referApply.OrderContent["OPEN"])
							logrus.Infoln("replace open order... at ", data.TradingTime)
							// OTA 第一階段都還沒買入就逾期 直接放棄此次交易 直接清空狀態 進入下一輪 重新開始
							//MyTrading.OrderStatusTest.ClearOrderStatus()

						} else if MyTrading.OrderStatusTest.Stage == 2 {

							MyTrading.OrderStatusTest.ReplaceCloseOrder(expiredOrder)
							logrus.Infoln("replace close order... at ", data.TradingTime)
						}

						// 處理完 ota 過期 ,結束此循環
						continue

					}

					// replace NormalOrder
					// 延遲normal order, 價格需用 openjudge更新

					if MyTrading.OrderStatusTest.Stage == 1 {

						newApply := MyTrading.OpenJudge(MyStrategy)

						// 若逾期 此時還是適合開單 replace
						if newApply != nil {
							logrus.Infoln("replace open order... at ", data.TradingTime)
							newPrice := newApply.DefaultContent.Price
							expiredOrder.Price = newPrice
							expiredOrder.CreateTime = data.TradingTime
							MyTrading.OrderStatusTest.ReplaceOpenOrder(expiredOrder)
							continue
						}
						// 若逾期 openJudge 判斷不適合開單 則直接撤銷
						logrus.Infoln("not the good time to open order, clear order....")
						MyTrading.OrderStatusTest.ClearOrderStatus() // 測試等待用replace的 減少 op 頻率

					} else if MyTrading.OrderStatusTest.Stage == 2 {

						// 預期 將手中東西盡可能快速出獲
						MyTrading.OrderStatusTest.ReplaceCloseOrder(expiredOrder)
						logrus.Infoln("replace close order... at ", data.TradingTime)
					}

					// 處理完 normal 過期order 結束此輪
					continue

				}

				// 無過期則下一輪
				continue
			}

			if MyTrading.OrderStatusTest.Stage == 0 {

				// workorder 不存在 但 stage == 0 , 表示之前所有order已經被 filled , 清空狀態

				MyTrading.OrderStatusTest.ClearOrderStatus()
				MyTrading.OrderStatusTest.Put_POST_DELETE.Sub(decimal.NewFromInt(1)) // 無實際訂單操作

				// 準備決策是否 開 OpenOrder
				applyOrder := MyTrading.OpenJudge(MyStrategy)
				// order不存在 直接結束
				if applyOrder != nil {
					logrus.Infoln("CurrentPrice: ", MyTrading.SQ.LatestData.Price)
					MyOrder := MyTrading.OrderProcess(applyOrder)
					MyTrading.OrderStatusTest.CreateOrder(MyOrder)
					logrus.Infoln("CreateOrderTime: ", MyTrading.SQ.LatestData.TradingTime)
				}

				// 結束此循環
				continue

			}

			// workingOrder 不存在 但 stage 不是0 表示 open filled 準備決策是否開 closeOrder
			applyOrder := MyTrading.CloseJudge(MyStrategy)
			if applyOrder != nil {
				logrus.Infoln("CurrentPrice: ", MyTrading.SQ.LatestData.Price)
				MyOrder := MyTrading.OrderProcess(applyOrder)
				MyTrading.OrderStatusTest.CreateOrder(MyOrder)
				logrus.Infoln("CreateOrderTime: ", MyTrading.SQ.LatestData.TradingTime)

			}

			// 結束此循環
			continue
		}

		// 無任何order 直接判斷是否開單

		applyOrder := MyTrading.OpenJudge(MyStrategy)
		// order不存在 直接結束
		if applyOrder != nil {
			logrus.Infoln("CurrentPrice: ", MyTrading.SQ.LatestData.Price)
			MyOrder := MyTrading.OrderProcess(applyOrder)
			MyTrading.OrderStatusTest.CreateOrder(MyOrder)
			logrus.Infoln("CreateOrderTime: ", MyTrading.SQ.LatestData.TradingTime)
		}

	}
	// 最後一天
	staticPerformance.TotalDay = staticPerformance.TotalDay.Add(decimal.NewFromInt(1))
	staticPerformance.Put_POST_DELETE = staticPerformance.Put_POST_DELETE.Add(MyTrading.OrderStatusTest.Put_POST_DELETE)
	staticPerformance.TotalWin = staticPerformance.TotalWin.Add(MyTrading.OrderStatusTest.Win)
	staticPerformance.TotalEvent = staticPerformance.TotalEvent.Add(MyTrading.OrderStatusTest.Event)
	staticPerformance.Append(MyTrading.SQ.LatestData.TradingTime.Format("2006-01-02"), MyTrading.OrderStatusTest.SumEarns)
	//staticPerformance.Report()

	PerformanceOutput <- &staticPerformance

}
