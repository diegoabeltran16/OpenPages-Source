// models/tiddler.go – Definición del Tiddler crudo exportado de TiddlyWiki
// -----------------------------------------------------------
// Package models contiene las estructuras de datos centrales para el pipeline.
package models

import (
	"encoding/json"
	"strings"
)

// Tiddler representa un elemento exportado de TiddlyWiki.
// Cada campo refleja el JSON original:
//   - Title:    Título único del tiddler.
//   - Text:     Contenido bruto (puede incluir JSON embebido, Markdown, texto plano).
//   - Type:     Tipo MIME del contenido (e.g., application/json, text/markdown).
//   - Tags:     Cadena con etiquetas en formato [[tag1]] [[tag2]].
//   - Created:  Timestamp de creación en formato TiddlyWiki.
//   - Modified: Timestamp de última modificación en formato TiddlyWiki.
//   - Color:    Color asociado (opcional).
//   - TmapID:   Identificador interno de TiddlyMap (opcional).
type Tiddler struct {
	Title     string                 `json:"title"`
	Text      string                 `json:"text"`
	Type      string                 `json:"type"`
	Tags      any                    `json:"tags"`
	Created   string                 `json:"created"`
	Modified  string                 `json:"modified"`
	Color     string                 `json:"color,omitempty"`
	Path      string                 `json:"path,omitempty"` // <--- AGREGA ESTA LÍNEA
	TmapID    string                 `json:"tmap.id,omitempty"`
	Relations map[string]interface{} `json:"relations,omitempty"`
	// Campos opcionales para compatibilidad
	TextMarkdown string                 `json:"textMarkdown,omitempty"`
	Content      map[string]interface{} `json:"content,omitempty"`
	Meta         *Meta                  `json:"meta,omitempty"` // Usa el struct tipado de record.go
	ExtraFields  map[string]interface{} `json:"-"`              // No se serializa automáticamente
	TagsList     []string               `json:"tags_list,omitempty"`
}

// UnmarshalJSON implementa unmarshaling personalizado para capturar campos dinámicos
func (t *Tiddler) UnmarshalJSON(data []byte) error {
	type Alias Tiddler
	aux := &struct {
		*Alias
		Tags any `json:"tags"`
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Manejo robusto de tags
	switch v := aux.Tags.(type) {
	case string:
		t.Tags = v
	case []interface{}:
		tags := make([]string, 0, len(v))
		for _, tag := range v {
			if s, ok := tag.(string); ok {
				tags = append(tags, s)
			}
		}
		t.Tags = tags
	}
	// Unmarshal a map para capturar campos adicionales
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Inicializar ExtraFields
	t.ExtraFields = make(map[string]interface{})

	// Capturar campos que no están en la estructura
	knownFields := map[string]bool{
		"title": true, "text": true, "type": true, "tags": true,
		"created": true, "modified": true, "color": true, "tmap.id": true,
		"path":         true, // <-- Añade path
		"tags_list":    true, // <-- Añade tags_list
		"relations":    true, // <-- Añade relations
		"textMarkdown": true, // <-- Añade textMarkdown si lo usas
		"content":      true, // <-- Añade content si lo usas
		"meta":         true, // <-- Añade meta si lo usas
	}

	for key, value := range raw {
		if !knownFields[key] {
			t.ExtraFields[key] = value
		}
	}

	return nil
}

// MarshalJSON implementa marshaling personalizado para incluir ExtraFields
func (t *Tiddler) MarshalJSON() ([]byte, error) {
	// Marshal los campos conocidos
	type Alias Tiddler
	base := make(map[string]interface{})
	b, err := json.Marshal((*Alias)(t))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &base); err != nil {
		return nil, err
	}
	// Agrega los campos extra
	for k, v := range t.ExtraFields {
		base[k] = v
	}
	return json.Marshal(base)
}

// TagsAsSlice devuelve las etiquetas del Tiddler como un slice de strings.
func (t *Tiddler) TagsAsSlice() []string {
	switch v := t.Tags.(type) {
	case string:
		return parseTags(v) // Usa tu función parseTags
	case []string:
		return v
	case []interface{}:
		tags := make([]string, 0, len(v))
		for _, tag := range v {
			if s, ok := tag.(string); ok {
				tags = append(tags, s)
			}
		}
		return tags
	default:
		return nil
	}
}

func parseTags(tags string) []string {
	tags = strings.TrimSpace(tags)
	if tags == "" {
		return nil
	}
	var result []string
	for _, tag := range strings.Fields(tags) {
		tag = strings.Trim(tag, "[]")
		if tag != "" {
			result = append(result, tag)
		}
	}
	return result
}

func (t *Tiddler) GetCreated() string {
	created := t.Created
	modified := t.Modified
	color := t.Color
	tmapid := t.TmapID

	if t.Meta != nil {
		if created == "" {
			created = t.Meta.Created
		}
		if modified == "" {
			modified = t.Meta.Modified
		}
		if color == "" {
			color = t.Meta.Color
		}
		if t.Meta.Extra != nil {
			if tmapid == "" {
				tmapid = t.Meta.Extra["tmap.id"]
			}
			if color == "" {
				color = t.Meta.Extra["color"]
			}
		}
	}
	return created
}

func (t *Tiddler) GetTmapID() string {
	tmapid := t.TmapID
	if tmapid == "" {
		if t.Relations != nil {
			if val, ok := t.Relations["tmap.id"]; ok {
				if id, ok := val.(string); ok {
					tmapid = id
				}
			}
		}
	}
	return tmapid
}

func (t *Tiddler) GetColor() string {
	color := t.Color
	if t.Meta != nil {
		if color == "" {
			color = t.Meta.Color
		}
		if t.Meta.Extra != nil && color == "" {
			color = t.Meta.Extra["color"]
		}
	}
	return color
}

func (t *Tiddler) GetModified() string {
	modified := t.Modified
	if t.Meta != nil {
		if modified == "" {
			modified = t.Meta.Modified
		}
	}
	return modified
}

// Meta representa la metadata asociada a un Tiddler, utilizada en el campo Meta de Tiddler.
type Meta struct {
	Title    string            `json:"title"`
	Tags     []string          `json:"tags,omitempty"`
	Created  string            `json:"created,omitempty"`  // Usa string para simplificar
	Modified string            `json:"modified,omitempty"` // Usa string para simplificar
	Color    string            `json:"color,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}
