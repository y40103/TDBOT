package utils

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type LocalOrderStatus struct {
	redisCli   *redis.Client
	pipe       redis.Pipeliner
	LastUpdate time.Time
}

func (self *LocalOrderStatus) AddOrderOperationNum(num int) {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	if self.pipe == nil {
		self.pipe = self.redisCli.Pipeline()
	}
	key := "OrderOperationNum"
	val := self.redisCli.Get(context.Background(), key)
	if val.Val() == "" {
		res0 := self.redisCli.Set(context.Background(), key, num, 15*3600*time.Second) // 壽命15小時
		logrus.Infoln("AddOrderOperationNum: SetOrderOperationNum ", res0)
		return
	}

	res1 := self.redisCli.IncrBy(context.Background(), key, int64(num))
	logrus.Infoln("AddOrderOperationNum: SetOrderOperationNum += ", res1)

}

// 回傳 order operation 次數
func (self *LocalOrderStatus) GetOrderOperationNum() int {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	if self.pipe == nil {
		self.pipe = self.redisCli.Pipeline()
	}
	key := "OrderOperationNum"
	val := self.redisCli.Get(context.Background(), key)
	if val.Val() == "" {
		res0 := self.redisCli.Set(context.Background(), key, 0, 15*3600*time.Second) // 壽命15小時
		logrus.Infoln("AddOrderOperationNum: SetOrderOperationNum ", res0)
		return 0
	}

	intVal, err := val.Int()

	if err != nil {
		logrus.Warnln("fail to get Order Operation Num")
	}

	return intVal

}

// 添加Order
// redis db: 1
// 所有OrderID 放於 Set AllOrderID
// Set 存放所有OrderID >>>  查詢 smembers AllOrderID
// Hash 放置各Order 屬性  HashKey 為 OrderID >> 查詢 Hmget <OrderID> <field1> <field2>
func (self *LocalOrderStatus) CreateOrderStatus(ctx context.Context, order UnitLimitOrder) {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	if self.pipe == nil {
		self.pipe = self.redisCli.Pipeline()
	}
	if order.OrderID == 0 {
		logrus.Warningln("orderID is ZERO")
		panic("orderID is ZERO")
	}
	logrus.Infoln("Start to Add OrderStatus")

	trackKey := "TrackingSymbol"
	res0 := self.pipe.SAdd(ctx, trackKey, order.Symbol)
	logrus.Infoln("PIPELINE ADD:", res0)
	key := order.Symbol + "_" + "AllOrderID"
	res1 := self.pipe.SAdd(ctx, key, order.OrderID)
	logrus.Infoln("PIPELINE ADD:", res1)
	res2 := self.pipe.HSet(ctx, fmt.Sprintf("%v", order.OrderID),
		"Symbol", order.Symbol,
		"OrderType", order.OrderType,
		"OrderID", order.OrderID,
		"Quantity", order.Quantity,
		"Price", order.Price,
		"Status", order.Status,
		"CreateTime", order.CreateTime,
		"Cancelable", order.Cancelable,
		"Editable", order.Editable,
		"Description", order.Description,
	)
	logrus.Infoln("PIPELINE ADD:", res2)

	keyOTA := order.Symbol + "_" + "OTA"
	res5 := self.pipe.Set(ctx, keyOTA, false, 0)
	logrus.Infoln("PIPELINE ADD:", res5)

	_, err := self.pipe.Exec(ctx)
	if err != nil {
		logrus.Warningln(err)
		panic(err)
	}

	self.pipe = nil

	logrus.Infoln("Commit PIPELINE")
	logrus.Infoln("Success to Add OrderStatus")

	allOrderID := self.GetAllOrderID(ctx, order.Symbol)
	logrus.Infof("After create: confirm AllOrderID %v", allOrderID)

	self.AddOrderOperationNum(1)

}

func (self *LocalOrderStatus) SetOTAStatus(ctx context.Context, Symbol string) {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}

	self.redisCli.Set(ctx, Symbol+"_OTA", true, 0)
	logrus.Infoln("Set OTA Tag")

}

