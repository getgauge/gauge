#!/bin/sh

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

# Usage:
# ./build/package.sh [--nightly]

if [[ -z $GOPATH ]]; then
    export GOPATH=`pwd`
fi
if [[ -z $GOBIN ]]; then
    export GOBIN="$GOPATH/bin"
fi

cd $GOPATH/src/github.com/getgauge/gauge

go run build/make.go --all-platforms $1

chmod +x bin/**/* && rm -rf deploy

go run build/make.go --distro --all-platforms --skip-windows $1
