# PIP Installer for getgauge-cli tool

This `setup.py` file is used for installed getgauge-cli tool.
the version number is manually bumped for now.

Python package can be installed through multiple ways

#### using setup.py
- navigate to the current directory `gauge\build\pip\`
- run `python setup.py install` on command prompt

#### using pip
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
  refactor    Refactor steps
  run         Run specs
  telemetry   Configure options for sending anonymous usage stats
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

## Installing an older version gauge 
Easiest way to deal with an older version of gauge is using setup.py

### Using setup.py
- Check through and obtain a valid tag/build number from [releases](https://github.com/getgauge/gauge/releases)
- Navigate to line # 15 of `setup.py` file and update the the value of `release_ver` variable to be something similar to  release_ver = '1.0.2'
- save the changes made to file and close the editor
- navigate to `setup.py` on command prompt/terminal 
- run the command `python setup.py install`
- This would install the version as specified along with latest release of Gauge-CLI Version

## Uninstalling Gauge CLI
Gauge CLI uninstall should be done manually for now.
Run the following command on your prompt
```
$ pip uninstall gauge-cli
$ python

# This Launches the Interactive Python  

