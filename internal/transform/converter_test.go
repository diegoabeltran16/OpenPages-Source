// internal/transform/converter_test.go – Tests para transform.ConvertTiddlers, ConvertTiddlersV2 y ConvertTiddlersV3
// --------------------------------------------------------------------------------
// Estas pruebas viven en el **mismo paquete** (`transform`) para acceder a parseTags y parseTWDate.
// Verifican:
//   1. Extracción correcta de etiquetas con parseTags.
//   2. ConvertTiddlers (v1): Tiddler → models.Record.
//   3. ConvertTiddlersV2 (v2): Tiddler → models.RecordV2, campos meta/content.
//   4. ConvertTiddlersV3 (v3): Tiddler → map[string]any minimalista para JSONL.
// --------------------------------------------------------------------------------

package transform

import (
	"reflect"
	"testing"
	"time"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// ----------------------------- parseTags -----------------------------
// Test_parseTags comprueba la extracción de etiquetas, incluyendo espacios.
func Test_parseTags(t *testing.T) {
	raw := "[[tag1]] [[tag 2]] [[tag3]]"
	want := []string{"tag1", "tag 2", "tag3"}

	got := parseTags(raw)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseTags(%q) = %v, want %v", raw, got, want)
	}
}

// ----------------------------- ConvertTiddlers (v1) -----------------------------
func TestConvertTiddlers(t *testing.T) {
	tiddlers := []models.Tiddler{
		{
			Title:    "Foo",
			Text:     "plain text",
			Tags:     "[[a]] [[b]]",
			Created:  "20250101",
			Modified: "20250102",
			Type:     "text/plain",
		},
		{
			Title:    "Bar",
			Text:     "{\"key\":\"value\"}",
			Tags:     "[[x]]",
			Created:  "20250103",
			Modified: "20250104",
			Type:     "application/json",
		},
	}

	want := []models.Record{
		{
			ID:           "Foo",
			Tags:         []string{"a", "b"},
			ContentType:  "text/plain",
			TextMarkdown: "plain text",
			TextPlain:    "plain text",
			CreatedAt:    "20250101",
			ModifiedAt:   "20250102",
		},
		{
			ID:           "Bar",
			Tags:         []string{"x"},
			ContentType:  "application/json",
			TextMarkdown: "{\n  \"key\": \"value\"\n}",
			TextPlain:    "{\n  \"key\": \"value\"\n}",
			CreatedAt:    "20250103",
			ModifiedAt:   "20250104",
		},
	}

	got := ConvertTiddlers(tiddlers)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ConvertTiddlers() = %+v, want %+v", got, want)
	}
}

// ----------------------------- ConvertTiddlersV2 (v2) -----------------------------
func TestConvertTiddlersV2(t *testing.T) {
	tiddlers := []models.Tiddler{
		{
			Title:    "Alpha",
			Text:     "some markdown",
			Tags:     "[[one]] [[two]]",
			Created:  "20250101",
			Modified: "20250102",
			Type:     "text/x-markdown",
			Color:    "#ff0000",
			TmapID:   "uuid-alpha",
		},
		{
			Title:    "Beta",
			Text:     "{\"field\":123}",
			Tags:     "[[X]]",
			Created:  "20250203",
			Modified: "20250204",
			Type:     "application/json",
			Color:    "#00ff00",
			TmapID:   "uuid-beta",
		},
	}

	recs := ConvertTiddlersV2(tiddlers)
	if len(recs) != 2 {
		t.Fatalf("ConvertTiddlersV2: longitud = %d, want 2", len(recs))
	}

	// Verificar campos meta y content del primer record
	m0 := recs[0].Meta
	if m0.Title != "Alpha" {
		t.Errorf("v2 Meta.Title = %q, want %q", m0.Title, "Alpha")
	}
	// parseTWDate("20250101") → 2025-01-01 00:00:00 UTC
	expectedCreated0 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !m0.Created.Equal(expectedCreated0) {
		t.Errorf("v2 Meta.Created = %v, want %v", m0.Created, expectedCreated0)
	}
	if m0.Color != "#ff0000" {
		t.Errorf("v2 Meta.Color = %q, want %q", m0.Color, "#ff0000")
	}
	if extraID, ok := m0.Extra["tmap.id"]; !ok || extraID != "uuid-alpha" {
		t.Errorf("v2 Meta.Extra[\"tmap.id\"] = %q, want %q", extraID, "uuid-alpha")
	}
	// Tags
	wantTags0 := []string{"one", "two"}
	if !reflect.DeepEqual(recs[0].Meta.Tags, wantTags0) {
		t.Errorf("v2 Meta.Tags = %v, want %v", recs[0].Meta.Tags, wantTags0)
	}
	// Content Markdown
	if recs[0].Content.Markdown != "some markdown" {
		t.Errorf("v2 Content.Markdown = %q, want %q", recs[0].Content.Markdown, "some markdown")
	}

	// Segundo record: JSON deserializado
	c1 := recs[1].Content.JSON
	if val, ok := c1["field"]; !ok || val.(float64) != 123 {
		t.Errorf("v2 Content.JSON[\"field\"] = %v, want 123", val)
	}
}

