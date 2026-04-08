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
set /a dapr_http_port1=4000 + (%app_port% %% 100)
set /a dapr_grpc_port1=30000 + (%app_port% %% 100)
set /a dapr_internal_grpc_port1=20000 + (%app_port% %% 1000)
set /a dapr_http_port2=4100 + (%app_port2% %% 100)
set /a dapr_grpc_port2=30100 + (%app_port2% %% 100)
set /a dapr_internal_grpc_port2=20100 + (%app_port2% %% 1000)

 echo [1/2] Starting App1 ...
 start "Dapr1" cmd /k "daprd.exe run --app-id %appid% --app-port %app_port% --dapr-http-port %dapr_http_port1% --dapr-grpc-port %dapr_grpc_port1% --dapr-internal-grpc-port %dapr_internal_grpc_port1% --enable-metrics=false "
 start "Dapr-App1" cmd /k " python -m http.server %app_port% -d ./ "

 timeout /t 4  /nobreak > nul

 echo [2/2] Starting App2 ...
 start "Dapr2" cmd /k "daprd.exe run --app-id %appid% --app-port %app_port2% --dapr-http-port %dapr_http_port2% --dapr-grpc-port %dapr_grpc_port2% --dapr-internal-grpc-port %dapr_internal_grpc_port2% --enable-metrics=false "
 start "Dapr-App2" cmd /k "python -m http.server %app_port2% -d ./"

 echo DONE
 pause
