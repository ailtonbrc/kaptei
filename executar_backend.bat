@echo off
setlocal
echo ==============================================
echo Iniciando o Backend Kaptei (Go)...
echo ==============================================

cd /d "%~dp0"
if errorlevel 1 (
    echo ERRO: Nao foi possivel acessar a pasta do Kaptei.
    goto :falha
)

where go >nul 2>&1
if errorlevel 1 (
    echo ERRO: Go nao foi encontrado no PATH.
    goto :falha
)

where powershell.exe >nul 2>&1
if errorlevel 1 (
    echo ERRO: Windows PowerShell nao foi encontrado.
    goto :falha
)

if not exist "backend\.env" (
    echo ERRO: backend\.env nao existe.
    echo Crie o arquivo a partir de backend\.env.example sem versionar segredos.
    goto :falha
)

echo.
echo Preparando configuracao segura do ambiente local...
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\preparar_ambiente_local.ps1"
if errorlevel 1 (
    echo ERRO: A preparacao segura do ambiente falhou.
    goto :falha
)

echo.
echo Verificando backup e migrations pendentes...
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\migrar_banco_seguro.ps1"
if errorlevel 1 (
    echo ERRO: Backup ou migrations falharam. A API nao sera iniciada.
    goto :falha
)

cd /d "%~dp0backend"
if errorlevel 1 goto :falha

echo.
echo Compilando a API...
if not exist "bin" mkdir "bin"
go build -o "bin\api.exe" ./cmd/api
if errorlevel 1 (
    echo ERRO: A compilacao falhou. O binario anterior nao sera executado.
    goto :falha
)

echo.
echo API pronta para iniciar em http://localhost:8080
"bin\api.exe"
set "KAPTEI_CODIGO=%ERRORLEVEL%"
if not "%KAPTEI_CODIGO%"=="0" (
    echo ERRO: A API foi encerrada com codigo %KAPTEI_CODIGO%.
    goto :falha_codigo
)
exit /b 0

:falha
set "KAPTEI_CODIGO=1"

:falha_codigo
echo.
pause
exit /b %KAPTEI_CODIGO%
