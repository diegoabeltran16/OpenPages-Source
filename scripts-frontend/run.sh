#!/usr/bin/env bash
# --------------------------------------------------------------------------------
# scripts-frontend/run.sh – Script interactivo para exportar y revertir tiddlers
# --------------------------------------------------------------------------------

set -euo pipefail

# ------------------------------------------- 1) Configuración base
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IN_DIR="$PROJECT_DIR/data/in"
OUT_DIR="$PROJECT_DIR/data/out"

mkdir -p "$IN_DIR"
mkdir -p "$OUT_DIR"

# ------------------------------------------- 2) Menú principal
while true; do
  echo ""
  echo "=========================================="
  echo "  OpenPages-Source - Menú de acciones"
  echo "=========================================="
  echo "  [1] Exportar tiddlers desde plantilla"
  echo "  [2] Revertir y actualizar textos desde JSONL"
  echo "  [3] Ejecutar ambos procesos (pipeline completo)"
  echo "  [0] Salir"
  echo ""
  read -rp "Elige una opción [0-3]: " choice

  case "$choice" in
    1)
      # ------------------------------------------- Exportar tiddlers
      echo ""
      mapfile -t INPUT_FILES < <(find "$IN_DIR" -maxdepth 1 -type f -name '*.json' | sort)
      if [ ${#INPUT_FILES[@]} -eq 0 ]; then
        echo "🔍 No se encontraron archivos .json en '$IN_DIR'."
        continue
      fi
      echo "Archivos disponibles en data/in/:"
      for idx in "${!INPUT_FILES[@]}"; do
        printf "    [%d] %s\n" "$((idx + 1))" "$(basename "${INPUT_FILES[$idx]}")"
      done
      while true; do
        read -rp "Elige el número del archivo JSON de entrada (1-${#INPUT_FILES[@]}): " sel
        if [[ "$sel" =~ ^[0-9]+$ ]] && (( sel >= 1 && sel <= ${#INPUT_FILES[@]} )); then
          INPUT_FILE="${INPUT_FILES[$((sel - 1))]}"
          break
        else
          echo "Opción inválida."
        fi
      done

      echo ""
      echo "Modos de exportación disponibles:"
      echo "  [1] v1"
      echo "  [2] v2"
      echo "  [3] v3"
      echo "  [4] hybrid"
      read -rp "Selecciona el modo por número [default: 3]: " sel_mode
      sel_mode="${sel_mode:-3}"
      case "$sel_mode" in
        1) MODE_FLAG="v1" ;;
        2) MODE_FLAG="v2" ;;
        3) MODE_FLAG="v3" ;;
        4) MODE_FLAG="hybrid" ;;
        *) MODE_FLAG="v3" ;;
      esac

      read -rp "¿Salida pretty? (y/n) [default: n]: " pretty
      if [[ "$pretty" =~ ^[Yy]$ ]]; then
        PRETTY_FLAG="-pretty"
      else
        PRETTY_FLAG=""
      fi

      DEFAULT_OUT_FILE="$OUT_DIR/$(basename "${INPUT_FILE%.*}.jsonl")"
      echo ""
      echo "Archivo de salida sugerido: $(basename "$DEFAULT_OUT_FILE")"
      read -rp "¿Deseas usar esta ruta? [S/n]: " useDefaultOut
      useDefaultOut="${useDefaultOut:-S}"
      if [[ "$useDefaultOut" =~ ^[Nn]$ ]]; then
        while true; do
          read -rp "Escribe el nombre de archivo (con extensión .jsonl o .json) dentro de data/out/: " customName
          if [[ "$customName" =~ ^[[:alnum:]._-]+\.jsonl$ ]] || [[ "$customName" =~ ^[[:alnum:]._-]+\.json$ ]]; then
            OUTPUT_FILE="$OUT_DIR/$customName"
            break
          else
            echo "Debe terminar en .jsonl o .json y no contener espacios."
          fi
        done
      else
        OUTPUT_FILE="$DEFAULT_OUT_FILE"
      fi

      mkdir -p "$(dirname "$OUTPUT_FILE")"
      echo ""
      echo "🚀 Ejecutando exportación:"
      echo "    Entrada: $INPUT_FILE"
      echo "    Salida:  $OUTPUT_FILE"
      echo "    Modo:    $MODE_FLAG"
      echo "    Pretty:  $PRETTY_FLAG"
      "$PROJECT_DIR/openpages_exporter.exe" -input "$INPUT_FILE" -output "$OUTPUT_FILE" -mode "$MODE_FLAG" $PRETTY_FLAG
      echo "✅ Exportación completada. Revisa '$OUTPUT_FILE'."
      ;;

    2)
      # ------------------------------------------- Revertir y actualizar textos
      echo ""
      mapfile -t PLANTILLA_FILES < <(find "$IN_DIR" -maxdepth 1 -type f -name '*.json' | sort)
      if [ ${#PLANTILLA_FILES[@]} -eq 0 ]; then
        echo "🔍 No se encontraron archivos .json en '$IN_DIR'."
        continue
      fi
      echo "Archivos disponibles en data/in/:"
      for idx in "${!PLANTILLA_FILES[@]}"; do
        printf "    [%d] %s\n" "$((idx + 1))" "$(basename "${PLANTILLA_FILES[$idx]}")"
      done
      while true; do
        read -rp "Selecciona el archivo plantilla por número (1-${#PLANTILLA_FILES[@]}): " sel_plantilla
        if [[ "$sel_plantilla" =~ ^[0-9]+$ ]] && (( sel_plantilla >= 1 && sel_plantilla <= ${#PLANTILLA_FILES[@]} )); then
          PLANTILLA_FILE="${PLANTILLA_FILES[$((sel_plantilla - 1))]}"
          break
        else
          echo "Opción inválida."
        fi
      done

      mapfile -t TEXTOS_FILES < <(find "$OUT_DIR" -maxdepth 1 -type f -name '*.jsonl' | sort)
      if [ ${#TEXTOS_FILES[@]} -eq 0 ]; then
        echo "🔍 No se encontraron archivos .jsonl en '$OUT_DIR'."
        continue
      fi
      echo "Archivos disponibles en data/out/:"
      for idx in "${!TEXTOS_FILES[@]}"; do
        printf "    [%d] %s\n" "$((idx + 1))" "$(basename "${TEXTOS_FILES[$idx]}")"
      done
      while true; do
        read -rp "Selecciona el archivo JSONL por número (1-${#TEXTOS_FILES[@]}): " sel_textos
        if [[ "$sel_textos" =~ ^[0-9]+$ ]] && (( sel_textos >= 1 && sel_textos <= ${#TEXTOS_FILES[@]} )); then
          TEXTOS_FILE="${TEXTOS_FILES[$((sel_textos - 1))]}"
          break
        else
          echo "Opción inválida."
        fi
      done

      REVERTED_FILE="$IN_DIR/$(basename "${TEXTOS_FILE%.*} (reverted).json")"
      echo ""
      echo "🚀 Ejecutando revertido y actualización de textos:"
      echo "    Plantilla: $PLANTILLA_FILE"
      echo "    Textos:    $TEXTOS_FILE"
      echo "    Salida:    $REVERTED_FILE"
      "$PROJECT_DIR/openpages_revert.exe" "$PLANTILLA_FILE" "$TEXTOS_FILE" "$REVERTED_FILE"
      echo "✅ Revertido completado. Revisa '$REVERTED_FILE'."
      ;;

    3)
      # ------------------------------------------- Pipeline completo
      echo ""
      mapfile -t INPUT_FILES < <(find "$IN_DIR" -maxdepth 1 -type f -name '*.json' | sort)
      if [ ${#INPUT_FILES[@]} -eq 0 ]; then
        echo "🔍 No se encontraron archivos .json en '$IN_DIR'."
        continue
      fi
      INPUT_FILE="${INPUT_FILES[0]}"
      OUTPUT_FILE="$OUT_DIR/$(basename "${INPUT_FILE%.*}.jsonl")"
      REVERTED_FILE="$IN_DIR/$(basename "${INPUT_FILE%.*} (reverted).json")"
      echo "🚀 Ejecutando pipeline completo:"
      "$PROJECT_DIR/openpages_exporter.exe" -input "$INPUT_FILE" -output "$OUTPUT_FILE" -mode v3
      "$PROJECT_DIR/openpages_revert.exe" "$INPUT_FILE" "$OUTPUT_FILE" "$REVERTED_FILE"
      echo "✅ Pipeline completado. Revisa '$REVERTED_FILE'."
      ;;

    0)
      echo "Saliendo del script. ¡Hasta luego!"
      exit 0
      ;;

    *)
      echo "Opción inválida. Intenta de nuevo."
      ;;
  esac
done