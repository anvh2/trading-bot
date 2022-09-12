// Copyright © 2017 Alessandro Sanino <saninoale@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package bot

import (
	"io/ioutil"
	"os"

	"github.com/anvh2/trading-bot/internal/models"
	"github.com/anvh2/trading-bot/internal/server"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts trading using saved configs",
	Long:  `Starts trading using saved configs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initConfigs(); err != nil {
			return err
		}

		srv := server.NewServer(botConfig.ExchangeConfig)
		return srv.Start()
	},
}

var botConfig models.BotConfig

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolVarP(&startFlags.Simulate, "simulate", "s", false, "Simulates the trades instead of actually doing them")
}

func initConfigs() error {
	configFile, err := os.Open(GlobalFlags.ConfigFile)
	if err != nil {
		return err
	}

	value, err := ioutil.ReadAll(configFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(value, &botConfig)
	if err != nil {
		return err
	}

	return nil
}
