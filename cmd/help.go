// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	helpCmd = &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long: `Help provides help for any command in the application.
Simply type ` + GaugeCmd.Name() + ` help [path to command] for full details.`,

		Run: func(c *cobra.Command, args []string) {
			if legacy {
				fmt.Println(legacyUsage)
				return
			}
			cmd, _, e := c.Root().Find(args)
			if cmd == nil || e != nil {
				c.Printf("Unknown help topic %#q\n", args)
				c.Root().Usage()
			} else {
				cmd.InitDefaultHelpFlag() // make possible 'help' flag to be shown
				cmd.Help()
			}
		},
	}
	legacy      bool
	legacyUsage = `Gauge 0.8.5 and below

Options:
  --add-plugin=""                  [DEPRECATED] Use 'gauge add <plugin name>'
  --api-port=""                    [DEPRECATED] Use 'gauge daemon <port>'
  --check-updates=false            [DEPRECATED] Use 'gauge update -c'
  --daemonize=false                [DEPRECATED] Use 'gauge daemon <port>'
  --docs=""                        [DEPRECATED] Use 'gauge docs <plugin name> specs/'
  --env="default"                  [DEPRECATED] Use 'gauge run -e <env name>'
  --failed=false                   [DEPRECATED] Use 'gauge run --failed'
  --file, -f=""                    [DEPRECATED] Use 'gauge install <plugin name> -f <zip file>'
  --format=""                      [DEPRECATED] Use 'gauge format specs/'
  -g, --group=-1                   [DEPRECATED] Use 'gauge run -n 5 -g 1 specs/'
  --init=""                        [DEPRECATED] Use 'gauge init <template name>'
  --install=""                     [DEPRECATED] Use 'gauge install <plugin name>'
  --install-all=false              [DEPRECATED] Use 'gauge install --all'
  --list-templates=false           [DEPRECATED] Use 'gauge init --templates'
  --machine-readable=false         [DEPRECATED] Use 'gauge version -m'
  -n=8                             [DEPRECATED] Use 'gauge run -p -n specs/'
  --parallel, -p=false             [DEPRECATED] Use 'gauge run -p specs/'
  --args=""                        [DEPRECATED] Use 'gauge add <plugin name> --args <args>'
  --plugin-version=""              [DEPRECATED] Use 'gauge [install|uninstall] <plugin name> -v <version>'
  --refactor=""                    [DEPRECATED] Use 'gauge refactor <old step> <new step>'
  --simple-console=false           [DEPRECATED] Use 'gauge run --simple-console'
  --sort, -s=false                 [DEPRECATED] Use 'gauge run -s specs'
  --strategy="lazy"                [DEPRECATED] Use 'gauge run -p --strategy=\"eager\"'
  --table-rows=""                  [DEPRECATED] Use 'gauge run --table-rows <rows>'
  --tags=""                        [DEPRECATED] Use 'gauge run --tags tag1,tag2 specs'
  --uninstall=""                   [DEPRECATED] Use 'gauge uninstall <plugin name>'
  --update=""                      [DEPRECATED] Use 'gauge update <plugin name>'
  --update-all=false               [DEPRECATED] Use 'gauge update -a'
  -v, --version, -version=false    [DEPRECATED] Use 'gauge version'
  --validate=false                 [DEPRECATED] Use 'gauge validate specs'
  --verbose=false                  [DEPRECATED] Use 'gauge run -v'`
)

func init() {
	helpCmd.Flags().BoolVarP(&legacy, "legacy", "", false, "Shows usage for old structure of gauge command")
	GaugeCmd.SetHelpCommand(helpCmd)
}
