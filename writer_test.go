// writer_test.go – Tests unitarios para WriteJSONL en writer.go
// ------------------------------------------------------------
package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// writeTempPath genera la ruta de un archivo temporal y lo cierra.
func writeTempPath(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "out-*.jsonl")
	if err != nil {
		t.Fatalf("Error creando archivo temporal: %v", err)
	}
	path := f.Name()
	f.Close()
	return path
}

func TestWriteJSONL_Success(t *testing.T) {
	// Datos de prueba: dos records simples
	recs := []models.Record{
		{
			ID:           "One",
			Tags:         []string{"a", "b"},
			ContentType:  "text/plain",
			TextMarkdown: "foo",
			TextPlain:    "foo",
			CreatedAt:    "20250101",
			ModifiedAt:   "20250102",
		},
		{
			ID:           "Two",
			Tags:         []string{"x"},
			ContentType:  "application/json",
			TextMarkdown: "{\"k\":1}",
			TextPlain:    "{\"k\":1}",
			CreatedAt:    "20250201",
			ModifiedAt:   "20250202",
		},
	}

	// Ruta temporal
	path := writeTempPath(t)
	defer os.Remove(path)

	// Llamar a la función
	if err := WriteJSONL(path, recs); err != nil {
		t.Fatalf("WriteJSONL devolvió error: %v", err)
	}

	// Leer el contenido resultante
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Error leyendo archivo de salida: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != len(recs) {
		t.Fatalf("Número de líneas = %d, want %d", len(lines), len(recs))
	}

	// Parsear y comparar cada línea
	for i, line := range lines {
		var got models.Record
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Errorf("Línea %d: error al parsear JSON: %v", i, err)
			continue
		}
		if !reflect.DeepEqual(got, recs[i]) {
			t.Errorf("Línea %d: got %+v, want %+v", i, got, recs[i])
		}
	}
}

func TestWriteJSONL_InvalidPath(t *testing.T) {
	// Intentar escribir en directorio inexistente
	err := WriteJSONL("/no/existe/salida.jsonl", []models.Record{})
	if err == nil {
		t.Errorf("Esperaba error al escribir en ruta inválida, pero fue nil")
	}
}
