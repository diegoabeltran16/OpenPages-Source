// models/record.go – Versiones v1 y v2 de Record
// --------------------------------------------------------------------------------
// Contexto pedagógico
// -------------------
// Este archivo concentra **ambas** representaciones que usa el pipeline:
//   • `Record`  (v1) → estructura compacta, utilizada hasta ahora.
//   • `RecordV2` (v2) → esquema "AI‑friendly" con meta ↔ content separados.
//
// Mantener los dos modelos en un solo archivo permite evolucionar gradualmente
// sin romper compatibilidad.  El conversor v1 sigue funcionando tal cual; el
// conversor v2 emitirá la nueva forma sólo cuando el usuario pase `-mode v2`.
// --------------------------------------------------------------------------------

package models

import "time"

// -----------------------------------------------------------------------------
// VERSIÓN 1 – Compacta (heredada)
// -----------------------------------------------------------------------------
// Usada por ConvertTiddlers (v1).  Se conserva para no romper flujos existentes.

type Record struct {
	ID           string   `json:"id"` // normalmente igual a Title
	Tags         []string `json:"tags,omitempty"`
	ContentType  string   `json:"type,omitempty"`
	TextMarkdown string   `json:"textMarkdown,omitempty"`
	TextPlain    string   `json:"textPlain,omitempty"`
	CreatedAt    string   `json:"createdAt,omitempty"` // formato yyyymmdd… (legacy)
	ModifiedAt   string   `json:"modifiedAt,omitempty"`
	Color        string   `json:"color,omitempty"`
}

// -----------------------------------------------------------------------------
// VERSIÓN 2 – “AI‑friendly” (meta vs content)
// -----------------------------------------------------------------------------
// Nuevas estructuras

type Content struct {
	Plain    string         `json:"plain,omitempty"`
	Markdown string         `json:"markdown,omitempty"`
	JSON     map[string]any `json:"json,omitempty"`
	Sections []Section      `json:"sections,omitempty"`
}

type Section struct {
	Level   int    `json:"level"`
	Heading string `json:"heading"`
	Text    string `json:"text"`
}

type Meta struct {
	Title    string            `json:"title"`
	Tags     []string          `json:"tags,omitempty"`
	Created  time.Time         `json:"created,omitempty"`
	Modified time.Time         `json:"modified,omitempty"`
	Color    string            `json:"color,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}

type RecordV2 struct {
	ID        string              `json:"id"`
	Type      string              `json:"type"` // "tiddler", "fragment", etc.
	Meta      Meta                `json:"meta"`
	Content   Content             `json:"content"`
	Relations map[string][]string `json:"relations,omitempty"`
}
