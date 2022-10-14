package main

import (
	"fmt"
	"os"

	cmd "github.com/anvh2/trading-bot/internal/servers/commander/cmd"
)

const (
	version = "0.1.0"
)

func main() {
	fmt.Println(os.Getenv("TOKEN"))
	cmd.SetVersion(version)
	cmd.Execute()
}
