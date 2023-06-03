package utils

import (
	"context"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
	"time"
)

var OrderTypeMap = map[int]string{10: "BuyLimit", 20: "SellShort", -10: "SellLimit", -20: "BuyToCover"}

type ResponseAccessToken struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type PlaceOrderBody struct {
	ComplexOrderStrategyType string `json:"complexOrderStrategyType"`
	OrderType                string `json:"orderType"`
	Session                  string `json:"session"`
	Price                    string `json:"price"`
	Duration                 string `json:"duration"`
	OrderStrategyType        string `json:"orderStrategyType"`
	OrderLegCollection       []struct {
		Instruction string `json:"instruction"`
		Quantity    int    `json:"quantity"`
		Instrument  struct {
			Symbol    string `json:"symbol"`
			AssetType string `json:"assetType"`
		} `json:"instrument"`
	} `json:"orderLegCollection"`
}

// PlaceOrderSchema
type PlaceOrder struct {
	ComplexOrderStrategyType string               `json:"complexOrderStrategyType"`
	OrderType                string               `json:"orderType"`
	Session                  string               `json:"session"`
	Price                    string               `json:"price"`
	Duration                 string               `json:"duration"`
	OrderStrategyType        string               `json:"orderStrategyType"`
	OrderLegCollection       []OrderLegCollection `json:"orderLegCollection"`
}

type OrderLegCollection struct {
	Instruction string     `json:"instruction"`
	Quantity    int        `json:"quantity"`
	Instrument  Instrument `json:"instrument"`
}

type Instrument struct {
	Symbol    string `json:"symbol"`
	AssetType string `json:"assetType"`
}

// 可看到所有
type OrderStatus []struct {
	Session                  string  `json:"session"`
	Duration                 string  `json:"duration"`
	OrderType                string  `json:"orderType"`
	ComplexOrderStrategyType string  `json:"complexOrderStrategyType"`
	Quantity                 float64 `json:"quantity"`
	FilledQuantity           float64 `json:"filledQuantity"`
	RemainingQuantity        float64 `json:"remainingQuantity"`
	RequestedDestination     string  `json:"requestedDestination"`
	DestinationLinkName      string  `json:"destinationLinkName"`
	Price                    float64 `json:"price"`
	OrderLegCollection       []struct {
		OrderLegType string `json:"orderLegType"`
		LegId        int    `json:"legId"`
		Instrument   struct {
			AssetType string `json:"assetType"`
			Cusip     string `json:"cusip"`
			Symbol    string `json:"symbol"`
		} `json:"instrument"`
		Instruction    string  `json:"instruction"`
		PositionEffect string  `json:"positionEffect"`
		Quantity       float64 `json:"quantity"`
	} `json:"orderLegCollection"`
	OrderStrategyType       string `json:"orderStrategyType"`
	OrderId                 int64  `json:"orderId"`
	Cancelable              bool   `json:"cancelable"`
	Editable                bool   `json:"editable"`
	Status                  string `json:"status"`
	EnteredTime             string `json:"enteredTime"`
	CloseTime               string `json:"closeTime"`
	Tag                     string `json:"tag"`
	AccountId               int    `json:"accountId"`
	OrderActivityCollection []struct {
		ActivityType           string  `json:"activityType"`
		ActivityId             int64   `json:"activityId"`
		ExecutionType          string  `json:"executionType"`
		Quantity               float64 `json:"quantity"`
		OrderRemainingQuantity float64 `json:"orderRemainingQuantity"`
		ExecutionLegs          []struct {
			LegId             int     `json:"legId"`
			Quantity          float64 `json:"quantity"`
			MismarkedQuantity float64 `json:"mismarkedQuantity"`
			Price             float64 `json:"price"`
			Time              string  `json:"time"`
		} `json:"executionLegs"`
	} `json:"orderActivityCollection"`
	ChildOrderStrategies []struct {
		Session                  string  `json:"session"`
		Duration                 string  `json:"duration"`
		OrderType                string  `json:"orderType"`
		ComplexOrderStrategyType string  `json:"complexOrderStrategyType"`
		Quantity                 float64 `json:"quantity"`
		FilledQuantity           float64 `json:"filledQuantity"`
		RemainingQuantity        float64 `json:"remainingQuantity"`
		RequestedDestination     string  `json:"requestedDestination"`
		DestinationLinkName      string  `json:"destinationLinkName"`
		Price                    float64 `json:"price"`
		OrderLegCollection       []struct {
			OrderLegType string `json:"orderLegType"`
			LegId        int    `json:"legId"`
			Instrument   struct {
				AssetType string `json:"assetType"`
				Cusip     string `json:"cusip"`
				Symbol    string `json:"symbol"`
			} `json:"instrument"`
			Instruction    string  `json:"instruction"`
			PositionEffect string  `json:"positionEffect"`
			Quantity       float64 `json:"quantity"`
		} `json:"orderLegCollection"`
		OrderStrategyType string `json:"orderStrategyType"`
		OrderId           int64  `json:"orderId"`
		Cancelable        bool   `json:"cancelable"`
		Editable          bool   `json:"editable"`
		Status            string `json:"status"`
		EnteredTime       string `json:"enteredTime"`
		Tag               string `json:"tag"`
		AccountId         int    `json:"accountId"`
	} `json:"childOrderStrategies"`
}

