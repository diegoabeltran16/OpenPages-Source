// internal/exporter/writer_test.go – Tests para exporter.WriteJSONL
// --------------------------------------------------------------------------------
// Comprueba los siguientes escenarios:
//   1. Éxito con slice de models.Record (v1).
//   2. Éxito con slice de models.RecordV2 (v2).
//   3. Éxito con slice de map[string]any (v3).
//   4. Ruta inválida devuelve error.
//   5. Creación automática de directorios anidados.
//   6. Error si el argumento no es un slice.
//   7. Diferencia entre modo ‘pretty’ (indentado) y ‘compacto’ (una sola línea).
//   8. Error de marshal al serializar tipos no serializables.
//
// Para ejecutar:
//   cd internal/exporter
//   go test
// --------------------------------------------------------------------------------

package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// tmpPath genera un archivo temporal y devuelve la ruta.
// Utilizado para crear archivos temporales en los tests.
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
		{ID: "1", Type: "tiddler", Meta: models.RecordMeta{Title: "Hello"}, Content: models.Content{Plain: "hola"}},
		{ID: "2", Type: "tiddler", Meta: models.RecordMeta{Title: "World"}, Content: models.Content{Plain: "mundo"}},
	}
	path := tmpPath(t)
	defer os.Remove(path)

	if err := WriteJSONL(context.Background(), path, recs, false); err != nil {
		t.Fatalf("WriteJSONL v2 err: %v", err)
	}

	verifyLines(t, path, recs)
}

// ----------------------------- caso éxito v3 -----------------------------
func TestWriteJSONL_V3(t *testing.T) {
	recs := []map[string]any{
		{"id": "X", "title": "X", "created": "2025-06-05T15:10:00-05:00", "modified": "2025-06-05T15:10:00-05:00", "tags": []string{"a"}, "tmap.id": "uuid1", "type": "text/plain", "text": "textoX"},
		{"id": "Y", "title": "Y", "created": "2025-06-05T15:20:00-05:00", "modified": "2025-06-05T15:20:00-05:00", "tags": []string{"b"}, "tmap.id": "uuid2", "type": "text/plain", "text": "textoY"},
	}
	path := tmpPath(t)
	defer os.Remove(path)

	if err := WriteJSONL(context.Background(), path, recs, false); err != nil {
		t.Fatalf("WriteJSONL v3 err: %v", err)
	}

	// Verifica cada línea deserializada
	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != len(recs) {
		t.Fatalf("líneas = %d, want %d", len(lines), len(recs))
	}

	for i, line := range lines {
		var got map[string]any
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("unmarshal línea %d: %v", i, err)
		}
		// Comparar campos primarios
		if got["id"] != recs[i]["id"] ||
			got["title"] != recs[i]["title"] ||
			got["created"] != recs[i]["created"] ||
			got["modified"] != recs[i]["modified"] ||
			got["tmap.id"] != recs[i]["tmap.id"] ||
			got["type"] != recs[i]["type"] ||
			got["text"] != recs[i]["text"] {
			t.Errorf("línea %d mismatch\n got:  %+v\n want: %+v", i, got, recs[i])
		}
		// Comparar tags (JSON deserializa a []any)
		gotTags, ok1 := got["tags"].([]any)
		wantTags, ok2 := recs[i]["tags"].([]string)
		if !ok1 || !ok2 || len(gotTags) != len(wantTags) {
			t.Errorf("tags mismatch en línea %d", i)
			continue
		}
		for j := range wantTags {
			if gotTags[j].(string) != wantTags[j] {
				t.Errorf("tags[%d] mismatch en línea %d: got %v want %v", j, i, gotTags[j], wantTags[j])
			}
		}
	}
}

// ----------------------------- ruta inválida -----------------------------
func TestWriteJSONL_InvalidPath(t *testing.T) {
	err := WriteJSONL(context.Background(), "/no/existe/out.jsonl", nil, false)
	if err == nil {
		t.Fatalf("esperaba error en ruta inválida")
	}
}

// ----------------------------- crea directorios automáticamente -----------------------------
func TestWriteJSONL_CreaDirectorios(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "sub1", "sub2", "out.jsonl")
	recs := []map[string]any{{"foo": "bar"}}

	err := WriteJSONL(context.Background(), nestedPath, recs, false)
	if err != nil {
		t.Fatalf("esperaba nil, obtuvo error: %v", err)
	}
	if _, err := os.Stat(nestedPath); err != nil {
		t.Fatalf("el archivo no fue creado: %v", err)
	}
}

// ----------------------------- error si no es slice -----------------------------
func TestWriteJSONL_ErrorSiNoSlice(t *testing.T) {
	tmpFile := tmpPath(t)
	notSlice := map[string]any{"foo": "bar"}

	err := WriteJSONL(context.Background(), tmpFile, notSlice, false)
	if err == nil {
		t.Fatal("esperaba error por tipo no slice, pero fue nil")
	}
}

// ----------------------------- pretty vs compacto -----------------------------
func TestWriteJSONL_PrettyYCompacto(t *testing.T) {
	tmpFile := tmpPath(t)
	recs := []map[string]any{{"foo": "bar"}}

	// Pretty: debe contener indentación interna ("\n  ")
	err := WriteJSONL(context.Background(), tmpFile, recs, true)
	if err != nil {
		t.Fatalf("error en pretty: %v", err)
	}
	content, _ := os.ReadFile(tmpFile)
	if !bytes.Contains(content, []byte("\n  ")) {
		t.Error("no se encontró indentación en pretty")
	}

	// Compacto: no debe contener indentación ("\n  ")
	tmpFile2 := tmpPath(t)
	err = WriteJSONL(context.Background(), tmpFile2, recs, false)
	if err != nil {
		t.Fatalf("error en compacto: %v", err)
	}
	content2, _ := os.ReadFile(tmpFile2)
	if bytes.Contains(content2, []byte("\n  ")) {
		t.Error("se encontró indentación en compacto")
	}
}

// ----------------------------- error de marshal -----------------------------
func TestWriteJSONL_ErrorMarshal(t *testing.T) {
	tmpFile := tmpPath(t)
	data := []any{make(chan int)} // canales no son serializables

	err := WriteJSONL(context.Background(), tmpFile, data, false)
	if err == nil {
		t.Fatal("esperaba error de marshal, pero fue nil")
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
