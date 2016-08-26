$zipsrc = $args[0]

$zipdst = $args[1]

If(Test-path $zipdst) {Remove-item $zipdst}

Add-Type -assembly "system.io.compression.filesystem"

[io.compression.zipfile]::CreateFromDirectory($zipsrc, $zipdst)
