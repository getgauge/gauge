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

$zipsrc = $args[0]
$zipdst = $args[1]

If(Test-path $zipdst) {Remove-item $zipdst}

Add-Type -assembly "system.io.compression.filesystem"

Write-Host "Creating zip : $zipdst"

try {
    [io.compression.zipfile]::CreateFromDirectory($zipsrc, $zipdst)
}
catch {
    Write-Host -ForegroundColor Red "Unable to create zip : $($_.Exception)"
    exit -1
}
