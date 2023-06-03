
[//]: # (## db schema 佈署 功能暫定 未實現 可忽略)

[//]: # ()
[//]: # (### account step)

[//]: # ()
[//]: # (```bash)

[//]: # ()
[//]: # (cd ../)

[//]: # (migrate create -ext sql -dir ./migration_sqlc -seq create_virtualaccount_table)

[//]: # ()
[//]: # (```)

[//]: # ()
[//]: # (填入sql schema)

[//]: # ()
[//]: # (create與drop該schema方法)

[//]: # ()
[//]: # (```bash)

[//]: # (export POSTGRESQL_URL="postgres://postgres:example@localhost:5432/virtualaccount?sslmode=disable")

[//]: # ()
[//]: # (migrate -database ${POSTGRESQL_URL} -path ./ up)

[//]: # (migrate -database ${POSTGRESQL_URL} -path ./ down)

[//]: # (# 使用單一檔案 進行migration)

[//]: # ()
[//]: # (# or )

[//]: # (cd migration_sqlc)

[//]: # ()
[//]: # (migrate -verbose -database ${POSTGRESQL_URL} -source file://migration up 1)

[//]: # (migrate -verbose -database ${POSTGRESQL_URL} -source file://migration down 1)

[//]: # (# 利用目錄 進行migration)

[//]: # ()
[//]: # ()
[//]: # (```)

[//]: # ()
[//]: # ()
[//]: # ()
[//]: # ()
[//]: # (```sql)

[//]: # ()
[//]: # (insert into accountinfo &#40;account_id,balance&#41; values &#40;0,10000&#41;;)

[//]: # ()
[//]: # (```)

[//]: # ()
[//]: # ()
[//]: # ()
[//]: # ()
[//]: # (### symbolStep)

[//]: # ()
[//]: # (```bash)

[//]: # (migrate create -ext sql -dir ./migration_sqlc -seq create_U_table)

[//]: # (```)

[//]: # ()
[//]: # ()
[//]: # (```bash)

[//]: # (export POSTGRESQL_URL="postgres://postgres:example@localhost:5432/SE?sslmode=disable")

[//]: # ()
[//]: # (cd migration_sqlc)

[//]: # ()
[//]: # (migrate -verbose -database ${POSTGRESQL_URL} -source file://migration up 1)

[//]: # (migrate -verbose -database ${POSTGRESQL_URL} -source file://migration down 1)

[//]: # (# 利用目錄 進行migration)

[//]: # ()
[//]: # ()
[//]: # (```)




## Gobot佈署


參考目錄
```bash
hccuse@us-east4-4c /pj> pwd
/pj
hccuse@us-east4-4c /pj [1]> tree -L 1
.
├── Dockerfile
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
	// 可多個標的 策略可自訂義 每個標的可使用不同策略 只需符合該街口
	// 策略自訂義 只需符合 model.StrategyInterface
	
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
