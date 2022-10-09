package helpers

func ResolvePositionSide(rsi float64) string {
	if rsi >= 70 {
		return "SHORT"
	} else if rsi <= 30 {
		return "LONG"
	}
	return ""
}
