package model

import (
	"GoBot/utils"
	"context"
	"fmt"
	"testing"
	"time"
)

func TestUpdateStatus(t *testing.T) {
	time.Sleep(time.Second * 2)

	orderSys := utils.OrderOperation{}
	LocalOrder := utils.LocalOrderStatus{}
	orderAPI := utils.TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "XXXXXXXX",
		ConsumerKey:  "OOOOOOOOOOOOOOOOOOOOOOOOOO",
		RefreshToken: "#####################################"}
	orderSys.OrderAPI = &orderAPI
	OpenOrder1 := utils.UnitLimitOrder{
		Symbol:    "GOOG",
		OrderType: 10,
		Quantity:  1,
		Price:     "82.0",
	}
	go UpdateAccessToken(orderAPI)
	OpenOrder2 := utils.UnitLimitOrder{
		Symbol:    "AAPL",
		OrderType: 10,
		Quantity:  1,
		Price:     "140",
	}

	orderSys.OrderAPI.AccessToken = AccessToken

	// 測試用觀察單
	ctx, _ := context.WithTimeout(context.Background(), time.Second*100)
	LocalOrder.ClearSymbolAllOrderStatus(ctx, OpenOrder1.Symbol)
	LocalOrder.ClearSymbolAllOrderStatus(ctx, OpenOrder2.Symbol)
	orderSys.OrderAPI.CreateLimitOrder(ctx, &OpenOrder1)
	SourceMainOrder, SourceTriggerOrder, _ := orderSys.UpdateLimitOrder(ctx, 30)
	matchOrder1 := orderSys.MatchOrderID(&OpenOrder1, SourceMainOrder, SourceTriggerOrder)
	fmt.Println(*matchOrder1, "######################")

	LocalOrder.CreateOrderStatus(ctx, *matchOrder1)
	orderSys.OrderAPI.CreateLimitOrder(ctx, &OpenOrder2)
	SourceMainOrder, SourceTriggerOrder, _ = orderSys.UpdateLimitOrder(ctx, 30)

	matchOrder2 := orderSys.MatchOrderID(&OpenOrder2, SourceMainOrder, SourceTriggerOrder)

	LocalOrder.CreateOrderStatus(ctx, *matchOrder2)
	defer orderSys.OrderAPI.DeleteOrder(context.Background(), fmt.Sprintf("%v", matchOrder1.OrderID))
	defer orderSys.OrderAPI.DeleteOrder(context.Background(), fmt.Sprintf("%v", matchOrder2.OrderID))

	go UpdateLocalOrderStatus(1000, orderAPI) // 一秒更新一次
	time.Sleep(time.Second * 60)
}

func TestUpdateStatus2(t *testing.T) {

	time.Sleep(time.Second * 2)

	orderSys := utils.OrderOperation{}
	orderAPI := utils.TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "XXXXXXXX",
		ConsumerKey:  "OOOOOOOOOOOOOOOOOOOOOOOOOO",
		RefreshToken: "#####################################"}
	orderSys.OrderAPI = &orderAPI
	go UpdateAccessToken(orderAPI)
	UpdateLocalOrderStatus(1000, orderAPI) // 一秒更新一次
	time.Sleep(time.Second * 60)
}
