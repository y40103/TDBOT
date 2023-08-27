package utils

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

// 將券商api基本功能封裝為自己常用需求
type OrderOperation struct {
	OrderAPI *TDOrder
}

func (self *OrderOperation) UpdateLimitOrder(ctx context.Context, RespMaxResult int) (mainOrder map[int64]*UnitLimitOrder, triggerOrder map[int64]*UnitLimitOrder, httpStatusCode int) {
	orderTypeMapNo := map[string]int{"BUY": 10, "SELL": -10, "SELL_SHORT": 20, "BUY_TO_COVER": -20}

	mainOrder = make(map[int64]*UnitLimitOrder)
	triggerOrder = make(map[int64]*UnitLimitOrder)

	endate := time.Now().In(Loc).Format("2006-01-02")
	startdate := ""
	status := ""
	maxResult := fmt.Sprintf("%v", RespMaxResult)
	httpCode, orders := self.OrderAPI.GetCurrentOrderStatus(ctx, startdate, endate, status, maxResult) // 抓下來之後 預設時區為utc
	if httpCode >= 200 && httpCode < 300 {
		// 更新本地order
		for _, val := range *orders {

			price := fmt.Sprintf("%v", val.Price)

			createTime, err := time.ParseInLocation("2006-01-02T15:04:05+0000", val.EnteredTime, time.UTC)
			createTime = createTime.In(Loc)
			if err != nil {
				logrus.Infoln(err)
			}
			NewOrder := UnitLimitOrder{
				Symbol:     val.OrderLegCollection[0].Instrument.Symbol,
				OrderType:  orderTypeMapNo[val.OrderLegCollection[0].Instruction],
				OrderID:    val.OrderId,
				Quantity:   int(val.Quantity),
				Price:      price,
				Status:     val.Status,
				CreateTime: createTime,
				Editable:   val.Editable,
				Cancelable: val.Cancelable,
			}
			mainOrder[val.OrderId] = &NewOrder

			// 若有trigger order
			if len(val.ChildOrderStrategies) > 0 {
				for _, childval := range val.ChildOrderStrategies {

					triggerVal := childval

					price = fmt.Sprintf("%v", triggerVal.Price)

					createTime, err = time.ParseInLocation("2006-01-02T15:04:05+0000", val.EnteredTime, time.UTC)
					createTime = createTime.In(Loc)
					if err != nil {
						logrus.Infoln(err)
					}
					NewTriggerOrder := UnitLimitOrder{
						Symbol:     triggerVal.OrderLegCollection[0].Instrument.Symbol,
						OrderType:  orderTypeMapNo[triggerVal.OrderLegCollection[0].Instruction],
						OrderID:    triggerVal.OrderId,
						Quantity:   int(triggerVal.Quantity),
						Price:      price,
						Status:     triggerVal.Status,
						CreateTime: createTime,
						Editable:   triggerVal.Editable,
						Cancelable: triggerVal.Cancelable,
					}
					triggerOrder[triggerVal.OrderId] = &NewTriggerOrder

				}

			}

		}

		logrus.Infoln("Update Order Status LastUpdate Timestamp")
	}

	return mainOrder, triggerOrder, httpCode

}

// update limit order 增加重連機制
func (self *OrderOperation) GetAllOrderStatusSourceFromAPI() (mainOrderAPISource map[int64]*UnitLimitOrder, triggerOrderAPISource map[int64]*UnitLimitOrder, httpCode int) {

	mainOrderAPISource = make(map[int64]*UnitLimitOrder)
	triggerOrderAPISource = make(map[int64]*UnitLimitOrder)
	count := 0
	maxResult := 45
	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*2500)
		mainOrderAPISource, triggerOrderAPISource, httpCode = self.UpdateLimitOrder(ctx, maxResult)
		if HttpSuccess(httpCode) {
			logrus.Infoln(httpCode)
			break
		} else if count == 3 {
			logrus.Warnln(httpCode)
			break
		}
		count += 1
		maxResult += 15
		logrus.Warnf("UpdateLocalOrderStatus attempt... %v", count)
		time.Sleep(time.Millisecond * time.Duration(count) * 100)
	}

	return mainOrderAPISource, triggerOrderAPISource, httpCode
}

