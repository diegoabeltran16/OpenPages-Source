// internal/transform/converter_test.go – Tests para transform.ConvertTiddlers
// --------------------------------------------------------------------------------
// Estas pruebas viven en el **mismo paquete** (`transform`) para acceder al
// helper no exportado `parseTags`.  Verifican:
//   1. Extracción correcta de etiquetas.
//   2. Conversión completa Tiddler → Record con indentado JSON.
// --------------------------------------------------------------------------------

package transform

import (
	"reflect"
	"testing"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// write helper innecesario: los datos están embebidos como literales JSON.

// Test_parseTags comprueba la extracción de etiquetas, incluyendo espacios.
func Test_parseTags(t *testing.T) {
	raw := "[[tag1]] [[tag 2]] [[tag3]]"
	want := []string{"tag1", "tag 2", "tag3"}

	got := parseTags(raw)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseTags(%q) = %v, want %v", raw, got, want)
	}
}

// TestConvertTiddlers verifica el flujo integral, incluyendo el pretty-print de JSON embebido.
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
