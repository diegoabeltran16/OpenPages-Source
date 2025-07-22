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