// 無法用於批配 OTA replace mainOrder
func (self *OrderOperation) MatchFullOTAOrderID(mainOrder *UnitLimitOrder, triggerOrder *UnitLimitOrder, SearchMainSource map[int64]*UnitLimitOrder, SearchTriggerSource map[int64]*UnitLimitOrder) (MatchMainOrder *UnitLimitOrder, MatchTriggerOrder *UnitLimitOrder) {
	if mainOrder.CreateTime.IsZero() || triggerOrder.CreateTime.IsZero() {
		logrus.Warningln("Main & Trigger don't contain create time...")
		panic("Main & Trigger don't contain create time...")
	}
	DealPriceIntPoint(mainOrder)
	DealPriceIntPoint(triggerOrder)
	MatchMainOrder = new(UnitLimitOrder)
	MatchTriggerOrder = new(UnitLimitOrder)

	// 更換成 decimal 比較 避免類似 10.7 != 10.70 的情況
	mainOrderPrice, _ := decimal.NewFromString(mainOrder.Price)
	triggerOrderPrice, _ := decimal.NewFromString(triggerOrder.Price)
	// mainOrder 批配
	for _, order := range SearchMainSource {
		orderPrice, _ := decimal.NewFromString(order.Price)

		logrus.Infof("candidate main: %+v\n", order)
		logrus.Infoln(order.Symbol == mainOrder.Symbol,
			orderPrice.LessThanOrEqual(mainOrderPrice), // 有時後會批配到 划算一點點的 ,怕放空有問題 試試看不用價格
			order.Quantity == mainOrder.Quantity,
			order.OrderType == mainOrder.OrderType,
			order.Status != "CANCELED",
			order.Status != "REPLACED",
			order.Status != "REJECTED",
			order.Status != "PENDING_REPLACE",
			order.Status != "EXPIRED",
			order.Status != "PENDING_CANCEL",
			order.Editable == false)
		logrus.Infoln(order.Symbol, mainOrder.Symbol,
			orderPrice, mainOrderPrice,
			order.Quantity, mainOrder.Quantity,
			order.OrderType, mainOrder.OrderType,
			order.Status != "CANCELED",
			order.Status != "REPLACED",
			order.Status != "REJECTED",
			order.Status != "PENDING_REPLACE",
			order.Status != "EXPIRED",
			order.Status != "PENDING_CANCEL",
			order.Editable == false)

		// 先比價格 價格誤差 0.1之內 , 若大於 0.1 直接排除
		if (orderPrice.Sub(mainOrderPrice)).Abs().GreaterThan(decimal.NewFromFloat32(0.3)) {
			logrus.Infof("ignore ,price difference great than 0.03.., candidate:%v,target:%v", orderPrice, mainOrderPrice)
			continue
		}

		if order.Symbol == mainOrder.Symbol &&
			orderPrice.LessThanOrEqual(mainOrderPrice) && // 有時後會批配到 划算一點點的 ,怕放空有問題 試試看不用價格
			order.Quantity == mainOrder.Quantity &&
			order.OrderType == mainOrder.OrderType &&
			order.Status != "CANCELED" &&
			order.Status != "REPLACED" &&
			order.Status != "REJECTED" &&
			order.Status != "PENDING_REPLACE" &&
			order.Status != "EXPIRED" &&
			order.Status != "PENDING_CANCEL" &&
			order.Editable == false { // 同 標的,交易數量,操作 挑出來比較

			//fmt.Println("#### ", order.OrderID, " ", (order.CreateTime.Sub(triggerOrder.CreateTime)).Abs())

			if MatchMainOrder.CreateTime.IsZero() { // order 創建時間與線上掛單創建時間最接近

				MatchMainOrder = order

			} else if (MatchMainOrder.CreateTime.Sub(mainOrder.CreateTime)).Abs() >= (order.CreateTime.Sub(mainOrder.CreateTime)).Abs() {

				// 若手上拿的 與 新來的 時間相同  那就用 orderID 比較 , OrderID比較大的數字(表示更新) 為批配對象
				if MatchMainOrder.CreateTime.Equal(order.CreateTime) && MatchMainOrder.OrderID > order.OrderID {
					continue
				}
				MatchMainOrder = order

			}
		}
	}

	for _, order := range SearchTriggerSource {
		orderPrice, _ := decimal.NewFromString(order.Price)
		logrus.Infof("candidate trigger %+v\n", order)
		logrus.Infoln(order.Symbol == triggerOrder.Symbol,
			orderPrice.LessThanOrEqual(triggerOrderPrice), // 有時後會批配到 划算一點點的 ,怕放空有問題 試試看不用價格
			order.Quantity == triggerOrder.Quantity,
			order.OrderType == triggerOrder.OrderType,
			order.Status != "CANCELED",
			order.Status != "REPLACED",
			order.Status != "REJECTED",
			order.Status != "PENDING_REPLACE",
			order.Status != "EXPIRED",
			order.Status != "PENDING_CANCEL",
			order.Editable == false)
		logrus.Infoln(order.Symbol, triggerOrder.Symbol,
			orderPrice, triggerOrderPrice,
			order.Quantity, triggerOrder.Quantity,
			order.OrderType, triggerOrder.OrderType,
			order.Status != "CANCELED",
			order.Status != "REPLACED",
			order.Status != "REJECTED",
			order.Status != "PENDING_REPLACE",
			order.Status != "EXPIRED",
			order.Status != "PENDING_CANCEL",
			order.Editable == false)

		// 先比價格 價格誤差 0.1之內 , 若大於 0.1 直接排除
		if (orderPrice.Sub(triggerOrderPrice)).Abs().GreaterThan(decimal.NewFromFloat32(0.3)) {
			logrus.Infof("ignore ,price difference great than 0.01.., candidate:%v,target:%v", orderPrice, mainOrderPrice)
			continue
		}

		if order.Symbol == triggerOrder.Symbol &&
			orderPrice.LessThanOrEqual(triggerOrderPrice) && // 有時後會批配到 划算一點點的 ,怕放空有問題 試試看不用價格
			order.Quantity == triggerOrder.Quantity &&
			order.OrderType == triggerOrder.OrderType &&
			order.Status != "CANCELED" &&
			order.Status != "REPLACED" &&
			order.Status != "REJECTED" &&
			order.Status != "PENDING_REPLACE" &&
			order.Status != "EXPIRED" &&
			order.Status != "PENDING_CANCEL" &&
			order.Editable == false { // 同 標的,價格,交易數量,操作 挑出來比較

			if MatchTriggerOrder.CreateTime.IsZero() { // order 創建時間與線上掛單創建時間最接近

				MatchTriggerOrder = order
			} else if (MatchTriggerOrder.CreateTime.Sub(mainOrder.CreateTime)).Abs() >= (order.CreateTime.Sub(mainOrder.CreateTime)).Abs() {

				// 若手上拿的 與 新來的 時間相同  那就用 orderID 比較 , OrderID比較大的數字(表示更新) 為批配對象
				if MatchTriggerOrder.CreateTime.Equal(order.CreateTime) && MatchTriggerOrder.OrderID > order.OrderID {
					continue
				}
				MatchTriggerOrder = order

			}
		}
	}

	logrus.Infof("matchMainOrder %+v", MatchMainOrder)
	logrus.Infof("matchTriggerOrder %+v", MatchTriggerOrder)

	return MatchMainOrder, MatchTriggerOrder

}