type ChildOrderStrategies struct {
	OrderType          string               `json:"orderType"`
	Session            string               `json:"session"`
	Price              string               `json:"price"`
	Duration           string               `json:"duration"`
	OrderStrategyType  string               `json:"orderStrategyType"`
	OrderLegCollection []OrderLegCollection `json:"orderLegCollection"`
}

type OTABody struct {
	OrderType            string                 `json:"orderType"`
	Session              string                 `json:"session"`
	Price                string                 `json:"price"`
	Duration             string                 `json:"duration"`
	OrderStrategyType    string                 `json:"orderStrategyType"`
	OrderLegCollection   []OrderLegCollection   `json:"orderLegCollection"`
	ChildOrderStrategies []ChildOrderStrategies `json:"childOrderStrategies"`
}

type UnitLimitOrder struct {
	Symbol      string    `redis:"Symbol"`
	OrderType   int       `redis:"OrderType"`
	OrderID     int64     `redis:"OrderID"`
	Quantity    int       `redis:"Quantity"`
	Price       string    `redis:"Price"`
	Status      string    `redis:"Status"`
	CreateTime  time.Time `redis:"time"`
	Cancelable  bool      `redis:"Cancelable"`
	Editable    bool      `redis:"Editable"`
	Description string    `redis:"Description"`
}

type TDOrder struct {
	AccountID                 string
	ConsumerKey               string
	Redirect_url              string
	RefreshToken              string
	AccessToken               string
	NextTokenUpdateTime       time.Time
	LastOrderStatusUpdateTime time.Time
}

type HisTransaction []TransactionUnit

type TransactionUnit struct {
	Type                  string  `json:"type"`
	SubAccount            string  `json:"subAccount"`
	SettlementDate        string  `json:"settlementDate"`
	NetAmount             float64 `json:"netAmount"`
	TransactionDate       string  `json:"transactionDate"`
	TransactionSubType    string  `json:"transactionSubType"`
	TransactionId         int64   `json:"transactionId"`
	CashBalanceEffectFlag bool    `json:"cashBalanceEffectFlag"`
	Description           string  `json:"description"`
	Fees                  struct {
		RFee          int     `json:"rFee"`
		AdditionalFee int     `json:"additionalFee"`
		CdscFee       int     `json:"cdscFee"`
		RegFee        float64 `json:"regFee"`
		OtherCharges  int     `json:"otherCharges"`
		Commission    int     `json:"commission"`
		OptRegFee     int     `json:"optRegFee"`
		SecFee        float64 `json:"secFee"`
	} `json:"fees"`
	TransactionItem struct {
		AccountId   int     `json:"accountId"`
		Cost        float64 `json:"cost"`
		Amount      int     `json:"amount,omitempty"`
		Price       float64 `json:"price,omitempty"`
		Instruction string  `json:"instruction,omitempty"`
		Instrument  struct {
			Symbol    string `json:"symbol"`
			Cusip     string `json:"cusip"`
			AssetType string `json:"assetType"`
		} `json:"instrument,omitempty"`
	} `json:"transactionItem"`
	OrderId   string `json:"orderId,omitempty"`
	OrderDate string `json:"orderDate,omitempty"`
}

