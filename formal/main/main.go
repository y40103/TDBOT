package main

import (
	"GoBot/model"
	"GoBot/utils"
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

var TradingDataChannels = make(map[string]chan *utils.TradingData)

var MySymbol = map[string]*model.SymbolBase{
	"AMD":  {Strategy: &model.MyDemoStrategy{}, Budget: decimal.NewFromInt(1600)},
	"AAPL": {Strategy: &model.MyDemoStrategy{}, Budget: decimal.NewFromInt(1600)},
	"NVDA": {Strategy: &model.MyDemoStrategy{}, Budget: decimal.NewFromInt(1600)},
}

func init() {

	logger := utils.Logger{Stdout: true, LogPath: "/pj/GoBot/log/Bot.log"}
	logger.Init()

	for symbol, _ := range MySymbol {
		TradingDataChannels[symbol] = make(chan *utils.TradingData, 1000)
	}
}

func main() {

	finnhub := utils.FinnToken{}
	token := finnhub.GetToken("/pj/finnhubToken/finn_token/finn_token.yaml")
	url := fmt.Sprintf("wss://ws.finnhub.io?token=%v", token)

	tradingSocket := model.TradingDataSocket{Symbols: MySymbol, Token: url}

	orderAPI := utils.TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "XXXXXXXX",
		ConsumerKey:  "OOOOOOOOOOOOOOOOOOOOOOOOOO",
		RefreshToken: "#####################################",
	}

	go model.UpdateAccessToken(orderAPI)

	// 等待AccessToken 初始化
	time.Sleep(time.Second * 3)

	go model.UpdateLocalOrderStatus(2000, orderAPI)

	for symbol, Info := range MySymbol {
		go model.MainTrading(orderAPI, symbol, Info.Strategy, Info.Budget, Info.Strategy.GetOrderExpiredTime(), TradingDataChannels[symbol])
	}

	tradingSocket.Streaming(TradingDataChannels)

}
