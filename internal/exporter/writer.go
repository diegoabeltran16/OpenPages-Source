// internal/exporter/writer.go – Persistencia de registros en JSONL (v3)
// --------------------------------------------------------------------------------
// Contexto pedagógico
// -------------------
// Esta versión robustecida de WriteJSONL garantiza:
//
//   1. Creación del directorio padre si no existe.
//   2. Named return para capturar errores al cerrar.
//   3. Impresión en consola de la cantidad de objetos que se escribirán.
//   4. Serialización de cualquier slice (p.ej. []map[string]any de v3) a JSONL estricto.
//   5. Opción “pretty” para inspección humana: aunque genere multilínea,
//      siempre agrega un solo '\n' al final de cada objeto.
//
// Firma:
//   WriteJSONL(ctx, path, records any, pretty bool) error
//     - ctx: contexto para cancelaciones futuras.
//     - path: ruta al archivo de salida (se crea su carpeta si falta).
//     - records: debe ser un slice (p.ej. []models.Record, []models.RecordV2 o []map[string]any).
//     - pretty: si true, MarshalIndent (multilínea); si false, Marshal compacto (una línea por objeto).
// --------------------------------------------------------------------------------

package exporter

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

// WriteJSONL serializa cualquier slice de elementos a un archivo JSONL.
// Cada elemento del slice se convierte en un JSON y se escribe como una línea.
// Si pretty==true, cada objeto queda indentado (multilínea) pero con un solo '\n' al final;
// si pretty==false, cada objeto ocupa exactamente una línea compacta.
//
// Ejemplo con v3:
//
//	recs := transform.ConvertTiddlersV3(tiddlers)  // []map[string]any
//	WriteJSONL(ctx, "out.jsonl", recs, false)
//
// Ejemplo con v2:
//
//	recsV2 := transform.ConvertTiddlersV2(tiddlers)  // []models.RecordV2
//	WriteJSONL(ctx, "out_pretty.json", recsV2, true)
func WriteJSONL(ctx context.Context, path string, records any, pretty bool) (err error) {
	_ = ctx // reservado para cancelaciones futuras

	// 1) Verificar que 'records' sea un slice
	v := reflect.ValueOf(records)
	if v.Kind() != reflect.Slice {
		return errors.New("records debe ser un slice")
	}
	count := v.Len()

	// 2) Asegurarnos de que el directorio padre exista
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if mkdirErr := os.MkdirAll(dir, 0o755); mkdirErr != nil {
			return fmt.Errorf("mkdirall '%s': %w", dir, mkdirErr)
		}
	}

	// 3) Crear (o truncar) el archivo de salida
	file, createErr := os.Create(path)
	if createErr != nil {
		return fmt.Errorf("crear '%s': %w", path, createErr)
	}
	// Named return para capturar posibles errores al cerrar
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("cerrar '%s': %w", path, cerr)
		}
	}()

	// 4) Preparar buffer para escritura
	w := bufio.NewWriter(file)

	// 5) Informar cuántos registros se escribirán
	fmt.Printf("💾 Escribiendo %d registros en '%s'...\n", count, path)

	// 6) Iterar sobre cada elemento del slice
	for i := 0; i < count; i++ {
		// Permitir cancelación si ctx se ha cancelado
		if ctx.Err() != nil {
			return ctx.Err()
		}

		elem := v.Index(i).Interface()

		// 6.1) Serializar a JSON
		var line []byte
		if pretty {
			line, err = json.MarshalIndent(elem, "", "  ")
		} else {
			line, err = json.Marshal(elem)
		}
		if err != nil {
			return fmt.Errorf("marshal elemento %d: %w", i, err)
		}

		// 6.2) Escribir el JSON y un solo '\n'
		if _, err = w.Write(line); err != nil {
			return fmt.Errorf("escribir elemento %d: %w", i, err)
		}
		if err = w.WriteByte('\n'); err != nil {
			return fmt.Errorf("newline elemento %d: %w", i, err)
		}
	}

	// 7) Forzar escritura en disco
	if err = w.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}

	return nil
}