func (self *TDOrder) GetAccessToken(ctx context.Context) (httpStatusCode int, accessToken string) {

	// 下次更新時間非預設值 或  時間更新與 現在時間比過期時間還大 , 從api取得token
	if self.NextTokenUpdateTime.IsZero() || (!self.NextTokenUpdateTime.IsZero() && time.Now().After(self.NextTokenUpdateTime)) {
		accessResp := &ResponseAccessToken{}
		url := fmt.Sprintf("https://api.tdameritrade.com/v1/oauth2/token")
		myclient := req.NewClient()
		request := myclient.NewRequest().SetContext(ctx)
		request = request.SetHeader("Content-Type", "application/x-www-form-urlencoded")
		bodyschema := make(map[string]string)
		bodyschema["grant_type"] = "refresh_token"
		bodyschema["refresh_token"] = self.RefreshToken
		bodyschema["access_type"] = ""
		bodyschema["code"] = ""
		bodyschema["client_id"] = self.ConsumerKey
		bodyschema["redirect_uri"] = ""
		request = request
		resp, err := request.SetFormData(bodyschema).SetSuccessResult(accessResp).Post(url)
		if err != nil {
			logrus.Warnln(err)
			return 500, ""
		}

		if resp.IsSuccessState() {
			self.NextTokenUpdateTime = time.Now().Add(time.Second * 1500)
			self.AccessToken = accessResp.AccessToken
			logrus.Infoln("set NextTokenUpdateTime: ", self.NextTokenUpdateTime)
			logrus.Infoln("success to get access token from API")
			return resp.StatusCode, self.AccessToken
		} else {
			logrus.Infoln("fail to get access token from API")
			return resp.StatusCode, ""
		}
	} else { // token 尚未過期

		logrus.Infoln("http status code: ", 200, "success to get access token from local")
		return 200, self.AccessToken

	}

}

// status: working,filled,canceled or "" ,
func (self *TDOrder) GetCurrentOrderStatus(ctx context.Context, startDate string, EndDate string, status string, maxResults string) (httpStatusCode int, orders *OrderStatus) {
	code, accessToken := self.GetAccessToken(ctx)
	if code < 200 && code > 299 {
		logrus.Infoln("can't get accessToken")
		panic("can't get accessToken")
	}
	respSchema := OrderStatus{}
	myclient := req.NewClient() //.DevMode()
	request := myclient.NewRequest().SetContext(ctx)
	myheader := make(map[string]string)
	myheader["Authorization"] = "Bearer " + accessToken
	request = request.SetHeaders(myheader).SetSuccessResult(&respSchema)
	url := fmt.Sprintf("https://api.tdameritrade.com/v1/orders?accountId=%v&fromEnteredTime=%v&toEnteredTime=%v&status=%v&maxResults=%v", self.AccountID, startDate, EndDate, status, maxResults)
	resp, err := request.Get(url)

	if err != nil {
		logrus.Warnln(err)
		return 500, nil
	}
	if resp.IsSuccessState() {
		logrus.Infoln(resp.StatusCode)
		//logrus.Infof("get order status: %+v", respSchema)
		self.LastOrderStatusUpdateTime = time.Now().In(Loc)
		return resp.StatusCode, &respSchema
	} else if resp.IsErrorState() {
		logrus.Warnln(resp.StatusCode)
	}

	return resp.StatusCode, nil
}

