package dbutils

import (
	"GoBot/utils"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestDb(t *testing.T) {
	pg := PsqlPool{SymbolList: []string{"GOOG", "AAPL"}, HOST: "localhost"}
	pg.Init()
	price := decimal.NewFromInt(100)
	GOOG_data := utils.TradingData{Price: price, Symbol: "GOOG", Volume: decimal.NewFromInt(100), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: true}
	APPL_data := utils.TradingData{Price: price, Symbol: "AAPL", Volume: decimal.NewFromInt(100), InsRate: int32(30), NextRate: int32(30), EventNum: int32(10), TradingTime: time.Now(), FormT: false}
	pg.InsertData(GOOG_data)
	pg.InsertData(APPL_data)
	pg.ClosePool()
}