func DealPriceIntPoint(mainOrder *UnitLimitOrder) {
	val, _ := strconv.ParseFloat(mainOrder.Price, 64)

	Scaleval := 100 * val
	if int(Scaleval)%100 == 0 {
		intval := fmt.Sprintf("%v", int(val))
		mainOrder.Price = intval
	}
}

// 無法用於批配 OTA replace mainOrder
func (self *OrderOperation) MatchOrderID(mainOrder *UnitLimitOrder, SearchMainSource map[int64]*UnitLimitOrder, SearchTriggerSource map[int64]*UnitLimitOrder) (MatchMainOrder *UnitLimitOrder) {
	if mainOrder.CreateTime.IsZero() {
		logrus.Warningln("Main & Trigger don't contain create time...")
		panic("Main & Trigger don't contain create time...")
	}

	DealPriceIntPoint(mainOrder)

	MatchMainOrder = new(UnitLimitOrder)
	mainOrderPrice, _ := decimal.NewFromString(mainOrder.Price)
	//// 需將創建時間加上去

	// mainOrder 批配
	for _, order := range SearchMainSource {
		orderPrice, _ := decimal.NewFromString(order.Price)

		// 創建時間差太多 則沒有比較必要
		if order.CreateTime.Add(time.Second * 60).Before(mainOrder.CreateTime) {
			logrus.Infof("ignore ,create order time diff too much , candidate:%+v,target:%+v", order.CreateTime, mainOrder.CreateTime)
			continue
		}

		logrus.Infof("%v+\n", order)
		logrus.Infoln(order.Symbol == mainOrder.Symbol,
			//orderPrice.LessThanOrEqual(mainOrderPrice),  // 有時後會批配到 划算一點點的 ,怕放空有問題 試試看不用價格
			order.Quantity == mainOrder.Quantity,
			order.OrderType == mainOrder.OrderType,
			order.Status != "CANCELED",
			order.Status != "REPLACED",
			order.Status != "PENDING_REPLACE",
			order.Status != "REJECTED",
			order.Status != "EXPIRED",
			order.Status != "PENDING_CANCEL",
		)
		logrus.Infoln(order.Symbol, mainOrder.Symbol,
			orderPrice, mainOrderPrice,
			order.Quantity, mainOrder.Quantity,
			order.OrderType, mainOrder.OrderType,
			order.Status != "CANCELED",
			order.Status != "REPLACED",
			order.Status != "PENDING_REPLACE",
			order.Status != "REJECTED",
			order.Status != "EXPIRED",
			order.Status != "PENDING_CANCEL",
		)

		if order.Symbol == mainOrder.Symbol &&
			//orderPrice.LessThanOrEqual(mainOrderPrice) && // 有時後會批配到 划算一點點的 ,怕放空有問題 試試看不用價格
			order.Quantity == mainOrder.Quantity &&
			order.OrderType == mainOrder.OrderType &&
			order.Status != "CANCELED" &&
			order.Status != "REPLACED" &&
			order.Status != "PENDING_REPLACE" &&
			order.Status != "REJECTED" &&
			order.Status != "EXPIRED" &&
			order.Status != "PENDING_CANCEL" { // 同 標的,交易數量,操作 挑出來比較

			if MatchMainOrder.CreateTime.IsZero() { // order 創建時間與線上掛單創建時間最接近

				MatchMainOrder = order

				// 判斷match方式 是比較 當前手上拿的線上order與mainOrder的創建時間差 與 需確認order 與 mainOrder的創建時間差 最接近的會match
			} else if (MatchMainOrder.CreateTime.Sub(mainOrder.CreateTime)).Abs() >= (order.CreateTime.Sub(mainOrder.CreateTime)).Abs() {

				// 若手上拿的 與 新來的 時間相同  那就用 orderID 比較 , OrderID比較大的數字(表示更新) 為批配對象
				if MatchMainOrder.CreateTime.Equal(order.CreateTime) && MatchMainOrder.OrderID > order.OrderID {
					continue
				}
				MatchMainOrder = order

			}
		}
	}

	logrus.Infof("FINAL MATCH: %+v", *MatchMainOrder)

	return MatchMainOrder

}

