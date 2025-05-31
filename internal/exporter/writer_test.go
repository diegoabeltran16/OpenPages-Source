// internal/exporter/writer_test.go – Tests para exporter.WriteJSONL
// --------------------------------------------------------------------------------
// Comprueba dos rutas:
//   1. Éxito: archivo temporal + 2 records → 2 líneas JSONL idénticas.
//   2. Falla: ruta imposible debe devolver error.
// --------------------------------------------------------------------------------
//
// --------------------------------------------------------------------------------
// Verifica que la función revisada (`records any, pretty bool`) maneje:
//   • Slice v1   → []models.Record.
//   • Slice v2   → []models.RecordV2.
//   • Ruta inválida → error.
// --------------------------------------------------------------------------------

package exporter

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// tmpPath genera un archivo temporal y devuelve la ruta.
func tmpPath(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "out-*.jsonl")
	if err != nil {
		t.Fatalf("tmpPath: %v", err)
	}
	name := f.Name()
	f.Close()
	return name
}

// ----------------------------- caso éxito v1 -----------------------------
func TestWriteJSONL_V1(t *testing.T) {
	recs := []models.Record{
		{ID: "A", Tags: []string{"x"}, ContentType: "text/plain", TextPlain: "foo"},
		{ID: "B", Tags: []string{"y"}, ContentType: "text/plain", TextPlain: "bar"},
	}
	path := tmpPath(t)
	defer os.Remove(path)

	if err := WriteJSONL(context.Background(), path, recs, false); err != nil {
		t.Fatalf("WriteJSONL v1 err: %v", err)
	}

	verifyLines(t, path, recs)
}

// ----------------------------- caso éxito v2 -----------------------------
func TestWriteJSONL_V2(t *testing.T) {
	recs := []models.RecordV2{
		{ID: "1", Type: "tiddler", Meta: models.Meta{Title: "Hello"}, Content: models.Content{Plain: "hola"}},
		{ID: "2", Type: "tiddler", Meta: models.Meta{Title: "World"}, Content: models.Content{Plain: "mundo"}},
	}
	path := tmpPath(t)
	defer os.Remove(path)

	if err := WriteJSONL(context.Background(), path, recs, false); err != nil {
		t.Fatalf("WriteJSONL v2 err: %v", err)
	}

	verifyLines(t, path, recs)
}

// ----------------------------- ruta inválida -----------------------------
func TestWriteJSONL_InvalidPath(t *testing.T) {
	err := WriteJSONL(context.Background(), "/no/existe/out.jsonl", nil, false)
	if err == nil {
		t.Fatalf("esperaba error en ruta inválida")
	}
}

// ----------------------------- helper de verificación --------------------
func verifyLines(t *testing.T, path string, wantSlice any) {
	t.Helper()
	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	v := reflect.ValueOf(wantSlice)
	if len(lines) != v.Len() {
		t.Fatalf("líneas = %d, want %d", len(lines), v.Len())
	}
	for i, l := range lines {
		gotPtr := reflect.New(v.Type().Elem()) // *T
		if err := json.Unmarshal([]byte(l), gotPtr.Interface()); err != nil {
			t.Fatalf("unmarshal línea %d: %v", i, err)
		}
		if !reflect.DeepEqual(gotPtr.Elem().Interface(), v.Index(i).Interface()) {
			t.Errorf("línea %d mismatch\n got:  %+v\n want: %+v", i, gotPtr.Elem(), v.Index(i))
		}
	}
}
