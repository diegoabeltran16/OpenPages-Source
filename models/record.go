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

import (
	"strings"
	"time"
)

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
	Name     string `json:"name"`
	RawValue string `json:"value"`
}

type RecordMeta struct {
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
	Meta      RecordMeta          `json:"meta"`
	Content   Content             `json:"content"`
	Relations map[string][]string `json:"relations,omitempty"`
}

// FlattenTags convierte el slice de tags en un string separado por coma.
func (r RecordV2) FlattenTags() string {
	if len(r.Meta.Tags) == 0 {
		return ""
	}
	return strings.Join(r.Meta.Tags, ",")
}

// FlattenRelations devuelve las relaciones "define" y "requiere" como strings separados por coma.
func (r RecordV2) FlattenRelations() (define string, requiere string) {
	if r.Relations == nil {
		return "", ""
	}
	if def, ok := r.Relations["define"]; ok && len(def) > 0 {
		define = strings.Join(def, ",")
	}
	if req, ok := r.Relations["requiere"]; ok && len(req) > 0 {
		requiere = strings.Join(req, ",")
	}
	return
}

// IsAIReady retorna true si el registro tiene los campos clave para IA.
func (r RecordV2) IsAIReady() bool {
	return r.ID != "" && r.Type != "" && r.Content.Plain != ""
}

// -----------------------------------------------------------------------------
// VERSIÓN HÍBRIDA – Para compatibilidad hacia adelante
// -----------------------------------------------------------------------------
// Estructura que combina elementos de v1 y v2 para facilitar la migración.

type RecordHybrid struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Created   string                 `json:"created"`
	Modified  string                 `json:"modified"`
	Tags      []string               `json:"tags"`
	Type      string                 `json:"type"`
	Text      string                 `json:"text"`
	TmapID    string                 `json:"tmap.id"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
	Content   map[string]interface{} `json:"content,omitempty"`
	Relations map[string]interface{} `json:"relations,omitempty"`
}