// ApplyOrder 格式 進行API掛單 , TimeOutMilliSec 是api timeout判斷的時間長度
func (self *OrderOperation) PlaceLimitOrder(applyOrder *ApplyOrder, TimeOutMilliSec int) (httpStatusCode int) {

	if TimeOutMilliSec < 200 {
		TimeOutMilliSec = 1500
	}

	if applyOrder == nil {
		return
	}

	attempt := 0
	httpStatusCode = 0
	ordertype := ""

	// 掛ota order, 失敗重新嘗試

	for {

		ctx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(TimeOutMilliSec))
		if applyOrder.OrderKind == "OTA" {

			ordertype = "OTA"
			httpStatusCode = self.OrderAPI.CreateOTAOrder(ctx, applyOrder.OrderContent["OPEN"], applyOrder.OrderContent["CLOSE"])

		} else if applyOrder.OrderKind == "Normal" || applyOrder.OrderKind == "NORMAL" {

			ordertype = OrderTypeMap[applyOrder.DefaultContent.OrderType]
			httpStatusCode = self.OrderAPI.CreateLimitOrder(ctx, applyOrder.DefaultContent)

		}
		if HttpSuccess(httpStatusCode) {

			logrus.Infof("Success to Create %v order by API", ordertype)

			return httpStatusCode
		} else if attempt == 2 {
			logrus.Warnf("fail to Create %v order by API", ordertype)
			return httpStatusCode
		}

		logrus.Warnf("Create %v ORDER BY API, attemp.... %v", ordertype, attempt)
		time.Sleep(time.Millisecond * time.Duration(100*attempt))
		attempt += 1

	}

}

