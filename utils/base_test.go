package utils

import (
	"fmt"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestStrategy(t *testing.T) {
	Data := make([]*TradingData, 0)
	Data = append(Data, &TradingData{Price: decimal.NewFromInt(150), Symbol: "GOOG", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true})
	Data = append(Data, &TradingData{Price: decimal.NewFromInt(100), Symbol: "GOOG", Volume: decimal.NewFromInt(1), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true})
	Data = append(Data, &TradingData{Price: decimal.NewFromInt(100), Symbol: "GOOG", Volume: decimal.NewFromInt(1), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true})
	Data = append(Data, &TradingData{Price: decimal.NewFromInt(100), Symbol: "GOOG", Volume: decimal.NewFromInt(2), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true})
	Data = append(Data, &TradingData{Price: decimal.NewFromInt(50), Symbol: "GOOG", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true})

	idt := &Indicator{}
	idt.Data = Data

	avg := idt.GetAvgPrice()
	fmt.Println(avg)

	wtavg := idt.GetWtAvgPrice()
	fmt.Println(wtavg)

}

func TestRSI(t *testing.T) {
	Data := make([]*TradingData, 0)
	idt := &Indicator{}
	Data = append(Data, &TradingData{Price: decimal.NewFromInt(150), Symbol: "GOOG", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true})
	Data = append(Data, &TradingData{Price: decimal.NewFromInt(80), Symbol: "GOOG", Volume: decimal.NewFromInt(1), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true})
	Data = append(Data, &TradingData{Price: decimal.NewFromInt(400), Symbol: "GOOG", Volume: decimal.NewFromInt(1), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true})
	Data = append(Data, &TradingData{Price: decimal.NewFromInt(140), Symbol: "GOOG", Volume: decimal.NewFromInt(2), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true})
	Data = append(Data, &TradingData{Price: decimal.NewFromFloat32(250), Symbol: "GOOG", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true})
	idt.Data = Data
	rsi := idt.GetRSI()
	fmt.Println(rsi)
}

func TestSQ(t *testing.T) {
	t1 := time.Now()
	sq := SQ{}
	sq.RedisCli = GetRedis("1")
	data1 := &TradingData{Price: decimal.NewFromInt(100), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true}
	data2 := &TradingData{Price: decimal.NewFromInt(101), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 1), FormT: true}
	data3 := &TradingData{Price: decimal.NewFromInt(102), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 5), FormT: true}
	data4 := &TradingData{Price: decimal.NewFromInt(101), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 7), FormT: true}
	data5 := &TradingData{Price: decimal.NewFromInt(102), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 20), FormT: true}
	data6 := &TradingData{Price: decimal.NewFromInt(102), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 21), FormT: true}
	data7 := &TradingData{Price: decimal.NewFromInt(102), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 22), FormT: true}
	data8 := &TradingData{Price: decimal.NewFromInt(102), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(-time.Second * 86405), FormT: true}
	data9 := &TradingData{Price: decimal.NewFromInt(102), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(-time.Second * 86401), FormT: true}
	data10 := &TradingData{Price: decimal.NewFromInt(102), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(-time.Second * 86400), FormT: true}
	sq.Append(data1)
	sq.Append(data2)
	sq.Append(data3)
	sq.Append(data4)
	sq.Append(data5)
	sq.Append(data6)
	sq.Append(data7)
	sq.Append(data8)
	sq.Append(data9)
	sq.Append(data10)
	for _, val := range sq.Data {
		fmt.Println(val.TradingTime)
	}
	fmt.Println(time.Since(t1))
}

