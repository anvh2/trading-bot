package cmd

import (
	"github.com/anvh2/trading-bot/internal/server/crawler"
	"github.com/spf13/cobra"
)

// crawlerCmd represents the start command
var crawlerCmd = &cobra.Command{
	Use:   "crawler",
	Short: "Start crawler service",
	Long:  "Start crawler service",
	RunE: func(cmd *cobra.Command, args []string) error {
		server := crawler.New()
		return server.Start()
	},
}

func init() {
	RootCmd.AddCommand(crawlerCmd)
}