// InstructionNumber: Buy:10, sell short:20 , sell: -10, buy_to_cover: -20
func (self *TDOrder) CreateLimitOrder(ctx context.Context, LimitOrder *UnitLimitOrder) (httpStatusCode int) {
	code, accessToken := self.GetAccessToken(ctx)
	if code < 200 && code > 299 {
		logrus.Infoln("can't get accessToken")
		panic("can't get accessToken")
	}
	logrus.Infoln("apply limit order")
	logrus.Infof("refer: %+v", LimitOrder)
	instruction := map[int]string{10: "BUY", 20: "SELL_SHORT", -10: "SELL", -20: "BUY_TO_COVER"}

	myinstructment := Instrument{Symbol: LimitOrder.Symbol, AssetType: "EQUITY"}

	ordercollection := OrderLegCollection{
		Instruction: instruction[LimitOrder.OrderType],
		Quantity:    LimitOrder.Quantity,
		Instrument:  myinstructment,
	}

	myorder := PlaceOrder{OrderType: "LIMIT",
		ComplexOrderStrategyType: "None",
		Session:                  "NORMAL",
		Price:                    LimitOrder.Price,
		Duration:                 "DAY",
		OrderStrategyType:        "SINGLE",
		OrderLegCollection:       []OrderLegCollection{ordercollection}}

	myclient := req.NewClient() //.DevMode()
	request := myclient.NewRequest().SetContext(ctx)
	myheader := make(map[string]string)
	myheader["Content-Type"] = "application/json"
	myheader["Authorization"] = "Bearer " + accessToken
	request = request.SetHeaders(myheader).SetBody(myorder)
	url := fmt.Sprintf("https://api.tdameritrade.com/v1/accounts/%v/orders", self.AccountID)
	resp, err := request.Post(url)
	if err != nil {
		logrus.Warnln(err)
		return 500
	}
	logrus.Infoln("commit limit order")

	if resp.IsSuccessState() {
		LimitOrder.CreateTime = time.Now().In(Loc)
		logrus.Infoln(resp.StatusCode)
	} else if resp.IsErrorState() {
		logrus.Warnln(resp.StatusCode)
	}

	return resp.StatusCode
}

// 只能同標的 一買一賣
func (self *TDOrder) CreateOTAOrder(ctx context.Context, MainLimitOrder *UnitLimitOrder, TriggerLimitOrder *UnitLimitOrder) (httpStatusCode int) {
	code, accessToken := self.GetAccessToken(ctx)
	if code < 200 && code > 299 {
		logrus.Infoln("can't get accessToken")
		panic("can't get accessToken")
	}
	logrus.Infoln("apply OTA")
	logrus.Infof("main refer: %+v", MainLimitOrder)
	logrus.Infof("trigger refer: %+v", TriggerLimitOrder)

	instruction := map[int]string{10: "BUY", 20: "SELL_SHORT", -10: "SELL", -20: "BUY_TO_COVER"}

	MainInstrument := Instrument{Symbol: MainLimitOrder.Symbol, AssetType: "EQUITY"}
	TriggerInstrument := Instrument{Symbol: TriggerLimitOrder.Symbol, AssetType: "EQUITY"}

	mainordercollection := OrderLegCollection{
		Instruction: instruction[MainLimitOrder.OrderType],
		Quantity:    MainLimitOrder.Quantity,
		Instrument:  MainInstrument,
	}

	triggerordercollection := OrderLegCollection{
		Instruction: instruction[TriggerLimitOrder.OrderType],
		Quantity:    TriggerLimitOrder.Quantity,
		Instrument:  TriggerInstrument,
	}

	Triggerorder := ChildOrderStrategies{
		OrderType:          "LIMIT",
		Session:            "NORMAL",
		Price:              TriggerLimitOrder.Price,
		Duration:           "DAY",
		OrderStrategyType:  "SINGLE",
		OrderLegCollection: []OrderLegCollection{triggerordercollection},
	}

	body := &OTABody{
		OrderType:            "LIMIT",
		Session:              "NORMAL",
		Price:                MainLimitOrder.Price,
		Duration:             "DAY",
		OrderStrategyType:    "TRIGGER",
		OrderLegCollection:   []OrderLegCollection{mainordercollection},
		ChildOrderStrategies: []ChildOrderStrategies{Triggerorder},
	}
	url := fmt.Sprintf("https://api.tdameritrade.com/v1/accounts/%v/orders", self.AccountID)
	myclient := req.NewClient() //.DevMode()
	request := myclient.NewRequest().SetContext(ctx)
	myheader := make(map[string]string)
	myheader["Authorization"] = "Bearer " + accessToken
	myheader["Content-Type"] = "application/json"
	request = request.SetHeaders(myheader).SetBody(body)
	resp, err := request.Post(url)
	if err != nil {
		logrus.Warnln(err)
		return 500
	}
	logrus.Infoln("Commit OTA Order")

	if resp.IsSuccessState() {
		now := time.Now().In(Loc)
		MainLimitOrder.CreateTime = now
		TriggerLimitOrder.CreateTime = now
		logrus.Infoln(resp.StatusCode)
	} else if resp.IsErrorState() {
		logrus.Warnln(resp.StatusCode)
		logrus.Warnln(err)
	}

	return resp.StatusCode

}

