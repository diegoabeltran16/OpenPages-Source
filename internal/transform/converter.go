// internal/transform/converter.go – v3
// --------------------------------------------------------------------------------
// Esta versión potencia el conversor con las siguientes mejoras:
//
//  1. Reconoce tiddlers de código (los que tienen Title que empieza con "-").
//  2. Detecta automáticamente el lenguaje de un archivo de código según su extensión.
//  3. Corrige el parseo de fechas que incluyen milisegundos (yyyymmddHHMMSSmmm).
//  4. Clasifica cada RecordV2 como "code" o "tiddler" para que puedas filtrarlos fácilmente.
//  5. Genera un RecordV2 (AI-friendly) que separa Meta ↔ Content, enriqueciendo la salida.
//
// Estructura general:
//   • parseTags     → Extrae las etiquetas [[tag]] de un string.
//   • parseTWDate   → Intenta convertir "yyyymmddHHMMSSmmm" o "yyyymmdd" en time.Time.
//   • detectLanguageByExtension → Dado un nombre de archivo (.go, .py, etc.), devuelve el string "go", "python"….
//   • ConvertTiddlersV3 → Toma []models.Tiddler y devuelve []models.RecordV2 siguiendo la lógica descrita.
//

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

// parseTags recibe un string como "[[foo]] [[bar baz]]" y devuelve []string{"foo", "bar baz"}
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

// -----------------------------------------------------------------------------
// 2. Parseador de fechas de TiddlyWiki, incluyendo milisegundos
// -----------------------------------------------------------------------------

// parseTWDate intenta convertir formatos "yyyymmddHHMMSSmmm", "yyyymmddHHMMSS" o "yyyymmdd"
// Devuelve time.Time y true si tuvo éxito, o time.Time{} y false en caso contrario.
func parseTWDate(raw string) (time.Time, bool) {
	layouts := []string{
		"20060102150405000", // yyyymmddHHMMSSmmm (17 dígitos)
		"20060102150405",    // yyyymmddHHMMSS  (14 dígitos)
		"20060102",          // yyyymmdd       (8 dígitos)
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, raw); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// -----------------------------------------------------------------------------
// 3. Detección de lenguaje por extensión de archivo (solo para tiddlers de código)
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
// 4. ConvertTiddlersV3 – la función principal de conversión
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
		}

		// ---------------------------------------------------------------------
		// 4.6) Construir el RecordV2 final
		// ---------------------------------------------------------------------
		rec := models.RecordV2{
			ID:        t.Title,
			Type:      recType,
			Meta:      meta,
			Content:   content,
			Relations: nil, // Por ahora no implementamos relaciones en v3
		}
		recs = append(recs, rec)
	}

	return recs
}
