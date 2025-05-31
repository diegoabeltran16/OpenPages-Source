// internal/exporter/writer.go – Persistencia de records en JSONL
// --------------------------------------------------------------------------------
// Contexto pedagógico
// -------------------
// Este archivo vive en el paquete **exporter** dentro de `internal/`.  Expone
// `WriteJSONL`, la pieza final del pipeline (importer → transform → exporter).
// Su misión es volcar `[]models.Record` al disco usando el formato **JSONL**
// (una línea JSON compacta por registro).
//
// Cambios respecto a la versión monolítica:
//   • paquete `exporter` (no `main`).
//   • Firma ahora acepta `context.Context` para futuras cancelaciones.
//   • Código sigue 100 % determinista: si algo falla, retorna `error`.
// --------------------------------------------------------------------------------
//
// Cambios clave
//   • Firma: WriteJSONL(ctx, path, records any, pretty bool)
//   • `records` debe ser un slice (v1 o v2) – se itera vía reflect.
//   • Flag `pretty` decide entre Marshal y MarshalIndent.
// --------------------------------------------------------------------------------

package exporter

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
)

// WriteJSONL serializa cualquier slice de structs a JSONL.
//
//   - records – debe ser un slice (p. ej. []models.Record o []models.RecordV2)
//   - pretty  – true → MarshalIndent (legible); false → Marshal (compacto)
func WriteJSONL(ctx context.Context, path string, records any, pretty bool) error {
	_ = ctx // reservado para cancelaciones futuras

	v := reflect.ValueOf(records)
	if v.Kind() != reflect.Slice {
		return errors.New("records debe ser slice")
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("crear '%s': %w", path, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("cerrar '%s': %w", path, cerr)
		}
	}()

	w := bufio.NewWriter(file)

	for i := 0; i < v.Len(); i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		elem := v.Index(i).Interface()

		var line []byte
		if pretty {
			line, err = json.MarshalIndent(elem, "", "  ")
		} else {
			line, err = json.Marshal(elem)
		}
		if err != nil {
			return fmt.Errorf("marshal elemento %d: %w", i, err)
		}
		if _, err := w.Write(line); err != nil {
			return fmt.Errorf("escribir elemento %d: %w", i, err)
		}
		if err := w.WriteByte('\n'); err != nil {
			return fmt.Errorf("newline elemento %d: %w", i, err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	return nil
}
