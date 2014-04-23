
@ECHO OFF
setlocal


if "%1" == "install" (	
	md "%APPDATA%\gauge\share\twist2\skel\env"
	copy twist2.exe "%APPDATA%\gauge"
	copy skel\hello_world.spec "%APPDATA%\gauge\share\twist2\skel"
	copy skel\default.properties "%APPDATA%\gauge\share\twist2\skel\env"
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
	copy src.exe ..\twist2.exe
	cd ..
)


endlocal
echo.
echo Build complete
