package model

import (
	"GoBot/utils"
	"context"
	"github.com/sirupsen/logrus"
	"time"
)

var AccessToken string

// GET ACCESS TOKEN  function 已設定下次更新時間 若時間尚未到達 會直接返回之前的token
// 這邊邏輯為 每60秒讓它去 call function 時間到 則會自動找api 否則就是返回最近一次的token
func UpdateAccessToken(OrderAPI utils.TDOrder) {
	orderAPI := &OrderAPI
	for {

		attempt := 0
		for { // 防止失敗 最多重試3次

			logrus.Infoln("try to update Accesstoken")
			ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*5000)
			httpCode, accesstoken := orderAPI.GetAccessToken(ctx)
			if utils.HttpSuccess(httpCode) {
				AccessToken = accesstoken
				logrus.Infoln("update AccessToken success: ", AccessToken)
				break
			} else if attempt == 3 {
				logrus.Warnf("fail to get accesstoken, even attempt %v times", attempt)
				break
			}
			attempt += 1
			time.Sleep(time.Millisecond * time.Duration(attempt*1000))
			logrus.Warnf("accesstoken attempt .... %v", attempt)
		}

		time.Sleep(time.Second * 60) // 60秒 確認 token 是否到了更新時間

	}

}
