// internal/transform/converter.go – Conversión de Tiddler → Record
// --------------------------------------------------------------------------------
// Contexto pedagógico
// -------------------
// Este archivo vive en **`internal/transform`** y expone la función
// `ConvertTiddlers`, encargada de transformar la estructura de dominio
// `models.Tiddler` en `models.Record`, lista para persistir como JSONL.
//
// Responsabilidades clave
// -----------------------
// 1. **Normalizar etiquetas**: "[[foo]] [[bar baz]]" → []string{"foo", "bar baz"}.
// 2. **Copiar metadatos**: Title → ID, Created → CreatedAt, etc.
// 3. **Embellecer JSON embebido** (`application/json`) con indentado.
//
// No altera la lógica previa; sólo cambia:
//   • *package transform* (en vez de main).
//   • Los imports se reducen a stdlib + models.
//
// --------------------------------------------------------------------------------

package transform

import (
	"bytes"
	"encoding/json"
	"regexp"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// tagRe detecta etiquetas [[...]].  El sub‑grupo captura el texto interno.
var tagRe = regexp.MustCompile(`\[\[([^]]+)\]\]`)

// parseTags convierte la cadena raw de tags en un slice limpio.
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

// ConvertTiddlers itera sobre los tiddlers y genera []models.Record.
// No retorna error porque la operación es puramente determinística y cualquier
// fallo se degrada (e.g. JSON malformado → se copia tal cual).
func ConvertTiddlers(ts []models.Tiddler) []models.Record {
	recs := make([]models.Record, 0, len(ts))

	for _, t := range ts {
		rec := models.Record{
			ID:          t.Title,
			Tags:        parseTags(t.Tags),
			ContentType: t.Type,
			CreatedAt:   t.Created,
			ModifiedAt:  t.Modified,
		}

		// Formateo especial si el cuerpo es JSON puro.
		if t.Type == "application/json" {
			var buf bytes.Buffer
			if err := json.Indent(&buf, []byte(t.Text), "", "  "); err == nil {
				rec.TextMarkdown = buf.String()
				rec.TextPlain = buf.String()
			} else {
				rec.TextMarkdown = t.Text // JSON inválido → copiar sin tocar
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
