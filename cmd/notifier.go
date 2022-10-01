package cmd

import (
	"github.com/anvh2/trading-bot/internal/server/notifier"
	"github.com/spf13/cobra"
)

// notifierCmd represents the start command
var notifierCmd = &cobra.Command{
	Use:   "notifier",
	Short: "Start notifier service",
	Long:  "Start notifier service",
	RunE: func(cmd *cobra.Command, args []string) error {
		server := notifier.New()
		return server.Start()
	},
}

func init() {
	RootCmd.AddCommand(notifierCmd)
}