func (self *TDOrder) DeleteOrder(ctx context.Context, orderID string) int {
	code, accessToken := self.GetAccessToken(ctx)
	if code < 200 && code > 299 {
		logrus.Infoln("can't get accessToken")
		panic("can't get accessToken")
	}
	url := fmt.Sprintf("https://api.tdameritrade.com/v1/accounts/%v/orders/%v", self.AccountID, orderID)
	myclient := req.NewClient() //.DevMode()
	request := myclient.NewRequest().SetContext(ctx)
	myheader := make(map[string]string)
	myheader["Authorization"] = "Bearer " + accessToken
	resp, err := request.SetHeaders(myheader).Delete(url)
	if err != nil {
		logrus.Warnln(err)
		return 500
	}
	logrus.Infoln("API commit delete order orderID: ", orderID)

	if resp.IsSuccessState() {
		logrus.Infoln(resp.StatusCode)
	} else if resp.IsErrorState() {
		logrus.Warnln(resp.StatusCode)
	}

	return resp.StatusCode
}

func (self *TDOrder) ReplaceOrder(ctx context.Context, OldOrderID string, LimitOrder *UnitLimitOrder) (httpStatusCode int) {
	code, accessToken := self.GetAccessToken(ctx)
	if code < 200 && code > 299 {
		logrus.Infoln("can't get accessToken")
		panic("can't get accessToken")
	}
	instruction := map[int]string{10: "BUY", 20: "SELL_SHORT", -10: "SELL", -20: "BUY_TO_COVER"}

	myinstructment := Instrument{Symbol: LimitOrder.Symbol, AssetType: "EQUITY"}

	ordercollection := OrderLegCollection{
		Instruction: instruction[LimitOrder.OrderType],
		Quantity:    LimitOrder.Quantity,
		Instrument:  myinstructment,
	}

	myorder := PlaceOrder{OrderType: "LIMIT",
		ComplexOrderStrategyType: "None",
		Session:                  "NORMAL",
		Price:                    LimitOrder.Price,
		Duration:                 "DAY",
		OrderStrategyType:        "SINGLE",
		OrderLegCollection:       []OrderLegCollection{ordercollection}}

	url := fmt.Sprintf("https://api.tdameritrade.com/v1/accounts/%v/orders/%v", self.AccountID, OldOrderID)

	myclient := req.NewClient() //.DevMode()
	request := myclient.NewRequest().SetContext(ctx)
	myheader := make(map[string]string)
	myheader["Authorization"] = "Bearer " + accessToken
	myheader["Content-Type"] = "application/json"
	request = request.SetHeaders(myheader).SetBody(&myorder)

	resp, err := request.Put(url)

	if err != nil {
		logrus.Warnln(err)
		return 500
	}

	logrus.Infoln("Replace OldOrderID: ", OldOrderID)
	logrus.Infoln("Refer NewOrder: ", myorder)

	if resp.IsSuccessState() {
		logrus.Infoln(resp.StatusCode)
		LimitOrder.CreateTime = time.Now().In(Loc)
	} else if resp.IsErrorState() {
		logrus.Warnln(resp.StatusCode, resp.Err)
	}

	return resp.StatusCode
}

