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

param (
    [switch]$nightly = $false
)

if ("$env:GOPATH" -eq "") {
    $env:GOPATH=$pwd
}
if ("$env:GOBIN" -eq "") {
    $env:GOBIN="$env:GOPATH\bin"
}

Set-Location -Path "$env:GOPATH\src\github.com\getgauge\gauge"

Push-Location "$pwd\bin\windows_amd64"
signtool sign /tr http://timestamp.digicert.com /td sha256 /fd sha256 /f $env:CERT_FILE /p "$env:CERT_FILE_PWD" gauge.exe
if ($LastExitCode -ne 0) {
     throw "gauge.exe signing failed"
}
Pop-Location

Push-Location "$pwd\bin\windows_386"
signtool sign /tr http://timestamp.digicert.com /td sha256 /fd sha256 /f $env:CERT_FILE /p "$env:CERT_FILE_PWD" gauge.exe
if ($LastExitCode -ne 0) {
     throw "gauge.exe signing failed"
}
Pop-Location

$nightlyFlag = If ($nightly) {"--nightly"} Else {""}
& go run build/make.go --distro --certFile $env:CERT_FILE --certFilePwd "$env:CERT_FILE_PWD" --bin-dir bin\windows_amd64 $nightlyFlag
& go run build/make.go --distro --certFile $env:CERT_FILE --certFilePwd "$env:CERT_FILE_PWD" --bin-dir bin\windows_386 $nightlyFlag

mkdir test_installers 

& cmd "/c" "copy /B deploy\gauge-*-darwin.x86_64.zip test_installers\gauge-darwin.x86_64.zip"
& cmd "/c" "copy /B deploy\gauge-*-linux.x86_64.zip test_installers\gauge-linux.x86_64.zip"
& cmd "/c" "copy /B deploy\gauge-*-windows.x86_64.zip test_installers\gauge-windows.x86_64.zip"