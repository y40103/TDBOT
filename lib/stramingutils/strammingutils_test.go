package dbutils

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"testing"
)

func TestStreamingParser(t *testing.T) {
	w, _, err := websocket.DefaultDialer.Dial("wss://ws.finnhub.io?token=ceupcdaad3ibo9vbgra0ceupcdaad3ibo9vbgrag", nil)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	symbols := []string{"GOOG", "AAPL"} // "BINANCE:BTCUSDT"
	for _, s := range symbols {
		msg, _ := json.Marshal(map[string]interface{}{"type": "subscribe", "symbol": s})
		w.WriteMessage(websocket.TextMessage, msg) // socket 取得的資料 綁至msg
	}

	var msg interface{}
	for {
		Data := new(TradingDataSet)
		Data.Symbols = symbols
		Data.Init()
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@")
		err := w.ReadJSON(&msg) // parse msg , 可以理解\不停從socket種 讀取一組訊息
		if err != nil {
			panic(err)
		}
		Data.Parser(msg)
	}
}
