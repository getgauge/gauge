# PIP Installer for getgauge-cli tool

This `setup.py` file is used for installed getgauge-cli tool.
the version number is manually bumped for now.

Python package can be installed through multiple ways

## using setup.py

- Navigate to the current directory `gauge\build\pip\`
- Run `python build.py --install` on command prompt (it will install the latest released version)
- To install a specific version set `GAUGE_VERSION` environment variable.

## using pip

- pre-requisite: pip should be installed and available on machine
- run the command `pip install getgauge-cli`

## Check to ensure getgauge-cli is installed

- Once the package has been setup. please exit the current terminal and relaunch terminal again
- run the command `gauge` , you should be able to see similar output
```
Usage:
  gauge <command> [flags] [args]

Examples:
  gauge run specs/
  gauge run --parallel specs/

Commands:
  config      Change global configurations
  daemon      Run as a daemon
  docs        Generate documentation using specified plugin
  format      Formats the specified spec files
  help        Help about any command
  init        Initialize project structure in the current directory
  install     Download and install plugin(s)
  list        List specifications, scenarios or tags for a gauge project
  man         Generate man pages
  run         Run specs
  uninstall   Uninstalls a plugin
  update      Updates a plugin
  validate    Check for validation and parse errors
  version     Print Gauge and plugin versions

Flags:
  -d, --dir string         Set the working directory for the current command, accepts a path relative to current directory (default ".")
  -h, --help               Help for gauge
  -l, --log-level string   Set level of logging to debug, info, warning, error or critical (default "info")
  -m, --machine-readable   Prints output in JSON format
  -v, --version            Print Gauge and plugin versions

Use "gauge [command] --help" for more information about a command.
Complete manual is available at https://manpage.gauge.org/.
```

## Using setup.py
- Install required dep by running `python/python3 -m pip install requirements.txt`.
- Check through and obtain a valid tag/build number from [releases](https://github.com/getgauge/gauge/releases)
- Run the command `python build.py --install` (set GAUGE_VERSION env to install specific version)
- This would install the version as specified along with latest release of Gauge-CLI Version

- Run the command `python build.py --dist` to generate the PyPi distrubutable (set GAUGE_VERSION env to install specific version)

## Uninstalling Gauge CLI

Gauge CLI uninstall should be done manually for now.
Run the following command on your prompt
```
$ pip uninstall gauge-cli
```
