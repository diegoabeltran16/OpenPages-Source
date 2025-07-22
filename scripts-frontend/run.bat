@echo off
chcp 65001 >nul
setlocal ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION

REM --------------------------------------
REM Configuracion de rutas
REM --------------------------------------
set PROJECT_DIR=%~dp0\..
set BIN=%PROJECT_DIR%\openpages_exporter.exe
set IN_DIR=%PROJECT_DIR%\data\in
set OUT_DIR=%PROJECT_DIR%\data\out

REM --------------------------------------
REM Verifica si el ejecutable existe
REM --------------------------------------
if not exist "%BIN%" (
    echo No se encontro %BIN%. Por favor compila con:
    echo     cd "%PROJECT_DIR%"
    echo     go build -o openpages_exporter.exe ./cmd/exporter
    pause
    exit /b 1
)

REM --------------------------------------
REM Seleccionar archivo de entrada
REM --------------------------------------
set "INPUT_FILE="
for %%f in ("%IN_DIR%\*.json") do (
    if not defined INPUT_FILE (
        set "INPUT_FILE=%%f"
    )
)
if not defined INPUT_FILE (
    echo No se encontro ningun archivo JSON en "%IN_DIR%"
    pause
    exit /b 1
)

echo.
echo 1. Exportar TiddlyWiki a JSONL (modo v3)
echo 2. Revertir JSONL enriquecido a JSON TiddlyWiki
set /p CHOICE="Selecciona una opcion [1-2]: "

if "%CHOICE%"=="1" (
    set OUTPUT_FILE=%OUT_DIR%\tiddlers.jsonl
    echo.
    echo Ejecutando conversion con:
    echo     Modo:     v3
    echo     Formato:  JSONL compacto
    echo     Entrada:  !INPUT_FILE!
    echo     Salida:   !OUTPUT_FILE!
    echo.
    "%BIN%" -input "!INPUT_FILE!" -output "!OUTPUT_FILE!" -mode v3
    echo.
    echo Conversion completada! Salida en: !OUTPUT_FILE!
    pause
    exit /b 0
)

if "%CHOICE%"=="2" (
    set INPUT_JSONL=%OUT_DIR%\tiddlers.jsonl
    set OUTPUT_JSON=%OUT_DIR%\tiddlers_revert.json

    echo Entrada: !INPUT_JSONL!
    echo Salida:  !OUTPUT_JSON!
    echo.

    "%BIN%" -reverse -input "!INPUT_JSONL!" -output "!OUTPUT_JSON!"
    if exist "!OUTPUT_JSON!" (
        echo Archivo generado: !OUTPUT_JSON!
    )
)
echo Opcion invalida.
pause
exit /b 1

echo DEBUG: INPUT_FILE = !INPUT_FILE!
