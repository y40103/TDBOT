package utils

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"
)

// Create or Update
func TestLocalOrderStatus_CreateOrderStatus_Commit(t *testing.T) {
	LocalOrder := LocalOrderStatus{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	NewOrder := UnitLimitOrder{
		Symbol:     "U",
		OrderType:  10,
		OrderID:    10705862626,
		Quantity:   5,
		Price:      "25.23",
		Status:     "WORKING",
		CreateTime: time.Now(),
		Editable:   true,
		Cancelable: true,
	}
	LocalOrder.CreateOrderStatus(ctx, NewOrder)

}

// Read
func TestLocalOrderStatus_GetOrderStatus(t *testing.T) {
	LocalOrder := LocalOrderStatus{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	//NewOrder := UnitLimitOrder{
	//	Symbol:     "U",
	//	OrderType:  10,
	//	OrderID:    10705862627,
	//	Quantity:   5,
	//	Price:      "25.23",
	//	Status:     "WORKING",
	//	CreateTime: time.Now(),
	//	Editable:   true,
	//	Cancelable: true,
	//}
	//LocalOrder.CreateOrderStatus(ctx, NewOrder)

	res := LocalOrder.GetOrderStatus(ctx, 10760941138)

	fmt.Println(*res)
}

// Delete
func TestLocalOrderStatus_ClearSymbolAllOrderStatus(t *testing.T) {
	LocalOrder := LocalOrderStatus{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	//mainOrder := UnitLimitOrder{
	//	Symbol:     "U",
	//	OrderType:  10,
	//	OrderID:    10705862626,
	//	Quantity:   5,
	//	Price:      "25.23",
	//	Status:     "WORKING",
	//	CreateTime: time.Now(),
	//	Editable:   true,
	//	Cancelable: true,
	//}
	//triggerOrder := UnitLimitOrder{
	//	Symbol:     "U",
	//	OrderType:  -10,
	//	OrderID:    10705862627,
	//	Quantity:   5,
	//	Price:      "25.23",
	//	Status:     "WORKING",
	//	CreateTime: time.Now(),
	//	Editable:   true,
	//	Cancelable: true,
	//}
	//LocalOrder.CreateOTAOrderStatus(ctx, mainOrder, triggerOrder)
	//
	//res1 := LocalOrder.GetOrderStatus(ctx, 10705862626)
	//res2 := LocalOrder.GetOrderStatus(ctx, 10705862627)
	//
	//fmt.Printf("%+v\n", *res1)
	//fmt.Printf("%+v\n", *res2)

	LocalOrder.ClearSymbolAllOrderStatus(ctx, "AAPL")
	LocalOrder.ClearSymbolAllOrderStatus(ctx, "TSLA")
	LocalOrder.ClearSymbolAllOrderStatus(ctx, "AMD")
	//res1 = LocalOrder.GetOrderStatus(ctx, 10705862626)
	//res2 = LocalOrder.GetOrderStatus(ctx, 10705862627)
	//
	//fmt.Printf("%+v\n", res1)
	//fmt.Printf("%+v\n", res2)
}

func TestLocalOrderStatus_GetOrderStatusSize(t *testing.T) {
	LocalOrder := LocalOrderStatus{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	//mainOrder := UnitLimitOrder{
	//	Symbol:     "U",
	//	OrderType:  10,
	//	OrderID:    10705862626,
	//	Quantity:   5,
	//	Price:      "25.23",
	//	Status:     "WORKING",
	//	CreateTime: time.Now(),
	//	Editable:   true,
	//	Cancelable: true,
	//}
	//triggerOrder := UnitLimitOrder{
	//	Symbol:     "U",
	//	OrderType:  -10,
	//	OrderID:    10705862627,
	//	Quantity:   5,
	//	Price:      "25.23",
	//	Status:     "WORKING",
	//	CreateTime: time.Now(),
	//	Editable:   true,
	//	Cancelable: true,
	//}
	//LocalOrder.CreateOTAOrderStatus(ctx, mainOrder, triggerOrder)
	//
	//res := LocalOrder.GetOrderStatusSize(ctx, mainOrder.Symbol)
	//fmt.Println("size=", res)
	//LocalOrder.ClearSymbolAllOrderStatus(ctx, mainOrder.Symbol)
	//res = LocalOrder.GetOrderStatusSize(ctx, mainOrder.Symbol)
	//fmt.Println("size=", res)
	//LocalOrder.ClearSymbolAllOrderStatus(ctx, mainOrder.Symbol)

	fmt.Println(LocalOrder.GetAllOrderID(ctx, "TSLA"))
}

func TestLocalOrderStatus_OrderIsOTA(t *testing.T) {
	LocalOrder := LocalOrderStatus{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	//mainOrder := UnitLimitOrder{
	//	Symbol:     "U",
	//	OrderType:  10,
	//	OrderID:    10705862626,
	//	Quantity:   5,
	//	Price:      "25.23",
	//	Status:     "WORKING",
	//	CreateTime: time.Now(),
	//	Editable:   true,
	//	Cancelable: true,
	//}
	//triggerOrder := UnitLimitOrder{
	//	Symbol:     "U",
	//	OrderType:  -10,
	//	OrderID:    10705862627,
	//	Quantity:   5,
	//	Price:      "25.23",
	//	Status:     "WORKING",
	//	CreateTime: time.Now(),
	//	Editable:   true,
	//	Cancelable: true,
	//}
	//LocalOrder.CreateOTAOrderStatus(ctx, mainOrder, triggerOrder)
	//
	//res := LocalOrder.OrderIsOTA(ctx, mainOrder.Symbol)
	//fmt.Println("OTA=", res)
	//
	//LocalOrder.ClearSymbolAllOrderStatus(ctx, mainOrder.Symbol)
	//
	//LocalOrder.CreateOrderStatus(ctx, mainOrder)
	//
	//res2 := LocalOrder.GetAllOrderID(ctx, mainOrder.Symbol)
	//fmt.Println(res2)
	//res3 := LocalOrder.OrderIsOTA(ctx, mainOrder.Symbol)
	//fmt.Println(res3)
	//LocalOrder.CreateOrderStatus(ctx, triggerOrder)
	//res2 = LocalOrder.GetAllOrderID(ctx, mainOrder.Symbol)
	//fmt.Println(res2)
	//res3 = LocalOrder.OrderIsOTA(ctx, mainOrder.Symbol)
	//fmt.Println(res3)
	//LocalOrder.ClearSymbolAllOrderStatus(ctx, mainOrder.Symbol)
	//res2 = LocalOrder.GetAllOrderID(ctx, mainOrder.Symbol)
	//fmt.Println(res2)
	//res3 = LocalOrder.OrderIsOTA(ctx, mainOrder.Symbol)
	//fmt.Println(res3)
	//LocalOrder.ClearSymbolAllOrderStatus(ctx, mainOrder.Symbol)
	fmt.Println(LocalOrder.OrderIsOTA(ctx, "MSFT"))
}

func TestLocalOrderStatus_GetOpenCloseOrder(t *testing.T) {
	LocalOrder := LocalOrderStatus{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	//LocalOrder.ClearSymbolAllOrderStatus(ctx, "U")
	//mainOrder := UnitLimitOrder{
	//	Symbol:     "U",
	//	OrderType:  10,
	//	OrderID:    10705862626,
	//	Quantity:   5,
	//	Price:      "30.23",
	//	Status:     "WORKING",
	//	CreateTime: time.Now(),
	//	Editable:   true,
	//	Cancelable: true,
	//}
	//LocalOrder.CreateOrderStatus(ctx, mainOrder)
	//size := LocalOrder.GetOrderStatusSize(ctx, "U")
	//fmt.Println("### ", size)
	//if size == 1 {
	//	openorder, _ := LocalOrder.GetOpenCloseOrder(ctx, "U")
	//	fmt.Printf("Open Order %+v\n", openorder)
	//}
	//
	//triggerOrder := UnitLimitOrder{
	//	Symbol:     "U",
	//	OrderType:  -10,
	//	OrderID:    10705862627,
	//	Quantity:   5,
	//	Price:      "100.23",
	//	Status:     "WORKING",
	//	CreateTime: time.Now(),
	//	Editable:   true,
	//	Cancelable: true,
	//}
	//
	//LocalOrder.CreateOrderStatus(ctx, triggerOrder)
	//size = LocalOrder.GetOrderStatusSize(ctx, "U")
	//if size == 2 {
	//	openOrder, closeOrder := LocalOrder.GetOpenCloseOrder(ctx, "U")
	//	fmt.Printf("Open Order %+v\n", openOrder)
	//	fmt.Printf("Close Order %+v\n", closeOrder)
	//}
	//LocalOrder.ClearSymbolAllOrderStatus(ctx, "U")
	openOrder, closeOrder := LocalOrder.GetOpenCloseOrder(ctx, "TSLA")
	fmt.Printf("Open Order %+v\n", openOrder)
	fmt.Printf("Close Order %+v\n", closeOrder)
}

func TestLocalOrderStatus_GetTrackingSymbol(t *testing.T) {
	LocalOrder := LocalOrderStatus{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	Order1 := UnitLimitOrder{
		Symbol:     "TSLA",
		OrderType:  10,
		OrderID:    10705862626,
		Quantity:   5,
		Price:      "125.23",
		Status:     "WORKING",
		CreateTime: time.Now(),
		Editable:   true,
		Cancelable: true,
	}

	LocalOrder.CreateOrderStatus(ctx, Order1)
	Order2 := UnitLimitOrder{
		Symbol:     "U",
		OrderType:  10,
		OrderID:    10705862628,
		Quantity:   5,
		Price:      "25.23",
		Status:     "WORKING",
		CreateTime: time.Now(),
		Editable:   true,
		Cancelable: true,
	}
	LocalOrder.CreateOrderStatus(ctx, Order2)

	res := LocalOrder.GetTrackingSymbol(ctx)

	fmt.Println("Tracking Symbol:")

	for _, val := range res {
		fmt.Println(val)
	}

	LocalOrder.ClearSymbolAllOrderStatus(ctx, Order1.Symbol)

	res = LocalOrder.GetTrackingSymbol(ctx)

	fmt.Println("Tracking Symbol:")

	for _, val := range res {
		fmt.Println(val)
	}

	res = LocalOrder.GetTrackingSymbol(ctx)

	fmt.Println("Tracking Symbol:")

	for _, val := range res {
		fmt.Println(val)
	}

	LocalOrder.ClearSymbolAllOrderStatus(ctx, Order2.Symbol)

	res = LocalOrder.GetTrackingSymbol(ctx)

	fmt.Println("Tracking Symbol:")

	for _, val := range res {
		fmt.Println(val)
	}

	Order3 := UnitLimitOrder{
		Symbol:     "AAL",
		OrderType:  10,
		OrderID:    10705862636,
		Quantity:   1,
		Price:      "10.23",
		Status:     "WORKING",
		CreateTime: time.Now(),
		Editable:   true,
		Cancelable: true,
	}

	Order4 := UnitLimitOrder{
		Symbol:     "AAL",
		OrderType:  -10,
		OrderID:    10705862637,
		Quantity:   1,
		Price:      "25.23",
		Status:     "WORKING",
		CreateTime: time.Now(),
		Editable:   true,
		Cancelable: true,
	}

	LocalOrder.CreateOTAOrderStatus(ctx, Order3, Order4)
	ota := LocalOrder.OrderIsOTA(ctx, "AAL")
	fmt.Println("OTA", ota)
	otasymboltracking := LocalOrder.GetTrackingSymbol(ctx)

	fmt.Println("Tracking Symbol: ")

	for _, val := range otasymboltracking {
		fmt.Println(val)
	}

	//
	LocalOrder.ClearSymbolAllOrderStatus(ctx, "AAL")

}

func TestOrderOperation_CreateOTA(t *testing.T) {
	LocalOrder := LocalOrderStatus{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	mainOrder := UnitLimitOrder{
		Symbol:     "U",
		OrderType:  10,
		OrderID:    10705862626,
		Quantity:   5,
		Price:      "25.23",
		Status:     "WORKING",
		CreateTime: time.Now(),
		Editable:   true,
		Cancelable: true,
	}
	triggerOrder := UnitLimitOrder{
		Symbol:     "U",
		OrderType:  -10,
		OrderID:    10705862627,
		Quantity:   5,
		Price:      "25.23",
		Status:     "WORKING",
		CreateTime: time.Now(),
		Editable:   true,
		Cancelable: true,
	}
	LocalOrder.CreateOTAOrderStatus(ctx, mainOrder, triggerOrder)
	res := LocalOrder.OrderIsOTA(ctx, mainOrder.Symbol)
	fmt.Println(res)
	LocalOrder.ClearSymbolAllOrderStatus(ctx, mainOrder.Symbol)
	LocalOrder.ClearSymbolAllOrderStatus(ctx, mainOrder.Symbol)
}

func TestLocalOrderStatus_TryReplaceOTATimeOutInOpenStage(t *testing.T) {
	orderSys := OrderOperation{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10000)
	defer cancel()
	orderAPI := TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	orderSys.OrderAPI = &orderAPI

	// OTA  order template
	LocalOrder := LocalOrderStatus{}
	mainOrder := UnitLimitOrder{
		Symbol:     "U",
		OrderType:  10,
		Quantity:   1,
		Price:      "25.23",
		CreateTime: time.Now(),
		Editable:   false,
		Cancelable: true,
	}
	triggerOrder := UnitLimitOrder{
		Symbol:     "U",
		OrderType:  -10,
		Quantity:   1,
		Price:      "40.23",
		CreateTime: time.Now(),
		Editable:   false,
		Cancelable: true,
	}

	LocalOrder.ClearSymbolAllOrderStatus(ctx, "U")

	// API 申請掛單
	tt := time.Now()
	code := orderSys.OrderAPI.CreateOTAOrder(ctx, &mainOrder, &triggerOrder)
	fmt.Println("申請掛單", time.Since(tt))
	if code >= 200 && code < 300 {
		t := time.Now()
		// 列出目前線上所有掛單
		mainOrderSource, triggerOrderSource, _ := orderSys.UpdateLimitOrder(ctx, 45)
		fmt.Println("取得更新", time.Since(t))
		// 從線上狀態批配剛剛申請的掛單
		MatchMainOrder, MatchTriggerOrder := orderSys.MatchFullOTAOrderID(&mainOrder, &triggerOrder, mainOrderSource, triggerOrderSource)

		// 將剛剛掛單更新至local
		LocalOrder.CreateOTAOrderStatus(ctx, *MatchMainOrder, *MatchTriggerOrder)

		// local 確認掛單狀態
		ota := LocalOrder.OrderIsOTA(ctx, MatchMainOrder.Symbol)
		fmt.Println("OTA: ", ota)
		openOrder, closeOrder := LocalOrder.GetOpenCloseOrder(ctx, "U")
		fmt.Println("currentOpen", openOrder.OrderID, openOrder.Price, openOrder.Status)
		fmt.Println("currentClose", closeOrder.OrderID, closeOrder.Price, closeOrder.Status)

		// 假設逾時 目前為open階段
		if openOrder.Status == "PENDING_ACTIVATION" { // 當時在假日測試 非開市時間
			t = time.Now()
			orderSys.OrderAPI.DeleteOrder(ctx, fmt.Sprintf("%v", openOrder.OrderID)) // 刪除整個ota
			fmt.Println("刪除order", time.Since(t))
			LocalOrder.ClearSymbolAllOrderStatus(ctx, "U")
			openOrder.Price = "26.23" // 更新價格
			t = time.Now()
			code = orderSys.OrderAPI.CreateOTAOrder(ctx, openOrder, closeOrder)
			fmt.Println("申請order", time.Since(t))
			if code >= 200 && code < 300 {
				// 查詢目前線上掛單狀態
				t = time.Now()
				mainOrderSource, triggerOrderSource, _ = orderSys.UpdateLimitOrder(ctx, 45)
				fmt.Println("取得更新", time.Since(t))
				// 從線上狀態批配剛剛申請的掛單
				MatchMainOrder, MatchTriggerOrder = orderSys.MatchFullOTAOrderID(openOrder, closeOrder, mainOrderSource, triggerOrderSource)
				fmt.Println(*MatchMainOrder)
				fmt.Println(*MatchTriggerOrder)
				// 將剛剛掛單更新至local
				LocalOrder.CreateOTAOrderStatus(ctx, *MatchMainOrder, *MatchTriggerOrder)

				// local 確認掛單狀態
				tracking := LocalOrder.GetTrackingSymbol(ctx)
				fmt.Println("tracking", tracking)
				ids := LocalOrder.GetAllOrderID(ctx, "U")
				fmt.Println(ids)
				ota = LocalOrder.OrderIsOTA(ctx, MatchMainOrder.Symbol)
				fmt.Println("OTA: ", ota)
				NewOpen := LocalOrder.GetOrderStatus(ctx, MatchMainOrder.OrderID)
				NewClose := LocalOrder.GetOrderStatus(ctx, MatchTriggerOrder.OrderID)
				fmt.Println("After Replacing Open", NewOpen.OrderID, NewOpen.Price, NewOpen.Status)
				fmt.Println("After Replacing Close", NewClose.OrderID, NewClose.Price, NewClose.Status)
				size := LocalOrder.GetOrderStatusSize(ctx, "U")
				fmt.Println("size=", size)
				orderSys.OrderAPI.DeleteOrder(ctx, fmt.Sprintf("%v", NewOpen.OrderID)) // 結束刪除刪除
				LocalOrder.ClearSymbolAllOrderStatus(ctx, "U")
				fmt.Println("刪除後")
				tracking = LocalOrder.GetTrackingSymbol(ctx)
				fmt.Println("tracking", tracking)
				ids = LocalOrder.GetAllOrderID(ctx, "U")
				fmt.Println(ids)
				size = LocalOrder.GetOrderStatusSize(ctx, "U")
				fmt.Println("size=", size)
				ota = LocalOrder.OrderIsOTA(ctx, MatchMainOrder.Symbol)
				fmt.Println("OTA: ", ota)
			}
		}

	}

}

func TestLocalOrderStatus_UpdateLocalOrderStatusFromAPIResponse(t *testing.T) {
	orderSys := OrderOperation{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10000)
	defer cancel()
	orderAPI := TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	orderSys.OrderAPI = &orderAPI

	// OTA  order template
	LocalOrder := LocalOrderStatus{}
	mainOrder := UnitLimitOrder{
		Symbol:     "U",
		OrderType:  10,
		Quantity:   1,
		Price:      "25.23",
		CreateTime: time.Now(),
		Editable:   false,
		Cancelable: true,
	}

	triggerOrder := UnitLimitOrder{
		Symbol:     "U",
		OrderType:  -10,
		Quantity:   1,
		Price:      "60.23",
		CreateTime: time.Now(),
		Editable:   false,
		Cancelable: true,
	}

	LocalOrder.ClearSymbolAllOrderStatus(ctx, "U")

	// 隨便創建一個ota
	code := orderSys.OrderAPI.CreateOTAOrder(ctx, &mainOrder, &triggerOrder)

	if code >= 200 && code < 300 {
		mainorderFromAPI, triggerOrderFromAPI, _ := orderSys.UpdateLimitOrder(ctx, 45)
		matchMainOder, matchTriggerOrder := orderSys.MatchFullOTAOrderID(&mainOrder, &triggerOrder, mainorderFromAPI, triggerOrderFromAPI)
		fmt.Printf("%+v\n", *matchMainOder)
		fmt.Printf("%+v\n", *matchTriggerOrder)
		LocalOrder.CreateOTAOrderStatus(ctx, *matchMainOder, *matchTriggerOrder)
		localmain := LocalOrder.GetOrderStatus(ctx, matchMainOder.OrderID)
		localtrigger := LocalOrder.GetOrderStatus(ctx, matchTriggerOrder.OrderID)
		fmt.Println("OTA=", LocalOrder.OrderIsOTA(ctx, "U"))
		fmt.Printf("main %+v\n", localmain)
		fmt.Printf("trigger %+v\n", localtrigger)
		fmt.Println("size=", LocalOrder.GetOrderStatusSize(ctx, "U"))

		// 將ota 刪掉 使狀態變canceled
		code = orderSys.OrderAPI.DeleteOrder(ctx, fmt.Sprintf("%v", localmain.OrderID))

		if code >= 200 && code < 300 {
			mainorderFromAPI, triggerOrderFromAPI, _ = orderSys.UpdateLimitOrder(ctx, 45)
			allSymbol := LocalOrder.GetTrackingSymbol(ctx)
			for _, s := range allSymbol {
				orderIDS := LocalOrder.GetAllOrderID(ctx, s)
				fmt.Println(orderIDS)
			}

			LocalOrder.UpdateLocalOrderStatusFromAPIResponse(ctx, mainorderFromAPI, triggerOrderFromAPI)
			localmain = LocalOrder.GetOrderStatus(ctx, matchMainOder.OrderID)
			localtrigger = LocalOrder.GetOrderStatus(ctx, matchTriggerOrder.OrderID)
			fmt.Println("OTA=", LocalOrder.OrderIsOTA(ctx, "U"))
			fmt.Printf("main %+v\n", localmain)
			fmt.Printf("trigger %+v\n", localtrigger)
			fmt.Println("size=", LocalOrder.GetOrderStatusSize(ctx, "U"))
			LocalOrder.ClearSymbolAllOrderStatus(ctx, "U")
		}

	}

}

func TestLocalOrderStatus_GetOrderStage(t *testing.T) {
	//logger := Logger{Stdout: true, LogPath: "./test123.log"}
	//
	//logger.Init()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10000)
	defer cancel()
	LocalOrder := LocalOrderStatus{}

	//// OTA  order template
	//OpenOrder := UnitLimitOrder{
	//	Symbol:     "U",
	//	OrderType:  10,
	//	Status:     "FILLED",
	//	OrderID:    10705812678,
	//	Quantity:   1,
	//	Price:      "25.23",
	//	CreateTime: time.Now(),
	//	Editable:   false,
	//	Cancelable: true,
	//}
	//
	//CloseOrder := UnitLimitOrder{
	//	Symbol:     "U",
	//	OrderType:  -10,
	//	OrderID:    107058126279,
	//	Status:     "WORKING",
	//	Quantity:   1,
	//	Price:      "60.23",
	//	CreateTime: time.Now(),
	//	Editable:   false,
	//	Cancelable: true,
	//}
	//
	//// 一般order test
	//LocalOrder.ClearSymbolAllOrderStatus(ctx, OpenOrder.Symbol)
	//
	//LocalOrder.CreateOrderStatus(ctx, OpenOrder)
	//LocalOrder.CreateOrderStatus(ctx, CloseOrder)
	//
	//LocalOrder.GetOrderStatus(ctx, OpenOrder.OrderID)
	//LocalOrder.GetOrderStatus(ctx, CloseOrder.OrderID)
	//
	//stage, workingOrder := LocalOrder.GetOrderStage(ctx, OpenOrder.Symbol)
	//
	//stagemap := map[int]string{1: "OPEN stage", 2: "CLOSE stage", 0: "ALL FILLED", -1: "UNKNOWN"}
	//
	//fmt.Println(LocalOrder.OrderIsOTA(ctx, OpenOrder.Symbol))
	//fmt.Println(stagemap[stage])
	//fmt.Printf("%+v\n", *workingOrder)
	//
	//LocalOrder.ClearSymbolAllOrderStatus(ctx, OpenOrder.Symbol)
	//
	//// ota order test
	//
	//LocalOrder.CreateOTAOrderStatus(ctx, OpenOrder, CloseOrder)
	//stage, workingOrder = LocalOrder.GetOrderStage(ctx, OpenOrder.Symbol)
	//
	//fmt.Println(LocalOrder.OrderIsOTA(ctx, OpenOrder.Symbol))
	//fmt.Println(stagemap[stage])
	//fmt.Printf("%+v\n", *workingOrder)
	//
	//LocalOrder.ClearSymbolAllOrderStatus(ctx, OpenOrder.Symbol)

	fmt.Println(LocalOrder.GetOrderStage(ctx, "TSLA"))
}

func TestLocalOrderStatus_ReplaceOrder(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10000)
	defer cancel()

	// OTA  order template
	LocalOrder := LocalOrderStatus{}
	OpenOrder := UnitLimitOrder{
		Symbol:     "U",
		OrderType:  10,
		Status:     "FILLED",
		OrderID:    10705812678,
		Quantity:   1,
		Price:      "25.23",
		CreateTime: time.Now(),
		Editable:   false,
		Cancelable: true,
	}

	CloseOrder := UnitLimitOrder{
		Symbol:     "U",
		OrderType:  -10,
		OrderID:    107058126279,
		Status:     "WORKING",
		Quantity:   1,
		Price:      "60.23",
		CreateTime: time.Now(),
		Editable:   false,
		Cancelable: true,
	}

	NewCloseOrder := UnitLimitOrder{
		Symbol:     "U",
		OrderType:  -10,
		OrderID:    107058126280,
		Status:     "WORKING",
		Quantity:   1,
		Price:      "70.23",
		CreateTime: time.Now(),
		Editable:   false,
		Cancelable: true,
	}

	// 一般order test
	LocalOrder.ClearSymbolAllOrderStatus(ctx, OpenOrder.Symbol)
	LocalOrder.CreateOrderStatus(ctx, OpenOrder)
	LocalOrder.CreateOrderStatus(ctx, CloseOrder)

	LocalOrder.ReplaceOrder(ctx, &NewCloseOrder) // 更新價格

	size := LocalOrder.GetOrderStatusSize(ctx, OpenOrder.Symbol)
	fmt.Println(size)
	allorder := LocalOrder.GetAllOrderID(ctx, OpenOrder.Symbol)
	fmt.Println(allorder)
	tracking := LocalOrder.GetTrackingSymbol(ctx)
	fmt.Println(tracking)

	for _, val := range allorder {
		intid, _ := strconv.Atoi(val)
		fmt.Printf("%+v", *LocalOrder.GetOrderStatus(ctx, int64(intid)))
	}

	LocalOrder.ClearSymbolAllOrderStatus(ctx, OpenOrder.Symbol)

	//// 嘗試更新ota 看看會不會error
	//
	//LocalOrder.CreateOTAOrderStatus(ctx, OpenOrder, CloseOrder)
	//LocalOrder.ReplaceOrder(ctx, &NewCloseOrder) // 更新價格
	//LocalOrder.ClearSymbolAllOrderStatus(ctx, OpenOrder.Symbol)

}

func TestLocalOrderStatus_InspectTimeOut(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10000)
	defer cancel()

	LocalOrder := LocalOrderStatus{}
	OpenOrder := UnitLimitOrder{
		Symbol:     "U",
		OrderType:  10,
		Status:     "WORKING",
		OrderID:    10705812678,
		Quantity:   1,
		Price:      "25.23",
		CreateTime: time.Now(),
		Editable:   false,
		Cancelable: true,
	}

	// 一般order test
	LocalOrder.ClearSymbolAllOrderStatus(ctx, OpenOrder.Symbol)
	LocalOrder.CreateOrderStatus(ctx, OpenOrder)

	LocalOrder.InspectTimeOut(ctx, "U", 5)

	time.Sleep(time.Second * 6)

	LocalOrder.InspectTimeOut(ctx, "U", 5)

	LocalOrder.ClearSymbolAllOrderStatus(ctx, OpenOrder.Symbol)

}

func TestLocalOrderStatus_ClearOrderStatus(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10000)
	defer cancel()

	LocalOrder := LocalOrderStatus{}

	LocalOrder.ClearOrderStatus(ctx, "10761675558")
}

func TestLocalOrderStatus_GetLockStatus(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10000)
	defer cancel()

	LocalOrder := LocalOrderStatus{}

	s1 := LocalOrder.GetOpenOrderLockStatus(ctx, "GDX")

	fmt.Println(s1)

	LocalOrder.SetOpenOrderLock(ctx, "GDX")

	s2 := LocalOrder.GetOpenOrderLockStatus(ctx, "GDX")

	fmt.Println(s2)

}

func TestLocalOrderStatus_GetOTAMainTriggerStatus(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10000)
	defer cancel()

	LocalOrder := LocalOrderStatus{}
	OrderMain := &UnitLimitOrder{Symbol: "AAL", OrderID: 1234456, OrderType: 10, Quantity: 1, Price: "12.02", CreateTime: time.Now().In(Loc)}

	OrderTrigger := &UnitLimitOrder{Symbol: "AAL", OrderID: 2345561, OrderType: -10, Quantity: 1, Price: "18.99", CreateTime: time.Now().In(Loc)}

	LocalOrder.CreateOTAOrderStatus(ctx, *OrderMain, *OrderTrigger)

	mainOrder, triggerOrder := LocalOrder.GetOTAMainTriggerStatus(ctx, "AAL")

	fmt.Printf("%+v\n", *mainOrder)
	fmt.Printf("%+v\n", *triggerOrder)

}

func TestLocalOrderStatus_AddOrderOperationNum(t *testing.T) {

	LocalOrder := LocalOrderStatus{}
	res2 := LocalOrder.GetOrderOperationNum()
	fmt.Println("today order op Num: ", res2)
	LocalOrder.AddOrderOperationNum(3)
	LocalOrder.AddOrderOperationNum(2)
	LocalOrder.AddOrderOperationNum(3)
	LocalOrder.AddOrderOperationNum(4)
	LocalOrder.AddOrderOperationNum(5)
	time.Sleep(time.Second * 2)

	res2 = LocalOrder.GetOrderOperationNum()

	fmt.Println("today order op Num: ", res2)

	res := LocalOrder.redisCli.TTL(context.Background(), "OrderOperationNum")
	fmt.Println(res)

	LocalOrder.redisCli.Del(context.Background(), "OrderOperationNum")

}
