package main

import (
	"GoBot/model"
	"GoBot/utils"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

func Single() {
	var WG sync.WaitGroup
	Symbol := "SE"
	logrus.SetLevel(logrus.DebugLevel)
	//MyStrategy := &model.WaveStrategy{}
	MyStrategy := &model.TestStrategy{}

	psql := utils.PsqlClient{}
	//sq := utils.SQ{RedisCli: utils.GetRedis("0"), TimeAllowSec: time.Second * 5}

	//data := psql.GetHisDataBackTest(Symbol, "2023-04-07", 1)
	data := psql.GetHisDataBackTest(Symbol, "2023-05-09", 5)

	//DevTrading := model.TradingTest{SQ: &utils.SQ{}}
	DataChan := make(chan *utils.TradingData)
	WG.Add(1)
	OutPutPerformance := make(chan *model.BackTestPerformance, 1000)

	// 目前為日內交易使用, 輸入日期可以為一段時間  1-5號之類的 績效會用一段時間來進行統計
	go model.DevTrading(MyStrategy, OutPutPerformance, 5, MyStrategy.GetOrderExpiredTime(), Symbol, decimal.NewFromInt(int64(1000)), DataChan, &WG)

	for _, val := range data {
		DataChan <- val
	}
	close(DataChan)

	WG.Wait()

	close(OutPutPerformance)

	for val := range OutPutPerformance {

		val.Report()

	}

}

type MultiPerformance struct {
	Symbol                                  string
	SingleEvent                             decimal.Decimal   // 總single數目
	EachSingleAvgDayPutPostDeleteRequestNum []decimal.Decimal // 每個single 日平均 order PUT/POST/DELETE 操作次數
	EachSinglePerformance                   []decimal.Decimal // 每個single 總收益
	EachSingleDate                          []string          // 每個single 其中之一日期
	EachSingleEarns                         []decimal.Decimal // 參照預算 收入金額
	Budget                                  decimal.Decimal   // 每個single 使用預算
	TotalTrade                              []decimal.Decimal // 每個single產生的交易事件
	TotalWinTrade                           []decimal.Decimal // 每個single正向收益的交易事件
	TotalPerformance                        decimal.Decimal   // 總收益
	SumTotalTrade                           decimal.Decimal
	SumTotalWinTrade                        decimal.Decimal
	SumEarns                                decimal.Decimal
}

func (self *MultiPerformance) ADD(EventPerformance *model.BackTestPerformance) {
	if self.Symbol == "" {
		self.Symbol = EventPerformance.Symbol
	}
	if self.EachSingleDate == nil {
		self.EachSingleAvgDayPutPostDeleteRequestNum = make([]decimal.Decimal, 0)
		self.EachSinglePerformance = make([]decimal.Decimal, 0)
		self.EachSingleDate = make([]string, 0)
		self.EachSingleEarns = make([]decimal.Decimal, 0)
		self.TotalTrade = make([]decimal.Decimal, 0)
		self.TotalWinTrade = make([]decimal.Decimal, 0)
		self.Budget = EventPerformance.DayBudget
	}

	self.SingleEvent = self.SingleEvent.Add(decimal.NewFromInt(1))

	self.EachSingleAvgDayPutPostDeleteRequestNum = append(self.EachSingleAvgDayPutPostDeleteRequestNum, EventPerformance.Put_POST_DELETE.Div(EventPerformance.DayEvent))
	self.EachSinglePerformance = append(self.EachSinglePerformance, EventPerformance.DayEarnsSum.Div(EventPerformance.DayBudget))
	self.EachSingleDate = append(self.EachSingleDate, EventPerformance.Date)
	self.TotalTrade = append(self.TotalTrade, EventPerformance.TotalEvent)
	self.TotalWinTrade = append(self.TotalWinTrade, EventPerformance.TotalWin)
	self.EachSingleEarns = append(self.EachSingleEarns, EventPerformance.DayEarnsSum)
}

func (self *MultiPerformance) Report() {
	fmt.Printf("\n%v EachEvent Profit: \n", self.Symbol)
	for index, date := range self.EachSingleDate {

		if self.TotalTrade[index].Equal(decimal.Zero) {
			continue
		}

		fmt.Printf("%v: Perforance: %v / %v = %v %%, AvgPutPostDeleteNum: %v / Day, TradeWinRate: %v / %v = %v %% \n",
			date,
			self.EachSingleEarns[index],
			self.Budget,
			self.EachSinglePerformance[index].Mul(decimal.NewFromInt(100)),
			self.EachSingleAvgDayPutPostDeleteRequestNum[index],
			self.TotalWinTrade[index],
			self.TotalTrade[index],
			self.TotalWinTrade[index].Div(self.TotalTrade[index]).Mul(decimal.NewFromInt(100)),
		)

	}

	avgPerformance := decimal.Zero
	allPerformance := decimal.Zero
	avgPutPostDeleteNum := decimal.Zero
	SingleWin := decimal.Zero

	for index := 0; index < int(self.SingleEvent.IntPart()); index++ {

		avgPerformance = avgPerformance.Add(self.EachSinglePerformance[index])
		allPerformance = allPerformance.Add(self.EachSinglePerformance[index])
		avgPutPostDeleteNum = avgPutPostDeleteNum.Add(self.EachSingleAvgDayPutPostDeleteRequestNum[index])
		self.SumTotalTrade = self.SumTotalTrade.Add(self.TotalTrade[index])
		self.SumTotalWinTrade = self.SumTotalWinTrade.Add(self.TotalWinTrade[index])
		self.SumEarns = self.SumEarns.Add(self.EachSingleEarns[index])
		if self.EachSinglePerformance[index].GreaterThan(decimal.Zero) {
			SingleWin = SingleWin.Add(decimal.NewFromInt(1))
		}

	}

	avgPerformance = avgPerformance.Div(self.SingleEvent)
	avgPutPostDeleteNum = avgPutPostDeleteNum.Div(self.SingleEvent)
	self.TotalPerformance = self.SumEarns.Div(self.Budget).Mul(decimal.NewFromInt(100))

	fmt.Printf("\n%v Total Profit:\n", self.Symbol)
	fmt.Printf("Total Performance: %v / %v = %v %% \n", self.SumEarns, self.Budget, self.TotalPerformance)
	fmt.Printf("Average Performance Per Single: %v %%\n", avgPerformance.Mul(decimal.NewFromInt(100)))
	fmt.Printf("Average DayPutPostDeleteNum Per Event: %v / Day\n", avgPutPostDeleteNum)

	// 統計 各single 所有交易統計勝率 總平均
	if !self.SumTotalTrade.Equal(decimal.Zero) {
		fmt.Printf("Average SingleTradeWinRate = %v / %v = %v %% \n", self.SumTotalWinTrade, self.SumTotalTrade, decimal.NewFromInt(100).Mul(self.SumTotalWinTrade.Div(self.SumTotalTrade)))
		fmt.Printf("SingleWinRate = %v / %v = %v %% \n", SingleWin, self.SingleEvent, SingleWin.Div(self.SingleEvent).Mul(decimal.NewFromInt(100)))

	}

	//
	if !self.SumTotalWinTrade.Equal(decimal.Zero) {
		fmt.Printf("Average Earns Per Win: %v %% \n", self.SumEarns.Div(self.SumTotalWinTrade).Div(self.Budget).Mul(decimal.NewFromInt(100)))
	}

}

type MultiBackTesting struct {
}

// for multiple  single goroutine
func (self *MultiBackTesting) SingeDateTask(OutPutPerformance chan *model.BackTestPerformance, Symbol string, MyStrategy model.StrategyInterface, Date string, SQAllowSec int, OrderExpired int) {

	psql := utils.PsqlClient{}
	wg2 := sync.WaitGroup{}
	data := psql.GetHisDataBackTest(Symbol, Date, 1)
	psql.CloseSession()
	DataChan := make(chan *utils.TradingData)
	wg2.Add(1)

	// 目前為日內交易使用, 輸入日期可以為一段時間  1-5號之類的 績效會用一段時間來進行統計
	go model.DevTrading(MyStrategy, OutPutPerformance, SQAllowSec, OrderExpired, Symbol, decimal.NewFromInt(int64(1000)), DataChan, &wg2)

	for _, val := range data {
		DataChan <- val
	}
	close(DataChan)
	wg2.Wait()
}

func (self *MultiBackTesting) MultipleQueue(taskDateQueue chan string, OutPutPerformance chan *model.BackTestPerformance, WG *sync.WaitGroup, Symbol string, MyStrategy model.StrategyInterface, SQAllowSec int, OrderExpired int) {
	defer WG.Done()

	for {
		date, ok := <-taskDateQueue

		if !ok {

			return
		}

		self.SingeDateTask(OutPutPerformance, Symbol, MyStrategy, date, SQAllowSec, OrderExpired)
		fmt.Println("Process Date ", date)
	}

}

func (self *MultiBackTesting) Multiple(date string, Symbos []string, MyStrategy model.StrategyInterface, BackTestDay int, SQAllowSec int, OrderExpired int) (AverageSymbolsPerformance decimal.Decimal) {
	logrus.SetLevel(logrus.WarnLevel)

	t, _ := time.Parse("20060102", date)

	result := make([]*MultiPerformance, 0)

	for _, Symbol := range Symbos {
		var WG sync.WaitGroup
		End := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, utils.Loc).Format("2006-01-02")
		psql := utils.PsqlClient{}
		dates := psql.GetActivityDateRange(Symbol, End, BackTestDay)
		psql.CloseSession()
		maxCPU := 16
		QueueDay := make(chan string, 0)
		OutPutPerformance := make(chan *model.BackTestPerformance, 1000)

		WG.Add(maxCPU)

		// 輸出 所有 goroutine tasks 參數  至 tasks chan 進行排隊
		go func(dates []time.Time, Queue chan string) {
			for _, eachDate := range dates {
				QueueDay <- eachDate.Format("2006-01-02")
			}

			close(QueueDay)
		}(dates, QueueDay)

		// 開啟 MaxCPU 組 goroutine
		for i := 0; i < maxCPU; i++ {

			go self.MultipleQueue(QueueDay, OutPutPerformance, &WG, Symbol, MyStrategy, SQAllowSec, OrderExpired)

		}

		// 維持goroutine
		go func(WG *sync.WaitGroup) {
			WG.Wait()
			close(OutPutPerformance)
		}(&WG)

		totalPerformance := MultiPerformance{}
		for eachDatePerformance := range OutPutPerformance {

			totalPerformance.ADD(eachDatePerformance)

		}
		totalPerformance.Report()
		result = append(result, &totalPerformance)
	}

	avgPerformance := decimal.Zero
	alltrade := decimal.Zero
	allwin := decimal.Zero
	allearns := decimal.Zero

	for _, res := range result {
		allearns = allearns.Add(res.SumEarns)
		avgPerformance = avgPerformance.Add(res.TotalPerformance)
		alltrade = alltrade.Add(res.SumTotalTrade)
		allwin = allwin.Add(res.SumTotalWinTrade)
	}

	fmt.Println("")
	fmt.Println(MyStrategy.GetID(), ":")
	fmt.Println("Average Symbols Performance ", " = ", avgPerformance.Div(decimal.NewFromInt(int64(len(result)))).RoundFloor(2), " %")

	if !allwin.Equal(decimal.Zero) && !alltrade.Equal(decimal.Zero) {
		// 所有single 交易事件 正獲利事件 勝率
		fmt.Printf("Strategy WinRate = %v / %v = %v %%\n", allwin, alltrade, allwin.Div(alltrade).Mul(decimal.NewFromInt(100)))
		fmt.Printf("Earns / Per Win = %v %%\n", allearns.Div(allwin).Div(result[0].Budget).Mul(decimal.NewFromInt(100)))
	}

	return avgPerformance.Div(decimal.NewFromInt(int64(len(result)))).RoundFloor(2)

}

