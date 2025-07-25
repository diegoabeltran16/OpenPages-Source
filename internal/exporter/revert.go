package exporter

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
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

// ExportAllFromJSONL exporta todos los tiddlers desde un archivo JSONL a un archivo de salida
func ExportAllFromJSONL(jsonlPath, outPath string) error {
	file, err := os.Open(jsonlPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var resultArr []map[string]any
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var obj map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &obj); err != nil {
			log.Printf("invalid JSONL line: %v", err)
			continue
		}
		title := getString(obj, "title")
		if title == "" || title == "<nil>" {
			continue
		}

		// tags: array de strings (si es string, parsear)
		var tagsArr []string
		if arr, ok := obj["tags"].([]interface{}); ok {
			for _, tag := range arr {
				if s, ok := tag.(string); ok {
					tagsArr = append(tagsArr, s)
				}
			}
		} else if s, ok := obj["tags"].(string); ok {
			tagsArr = parseTags(s)
		} else {
			tagsArr = []string{}
		}

		// relations
		relations := map[string]any{}
		if v, ok := obj["relations"].(map[string]any); ok {
			relations = v
		}

		// tags_list: copiar tal cual si existe, si no, array vacío
		var tagsList []string
		if arr, ok := obj["tags_list"].([]interface{}); ok {
			for _, tag := range arr {
				if s, ok := tag.(string); ok {
					tagsList = append(tagsList, s)
				}
			}
		} else if arr, ok := obj["tags_list"].([]string); ok {
			tagsList = arr
		} else {
			tagsList = []string{}
		}

		// text: todo lo que no sea campo estándar, serializado como JSON (MOVER AQUÍ)
		standardFields := map[string]bool{
			"title": true, "type": true, "tags": true, "tags_list": true,
			"hash": true, "path": true, "color": true, "tmap.id": true, "relations": true,
			"created": true, "modified": true, "id": true, "created_rfc": true, "modified_rfc": true,
		}
		textMap := make(map[string]any)
		for k, v := range obj {
			if !standardFields[k] {
				textMap[k] = v
			}
		}
		textBytes, _ := json.Marshal(textMap)
		text := string(textBytes)
		if len(textMap) == 0 {
			text = "{}"
		}

		// color, created, modified, path, tmap.id: siempre string ("" si no existe)
		color := getString(obj, "color")
		created := getString(obj, "created")
		modified := getString(obj, "modified")
		path := getString(obj, "path")
		tmapid := getString(obj, "tmap.id")

		// --- Poblar desde meta y extra si están vacíos ---
		if meta, ok := obj["meta"].(map[string]interface{}); ok {
			if color == "" {
				color = getString(meta, "color")
			}
			if created == "" {
				created = getString(meta, "created")
			}
			if modified == "" {
				modified = getString(meta, "modified")
			}
			if tmapid == "" {
				tmapid = getString(meta, "tmap.id")
			}
			if path == "" {
				path = getString(meta, "path")
			}
			// Buscar en meta.extra
			if extra, ok := meta["extra"].(map[string]interface{}); ok {
				if tmapid == "" {
					tmapid = getString(extra, "tmap.id")
				}
				if color == "" {
					color = getString(extra, "color")
				}
				if path == "" {
					path = getString(extra, "path")
				}
			}
		}

		// Si aún están vacíos, intentar extraer desde el campo "text" (JSON serializado)
		if looksLikeJSON(text) {
			var inner map[string]interface{}
			if err := json.Unmarshal([]byte(text), &inner); err == nil {
				if created == "" {
					created = getString(inner, "created")
				}
				if modified == "" {
					modified = getString(inner, "modified")
				}
				if color == "" {
					color = getString(inner, "color")
				}
				if tmapid == "" {
					tmapid = getString(inner, "tmap.id")
				}
				if path == "" {
					path = getString(inner, "path")
				}
				// Buscar en inner.meta.extra
				if meta, ok := inner["meta"].(map[string]interface{}); ok {
					if color == "" {
						color = getString(meta, "color")
					}
					if created == "" {
						created = getString(meta, "created")
					}
					if modified == "" {
						modified = getString(meta, "modified")
					}
					if tmapid == "" {
						tmapid = getString(meta, "tmap.id")
					}
					if path == "" {
						path = getString(meta, "path")
					}
					if extra, ok := meta["extra"].(map[string]interface{}); ok {
						if tmapid == "" {
							tmapid = getString(extra, "tmap.id")
						}
						if color == "" {
							color = getString(extra, "color")
						}
						if path == "" {
							path = getString(extra, "path")
						}
					}
				}
			}
		}

		// hash
		hash := hashSHA256(text)

		// Si tagsArr y tagsList están vacíos, intenta extraerlos desde meta.tags en el campo "text"
		if len(tagsArr) == 0 && len(tagsList) == 0 && looksLikeJSON(text) {
			var inner map[string]interface{}
			if err := json.Unmarshal([]byte(text), &inner); err == nil {
				// Buscar en inner["meta"]["tags"]
				if meta, ok := inner["meta"].(map[string]interface{}); ok {
					if tagsRaw, ok := meta["tags"]; ok {
						switch tags := tagsRaw.(type) {
						case []interface{}:
							for _, tag := range tags {
								if s, ok := tag.(string); ok {
									tagsArr = append(tagsArr, s)
									tagsList = append(tagsList, s)
								}
							}
						case []string:
							tagsArr = append(tagsArr, tags...)
							tagsList = append(tagsList, tags...)
						}
					}
				}
			}
		}

		// --- Unificación y sincronización final de tags ---
		if len(tagsArr) == 0 && len(tagsList) > 0 {
			tagsArr = append(tagsArr, tagsList...)
		}
		if len(tagsList) == 0 && len(tagsArr) > 0 {
			// (No action needed: tagsList is set to uniqueTags below)
		}
		// Eliminar duplicados y mantener orden
		tagsSeen := make(map[string]struct{})
		uniqueTags := make([]string, 0, len(tagsArr))
		for _, tag := range tagsArr {
			if _, ok := tagsSeen[tag]; !ok {
				tagsSeen[tag] = struct{}{}
				uniqueTags = append(uniqueTags, tag)
			}
		}
		tagsArr = uniqueTags
		tagsList = uniqueTags

		// Construir tagsTW usando helper
		tagsTW := buildTagsTW(tagsList, tagsArr)

		tiddler := map[string]any{
			"title":     title,
			"text":      text,
			"type":      "application/json",
			"tags":      tagsTW,   // <-- string TiddlyWiki
			"tags_list": tagsList, // <-- array
			"created":   created,
			"modified":  modified,
			"hash":      hash,
			"path":      path,
			"color":     color,
			"tmap.id":   tmapid,
			"relations": relations,
		}
		resultArr = append(resultArr, tiddler)
	}

	// Verificar errores del scanner
	if err := scanner.Err(); err != nil {
		return err
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

// buildTagsTW construye el string TiddlyWiki de tags
func buildTagsTW(tagsList, tagsArr []string) string {
	var tagsTW string
	// Usar la lista más larga y no vacía
	var tags []string
	if len(tagsList) > 0 {
		tags = tagsList
	} else {
		tags = tagsArr
	}
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		// Siempre envolver en [[...]] si no lo está
		if !strings.HasPrefix(trimmed, "[[") || !strings.HasSuffix(trimmed, "]]") {
			tagsTW += "[[" + trimmed + "]] "
		} else {
			tagsTW += trimmed + " "
		}
	}
	return strings.TrimSpace(tagsTW)
}

