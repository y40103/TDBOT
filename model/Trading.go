package model

import (
	"GoBot/utils"
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strings"
	"time"
)

// 若需修改策略 則是直接修改 Strategy部份
type Trading struct {
	Symbol            string
	Budget            decimal.Decimal
	Quantity          int
	SQ                *utils.SQ
	LocalOrderStatus  *utils.LocalOrderStatus
	OrderAPIOperation *utils.OrderOperation
	TradingLock       bool // 目前 用於一些交易意外 鎖住下單功能 or 蓄意停止交易能 , 	// 需滿滿足 //	GetID() string, GetOrderKind() string, OpenJudge(sq *utils.SQ) *utils.ApplyOrder, CloseJudge(sq *utils.SQ) *utils.ApplyOrder,
	DayTrade          bool
}

// 確認線上order狀態 , 若非活動 則刪除local Order Status

// 實際上還需做與倉位內的symbol做確認 有可能已經 openOrder filled 但等待時間準備 closeOrder 此時不會有order有在工作
func (self *Trading) ClearPreOrderStatus() {

	logrus.Infoln("Prepare to Clear PreOrderStatus ...")

	attemp := 0

	symbolList := make([]string, 0)

	for {

		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

		mainOrder, triggerOder, statusCode := self.OrderAPIOperation.UpdateLimitOrder(ctx, 50)

		logrus.Infoln("PreOrderStatus: ", statusCode)

		if utils.HttpSuccess(statusCode) {

			for _, each := range mainOrder {
				if each.Status == "WORKING" || each.Status == "QUEUED" || strings.Contains(each.Status, "PENDING") || strings.Contains(each.Status, "AWAITING") {

					symbolList = append(symbolList, each.Symbol)

				}

			}

			for _, each := range triggerOder {
				if each.Status == "WORKING" || each.Status == "QUEUED" || strings.Contains(each.Status, "PENDING") || strings.Contains(each.Status, "AWAITING") {

					symbolList = append(symbolList, each.Symbol)

				}

			}
			break
		}

		if attemp == 3 {
			logrus.Warnln("Fail to ClearPreOrderStatus")
			break
		}

		attemp += 1
		logrus.Warnf("retry to ClearPreOrderStatus %v...", attemp)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	TrackingSymbols := self.LocalOrderStatus.GetTrackingSymbol(ctx)

outer:
	for _, ts := range TrackingSymbols {

		for _, workingSymbol := range symbolList {
			if ts == workingSymbol {
				logrus.Infof("Keep Working Symbol...", ts)
				continue outer
			}
		}
		logrus.Infof("clear %v PreOrderStatus", ts)
		self.LocalOrderStatus.ClearSymbolAllOrderStatus(ctx, ts)
	}

	logrus.Infoln("PreOrderStatus clear task complete")

}

// SQ存在 Qantity 尚未有數值 進行數量評估
func (self *Trading) GetQuantity() {

	if self.Quantity == 0 && len(self.SQ.Data) > 0 {

		self.Quantity = int(self.Budget.Div(self.SQ.Data[0].Price).IntPart())

		logrus.Infof("Set %v Quantity is %v", self.Symbol, self.Quantity)

	}

}

// 至設定時間後  關閉openOrder功能  本次交易結束後 就無法開始下次transaction
func (self *Trading) InspectTradingTime(closeAfterTime time.Time) {

	// 假設其他原因導致已經被Lock 直接結束
	if self.LocalOrderStatus.GetOpenOrderLockStatus(context.Background(), self.Symbol) == true {
		return
	}

	if time.Now().In(utils.Loc).After(closeAfterTime) {

		self.TradingLock = true
	}

}

func (self *Trading) InspectOrderOperationLimit() {

	// 假設其他原因導致已經被Lock 直接結束
	if self.LocalOrderStatus.GetOpenOrderLockStatus(context.Background(), self.Symbol) == true {
		return
	}
	OrderOP := self.LocalOrderStatus.GetOrderOperationNum()
	if OrderOP > 375 {
		logrus.Infof("Already to today Trading Order Limit %v", OrderOP)
		self.TradingLock = true
	}

}

func (self *Trading) OpenJudge(Strategy StrategyInterface) *utils.ApplyOrder {
	data := &utils.JudgeReferData{SQ: self.SQ}
	applyOrder := Strategy.GetOpenJudgement(data)

	if self.TradingLock == true { // 用於通訊意外 鎖住交易功能
		applyOrder = nil
	}

	return applyOrder

}

func (self *Trading) CloseJudge(Strategy StrategyInterface) *utils.ApplyOrder {
	Open, _ := self.LocalOrderStatus.GetOpenCloseOrder(context.Background(), self.Symbol)
	data := &utils.JudgeReferData{SQ: self.SQ, ReferOrder: Open}
	applyOrder := Strategy.GetCloseJudgement(data)

	//if self.TradingLock == true { // 用於通訊意外 鎖住交易功能   // 主要鎖住openJudge
	//	applyOrder = nil
	//}

	return applyOrder
}

// 處理api掛單 批配掛單 更新OrderStatus至Local
func (self *Trading) ProcessApplyOrder(applyOrder *utils.ApplyOrder) {

	var ctx context.Context
	var mainOrderAPISource, triggerOrderAPISource map[int64]*utils.UnitLimitOrder

	if applyOrder != nil {

		logrus.Infof("Process Apply Order %+v", *applyOrder)

		// 設定交易數量
		if applyOrder.DefaultContent == nil {
			applyOrder.OrderContent["OPEN"].Quantity = self.Quantity
			applyOrder.OrderContent["CLOSE"].Quantity = self.Quantity
		} else {
			applyOrder.DefaultContent.Quantity = self.Quantity
		}

		// API 掛單
		httpCode := self.OrderAPIOperation.PlaceLimitOrder(applyOrder, 1500)

		// API取得所有掛單 & Match掛單 & 更新order線上狀態至local
		// 剛掛出去的單 有時候API抓取的資料 還沒更新進去 Match會失敗若 match 無內容 嘗試重抓
		if utils.HttpSuccess(httpCode) {

			// api 抓取掛單資訊
			mainOrderAPISource, triggerOrderAPISource, httpCode = self.OrderAPIOperation.GetAllOrderStatusSourceFromAPI()

			if applyOrder.OrderKind == "OTA" {

				matchMainOrder, matchTriggerOrder := self.OrderAPIOperation.MatchFullOTAOrderID(applyOrder.OrderContent["OPEN"], applyOrder.OrderContent["CLOSE"], mainOrderAPISource, triggerOrderAPISource)

				// 若提交order後 線上order狀態還沒即時更新 嘗試重試
				attempt := 1
				for {

					time.Sleep(time.Duration(attempt) * time.Millisecond * 1000)
					if matchMainOrder.OrderID == 0 || matchTriggerOrder.OrderID == 0 {
						mainOrderAPISource, triggerOrderAPISource, httpCode = self.OrderAPIOperation.GetAllOrderStatusSourceFromAPI()
						matchMainOrder, matchTriggerOrder = self.OrderAPIOperation.MatchFullOTAOrderID(applyOrder.OrderContent["OPEN"], applyOrder.OrderContent["CLOSE"], mainOrderAPISource, triggerOrderAPISource)
					}

					if matchMainOrder.OrderID != 0 && matchTriggerOrder.OrderID != 0 {
						logrus.Infoln("Success to match OTA ORDER")
						logrus.Infof("main %+v", matchMainOrder)
						logrus.Infof("trigger %+v", matchTriggerOrder)
						break
					} else if attempt == 3 {
						logrus.Warnln("多次取得API更新 失敗 StatusCode: ", httpCode)
						logrus.Warnln("產生結果為 create OTA, local no record status, can't detect timeout," +
							" if next times keep fail, it will make apply the infinity orders....")
						// 鎖住 OpenOrder 功能
						self.TradingLock = true
						self.LocalOrderStatus.SetOpenOrderLock(context.Background(), self.Symbol)
						logrus.Warnf("%v Lock OpenOrder Capability", self.Symbol)

						return

					}

					attempt += 1
					logrus.Warnf("FAIL TO MATCH OTA ORDER , attempt... %v", attempt)

				}

				ctx, _ = context.WithTimeout(context.Background(), time.Millisecond*1500)
				self.LocalOrderStatus.CreateOTAOrderStatus(ctx, *matchMainOrder, *matchTriggerOrder)
			}

			// TODO: 待實現重啟機制
			if applyOrder.OrderKind == "Normal" || applyOrder.OrderKind == "NORMAL" {
				matchMainOrder := new(utils.UnitLimitOrder)
				attempt := 0
				for {
					matchMainOrder = self.OrderAPIOperation.MatchOrderID(applyOrder.DefaultContent, mainOrderAPISource, triggerOrderAPISource)

					if matchMainOrder.OrderID != 0 {
						logrus.Infoln("Success to Match Normal ORDER")
						logrus.Infof("main %+v", matchMainOrder)
						break
					} else if attempt == 3 {
						logrus.Warnln("Match Normal Order fail , lock....")
						self.TradingLock = true
						self.LocalOrderStatus.SetOpenOrderLock(context.Background(), self.Symbol)
						return
					}

					attempt += 1
					logrus.Warnf("FAIL TO MATCH NORMAL ORDER , attempt... %v", attempt)
				}

				ctx, _ = context.WithTimeout(context.Background(), time.Millisecond*1500)
				self.LocalOrderStatus.CreateOrderStatus(ctx, *matchMainOrder)

				logrus.Infof("Success OpenOrder %+v", *applyOrder)
				return
			}

			return
		}

		logrus.Warnln("OpenJudge PlaceOrder is failed")

		return

	}
}

func MainTrading(OrderAPI utils.TDOrder, Symbol string, MyStrategy StrategyInterface, Budget decimal.Decimal, OrderTimeOut int, DataChannel chan *utils.TradingData) {
	orderAPI := &OrderAPI
	OrderAPIOperate := utils.OrderOperation{}
	OrderAPIOperate.OrderAPI = orderAPI
	sq := utils.SQ{RedisCli: utils.GetRedis("0"), TimeAllowSec: time.Second * 30}
	MyTrading := &Trading{Symbol: Symbol,
		DayTrade:          true,
		Budget:            Budget,
		SQ:                &sq,
		LocalOrderStatus:  &utils.LocalOrderStatus{},
		OrderAPIOperation: &OrderAPIOperate}

	MyTrading.TradingLock = MyTrading.LocalOrderStatus.GetOpenOrderLockStatus(context.Background(), Symbol)
	// 初始化 OpenOrderLock, 若重啟後 依然不進行交易 , 該狀態12hr後會自動解除

	today := time.Now().In(utils.Loc)
	// 實際開市 0830
	marketTime := time.Date(today.Year(), today.Month(), today.Day(), 9, 30, 0, 0, utils.Loc)

	MyTrading.ClearPreOrderStatus()

	for {

		NewStreamingData := <-DataChannel

		if NewStreamingData.FormT == true || time.Now().In(utils.Loc).Before(marketTime) { // 盤前後 不交易
			rand.Seed(time.Now().UnixNano())
			logrus.Infof("EXTEND TIME... now: %v, marketTime: %v", time.Now().In(utils.Loc), marketTime)
			logrus.Infoln(NewStreamingData.FormT == true, time.Now().In(utils.Loc).Before(marketTime))
			continue
		}

		// 鎖住openOder, 該次交易為最後一次交易, 不會再開單
		if MyTrading.DayTrade == true {
			MyTrading.InspectTradingTime(time.Date(today.Year(), today.Month(), today.Day(), 11, 30, 0, 0, utils.Loc))
		}

		// 確認今日操作次數
		MyTrading.InspectOrderOperationLimit()

		MyTrading.SQ.Append(NewStreamingData)

		MyTrading.GetQuantity()
		// 初始化Symbol交易數量

		//logrus.Infoln(NewStreamingData.Symbol, NewStreamingData.Price, NewStreamingData.Volume, NewStreamingData.EventNum)

		// 更新 API Access token
		MyTrading.OrderAPIOperation.OrderAPI.AccessToken = AccessToken

		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		size := MyTrading.LocalOrderStatus.GetOrderStatusSize(ctx, Symbol)

		// 手上沒有任何order 狀態 , 判斷是否開單
		if size == 0 {

			applyOrder := MyTrading.OpenJudge(MyStrategy)
			MyTrading.ProcessApplyOrder(applyOrder)
			continue

		}

		// 手上已有order, 查看當前階段
		logrus.Infoln("Hold Order, Get Order Stage... ")
		stage, WorkingOrder := MyTrading.LocalOrderStatus.GetOrderStage(ctx, Symbol)
		logrus.Infof("stage =%v  WorkingOrder = %v ", stage, WorkingOrder)

		// local紀錄手上有order, 但查看後 發現已經全部Filled , 清空Local狀態 , 重新判斷是否開單
		if stage == 0 {

			logrus.Infoln("Stage=0, All Order Filled, 判斷是否掛OpenOrder")
			ctx, _ = context.WithTimeout(context.Background(), time.Second*5)

			logrus.Infoln("Clear All Filled Order From Local")
			MyTrading.LocalOrderStatus.ClearSymbolAllOrderStatus(ctx, Symbol)

			applyOrder := MyTrading.OpenJudge(MyStrategy)

			MyTrading.ProcessApplyOrder(applyOrder)

			continue

			// order 尚未關閉 判斷是否需要開單
		} else if stage > 0 && WorkingOrder == nil {

			logrus.Infoln("Stage > 0, 判斷是掛closeOrder")
			applyOrder := MyTrading.CloseJudge(MyStrategy)
			MyTrading.ProcessApplyOrder(applyOrder)

			continue

		} else if stage == -1 {

			logrus.Warnln("有bug ")
			logrus.Warnf("size = %v, stage = %v, workingOrder = %+v", size, stage, WorkingOrder)
			panic("ERROR")
			continue

		}

		// order 存在 檢測 timeout
		// expired = -1 表示尚未過期 , WorkingOrder != nil , 表示有order正在掛單

		logrus.Infof("Inspect if the existing orders are overdue")
		ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
		stage, WorkingOrder = MyTrading.LocalOrderStatus.InspectTimeOut(ctx, Symbol, OrderTimeOut)

		// 訂單存在 , 尚未過期
		if stage == 100 {

			continue

		} else if WorkingOrder != nil { // 狀態 WORKING Order 過期處理

			ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
			OTA := MyTrading.LocalOrderStatus.OrderIsOTA(ctx, Symbol)

			// 去除price小數點 創建newOrder instance
			price, _ := sq.Data[len(sq.Data)-1].Price.Float64()
			pricefloor := utils.FloorFloat(price, 2)
			newOrder := &utils.UnitLimitOrder{
				Symbol:     Symbol,
				Quantity:   WorkingOrder.Quantity,
				OrderType:  WorkingOrder.OrderType,
				Price:      fmt.Sprintf("%v", pricefloor),
				CreateTime: NewStreamingData.TradingTime}
			// 更換訂單後 需更新更換時間 確認下次逾期用

			if OTA {

				// OTA OpenOrder 階段逾期 替換 或是 當下openJudge 不適合 直接砍單
				// // TODO: 若中間突然交易成功 有可能需要做一些處理  FILLED 的order 不確定能不能被刪掉 , 若刪不掉就不會有異常 , 若刪的掉 有可能trigger 被 open , 但local還有紀錄 待觀察
				if stage == 1 {

					// 產生心藥替換的MainOrder
					applyOrder := MyTrading.OpenJudge(MyStrategy)

					if applyOrder == nil { // 如果逾期 但當下也不符合開單條件 直接撤銷
						logrus.Infoln("try to replaceOpenOTAOrder, but not the good time to open order..")
						statusCode := MyTrading.OrderAPIOperation.DeleteOTAFullOrder(WorkingOrder)

						// 刪除local status
						if utils.HttpSuccess(statusCode) {

							ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
							MyTrading.LocalOrderStatus.ClearSymbolAllOrderStatus(ctx, WorkingOrder.Symbol)

							MyTrading.LocalOrderStatus.AddOrderOperationNum(1) // 刪除算是操作

							logrus.Infof("TimeOutFullOTAOrder Clear %v all localorder Status", Symbol)
						} else {
							logrus.Warnln("TimeOutFullOTAOrder Fail: 有可能正要刪時 突然成交", statusCode)
							// 不確定到底有沒有成功刪除 若狀態有變為cancel 就用跟用把FILLED order 從 local 刪除相同方式
						}

						continue
					}

					// 若為適合開單時間 更換order
					newOrder = applyOrder.OrderContent["OPEN"]
					newOrder.Quantity = MyTrading.Quantity
					logrus.Infof("Start to replaceOpenOTAOrder.....")
					ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
					oldMainOrder, oldTriggerOrder := MyTrading.LocalOrderStatus.GetOTAMainTriggerStatus(ctx, Symbol)
					newMainOrder, newTriggerOrder := MyTrading.OrderAPIOperation.ReplaceOpenOTAOrder(oldMainOrder, oldTriggerOrder, newOrder)

					// 清除 old order from local
					ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
					MyTrading.LocalOrderStatus.ClearSymbolAllOrderStatus(ctx, Symbol)

					// replace時 有機會exceed 有可能有replace成功 嘗試match, 若＼match有結果 表示成功, 若match失敗
					// 可能replace進行時 中途被filled  所以replace失敗 表示沒有必要replace了... 直接結束此輪
					if newMainOrder == nil && newTriggerOrder == nil {
						continue
					}

					// create local order
					ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
					MyTrading.LocalOrderStatus.CreateOTAOrderStatus(ctx, *newMainOrder, *newTriggerOrder)

					logrus.Infof("Replace New OTA OPEN: %+v\n", newMainOrder)

					continue // OTA 逾期處理結束  進入下一個循環

					// OTA CloseOrder 階段逾期
				} else if stage == 2 {

					// replace order by API
					logrus.Infof("Start to replaceCloseOTAOrder.....")
					NewOrder := MyTrading.OrderAPIOperation.ReplaceCloseOTAOrder(WorkingOrder, newOrder)

					if NewOrder == nil || NewOrder.OrderID == 0 {
						logrus.Warnln("Status is missing", NewOrder)
						// 假設match不到 , 有機會是oldOrder 突然被filled掉 所以沒辦法再被replace 400, or replace 200, 但提交出去狀態為rejected
						// 也有可能 match 時 該列表沒更新 完全找不到 此時情況有可能是  stage1 filled, stage2 replaced
						continue
					}
					// match錯誤 才會時間錯亂, 理論上不會出現更新時間問題 會有線上create時間 怕批配到舊資料

					// 清除 old order from local
					ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
					MyTrading.LocalOrderStatus.ClearOrderStatus(ctx, fmt.Sprintf("%v", WorkingOrder.OrderID))

					// create local order
					ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
					MyTrading.LocalOrderStatus.CreateOrderStatus(ctx, *NewOrder)
					MyTrading.LocalOrderStatus.SetOTAStatus(ctx, NewOrder.Symbol)
					// CreateOrderStatus local中的OTA tag 會不見, 這個通常是用來創建NormalOrder的 , 但因CreateOTAOrder 需一組 , 這邊手動補

					continue // OTA 逾期處理結束  進入下一個循環

				}

			}

			// Normal Order 逾期處理

			if !OTA {

				applyOrder := MyTrading.OpenJudge(MyStrategy)

				if stage == 1 {

					if applyOrder == nil { // 如果逾期 但當下也不符合開單條件 直接撤銷
						logrus.Infoln("try to replaceOpenNormalOrder, but not the good time to open order..")
						statusCode := MyTrading.OrderAPIOperation.DeleteNormalOrder(WorkingOrder)

						// 刪除local status
						if utils.HttpSuccess(statusCode) {

							ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
							MyTrading.LocalOrderStatus.ClearSymbolAllOrderStatus(ctx, WorkingOrder.Symbol)

							MyTrading.LocalOrderStatus.AddOrderOperationNum(1) // 刪除算是操作

							logrus.Infof("TimeOutNormalOrder Clear %v all localorder Status", Symbol)
						} else {
							// 清調 就當作多買
							logrus.Warnln("TimeOutNormalOrder Fail: 有可能正要刪時 突然成交", statusCode)
							// 不確定到底有沒有成功刪除 若狀態有變為cancel 就用跟用把FILLED order 從 local 刪除相同方式
						}

						continue
					}

					// 若適合重新替換
					newOrder = applyOrder.DefaultContent
					NewOrder := MyTrading.OrderAPIOperation.ReplaceNormalOrder(WorkingOrder, newOrder)
					if NewOrder == nil || NewOrder.OrderID == 0 { // 假設match不到 , 有機會是oldOrder 突然被filled掉 所以沒辦法再被replace 400, or replace 200, 但提交出去狀態為rejected

						// try to reset
						statusCode := MyTrading.OrderAPIOperation.DeleteNormalOrder(WorkingOrder)

						// 刪除local status
						if utils.HttpSuccess(statusCode) {

							ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
							MyTrading.LocalOrderStatus.ClearSymbolAllOrderStatus(ctx, WorkingOrder.Symbol)

							MyTrading.LocalOrderStatus.AddOrderOperationNum(1) // 刪除算是操作

							logrus.Infof("TimeOutNormalOrder Clear %v all localorder Status", Symbol)
						} else {
							// 清調 就當作多買
							logrus.Warnln("TimeOutNormalOrder Fail: 有可能正要刪時 突然成交", statusCode)
							// 不確定到底有沒有成功刪除 若狀態有變為cancel 就用跟用把FILLED order 從 local 刪除相同方式
						}

						continue
					}
					ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
					MyTrading.LocalOrderStatus.ClearOrderStatus(ctx, fmt.Sprintf("%v", WorkingOrder.OrderID))
					MyTrading.LocalOrderStatus.CreateOrderStatus(ctx, *NewOrder)
					logrus.Infof("Replace Normal Open Order Success .... %+v", *NewOrder)
					continue

				} else if stage == 2 {

					NewOrder := MyTrading.OrderAPIOperation.ReplaceNormalOrder(WorkingOrder, newOrder)
					if NewOrder == nil || NewOrder.OrderID == 0 { // 假設match不到 , 有機會是oldOrder 突然被filled掉 所以沒辦法再被replace 400, or replace 200, 但提交出去狀態為rejected
						continue
					}
					ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
					MyTrading.LocalOrderStatus.ClearOrderStatus(ctx, fmt.Sprintf("%v", WorkingOrder.OrderID))
					MyTrading.LocalOrderStatus.CreateOrderStatus(ctx, *NewOrder)
					logrus.Infof("Replace Normal Close Order Success .... %+v", *NewOrder)
					continue

				}

			}

		}

	}

}