// ----------------------------- ConvertTiddlersV3 (v3) -----------------------------
func TestConvertTiddlersV3(t *testing.T) {
	tiddlers := []models.Tiddler{
		{
			Title:    "Gamma",
			Text:     "hello gamma",
			Tags:     "[[g1]] [[g2]]",
			Created:  "20250301",
			Modified: "20250302",
			Type:     "text/plain",
			Color:    "#0000ff",
			TmapID:   "uuid-gamma",
		},
		// Sin campos Created/Modified parseables (ej. "invalid"), para probar fallback
		{
			Title:    "Delta",
			Text:     "hello delta",
			Tags:     "[[d1]]",
			Created:  "invalid",
			Modified: "invalid",
			Type:     "text/plain",
			Color:    "#ffffff",
			TmapID:   "uuid-delta",
		},
	}

	recs := ConvertTiddlersV3(tiddlers)
	if len(recs) != 2 {
		t.Fatalf("ConvertTiddlersV3: longitud = %d, want 2", len(recs))
	}

	// Primer objeto: fechas parseadas en RFC3339
	obj0 := recs[0]
	// created "20250301" → 2025-03-01T00:00:00Z (UTC) formateado con offset
	parsedCreated0, _ := parseTWDate("20250301")
	want0 := parsedCreated0.Format("2006-01-02T15:04:05-07:00")
	if obj0["created"] != want0 {
		t.Errorf("v3 obj0[\"created\"] = %q, want %q", obj0["created"], want0)
	}
	// fields básicos
	if obj0["id"] != "Gamma" || obj0["title"] != "Gamma" {
		t.Errorf("v3 obj0 id/title = %q/%q, want %q/%q", obj0["id"], obj0["title"], "Gamma", "Gamma")
	}
	// Tags
	gotTags0, ok0 := obj0["tags"].([]string)
	wantTags0 := []string{"g1", "g2"}
	if !ok0 || !reflect.DeepEqual(gotTags0, wantTags0) {
		t.Errorf("v3 obj0[\"tags\"] = %v, want %v", gotTags0, wantTags0)
	}
	// TmapID y text
	if obj0["tmap.id"] != "uuid-gamma" {
		t.Errorf("v3 obj0[\"tmap.id\"] = %q, want %q", obj0["tmap.id"], "uuid-gamma")
	}
	if obj0["text"] != "hello gamma" {
		t.Errorf("v3 obj0[\"text\"] = %q, want %q", obj0["text"], "hello gamma")
	}

	// Segundo objeto: fechas “invalid” deben haber caído en Now()
	obj1 := recs[1]
	// created/modified no deben ser “0001-01-01...”; deben formatearse con Now()
	created1 := obj1["created"].(string)
	modified1 := obj1["modified"].(string)
	if created1 == "" || created1 == "0001-01-01T00:00:00Z" {
		t.Errorf("v3 obj1[\"created\"] no debería ser vacío ni valor cero, got %q", created1)
	}
	if modified1 == "" || modified1 == "0001-01-01T00:00:00Z" {
		t.Errorf("v3 obj1[\"modified\"] no debería ser vacío ni valor cero, got %q", modified1)
	}

	// Campos básicos para Delta
	if obj1["id"] != "Delta" || obj1["title"] != "Delta" {
		t.Errorf("v3 obj1 id/title = %q/%q, want %q/%q", obj1["id"], obj1["title"], "Delta", "Delta")
	}
	gotTags1, ok1 := obj1["tags"].([]string)
	wantTags1 := []string{"d1"}
	if !ok1 || !reflect.DeepEqual(gotTags1, wantTags1) {
		t.Errorf("v3 obj1[\"tags\"] = %v, want %v", gotTags1, wantTags1)
	}
	if obj1["tmap.id"] != "uuid-delta" {
		t.Errorf("v3 obj1[\"tmap.id\"] = %q, want %q", obj1["tmap.id"], "uuid-delta")
	}
}

// ----------------------------- parseTWDate fallback -----------------------------
// Asegura que parseTWDate retorne (Time{}, false) si el formato no coincide.
func Test_parseTWDate_Invalid(t *testing.T) {
	if _, ok := parseTWDate("notadate"); ok {
		t.Errorf("parseTWDate(\"notadate\") devolvió ok=true, se esperaba false")
	}
}
