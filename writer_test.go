// writer_test.go – Tests unitarios para writer.go
// ------------------------------------------------
// Pruebas de WriteJSONL para asegurar escritura correcta de JSONL
package main

import (
	"bufio"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"diegoabeltran16/OpenPages-Source/models"
)

func TestWriteJSONL_Success(t *testing.T) {
	// Preparar registros de ejemplo
	records := []models.Record{
		{
			ID:           "id1",
			Title:        "Primer Título",
			ContentType:  "text/plain",
			Tags:         []string{"a", "b"},
			TextMarkdown: "Texto **fuerte**",
			TextPlain:    "Texto fuerte",
			CreatedAt:    "2025-01-01T12:00:00Z",
			ModifiedAt:   "2025-01-02T13:00:00Z",
		},
		{
			ID:           "id2",
			Title:        "Segundo Título",
			ContentType:  "text/markdown",
			Tags:         []string{"x", "y", "z"},
			TextMarkdown: "# Encabezado",
			TextPlain:    "Encabezado",
			CreatedAt:    "2025-02-01T08:30:00Z",
			ModifiedAt:   "2025-02-01T09:45:00Z",
		},
	}

	// Crear archivo temporal de salida
	tmpfile, err := os.CreateTemp("", "records-*.jsonl")
	if err != nil {
		t.Fatalf("no se pudo crear archivo temporal: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Llamar a WriteJSONL
	if err := WriteJSONL(tmpfile.Name(), records); err != nil {
		t.Fatalf("WriteJSONL devolvió error: %v", err)
	}

	// Abrir el archivo para lectura
	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatalf("no se pudo abrir archivo temporal: %v", err)
	}
	defer file.Close()

	// Leer línea por línea y deserializar
	scanner := bufio.NewScanner(file)
	var got []models.Record
	for scanner.Scan() {
		var rec models.Record
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			t.Fatalf("error al parsear JSONL: %v", err)
		}
		got = append(got, rec)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("error al escanear archivo: %v", err)
	}

	// Verificar número de registros
	if len(got) != len(records) {
		t.Fatalf("se esperaban %d registros, pero se obtuvieron %d", len(records), len(got))
	}

	// Comparar cada registro
	for i := range records {
		if !reflect.DeepEqual(got[i], records[i]) {
			t.Errorf("Registro %d diferente:\n got= %+v\n want=%+v", i, got[i], records[i])
		}
	}
}

func TestWriteJSONL_ErrorCreatingFile(t *testing.T) {
	// Intentar escribir en un directorio inexistente
	badPath := "/ruta/que/no/existe/output.jsonl"
	recs := []models.Record{}
	if err := WriteJSONL(badPath, recs); err == nil {
		t.Error("WriteJSONL no devolvió error al intentar crear archivo en ruta inválida")
	}
}
