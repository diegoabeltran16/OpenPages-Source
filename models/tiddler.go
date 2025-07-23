// models/tiddler.go – Definición del Tiddler crudo exportado de TiddlyWiki
// -----------------------------------------------------------
// Package models contiene las estructuras de datos centrales para el pipeline.
package models

import "encoding/json"

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
	Title       string                 `json:"title"`
	Text        string                 `json:"text"`
	Type        string                 `json:"type"`
	Tags        string                 `json:"tags"`
	Created     string                 `json:"created"`
	Modified    string                 `json:"modified"`
	Color       string                 `json:"color,omitempty"`
	TmapID      string                 `json:"tmap.id,omitempty"`
	ExtraFields map[string]interface{} `json:"-"`                   // No se serializa automáticamente
	Relations   []string               `json:"relations,omitempty"` // Campo para relaciones
}

// UnmarshalJSON implementa unmarshaling personalizado para capturar campos dinámicos
func (t *Tiddler) UnmarshalJSON(data []byte) error {
	// Estructura auxiliar con todos los campos conocidos
	type Alias Tiddler
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	// Unmarshal normal
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
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
