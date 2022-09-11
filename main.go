package main

import (
	bot "github.com/anvh2/trading-boy/cmd"
)

const (
	version = "0.0.1-pre-alpha"
)

func main() {
	bot.SetVersion(version)
	bot.Execute()
}
