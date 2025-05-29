// converter.go – Conversión de Tiddler a Record con tags parseados y JSON embebido legible
// -----------------------------------------------------------
// Ubicación: raíz del proyecto.
// Responsabilidad: transformar cada models.Tiddler en un models.Record,
// extrayendo etiquetas y formateando contenido JSON para mejor lectura.
package main

import (
	"bytes"
	"encoding/json"
	"regexp"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// tagRe captura texto dentro de [[...]]
var tagRe = regexp.MustCompile(`\[\[([^]]+)\]\]`)

// parseTags extrae etiquetas de raw ("[[tag1]] [[tag2]]") a slice de strings.
func parseTags(raw string) []string {
	matches := tagRe.FindAllStringSubmatch(raw, -1)
	tags := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			tags = append(tags, m[1])
		}
	}
	return tags
}

// ConvertTiddlers convierte un slice de models.Tiddler en []models.Record
// listo para exportación JSONL.
func ConvertTiddlers(tiddlers []models.Tiddler) []models.Record {
	recs := make([]models.Record, 0, len(tiddlers))
	for _, t := range tiddlers {
		rec := models.Record{
			ID:          t.Title,
			Tags:        parseTags(t.Tags),
			ContentType: t.Type,
			CreatedAt:   t.Created,
			ModifiedAt:  t.Modified,
		}

		// Formatear JSON embebido con indentación si aplica
		if t.Type == "application/json" {
			var buf bytes.Buffer
			if err := json.Indent(&buf, []byte(t.Text), "", "  "); err == nil {
				rec.TextMarkdown = buf.String()
				rec.TextPlain = buf.String()
			} else {
				rec.TextMarkdown = t.Text
				rec.TextPlain = t.Text
			}
		} else {
			rec.TextMarkdown = t.Text
			rec.TextPlain = t.Text
		}

		recs = append(recs, rec)
	}
	return recs
}
