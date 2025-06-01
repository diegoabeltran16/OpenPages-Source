#!/bin/bash

# --------------------------------------------------------------------------------
# scripts/run.sh – Script interactivo para ejecutar el pipeline de exportación
# --------------------------------------------------------------------------------
# Permite al usuario seleccionar:
#   - Modo de exportación: v1 (simple) o v2 (estructurado)
#   - Formato: JSONL (para IA) o JSON (indentado para humanos)
#   - Archivos de entrada y salida (predefinidos o personalizados)
# --------------------------------------------------------------------------------

set -e

# ------------------------------------------- Configuración base
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IN_DIR="$PROJECT_DIR/data/in"
OUT_DIR="$PROJECT_DIR/data/out"
DEFAULT_IN_FILE="$(find "$IN_DIR" -name '*.json' | head -n1)"
DEFAULT_OUT_FILE="$OUT_DIR/out.jsonl"

echo "📁 Carpeta del proyecto: $PROJECT_DIR"
echo "📥 Archivo de entrada sugerido: $DEFAULT_IN_FILE"

# ------------------------------------------- Selección de modo
echo ""
echo "🧠 ¿Qué modo deseas usar para transformar los tiddlers?"
select MODE in "v1 (básico)" "v2 (estructurado con meta y content)" "Cancelar"; do
  case $REPLY in
    1) MODE_FLAG="v1"; break ;;
    2) MODE_FLAG="v2"; break ;;
    3) echo "🚪 Cancelado."; exit 0 ;;
    *) echo "Opción inválida. Intenta de nuevo." ;;
  esac
done

# ------------------------------------------- Selección de formato de salida
echo ""
echo "📤 ¿Qué formato de salida deseas?"
select PRETTY in "JSONL (una línea por entrada, ideal para IA)" "JSON indentado (solo para inspección humana)" "Cancelar"; do
  case $REPLY in
    1) PRETTY_FLAG=""; break ;;
    2) PRETTY_FLAG="-pretty"; break ;;
    3) echo "🚪 Cancelado."; exit 0 ;;
    *) echo "Opción inválida. Intenta de nuevo." ;;
  esac
done


# ------------------------------------------- Confirmación de archivo de entrada
echo ""
echo "🔍 Archivo de entrada detectado: $DEFAULT_IN_FILE"
read -rp "¿Deseas usar esta ruta? [S/n]: " CONFIRM
CONFIRM="${CONFIRM:-S}"

if [[ "$CONFIRM" =~ ^[Nn] ]]; then
  read -rp "Escribe la nueva ruta completa del archivo de entrada: " INPUT_FILE
else
  INPUT_FILE="$DEFAULT_IN_FILE"
fi

echo ""
echo "💾 Archivo de salida sugerido: $DEFAULT_OUT_FILE"
read -rp "¿Deseas usar esta ruta? [S/n]: " OUT_CONFIRM
OUT_CONFIRM="${OUT_CONFIRM:-S}"

if [[ "$OUT_CONFIRM" =~ ^[Nn] ]]; then
  read -rp "Escribe la nueva ruta completa del archivo de salida: " OUTPUT_FILE
else
  OUTPUT_FILE="$DEFAULT_OUT_FILE"
fi

# ------------------------------------------- Ejecución
echo ""
echo "🚀 Ejecutando exportación..."
echo "    Modo:   $MODE_FLAG"
echo "    Formato: $( [ -n "$PRETTY_FLAG" ] && echo 'JSON indentado' || echo 'JSONL plano')"
echo "    Entrada: $INPUT_FILE"
echo "    Salida:  $OUTPUT_FILE"
echo ""

go run "$PROJECT_DIR/cmd/exporter" \
  -input "$INPUT_FILE" \
  -output "$OUTPUT_FILE" \
  -mode "$MODE_FLAG" \
  $PRETTY_FLAG

echo ""
echo "✅ Exportación completada."
