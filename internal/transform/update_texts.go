package transform

import (
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
