package model

import (
	"GoBot/utils"
	"fmt"
	"github.com/shopspring/decimal"
)

type StrategyInterface interface {
	GetID() string
	GetOrderKind() string
	GetOrderExpiredTime() int
	GetDayTrade() bool
	GetOpenJudgement(data *utils.JudgeReferData) *utils.ApplyOrder
	GetCloseJudgement(data *utils.JudgeReferData) *utils.ApplyOrder
}

// demo rsi < 20 , 買 , rsi > 80, 賣  僅Demo ,非有效策略
// 交易策略
type MyDemoStrategy struct {
}

// 設置策略代號
func (self *MyDemoStrategy) GetID() string {

	return "S"
}

// 是否為日內交易策略
func (self *MyDemoStrategy) GetDayTrade() bool {

	return true
}

// OTA(複合訂單) or Normal
func (self *MyDemoStrategy) GetOrderKind() string {

	return "Normal"
}

// 掛出訂單 若無成交 下次更新價格時間 , OpenOrder 會重新用 GetOpenJudgement 判斷 , CloseOrder 只要之前符合一次條件 逾期就用新進價格取代
func (self *MyDemoStrategy) GetOrderExpiredTime() int {

	return 100
}

// 開單判斷
func (self *MyDemoStrategy) GetOpenJudgement(Data *utils.JudgeReferData) *utils.ApplyOrder {

	sq := Data.SQ // 類似queue , 但只會保留某段時間內的交易資料, e.g. sq  TimeAllowSec =5 , 只保留近5秒內的交易資料

	ind := utils.Indicator{sq.Data}

	rsi := ind.GetRSI()

	if rsi < 20 {

		applyOrder := self.LongApply(Data)

		return applyOrder

	} else {
		return nil
	}

}

// buy long
func (self *MyDemoStrategy) LongApply(Data *utils.JudgeReferData) *utils.ApplyOrder {

	sq := Data.SQ

	fix := 0.
	targetPrice, _ := sq.Data[len(sq.Data)-1].Price.Float64()
	targetPrice = utils.FloorFloat(targetPrice, 2)                                    // 目標價四捨五入
	if decimal.NewFromFloat(targetPrice).GreaterThan(sq.Data[len(sq.Data)-1].Price) { // 若比最新平均價格還貴 -0.01
		targetPrice -= 0.01
	}

	targetPrice_string := fmt.Sprintf("%.2f", targetPrice-fix) //交易系統只能到小數第二位 ,

	applyOrder := &utils.ApplyOrder{Strategy: self.GetID(), OrderContent: make(map[string]*utils.UnitLimitOrder)}

	applyOrder.OrderKind = "Normal"

	applyOrder.DefaultContent = &utils.UnitLimitOrder{Symbol: sq.Data[0].Symbol, Price: targetPrice_string, OrderType: 10, CreateTime: sq.LatestData.TradingTime.In(utils.Loc)}

	return applyOrder
}

// 關單判斷
func (self *MyDemoStrategy) GetCloseJudgement(Data *utils.JudgeReferData) *utils.ApplyOrder {

	sq := Data.SQ // 類似queue , 但只會保留某段時間內的交易資料, e.g. sq  TimeAllowSec =5 , 只保留近5秒內的交易資料

	ind := utils.Indicator{sq.Data}

	rsi := ind.GetRSI()

	if rsi > 80 {

		applyOrder := self.LongApply(Data)

		return applyOrder

	} else {
		return nil
	}

	return nil

}
