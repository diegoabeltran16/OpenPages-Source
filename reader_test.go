// reader_test.go – Tests unitarios para reader.go
// ------------------------------------------------
// Pruebas de ReadTiddlers para asegurar lectura, manejo de JSON
// malformado y errores de archivo inexistente.
package main

import (
	"encoding/json"
	"os"
	"testing"

	"openpages-source/models"
)

func TestReadTiddlers_Success(t *testing.T) {
	// Crear archivo temporal con JSON válido
	tmpfile, err := os.CreateTemp("", "tiddlers-*.json")
	if err != nil {
		t.Fatalf("no se pudo crear archivo temporal: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Preparar payload de ejemplo
	payload := struct {
		Tiddlers []models.Tiddler `json:"tiddlers"`
	}{
		Tiddlers: []models.Tiddler{
			{
				Title:    "Título",
				Type:     "text/markdown",
				Tags:     []string{"uno", "dos"},
				Text:     "Contenido ejemplo",
				Created:  "20250101120000000",
				Modified: "20250102130000000",
			},
		},
	}

	// Serializar y escribir en archivo
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error al serializar payload: %v", err)
	}
	if err := os.WriteFile(tmpfile.Name(), data, 0644); err != nil {
		t.Fatalf("error al escribir en archivo temporal: %v", err)
	}

	// Ejecutar ReadTiddlers
	tiddlers, err := ReadTiddlers(tmpfile.Name())
	if err != nil {
		t.Fatalf("ReadTiddlers devolvió error inesperado: %v", err)
	}

	// Verificar resultados
	if len(tiddlers) != 1 {
		t.Fatalf("se esperaban 1 tiddler, obtenidos %d", len(tiddlers))
	}
	got := tiddlers[0]
	want := payload.Tiddlers[0]
	if got.Title != want.Title {
		t.Errorf("Title = %q; se esperaba %q", got.Title, want.Title)
	}
	if got.Type != want.Type {
		t.Errorf("Type = %q; se esperaba %q", got.Type, want.Type)
	}
	if len(got.Tags) != len(want.Tags) {
		t.Errorf("Tags length = %d; se esperaba %d", len(got.Tags), len(want.Tags))
	}
	if got.Text != want.Text {
		t.Errorf("Text = %q; se esperaba %q", got.Text, want.Text)
	}
	if got.Created != want.Created {
		t.Errorf("Created = %q; se esperaba %q", got.Created, want.Created)
	}
	if got.Modified != want.Modified {
		t.Errorf("Modified = %q; se esperaba %q", got.Modified, want.Modified)
	}
}

func TestReadTiddlers_MalformedJSON(t *testing.T) {
	// JSON malformado debe producir error
	tmpfile, err := os.CreateTemp("", "malformed-*.json")
	if err != nil {
		t.Fatalf("no se pudo crear archivo temporal: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Escribir contenido inválido
	if err := os.WriteFile(tmpfile.Name(), []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("error al escribir JSON malformado: %v", err)
	}

	_, err = ReadTiddlers(tmpfile.Name())
	if err == nil {
		t.Error("ReadTiddlers no devolvió error con JSON malformado")
	}
}

func TestReadTiddlers_FileNotFound(t *testing.T) {
	// Archivo inexistente debe producir error
	_, err := ReadTiddlers("no_existente.json")
	if err == nil {
		t.Error("ReadTiddlers no devolvió error con archivo inexistente")
	}
}
