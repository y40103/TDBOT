package main

import (
	"GoBot/utils"
	"fmt"
	"time"
)

func main() {
	t1 := time.Now()
	//enddate := time.Now().Format("2006-01-02")
	enddate := "2023-05-01"
	numDay := 7
	//logrus.Infof("Exec Date: %v, Num: %v", enddate, numDay)
	//
	Symbols := []string{"INTC"}

	psqlCli := &utils.PsqlClient{}
	redisCli := utils.GetRedis("0")
	static := utils.SymbolStatic{Symbols: Symbols, AnyDBClient: psqlCli, RedisCLi: redisCli}

	//static.StaticInfo(startdate, enddate)    // 統計過去一段時間 交易量 交易事件
	//static.StaticWave(5, "2023-03-17", "2023-03-31") // 統計過去一段時間波動率 (所有時段波動率平均)

	static.StaticUnitWave(5, enddate, numDay, 5)
	//static.StaticOpenUnitWave(5, enddate, numDay, 5)

	//static.StaticAllTradingTimes(enddate)               // 統計所有時間各價格交易次數
	//static.StaticNDayTradingTimes(enddate, 10)          // 統計10天內各價格交易次數
	//static.StaticNDayTradingTimesRank(symbol, 10, "1d") // 將10天各價格交易次數 進行猶大至小排序 精度為各位數
	//static.StaticNDayTradingTimesRank(symbol, 10, "2d") // 將10天各價格交易次數 進行猶大至小排序 精度十位數
	//static.StaticNDayTradingTimesRank(symbol, 10, "1f") // 將10天各價格交易次數 進行猶大至小排序 精度小數一位
	fmt.Println("cost: ", time.Since(t1))
}
