// internal/importer/reader.go – Lectura de tiddlers desde JSON exportado de TiddlyWiki
// ----------------------------------------------------------------------------------------------------
// Contexto pedagógico
// -------------------
// Este archivo vive en el paquete **importer** dentro de `internal/`, por lo que no se puede importar
// desde fuera del módulo.  Expone la función `Read`, encargada de convertir un export de TiddlyWiki
// (JSON) en un slice de `models.Tiddler` homogéneo.
//
// Firma pública:
//   Read(ctx context.Context, path string) ([]models.Tiddler, error)
//
// · `ctx` permite, en una futura versión streaming, cancelar la operación.
// · `path` es la ruta del archivo a leer.
//
// El algoritmo detecta automáticamente dos formatos de exportación:
//   1. Array JSON   → `[ {...}, {...} ]`
//   2. Objeto plano → `{ "id": {...}, "id2": {...} }`
// ----------------------------------------------------------------------------------------------------

package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// Read abre y deserializa el archivo indicado en `path`.
//
// Valores de retorno
// ------------------
//   - []models.Tiddler – tiddlers listos para procesar aguas abajo.
//   - error            – nil en éxito; descriptivo en caso de fallo.
func Read(ctx context.Context, path string) ([]models.Tiddler, error) {
	// Por el momento `ctx` no se usa porque la lectura se hace de un solo golpe.
	// Se acepta como parámetro para soportar cancelaciones cuando se implemente
	// el modo streaming con json.Decoder.
	_ = ctx

	// ---------------------------------------------------------------------
	// 1) Lectura de archivo completo
	// ---------------------------------------------------------------------
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo '%s': %w", path, err)
	}

	// ---------------------------------------------------------------------
	// 2) Intento: array JSON
	// ---------------------------------------------------------------------
	var list []models.Tiddler
	if err := json.Unmarshal(data, &list); err == nil {
		if len(list) == 0 {
			fmt.Println("⚠️  Archivo válido, pero el array de tiddlers está vacío.")
		}
		return list, nil
	}

	// ---------------------------------------------------------------------
	// 3) Intento: objeto plano (map)
	// ---------------------------------------------------------------------
	var mp map[string]models.Tiddler
	if err := json.Unmarshal(data, &mp); err == nil {
		tiddlers := make([]models.Tiddler, 0, len(mp))
		for _, t := range mp {
			tiddlers = append(tiddlers, t)
		}
		if len(tiddlers) == 0 {
			fmt.Println("⚠️  Archivo válido, pero el mapa de tiddlers está vacío.")
		}
		return tiddlers, nil
	}

	// ---------------------------------------------------------------------
	// 4) Formato desconocido
	// ---------------------------------------------------------------------
	return nil, fmt.Errorf("error al parsear JSON de tiddlers: no es ni array ni objeto plano válido")
}
