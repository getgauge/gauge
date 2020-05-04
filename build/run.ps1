# ----------------------------------------------------------------
#   Copyright (c) ThoughtWorks, Inc.
#   Licensed under the Apache License, Version 2.0
#   See LICENSE.txt in the project root for license information.
# ----------------------------------------------------------------

param (
    [Parameter(Mandatory=$true)][string]$task = "build"
)

function checkLasterror {
    if ($LastExitCode -ne 0) {
        exit $LastExitCode
    }
}

function build {
    & go run build/make.go
    checkLasterror
}

function test {
    & go test .\... -v
    checkLasterror
}

switch ($task) {
    "test" { 
        test
    }
    Default {
        build
    }
}
