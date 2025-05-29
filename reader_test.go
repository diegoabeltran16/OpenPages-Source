// reader_test.go – Tests unitarios para ReadTiddlers en reader.go
// ---------------------------------------------------------------
package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// writeTempFile crea un archivo temporal con el contenido dado y devuelve su ruta.
func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "tiddlers-*.json")
	if err != nil {
		t.Fatalf("Error al crear archivo temporal: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Error al escribir en archivo temporal: %v", err)
	}
	return f.Name()
}

func TestReadTiddlersArray(t *testing.T) {
	jsonData := `[
	  {
	    "title": "Foo",
	    "text": "Contenido de Foo",
	    "type": "text/plain",
	    "tags": "[[alpha]] [[beta]]",
	    "created": "20250101",
	    "modified": "20250102",
	    "color": "#ff0000",
	    "tmap.id": "id-foo"
	  }
	]`
	path := writeTempFile(t, jsonData)
	defer os.Remove(path)

	want := []models.Tiddler{
		{
			Title:    "Foo",
			Text:     "Contenido de Foo",
			Type:     "text/plain",
			Tags:     "[[alpha]] [[beta]]",
			Created:  "20250101",
			Modified: "20250102",
			Color:    "#ff0000",
			TmapID:   "id-foo",
		},
	}

	got, err := ReadTiddlers(path)
	if err != nil {
		t.Fatalf("ReadTiddlers(array) devolvió error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadTiddlers(array) = %+v, want %+v", got, want)
	}
}

func TestReadTiddlersMap(t *testing.T) {
	jsonData := `{ 
	  "Bar": {
	    "title": "Bar",
	    "text": "Contenido de Bar",
	    "type": "application/json",
	    "tags": "[[x]]",
	    "created": "20250401",
	    "modified": "20250402",
	    "color": "#00ff00",
	    "tmap.id": "id-bar"
	  }
	}`
	path := writeTempFile(t, jsonData)
	defer os.Remove(path)

	want := []models.Tiddler{
		{
			Title:    "Bar",
			Text:     "Contenido de Bar",
			Type:     "application/json",
			Tags:     "[[x]]",
			Created:  "20250401",
			Modified: "20250402",
			Color:    "#00ff00",
			TmapID:   "id-bar",
		},
	}

	got, err := ReadTiddlers(path)
	if err != nil {
		t.Fatalf("ReadTiddlers(map) devolvió error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadTiddlers(map) = %+v, want %+v", got, want)
	}
}
