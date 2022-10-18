@ECHO OFF
MKDIR "%WINDIR%\..\tools"
@REM Copy ".\QuickGo.exe" exe to C:\Tools
COPY ".\QuickGo.exe" "%WINDIR%\..\tools\QuickGo.exe"
@REM Copy ".\conf\*" to C:\Tools
XCOPY ".\conf" "%WINDIR%\..\tools\conf" /E /I /Y
@REM Set environment variable for the tools folder
SETX PATH "%PATH%;%WINDIR%\..\tools"