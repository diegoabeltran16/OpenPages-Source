@echo off
chcp 65001 >nul
setlocal ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION
cd /d "%~dp0.."

REM === OpenPages-Source Frontend Script ===
REM Script interactivo para exportar y revertir tiddlers

set EXPORTER_BIN=.\openpages_exporter.exe
set REVERT_BIN=.\openpages_revert.exe

REM Compilar los binarios si no existen
if not exist %EXPORTER_BIN% (
    echo Compilando exporter...
    go build -o %EXPORTER_BIN% cmd\exporter
)
if not exist %REVERT_BIN% (
    echo Compilando revert...
    go build -o %REVERT_BIN% cmd\revert
)

:MENU
echo.
echo ==========================================
echo   OpenPages-Source - Menu de acciones
echo ==========================================
echo [1] Exportar tiddlers desde plantilla
echo [2] Revertir y actualizar textos desde JSONL
echo [3] Ejecutar ambos procesos (pipeline completo)
echo [0] Salir
echo.
set /p choice=Elige una opcion [0-3]:

if "%choice%"=="1" goto EXPORT
if "%choice%"=="2" goto REVERT
if "%choice%"=="3" goto PIPELINE
if "%choice%"=="0" goto END

echo Opcion invalida. Intenta de nuevo.
goto MENU

:EXPORT
echo.
REM Listar archivos en data\in para selección
set /a idx=0
echo Archivos disponibles en data\in:
for %%f in ("data\in\*.json") do (
    set /a idx+=1
    set "file_in[!idx!]=%%f"
    echo   [!idx!] %%f
)
if "%idx%"=="0" (
    echo No se encontraron archivos en data\in.
    pause
    goto MENU
)
set /p sel_in=Selecciona el archivo de entrada por número:
if not defined file_in[%sel_in%] (
    echo Selección inválida.
    pause
    goto MENU
)
set input_local=!file_in[%sel_in%]!
for %%A in ("%input_local%") do (
    set "output_local=data\out\%%~nA.jsonl"
)
set "input=%input_local%"
set "output=data\out"

echo Modos de exportacion disponibles:
echo   [1] v1
echo   [2] v2
echo   [3] v3
echo   [4] hybrid
set /p sel_mode=Selecciona el modo por número [default: 3]:
if "%sel_mode%"=="" set sel_mode=3

if "%sel_mode%"=="1" set mode=v1
if "%sel_mode%"=="2" set mode=v2
if "%sel_mode%"=="3" set mode=v3
if "%sel_mode%"=="4" set mode=hybrid

REM Validar selección
if not "%mode%"=="v1" if not "%mode%"=="v2" if not "%mode%"=="v3" if not "%mode%"=="hybrid" (
    echo Selección de modo inválida.
    pause
    goto MENU
)

set /p pretty=¿Salida pretty? (y/n) [default: n]:
if /i "%pretty%"=="y" (
    set pretty=-pretty
) else (
    set pretty=
)
echo.
echo Parámetros seleccionados:
echo   Entrada: %input%
echo   Salida: %output%
echo   Modo: %mode%
echo   Pretty: %pretty%
echo.
echo Ejecutando exportacion...
%EXPORTER_BIN% -input "%input%" -output "%output%" -mode %mode% %pretty%
if errorlevel 1 (
    echo Error en exportacion. Abortando.
    pause
    goto MENU
)
echo Exportacion completada.
pause
goto MENU

:REVERT
echo.
REM Selección de archivo plantilla JSON en data\in
set /a idx=0
echo Archivos disponibles en data\in:
for %%f in ("data\in\*.json") do (
    set /a idx+=1
    set "file_in[!idx!]=%%f"
    echo   [!idx!] %%f
)
if "%idx%"=="0" (
    echo No se encontraron archivos en data\in.
    pause
    goto MENU
)
set /p sel_in=Selecciona el archivo plantilla por número:
if not defined file_in[%sel_in%] (
    echo Selección inválida.
    pause
    goto MENU
)
set plantilla=!file_in[%sel_in%]!

REM Selección de archivo JSONL en data\out
set /a idx=0
echo Archivos disponibles en data\out:
for %%f in ("data\out\*.jsonl") do (
    set /a idx+=1
    set "file_out[!idx!]=%%f"
    echo   [!idx!] %%f
)
if "%idx%"=="0" (
    echo No se encontraron archivos en data\out.
    pause
    goto MENU
)
set /p sel_out=Selecciona el archivo JSONL por número:
if not defined file_out[%sel_out%] (
    echo Selección inválida.
    pause
    goto MENU
)
set input_local=!file_out[%sel_out%]!
for %%A in ("%input_local%") do (
    set "output_local=data\in\%%~nA (reverted).json"
)
set input=%input_local%
set output=%output_local%

echo.
echo Parámetros seleccionados:
echo   Plantilla: %plantilla%
echo   Textos: %input%
echo   Salida: %output%
echo.
REM Asegurar que la carpeta de destino existe
if not exist "data\in" (
    mkdir "data\in"
)
echo Ejecutando revertido y actualizacion de textos...
%REVERT_BIN% "%plantilla%" "%input%" "%output%"
if errorlevel 1 (
    echo Error en revertido. Abortando.
    pause
    goto MENU
)
echo Revertido completado.
pause
goto MENU

:PIPELINE
echo Ejecutando pipeline completo...
set input_plantilla=data\in\Plantilla (Estudiar OpenPages).json
set output_dir=data\out
set output_revert=data\in\Plantilla (Estudiar OpenPages) (reverted).json

%EXPORTER_BIN% -input "%input_plantilla%" -output "%output_dir%" -mode v3
REM Ahora debes buscar el archivo generado para pasarlo al revert:
for %%f in ("%output_dir%\Plantilla (Estudiar OpenPages)_v3*.jsonl") do (
    set "input_jsonl=%%f"
    goto found_jsonl
)
:found_jsonl
%REVERT_BIN% "%input_plantilla%" "%input_jsonl%" "%output_revert%"
if errorlevel 1 (
    echo Error en revertido. Abortando.
    pause
    goto MENU
)
echo Pipeline completado correctamente.
pause
goto MENU

:END
echo Saliendo del script. ¡Hasta luego!
exit /b 0
