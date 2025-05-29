// reader.go – Lectura de tiddlers desde JSON exportado de TiddlyWiki
// -----------------------------------------------------------
// Ubicación: raíz del proyecto.
// Responsabilidad: leer un archivo JSON exportado de TiddlyWiki,
// soportando dos formatos de entrada:
//  1. Array de objetos JSON (`[ {...}, {...} ]`)
//  2. Objeto plano con claves como IDs (`{ "ID1": {...}, "ID2": {...} }`)
//
// Devuelve un slice de models.Tiddler listo para procesamiento.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// ReadTiddlers abre y deserializa un archivo JSON exportado de TiddlyWiki.
// - Intenta primero leer como slice plano de Tiddlers.
// - Si falla, intenta leer como mapa de ID→Tiddler.
// Parámetros:
//   - path: ruta al archivo JSON de TiddlyWiki.
//
// Retorna:
//   - []models.Tiddler: slice de tiddlers leídos.
//   - error: información detallada si falla la lectura o el parseo.
func ReadTiddlers(path string) ([]models.Tiddler, error) {
	// Leer contenido del archivo
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo '%s': %w", path, err)
	}

	// 1) Intentar parsear como array JSON
	var list []models.Tiddler
	if err := json.Unmarshal(data, &list); err == nil {
		if len(list) == 0 {
			fmt.Println("⚠️  Archivo leído correctamente, pero no contiene tiddlers en formato array.")
		}
		return list, nil
	}

	// 2) Intentar parsear como objeto plano
	var mp map[string]models.Tiddler
	if err := json.Unmarshal(data, &mp); err == nil {
		tiddlers := make([]models.Tiddler, 0, len(mp))
		for _, t := range mp {
			tiddlers = append(tiddlers, t)
		}
		if len(tiddlers) == 0 {
			fmt.Println("⚠️  Archivo leído correctamente, pero no contiene tiddlers en formato map.")
		}
		return tiddlers, nil
	}

	// 3) Ningún formato válido
	return nil, fmt.Errorf("error al parsear JSON de tiddlers: no es ni array ni objeto plano válido")
}