func EachDateStaticMultiple(MyStrategy model.StrategyInterface, Symbols []string, EndDate string, DayNum int) {
	t, _ := time.Parse("20060102", EndDate)
	psqlCli := &utils.PsqlClient{}
	redisCli := utils.GetRedis("0")
	static := utils.SymbolStatic{Symbols: Symbols, AnyDBClient: psqlCli, RedisCLi: redisCli}
	dates := static.AnyDBClient.GetActivityDateRange(Symbols[0], t.Format("2006-01-02"), DayNum)
	psqlCli.CloseSession()
	AvgSum := decimal.Zero
	num := decimal.Zero
	for _, date := range dates {

		backtesting := &MultiBackTesting{}

		numDay := 7
		psqlCli = &utils.PsqlClient{}
		static = utils.SymbolStatic{Symbols: Symbols, AnyDBClient: psqlCli, RedisCLi: redisCli}
		static.StaticUnitWave(5, date.Format("2006-01-02"), numDay, 5)
		psqlCli.CloseSession()
		ASP := backtesting.Multiple(date.Add(time.Hour*24).Format("20060102"), Symbols, MyStrategy, 1, 5, MyStrategy.GetOrderExpiredTime())

		AvgSum = AvgSum.Add(ASP)
		num = num.Add(decimal.NewFromInt(1))
	}

	fmt.Printf("Average Performance = %v\n", AvgSum.Div(num).RoundFloor(2))

}