// OTA MAIN ORDER , 就直接刪掉 重新創建 , 因若replace OTA main order, 它會把原本的 連trigger狀態改為replace, 等同事把原本的刪掉 新增一組新的
func (self *LocalOrderStatus) ReplaceOTATriggerOrder(ctx context.Context, NewTriggerOrder *UnitLimitOrder) {

	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	if self.pipe == nil {
		self.pipe = self.redisCli.Pipeline()
	}

	if NewTriggerOrder.OrderID == 0 {
		logrus.Warningln("orderID is ZERO")
		panic("orderID is ZERO")
	}
	logrus.Infoln("ReplaceOTATriggerOrder....")
	_, Close := self.GetOpenCloseOrder(ctx, NewTriggerOrder.Symbol)

	res0 := self.pipe.Del(ctx, fmt.Sprintf("%v", Close.OrderID))
	logrus.Infoln("PIPELINE ADD:", res0)

	res1 := self.pipe.HSet(ctx, fmt.Sprintf("%v", NewTriggerOrder.OrderID),
		"Symbol", NewTriggerOrder.Symbol,
		"OrderType", NewTriggerOrder.OrderType,
		"OrderID", NewTriggerOrder.OrderID,
		"Quantity", NewTriggerOrder.Quantity,
		"Price", NewTriggerOrder.Price,
		"Status", NewTriggerOrder.Status,
		"CreateTime", NewTriggerOrder.CreateTime,
		"Cancelable", NewTriggerOrder.Cancelable,
		"Editable", NewTriggerOrder.Editable,
		"Description", NewTriggerOrder.Description)
	logrus.Infoln("PIPELINE ADD:", res1)

	_, err := self.pipe.Exec(ctx)
	if err != nil {
		logrus.Warningln(err)
		panic(err)
	}
	self.pipe = nil

	logrus.Infoln("Commit PIPELINE")
	logrus.Infoln("Success to Add OrderStatus")
	self.AddOrderOperationNum(1)
}

func (self *LocalOrderStatus) ReplaceOrder(ctx context.Context, NewOrder *UnitLimitOrder) {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	if self.pipe == nil {
		self.pipe = self.redisCli.Pipeline()
	}

	logrus.Infoln("ReplaceOrder....")

	if NewOrder.OrderID == 0 {
		logrus.Warningln("orderID is ZERO")
		panic("orderID is ZERO")
	}

	OTA := self.OrderIsOTA(ctx, NewOrder.Symbol)

	if OTA {
		logrus.Warnln("CAN'T REPLACE OTA ORDER")
		panic("CAN'T REPLACE OTA ORDER")
	}

	self.pipe = self.redisCli.Pipeline()

	_, CurrentWorkingOrder := self.GetOrderStage(ctx, NewOrder.Symbol)

	res0 := self.pipe.Del(ctx, fmt.Sprintf("%v", CurrentWorkingOrder.OrderID))
	logrus.Infoln("PIPELINE ADD:", res0)
	key := CurrentWorkingOrder.Symbol + "_AllOrderID"
	res1 := self.pipe.SRem(ctx, key, CurrentWorkingOrder)
	logrus.Infoln("PIPELINE ADD:", res1)

	_, err := self.pipe.Exec(ctx)
	if err != nil {
		logrus.Warningln(err)
		panic(err)
	}
	self.pipe = nil

	logrus.Infof("replace order %v to %v", CurrentWorkingOrder.OrderID, NewOrder.OrderID)
	self.AddOrderOperationNum(1)
}

