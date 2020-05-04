# ----------------------------------------------------------------
#   Copyright (c) ThoughtWorks, Inc.
#   Licensed under the Apache License, Version 2.0
#   See LICENSE.txt in the project root for license information.
# ----------------------------------------------------------------

param (
    [switch]$nightly = $false
)

$nightlyFlag = If ($nightly) {"--nightly"} Else {""}
& go run build/make.go --distro --certFile $env:CERT_FILE --bin-dir bin\windows_amd64 $nightlyFlag
& go run build/make.go --distro --certFile $env:CERT_FILE --bin-dir bin\windows_386 $nightlyFlag
