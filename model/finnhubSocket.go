package model

import (
	"GoBot/lib/dbutils"
	"GoBot/utils"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

type SymbolBase struct {
	Symbol   string
	Strategy StrategyInterface
	Budget   decimal.Decimal
}

type TradingDataSocket struct {
	Symbols map[string]*SymbolBase
	Token   string
}

func (self TradingDataSocket) Streaming(streamingChan map[string]chan *utils.TradingData) {

	MySymbol := make([]string, 0)
	for symbol, _ := range self.Symbols {
		MySymbol = append(MySymbol, symbol)
	}

	for {
		//startTIme := time.Now()

		w, _, err := websocket.DefaultDialer.Dial(self.Token, nil)
		if err != nil {
			panic(err)
		}
		defer w.Close()
		for _, s := range MySymbol {
			msg, _ := json.Marshal(map[string]interface{}{"type": "subscribe", "symbol": s})
			w.WriteMessage(websocket.TextMessage, msg) // socket 取得的資料 綁至msg
		}

		pg := dbutils.PsqlPool{SymbolList: MySymbol, HOST: "localhost"}
		pg.Init()

		var msg interface{}
		for {
			Data := new(utils.TradingDataSet)
			Data.Symbols = MySymbol
			Data.Init()
			//fmt.Println("##")
			err := w.ReadJSON(&msg) // parse msg , 可以理解\不停從socket種 讀取一組訊息
			if err != nil {
				panic(err)
			}
			Data.Parser(msg)
			for _, symbol := range Data.Symbols {
				if Data.Data[symbol].EventNum > 0 {
					//fmt.Println("-", *Data.Data[symbol])
					streamingChan[symbol] <- Data.Data[symbol]
					//pg.InsertData(*Data.Data[symbol])

				}

			}

			// 每25分鐘嘗試使他自己自動重連
			//if time.Now().After(startTIme.Add(time.Minute * time.Duration(15))) {
			//	logrus.Warn("reset webSocket.......")
			//	break
			//}

		}

	}

}