func (self *LocalOrderStatus) CreateOTAOrderStatus(ctx context.Context, MainOrder UnitLimitOrder, TriggerOrder UnitLimitOrder) {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	if self.pipe == nil {
		self.pipe = self.redisCli.Pipeline()
	}
	if MainOrder.OrderID == 0 || TriggerOrder.OrderID == 0 {
		logrus.Warningln("orderID is ZERO")
		panic("orderID is ZERO")
	}
	logrus.Infoln("Start to Add OrderStatus")

	trackKey := "TrackingSymbol"
	res00 := self.pipe.SAdd(ctx, trackKey, MainOrder.Symbol)
	logrus.Infoln("PIPELINE ADD:", res00)
	res01 := self.pipe.SAdd(ctx, trackKey, TriggerOrder.Symbol)
	logrus.Infoln("PIPELINE ADD:", res01)

	key := MainOrder.Symbol + "_" + "AllOrderID"
	res0 := self.pipe.SAdd(ctx, key, MainOrder.OrderID)
	logrus.Infoln("PIPELINE ADD:", res0)
	res1 := self.pipe.SAdd(ctx, key, TriggerOrder.OrderID)
	logrus.Infoln("PIPELINE ADD:", res1)
	res3 := self.pipe.HSet(ctx, fmt.Sprintf("%v", MainOrder.OrderID),
		"Symbol", MainOrder.Symbol,
		"OrderType", MainOrder.OrderType,
		"OrderID", MainOrder.OrderID,
		"Quantity", MainOrder.Quantity,
		"Price", MainOrder.Price,
		"Status", MainOrder.Status,
		"CreateTime", MainOrder.CreateTime,
		"Cancelable", MainOrder.Cancelable,
		"Editable", MainOrder.Editable,
		"Description", MainOrder.Description,
	)
	logrus.Infoln("PIPELINE ADD:", res3)
	res4 := self.pipe.HSet(ctx, fmt.Sprintf("%v", TriggerOrder.OrderID),
		"Symbol", TriggerOrder.Symbol,
		"OrderType", TriggerOrder.OrderType,
		"OrderID", TriggerOrder.OrderID,
		"Quantity", TriggerOrder.Quantity,
		"Price", TriggerOrder.Price,
		"Status", TriggerOrder.Status,
		"CreateTime", TriggerOrder.CreateTime,
		"Cancelable", TriggerOrder.Cancelable,
		"Editable", TriggerOrder.Editable,
		"Description", TriggerOrder.Description,
	)
	logrus.Infoln("PIPELINE ADD:", res4)

	keyOTA := MainOrder.Symbol + "_" + "OTA"
	res5 := self.pipe.Set(ctx, keyOTA, true, 0)
	logrus.Infoln("PIPELINE ADD:", res5)

	_, err := self.pipe.Exec(ctx)
	if err != nil {
		logrus.Warningln(err)
		panic(err)
	}

	self.pipe = nil

	logrus.Infoln("Commit PIPELINE")
	logrus.Infoln("Success to Add OrderStatus")
	self.AddOrderOperationNum(2)
}

// 若ota key 不存在 返回 false, ota key 存在 又是 ota true, ota key存在 非 ota , false
func (self *LocalOrderStatus) OrderIsOTA(ctx context.Context, symbol string) bool {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	key := symbol + "_" + "OTA"
	res := self.redisCli.Get(ctx, key)

	if res != nil { // 存在key
		exists, err := res.Bool()
		if err != nil {
			logrus.Warningln(err)
			return false
		}
		return exists
	}

	return false
}

// 統計該symbol所有orderID 數量 redis key: <Symbol>_AllOrderID
func (self *LocalOrderStatus) GetOrderStatusSize(ctx context.Context, symbol string) int {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}

	key := symbol + "_" + "AllOrderID"
	size, err := self.redisCli.SCard(ctx, key).Result()
	if err != nil {
		logrus.Warningln(err)
	}
	return int(size)
}

// 取得 ota open order & close order
func (self *LocalOrderStatus) GetOTAMainTriggerStatus(ctx context.Context, Symbol string) (MainOrder *UnitLimitOrder, TriggerOrder *UnitLimitOrder) {

	allOrderID := self.GetAllOrderID(ctx, Symbol)

	MainOrder = new(UnitLimitOrder)
	TriggerOrder = new(UnitLimitOrder)

	for _, val := range allOrderID {
		intID, _ := strconv.Atoi(val)
		order := self.GetOrderStatus(ctx, int64(intID))
		if order.OrderType > 0 {
			MainOrder = order
		} else if order.OrderType < 0 {
			TriggerOrder = order
		}
	}

	return MainOrder, TriggerOrder

}

