@echo off
setlocal ENABLEEXTENSIONS

REM --------------------------------------
REM Configuración de rutas
REM --------------------------------------
set PROJECT_DIR=%~dp0\..
set BIN=%PROJECT_DIR%\openpages_exporter.exe
set IN_DIR=%PROJECT_DIR%\data\in
set OUT_DIR=%PROJECT_DIR%\data\out

REM --------------------------------------
REM Verifica si el ejecutable existe
REM --------------------------------------
if not exist "%BIN%" (
    echo ❌ No se encontró %BIN%. Por favor compila con:
    echo     cd ^"%PROJECT_DIR%^"
    echo     go build -o openpages_exporter.exe ./cmd/exporter
    pause
    exit /b 1
)

REM --------------------------------------
REM Seleccionar archivo de entrada (primer JSON en in/)
REM --------------------------------------
for %%f in ("%IN_DIR%\*.json") do (
    set INPUT_FILE=%%f
    goto FOUND
)
echo ❌ No se encontró ningún archivo JSON en "%IN_DIR%"
pause
exit /b 1

:FOUND
REM --------------------------------------
REM Construir ruta de salida
REM --------------------------------------
set OUTPUT_FILE=%OUT_DIR%\tiddlers.jsonl

echo.
echo 🚀 Ejecutando conversión con:
echo     Modo:     v3
echo     Formato:  JSONL compacto (una línea por objeto)
echo     Entrada:  %INPUT_FILE%
echo     Salida:   %OUTPUT_FILE%
echo.

"%BIN%" -input "%INPUT_FILE%" -output "%OUTPUT_FILE%" -mode v3

echo.
echo ✅ Conversión completada! Salida en: %OUTPUT_FILE%
pause
