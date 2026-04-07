@echo off

echo ========================================
echo   Stop All Dapr Instances
echo ========================================
echo.

taskkill /F /IM daprd.exe >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] daprd.exe killed
) else (
    echo [INFO] No daprd.exe found running
)

echo.
echo To stop Python HTTP servers, close their windows manually.
echo.
pause