func (self *LocalOrderStatus) GetOrderStatus(ctx context.Context, OrderID int64) *UnitLimitOrder {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}

	res := self.redisCli.HMGet(ctx, fmt.Sprintf("%d", OrderID),
		"Symbol",
		"OrderType",
		"OrderID",
		"Quantity",
		"Price",
		"Status",
		"CreateTime",
		"Cancelable",
		"Editable",
		"Description")
	logrus.Infof("get order interface: %+v", res)

	myorder := new(UnitLimitOrder)
	err := res.Scan(myorder)
	logrus.Infoln("Get Order status ", myorder)
	if err != nil {
		logrus.Warnln(err)
		panic(err)
	}

	// redis 無時間類型 另外補
	resp, err := res.Result()

	if err != nil {
		logrus.Warnln(err)
		panic(err)
	}

	if resp[2] == nil {
		logrus.Warnln("No key Exists")
		return nil
	}

	// 2006-01-02T15:04:05.000000000-07:00
	mytime, err := time.Parse("2006-01-02T15:04:05-07:00", resp[6].(string))
	if err != nil {
		logrus.Warnln(err)
	}
	mytime = mytime.In(Loc)
	myorder.CreateTime = mytime
	logrus.Infof("Get OrderID %v: %+v", OrderID, *myorder)

	return myorder

}

// 取得單一symbol 的所有orderID
func (self *LocalOrderStatus) GetAllOrderID(ctx context.Context, Symbol string) []string {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	key := Symbol + "_AllOrderID"
	res, err := self.redisCli.SMembers(ctx, key).Result()
	if err != nil {
		logrus.Warnln(err)
		panic(err)
	}
	return res
}

// 刪除單一orderID
// srm <Symbol>__AllOrderID <orderID> ,del <OrderID>,  若為最後一個orderID 相依刪除 OpenOrderLock srm <Symbol> , del <Symbol>_OTA , srm TrackingSymbol <symbol>,
func (self *LocalOrderStatus) ClearOrderStatus(ctx context.Context, OrderID string) {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	if self.pipe == nil {
		self.pipe = self.redisCli.Pipeline()
	}

	Symbol, _ := self.redisCli.HGet(ctx, OrderID, "Symbol").Result()
	allorderIDbefore := self.GetAllOrderID(ctx, Symbol)
	logrus.Infof("before clear = %v symbol = %v", allorderIDbefore, Symbol)

	res1 := self.pipe.SRem(ctx, Symbol+"_AllOrderID", OrderID) // 刪除該OrderID追蹤
	logrus.Infoln("PIPELINE SRem Filed:", res1)

	res2 := self.pipe.Del(ctx, OrderID) // 刪除該OrderID 內容
	logrus.Infoln("PIPELINE Del :", res2)

	if len(allorderIDbefore) <= 1 {
		//res3 := self.pipe.Del(ctx, Symbol+"_OTA")
		//logrus.Infoln("PIPELINE Del:", res3)
		res4 := self.pipe.SRem(ctx, "TrackingSymbol", Symbol) // 所有orderID 都取消後, 解除 對該symbol追蹤
		logrus.Infoln("PIPELINE SRem Field:", res4)
		res6 := self.pipe.Del(ctx, Symbol+"_AllOrderID") // 該symbol orderID都取消後 刪除該key
		logrus.Infoln("PIPELINE Del ", res6)
		res10 := self.pipe.SRem(ctx, "OpenOrderLock", Symbol)
		logrus.Infoln("PIPELINE SRem Field:", res10)
	}

	_, err := self.pipe.Exec(ctx)
	if err != nil {
		logrus.Warningln(err)
		panic(err)
	}

	self.pipe = nil
	logrus.Infoln("Commit PIPELINE")
	logrus.Infoln("Success to Clear OrderStatus")

	allorderID := self.GetAllOrderID(ctx, Symbol)
	logrus.Infof("After Clear %v, now: %v", OrderID, allorderID)

}

