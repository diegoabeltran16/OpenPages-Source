// reader_test.go – Tests unitarios para ReadTiddlers en reader.go
// --------------------------------------------------------------------------------
// Contexto pedagógico
// -------------------
// Este archivo acompaña a *reader.go* y demuestra, mediante **pruebas unitarias**
// escritas con el paquete `testing` estándar de Go, que la función `ReadTiddlers`
// interpreta correctamente los dos formatos de exportación que genera
// TiddlyWiki.
//
// Cada prueba sigue la estructura *Arrange → Act → Assert* aunque, para mantener
// la convención idiomática de Go, las secciones no se etiquetan explícitamente.
//
// --------------------------------------------------------------------------------

package importer

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// writeTempFile crea un archivo temporal con el contenido recibido y devuelve
// su ruta.  Cualquier fallo interrumpe la prueba.
func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "tiddlers-*.json")
	if err != nil {
		t.Fatalf("error creando archivo temporal: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("error escribiendo archivo temporal: %v", err)
	}
	return f.Name()
}

// TestRead_Array verifica la ruta feliz cuando el JSON es un array.
func TestRead_Array(t *testing.T) {
	// Arrange
	jsonData := `[
      {"title":"Foo","text":"txt","type":"text/plain","tags":"[[a]]","created":"20250101","modified":"20250102"}
    ]`
	path := writeTempFile(t, jsonData)
	defer os.Remove(path)

	want := []models.Tiddler{{
		Title:    "Foo",
		Text:     "txt",
		Type:     "text/plain",
		Tags:     "[[a]]",
		Created:  "20250101",
		Modified: "20250102",
	}}

	// Act
	got, err := Read(context.Background(), path)
	if err != nil {
		t.Fatalf("Read(array) devolvió error: %v", err)
	}

	// Assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Read(array) = %+v, want %+v", got, want)
	}
}

// TestRead_Map verifica el parseo cuando el JSON es un objeto plano.
func TestRead_Map(t *testing.T) {
	jsonData := `{"Bar":{"title":"Bar","text":"x","type":"application/json","tags":"[[x]]","created":"20250401","modified":"20250402"}}`
	path := writeTempFile(t, jsonData)
	defer os.Remove(path)

	want := []models.Tiddler{{
		Title:    "Bar",
		Text:     "x",
		Type:     "application/json",
		Tags:     "[[x]]",
		Created:  "20250401",
		Modified: "20250402",
	}}

	got, err := Read(context.Background(), path)
	if err != nil {
		t.Fatalf("Read(map) devolvió error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Read(map) = %+v, want %+v", got, want)
	}
}
