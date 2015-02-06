# Copyright 2014 ThoughtWorks, Inc.

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
if [ -z "$prefix" ];
then
	prefix=/usr/local
  echo "creating $prefix if it doesn't exist"
  [ -d $prefix ] || mkdir $prefix
  echo "Copying files to $prefix"
  sudo /bin/sh -c "cp -rf bin share $prefix"
else
  echo "creating $prefix if it doesn't exist"
  [ -d $prefix ] || mkdir $prefix
  echo "Copying files to $prefix"
  cp -rf bin share $prefix
fi

echo "Done installing Gauge to $prefix"
echo "Installing plugin - html-report. This may take a few minutes"
$prefix/bin/gauge --install html-report
