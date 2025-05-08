// converter_test.go – Tests unitarios para converter.go
// ------------------------------------------------------
// Pruebas de ConvertToRecord y funciones auxiliares stubs
package main

import (
	"reflect"
	"testing"

	"openpages-source/models"
)

// TestConvertToRecord verifica que ConvertToRecord mapea correctamente
// los campos básicos de models.Tiddler a models.Record usando funciones stub.
func TestConvertToRecord(t *testing.T) {
	tiddler := models.Tiddler{
		Title:    "Ejemplo de Título",
		Type:     "text/markdown",
		Tags:     []string{"tag1", "tag2"},
		Text:     "Contenido **resaltado**",
		Created:  "20250101120000000", // formato TiddlyWiki
		Modified: "20250102130000000",
	}

	rec := ConvertToRecord(tiddler)

	// Comprobación de mapeo directo
	if rec.ID != tiddler.Title {
		t.Errorf("ID = %q; se esperaba %q", rec.ID, tiddler.Title)
	}
	if rec.Title != tiddler.Title {
		t.Errorf("Title = %q; se esperaba %q", rec.Title, tiddler.Title)
	}
	if rec.ContentType != tiddler.Type {
		t.Errorf("ContentType = %q; se esperaba %q", rec.ContentType, tiddler.Type)
	}
	if !reflect.DeepEqual(rec.Tags, tiddler.Tags) {
		t.Errorf("Tags = %v; se esperaba %v", rec.Tags, tiddler.Tags)
	}
	if rec.TextMarkdown != tiddler.Text {
		t.Errorf("TextMarkdown = %q; se esperaba %q", rec.TextMarkdown, tiddler.Text)
	}
	if rec.TextPlain != tiddler.Text {
		t.Errorf("TextPlain = %q; se esperaba %q", rec.TextPlain, tiddler.Text)
	}
	if rec.CreatedAt != tiddler.Created {
		t.Errorf("CreatedAt = %q; se esperaba %q", rec.CreatedAt, tiddler.Created)
	}
	if rec.ModifiedAt != tiddler.Modified {
		t.Errorf("ModifiedAt = %q; se esperaba %q", rec.ModifiedAt, tiddler.Modified)
	}
}
