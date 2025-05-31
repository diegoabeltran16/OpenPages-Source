// internal/transform/converter.go – v1 + v2
// --------------------------------------------------------------------------------
// Este archivo ahora expone **dos** funciones públicas:
//   • ConvertTiddlers   → genera []models.Record     (esquema heredado).
//   • ConvertTiddlersV2 → genera []models.RecordV2   (nuevo esquema AI‑friendly).
// Ambas conviven para permitir que el CLI elija entre `-mode v1` y `-mode v2`.
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

// TiddlyWiki suele usar yyyymmddhhMMSSmmm o yyyymmdd.
func parseTWDate(raw string) (time.Time, bool) {
	layouts := []string{"20060102150405", "20060102"}
	for _, l := range layouts {
		if t, err := time.Parse(l, raw); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// -----------------------------------------------------------------------------
// Versión 1 – lógica intacta
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
// Versión 2 – esquema meta/content
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
		default: // text/plain y otros
			content.Plain = t.Text
		}

		rec := models.RecordV2{
			ID:      t.Title,   // se podría slugificar; se deja igual por simplicidad
			Type:    "tiddler", // valor fijo; futuros conversores pueden clasificar
			Meta:    meta,
			Content: content,
			// Relations pendiente: populate si tu Tiddler trae esa info
		}
		recs = append(recs, rec)
	}
	return recs
}
