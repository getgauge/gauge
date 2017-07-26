# Copyright [2014, 2015] [ThoughtWorks Inc.](www.thoughtworks.com)
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

$env:GOPATH=$pwd
$GOBIN="$env:GOPATH\bin"
Set-Location $GOPATH\src\github.com\getgauge\gauge
go get github.com/tools/godep && $GOBIN/godep restore


Push-Location "$pwd\bin\windows_amd64"
signtool sign /f $env:CERT_FILE /p "$env:CERT_FILE_PWD" gauge.exe
signtool sign /f $env:CERT_FILE /p "$env:CERT_FILE_PWD" gauge_screenshot.exe
Pop-Location

Push-Location "$pwd\bin\windows_386"
signtool sign /f $env:CERT_FILE /p "$env:CERT_FILE_PWD" gauge.exe
signtool sign /f $env:CERT_FILE /p "$env:CERT_FILE_PWD" gauge_screenshot.exe
Pop-Location

go run build/make.go --distro --certFile $env:CERT_FILE --certFilePwd "$env:CERT_FILE_PWD" --bin-dir bin\windows_amd64
go run build/make.go --distro --certFile $env:CERT_FILE --certFilePwd "$env:CERT_FILE_PWD" --bin-dir bin\windows_386

mkdir test_installers 

copy /B deploy\gauge-*-darwin.x86_64.zip test_installers\gauge-darwin.x86_64.zip 
copy /B deploy\gauge-*-linux.x86_64.zip test_installers\gauge-linux.x86_64.zip 
copy /B deploy\gauge-*-windows.x86_64.exe test_installers\gauge-windows.x86_64.exe