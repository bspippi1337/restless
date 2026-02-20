@echo off
setlocal
cd /d "%~dp0\.."
powershell -NoProfile -ExecutionPolicy Bypass -File "scripts\build-winx64.ps1" %*
echo.
echo Done. Press any key...
pause >nul
