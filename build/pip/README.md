# PIP Installer for getgauge-cli tool

This `setup.py` file is used for installing the getgauge-cli tool. The Python package can be installed multiple ways:

## using pip

- pre-requisite: pip should be installed and available on machine
- run the command `pip install getgauge-cli`

## using setup.py

- Navigate to the current directory `gauge\build\pip\`
- Run `python build.py --install` on command prompt (it will install the latest released version)
- To install a specific version set `GAUGE_VERSION` environment variable.

## Check to ensure getgauge-cli is installed

- Once the package has been setup. please exit the current terminal and relaunch terminal again
- run the command `gauge` , you should be able to see similar output
```
Usage:
  gauge <command> [flags] [args]

# ...etc
```

## Using setup.py
- Install required dep by running `python/python3 -m pip install requirements.txt`.
- Check through and obtain a valid tag/build number from [releases](https://github.com/getgauge/gauge/releases)
- Run the command `python build.py --install` (set GAUGE_VERSION env to install specific version)
- This would install the version as specified along with latest release of Gauge-CLI Version

- Run the command `python build.py --dist` to generate the PyPi distributable (set GAUGE_VERSION env to install specific version)

## Uninstalling Gauge CLI

Gauge CLI uninstall should be done manually for now.
Run the following command on your prompt
```
$ pip uninstall gauge-cli
```
