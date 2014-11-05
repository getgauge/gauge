mkdir out\%major_version%.%minor_version%.%patch_version%
makensis.exe "/DPRODUCT_VERSION=%major_version%.%minor_version%.%patch_version%" "/DGAUGE_DISTRIBUTABLES_DIR=%gauge_distributables_dir%" "/DOUTPUT_FILE_NAME=out\%major_version%.%minor_version%.%patch_version%\gauge-%major_version%.%minor_version%.%patch_version%-%os%.%arch%.exe"  gauge-install.nsi