func main() {

	now := time.Now()
	//Single()

	Symbols := []string{"AAPL", "AMD", "NVDA"}
	date := "20230508"
	//h
	MyStrategy := &model.WaveStrategySingle4{}
	backtesting := &MultiBackTesting{}

	backtesting.Multiple(date, Symbols, MyStrategy, 70, 5, MyStrategy.GetOrderExpiredTime())

	//EachDateStaticMultiple(&model.WaveStrategySingle{}, Symbols, date, 10)
	//

	fmt.Println("\n##################")
	fmt.Printf("cost time = %v\n", time.Since(now))

}

// 30day
// ~ 30min 1.04
// ~ 60min 4.01 0900
// ~ 90min 4.85
// ~ 120min 7.28
// ~ 150min 8.66
// ~ 180min 10.16 1200
// ~ 210min 10.99
// ~ 240mub 12.06
// ~ 270mub 12.44
// ~ 270mub 12.64
// ~ 270mub 13.38
// ~ 270mub 14.17
// ~ 270mub 14.88

// expose = 75 sell_ratio=2 static=10
// cost time = 1h58m 28 大概要 2-3hr , nvda 2h25,msft 1h58,aapl 1h52 ,tsla 4h21m ,intc 43m ,amd 59m
//MSFT
// 0509 每天統計前7天 70day  p= 0.44  expose=75 sell_ratio=2,gap<0.7
//NVDA
// 0509 每天統計前7天 70day  p= 0.67   expose=75 sell_ratio=2,gap<0.7
//AAPL
// 0509 每天統計前7天 70day  p= 0.59   expose=75 sell_ratio=2,gap<0.7
//INTC
// 0509 每天統計前7天 70day  p= 0.35   expose=75 sell_ratio=2,gap<0.7
//AMD
// 0509 每天統計前7天 70day  p= 1.32   expose=75 sell_ratio=2,gap<0.7
//TSLA
// 0509 每天統計前7天 70day  p= 1.18   expose=75 sell_ratio=2,gap<0.7

