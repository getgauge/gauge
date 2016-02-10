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
    echo "Installing plugin - $1..."
    $prefix/bin/gauge --install $1
}

display_usage() {
	echo "On Linux, this script takes an optional install path."
	echo -e "\nUsage:\n$0 [path] \n"
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
        echo "export GAUGE_ROOT=$prefix"  >> ~/.profile
        updated_profile=1
    fi
    if [ $updated_profile ] ; then
        source ~/.profile
    fi
}

install_gauge() {
    echo "Installing gauge at $prefix"
    echo "Creating $prefix if it doesn't exist"
    [ -d $prefix ] || mkdir $prefix

    # check for write permissions
    if [ ! -w "$prefix" -a "$prefix" = "/usr/local" ]; then
        echo
        echo "Installation failed..."
        echo "You do not have write permissions for $prefix"
        echo "Please run this script as sudo or pass a custom location where you want to install Gauge."
        echo "Example: ./install.sh /home/gauge/local/gauge_install_dir"
        echo
        exit 1
    fi

    # do the installation
    echo "Copying files to $prefix"
    gaugePropertiesFile=$prefix/share/gauge/gauge.properties
    if [ -f $prefix/share/gauge/timestamp.txt ] ; then
        currentTimeStamp=`date +%s -r $gaugePropertiesFile`
        oldTimeStamp=`cat $prefix/share/gauge/timestamp.txt`
        if [ $currentTimeStamp != $oldTimeStamp ] ; then
            backupFile=$prefix/share/gauge/gauge.properties.bak
            echo "If you have Gauge installed already and there are any manual changes in gauge.properties file, a backup of it has been taken at GAUGE_INSTALL_LOCATION\share\gauge\gauge.properties.bak. You can restore these configurations later."
            rm -rf $backupFile
            cat $gaugePropertiesFile > $backupFile
        fi
    fi

    cp -rf bin share $prefix
    date +%s -r $gaugePropertiesFile > $prefix/share/gauge/timestamp.txt

    set_gaugeroot
    echo "Gauge core successfully installed.\n"
}

if [ -z "$1" ]; then
    prefix=/usr/local
else
    prefix=$1
fi

# if more than one arguments supplied, display usage
if [ $# -gt 1 ]
then
    display_usage
    exit 1
fi

# check whether user has supplied -h or --help . If yes display usage
if [[ ( $@ == "--help") || $@ == "-h" ]]
then
    display_usage
    exit 0
fi

install_gauge $prefix
install_plugin html-report