// 刪除某Symbol 所有order狀態
// del <Symbol>_allorderID , del <iter 某Symbol所有OrderID> ,del <Symbol>_OTA, srm TrackingSymbol <Symbol> , srm OpenOrderLock <Symbol>
func (self *LocalOrderStatus) ClearSymbolAllOrderStatus(ctx context.Context, Symbol string) {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	if self.pipe == nil {
		self.pipe = self.redisCli.Pipeline()
	}

	allorderID := self.GetAllOrderID(ctx, Symbol)

	res1 := self.pipe.Del(ctx, Symbol+"_AllOrderID")
	logrus.Infoln("PIPELINE Del:", res1)
	for _, orderid := range allorderID {
		res2 := self.pipe.Del(ctx, orderid)
		logrus.Infoln("PIPELINE Del:", res2)
	}
	res3 := self.pipe.Del(ctx, Symbol+"_OTA")
	logrus.Infoln("PIPELINE Del:", res3)

	res4 := self.pipe.SRem(ctx, "TrackingSymbol", Symbol) // 刪除追蹤標籤
	logrus.Infoln("PIPELINE Del SRem field:", res4)

	res5 := self.pipe.SRem(ctx, "OpenOrderLock", Symbol)
	logrus.Infoln("PIPELINE SRem Field:", res5)

	_, err := self.pipe.Exec(ctx)
	if err != nil {
		logrus.Warningln(err)
		panic(err)
	}

	self.pipe = nil
	logrus.Infoln("Commit PIPELINE")
	logrus.Infoln("Success to Clear OrderStatus")
}

func (self *LocalOrderStatus) GetTrackingSymbol(ctx context.Context) (trackingSymbol []string) {

	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	key := "TrackingSymbol"
	res, err := self.redisCli.SMembers(ctx, key).Result()
	if err != nil {
		logrus.Warnln(err)
		panic(err)
	}
	return res

}

// 回傳 open order, close order , 若目前只有open order, closeOrder部份回傳nil
func (self *LocalOrderStatus) GetOpenCloseOrder(ctx context.Context, Symbol string) (openOrder *UnitLimitOrder, closeOrder *UnitLimitOrder) {

	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	logrus.Infoln("GetOpenClose: dependency GetAllOrderID")
	AllorderID := self.GetAllOrderID(ctx, Symbol)
	logrus.Infoln("GetOpenCloseOrder: Current AllOrderID: ", AllorderID)
	if AllorderID == nil {
		return nil, nil
	} else if len(AllorderID) == 1 { // 只有一個 表示只有open
		id, err := strconv.Atoi(AllorderID[0])
		if err != nil {
			logrus.Warnln(err)
			return nil, nil
		}
		logrus.Infoln("GetOpenClose: dependency GetOrderStatus1")
		order := self.GetOrderStatus(ctx, int64(id))
		return order, nil
	} else if len(AllorderID) > 2 {
		msg := fmt.Sprintf("GetOpenClose: %v AllOrderID > 2", Symbol)
		panic(msg)
	}

	allOrder := make([]*UnitLimitOrder, 0)

	for _, orderID := range AllorderID {
		id, err := strconv.Atoi(orderID)
		if err != nil {
			logrus.Warnln(err)
			return nil, nil
		}
		logrus.Infoln("GetOpenClose: dependency GetOrderStatus2")
		order := self.GetOrderStatus(ctx, int64(id))
		if order == nil || order.Status == "REPLACED" {
			continue
		}
		allOrder = append(allOrder, order)

	}

	if len(allOrder) != 2 {
		logrus.Warnf("GetOpenCloseOrder: all order: %v", allOrder)
		if len(allOrder) == 0 {
			return nil, nil
		}
	}

	if allOrder[0].CreateTime.After(allOrder[1].CreateTime) {
		allOrder[0], allOrder[1] = allOrder[1], allOrder[0]
	}

	return allOrder[0], allOrder[1]

}

