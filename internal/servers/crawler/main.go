package main

import (
	cmd "github.com/anvh2/trading-bot/internal/servers/crawler/cmd"
)

const (
	version = "0.1.0"
)

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
