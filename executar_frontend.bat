@echo off
setlocal
echo ==============================================
echo Iniciando o Frontend Kaptei (React/Vite)...
echo ==============================================

cd /d "%~dp0frontend"
if errorlevel 1 (
    echo ERRO: Nao foi possivel acessar a pasta frontend.
    goto :falha
)

where node >nul 2>&1
if errorlevel 1 (
    echo ERRO: Node.js nao foi encontrado no PATH.
    goto :falha
)

where npm.cmd >nul 2>&1
if errorlevel 1 (
    echo ERRO: npm nao foi encontrado no PATH.
    goto :falha
)

echo Verificando dependencias do Node...
if not exist "node_modules" (
    echo Instalando dependencias bloqueadas pelo package-lock.json...
    call npm.cmd ci --no-audit --no-fund
    if errorlevel 1 (
        echo ERRO: Nao foi possivel instalar as dependencias.
        goto :falha
    )
)

echo.
echo Iniciando o servidor de desenvolvimento...
call npm.cmd run dev
set "KAPTEI_CODIGO=%ERRORLEVEL%"
if not "%KAPTEI_CODIGO%"=="0" goto :falha_codigo
exit /b 0

:falha
set "KAPTEI_CODIGO=1"

:falha_codigo
echo.
pause
exit /b %KAPTEI_CODIGO%
