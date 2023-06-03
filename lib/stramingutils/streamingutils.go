package dbutils

import (
	"github.com/shopspring/decimal"
	"reflect"
	"time"
)

type TradingDataSet struct {
	Symbols []string
	Data    map[string]*TradingData
}

func (self *TradingDataSet) Init() {
	self.Data = make(map[string]*TradingData)
	for _, each_data := range self.Symbols {
		self.Data[each_data] = new(TradingData)
		self.Data[each_data].Data = make([]symbolData, 0)
	}
}

func (self *TradingDataSet) GetField() {

	for _, symbol := range self.Symbols {
		if len(self.Data[symbol].Data) > 0 {
			self.Data[symbol].GetField()
			// fmt.Println("symbol ", self.Data[symbol].Symbol, "price ", self.Data[symbol].Price, "vol ", self.Data[symbol].Volume, "evenNum ", self.Data[symbol].EventNum, "insRate ", self.Data[symbol].InsRate, "time ", self.Data[symbol].TradingTime)
		}
	}

}

// 同一個時間 單symbol 總資料
type TradingData struct {
	Symbol      string
	Price       decimal.Decimal
	Volume      decimal.Decimal
	InsRate     uint8 // c 含有8的比例
	NextRate    uint8
	EventNum    uint16
	TradingTime time.Time
	FormT       bool
	Data        []symbolData
}

// 計算同一個時間 該標的的統計資料
func (self *TradingData) GetField() {
	insvol := decimal.NewFromInt(0)
	nextvol := decimal.NewFromInt(0)
	totalPV := decimal.NewFromInt(0)
	for _, val := range self.Data {
		if val.ins == true {
			insvol = insvol.Add(val.vol)
		}
		if val.nextDay == true {
			nextvol = nextvol.Add(val.vol)
		}
		self.Volume = self.Volume.Add(val.vol)
		totalPV = totalPV.Add(val.vol.Mul(val.price))
		self.EventNum += 1
		self.TradingTime = val.tradingtime
		self.FormT = val.formT
		self.Symbol = val.symbol
	}
	self.InsRate = uint8(insvol.Div(self.Volume).Mul(decimal.NewFromInt(100)).IntPart())
	self.NextRate = uint8(nextvol.Div(self.Volume).Mul(decimal.NewFromInt(100)).IntPart())
	self.Price = totalPV.Div(self.Volume).Round(4)
}

// 同一個時間 symbol 單筆資料

type symbolData struct {
	symbol      string
	price       decimal.Decimal
	vol         decimal.Decimal
	ins         bool
	formT       bool
	nextDay     bool
	tradingtime time.Time
}

func (self *TradingDataSet) Parser(msg interface{}) {
	val, ok := msg.(map[string]interface{})
	if ok {
		for k0, v0 := range val {
			if k0 == "data" {
				v0 := v0.([]interface{})
				//size := len(v0)
				for _, val := range v0 {
					//fmt.Println(val)
					val := val.(map[string]interface{})
					data := symbolData{}
					data.symbol = val["s"].(string)
					if data.symbol == "" {
						continue
					}
					data.vol = decimal.NewFromFloat(val["v"].(float64))
					data.price = decimal.NewFromFloat(val["p"].(float64))
					data.tradingtime = time.UnixMilli(int64(val["t"].(float64)))
					r_slice := reflect.ValueOf(val["c"])
					for i := 0; i < r_slice.Len(); i++ {
						id := r_slice.Index(i).Interface().(string)
						switch id {
						case "8":
							data.ins = true
						case "24":
							data.formT = true
						case "17":
							data.nextDay = true
						}
					}
					self.Data[data.symbol].Data = append(self.Data[data.symbol].Data, data)
				}
				self.GetField()
			}
		}
	}

	//GOOG := *self.Data["GOOG"]
	//AAPL := *self.Data["AAPL"]
	//fmt.Println(GOOG.Price, GOOG.EventNum, GOOG.InsRate, GOOG.TradingTime, GOOG.Volume, GOOG.FormT)
	//fmt.Println(AAPL.Price, AAPL.EventNum, GOOG.InsRate, GOOG.TradingTime, AAPL.Volume, AAPL.FormT)
	//fmt.Println("symbol: ", self.Symbol, " price: ", self.Price, " vol: ", self.Volume, " eventNum: ", self.EventNum, " insRate: ", self.InsRate, " time: ", self.TradingTime, " formT:", self.FormT)
}
