// models/tiddler.go – Definición del Tiddler crudo exportado de TiddlyWiki
// -----------------------------------------------------------
// Package models contiene las estructuras de datos centrales para el pipeline.
package models

// Tiddler representa un elemento exportado de TiddlyWiki.
// Cada campo refleja el JSON original:
//  - Title:    Título único del tiddler.
//  - Text:     Contenido bruto (puede incluir JSON embebido, Markdown, texto plano).
//  - Type:     Tipo MIME del contenido (e.g., application/json, text/markdown).
//  - Tags:     Cadena con etiquetas en formato [[tag1]] [[tag2]].
//  - Created:  Timestamp de creación en formato TiddlyWiki.
//  - Modified: Timestamp de última modificación en formato TiddlyWiki.
//  - Color:    Color asociado (opcional).
//  - TmapID:   Identificador interno de TiddlyMap (opcional).
type Tiddler struct {
	Title    string `json:"title"`
	Text     string `json:"text"`
	Type     string `json:"type"`
	Tags     string `json:"tags"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
	Color    string `json:"color,omitempty"`
	TmapID   string `json:"tmap.id,omitempty"`
}
