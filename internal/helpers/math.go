package helpers

import (
	"fmt"
	"strconv"
)

func StringToFloat(val string) float64 {
	result, _ := strconv.ParseFloat(val, 64)
	return result
}

func AddFloat(data ...string) float64 {
	result := 0.0
	for _, val := range data {
		result += StringToFloat(val)
	}

	return result / float64(len(data))
}

func Div(fraction, numerator string) string {
	f, _ := strconv.ParseFloat(fraction, 64)
	n, _ := strconv.ParseFloat(numerator, 64)
	return fmt.Sprint(f / n)
}
