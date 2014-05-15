
@ECHO OFF
setlocal

set programName="gauge"
if "%1" == "install" (	
	md "%APPDATA%\%programName%\share\%programName%\skel\env"
	copy %programName%.exe "%APPDATA%\%programName%"
	copy skel\hello_world.spec "%APPDATA%\%programName%\share\%programName%\skel"
	copy skel\default.properties "%APPDATA%\%programName%\share\%programName%\skel\env"
	echo Installation successful
) else if "%1" == "test" (
	SET GOPATH="%cd%"
	cd src
	go test
	cd ..
) else (
	SET GOPATH="%cd%"
	cd src
	go build
	copy src.exe ..\%programName%.exe
	cd ..
)


endlocal
echo.
echo Build complete
