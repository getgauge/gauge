@ECHO OFF
set GAUGE_PROPERTIES_FILE=%GAUGE_ROOT%\share\gauge\gauge.properties
set TIMESTAMP_FILE=%GAUGE_ROOT%\share\gauge\timestamp.txt


IF EXIST "%TIMESTAMP_FILE%" (
  set /p OLD_TIMESTAMP=<"%TIMESTAMP_FILE%"
  FOR %%f IN ("%GAUGE_PROPERTIES_FILE%") DO SET CURRENT_TIMESTAMP=%%~tf

  IF NOT "%OLD_TIMESTAMP%" == "%CURRENT_TIMESTAMP%" (
      echo "There could be some changes in gauge.properties file. Taking a backup of it in %GAUGE_PROPERTIES_FILE%.bak..."
      del "%GAUGE_PROPERTIES_FILE%.bak"
      copy "%GAUGE_PROPERTIES_FILE%" "%GAUGE_PROPERTIES_FILE%.bak"
  )
)
