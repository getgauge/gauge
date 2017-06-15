package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var oldCmd = &cobra.Command{
	Use:    "old-usage",
	Short:  "Shows usage for old structure of gauge command",
	Long:   `Shows usage for old structure of gauge command`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`Gauge 0.8.4 and below

Options:
  --add-plugin=""                  [DEPRECATED] Use 'gauge add <plugin name>'
  --api-port=""                    [DEPRECATED] Use 'gauge daemon <port>'
  --check-updates=false            [DEPRECATED] Use 'gauge update -c'
  --daemonize=false                [DEPRECATED] Use 'gauge daemon'
  --docs=""                        [DEPRECATED] Use 'gauge docs <plugin name> specs/'
  --env="default"                  [DEPRECATED] Use 'gauge run -e <env name>'
  --failed=false                   [DEPRECATED] Use 'gauge run --failed'
  --file, -f=""                    [DEPRECATED] Use 'gauge install <plugin name> -f <zip file>'
  --format=""                      [DEPRECATED] Use 'gauge format specs/'
  -g, --group=-1                   [DEPRECATED] Use 'gauge -n 5 -g 1 specs/'
  --init=""                        [DEPRECATED] Use 'gauge init <template name>'
  --install=""                     [DEPRECATED] Use 'gauge install <plugin name>'
  --install-all=false              [DEPRECATED] Use 'gauge install --all'
  --list-templates=false           [DEPRECATED] Use 'gauge list-templates'
  --machine-readable=false         [DEPRECATED] Use 'gauge version -m'
  -n=8                             [DEPRECATED] Use 'gauge run -p -n specs/'
  --parallel, -p=false             [DEPRECATED] Use 'gauge run -p specs/'
  --plugin-args=""                 [DEPRECATED] Use 'gauge add <plugin name> --plugin-args <args>'
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
  --verbose=false                  [DEPRECATED] Use 'gauge run -v'
`)
	},
}

func init() {
	GaugeCmd.AddCommand(oldCmd)
}
