// internal/transform/converter_test.go
// --------------------------------------------------------------------------------
// Aquí actualizamos los tests para que ya no invoquen ConvertTiddlers (v1) ni
// ConvertTiddlersV2, sino únicamente la nueva función ConvertTiddlersV3. Además,
// agregamos un test para parseTWDate (para asegurarnos de que parsea fechas con
// milisegundos) y revisamos parseTags como antes.
//
// Estructura del archivo:
//
//   1. import de paquetes necesarios
//   2. Test_parseTags (sin cambios)
//   3. Test_parseTWDate (nuevo test para fechas con y sin milisegundos)
//   4. TestConvertTiddlersV3 (test principal para la lógica v3)
// --------------------------------------------------------------------------------

package transform

import (
	"reflect"
	"testing"
	"time"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// -----------------------------------------------------------------------------
// 1. Test para parseTags
// -----------------------------------------------------------------------------

// Test_parseTags verifica que parseTags extrae correctamente todas las etiquetas
// del formato [[tag]]. Tanto etiquetas con espacios internos como sin espacios.
func Test_parseTags(t *testing.T) {
	raw := "[[foo]] [[bar baz]] [[123]]"
	want := []string{"foo", "bar baz", "123"}

	got := parseTags(raw)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseTags(%q) = %v; queríamos %v", raw, got, want)
	}
}

// -----------------------------------------------------------------------------
// 2. Test para parseTWDate
// -----------------------------------------------------------------------------

// Test_parseTWDate valida que parseTWDate convierta correctamente strings de fecha
// en formatos yyyymmddHHMMSSmmm (17 dígitos), yyyymmddHHMMSS (14 dígitos) y yyyymmdd (8 dígitos).
func Test_parseTWDate(t *testing.T) {
	tests := []struct {
		raw     string
		layout  string
		wantStr string // la representación deseada en RFC3339 para compararlo fácilmente
	}{
		{
			raw:     "20251231115959123",        // formato yyyymmddHHMMSSmmm
			layout:  "2006-01-02T15:04:05.000Z", // con milisegundos
			wantStr: "2025-12-31T11:59:59.123Z",
		},
		{
			raw:     "20251231115959",       // formato yyyymmddHHMMSS
			layout:  "2006-01-02T15:04:05Z", // sin milisegundos
			wantStr: "2025-12-31T11:59:59Z",
		},
		{
			raw:     "20251231",             // formato yyyymmdd
			layout:  "2006-01-02T00:00:00Z", // solo fecha, hora = 00:00:00
			wantStr: "2025-12-31T00:00:00Z",
		},
		{
			raw:     "invalid", // no se puede parsear
			layout:  "",        // no importa
			wantStr: "",
		},
	}

	for _, tc := range tests {
		gotTime, ok := parseTWDate(tc.raw)
		if tc.wantStr == "" {
			// En caso inválido, ok debe ser false y gotTime == zero
			if ok || !gotTime.IsZero() {
				t.Errorf("parseTWDate(%q) devolvió (%v, %v); queríamos (zero, false)", tc.raw, gotTime, ok)
			}
			continue
		}
		if !ok {
			t.Errorf("parseTWDate(%q) devolvió ok=false, queríamos true", tc.raw)
			continue
		}
		// Convertir gotTime a RFC3339 (UTC) para comparar cadenas
		gotStr := gotTime.UTC().Format(tc.layout)
		if gotStr != tc.wantStr {
			t.Errorf("parseTWDate(%q) = %q; queríamos %q", tc.raw, gotStr, tc.wantStr)
		}
	}
}

// -----------------------------------------------------------------------------
// 3. Test para ConvertTiddlersV3
// -----------------------------------------------------------------------------

