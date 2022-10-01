package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start order-status-consumer service",
	Long:  "Start order-status-consumer service",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Please input your service name ...")
		return errors.New("invalid arguments")
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