// 只能用來更新 不能用於新增 不會有ota tag
// 更新追蹤symbol的所有order狀態
func (self *LocalOrderStatus) UpdateLocalOrderStatusFromAPIResponse(ctx context.Context, ResponseMainOrder map[int64]*UnitLimitOrder, ResponseTriggerOrder map[int64]*UnitLimitOrder) bool {

	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	if self.pipe == nil {
		self.pipe = self.redisCli.Pipeline()
	}

	TrackingOrder := make(map[string]*UnitLimitOrder)
	TargetID := make([]string, 0)

	allsymbol := self.GetTrackingSymbol(ctx)

	logrus.Infof("Prepare to update symbol: %v num: %v", allsymbol, len(allsymbol))

	for _, symbol := range allsymbol {
		res := self.GetAllOrderID(ctx, symbol)

		TargetID = append(TargetID, res...)
		logrus.Infoln("symbol: ", symbol, "orderID: ", res)
	}

	self.pipe = self.redisCli.Pipeline()

	// 目標order id , 先查找main order, 找不到再找 trigger
	for _, orderID := range TargetID {

		intOrderID, err := strconv.Atoi(orderID)
		if err != nil {
			logrus.Warnln(err)
		}

		order, ok := ResponseMainOrder[int64(intOrderID)]

		if ok {
			TrackingOrder[orderID] = order
		} else if order, ok = ResponseTriggerOrder[int64(intOrderID)]; ok {

			if ok {
				TrackingOrder[orderID] = order
			} else {

				logrus.Warnln("CAN'T GET TRACKING ORDER FROM API RESPONSE ")
				logrus.Warnln("Miss ORDER ID ", orderID)
				continue

			}

		}

		if order == nil {
			logrus.Warnf("Can't Get OrderID %v", orderID)
			return false
		}

		res := self.pipe.HSet(ctx, orderID,
			"Symbol", order.Symbol,
			"OrderType", order.OrderType,
			"OrderID", order.OrderID,
			"Quantity", order.Quantity,
			"Price", order.Price,
			"Status", order.Status,
			"CreateTime", order.CreateTime,
			"Cancelable", order.Cancelable,
			"Editable", order.Editable,
			"Description", order.Description)
		logrus.Infoln("PIPELINE ADD:", res)

	}

	_, err := self.pipe.Exec(ctx)
	if err != nil {
		logrus.Warnln(err)
		panic(err)
	}
	self.pipe = nil
	logrus.Infoln("Commit PIPELINE")
	logrus.Infoln("Update Order Status to local")
	self.LastUpdate = time.Now().In(Loc)
	return true

}

// 1 > open working , 2 > close working ,  0 > all fiiled , -1 else
// stageOrder 表示當前working order
func (self *LocalOrderStatus) GetOrderStage(ctx context.Context, symbol string) (Stage int, StageOrder *UnitLimitOrder) {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}

	ota := self.OrderIsOTA(ctx, symbol)
	logrus.Infoln("Get Order Stage....")
	openOrder, closeOrder := self.GetOpenCloseOrder(ctx, symbol)

	// Symbol order不存在
	if openOrder == nil && closeOrder == nil {
		logrus.Warnln("NO SYMBOL ORDER: ", symbol)
		return -1, nil
	}

	if ota {

		// 曾經出現過 ota 兩個都是working, 有可能open讀近來後 剛好filled , close近來也是working

		if openOrder != nil && closeOrder.Status == "AWAITING_PARENT_ORDER" {
			logrus.Infof("OTA = %v", ota)
			logrus.Infof("%+v", openOrder)
			logrus.Infof("%+v", closeOrder)
			logrus.Infoln("stage = ", 1)
			return 1, openOrder
		} else if (closeOrder.Status == "WORKING" || closeOrder.Status == "QUEUED" || closeOrder.Status == "ACCEPTED" || closeOrder.Status == "AWAITING_UR_OUT") && closeOrder != nil {
			logrus.Infof("OTA = %v", ota)
			logrus.Infof("%+v", openOrder)
			logrus.Infof("%+v", closeOrder)
			logrus.Infoln("stage = ", 2)
			return 2, closeOrder
		}
	}

	if ota == false {

		// 非 ota 情況
		if (openOrder.Status == "WORKING" || openOrder.Status == "QUEUED" || openOrder.Status == "ACCEPTED" || openOrder.Status == "AWAITING_UR_OUT") && closeOrder == nil { // 一般order, 剛掛單 尚未賣出
			logrus.Infof("OTA = %v", ota)
			logrus.Infof("%+v", openOrder)
			logrus.Infof("%+v", closeOrder)
			logrus.Infoln("stage = ", 1)
			return 1, openOrder

		} else if openOrder.Status == "FILLED" && closeOrder == nil { // 一般order, 剛賣出 尚未掛單
			logrus.Infof("OTA = %v", ota)
			logrus.Infof("%+v", openOrder)
			logrus.Infof("%+v", closeOrder)
			logrus.Infoln("stage = ", 2)
			return 2, nil

		} else if openOrder.Status == "FILLED" && closeOrder.Status != "FILLED" { // 一般order, waiting close order
			logrus.Infof("OTA = %v", ota)
			logrus.Infof("%+v", openOrder)
			logrus.Infof("%+v", closeOrder)
			logrus.Infoln("stage = ", 2)
			return 2, closeOrder
		}

	}

	// 共有情況
	if openOrder.Status == "FILLED" && closeOrder.Status == "FILLED" { // 一般或ota 完成交易
		logrus.Infof("OTA = %v", ota)
		logrus.Infof("%+v", openOrder)
		logrus.Infof("%+v", closeOrder)
		logrus.Infoln("stage = ", 0)
		return 0, nil
	} else if openOrder.Status == "CANCELED" && closeOrder.Status == "CANCELED" { //可能一些意外導致 order 都被 canceled 同交易完成處理
		logrus.Infof("OTA = %v", ota)
		logrus.Infof("openOrder: %+v", openOrder)
		logrus.Infof("closeOrder: %+v", closeOrder)
		logrus.Infoln("stage = ", 0)
		return 0, nil
	}

	logrus.Warnln("ERROR STATUS ")
	logrus.Warnf("OTA = %v", ota)
	logrus.Warnf("open = %+v", openOrder)
	logrus.Warnf("close = %+v", closeOrder)

	return -1, nil
}

