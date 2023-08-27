package model

import (
	"GoBot/utils"
	"fmt"
	"testing"
	"time"
)

func TestUpdateAccessToken(t *testing.T) {
	orderAPI := utils.TDOrder{Redirect_url: "https://localhost:8080",
		AccountID:    "XXXXXXXX",
		ConsumerKey:  "OOOOOOOOOOOOOOOOOOOOOOOOOO",
		RefreshToken: "#####################################",
	}

	go UpdateAccessToken(orderAPI)

	go func() {

		for {
			time.Sleep(time.Second * 3)

			fmt.Println("#1 Token: ", AccessToken)

		}
	}()

	for {
		time.Sleep(time.Second * 3)

		fmt.Println("#2 Token: ", AccessToken)

	}

}
