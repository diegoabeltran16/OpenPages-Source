// models/record.go – Definición del Record procesado para exportación en JSONL
// -----------------------------------------------------------
// Package models contiene las estructuras de datos centrales para el pipeline.
package models

// Record representa la versión procesada de un Tiddler,
// lista para exportarse en formato JSONL.
// - ID:           Título original del tiddler
// - Tags:         Lista de etiquetas extraídas
// - ContentType:  Tipo MIME del contenido
// - TextMarkdown: Contenido con sintaxis conservada (Markdown o JSON formateado)
// - TextPlain:    Versión en texto plano
// - CreatedAt:    Timestamp de creación original
// - ModifiedAt:   Timestamp de última modificación
type Record struct {
	ID           string   `json:"id"`
	Tags         []string `json:"tags"`
	ContentType  string   `json:"contentType"`
	TextMarkdown string   `json:"textMarkdown"`
	TextPlain    string   `json:"textPlain"`
	CreatedAt    string   `json:"createdAt"`
	ModifiedAt   string   `json:"modifiedAt"`
}
