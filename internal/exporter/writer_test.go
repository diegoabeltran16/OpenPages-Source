// internal/exporter/writer_test.go – Tests para exporter.WriteJSONL
// --------------------------------------------------------------------------------
// Comprueba dos rutas:
//   1. Éxito: archivo temporal + 2 records → 2 líneas JSONL idénticas.
//   2. Falla: ruta imposible debe devolver error.
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

// helper para generar un filepath temporal (cerrado) que luego se re‑abre.
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

func TestWriteJSONL_Success(t *testing.T) {
	recs := []models.Record{
		{ID: "One", Tags: []string{"a", "b"}, ContentType: "text/plain", TextMarkdown: "foo", TextPlain: "foo"},
		{ID: "Two", Tags: []string{"x"}, ContentType: "application/json", TextMarkdown: "{\"k\":1}", TextPlain: "{\"k\":1}"},
	}
	path := tmpPath(t)
	defer os.Remove(path)

	if err := WriteJSONL(context.Background(), path, recs); err != nil {
		t.Fatalf("WriteJSONL err: %v", err)
	}

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != len(recs) {
		t.Fatalf("líneas = %d, want %d", len(lines), len(recs))
	}
	for i, l := range lines {
		var got models.Record
		if err := json.Unmarshal([]byte(l), &got); err != nil {
			t.Fatalf("unmarshal línea %d: %v", i, err)
		}
		if !reflect.DeepEqual(got, recs[i]) {
			t.Errorf("línea %d mismatch: got %+v, want %+v", i, got, recs[i])
		}
	}
}

func TestWriteJSONL_InvalidPath(t *testing.T) {
	err := WriteJSONL(context.Background(), "/no/existe/out.jsonl", nil)
	if err == nil {
		t.Fatalf("esperaba error en ruta inválida")
	}
}
