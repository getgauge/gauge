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

install_plugin() {
    oldIFS="$IFS"
    IFS=","
    IFS=${IFS:0:1} # this is useful to format your code with tabs
    pluginsList=( $plugins )
    IFS="$oldIFS"

    for plugin in "${pluginsList[@]}"
    do
        echo "Installing plugin $plugin ..."
        $prefix/bin/gauge --install $plugin
    done
}

display_usage() {
	echo -e "On Linux, this script installs gauge and it's plugins.\n\nUsage:\n./install.sh\n\nSet PREFIX env to install gauge at custom location.
Set PLUGINS env to install plugins alogn with gauge.
Exp:-
    PREFIX=my/custom/path ./install.sh
    PLUGINS=java,ruby,spectacle ./install.sh
    PREFIX=my/custom/path PLUGINS=xml-report,java ./install.sh"
}

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
        echo "export GAUGE_ROOT=$config"  >> ~/.profile
        updated_profile=1
    fi
    if [ $updated_profile ] ; then
        source ~/.profile
    fi
}

create_prefix_if_doesnt_exist() {
    if [ "$prefix" == "$config" ]; then 
        echo "Creating $prefix if it doesn't exist"
        [ -d $prefix ] || mkdir $prefix
    else
        echo "Creating $prefix if it doesn't exist"
        [ -d $prefix ] || mkdir $prefix
        echo "Creating $config if it doesn't exist"
        [ -d $config ] || mkdir -p $config
    fi
}

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
    gaugePropertiesFile=$config/share/gauge/gauge.properties
    if [ -f $config/share/gauge/timestamp.txt ] ; then
        currentTimeStamp=`date +%s -r $gaugePropertiesFile`
        oldTimeStamp=`cat $config/share/gauge/timestamp.txt`
        if [ $currentTimeStamp != $oldTimeStamp ] ; then
            backupFile=$config/share/gauge/gauge.properties.bak
            echo "If you have Gauge installed already and there are any manual changes in gauge.properties file, a backup of it has been taken at GAUGE_INSTALL_LOCATION\share\gauge\gauge.properties.bak. You can restore these configurations later."
            rm -rf $backupFile
            cat $gaugePropertiesFile > $backupFile
        fi
    fi
    cp -rf share $config
    date +%s -r $gaugePropertiesFile > $config/share/gauge/timestamp.txt
}

install_gauge() {
    config="$HOME/.gauge/config"
    if [ "$prefix" != "/usr/local" ]; then
        config=$prefix
    fi
    echo "Installing gauge at $prefix/bin"
    if tty -s; then
        echo -e "Provide a custom location or press ENTER :-"
        read -e installLocatioan
        prefix=$installLocatioan
        config=$installLocatioan
    fi
    create_prefix_if_doesnt_exist
    copy_gauge_binaries
    copy_gauge_configuration_files
    set_gaugeroot
    echo "Gauge core successfully installed.\n"
}

if [ -z "$PREFIX" ]; then
    prefix=/usr/local
else
    prefix=$PREFIX
fi

if [ -z "$PLUGINS" ]; then
    plugins=html-report
else
    plugins=$PLUGINS
fi

# check whether user has supplied -h or --help . If yes display usage

if [[$# != 0 ]]; then
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
