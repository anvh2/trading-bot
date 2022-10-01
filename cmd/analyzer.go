package cmd

import (
	"github.com/anvh2/trading-bot/internal/server/analyzer"
	"github.com/spf13/cobra"
)

// analyzerCmd represents the start command
var analyzerCmd = &cobra.Command{
	Use:   "analyzer",
	Short: "Start analyzer service",
	Long:  "Start analyzer service",
	RunE: func(cmd *cobra.Command, args []string) error {
		server := analyzer.New()
		return server.Start()
	},
}

func init() {
	RootCmd.AddCommand(analyzerCmd)
}
