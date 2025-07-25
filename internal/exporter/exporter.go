package exporter

import (
	"encoding/json"
	"fmt"
	"os"
)

// WriteJSON vuelca “v” a disco en outputPath como JSON.
// Si pretty es true, usa indentación de 2 espacios; si no, JSON compacto.
func WriteJSON(outputPath string, v any, pretty bool) error {
	var (
		data []byte
		err  error
	)
	if pretty {
		data, err = json.MarshalIndent(v, "", "  ")
	} else {
		data, err = json.Marshal(v)
	}
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("write file %s: %w", outputPath, err)
	}
	return nil
}

func ExportToJSONL(tiddlers []Tiddler, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, t := range tiddlers {
		// Crea un objeto plano para exportar
		exportObj := map[string]any{
			"title":     t.Title,
			"type":      t.Type,
			"tags":      t.Tags,
			"created":   t.Created,
			"modified":  t.Modified,
			"color":     t.Color,
			"tmap.id":   t.TmapID,
			"textPlain": GetTextContent(t.Text),
		}
		if err := enc.Encode(exportObj); err != nil {
			return err
		}
	}
	return nil
}

func GetTextContent(text string) string {
	if len(text) > 0 && text[0] == '{' && text[len(text)-1] == '}' {
		var w map[string]any
		if err := json.Unmarshal([]byte(text), &w); err == nil {
			if c, ok := w["content"].(map[string]any); ok {
				if plain, ok := c["plain"].(string); ok && plain != "" {
					return plain
				}
				if markdown, ok := c["markdown"].(string); ok && markdown != "" {
					return markdown
				}
			}
		}
	}
	return text
}
