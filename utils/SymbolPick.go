package utils

import (
	"context"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
	"time"
)

// random
func RandSymbolIndex(GetIndexNum int, AllNum int) []int {

	res := make([]int, 0)

	// 查找今日是否已有紀錄
	redis := GetRedis("0")
	key := "TodayIndex"
	Allindex := redis.SMembers(context.Background(), key)
	allindexString, err := Allindex.Result()
	life := redis.TTL(context.Background(), key)

	if err == nil && len(allindexString) > 0 {

		for _, val := range allindexString {

			todayIndex, err := strconv.Atoi(val)

			if err != nil {
				logrus.Warnln(err)
			}

			res = append(res, todayIndex)

		}
		logrus.Infof("index: %v life: %v", res, life)
		return res
	}

	for i := 0; i < 10000; i++ {

		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(AllNum)

		// 初始項 直接加入
		if len(res) == 0 {
			res = append(res, n)
		}

		// 確認是否有重複項
		for index, val := range res {

			// 重複則放棄
			if n == val {
				break
			}

			// 全部結束表示無重複 則加入
			if index == len(res)-1 {
				res = append(res, n)
				break
			}
		}

		// 滿足size 輸出
		if len(res) == GetIndexNum {
			break
		}

	}

	// 寫入cache
	for _, index := range res {

		redis.SAdd(context.Background(), key, index)
	}

	setexp := redis.Expire(context.Background(), key, time.Second*3600*12)
	logrus.Infoln("set expired time ", setexp)

	return res

}
