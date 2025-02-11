package helper

import (
	"fmt"
	"strings"
)

func FancyFloat(nr float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", nr), "0"), ".")
}
