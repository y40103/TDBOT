package utils

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestSymbolOrderStatus_UpdateOrder(t *testing.T) {
	orderSys := OrderOperation{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10000)
	defer cancel()
	orderAPI := TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	orderSys.OrderAPI = &orderAPI
	mainOrder, triggerOrder, _ := orderSys.UpdateLimitOrder(ctx, 200)
	fmt.Println("main: #########")
	for k, v := range mainOrder {
		fmt.Printf("key: %v, val: %+v\n", k, *v)
	}
	fmt.Println("trigger: #########")
	for k, v := range triggerOrder {
		fmt.Printf("key: %v, val: %+v\n", k, *v)
	}

}

func TestOrderOperation_MatchFullOTAOrderID(t *testing.T) {
	orderSys := OrderOperation{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10000)
	defer cancel()
	orderAPI := &TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	orderSys.OrderAPI = orderAPI
	mainOrder := UnitLimitOrder{Symbol: "AAL", OrderType: 10, Quantity: 1, Price: "10.91"}
	triggerOrder := UnitLimitOrder{Symbol: "AAL", OrderType: -10, Quantity: 1, Price: "11.99"}
	code := orderSys.OrderAPI.CreateOTAOrder(ctx, &mainOrder, &triggerOrder)
	SourceMainOrder, SourceTriggerOrder, _ := orderSys.UpdateLimitOrder(ctx, 45)
	if code >= 200 && code < 300 {
		onlineMainOrder, onlineTriggerOrder := orderSys.MatchFullOTAOrderID(&mainOrder, &triggerOrder, SourceMainOrder, SourceTriggerOrder)
		fmt.Printf("MatchMain: %+v ,time: %v\n", *onlineMainOrder, onlineMainOrder.CreateTime)
		fmt.Printf("MatchTrigger: %+v, time: %v\n", *onlineTriggerOrder, onlineTriggerOrder.CreateTime)
		delcode := orderSys.OrderAPI.DeleteOrder(ctx, fmt.Sprintf("%v", onlineMainOrder.OrderID))
		if delcode >= 200 && code < 300 {
			fmt.Println("delete order: orderid=", onlineMainOrder.OrderID)
		}
	}

}

func TestOrderOperation_MatchOrderID(t *testing.T) {
	orderSys := OrderOperation{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10000)
	defer cancel()
	orderAPI := &TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	orderSys.OrderAPI = orderAPI
	mainOrder := UnitLimitOrder{Symbol: "TSLA", OrderType: 10, Quantity: 1, Price: "170.01"}
	code := orderAPI.CreateLimitOrder(ctx, &mainOrder)
	fmt.Println(time.Now().In(Loc))
	SourceMainOrder, SourceTriggerOrder, _ := orderSys.UpdateLimitOrder(ctx, 45)

	// 測試流程 創建limit match , replace new limit, match
	if code >= 200 && code < 300 {
		onlineMainOrder := orderSys.MatchOrderID(&mainOrder, SourceMainOrder, SourceTriggerOrder)
		fmt.Printf("MatchMain: %+v ,time: %v\n", *onlineMainOrder, onlineMainOrder.CreateTime)
		mainOrderRE := UnitLimitOrder{Symbol: "TSLA", OrderType: 10, Quantity: 1, Price: "150.01"}
		fmt.Println(time.Now().In(Loc))
		reCode := orderSys.OrderAPI.ReplaceOrder(ctx, fmt.Sprintf("%v", onlineMainOrder.OrderID), &mainOrderRE)
		if reCode >= 200 && code < 300 {
			SourceMainOrder, SourceTriggerOrder, _ = orderSys.UpdateLimitOrder(ctx, 45)
			onlineMainOrder = orderSys.MatchOrderID(&mainOrderRE, SourceMainOrder, SourceTriggerOrder)
			fmt.Printf("After replacing MatchMain: %+v ,time: %v\n", *onlineMainOrder, onlineMainOrder.CreateTime)
		}
		delcode := orderSys.OrderAPI.DeleteOrder(ctx, fmt.Sprintf("%v", onlineMainOrder.OrderID))
		if delcode >= 200 && code < 300 {
			fmt.Println("delete order: orderid=", onlineMainOrder.OrderID)
		}
	}

}

