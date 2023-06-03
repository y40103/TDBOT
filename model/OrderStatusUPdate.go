package model

import (
	"GoBot/utils"
	"context"
	"github.com/sirupsen/logrus"
	"time"
)

// unit is milliSecond, min = 1000
func UpdateLocalOrderStatus(interval int) {

	if interval < 1000 { //防止過度頻繁訪問API
		interval = 1000
	}

	orderSys := utils.OrderOperation{}
	LocalOrder := utils.LocalOrderStatus{}
	orderAPI := utils.TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		ConsumerKey:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx",
		RefreshToken: "OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO"}
	orderSys.OrderAPI = &orderAPI

	for {

		var mainorderAPISource, triggerOrderAPISource map[int64]*utils.UnitLimitOrder
		var httpCode int
		var ctx context.Context

		orderSys.OrderAPI.AccessToken = AccessToken
		symbols := LocalOrder.GetTrackingSymbol(context.Background())

		// 若沒有正在追蹤的標的 直接結束該次更新
		if len(symbols) == 0 {
			logrus.Infof("NO ANY ORDER, Symbols: %v", symbols)
			time.Sleep(time.Millisecond * 1000)
			continue
		}

		// 失敗重試最多三次
		count := 0
		resultNum := 100
		for {
			ctx, _ = context.WithTimeout(context.Background(), time.Millisecond*5000)
			mainorderAPISource, triggerOrderAPISource, httpCode = orderSys.UpdateLimitOrder(ctx, resultNum)
			if utils.HttpSuccess(httpCode) {
				logrus.Infoln(httpCode)

				// 更新至local
				ctx, _ = context.WithTimeout(context.Background(), time.Millisecond*5000)
				success := LocalOrder.UpdateLocalOrderStatusFromAPIResponse(ctx, mainorderAPISource, triggerOrderAPISource)

				if success == false { // 抓取數量不含內容 重試
					logrus.Warnln("UpdateLimitOrder not contain Order")
					count += 1
					resultNum += 25
					continue
				} else if count == 3 {
					break
				}

				break
			} else if count == 3 {
				logrus.Warnln("UpdateLocalOrderStatus: UpdateLimitOrder FAIL", httpCode)
				break
			}
			count += 1
			time.Sleep(time.Millisecond * time.Duration(count*1000))
			logrus.Warnf("UpdateLocalOrderStatus attempt... %v", count)
		}

		time.Sleep(time.Millisecond * time.Duration(interval))
	}

}
