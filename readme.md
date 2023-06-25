# TDBOT

基於 TD Ameritrade API 自動交易系統

TD Ameritrade 已被 Charles Schwab 併購
API合併整合中 若券商API更新 有可能隨時會失效

## TDBOT佈署

參考環境
```

hccuse@us-east4-4c /pj> cat /etc/os-release 
PRETTY_NAME="Ubuntu 22.04.2 LTS"
NAME="Ubuntu"
VERSION_ID="22.04"
VERSION="22.04.2 LTS (Jammy Jellyfish)"
VERSION_CODENAME=jammy
ID=ubuntu
ID_LIKE=debian
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
UBUNTU_CODENAME=jammy

hccuse@us-east4-4c /pj> go version
go version go1.20.3 linux/amd64

```


參考目錄
```bash
hccuse@us-east4-4c /pj> pwd
/pj
hccuse@us-east4-4c /pj [1]> tree -L 1
.
├── GoBot
├── docker-compose.yaml
├── postgresql
└── redis
```


SocketToken預設路徑: /pj/finnhubToken/finn_token/finn_token.yaml

需三組 若重啟後 會隨機選擇一組使用

```yaml

collect1: XXXXXXXXXX
collect2: OOOOOOOOOO
collect3: HHHHHHHHHH
update: '20230501'

```
token from: https://finnhub.io/



需先定義帳戶metadata 策略 金額 於 formal/main/main.go 修改
策略自訂義可參考 model/Strategy.go

```
    // TD Ameritrade 帳戶基本訊息
   	orderAPI := utils.TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "1234567",
		// 帳戶id
		ConsumerKey:  "ZZZZZZZZZZZZZZZZZZ",
		// devekop ConsumerKey
		RefreshToken: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
	    // reflesh token
	}
	
	// 標的: *model.SymbolBase{標的,策略,金額}
	// 這邊意思為 AAPL 使用策略為 model.MyDemoStrategy, 2000刀
	// 可多個標的 策略可自定義 每個標的可使用不同策略 只需符合該接口
	// 策略自訂義 只需符合 model.StrategyInterface , demo 可參考 model/Strategy.go
	
    var MySymbol = map[string]*model.SymbolBase{
	"AAPL": {Symbol: "AAPL", Strategy: &model.MyDemoStrategy{}, Budget: decimal.NewFromInt(2000)},
    }
```

修改完後重新編譯 formal/main/main.go

```

go build main.go

```


docker-compose 預設路徑: /pj/docker-compose.yaml

啟動
```
dcoker-compose up -d
```


關閉
```
dcoker-compose down
```
