package utils

import (
	"fmt"
	"testing"
)

func TestRandSymbolIndex(t *testing.T) {

	MyCandidate := []string{"AAPL", "MSFT", "NVDA", "TSLA", "AMZN"}

	AllNum := len(MyCandidate)
	PickNum := 1
	indexes := RandSymbolIndex(PickNum, AllNum)

	for _, val := range indexes {
		fmt.Println(MyCandidate[val])
	}
	fmt.Println(indexes)
}
