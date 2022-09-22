package cmd

import (
	"github.com/anvh2/trading-bot/internal/server"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start order-status-consumer service",
	Long:  "Start order-status-consumer service",
	RunE: func(cmd *cobra.Command, args []string) error {
		server := server.NewServer()
		if err := server.Setup(); err != nil {
			return err
		}

		return server.Start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