// CloneAndUpdateTexts es la función principal de revertido 100% robusta
func CloneAndUpdateTexts(plantillaPath, jsonlPath, outPath string) error {
	// 1. Leer la plantilla como array
	plantillaFile, err := os.Open(plantillaPath)
	if err != nil {
		return fmt.Errorf("abrir plantilla %s: %w", plantillaPath, err)
	}
	defer plantillaFile.Close()

	var plantillaArr []Tiddler
	if err := json.NewDecoder(plantillaFile).Decode(&plantillaArr); err != nil {
		return fmt.Errorf("decodificar plantilla: %w", err)
	}

	// 2. Leer los textos nuevos desde JSONL con máxima robustez
	updates := make(map[string]string)
	jsonlFile, err := os.Open(jsonlPath)
	if err != nil {
		return fmt.Errorf("abrir JSONL %s: %w", jsonlPath, err)
	}
	defer jsonlFile.Close()

	scanner := bufio.NewScanner(jsonlFile)
	for scanner.Scan() {
		var obj map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &obj); err != nil {
			log.Printf("invalid JSONL line: %v", err)
			continue
		}

		title := getString(obj, "title")
		if title == "" {
			title = getString(obj, "id")
		}
		if title == "" || title == "<nil>" {
			continue
		}

		text := ExtractTextFromJSONL(obj)
		if text != "" && !looksLikeJSON(text) {
			updates[title] = text
		}
	}

	// Verificar errores del scanner
	if err := scanner.Err(); err != nil {
		return err
	}

	// 3. Actualizar los textos en la plantilla preservando estructura
	now := time.Now().Format("20060102150405")
	var resultArr []Tiddler
	applied := 0 // Contador de actualizaciones aplicadas

	for _, t := range plantillaArr {
		if t.Title == "" || t.Title == "<nil>" {
			continue
		}

		if t.Text == "" && t.Type == "" && t.Tags == "" &&
			t.Created == "" && t.Modified == "" && t.Color == "" && t.TmapID != "" {
			continue
		}

		if strings.HasSuffix(t.Title, ".json") || strings.HasSuffix(t.Title, ".jsonl") {
			continue
		}

		if newText, hasUpdate := updates[t.Title]; hasUpdate {
			needsUpdate := false
			if looksLikeJSON(t.Text) {
				var wrapper map[string]interface{}
				if err := json.Unmarshal([]byte(t.Text), &wrapper); err == nil {
					if content, ok := wrapper["content"].(map[string]interface{}); ok {
						if currentPlain := getString(content, "plain"); currentPlain != newText {
							needsUpdate = true
						}
					} else {
						needsUpdate = true
					}
				}
			} else {
				needsUpdate = (t.Text != newText)
			}

			if needsUpdate {
				t = UpdateTiddlerWrapper(t, newText, "")
				t.Modified = now
				applied++
			}
		}

		resultArr = append(resultArr, t)
	}

	// 4. Guardar el resultado como array con formato bonito
	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("crear archivo salida %s: %w", outPath, err)
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resultArr); err != nil {
		return fmt.Errorf("codificar resultado: %w", err)
	}

	fmt.Printf("✅ Revertido completado: %d tiddlers procesados, %d actualizaciones aplicadas\n",
		len(resultArr), applied)
	return nil
}

