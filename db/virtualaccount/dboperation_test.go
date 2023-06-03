package virtualaccount

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"testing"
)

const (
	USERNAME = "postgres"
	PASSWD   = "example"
	TIMEZONE = "EST5EDT"
)

func GetQuery(DBNAME string) *Queries {
	dbinfo := fmt.Sprintf("port=5432 host=localhost user=%v password=%v dbname=%v sslmode=disable timezone=%v", USERNAME, PASSWD, DBNAME, TIMEZONE)
	db, err := sql.Open("postgres", dbinfo)
	// open db connection pool

	db.SetMaxOpenConns(30)

	if err != nil {
		fmt.Println(err)
	}
	query := New(db)
	return query
}

func GetTxQuery(DBNAME string) (*Queries, *sql.Tx) {
	dbinfo := fmt.Sprintf("port=5432 host=localhost user=%v password=%v dbname=%v sslmode=disable", USERNAME, PASSWD, DBNAME)
	db, err := sql.Open("postgres", dbinfo)
	// open db connection pool

	db.SetMaxOpenConns(30)

	if err != nil {
		fmt.Println(err)
	}
	query := New(db)

	tx, err := db.Begin()

	if err != nil {
		fmt.Println(err)
	}

	return query, tx
}

func TestCreateNewAccount(t *testing.T) {

	query := GetQuery("virtualaccount")

	ctx := context.Background()

	info, err := query.CreateNewAccount(ctx, CreateNewAccountParams{AccountID: "12345", Balance: "10000"})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("account = ", info)

}

func TestQueryBalance(t *testing.T) {

	query := GetQuery("virtualaccount")

	ctx := context.Background()

	balance, err := query.QueryBalance(ctx, "12345")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("balance = ", balance)

}

//func TestDeleteAccont(t *testing.T) {
//
//	query := GetQuery("virtualaccount")
//
//	ctx := context.Background()
//
//	err := query.DeleteAccount(ctx, 12345)
//
//	if err != nil {
//		fmt.Println(err)
//	}
//
//}

func TestUpdateBalance(t *testing.T) {

	query := GetQuery("virtualaccount")

	ctx := context.Background()

	balance, err := query.QueryBalance(ctx, "12345")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("balance = ", balance)

	versionNow, err := query.get_version(ctx, "12345")

	_, err = query.AddAccountBalance(ctx, AddAccountBalanceParams{AccountID: "12345", Balance: "10000", Version: versionNow})
	if err != nil {
		fmt.Println(err)
	}
	balance, err = query.QueryBalance(ctx, "12345")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("balance = ", balance)
}

func TestCreateNewTask(t *testing.T) {
	ctx := context.Background()
	query := GetQuery("virtualaccount")
	info, err := query.CreateNewTask(ctx, CreateNewTaskParams{TaskID: "54322", AccountID: "12345", Symbol: "SE"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(info)
}

func TestCreateHisBuy(t *testing.T) {

	ctx := context.Background()
	query := GetQuery("virtualaccount")

	taskinfo, err := query.CreateNewTask(ctx, CreateNewTaskParams{AccountID: "12345", TaskID: "54322", Symbol: "SE"})
	holdSymbolinfo, err := query.UpdateHoldSymbolStatus(ctx, UpdateHoldSymbolStatusParams{Symbol: "SE", Quantity: 100, AccountID: "12345"})
	info, err := query.CreateHisBuy(ctx, CreateHisBuyParams{AccountID: "12345", TaskID: "54322", Symbol: "SE", BuyPrice: "102.5", Quantity: 100, Trigger: "s1"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(taskinfo)
	fmt.Println(holdSymbolinfo)
	fmt.Println(info)

}

func TestGetAccountTaskStatus(t *testing.T) {

	query := GetQuery("virtualaccount")
	ctx := context.Background()
	info, err := query.GetAccountTaskStatus(ctx, "12345")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(info)

	for _, val := range info {
		fmt.Println(val.AccountID, val.TaskID, val.Symbol)
	}
}

//func TestCreateHisSell(t *testing.T) {
//
//	ctx := context.Background()
//	query := GetQuery()
//	var sellnum int32 = 0
//	holdSymbolinfo, err := query.UpdateHoldSymbolStatus(ctx, UpdateHoldSymbolStatusParams{Symbol: "SE", Quantity: sellnum * -1, AccountID: "12345"})
//	buyPrice, err := query.GetTaskBuyPrice(ctx, "54322")
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(holdSymbolinfo)
//	buyPriceD, err := decimal.NewFromString(buyPrice)
//	if err != nil {
//		fmt.Println(err)
//	}
//	sellPrice := "105.323"
//	sellPriceD, err := decimal.NewFromString(sellPrice)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	sellnumD := decimal.NewFromInt(int64(sellnum))
//	income := sellPriceD.Sub(buyPriceD).Mul(sellnumD)
//
//	info, err := query.CreateHisSell(ctx, CreateHisSellParams{AccountID: "12345", TaskID: "54322", Symbol: "SE", SellPrice: sellPrice, Quantity: sellnum, Trigger: "s1", Income: income.String()})
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(info)
//
//	err = query.CloseTask(ctx, "54322")
//	if err != nil {
//		fmt.Println(err)
//	}
//	err = query.CloseSymbol(ctx)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//}

func TestTransaction(t *testing.T) {
	ctx := context.Background()
	query, tx := GetTxQuery("virtualaccount")

	defer tx.Rollback() // 實際上不影響 若中途失敗 return , 尚未被commit , 不會更新db, 但若全部執行完後 commit, 有rollback也沒用

	qtx := query.WithTx(tx)

	var sellnum int32 = 100
	buyPrice, err := qtx.GetTaskBuyPrice(ctx, "54322")
	if err != nil {
		fmt.Println(err)
		return
	}
	buyPriceD, err := decimal.NewFromString(buyPrice)
	if err != nil {
		fmt.Println(err)
		return
	}
	sellnumD := decimal.NewFromInt(int64(sellnum))

	sellPrice := "105.323"
	sellPriceD, err := decimal.NewFromString(sellPrice)
	income := sellPriceD.Sub(buyPriceD).Mul(sellnumD)

	// 原始持有100
	// 賣掉100, update至db
	holdSymbolinfo, err := qtx.UpdateHoldSymbolStatus(ctx, UpdateHoldSymbolStatusParams{Symbol: "SE", Quantity: sellnum * -1, AccountID: "12345"})
	info, err := qtx.CreateHisSell(ctx, CreateHisSellParams{AccountID: "12345", TaskID: "54322", Symbol: "SE", SellPrice: sellPrice, Quantity: sellnum, Trigger: "s1", Income: income.String()})

	// 若symbol 數量為零 刪除symbol, buy and sell 數量相同 刪除task
	err = qtx.CloseTask(ctx, "54322")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = qtx.CloseSymbol(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(holdSymbolinfo.Symbol, holdSymbolinfo.Quantity)
	fmt.Println(info.Symbol, info.Quantity, info.SellPrice, info.Income)

	//return
	// 直接跳出 避免執行完成

	if err = tx.Commit(); err != nil { // 這邊會提交至db, 提交後無法rollback
		return
	}

}