// 處理API ORDER 部份 , Local紀錄另外處理 , 最後回傳 newOrder 包含orderid
func (self *OrderOperation) ReplaceCloseOTAOrder(OldOrder *UnitLimitOrder, NewOrder *UnitLimitOrder) *UnitLimitOrder {

	price, _ := strconv.ParseFloat(NewOrder.Price, 64)
	newPrice := fmt.Sprintf("%v", FloorFloat(price, 2))

	newOrder := &UnitLimitOrder{
		Symbol:     OldOrder.Symbol,
		Price:      newPrice,
		OrderType:  OldOrder.OrderType,
		Quantity:   OldOrder.Quantity,
		CreateTime: NewOrder.CreateTime}

	// 防止兩單當相同
	newD, _ := decimal.NewFromString(newOrder.Price)
	oldD, _ := decimal.NewFromString(OldOrder.Price)
	if newD.LessThanOrEqual(oldD) {
		logrus.Infoln("replace order equal old order, new one + 0.01")
		newD.Add(decimal.NewFromFloat(0.01))
	} else if newD.GreaterThan(oldD) {
		logrus.Infoln("replace order equal old order, new one - 0.01")
		newD.Add(decimal.NewFromFloat(-0.01))
	}
	newOrder.Price = newD.String()
	OldOrder.Price = oldD.String()

	attempt := 0

	logrus.Infof("Prepare replace old order: %+v", OldOrder)
	logrus.Infof("TO New order: %+v", NewOrder)

	// API replace OTA order
	for {

		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		httpStatusCode := self.OrderAPI.ReplaceOrder(ctx, fmt.Sprintf("%v", OldOrder.OrderID), newOrder)

		if HttpSuccess(httpStatusCode) {
			logrus.Infoln("success to replace OTACloseOrder")
			break
		} else if attempt == 2 {
			logrus.Warnln("fail to Replace OTACloseOrder, if exceed... can't not confirm status ... 200 or other ...") // 失敗直接跳出 沒有match必要
			// return nil      // context exceed 有可能已經生效 不確定 繼續往下走看看

			break
		}
		attempt += 1
		logrus.Warnf("ReplaceOTAOrder attempt... %v", attempt)
		time.Sleep(time.Millisecond * time.Duration(100*attempt))

	}

	// match失敗 重新嘗試
	attempt = 0
	for {
		var mainorderAPISource, triggerOrderAPISource map[int64]*UnitLimitOrder
		httpCode := 0
		mainorderAPISource = make(map[int64]*UnitLimitOrder)
		triggerOrderAPISource = make(map[int64]*UnitLimitOrder)
		// Get OrderStatus From API

		count := 0 // API updateOrderStatus, 若失敗 重新嘗試
		maxResult := 45
		for {
			ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(1500))
			mainorderAPISource, triggerOrderAPISource, httpCode = self.UpdateLimitOrder(ctx, maxResult)
			if HttpSuccess(httpCode) {
				break
			} else if count == 3 {
				break
			}
			count += 1
			maxResult += 15
			logrus.Warnf("UpdateLocalOrderStatus attempt... %v", count)
		}

		// match order
		matchOrder := self.MatchOrderID(newOrder, mainorderAPISource, triggerOrderAPISource)

		if matchOrder.OrderID != 0 {
			logrus.Infof("replace && match %+v", *matchOrder)
			return matchOrder
		}

		if attempt == 3 {
			logrus.Warnln("after replace OTACloseOrder..., but match NewOrder fail....")
			return matchOrder
		}

		time.Sleep(time.Millisecond * time.Duration(attempt*1000))
		attempt += 1
		logrus.Warnf("fail to match NewOrder, attempt... %v", attempt)

	}

}