// expose = 75 sell_ratio=2 static=10
//MSFT
// 0509 每天統計前5天 30day  p=0.08   expose=75 sell_ratio=2
//NVDA
// 0509 每天統計前5天 30day  p=0.46   expose=75 sell_ratio=2
//AAPL
// 0509 每天統計前5天 30day  p=0.26   expose=75 sell_ratio=2

// expose = 75 sell_ratio=2 static=7
//MSFT
// 0509 每天統計前5天 30day  p= 0.22  expose=75 sell_ratio=2
//NVDA
// 0509 每天統計前5天 30day  p= 0.51  expose=75 sell_ratio=2
//AAPL
// 0509 每天統計前5天 30day  p= 0.25   expose=75 sell_ratio=2

// expose = 75 sell_ratio=2 static=5
//MSFT
// 0509 每天統計前5天 30day  p=0.2   expose=75 sell_ratio=2
//NVDA
// 0509 每天統計前5天 30day  p=0.5   expose=75 sell_ratio=2
//AAPL
// 0509 每天統計前5天 30day  p=0.31   expose=75 sell_ratio=2

// expose = 75 sell_ratio=2 static=3
//MSFT
// 0509 每天統計前5天 30day  p=0.19   expose=75 sell_ratio=2
//NVDA
// 0509 每天統計前5天 30day  p=0.41   expose=75 sell_ratio=2
//AAPL
// 0509 每天統計前5天 30day  p=0.3   expose=75 sell_ratio=2

// static = 5

//MSFT
// 0509 每天統計前5天 30day  p=0.29   expose=100 sell_ratio=3
//AAPL
// 0509 每天統計前5天 30day  p=0,21   expose=100 sell_ratio=3
//NVDA
// 0509 每天統計前5天 30day  p=0.34   expose=100 sell_ratio=3

//MSFT

// 0509 每天統計前5天 10day  p= 0.42  expose=100 sell_ratio=3

// 0509 每天統計前5天 10day  p= 0.32  expose=125 sell_ratio=3

// 0509 每天統計前5天 10day  p= 0.55  expose=150 sell_ratio=3

// 0509 每天統計前5天 30day  p= 0.44  expose=150 sell_ratio=3
// 0509 每天統計前5天 30day  p= 0.38  expose=150 sell_ratio=2
// 0509 每天統計前5天 30day  p= -0.49  expose=150 sell_ratio=1

// NVDA

// 0509 每天統計前5天 10day  p= 0.67  expose=100 sell_ratio=3

// 0509 每天統計前5天 10day  p= 0.63  expose=125  sell_ratio=3

// 0509 每天統計前5天 10day  p= 0.68  expose=150 sell_ratio=3

// 0509 每天統計前5天 30day  p=0.38  expose=150 sell_ratio=3
// 0509 每天統計前5天 30day  p=0.53  expose=150 sell_ratio=2

// AAPL

// 0509 每天統計前5天 10day  p= 0.52  expose=100  sell_ratio=3

// 0509 每天統計前5天 10day  p= 0.46  expose=125  sell_ratio=3

// 0509 每天統計前5天 10day  p= 0.57  expose=150  sell_ratio=3

// 0509 每天統計前5天 30day  p=0.35  expose=150 sell_ratio=3
// 0509 每天統計前5天 30day  p=0.33  expose=150 sell_ratio=2

// TSLA
// 0509 每天統計前5天 30day  p=0.96  expose=150 sell_ratio=3

// GOOG
// 0509 每天統計前5天 30day  p=-0.15  expose=150 sell_ratio=3

// INTC
// 0509 每天統計前5天 30day  p=0.72  expose=150 sell_ratio=3

// AMD
// 0509 每天統計前5天 30day  p=0.89  expose=150 sell_ratio=3
