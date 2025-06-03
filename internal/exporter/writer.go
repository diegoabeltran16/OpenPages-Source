// internal/exporter/writer.go – Persistencia de records en JSONL
// --------------------------------------------------------------------------------
// Contexto pedagógico
// -------------------
// Aquí mejoramos el Writer para asegurarnos de que:
//
//   1. El directorio padre de `path` exista (creamos con os.MkdirAll).
//   2. Usamos un `err` nombrado para que el `defer` detecte problemas al cerrar.
//   3. Registramos (mensajito a stdout) cuántos registros vamos a escribir.
//   4. Retornamos errores claros si algo falla en cualquiera de las etapas.
//
// Firma:
//   WriteJSONL(ctx, path, records any, pretty bool) error
//     - `path` puede incluir subdirectorios: si no existen, los creamos.
//     - `records` debe ser un slice (por ejemplo []models.RecordV2).
//     - `pretty` decide entre json.Marshal (compacto) o json.MarshalIndent (legible).
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

// WriteJSONL serializa cualquier slice de structs a un archivo JSONL.
//
//   - ctx: contexto para abortar si se desea cancelar.
//   - path: ruta donde se escribirá el JSONL (se crea directorio padre si hace falta).
//   - records: debe ser un slice de structs (por ejemplo []models.RecordV2).
//   - pretty: si true, se usa MarshalIndent; si false, Marshal (una línea compacta por registro).
func WriteJSONL(ctx context.Context, path string, records any, pretty bool) (err error) {
	_ = ctx // reservado para posibles cancelaciones en el futuro

	// 1) Verificar que records sea slice
	v := reflect.ValueOf(records)
	if v.Kind() != reflect.Slice {
		return errors.New("records debe ser slice")
	}
	count := v.Len()

	// 2) Asegurarnos de que el directorio padre exista
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if mkdirErr := os.MkdirAll(dir, 0o755); mkdirErr != nil {
			return fmt.Errorf("mkdirall '%s': %w", dir, mkdirErr)
		}
	}

	// 3) Crear (o truncar) el archivo
	file, createErr := os.Create(path)
	if createErr != nil {
		return fmt.Errorf("crear '%s': %w", path, createErr)
	}
	// Usamos named return 'err' para que defer pueda capturar errores de Close
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("cerrar '%s': %w", path, cerr)
		}
	}()

	// 4) Preparar el writer con buffer
	w := bufio.NewWriter(file)

	// 5) Mensaje informativo: cuántos registros vamos a escribir
	fmt.Printf("💾 Escribiendo %d registros en '%s'...\n", count, path)

	// 6) Iterar sobre cada elemento en el slice
	for i := 0; i < count; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		elem := v.Index(i).Interface()

		// 6.1) Serializar a JSON (indentado o compacto según `pretty`)
		var line []byte
		if pretty {
			line, err = json.MarshalIndent(elem, "", "  ")
		} else {
			line, err = json.Marshal(elem)
		}
		if err != nil {
			return fmt.Errorf("marshal elemento %d: %w", i, err)
		}

		// 6.2) Escribir la línea y un salto de línea
		if _, err = w.Write(line); err != nil {
			return fmt.Errorf("escribir elemento %d: %w", i, err)
		}
		if err = w.WriteByte('\n'); err != nil {
			return fmt.Errorf("newline elemento %d: %w", i, err)
		}
	}

	// 7) Hacer flush del buffer para asegurar que todo esté en disco
	if err = w.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}

	return nil
}
