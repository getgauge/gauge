# Gauge packaging for Windows 

## Tools required

* Nullsoft Install System (http://nsis.sourceforge.net)
* NIS Edit (Optional, just for authoring the script files)

NSIS binaries should be added to path 

## Variables

* `PRODUCT_VERSION`: -> Version of Gauge
* `GAUGE_DISTRIBUTABLES_DIR`: Directory where gauge distributables are available. It should not end with `\`
*  `OUTPUT_FILE_NAME`: Name of the setup file

## Sample command

```
makensis.exe /DPRODUCT_VERSION=0.0.2 /DGAUGE_DISTRIBUTABLES_DIR=c:\gauge /DOUTPUT_DILE_NAME=gauge-0.0.2-windows-x86_64.exe  "gauge-install.nsi" 
```

This will generate a file called `gauge-0.0.2-windows-x86_64.exe` in the current path