// getString devuelve el valor string del campo o "" si es nil o no existe
func getString(obj map[string]interface{}, key string) string {
	if v, ok := obj[key]; ok && v != nil {
		return fmt.Sprint(v)
	}
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

// ExtractTextFromJSONL extrae el texto plano desde un objeto JSONL con máxima robustez
func ExtractTextFromJSONL(obj map[string]interface{}) string {
	var text string

	text = getString(obj, "textPlain")
	if text != "" {
		return text
	}

	text = getString(obj, "contentPlain")
	if text != "" {
		return text
	}

	if content, ok := obj["content"].(map[string]interface{}); ok {
		text = getString(content, "plain")
		if text != "" {
			return text
		}
		text = getString(content, "contentPlain")
		if text != "" {
			return text
		}
	}

	text = getString(obj, "textMarkdown")
	if text != "" {
		return text
	}

	text = getString(obj, "contentMarkdown")
	if text != "" {
		return text
	}

	if content, ok := obj["content"].(map[string]interface{}); ok {
		text = getString(content, "markdown")
		if text != "" {
			return text
		}
		text = getString(content, "contentMarkdown")
		if text != "" {
			return text
		}
	}

	if rawText, ok := obj["text"]; ok {
		if s, ok := rawText.(string); ok {
			if looksLikeJSON(s) {
				var inner map[string]interface{}
				if err := json.Unmarshal([]byte(s), &inner); err == nil {
					text = getString(inner, "plain")
					if text != "" {
						return text
					}
					text = getString(inner, "contentPlain")
					if text != "" {
						return text
					}
					text = getString(inner, "markdown")
					if text != "" {
						return text
					}
					text = getString(inner, "contentMarkdown")
					if text != "" {
						return text
					}
				}
			} else {
				return s
			}
		}
	}

	return ""
}

// UpdateTiddlerWrapper actualiza el wrapper JSON de un tiddler manteniendo la estructura original
func UpdateTiddlerWrapper(original Tiddler, newPlain, newMarkdown string) Tiddler {
	if !looksLikeJSON(original.Text) {
		original.Text = newPlain
		return original
	}

	var wrapper map[string]interface{}
	if err := json.Unmarshal([]byte(original.Text), &wrapper); err != nil {
		original.Text = newPlain
		return original
	}

	content, ok := wrapper["content"].(map[string]interface{})
	if !ok {
		content = make(map[string]interface{})
	}

	if newPlain != "" {
		content["plain"] = newPlain
	}
	if newMarkdown != "" {
		content["markdown"] = newMarkdown
	}

	if newPlain == "" && content["plain"] != nil {
		// Mantener el plain existente
	}
	if newMarkdown == "" && content["markdown"] != nil {
		// Mantener el markdown existente
	}

	wrapper["content"] = content

	b, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		original.Text = newPlain
		return original
	}

	original.Text = string(b)
	return original
}

// parseTags extrae etiquetas [[tag]] como array de strings
func parseTags(raw string) []string {
	var tags []string
	parts := strings.Fields(raw)
	for _, p := range parts {
		if strings.HasPrefix(p, "[[") && strings.HasSuffix(p, "]]") {
			tag := strings.TrimPrefix(p, "[[")
			tag = strings.TrimSuffix(tag, "]]")
			tags = append(tags, tag)
		}
	}
	return tags
}

// hashSHA256 calcula el hash SHA-256 de un string y lo devuelve en hex
func hashSHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// looksLikeJSON detecta si un string parece ser un objeto o array JSON serializado
func looksLikeJSON(s string) bool {
	s = strings.TrimSpace(s)
	return len(s) > 1 && (s[0] == '{' || s[0] == '[')
}

// (Eliminado: lógica de sincronización de tags fuera de función, lo cual no es válido en Go)
