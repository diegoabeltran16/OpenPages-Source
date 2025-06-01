@echo off
REM ------------------------------------------------------------------------------
REM scripts-frontend\run.bat – Script interactivo para ejecutar el pipeline en Windows
REM ------------------------------------------------------------------------------

setlocal EnableDelayedExpansion

REM ------------------------------------------- Configuración base
set PROJECT_DIR=%~dp0..
set IN_DIR=%PROJECT_DIR%\data\in
set OUT_DIR=%PROJECT_DIR%\data\out

REM Buscar primer archivo JSON en data\in
for %%F in (%IN_DIR%\*.json) do (
    set "DEFAULT_IN_FILE=%%F"
    goto :found_json
)
:found_json

set "DEFAULT_OUT_FILE=%OUT_DIR%\out.jsonl"

echo 📁 Carpeta del proyecto: %PROJECT_DIR%
echo 📥 Archivo de entrada sugerido: %DEFAULT_IN_FILE%
echo.

REM ------------------------------------------- Selección de modo
:mode_select
echo 🧠 ¿Qué modo deseas usar para transformar los tiddlers?
echo   1) v1 (básico)
echo   2) v2 (estructurado con meta y content)
echo   3) Cancelar
set /p MODE_CHOICE=Opción [1-3]: 

if "%MODE_CHOICE%"=="1" (
    set MODE_FLAG=v1
) else if "%MODE_CHOICE%"=="2" (
    set MODE_FLAG=v2
) else if "%MODE_CHOICE%"=="3" (
    echo 🚪 Cancelado.
    exit /b
) else (
    echo ❌ Opción inválida. Intenta de nuevo.
    goto :mode_select
)

REM ------------------------------------------- Selección de formato
:format_select
echo.
echo 📤 ¿Qué formato de salida deseas?
echo   1) JSONL (una línea por entrada, ideal para IA)
echo   2) JSON indentado (solo para inspección humana)
echo   3) Cancelar
set /p FORMAT_CHOICE=Opción [1-3]: 

if "%FORMAT_CHOICE%"=="1" (
    set PRETTY_FLAG=
) else if "%FORMAT_CHOICE%"=="2" (
    set PRETTY_FLAG=-pretty
) else if "%FORMAT_CHOICE%"=="3" (
    echo 🚪 Cancelado.
    exit /b
) else (
    echo ❌ Opción inválida. Intenta de nuevo.
    goto :format_select
)

REM ------------------------------------------- Confirmación entrada
echo.
echo 🔍 Archivo de entrada detectado: %DEFAULT_IN_FILE%
set /p CONFIRM_IN=¿Deseas usar esta ruta? [S/n]: 
if /I "%CONFIRM_IN%"=="n" (
    set /p INPUT_FILE=Escribe la nueva ruta completa del archivo de entrada: 
) else (
    set INPUT_FILE=%DEFAULT_IN_FILE%
)

REM ------------------------------------------- Confirmación salida
echo.
echo 💾 Archivo de salida sugerido: %DEFAULT_OUT_FILE%
set /p CONFIRM_OUT=¿Deseas usar esta ruta? [S/n]: 
if /I "%CONFIRM_OUT%"=="n" (
    set /p OUTPUT_FILE=Escribe la nueva ruta completa del archivo de salida: 
) else (
    set OUTPUT_FILE=%DEFAULT_OUT_FILE%
)

REM ------------------------------------------- Ejecución
echo.
echo 🚀 Ejecutando exportación...
echo     Modo:    %MODE_FLAG%
if defined PRETTY_FLAG (
    echo     Formato: JSON indentado
) else (
    echo     Formato: JSONL plano
)
echo     Entrada: %INPUT_FILE%
echo     Salida:  %OUTPUT_FILE%
echo.

REM Ejecutar el binario compilado (si no está compilado usar: go run ...)
openpages_exporter.exe -input "%INPUT_FILE%" -output "%OUTPUT_FILE%" -mode %MODE_FLAG% %PRETTY_FLAG%

echo.
echo ✅ Exportación completada.
pause
