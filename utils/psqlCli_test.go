package utils

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMarketOpen(t *testing.T) {
	ctx := context.Background()
	PsqlCLi := PsqlClient{}

	s := []string{"U", "TSLA"}

	for _, val := range s {
		PsqlCLi.GetSymbolQuery(val)
		count := 0
		dates := make([]time.Time, 0)
		for i := 0; i < 100; i++ {
			date := time.Date(2023, 3, 7, 0, 0, 0, 0, Loc).Add(time.Hour * -24 * time.Duration(i))
			res, err := PsqlCLi.Conn.HisDateMarketOpen(ctx, date)
			if err != nil {
				fmt.Println(err)
			}

			if res == true {
				count += 1
				dates = append(dates, date)
				if count == 10 {
					break
				}
			}

		}
		fmt.Println("latest 10 market days")
		for _, val := range dates {
			fmt.Println(val)
		}
		PsqlCLi.CloseSession()
	}

}

func TestPsqlClient_GetHisData(t *testing.T) {
	PsqlCLi := PsqlClient{}
	data := PsqlCLi.GetHisData("U", "2023-03-07", 10)
	for _, val := range data {
		fmt.Println(*val)
	}
}

func TestPsqlClient_GetNMinHisData(t *testing.T) {
	PsqlCLi := PsqlClient{}
	data := PsqlCLi.GetNMinHisData("U", "2023-02-22", "2023-02-23", 9, 10, 5)

	for _, v := range data {
		fmt.Println(v)
	}

}