// 待測試
// 處理API 替換掛單, 新掛單批配  , 返回最後替代後的掛單
func (self *OrderOperation) ReplaceNormalOrder(OldOrder *UnitLimitOrder, NewOrder *UnitLimitOrder) *UnitLimitOrder {

	attempt := 0
	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
		statusCode := self.OrderAPI.ReplaceOrder(ctx, fmt.Sprintf("%v", OldOrder.OrderID), NewOrder)

		if HttpSuccess(statusCode) {
			logrus.Infoln("TimeOut Normal Order Success to Replace Order")
			break
		} else if attempt == 3 {
			logrus.Warnln("TimeOut Normal Order Fail to Replace Order")
			break
		}

		attempt += 1
		logrus.Warnf("TimeOut Normal Order Replace attempt... %v", attempt)
		time.Sleep(time.Millisecond * time.Duration(attempt*100))
	}

	var mainorderAPISource, triggerOrderAPISource map[int64]*UnitLimitOrder
	statusCode := 0
	mainorderAPISource = make(map[int64]*UnitLimitOrder)
	triggerOrderAPISource = make(map[int64]*UnitLimitOrder)
	attempt = 0
	maxResult := 45
	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
		mainorderAPISource, triggerOrderAPISource, statusCode = self.UpdateLimitOrder(ctx, maxResult)

		if HttpSuccess(statusCode) {
			logrus.Infoln("ReplaceNormalOrder Success to Replace Order")
			break
		} else if attempt == 3 {
			logrus.Warnln("ReplaceNormalOrder Fail to Replace Order")
			break
		}

		attempt += 1
		maxResult += 15
		logrus.Warnf("ReplaceNormalOrder Replace attempt... %v", attempt)
		time.Sleep(time.Millisecond * time.Duration(100*attempt))
	}

	matchOrder := self.MatchOrderID(NewOrder, mainorderAPISource, triggerOrderAPISource)

	logrus.Infof("replace && match %+v", *matchOrder)

	return matchOrder

}

func (self *OrderOperation) ReplaceNormalOrderToMarketOrder(OldOrder *UnitLimitOrder, NewOrder *UnitLimitOrder) *UnitLimitOrder {

	attempt := 0
	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
		statusCode := self.OrderAPI.ReplaceToMarketOrder(ctx, fmt.Sprintf("%v", OldOrder.OrderID), NewOrder)

		if HttpSuccess(statusCode) {
			logrus.Infoln("TimeOut Normal Order Success to Replace Order")
			break
		} else if attempt == 3 {
			logrus.Warnln("TimeOut Normal Order Fail to Replace Order")
			break
		}

		attempt += 1
		logrus.Warnf("TimeOut Normal Order Replace attempt... %v", attempt)
		time.Sleep(time.Millisecond * time.Duration(attempt*100))
	}

	var mainorderAPISource, triggerOrderAPISource map[int64]*UnitLimitOrder
	statusCode := 0
	mainorderAPISource = make(map[int64]*UnitLimitOrder)
	triggerOrderAPISource = make(map[int64]*UnitLimitOrder)
	attempt = 0
	maxResult := 45
	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
		mainorderAPISource, triggerOrderAPISource, statusCode = self.UpdateLimitOrder(ctx, maxResult)

		if HttpSuccess(statusCode) {
			logrus.Infoln("ReplaceNormalOrder Success to Replace Order to Market Order")
			break
		} else if attempt == 3 {
			logrus.Warnln("ReplaceNormalOrder Fail to Replace Order")
			break
		}

		attempt += 1
		maxResult += 15
		logrus.Warnf("ReplaceNormalOrder Replace attempt... %v", attempt)
		time.Sleep(time.Millisecond * time.Duration(100*attempt))
	}

	matchOrder := self.MatchOrderID(NewOrder, mainorderAPISource, triggerOrderAPISource)

	logrus.Infof("replace && match %+v", *matchOrder)

	return matchOrder

}

func (self *OrderOperation) DeleteOTAFullOrder(OldOrder *UnitLimitOrder) (statusCode int) {

	attempt := 0

	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
		statusCode = self.OrderAPI.DeleteOrder(ctx, fmt.Sprintf("%v", OldOrder.OrderID))

		if HttpSuccess(statusCode) {
			logrus.Infof("TimeOutFullOTAOrder Success to Delete Old OTAFullOrder %v", statusCode)
			return statusCode
		} else if attempt == 3 {
			logrus.Warnf("TimeOutFullOTAOrder Fail to Delete Old OTAFullOrder %v", statusCode)
			return statusCode
		}

		attempt += 1
		logrus.Warnf("TimeOutFullOTAOrder Delete Old OTAFullOrder attempt... %v", attempt)
		time.Sleep(time.Millisecond * time.Duration(500*attempt))
	}
}

