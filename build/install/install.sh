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
	path_prefix=$1
    echo "creating $prefix if it doesn't exist"
    [ -d $prefix ] || mkdir $prefix
    echo "Copying files to $prefix"
    cp -rf bin share $prefix
    
    # ensure gauge is on path
    if ! type "gauge" > /dev/null; then
        export GAUGE_ROOT=$prefix
        
        echo "=========================================="
        echo "Please add $prefix/bin to your PATH env variable."
        echo "=========================================="
    else
        echo "$prefix is in path: $PATH"
    fi
    echo "Gauge core successfully installed."
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
