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

package exporter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// WriteJSONL serializa `records` al archivo `path` (crea o trunca).
//
// Parámetros
// ----------
//   - ctx     – permite cancelar cuando se implemente escritura chunked.
//   - path    – destino en disco (se crea/trunca).
//   - records – slice ordenado que se volcará tal cual.
//
// Retorna
// -------
//   - error – nil en éxito; descriptivo en cualquier fallo.
func WriteJSONL(ctx context.Context, path string, records []models.Record) error {
	_ = ctx // sin uso por ahora; se conservará para compatibilidad futura

	// 1) Crear/truncar archivo ------------------------------------------------
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("no se pudo crear '%s': %w", path, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("error al cerrar '%s': %w", path, cerr)
		}
	}()

	// 2) Buffer de escritura --------------------------------------------------
	w := bufio.NewWriter(file)

	for i, rec := range records {
		// Cancelación futura: if ctx.Err() != nil { return ctx.Err() }

		line, err := json.Marshal(rec)
		if err != nil {
			return fmt.Errorf("error serializando record #%d: %w", i, err)
		}
		if _, err := w.Write(line); err != nil {
			return fmt.Errorf("error escribiendo record #%d: %w", i, err)
		}
		if err := w.WriteByte('\n'); err != nil {
			return fmt.Errorf("error escribiendo \n en record #%d: %w", i, err)
		}
	}

	// 3) Flush ---------------------------------------------------------------
	if err := w.Flush(); err != nil {
		return fmt.Errorf("error al vaciar buffer: %w", err)
	}
	return nil
}