func (self *OrderOperation) DeleteNormalOrder(OldOrder *UnitLimitOrder) (statusCode int) {

	attempt := 0

	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
		statusCode = self.OrderAPI.DeleteOrder(ctx, fmt.Sprintf("%v", OldOrder.OrderID))

		if HttpSuccess(statusCode) {
			logrus.Infof("TimeOutNormalOrder Success to Delete Old NormalOrder %v", statusCode)
			return statusCode
		} else if attempt == 3 {
			logrus.Warnf("TimeOutNormalOrder Fail to Delete Old NormalOrder %v", statusCode)
			return statusCode
		}

		attempt += 1
		logrus.Warnf("TimeOutNormalOrder Delete Old NormalOrder attempt... %v", attempt)
		time.Sleep(time.Millisecond * time.Duration(500*attempt))
	}
}

// 修改OTA OPNE , CLOSE orderID會改變 , 但內容一樣  無法同時修改,  這邊傳入舊的 是為了重新match full OTA
func (self *OrderOperation) ReplaceOpenOTAOrder(OldMainOrder *UnitLimitOrder, OlDTriggerOlder *UnitLimitOrder, NewMainOrder *UnitLimitOrder) (matchOrder *UnitLimitOrder, triggerOrder *UnitLimitOrder) {

	price, _ := strconv.ParseFloat(NewMainOrder.Price, 64)
	newPrice := fmt.Sprintf("%v", FloorFloat(price, 2))

	newOrder := &UnitLimitOrder{
		Symbol:     NewMainOrder.Symbol,
		Price:      newPrice,
		OrderType:  NewMainOrder.OrderType,
		Quantity:   NewMainOrder.Quantity,
		CreateTime: NewMainOrder.CreateTime}

	// 防止兩單當相同
	newD, _ := decimal.NewFromString(newOrder.Price)
	oldD, _ := decimal.NewFromString(OldMainOrder.Price)
	if newD.LessThanOrEqual(oldD) {
		logrus.Infoln("replace order equal old order, new one + 0.01")
		newD.Add(decimal.NewFromFloat(0.01))
	} else if newD.LessThanOrEqual(oldD) {
		logrus.Infoln("replace order equal old order, new one - 0.01")
		newD.Add(decimal.NewFromFloat(-0.01))
	}
	newOrder.Price = newD.String()
	OldMainOrder.Price = oldD.String()

	attempt := 0

	logrus.Infof("Prepare replace OTA OPEN old order: %+v", OldMainOrder)
	logrus.Infof("TO New order: %+v", newOrder)

	// API replace OTA order
	for {

		ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
		httpStatusCode := self.OrderAPI.ReplaceOpenOTAOrder(ctx, fmt.Sprintf("%v", OldMainOrder.OrderID), newOrder)

		if HttpSuccess(httpStatusCode) {
			logrus.Infoln("success to replace OTAOpenOrder")
			break
		} else if attempt == 2 {
			logrus.Warnln("fail to Replace OTAOpenOrder, if exceed... can't not confirm status ... 200 or other ...") // 失敗直接跳出 沒有match必要
			// return nil
			//context exceed 有可能已經生效, 或是其實半路被filled了 不確定 繼續往下走看看 match 看看 失敗表示replace失敗
			break
		}
		attempt += 1
		logrus.Warnf("ReplaceOTAOpenOrder attempt... %v", attempt)
		time.Sleep(time.Millisecond * time.Duration(100*attempt))

	}

	// match失敗 重新嘗試
	attempt = 0
	for {
		var mainorderAPISource, triggerOrderAPISource map[int64]*UnitLimitOrder
		httpCode := 0
		mainorderAPISource = make(map[int64]*UnitLimitOrder)
		triggerOrderAPISource = make(map[int64]*UnitLimitOrder)
		// Get OrderStatus From API

		count := 0 // API updateOrderStatus, 若失敗 重新嘗試
		maxResult := 45
		for {
			ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(1500))
			mainorderAPISource, triggerOrderAPISource, httpCode = self.UpdateLimitOrder(ctx, maxResult)
			if HttpSuccess(httpCode) {
				break
			} else if count == 3 {
				break
			}
			count += 1
			maxResult += 15
			logrus.Warnf("UpdateLocalOrderStatus attempt... %v", count)
		}

		// match order
		matchOrder, triggerOrder = self.MatchFullOTAOrderID(NewMainOrder, OlDTriggerOlder, mainorderAPISource, triggerOrderAPISource)

		if matchOrder.OrderID != 0 {
			logrus.Infof("replace OTA opne && match %+v", *matchOrder)
			return matchOrder, triggerOrder
		}

		if attempt == 3 {
			logrus.Warnln("after replace OTAOpenOrder..., but match NewOrder fail....")
			return nil, nil
		}

		time.Sleep(time.Millisecond * time.Duration(attempt*500))
		attempt += 1
		logrus.Warnf("fail to match NewOrder, attempt... %v", attempt)

	}

}

