// reader.go – Lectura de tiddlers desde JSON exportado de TiddlyWiki
// -----------------------------------------------------------
// Ubicación: raíz del proyecto.
// Responsabilidad: leer el archivo JSON de TiddlyWiki y devolver un slice de models.Tiddler,
// manejando errores de forma clara y documentada.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"diegoabeltran16/OpenPages-Source/models"
)

// ReadTiddlers abre y deserializa un archivo JSON exportado de TiddlyWiki.
// Parámetros:
//   - path: ruta al archivo JSON de TiddlyWiki.
// Retorna:
//   - []models.Tiddler: slice de tiddlers leídos.
//   - error: información detallada si falla la lectura o el parseo.
func ReadTiddlers(path string) ([]models.Tiddler, error) {
	// 1) Leer todo el contenido del archivo
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo '%s': %w", path, err)
	}

	// 2) Definir estructura temporal para el JSON
	var payload struct {
		Tiddlers []models.Tiddler `json:"tiddlers"`
	}

	// 3) Deserializar JSON en la estructura
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("error al parsear JSON de tiddlers: %w", err)
	}

	// 4) Retornar los tiddlers extraídos
	return payload.Tiddlers, nil
}