func TestCMP_GetGradient(t *testing.T) {
	//rc := GetRedis("0")
	sq := SQ{}
	sq.TimeAllowSec = time.Second * 10
	data1 := &TradingData{Price: decimal.NewFromInt(100), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true}
	data2 := &TradingData{Price: decimal.NewFromInt(101), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 1), FormT: true}
	data3 := &TradingData{Price: decimal.NewFromInt(102), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 2), FormT: true}
	data4 := &TradingData{Price: decimal.NewFromInt(101), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 3), FormT: true}
	data5 := &TradingData{Price: decimal.NewFromInt(102), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 4), FormT: true}
	data6 := &TradingData{Price: decimal.NewFromInt(102), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 5), FormT: true}
	data7 := &TradingData{Price: decimal.NewFromInt(102), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 6), FormT: true}
	data8 := &TradingData{Price: decimal.NewFromInt(102), Symbol: "U", Volume: decimal.NewFromInt(3), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 9), FormT: true}
	sq.Append(data1)
	sq.Append(data2)
	sq.Append(data3)
	sq.Append(data4)
	sq.Append(data5)
	sq.Append(data6)
	sq.Append(data7)
	sq.Append(data8)

	sq2 := SQ{}
	sq2.TimeAllowSec = time.Second * 10
	data11 := &TradingData{Price: decimal.NewFromInt(110), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(5), TradingTime: time.Now(), FormT: true}
	data22 := &TradingData{Price: decimal.NewFromInt(111), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(4), TradingTime: time.Now().Add(time.Second * 1), FormT: true}
	data33 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(3), TradingTime: time.Now().Add(time.Second * 2), FormT: true}
	data44 := &TradingData{Price: decimal.NewFromInt(111), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 3), FormT: true}
	data55 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 4), FormT: true}
	data66 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(8), TradingTime: time.Now().Add(time.Second * 5), FormT: true}
	data77 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(1), TradingTime: time.Now().Add(time.Second * 6), FormT: true}
	data88 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(7), TradingTime: time.Now().Add(time.Second * 9), FormT: true}
	sq2.Append(data11)
	sq2.Append(data22)
	sq2.Append(data33)
	sq2.Append(data44)
	sq2.Append(data55)
	sq2.Append(data66)
	sq2.Append(data77)
	sq2.Append(data88)
	cmp := CMP{}

	res := cmp.GetGradient(sq2.Data, sq.Data)
	fmt.Println("兩組資料 單位時間內差距")
	fmt.Println(res.Price, res.Volume, res.EventNum)

}

func TestReferWave(t *testing.T) {
	sq2 := SQ{}
	sq2.TimeAllowSec = time.Second * 10
	data11 := &TradingData{Price: decimal.NewFromInt(110), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(5), TradingTime: time.Now(), FormT: true}
	data22 := &TradingData{Price: decimal.NewFromInt(111), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(4), TradingTime: time.Now().Add(time.Second * 1), FormT: true}
	data33 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(3), TradingTime: time.Now().Add(time.Second * 2), FormT: true}
	data44 := &TradingData{Price: decimal.NewFromInt(111), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 3), FormT: true}
	data55 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 4), FormT: true}
	data66 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(8), TradingTime: time.Now().Add(time.Second * 5), FormT: true}
	data77 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(1), TradingTime: time.Now().Add(time.Second * 6), FormT: true}
	data88 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "U", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(7), TradingTime: time.Now().Add(time.Second * 9), FormT: true}
	sq2.Append(data11)
	sq2.Append(data22)
	sq2.Append(data33)
	sq2.Append(data44)
	sq2.Append(data55)
	sq2.Append(data66)
	sq2.Append(data77)
	sq2.Append(data88)
	sq2.RedisCli = GetRedis("0")
}

func TestReferNMinWave(t *testing.T) {
	sq2 := SQ{}
	sq2.RedisCli = GetRedis("0")
	sq2.TimeAllowSec = time.Second * 10
	t0 := time.Now()
	data11 := &TradingData{Price: decimal.NewFromInt(110), Symbol: "AAPL", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(5), TradingTime: time.Now(), FormT: true}
	data22 := &TradingData{Price: decimal.NewFromInt(111), Symbol: "AAPL", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(4), TradingTime: time.Now().Add(time.Second * 1), FormT: true}
	data33 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "AAPL", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(3), TradingTime: time.Now().Add(time.Second * 2), FormT: true}
	data44 := &TradingData{Price: decimal.NewFromInt(111), Symbol: "AAPL", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 3), FormT: true}
	data55 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "AAPL", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now().Add(time.Second * 4), FormT: true}
	data66 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "AAPL", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(8), TradingTime: time.Now().Add(time.Second * 5), FormT: true}
	data77 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "AAPL", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(1), TradingTime: time.Now().Add(time.Second * 6), FormT: true}
	data88 := &TradingData{Price: decimal.NewFromInt(112), Symbol: "AAPL", Volume: decimal.NewFromInt(6), InsRate: int32(30), NextRate: int32(30), EventNum: int32(7), TradingTime: time.Now().Add(time.Second * 9), FormT: true}
	sq2.Append(data11)
	sq2.Append(data22)
	sq2.Append(data33)
	sq2.Append(data44)
	sq2.Append(data55)
	sq2.Append(data66)
	sq2.Append(data77)
	sq2.Append(data88)
	t1 := time.Now()
	Pwave, Nwave := sq2.GetUnitReferNMinWave(5)
	fmt.Println(time.Since(t0))
	fmt.Println(time.Since(t1))

	fmt.Println(Pwave, Nwave)
}
