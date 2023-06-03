package model

import (
	"fmt"
	"testing"
	"time"
)

func TestUpdateAccessToken(t *testing.T) {
	go UpdateAccessToken()

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
