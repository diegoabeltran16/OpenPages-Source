// internal/transform/converter.go – v1, v2 y nueva versión v3 con esquema mínimo
// --------------------------------------------------------------------------------
// Este archivo expone tres funciones públicas:
//   • ConvertTiddlers   → genera []models.Record     (esquema heredado v1).
//   • ConvertTiddlersV2 → genera []models.RecordV2   (esquema AI-friendly v2).
//   • ConvertTiddlersV3 → genera []map[string]any    (esquema mínimo para JSONL estricto v3).
//
// La versión v3 produce objetos JSON planos que cumplen con:
//
//   - Una sola línea por objeto (ideal para JSONL).
//   - Campos esenciales: id, title, created, modified, tags, tmap.id, relations, type, text.
//   - Fechas en RFC3339 con zona (por ejemplo "2025-06-05T15:10:00-05:00").
//   - Sin duplicación de tags ni niveles de anidación innecesarios.
//
// De esta manera, un JSONL estricto tendrá líneas como:
//
//   {"id":"_____BirdsColor","title":"_____BirdsColor","created":"2025-06-05T15:10:00-05:00", ... }
//
// --------------------------------------------------------------------------------

package transform

import (
	"bytes"
	"encoding/json"
	"regexp"
	"time"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// -----------------------------------------------------------------------------
// Utilidades compartidas
// -----------------------------------------------------------------------------

var tagRe = regexp.MustCompile(`\[\[([^]]+)\]\]`)

// parseTags extrae todas las etiquetas de la forma [[etiqueta]] de un string.
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

// parseTWDate intenta parsear un string TiddlyWiki (yyyymmddhhMMSS o yyyymmdd).
// Devuelve time.Time y true si tuvo éxito; de lo contrario, time.Time{} y false.
func parseTWDate(raw string) (time.Time, bool) {
	layouts := []string{"20060102150405", "20060102"}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, raw, time.UTC); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// formatISO8601 formatea un time.Time en RFC3339 con offset, p.ej. "2025-06-05T15:10:00-05:00".
func formatISO8601(t time.Time) string {
	// Si t es cero, usamos la hora actual
	if t.IsZero() {
		return time.Now().Format("2006-01-02T15:04:05-07:00")
	}
	return t.Format("2006-01-02T15:04:05-07:00")
}

// -----------------------------------------------------------------------------
// Versión 1 – lógica intacta (esquema heredado)
// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------
// Versión 2 – esquema meta/content (AI-friendly)
// -----------------------------------------------------------------------------

func ConvertTiddlersV2(ts []models.Tiddler) []models.RecordV2 {
	recs := make([]models.RecordV2, 0, len(ts))

	for _, t := range ts {
		// Meta
		created, _ := parseTWDate(t.Created)
		modified, _ := parseTWDate(t.Modified)

		meta := models.Meta{
			Title:    t.Title,
			Tags:     parseTags(t.Tags),
			Created:  created,
			Modified: modified,
			Color:    t.Color,
			Extra: map[string]string{
				"tmap.id": t.TmapID,
			},
		}

		// Content
		var content models.Content
		switch t.Type {
		case "application/json":
			var obj map[string]any
			if err := json.Unmarshal([]byte(t.Text), &obj); err == nil {
				content.JSON = obj
			} else {
				content.Plain = t.Text
			}
		case "text/x-markdown":
			content.Markdown = t.Text
		default:
			content.Plain = t.Text
		}

		rec := models.RecordV2{
			ID:        t.Title,
			Type:      "tiddler",
			Meta:      meta,
			Content:   content,
			Relations: nil,
		}
		recs = append(recs, rec)
	}
	return recs
}

// -----------------------------------------------------------------------------
// Versión 3 – esquema mínimo para JSONL estricto (una línea por objeto)
// -----------------------------------------------------------------------------

// ConvertTiddlersV3 recibe []models.Tiddler y devuelve []map[string]any
// donde cada map corresponde a un JSON plano sin saltos de línea internos.
// Campos incluidos:
//   - "id", "title": ambos iguales a t.Title
//   - "created", "modified": ISO8601 con zona, o ahora si no se parsea
//   - "tags": []string (de parseTags)
//   - "tmap.id": string
//   - "relations": map[string][]string (si aplica; aquí nil)
//   - "type": t.Type
//   - "text": t.Text (plano o markdown)
//
// No se duplica tags en otro nivel. Ideal para JSONL.
func ConvertTiddlersV3(ts []models.Tiddler) []map[string]any {
	recs := make([]map[string]any, 0, len(ts))

	for _, t := range ts {
		// 1) Fechas en time.Time
		createdTime, okC := parseTWDate(t.Created)
		modifiedTime, okM := parseTWDate(t.Modified)

		// 2) Formatear fechas a string ISO8601
		createdStr := formatISO8601(createdTime)
		if !okC {
			// Si parse falló, usamos ahora
			createdStr = formatISO8601(time.Now())
		}
		modifiedStr := formatISO8601(modifiedTime)
		if !okM {
			modifiedStr = formatISO8601(time.Now())
		}

		// 3) Extraer tags
		tags := parseTags(t.Tags)

		// 4) Construir el objeto JSON mínimo
		obj := map[string]any{
			"id":       t.Title,
			"title":    t.Title,
			"created":  createdStr,
			"modified": modifiedStr,
			"tags":     tags,
			"tmap.id":  t.TmapID,
			"type":     t.Type,
			"text":     t.Text,
		}

		// 5) Si tu modelo Tiddler incluyera relaciones explícitas,
		// podrías agregarlas así (aquí se deja como nil/simplemente no se incluye):
		// obj["relations"] = map[string][]string{ ... }

		recs = append(recs, obj)
	}
	return recs
}

// ELIMINAR: La función ReverseJSONLToTiddlyJSON que estaba aquí
// Ahora está en reverse.go como archivo separado
