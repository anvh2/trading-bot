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
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/anvh2/trading-bot/internal/models"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the bot to trade",
	Long: `Initializes the trading bot: it will ask several questions to properly create a conf file.
	It must be run prior any other command if config file is not present.`,
	Run: executeInitCommand,
}

func init() {
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initFlags.ConfigFile, "import", "", "imports configuration from a file.")
}

func executeInitCommand(cmd *cobra.Command, args []string) {
	initConfig()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if initFlags.ConfigFile != "" {
		//try first to unmarshal the file to check if it is correct format.
		content, err := ioutil.ReadFile(initFlags.ConfigFile)
		if err != nil {
			fmt.Print("Error while opening the config file provided")
			if GlobalFlags.Verbose > 0 {
				fmt.Printf(": %s", err.Error())
			}
			fmt.Println()
			return
		}
		var checker models.BotConfig
		err = yaml.Unmarshal(content, &checker)
		if err != nil {
			fmt.Print("Cannot load provided configuration file")
			if GlobalFlags.Verbose > 0 {
				fmt.Printf(": %s", err.Error())
			}
			fmt.Println()
			return
		}
		err = ioutil.WriteFile("./.bot_config.yml", content, 0666)
		if err != nil {
			fmt.Print("Cannot write new configuration file")
			if GlobalFlags.Verbose > 0 {
				fmt.Printf(": %s", err.Error())
			}
			fmt.Println()
			return
		}
	}
}
