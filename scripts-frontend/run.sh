#!/usr/bin/env bash
# --------------------------------------------------------------------------------
# scripts/run.sh – Script interactivo para ejecutar el pipeline de exportación
# --------------------------------------------------------------------------------
# Este script te permite:
#   1) Seleccionar un archivo JSON de entrada (tiddlers) desde data/in/.
#   2) Definir el archivo de salida en data/out/ (JSONL o JSON “pretty”).
#   3) Elegir el modo de conversión (v1, v2, o v3).
#   4) Ejecutar el comando go run cmd/exporter/main.go con los flags adecuados.
#
# Para usarlo:
#   chmod +x scripts/run.sh
#   ./scripts/run.sh
#
# --------------------------------------------------------------------------------

set -euo pipefail

# ------------------------------------------- 1) Configuración base
# Determinar la carpeta raíz del proyecto (dos niveles arriba de este script).
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IN_DIR="$PROJECT_DIR/data/in"
OUT_DIR="$PROJECT_DIR/data/out"

# Asegurarnos de que IN_DIR y OUT_DIR existan
mkdir -p "$IN_DIR"
mkdir -p "$OUT_DIR"

# Buscar archivos JSON en data/in/ (solo extensión .json)
mapfile -t INPUT_FILES < <(find "$IN_DIR" -maxdepth 1 -type f -name '*.json' | sort)

if [ ${#INPUT_FILES[@]} -eq 0 ]; then
  echo "🔍 No se encontraron archivos .json en '$IN_DIR'. Coloca al menos uno y vuelve a intentarlo."
  exit 1
fi

echo "📁 Carpeta del proyecto: $PROJECT_DIR"
echo "📥 Archivos disponibles en data/in/:"
for idx in "${!INPUT_FILES[@]}"; do
  # Mostrar índice + 1 para que el usuario elija “1, 2, 3…”
  printf "    [%d] %s\n" "$((idx + 1))" "$(basename "${INPUT_FILES[$idx]}")"
done

# ------------------------------------------- 2) Selección interactiva de archivo de entrada
echo ""
while true; do
  read -rp "✏️  Elige el número del archivo JSON de entrada (1-${#INPUT_FILES[@]}): " sel
  # Validar que sea número entre 1 y cantidad de archivos
  if [[ "$sel" =~ ^[0-9]+$ ]] && (( sel >= 1 && sel <= ${#INPUT_FILES[@]} )); then
    INPUT_FILE="${INPUT_FILES[$((sel - 1))]}"
    echo "✅ Has seleccionado: $(basename "$INPUT_FILE")"
    break
  else
    echo "⚠️  Opción inválida. Ingresa un número entre 1 y ${#INPUT_FILES[@]}."
  fi
done

# ------------------------------------------- 3) Selección interactiva de modo de conversión
echo ""
echo "🧠 ¿Qué modo deseas usar para transformar los tiddlers?"
PS3="Selecciona una opción (1-3): "
options=("v1 (básico)" "v2 (estructurado con meta y content)" "v3 (AI-friendly)" "Cancelar")
select opt in "${options[@]}"; do
  case "$REPLY" in
    1) MODE_FLAG="v1"; break ;;
    2) MODE_FLAG="v2"; break ;;
    3) MODE_FLAG="v3"; break ;;
    4) echo "🚪 Cancelado."; exit 0 ;;
    *) echo "⚠️  Opción inválida. Intenta de nuevo." ;;
  esac
done

# ------------------------------------------- 4) Selección interactiva de formato de salida
echo ""
echo "📤 ¿Qué formato de salida deseas?"
PS3="Selecciona una opción (1-3): "
formatOptions=("JSONL (una línea por entrada, ideal para IA)" "JSON indentado (solo para inspección humana)" "Cancelar")
select optFmt in "${formatOptions[@]}"; do
  case "$REPLY" in
    1) PRETTY_FLAG=""; break ;;
    2) PRETTY_FLAG="-pretty"; break ;;
    3) echo "🚪 Cancelado."; exit 0 ;;
    *) echo "⚠️  Opción inválida. Intenta de nuevo." ;;
  esac
done

# ------------------------------------------- 5) Definir ruta de salida predeterminada y/o personalizada
DEFAULT_OUT_FILE="$OUT_DIR/$(basename "${INPUT_FILE%.*}.jsonl")"

echo ""
echo "💾 Archivo de salida sugerido: $(basename "$DEFAULT_OUT_FILE")"
read -rp "¿Deseas usar esta ruta? [S/n]: " useDefaultOut
useDefaultOut="${useDefaultOut:-S}"

if [[ "$useDefaultOut" =~ ^[Nn] ]]; then
  while true; do
    read -rp "✏️  Escribe el nombre de archivo (con extensión .jsonl o .json) dentro de data/out/: " customName
    # Aceptamos únicamente que termine en .jsonl o .json
    if [[ "$customName" =~ ^[[:alnum:]._-]+\.jsonl$ ]] || [[ "$customName" =~ ^[[:alnum:]._-]+\.json$ ]]; then
      OUTPUT_FILE="$OUT_DIR/$customName"
      break
    else
      echo "⚠️  Debe terminar en .jsonl o .json y no contener espacios."
    fi
  done
else
  OUTPUT_FILE="$DEFAULT_OUT_FILE"
fi

# Asegurar que la carpeta de salida exista
mkdir -p "$(dirname "$OUTPUT_FILE")"

# ------------------------------------------- 6) Mostrar resumen y ejecutar
echo ""
echo "🚀 Ejecutando exportación con los siguientes parámetros:"
echo "    Modo de conversión:     $MODE_FLAG"
if [[ -n "$PRETTY_FLAG" ]]; then
  echo "    Formato de salida:      JSON indentado"
else
  echo "    Formato de salida:      JSONL plano"
fi
echo "    Archivo de entrada:     $INPUT_FILE"
echo "    Archivo de salida:      $OUTPUT_FILE"
echo ""

# Comando para correr Go
go run "$PROJECT_DIR/cmd/exporter" \
  -input "$INPUT_FILE" \
  -output "$OUTPUT_FILE" \
  $PRETTY_FLAG

echo ""
echo "✅ Exportación completada. Revisa '$OUTPUT_FILE'."