func TestOrderOperation_ReplaceCloseOTAOrder(t *testing.T) {

	orderSys := OrderOperation{}
	orderAPI := &TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	orderSys.OrderAPI = orderAPI
	newOrder := &UnitLimitOrder{Symbol: "AAL", OrderType: -10, Quantity: 1, Price: "16.08"}
	oldOrder := &UnitLimitOrder{Symbol: "AAL", OrderType: -10, Quantity: 1, Price: "16.02", OrderID: 10789523152}
	matchOrder := orderSys.ReplaceCloseOTAOrder(oldOrder, newOrder)
	fmt.Printf("New: %+v", matchOrder)
}

func TestOrderOperation_ReplaceOpenOTAOrder(t *testing.T) {

	orderSys := OrderOperation{}
	orderAPI := &TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	orderSys.OrderAPI = orderAPI
	LocalOrder := LocalOrderStatus{}
	oldOrderMain := &UnitLimitOrder{Symbol: "AAL", OrderType: 10, Quantity: 1, Price: "12.10", OrderID: 10789523445, CreateTime: time.Now().In(Loc)}
	newOrderMain := &UnitLimitOrder{Symbol: "AAL", OrderType: 10, Quantity: 1, Price: "12.01", CreateTime: time.Now().In(Loc)}

	OldOrderTrigger := &UnitLimitOrder{Symbol: "AAL", OrderType: -10, Quantity: 1, Price: "18.99", CreateTime: time.Now().In(Loc)}

	matchOrder, triggerOder := orderSys.ReplaceOpenOTAOrder(oldOrderMain, OldOrderTrigger, newOrderMain)

	fmt.Printf("New main: %+v\n", matchOrder)
	fmt.Printf("New trigger : %+v\n", triggerOder)

	LocalOrder.CreateOTAOrderStatus(context.Background(), *matchOrder, *triggerOder)

	mainFromLocal, triggerFromLocal := LocalOrder.GetOTAMainTriggerStatus(context.Background(), "AAL")

	newOrder := *mainFromLocal

	newOrder.Price = "12.02"

	matchOrder2, triggerOder2 := orderSys.ReplaceOpenOTAOrder(mainFromLocal, triggerFromLocal, &newOrder)

	LocalOrder.ClearSymbolAllOrderStatus(context.Background(), "AAL")

	LocalOrder.CreateOTAOrderStatus(context.Background(), *matchOrder2, *triggerOder2)

	mainFromLocal2, triggerFromLocal2 := LocalOrder.GetOTAMainTriggerStatus(context.Background(), "AAL")

	fmt.Printf("New main: %+v\n", mainFromLocal2)
	fmt.Printf("New trigger : %+v\n", triggerFromLocal2)

	LocalOrder.ClearSymbolAllOrderStatus(context.Background(), "AAL")

}

func TestOrderOperation_ReplaceNormalOrder(t *testing.T) {

	orderSys := OrderOperation{}
	orderAPI := &TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	orderSys.OrderAPI = orderAPI
	oldOrder := &UnitLimitOrder{Symbol: "AAL", OrderType: 10, Quantity: 1, Price: "10.5", OrderID: 10740097391}
	newOrder := &UnitLimitOrder{Symbol: "AAL", OrderType: 10, Quantity: 1, Price: "10.7"}
	matchOrder := orderSys.ReplaceNormalOrder(oldOrder, newOrder)
	fmt.Printf("New: %+v", matchOrder)
}

func TestOrderOperation_GetDayPerformance(t *testing.T) {

	orderSys := OrderOperation{}
	orderAPI := &TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	orderSys.OrderAPI = orderAPI

	startDate := "2023-05-15"
	endDate := "2023-05-20"

	res := orderSys.GetPeriodTransactionPerformance("2023-05-15", "2023-05-20")

	SumReferPrice := decimal.Zero
	SumEarnings := decimal.Zero
	for key, val := range res {

		fmt.Printf("%+v %+v\n", key, *val)
		SumEarnings = SumEarnings.Add(val.SumEarns)
		SumReferPrice = SumReferPrice.Add(val.referPrice)
	}
	fmt.Printf("%v - %v\n", startDate, endDate)
	fmt.Printf("Period Performance = %v / %v = %v %%\n", SumEarnings, SumReferPrice, SumEarnings.Div(SumReferPrice).Mul(decimal.NewFromInt(100)))

}
