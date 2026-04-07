@REM @echo off
@REM
@REM echo [1/2] Starting App1 ...
@REM start "Dapr1" cmd /k "daprd.exe run --app-id shaogd-app --app-port 8100 --dapr-http-port 3600 --dapr-grpc-port 52001 --dapr-internal-grpc-port 48444 --enable-metrics=false --log-level=debug "
@REM start "Dapr-App1" cmd /k " python -m http.server 8100 -d ./ "
@REM
@REM timeout /t 4 /nobreak >nul
@REM
@REM echo [2/2] Starting App2 ...
@REM start "Dapr2" cmd /k "daprd.exe run --app-id shaogd-app --app-port 8101 --dapr-http-port 3601 --dapr-grpc-port 52002 --dapr-internal-grpc-port 48445 --enable-metrics=false "
@REM start "Dapr-App2" cmd /k "python -m http.server 8101 -d ./"
@REM
@REM echo DONE
@REM pause
@echo off

 if "%1"=="" (
    echo Usage: start-dapr.bat [appid] [app-port]
    pause
    exit /b
 )

 if "%2"=="" (
    echo Usage: start-dapr.bat [appid] [app-port]
    pause
    exit /b
 )

 set appid=%1
 set app_port=%2
 set /a app_port2=%app_port% + 1
 set /a dapr_http_port1=%app_port% - 4500
 set /a dapr_grpc_port1=52000 + (%app_port% - 8100 + 1)
 set /a dapr_internal_grpc_port1=48444 + (%app_port% - 8100)
 set /a dapr_http_port2=%app_port2% - 4500
 set /a dapr_grpc_port2=52000 + (%app_port2% - 8100 + 1)
 set /a dapr_internal_grpc_port2=48444 + (%app_port2% - 8100)

 echo [1/2] Starting App1 ...
 start "Dapr1" cmd /k "daprd.exe run --app-id %appid% --app-port %app_port% --dapr-http-port %dapr_http_port1% --dapr-grpc-port %dapr_grpc_port1% --dapr-internal-grpc-port %dapr_internal_grpc_port1% --enable-metrics=false --log-level=debug "
 start "Dapr-App1" cmd /k " python -m http.server %app_port% -d ./ "

 timeout /t 4  /nobreak > nul

 echo [2/2] Starting App2 ...
 start "Dapr2" cmd /k "daprd.exe run --app-id %appid% --app-port %app_port2% --dapr-http-port %dapr_http_port2% --dapr-grpc-port %dapr_grpc_port2% --dapr-internal-grpc-port %dapr_internal_grpc_port2% --enable-metrics=false "
 start "Dapr-App2" cmd /k "python -m http.server %app_port2% -d ./"

 echo DONE
 pause
