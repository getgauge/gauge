# ----------------------------------------------------------------
#   Copyright (c) ThoughtWorks, Inc.
#   Licensed under the Apache License, Version 2.0
#   See LICENSE.txt in the project root for license information.
# ----------------------------------------------------------------

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