type SymbolTransaction struct {
	Symbol             string
	TransactionPrice   []decimal.Decimal
	Earns              []decimal.Decimal
	SumEarns           decimal.Decimal
	Performance        decimal.Decimal // 香對於referPrice 百分比績效
	WinRate            decimal.Decimal
	FullTransactionNum decimal.Decimal // 完整 買入買出的循環 次數\
	referPrice         decimal.Decimal // 使用第一次購買的金額作為參考
}

func (self *SymbolTransaction) Append(unit TransactionUnit) {

	transactionPrice := decimal.NewFromFloat(unit.NetAmount)
	self.TransactionPrice = append(self.TransactionPrice, transactionPrice)

	if self.Symbol == "" {
		self.Symbol = unit.TransactionItem.Instrument.Symbol
	}

	if self.referPrice.Equal(decimal.Zero) {
		self.referPrice = transactionPrice
	}

}

// 必須滿足當日沖銷所有艙位 才能進行統計
func (self *SymbolTransaction) Static() {

	self.Earns = make([]decimal.Decimal, 0)
	sum := decimal.Zero
	for _, eachEarns := range self.TransactionPrice {
		sum = sum.Add(eachEarns)

		// 累積價差在價格20%以內 表示close order
		if sum.Abs().LessThan(eachEarns.Mul(decimal.NewFromFloat(0.2)).Abs()) {
			self.Earns = append(self.Earns, sum)
			self.FullTransactionNum = self.FullTransactionNum.Add(decimal.NewFromInt(1))
			sum = decimal.Zero
		}
	}

	win := decimal.Zero
	for _, val := range self.Earns {
		self.SumEarns = self.SumEarns.Add(val)
		if val.GreaterThan(decimal.Zero) {
			win = win.Add(decimal.NewFromInt(1))
		}
	}

	self.WinRate = win.Div(decimal.NewFromInt(int64(len(self.Earns)))).Mul(decimal.NewFromFloat(100)).RoundFloor(2)
	self.Performance = self.SumEarns.Div(self.referPrice).Mul(decimal.NewFromInt(100)).RoundFloor(2)
}

func (self *OrderOperation) GetPeriodTransactionPerformance(startDate string, endDate string, noStatic []string) (AllSymbolStatic map[string]*SymbolTransaction) {

	AllSymbolStatic = make(map[string]*SymbolTransaction)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	parm := UnitHisTransaction{TransactionType: "ALL", Symbol: "", StartDate: startDate, EndDate: endDate}
	httpCode, res := self.OrderAPI.GetTransactionHistory(ctx, parm)
	if HttpSuccess(httpCode) {

	outer:
		for _, val := range *res {

			if val.TransactionItem.Instrument.Symbol != "" {

				for _, noSymbol := range noStatic {
					if noSymbol == val.TransactionItem.Instrument.Symbol {
						continue outer
					}
				}

				_, ok := AllSymbolStatic[val.TransactionItem.Instrument.Symbol]
				if !ok {
					AllSymbolStatic[val.TransactionItem.Instrument.Symbol] = new(SymbolTransaction)
				}
				AllSymbolStatic[val.TransactionItem.Instrument.Symbol].Append(val)
			}

		}

		for _, symbol := range AllSymbolStatic {
			symbol.Static()
		}

	}

	return AllSymbolStatic

}

func (self *OrderOperation) GetPosition() (currentPosition []string) {
	symbol := []string{}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	httpCode, res := self.OrderAPI.GetAccountInfo(ctx)
	if HttpSuccess(httpCode) {

		for _, val := range res.SecuritiesAccount.Positions {

			symbol = append(symbol, val.Instrument.Symbol)
		}
	}

	return symbol

}
