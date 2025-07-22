// internal/transform/reverse.go – Reversión de JSONL enriquecido a JSON TiddlyWiki
// --------------------------------------------------------------------------------
// Contexto pedagógico
// -------------------
// Esta función complementaria permite la **reversión** del pipeline principal:
//   TiddlyWiki JSON → JSONL (v3) → TiddlyWiki JSON
//
// Casos de uso:
//   1. Verificar que el pipeline es bidireccional y sin pérdida de datos.
//   2. Permitir ediciones en JSONL y reimportar a TiddlyWiki.
//   3. Migrar entre instancias de TiddlyWiki usando JSONL como formato intermedio.
//
// Algoritmo:
//   1. Lee archivo JSONL línea por línea.
//   2. Parsea cada línea como map[string]any.
//   3. Convierte campos enriquecidos de vuelta al formato TiddlyWiki:
//      - RFC3339 → formato TiddlyWiki (yyyymmddhhMMSS)
//      - []string tags → "[[tag1]] [[tag2]]"
//      - Campos simples directos
//   4. Serializa como array JSON con indentación.
//
// Firma:
//   ReverseJSONLToTiddlyJSON(inputPath, outputPath string) error
// --------------------------------------------------------------------------------

package transform

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// ReverseJSONLToTiddlyJSON lee un archivo JSONL (generado por ConvertTiddlersV3)
// y lo convierte de vuelta al formato JSON de TiddlyWiki compatible.
//
// Transformaciones aplicadas:
//   - RFC3339 dates → TiddlyWiki format (20060102150405)
//   - []string tags → "[[tag1]] [[tag2]]" format
//   - map[string]any → models.Tiddler structs
//   - JSONL lines → JSON array with indentation
//
// Ejemplo:
//
//	ReverseJSONLToTiddlyJSON("data/out/tiddlers.jsonl", "data/out/restored.json")
func ReverseJSONLToTiddlyJSON(inputPath, outputPath string) error {
	// 1) Abrir archivo JSONL de entrada
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("no se pudo abrir archivo JSONL '%s': %w", inputPath, err)
	}
	defer file.Close()

	var tiddlers []models.Tiddler
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	// 2) Procesar cada línea del JSONL
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Saltar líneas vacías
		if line == "" {
			continue
		}

		// 3) Parsear línea JSON
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return fmt.Errorf("error parseando línea %d: %w", lineNumber, err)
		}

		// 4) Convertir registro de vuelta a Tiddler
		tiddler, err := recordToTiddler(record)
		if err != nil {
			return fmt.Errorf("error convirtiendo línea %d: %w", lineNumber, err)
		}

		tiddlers = append(tiddlers, tiddler)
	}

	// 5) Verificar errores de lectura
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error leyendo archivo JSONL: %w", err)
	}

	// 6) Crear archivo de salida
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("no se pudo crear archivo de salida '%s': %w", outputPath, err)
	}
	defer outputFile.Close()

	// 7) Serializar como JSON con indentación (formato TiddlyWiki)
	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(tiddlers); err != nil {
		return fmt.Errorf("error escribiendo JSON de salida: %w", err)
	}

	fmt.Printf("🔄 Reversión completada: %d tiddlers convertidos\n", len(tiddlers))
	return nil
}

// recordToTiddler convierte un map[string]any (de JSONL v3) de vuelta a models.Tiddler
func recordToTiddler(record map[string]any) (models.Tiddler, error) {
	tiddler := models.Tiddler{}

	// Campos string simples
	if id, ok := record["id"].(string); ok {
		tiddler.Title = id
	}
	if text, ok := record["text"].(string); ok {
		tiddler.Text = text
	}
	if typ, ok := record["type"].(string); ok {
		tiddler.Type = typ
	}
	if tmapID, ok := record["tmap.id"].(string); ok {
		tiddler.TmapID = tmapID
	}

	// Fechas: convertir de RFC3339 a formato TiddlyWiki
	if created, ok := record["created"].(string); ok {
		if t, err := parseRFC3339ToTW(created); err == nil {
			tiddler.Created = t
		} else {
			// Fallback: usar fecha actual si el parseo falla
			tiddler.Created = time.Now().Format("20060102150405")
		}
	}

	if modified, ok := record["modified"].(string); ok {
		if t, err := parseRFC3339ToTW(modified); err == nil {
			tiddler.Modified = t
		} else {
			// Fallback: usar fecha actual si el parseo falla
			tiddler.Modified = time.Now().Format("20060102150405")
		}
	}

	// Tags: convertir de []interface{} a formato TiddlyWiki "[[tag1]] [[tag2]]"
	if tags, ok := record["tags"].([]interface{}); ok {
		var tagStrings []string
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				// Envolver cada tag en [[ ]]
				tagStrings = append(tagStrings, fmt.Sprintf("[[%s]]", tagStr))
			}
		}
		tiddler.Tags = strings.Join(tagStrings, " ")
	}

	// Campos opcionales que podrían estar presentes
	if color, ok := record["color"].(string); ok {
		tiddler.Color = color
	}

	return tiddler, nil
}

// parseRFC3339ToTW convierte una fecha RFC3339 de vuelta al formato TiddlyWiki
func parseRFC3339ToTW(rfc3339Str string) (string, error) {
	// Intentar varios formatos RFC3339
	layouts := []string{
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05.000-07:00",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, rfc3339Str); err == nil {
			// Convertir a formato TiddlyWiki: yyyymmddhhMMSS
			return t.Format("20060102150405"), nil
		}
	}

	return "", fmt.Errorf("formato de fecha no reconocido: %s", rfc3339Str)
}
