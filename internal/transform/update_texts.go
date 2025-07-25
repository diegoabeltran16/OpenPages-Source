package transform

import (
	"testing"
	"time"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// UpdateTexts actualiza solo los campos "text" y "modified" en la plantilla usando los tiddlers de updates.
// Si el "text" cambió, actualiza "text" y "modified" (con fecha actual TiddlyWiki).
func UpdateTexts(template []models.Tiddler, updates []models.Tiddler) []models.Tiddler {
	// Indexar actualizaciones por título
	updatesByTitle := make(map[string]string)
	for _, upd := range updates {
		updatesByTitle[upd.Title] = upd.Text
	}
	now := time.Now().Format("20060102150405") // formato TiddlyWiki

	for i := range template {
		title := template[i].Title
		if newText, ok := updatesByTitle[title]; ok && template[i].Text != newText {
			template[i].Text = newText
			template[i].Modified = now
		}
	}
	return template
}

func TestUpdateTexts(t *testing.T) {
	plantilla := []models.Tiddler{
		{Title: "A", Text: "foo", Modified: "20240101"},
		{Title: "B", Text: "bar", Modified: "20240102"},
	}
	updates := []models.Tiddler{
		{Title: "A", Text: "foo"},         // igual, no debe cambiar
		{Title: "B", Text: "nuevo texto"}, // diferente, debe actualizarse
	}

	result := UpdateTexts(plantilla, updates)

	// El texto de A no cambia, ni la fecha
	if result[0].Text != "foo" || result[0].Modified != "20240101" {
		t.Errorf("No debe cambiar el tiddler A")
	}

	// El texto de B cambia, y la fecha debe ser "ahora" (formato TiddlyWiki)
	if result[1].Text != "nuevo texto" {
		t.Errorf("Debe actualizar el texto de B")
	}
	if result[1].Modified == "20240102" {
		t.Errorf("Debe actualizar la fecha de B")
	}
	if len(result) != 2 {
		t.Errorf("No debe cambiar la cantidad de tiddlers")
	}

	// Verifica formato TiddlyWiki (14 dígitos numéricos)
	if result[1].Modified == "" || len(result[1].Modified) != 14 {
		t.Errorf("La fecha modificada debe tener formato TiddlyWiki (yyyymmddhhMMSS)")
	}
}
