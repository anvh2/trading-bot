package models

import "encoding/json"

type Stoch struct {
	RSI   float64 `json:"rsi"`
	SlowK float64 `json:"slow_k"`
	SlowD float64 `json:"slow_d"`
}

type Oscillator struct {
	Symbol string            `json:"symbol"`
	Stoch  map[string]*Stoch `json:"stoch"`
}

func (s *Oscillator) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}
