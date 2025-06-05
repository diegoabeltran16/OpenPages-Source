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
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// -----------------------------------------------------------------------------
// 1. Extracción de etiquetas en formato [[tag]]
// -----------------------------------------------------------------------------

// tagRe compila una expresión regular que matchea [[cualquier_texto]]
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
	layouts := []string{
		"20060102150405000", // yyyymmddHHMMSSmmm (17 dígitos)
		"20060102150405",    // yyyymmddHHMMSS  (14 dígitos)
		"20060102",          // yyyymmdd       (8 dígitos)
	}
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

// detectLanguageByExtension recibe un nombre de archivo (p.ej. "foo.go")
// y devuelve un string como "go", "python", "bash", etc. Si no reconoce la extensión,
// devuelve cadena vacía.
func detectLanguageByExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".sh", ".bash":
		return "bash"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".java":
		return "java"
	case ".rs":
		return "rust"
	case ".rb":
		return "ruby"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md":
		return "markdown"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	default:
		return ""
	}
}

// -----------------------------------------------------------------------------
// Versión 2 – esquema meta/content (AI-friendly)
// -----------------------------------------------------------------------------

// ConvertTiddlersV3 recorre cada models.Tiddler y genera un models.RecordV2,
// separando Meta ↔ Content y clasificando el tipo de tiddler ("code" o "tiddler").
//
// Parámetros:
//   - ts: slice de models.Tiddler (cada tiddler viene con Title, Text, Tags, Created, Modified, Type, Color, TmapID).
//
// Retorno:
//   - []models.RecordV2: slice de registros listos para exportar en JSONL.
//     Cada RecordV2 tiene:
//   - ID:       mismo que Title (se podría slugificar si se desea).
//   - Type:     "code"  si Title empieza con "-", o
//     "tiddler" en cualquier otro caso.
//   - Meta:     { Title, Tags, Created(time.Time), Modified(time.Time), Color, Extra{"tmap.id": TmapID} }.
//   - Content:  Si Type="code": Plain = Text completo, y
//     Sections = [{"name":"language", "value":"<lenguaje>"}].
//     Si Type="tiddler": según t.Type (application/json → JSON;
//     text/markdown → Markdown;
//     otro → Plain).
//   - Relations: nil (en esta versión no poblamos relaciones).
func ConvertTiddlersV3(ts []models.Tiddler) []models.RecordV2 {
	recs := make([]models.RecordV2, 0, len(ts))

	for _, t := range ts {
		// ---------------------------------------------------------------------
		// 4.1) Parsear las fechas "Created" y "Modified"
		//     (pueden venir en formato yyyymmddHHMMSSmmm, yyyymmddHHMMSS o yyyymmdd)
		// ---------------------------------------------------------------------
		createdTime, okC := parseTWDate(t.Created)
		if !okC {
			createdTime = time.Time{} // zero value si no pudo parsear
		}
		modifiedTime, okM := parseTWDate(t.Modified)
		if !okM {
			modifiedTime = time.Time{}
		}

		// ---------------------------------------------------------------------
		// 4.2) Extraer las etiquetas del string t.Tags
		// ---------------------------------------------------------------------
		parsedTags := parseTags(t.Tags)

		// ---------------------------------------------------------------------
		// 4.3) Determinar si este tiddler es "código"
		//      Se asume que TODO archivo de código en el proyecto lleva “-” como primer carácter.
		// ---------------------------------------------------------------------
		isCode := strings.HasPrefix(t.Title, "-")
		recType := "tiddler"
		if isCode {
			recType = "code"
		}

		// ---------------------------------------------------------------------
		// 4.4) Construir el objeto Meta
		// ---------------------------------------------------------------------
		meta := models.Meta{
			Title:    t.Title,
			Tags:     parsedTags,
			Created:  createdTime,
			Modified: modifiedTime,
			Color:    t.Color,
			Extra: map[string]string{
				"tmap.id": t.TmapID,
			},
		}

		// ---------------------------------------------------------------------
		// 4.5) Construir el objeto Content
		//      - Si es código: lo guardamos en Plain, y agregamos sección "language".
		//      - Si es tiddler conceptual: separamos JSON, Markdown o Plain según t.Type.
		// ---------------------------------------------------------------------
		var content models.Content

		if isCode {
			// --- Tiddler de código: guardamos texto completo en Plain
			content.Plain = t.Text

			//    Detectar idioma (por extensión) y almacenar en Sections
			lang := detectLanguageByExtension(t.Title)
			if lang != "" {
				content.Sections = []models.Section{{
					Name:     "language",
					RawValue: lang,
				}}
			}
		} else {
			// --- Tiddler conceptual: tomamos en cuenta t.Type
			switch t.Type {
			case "application/json":
				var obj map[string]any
				if err := json.Unmarshal([]byte(t.Text), &obj); err == nil {
					content.JSON = obj
				} else {
					// Si no se pudo parsear JSON, lo dejamos en Plain
					content.Plain = t.Text
				}
			case "text/x-markdown", "text/markdown":
				content.Markdown = t.Text
			default:
				// Cualquier otro tipo (text/plain, text/markdown u otro),
				// lo dejamos en Plain sin más
				content.Plain = t.Text
			}
		case "text/x-markdown":
			content.Markdown = t.Text
		default:
			content.Plain = t.Text
		}

		// ---------------------------------------------------------------------
		// 4.6) Construir el RecordV2 final
		// ---------------------------------------------------------------------
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