func (self *TDOrder) ReplaceOpenOTAOrder(ctx context.Context, OldOrderID string, MainLimitOrder *UnitLimitOrder) (httpStatusCode int) {
	code, accessToken := self.GetAccessToken(ctx)
	if code < 200 && code > 299 {
		logrus.Infoln("can't get accessToken")
		panic("can't get accessToken")
	}
	logrus.Infoln("Replace OTA")
	logrus.Infof("main refer: %+v", MainLimitOrder)

	instruction := map[int]string{10: "BUY", 20: "SELL_SHORT", -10: "SELL", -20: "BUY_TO_COVER"}

	MainInstrument := Instrument{Symbol: MainLimitOrder.Symbol, AssetType: "EQUITY"}

	mainordercollection := OrderLegCollection{
		Instruction: instruction[MainLimitOrder.OrderType],
		Quantity:    MainLimitOrder.Quantity,
		Instrument:  MainInstrument,
	}

	myOrderBody := &OTABody{
		OrderType:          "LIMIT",
		Session:            "NORMAL",
		Price:              MainLimitOrder.Price,
		Duration:           "DAY",
		OrderStrategyType:  "TRIGGER",
		OrderLegCollection: []OrderLegCollection{mainordercollection},
	}
	url := fmt.Sprintf("https://api.tdameritrade.com/v1/accounts/%v/orders/%v", self.AccountID, OldOrderID)

	myclient := req.NewClient() //.DevMode()
	request := myclient.NewRequest().SetContext(ctx)
	myheader := make(map[string]string)
	myheader["Authorization"] = "Bearer " + accessToken
	myheader["Content-Type"] = "application/json"
	request = request.SetHeaders(myheader).SetBody(&myOrderBody)

	resp, err := request.Put(url)

	if err != nil {
		logrus.Warnln(err)
		return 500
	}

	logrus.Infoln("Replace OldOrderID: ", OldOrderID)
	logrus.Infoln("Refer NewOrder: ", myOrderBody)

	if resp.IsSuccessState() {
		logrus.Infoln(resp.StatusCode)
		MainLimitOrder.CreateTime = time.Now().In(Loc)
	} else if resp.IsErrorState() {
		logrus.Warnln(resp.StatusCode, resp.Err)
	}

	return resp.StatusCode

}

// TransactionType: ALL,TRADE,BUY_ONLY,SELL_ONLY
// Symbol: U,TSLA,GOOG... or ""
// TransactionType : BUY_ONLY , SELL_ONLY , 用其他schema 不同 會無法marshal近來
type UnitHisTransaction struct {
	Symbol          string
	TransactionType string
	StartDate       string
	EndDate         string
}

func (self *TDOrder) GetTransactionHistory(ctx context.Context, hisTransaction UnitHisTransaction) (httpStatusCode int, transaction *HisTransaction) {
	code, accessToken := self.GetAccessToken(ctx)
	if code < 200 && code > 299 {
		logrus.Infoln("can't get accessToken")
		panic("can't get accessToken")
	}
	transactions := HisTransaction{}
	url := fmt.Sprintf("https://api.tdameritrade.com/v1/accounts/%v/transactions?type=%v&symbol=%v&startDate=%v&endDate=%v", self.AccountID, hisTransaction.TransactionType, hisTransaction.Symbol, hisTransaction.StartDate, hisTransaction.EndDate)
	myclient := req.NewClient() //.DevMode()
	request := myclient.NewRequest().SetContext(ctx)
	myheader := make(map[string]string)
	myheader["Authorization"] = "Bearer " + accessToken
	resp, err := request.SetHeaders(myheader).SetSuccessResult(&transactions).Get(url)
	if err != nil {
		logrus.Infoln(err)
	}

	logrus.Infoln("Commit Get History Transaction")
	if resp.IsSuccessState() {
		logrus.Infoln(resp.StatusCode)
	} else if resp.IsErrorState() {
		logrus.Warnln(resp.StatusCode)
	}

	return resp.StatusCode, &transactions
}
