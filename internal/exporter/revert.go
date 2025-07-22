package exporter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// RevertToSingleTiddler exporta solo el tiddler raíz como objeto único.
// inputPath: ruta al archivo JSON (array de tiddlers revertidos)
// outputPath: ruta al archivo de salida (objeto único)
// rootTitle: título del tiddler raíz (ej: "_____Nombre del Proyecto")
func RevertToSingleTiddler(ctx context.Context, inputPath, outputPath, rootTitle string) error {
	_ = ctx // reservado para cancelaciones futuras

	// Leer el array de tiddlers
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("leer %s: %w", inputPath, err)
	}
	var tiddlers []map[string]any
	if err := json.Unmarshal(data, &tiddlers); err != nil {
		return fmt.Errorf("parsear array: %w", err)
	}

	// Buscar el tiddler raíz por título
	var root map[string]any
	for _, t := range tiddlers {
		title, ok := t["title"].(string)
		if ok && title == rootTitle {
			root = t
			break
		}
	}
	if root == nil {
		return fmt.Errorf("no se encontró tiddler raíz: %s", rootTitle)
	}

	// Construir objeto único con la clave original
	obj := map[string]any{rootTitle: root}

	// Serializar y guardar el resultado
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("crear %s: %w", outputPath, err)
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(obj); err != nil {
		return fmt.Errorf("serializar objeto único: %w", err)
	}

	fmt.Printf("✅ Tiddler raíz exportado como objeto único en '%s'\n", outputPath)
	return nil
}

// Tiddler estructura para representar un tiddler en formato JSONL
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

// FromJSONLToJSON convierte un archivo JSONL de tiddlers a un archivo JSON
func FromJSONLToJSON(inFile, outFile string) error {
	f, err := os.Open(inFile)
	if err != nil {
		return err
	}
	defer f.Close()

	tiddlers := make(map[string]Tiddler)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var obj map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &obj); err != nil {
			continue
		}

		// Reconstruir tags en formato TiddlyWiki
		tags := ""
		if arr, ok := obj["tags"].([]interface{}); ok {
			for _, tag := range arr {
				tags += "[[" + fmt.Sprint(tag) + "]] "
			}
			tags = strings.TrimSpace(tags)
		} else if str, ok := obj["tags"].(string); ok {
			tags = str
		}

		// Convertir fechas a formato TiddlyWiki (solo yyyyMMdd)
		created := revertDate(getString(obj, "created"))
		modified := revertDate(getString(obj, "modified"))

		title := getString(obj, "title")
		if title == "" {
			title = getString(obj, "id")
		}

		tiddler := Tiddler{
			Title:    title,
			Text:     getString(obj, "text"),
			Type:     getString(obj, "type"),
			Tags:     tags,
			Created:  created,
			Modified: modified,
			Color:    getString(obj, "color"),
			TmapID:   getString(obj, "tmap.id"),
		}

		if tiddler.Title != "" {
			tiddlers[tiddler.Title] = tiddler
		}
	}

	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(tiddlers)
}

// getString devuelve el valor string del campo o "" si es nil o no existe
func getString(obj map[string]interface{}, key string) string {
	if v, ok := obj[key]; ok && v != nil {
		return fmt.Sprint(v)
	}
	// Mapeo alternativo para campos comunes
	switch key {
	case "title":
		if v, ok := obj["id"]; ok && v != nil {
			return fmt.Sprint(v)
		}
	case "text":
		if v, ok := obj["body"]; ok && v != nil {
			return fmt.Sprint(v)
		}
	}
	return ""
}

// revertDate convierte "2025-06-05T15:10:00-05:00" → "20250605"
func revertDate(iso string) string {
	t, err := time.Parse("2006-01-02T15:04:05-07:00", iso)
	if err != nil {
		return ""
	}
	return t.Format("20060102")
}

func CloneAndUpdateTexts(plantillaPath, jsonlPath, outPath string) error {
	// 1. Leer la plantilla como array
	plantillaFile, err := os.Open(plantillaPath)
	if err != nil {
		return err
	}
	defer plantillaFile.Close()
	var plantillaArr []Tiddler
	if err := json.NewDecoder(plantillaFile).Decode(&plantillaArr); err != nil {
		return err
	}

	// Convertir array a objeto por título
	plantilla := make(map[string]Tiddler)
	for _, t := range plantillaArr {
		plantilla[t.Title] = t
	}

	// 2. Leer los textos nuevos desde JSONL
	updates := make(map[string]string)
	jsonlFile, err := os.Open(jsonlPath)
	if err != nil {
		return err
	}
	defer jsonlFile.Close()
	scanner := bufio.NewScanner(jsonlFile)
	for scanner.Scan() {
		var obj map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &obj); err != nil {
			continue
		}
		title := getString(obj, "title")
		if title == "" {
			title = getString(obj, "id")
		}
		text := getString(obj, "text")
		if text != "" {
			// Si el campo "text" en JSONL es un objeto, conviértelo a string plano
			if isJSONString(text) {
				// Si es un JSON serializado, puedes extraer el campo "plain" si existe
				var temp map[string]interface{}
				if err := json.Unmarshal([]byte(text), &temp); err == nil {
					if plain, ok := temp["plain"].(string); ok {
						text = plain
					}
				}
			}
			updates[title] = text
		}
	}

	// 3. Actualizar los textos en la plantilla
	for k, t := range plantilla {
		if newText, ok := updates[k]; ok {
			t.Text = newText
			plantilla[k] = t
		}
	}

	// 4. Guardar el resultado como array
	var resultArr []Tiddler
	for _, t := range plantilla {
		resultArr = append(resultArr, t)
	}

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(resultArr)
}

// isJSONString verifica si una cadena es un JSON válido
func isJSONString(s string) bool {
	return strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")
}