// stage = 100, 表示尚未過期  else 表示過期 , order 為過期掛單資訊
func (self *LocalOrderStatus) InspectTimeOut(ctx context.Context, Symbol string, Duration int) (stage int, order *UnitLimitOrder) {
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}

	stage, order = self.GetOrderStage(ctx, Symbol)

	if order != nil && !order.CreateTime.IsZero() {

		// expired
		if (order.CreateTime.Add(time.Second * time.Duration(Duration))).Before(time.Now()) {

			logrus.Infof("%v expired: stage %v order %+v", Symbol, stage, order)

			return stage, order
		}

		logrus.Infof("%v order is under valid duration, stage = %v", Symbol, stage)

	}

	return 100, nil

}

// 鎖住OpenLock 半後解除  1 == lock , nil == unlock
func (self *LocalOrderStatus) SetOpenOrderLock(ctx context.Context, Symbol string) {

	logrus.Infof("Start to Set %v OpenOrderLockStatus", Symbol)
	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}
	if self.pipe == nil {
		self.pipe = self.redisCli.Pipeline()
	}

	res1 := self.pipe.SAdd(ctx, "OpenOrderLock", Symbol)
	logrus.Infoln("PIPELINE: ", res1)
	res2 := self.pipe.Expire(ctx, "OpenOderLock", time.Hour*12)
	logrus.Infoln("PIPELINE: ", res2)

	_, err := self.pipe.Exec(ctx)
	if err != nil {
		logrus.Warningln(err)
		panic(err)
	}

	logrus.Warnf("Lock Symbol %v OpenOrder Capability 12hr", Symbol)

}

// not == nil , 表示沒有鎖住
func (self *LocalOrderStatus) GetOpenOrderLockStatus(ctx context.Context, Symbol string) (status bool) {

	logrus.Infof("Start to Get %v OpenOrderLockStatus", Symbol)

	if self.redisCli == nil {
		self.redisCli = GetRedis("1")
	}

	resp, err := self.redisCli.SMembers(ctx, "OpenOrderLock").Result()

	if err != nil {
		logrus.Warn(err)
	}

	for _, lockSymbol := range resp {

		if Symbol == lockSymbol {
			status = true
			break
		}

	}

	logrus.Infof("Get Symbol %v OpenOrderLockStatus: %v", Symbol, status)

	return status
}
