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
func parseTags(raw any) []string {
	switch v := raw.(type) {
	case string:
		matches := tagRe.FindAllStringSubmatch(v, -1)
		tags := make([]string, 0, len(matches))
		for _, m := range matches {
			if len(m) > 1 {
				tags = append(tags, m[1])
			}
		}
		return tags
	case []interface{}:
		tags := make([]string, 0, len(v))
		for _, tag := range v {
			if s, ok := tag.(string); ok {
				tags = append(tags, s)
			}
		}
		return tags
	case []string:
		return v
	default:
		return nil
	}
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

// orEmpty devuelve el valor o "" si está vacío
func orEmpty(s string) string {
	if s == "" {
		return ""
	}
	return s
}

// -----------------------------------------------------------------------------
// Versión 1 – lógica intacta (esquema heredado)
// -----------------------------------------------------------------------------

func ConvertTiddlers(ts []models.Tiddler) []models.Record {
	recs := make([]models.Record, 0, len(ts))

	for _, t := range ts {
		// --- Extracción robusta de campos secundarios ---
		created := t.Created
		modified := t.Modified
		color := t.Color

		// Buscar en Meta si están vacíos
		if t.Meta != nil {
			if created == "" {
				created = t.Meta.Created
			}
			if color == "" {
				color = t.Meta.Color
			}
		}

		rec := models.Record{
			ID:          t.Title,
			Tags:        parseTags(t.Tags),
			ContentType: t.Type,
			CreatedAt:   created,
			ModifiedAt:  modified,
			Color:       color,
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
		// --- Extracción robusta de campos secundarios ---
		created := t.Created
		modified := t.Modified
		color := t.Color
		tmapid := t.TmapID

		// Buscar en Meta si están vacíos
		if t.Meta != nil {
			if created == "" {
				created = t.Meta.Created
			}
			if modified == "" {
				modified = t.Meta.Modified
			}
			if color == "" {
				color = t.Meta.Color
			}
			// Buscar en Meta.Extra
			if t.Meta.Extra != nil {
				if tmapid == "" {
					tmapid = t.Meta.Extra["tmap.id"]
				}
				if color == "" {
					color = t.Meta.Extra["color"]
				}
			}
		}

		// Meta
		createdTime, _ := parseTWDate(created)
		modifiedTime, _ := parseTWDate(modified)

		meta := models.RecordMeta{
			Title:    t.Title,
			Tags:     parseTags(t.Tags),
			Created:  createdTime,
			Modified: modifiedTime,
			Color:    color,
			Extra: map[string]string{
				"tmap.id": tmapid,
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

		// 3) Extraer tags y tags_list
		tags := parseTags(t.Tags)
		var tagsList []string
		if t.TagsList != nil {
			tagsList = t.TagsList
		} else {
			tagsList = []string{}
		}

		// --- Extracción robusta de campos secundarios ---
		created := t.Created
		modified := t.Modified
		color := t.Color
		tmapid := t.TmapID
		path := t.Path

		// Buscar en Meta si están vacíos
		if t.Meta != nil {
			if created == "" {
				created = t.Meta.Created
			}
			if modified == "" {
				modified = t.Meta.Modified
			}
			if color == "" {
				color = t.Meta.Color
			}
			// Buscar en Meta.Extra
			if t.Meta.Extra != nil {
				if tmapid == "" {
					tmapid = t.Meta.Extra["tmap.id"]
				}
				if color == "" {
					color = t.Meta.Extra["color"]
				}
				if path == "" {
					path = t.Meta.Extra["path"]
				}
			}
		}

		// 4) Construir el objeto JSON mínimo y robusto
		obj := map[string]any{
			"id":           t.Title,
			"title":        t.Title,
			"created":      orEmpty(created),
			"created_rfc":  createdStr,
			"modified":     orEmpty(modified),
			"modified_rfc": modifiedStr,
			"tags":         tags,
			"tags_list":    tagsList,
			"tmap.id":      orEmpty(tmapid),
			"type":         orEmpty(t.Type),
			"text":         GetTextContent(t.Text),
			"color":        orEmpty(color),
			"path":         orEmpty(path),
		}

		// 5) Relaciones explícitas si aplica
		if t.Relations != nil {
			obj["relations"] = t.Relations
		} else {
			obj["relations"] = map[string]any{}
		}

		recs = append(recs, obj)
	}
	return recs
}

// ConvertTiddlersHybrid genera un slice de objetos planos ideales para IA/RAG.
func ConvertTiddlersHybrid(ts []models.Tiddler) []models.Record {
	recs := make([]models.Record, 0, len(ts))
	for _, t := range ts {
		// --- Extracción robusta de campos secundarios ---
		created := t.Created
		modified := t.Modified
		color := t.Color

		// Buscar en Meta si están vacíos
		if t.Meta != nil {
			if created == "" {
				created = t.Meta.Created
			}
			if modified == "" {
				modified = t.Meta.Modified
			}
			if color == "" {
				color = t.Meta.Color
			}
		}

		rec := models.Record{
			ID:           t.Title,
			Tags:         parseTags(t.Tags),
			ContentType:  t.Type,
			TextMarkdown: GetTextContent(t.Text),
			TextPlain:    GetTextContent(t.Text),
			CreatedAt:    created,
			ModifiedAt:   modified,
			Color:        color,
		}
		recs = append(recs, rec)
	}
	return recs
}

// GetTextContent extrae el texto del contenido, manejando tanto JSON como texto plano
func GetTextContent(text string) string {
	if len(text) > 0 && text[0] == '{' && text[len(text)-1] == '}' {
		var w map[string]any
		if err := json.Unmarshal([]byte(text), &w); err == nil {
			if c, ok := w["content"].(map[string]any); ok {
				if plain, ok := c["plain"].(string); ok && plain != "" {
					return plain
				}
				if markdown, ok := c["markdown"].(string); ok && markdown != "" {
					return markdown
				}
			}
		}
	}
	return text
}