// TestConvertTiddlersV3 valida casos mixtos: tiddler conceptual (Markdown) y tiddler
// de código (prefijo "-" y archivo .go con milisegundos en la fecha).
func TestConvertTiddlersV3(t *testing.T) {
	// 3.1) Construimos dos tiddlers de ejemplo:
	//      - t1: tiddler conceptual en Markdown
	//      - t2: tiddler de código Go
	t1 := models.Tiddler{
		Title:    "ExampleConcept.md",
		Text:     "# Título\nContenido en **Markdown**.",
		Tags:     "[[alpha]] [[beta gamma]]",
		Created:  "20250101",        // yyyymmdd (solo fecha)
		Modified: "20250102",        // yyyymmdd
		Type:     "text/x-markdown", // forzamos Markdown
		Color:    "#abcdef",
		TmapID:   "tmap-concept-123",
	}
	t2 := models.Tiddler{
		Title:    "-foo_main.go",
		Text:     "package main\n\nfunc main() {}",
		Tags:     "[[code]]",
		Created:  "20251231115959123", // yyyymmddHHMMSSmmm (17 dígitos)
		Modified: "20251231120000123", // igual con milisegundos
		Type:     "text/plain",        // aunque es código .go, el import lo pone como plain
		Color:    "#123456",
		TmapID:   "tmap-code-456",
	}

	// 3.2) Invocamos el conversor V3
	got := ConvertTiddlersV3([]models.Tiddler{t1, t2})

	// Debemos obtener 2 registros
	if len(got) != 2 {
		t.Fatalf("ConvertTiddlersV3 produjo %d registros; queríamos 2", len(got))
	}

	// 3.3) Verificar el primer registro (t1, conceptual)
	r1 := got[0]
	if r1.Type != "tiddler" {
		t.Errorf("r1.Type = %q; queríamos \"tiddler\"", r1.Type)
	}
	if r1.ID != "ExampleConcept.md" {
		t.Errorf("r1.ID = %q; queríamos \"ExampleConcept.md\"", r1.ID)
	}
	// Meta:
	if r1.Meta.Title != "ExampleConcept.md" {
		t.Errorf("r1.Meta.Title = %q; queríamos \"ExampleConcept.md\"", r1.Meta.Title)
	}
	// Fecha sin milisegundos: parseTWDate("20250101") debe dar "2025-01-01T00:00:00Z"
	wantCreated1, _ := time.Parse("2006-01-02T15:04:05Z", "2025-01-01T00:00:00Z")
	if !r1.Meta.Created.Equal(wantCreated1) {
		t.Errorf("r1.Meta.Created = %v; queríamos %v", r1.Meta.Created, wantCreated1)
	}
	wantModified1, _ := time.Parse("2006-01-02T15:04:05Z", "2025-01-02T00:00:00Z")
	if !r1.Meta.Modified.Equal(wantModified1) {
		t.Errorf("r1.Meta.Modified = %v; queríamos %v", r1.Meta.Modified, wantModified1)
	}
	// Tags:
	expectedTags1 := []string{"alpha", "beta gamma"}
	if !reflect.DeepEqual(r1.Meta.Tags, expectedTags1) {
		t.Errorf("r1.Meta.Tags = %v; queríamos %v", r1.Meta.Tags, expectedTags1)
	}
	// Color y TmapID:
	if r1.Meta.Color != "#abcdef" {
		t.Errorf("r1.Meta.Color = %q; queríamos \"#abcdef\"", r1.Meta.Color)
	}
	if r1.Meta.Extra["tmap.id"] != "tmap-concept-123" {
		t.Errorf("r1.Meta.Extra[\"tmap.id\"] = %q; queríamos \"tmap-concept-123\"", r1.Meta.Extra["tmap.id"])
	}
	// Content: debería estar en Markdown
	if r1.Content.Markdown != "# Título\nContenido en **Markdown**." {
		t.Errorf("r1.Content.Markdown = %q; queríamos el texto completo en Markdown", r1.Content.Markdown)
	}
	// Aseguramos que Plain y Secciones estén vacíos
	if r1.Content.Plain != "" {
		t.Errorf("r1.Content.Plain = %q; queríamos cadena vacía para tiddler conceptual", r1.Content.Plain)
	}
	if len(r1.Content.Sections) != 0 {
		t.Errorf("r1.Content.Sections = %v; queríamos ninguna sección para tiddler conceptual", r1.Content.Sections)
	}

	// 3.4) Verificar el segundo registro (t2, código Go)
	r2 := got[1]
	if r2.Type != "code" {
		t.Errorf("r2.Type = %q; queríamos \"code\"", r2.Type)
	}
	if r2.ID != "-foo_main.go" {
		t.Errorf("r2.ID = %q; queríamos \"-foo_main.go\"", r2.ID)
	}
	// Meta:
	if r2.Meta.Title != "-foo_main.go" {
		t.Errorf("r2.Meta.Title = %q; queríamos \"-foo_main.go\"", r2.Meta.Title)
	}
	// Fecha con milisegundos: parseTWDate("20251231115959123")
	wantCreated2, _ := time.Parse("2006-01-02T15:04:05.000Z", "2025-12-31T11:59:59.123Z")
	if !r2.Meta.Created.Equal(wantCreated2) {
		t.Errorf("r2.Meta.Created = %v; queríamos %v", r2.Meta.Created, wantCreated2)
	}
	wantModified2, _ := time.Parse("2006-01-02T15:04:05.000Z", "2025-12-31T12:00:00.123Z")
	if !r2.Meta.Modified.Equal(wantModified2) {
		t.Errorf("r2.Meta.Modified = %v; queríamos %v", r2.Meta.Modified, wantModified2)
	}
	// Tags:
	expectedTags2 := []string{"code"}
	if !reflect.DeepEqual(r2.Meta.Tags, expectedTags2) {
		t.Errorf("r2.Meta.Tags = %v; queríamos %v", r2.Meta.Tags, expectedTags2)
	}
	// Color y TmapID:
	if r2.Meta.Color != "#123456" {
		t.Errorf("r2.Meta.Color = %q; queríamos \"#123456\"", r2.Meta.Color)
	}
	if r2.Meta.Extra["tmap.id"] != "tmap-code-456" {
		t.Errorf("r2.Meta.Extra[\"tmap.id\"] = %q; queríamos \"tmap-code-456\"", r2.Meta.Extra["tmap.id"])
	}
	// Content: Plain debe tener el código completo
	if r2.Content.Plain != "package main\n\nfunc main() {}" {
		t.Errorf("r2.Content.Plain = %q; queríamos \"package main\\n\\nfunc main() {}\"", r2.Content.Plain)
	}
	// Debe haber exactamente una Section indicando el lenguaje "go"
	if len(r2.Content.Sections) != 1 {
		t.Fatalf("r2.Content.Sections tiene longitud %d; queríamos 1", len(r2.Content.Sections))
	}
	sec := r2.Content.Sections[0]
	if sec.Name != "language" || sec.RawValue != "go" {
		t.Errorf("r2.Content.Sections[0] = %v; queríamos {Name:\"language\", RawValue:\"go\"}", sec)
	}
	// El campo Markdown y JSON deben estar vacíos
	if r2.Content.Markdown != "" {
		t.Errorf("r2.Content.Markdown = %q; queríamos cadena vacía para tiddler de código", r2.Content.Markdown)
	}
	if r2.Content.JSON != nil {
		t.Errorf("r2.Content.JSON = %v; queríamos nil para tiddler de código", r2.Content.JSON)
	}
}
