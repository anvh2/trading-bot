package helpers

import (
	"fmt"
	"math"
	"strconv"
)

func AlignPrice(rawPrice float64, stepSize string) float64 {
	step, _ := strconv.ParseFloat(stepSize, 64)
	precision := -math.Log10(step)
	round := math.Pow10(int(precision))
	return math.Round(rawPrice*round) / round
}

func AlignPriceToString(rawPrice float64, stepSize string) string {
	step, _ := strconv.ParseFloat(stepSize, 64)
	precision := -math.Log10(step)
	round := math.Pow10(int(precision))
	return fmt.Sprint(math.Round(rawPrice*round) / round)
}

func AlignQuantity(quantity float64, stepSize string) float64 {
	step, _ := strconv.ParseFloat(stepSize, 64)
	precision := -math.Log10(step)
	round := math.Pow10(int(precision))
	return math.Round((quantity+0.5)*round) / round
}

func AlignQuantityToString(quantity float64, stepSize string) string {
	step, _ := strconv.ParseFloat(stepSize, 64)
	precision := -math.Log10(step)
	round := math.Pow10(int(precision))
	return fmt.Sprint(math.Round((quantity+0.5)*round) / round)
}

func AmountToLotSize(lot float64, precision int, amount float64) float64 {
	return math.Trunc(math.Floor(amount/lot)*lot*math.Pow10(precision)) / math.Pow10(precision)
}
