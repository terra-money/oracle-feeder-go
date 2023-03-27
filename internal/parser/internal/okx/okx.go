package okx

import (
	"fmt"
	"strings"
)

func ParseSymbol(symbol string) (string, string, error) {
	symbol = strings.ToUpper(symbol)
	arr := strings.Split(symbol, "-")
	if len(arr) != 2 {
		return "", "", fmt.Errorf("failed to parse okx %s", symbol)
	}
	return arr[0], arr[1], nil
}
