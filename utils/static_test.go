package utils

import (
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
	"testing"
)

func TestSymbolStatic_StaticInfo(t *testing.T) {
	psqlCli := &PsqlClient{}
	redisCli := GetRedis("0")
	static := SymbolStatic{Symbols: []string{"U"}, AnyDBClient: psqlCli, RedisCLi: redisCli}
	static.StaticInfo("2023-02-22", "2023-03-07")

}

func TestSymbolStatic_StaticCandle(t *testing.T) {
	psqlCli := &PsqlClient{}
	redisCli := GetRedis("0")
	static := SymbolStatic{Symbols: []string{"U"}, AnyDBClient: psqlCli, RedisCLi: redisCli}
	static.StaticWave(5, "2023-02-22", "2023-03-07")
}

func TestSymbolStatic_StaticAllTradingTimes(t *testing.T) {
	psqlCli := &PsqlClient{}
	redisCli := GetRedis("0")
	static := SymbolStatic{Symbols: []string{"NET"}, AnyDBClient: psqlCli, RedisCLi: redisCli}
	//static.StaticAllTradingTimes("2023-03-01")
	//static.StaticAllTradingTimes("2023-03-07")
	//
	//redisCli.FlushAll(ctx)
	//fmt.Println("second----------")
	static.StaticAllTradingTimes("2023-03-10")

}

func TestSymbolStatic_StaticNDayTradingTimes(t *testing.T) {
	psqlCli := &PsqlClient{}
	redisCli := GetRedis("0")
	static := SymbolStatic{Symbols: []string{"U"}, AnyDBClient: psqlCli, RedisCLi: redisCli}
	static.StaticNDayTradingTimes("2023-03-07", 10)

}

// 10日 avg為限
// 5日 當日 avg 為限
func TestSymbolStatic_StaticNDayTradingTimesRank(t *testing.T) {
	symbol := "TSLA"
	psqlCli := &PsqlClient{}
	redisCli := GetRedis("0")
	static := SymbolStatic{Symbols: []string{symbol}, AnyDBClient: psqlCli, RedisCLi: redisCli}
	//static.StaticAllTradingTimes("2023-04-07")
	// 統計所有交易次數
	static.StaticNDayTradingTimes("2023-02-28", 10)
	//// 統計n天內交易次數
	//
	//static.StaticNDayTradingTimesRank(symbol, 10, "2d")
	//// 針對不同精度單位 排序
	//
	res := static.StaticNDayTradingTimesRank(symbol, 10, "1f")
	fmt.Println(res.Price)
	fmt.Println(res.TradingTimes)

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
	fmt.Println(Wprice.Div(sumvol))

	//static.StaticNDayTradingTimesRank(symbol, 10, "1d")
	//static.StaticNDayTradingTimesRank(symbol, 10, "2d")

	//static.StaticNDayTradingTimesRank(symbol, 0, "1d")

}

func TestSymbolStaticAll(t *testing.T) {
	enddate := "2023-03-31"
	startdate := "2023-03-27"
	symbol := []string{"AAPL", "GOOG", "TSLA", "AMD", "INTC", "MSFT", "SPY", "NVDA"}
	psqlCli := &PsqlClient{}
	redisCli := GetRedis("0")
	static := SymbolStatic{Symbols: symbol, AnyDBClient: psqlCli, RedisCLi: redisCli}
	static.StaticInfo(startdate, enddate)    // 統計過去一段時間 交易量 交易事件
	static.StaticWave(5, startdate, enddate) // 統計過去一段時間波動率
	//static.StaticAllTradingTimes(enddate)               // 統計所有時間各價格交易次數
	//static.StaticNDayTradingTimes(enddate, 10)          // 統計10天內各價格交易次數
	//static.StaticNDayTradingTimesRank(symbol, 10, "1d") // 將10天各價格交易次數 進行猶大至小排序 精度為各位數
	//static.StaticNDayTradingTimesRank(symbol, 10, "2d") // 將10天各價格交易次數 進行猶大至小排序 精度十位數
	//static.StaticNDayTradingTimesRank(symbol, 10, "1f") // 將10天各價格交易次數 進行猶大至小排序 精度小數一位
}

func TestSymbolStatic_StaticUnitWave(t *testing.T) {
	symbol := []string{"AAPL", "GOOG", "TSLA", "AMD", "INTC", "MSFT", "SPY", "NVDA"}
	psqlCli := &PsqlClient{}
	redisCli := GetRedis("0")
	static := SymbolStatic{Symbols: symbol, AnyDBClient: psqlCli, RedisCLi: redisCli}

	enddate := "2023-03-31"

	static.StaticUnitWave(5, enddate, 5, 5)
}

// 只統計前五分鐘
func TestSymbolStatic_StaticOpenUnitWave(t *testing.T) {
	symbol := []string{"AAPL", "GOOG", "TSLA", "AMD", "INTC", "MSFT", "SPY", "NVDA"}
	psqlCli := &PsqlClient{}
	redisCli := GetRedis("0")
	static := SymbolStatic{Symbols: symbol, AnyDBClient: psqlCli, RedisCLi: redisCli}

	enddate := "2023-03-31"

	static.StaticOpenUnitWave(5, enddate, 5, 5)
}

func TestSymbolStatic_StaticNDayBuffer(t *testing.T) {
	symbol := []string{"U"}
	psqlCli := &PsqlClient{}
	redisCli := GetRedis("0")
	static := SymbolStatic{Symbols: symbol, AnyDBClient: psqlCli, RedisCLi: redisCli}

	enddate := "2023-04-20"

	static.StaticNDayBuffer(symbol, enddate, 3, "1f")

	static.StaticNDayBuffer(symbol, enddate, 7, "1f")
}

func TestSymbolStatic_StaticOpenGap(t *testing.T) {
	symbol := []string{"GOOG"}
	psqlCli := &PsqlClient{}
	redisCli := GetRedis("0")
	static := SymbolStatic{Symbols: symbol, AnyDBClient: psqlCli, RedisCLi: redisCli}
	endDate := "2023-05-07"
	static.StaticOpenGap(endDate, 20) // 三個月

}
