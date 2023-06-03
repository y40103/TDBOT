package utils

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGetAPIAccessToken(t *testing.T) {
	orderAPI := TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*2000)
	defer cancel()
	httpCode, accesstoken := orderAPI.GetAccessToken(ctx)
	fmt.Println(httpCode, accesstoken)
	httpCode, accesstoken = orderAPI.GetAccessToken(ctx)
	fmt.Println(httpCode, accesstoken)
}

func TestLimitOrder(t *testing.T) {
	orderAPI := TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	mainOrder := UnitLimitOrder{Symbol: "TSLA", OrderType: 10, Quantity: 1, Price: "170.01"}
	// 10 buy 20 sell short -10 sell -20 buy to cover
	code := orderAPI.CreateLimitOrder(ctx, &mainOrder)
	fmt.Println(code)
}

// OTA狀態機制  掛單出去 為一個複合的 若main被filled 會掛出第二個 此時若將返回結果限制1 就會已一般sell的order顯示 但若 返回結果顯示2 則會以ota狀態返回
func TestGetOrderStatus(t *testing.T) {
	orderAPI := TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	code, order := orderAPI.GetCurrentOrderStatus(ctx, "2023-04-12", "2023-04-12", "", "1")

	fmt.Println(code)
	fmt.Println(*order)

}

// 似乎只能同標的
func TestOTA(t *testing.T) {
	orderAPI := TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	mainOrder := UnitLimitOrder{Symbol: "AAL", OrderType: 10, Quantity: 1, Price: "12.16"}
	triggerOrder := UnitLimitOrder{Symbol: "AAL", OrderType: -10, Quantity: 1, Price: "18.99"}
	code := orderAPI.CreateOTAOrder(ctx, &mainOrder, &triggerOrder)

	fmt.Println(code)
	fmt.Println(time.Now().In(Loc))
}

func TestDeleteOrder(t *testing.T) {
	orderAPI := TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	code := orderAPI.DeleteOrder(ctx, "10714858319")
	fmt.Println(code)
}

func TestTDOrder_ReplaceOrder(t *testing.T) {
	orderAPI := TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	NewOrder := UnitLimitOrder{Symbol: "AAL", OrderType: -10, Quantity: 1, Price: "13.16"}
	code := orderAPI.ReplaceOrder(ctx, "10760942643", &NewOrder)
	fmt.Println(code)
}

func TestTDOrder_ReplaceOTAOpenOrder(t *testing.T) {
	orderAPI := TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()

	newMain := ApplyOrder{OrderKind: "OTA"}
	newMain.OrderContent = make(map[string]*UnitLimitOrder)
	//newMain.OrderContent["OPEN"] = new(UnitLimitOrder)
	//newMain.OrderContent["CLOSE"] = new(UnitLimitOrder)
	newMain.OrderContent["OPEN"] = &UnitLimitOrder{Symbol: "AAL", Price: "12.12", OrderType: 10, CreateTime: time.Now().In(Loc)}
	code := orderAPI.ReplaceOrder(ctx, "10788948572", newMain.OrderContent["OPEN"])
	fmt.Println(code)
}

func TestTDOrder_GetTransactionHistory(t *testing.T) {
	orderAPI := TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*4000)
	defer cancel()
	Param := UnitHisTransaction{TransactionType: "ALL", Symbol: "", StartDate: "2023-05-19", EndDate: "2023-05-21"}
	code, res := orderAPI.GetTransactionHistory(ctx, Param)

	fmt.Println(code)
	fmt.Printf("%+v\n", *res)
	fmt.Println(len(*res))
	for _, val := range *res {
		fmt.Println(val)
	}

}
