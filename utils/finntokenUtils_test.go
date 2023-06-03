package utils

import (
	"fmt"
	"testing"
)

func TestFinnToken(t *testing.T) {
	file := "/home/hccuse/Insync/y40103@gmail.com/Google Drive/hccuse/hccuse/learn/quan/GoBot/finn_token.yaml"

	token := FinnToken{}

	fmt.Println(token.GetToken(file))
}
