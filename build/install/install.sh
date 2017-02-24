#!/bin/bash

# Copyright 2015 ThoughtWorks, Inc.

# This file is part of Gauge.

# Gauge is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.

# Gauge is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with Gauge.  If not, see <http://www.gnu.org/licenses/>.
set -e


# Install all the plugins mentioned in $PLUGINS
install_plugin() {
    oldIFS="$IFS"
    IFS=","
    IFS=${IFS:0:1} # this is useful to format your code with tabs
    pluginsList=( $plugins )
    IFS="$oldIFS"

    for plugin in "${pluginsList[@]}"
    do
        echo "Installing plugin $plugin ..."
        gauge --install $plugin
    done
}

# Print usage of this script
display_usage() {
	echo -e "On Linux, this script installs gauge and it's plugins.\n\nUsage:\n./install.sh\n\nSet PREFIX env to install gauge at custom location.
Set PLUGINS env to install plugins alogn with gauge.
Exp:-
    PREFIX=my/custom/path ./install.sh
    PLUGINS=java,ruby,spectacle ./install.sh
    PREFIX=my/custom/path PLUGINS=xml-report,java ./install.sh"
}

# Find absolute path
get_absolute_path (){
    [[ -d $1 ]] && { cd "$1"; echo "$(pwd -P)"; } ||
    { cd "$(dirname "$1")" || exit 1; echo "$(pwd -P)/$(basename "$1")"; }
}

# Set GAUGE_ROOT environment variable
set_gaugeroot() {
    # ensure gauge is on PATH
    if [ -z "$(which gauge)" ]; then
        echo "Adding gauge to system path..."
        echo "PATH=$PATH:$prefix/bin" >> ~/.profile
        updated_profile=1
    fi
    # ensure GAUGE_ROOT is set
    if [ -z "$GAUGE_ROOT" ]; then
        echo "Adding GAUGE_ROOT to environment..."
        echo "export GAUGE_ROOT=$configPrefix"  >> ~/.profile
        updated_profile=1
    fi
    if [ $updated_profile ] ; then
        source ~/.profile
    fi
    echo -e "GAUGE_ROOT has been set. If you face errors, run '$ source ~/.profile'\n"
}


# Creates installation prefix and configuration dirs if doesn't exist
create_prefix_if_does_not_exist() {
      [ -d $prefix ] || echo "Creating $prefix ..." && mkdir -p $prefix
      [ -d $config ] || echo "Creating $config ..." && mkdir -p $config
}


# Copy gauge binaries in $prefix dir
copy_gauge_binaries() {
    # check for write permissions and Install gauge, asks for sudo access if not permitted
    if [ ! -w "$prefix" -a "$prefix" = "/usr/local" ]; then
        echo
        echo "You do not have write permissions for $prefix"
        echo "Running script as sudo "
        sudo cp -rf bin $prefix
        echo "Installed gauge binaries at $prefix"
        sudo -k
    else
        cp -rf bin $prefix
        echo "Installed gauge binaries at $prefix"
    fi
}

# copy gauge configuration at $config
copy_gauge_configuration_files() {
    gaugePropertiesFile=$config/gauge.properties
    if [ -f $config/timestamp.txt ] ; then
        currentTimeStamp=`date +%s -r $gaugePropertiesFile`
        oldTimeStamp=`cat $config/timestamp.txt`
        if [ $currentTimeStamp != $oldTimeStamp ] ; then
            backupFile=$config/gauge.properties.bak
            echo "If you have Gauge installed already and there are any manual changes in gauge.properties file, a backup of it has been taken at GAUGE_INSTALL_LOCATION\share\gauge\gauge.properties.bak. You can restore these configurations later."
            rm -rf $backupFile
            cat $gaugePropertiesFile > $backupFile
        fi
    fi
    cp -rf share/gauge/* $config
    date +%s -r $gaugePropertiesFile > $config/timestamp.txt
}

# Do the installation
install_gauge() {
    configPrefix="$HOME/.gauge"
    if [ "$prefix" != "/usr/local" ]; then
        configPrefix=$prefix
    fi
    echo "Installing gauge at $prefix/bin"
    if tty -s; then
        echo -e "Press [ENTER] to continue or provide a custom location to install gauge at that location:-"
        read -e installLocation
        if [[ ! -z $installLocation ]]; then
          prefix=$(get_absolute_path ${installLocation/\~/$HOME})
          configPrefix=$prefix
        fi
    fi

    config=$configPrefix/config
    create_prefix_if_does_not_exist
    copy_gauge_binaries
    copy_gauge_configuration_files
    set_gaugeroot
    source ~/.profile
    echo "Gauge core successfully installed.\n"
}


# Set install location to /usr/local/bin if $PREFIX is not set.
if [ -z "$PREFIX" ]; then
    prefix=/usr/local
else
    prefix=$PREFIX
fi


# Set html-report as default plugin in plugin list if $PLUGINS is not set.
if [ -z "$PLUGINS" ]; then
    plugins=html-report
else
    plugins=$PLUGINS
fi

# check whether user has supplied -h or --help . If yes display usage if no diplay usage with an error
if [[ $# != 0 ]]; then
    if [[ ( $@ == "--help") || $@ == "-h" ]]
    then
        display_usage
        exit 0
    else
        echo -e "unknown option $@. \n"
        display_usage
        exit 1
    fi
fi

install_gauge $prefix
install_plugin $plugins
