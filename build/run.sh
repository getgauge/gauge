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

if [[ -z $GOPATH ]]; then
    export GOPATH=`pwd`
fi
if [[ -z $GOBIN ]]; then
    export GOBIN="$GOPATH/bin"
fi

option="${1}"
case ${option} in
    test)
        go test ./... -v
        ;;
    build|"")
        go run build/make.go
        ;;
    *)
        echo "`basename ${0}`:usage: [build|test]"
        exit 1
        ;;
esac
