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

export GOPATH=`pwd`
export GOBIN="$GOPATH/bin"

cd $GOPATH/src/github.com/getgauge/gauge

go get github.com/tools/godep && $GOBIN/godep restore

go run build/make.go --all-platforms

chmod +x build/install/install.sh && chmod +x bin/**/* && rm -rf deploy

security unlock-keychain -p $KEYCHAIN_PASSWORD login.keychain && security import /vagrant/Gauge_Osx_Cert.p12 -P "$CERT_PASSWORD" -A -k login.keychain && go run build/make.go --distro --all-platforms --skip-windows