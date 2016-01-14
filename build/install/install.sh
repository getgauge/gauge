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

#!/bin/bash

set -e

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
    cp -rf bin share $prefix

    # ensure gauge is on path
    if [ -z "$(which gauge)" ]; then
        echo "Adding gauge to system path..."
        echo "PATH=$PATH:$prefix/bin" >> ~/.profile
        echo "export GAUGE_ROOT=$prefix"  >> ~/.profile
        source ~/.profile
    fi
    echo "Gauge core successfully installed."
    echo
}

install_plugin() {
    echo "Installing plugin - $1..."
    $prefix/bin/gauge --install $1
}

if [ -z "$1" ]; then
    prefix=/usr/local
else
    prefix=$1
fi

install_gauge $prefix
install_plugin html-report

