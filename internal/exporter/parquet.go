package exporter

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

// ParquetNode representa el esquema Parquet alineado al diseño semántico.
type ParquetNode struct {
	ID           string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Rol          string `parquet:"name=rol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Tags         string `parquet:"name=tags, type=BYTE_ARRAY, convertedtype=UTF8"`
	Content      string `parquet:"name=content_plain, type=BYTE_ARRAY, convertedtype=UTF8"`
	Define       string `parquet:"name=define, type=BYTE_ARRAY, convertedtype=UTF8"`
	Requiere     string `parquet:"name=requiere, type=BYTE_ARRAY, convertedtype=UTF8"`
	IsAIReady    bool   `parquet:"name=is_ai_ready, type=BOOLEAN"`
	HasRelations bool   `parquet:"name=has_relations, type=BOOLEAN"`
}

// MapRecordToParquet convierte un registro JSONL genérico a ParquetNode.
// Admite map[string]interface{} para máxima compatibilidad.
func MapRecordToParquet(m map[string]interface{}) ParquetNode {
	// Helper para extraer string seguro
	getStr := func(key string) string {
		if v, ok := m[key]; ok && v != nil {
			return fmt.Sprint(v)
		}
		return ""
	}
	// Helper para extraer y aplanar listas
	flattenList := func(key string) string {
		if v, ok := m[key]; ok && v != nil {
			switch vv := v.(type) {
			case []interface{}:
				var out []string
				for _, item := range vv {
					out = append(out, fmt.Sprint(item))
				}
				return strings.Join(out, ",")
			case string:
				return vv
			}
		}
		return ""
	}
	// Relaciones: define y requiere
	var define, requiere string
	if rels, ok := m["relations"].(map[string]interface{}); ok {
		if d, ok := rels["define"]; ok {
			define = flattenListFromAny(d)
		}
		if r, ok := rels["requiere"]; ok {
			requiere = flattenListFromAny(r)
		}
	}
	// Si relations es plano (string o lista)
	if define == "" {
		define = flattenList("define")
	}
	if requiere == "" {
		requiere = flattenList("requiere")
	}
	// is_ai_ready: heurística simple (tiene id, rol, content)
	isAIReady := getStr("id") != "" && getStr("rol") != "" && getStr("contentPlain") != ""
	// has_relations: define o requiere no vacío
	hasRelations := define != "" || requiere != ""

	return ParquetNode{
		ID:           getStr("id"),
		Rol:          getStr("rol"),
		Tags:         flattenList("tags"),
		Content:      getStr("contentPlain"),
		Define:       define,
		Requiere:     requiere,
		IsAIReady:    isAIReady,
		HasRelations: hasRelations,
	}
}

// flattenListFromAny convierte cualquier lista a string separado por coma.
func flattenListFromAny(val interface{}) string {
	switch vv := val.(type) {
	case []interface{}:
		var out []string
		for _, item := range vv {
			out = append(out, fmt.Sprint(item))
		}
		return strings.Join(out, ",")
	case string:
		return vv
	}
	return ""
}

// ConvertJSONLToParquet convierte un archivo .jsonl a .parquet alineado al diseño semántico.
func ConvertJSONLToParquet(inputPath string, outputPath string) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("abrir input: %w", err)
	}
	defer f.Close()

	// Crear archivo Parquet
	fw, err := local.NewLocalFileWriter(outputPath)
	if err != nil {
		return fmt.Errorf("crear parquet: %w", err)
	}
	defer fw.Close()

	pw, err := writer.NewParquetWriter(fw, new(ParquetNode), 4)
	if err != nil {
		return fmt.Errorf("parquet writer: %w", err)
	}
	pw.RowGroupSize = 128 * 1024 * 1024 // 128MB
	pw.CompressionType = parquet.CompressionCodec_SNAPPY

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		count++

		// DEBUG: Imprimir línea problemática
		if count >= 30 && count <= 40 {
			fmt.Printf("DEBUG línea %d: %s\n", count, line[:min(100, len(line))])
		}

		var m map[string]interface{}
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			fmt.Printf("ERROR línea %d: %s\n", count, line)
			return fmt.Errorf("jsonl línea %d: %w", count, err)
		}
		node := MapRecordToParquet(m)
		if err := pw.Write(node); err != nil {
			return fmt.Errorf("escribiendo parquet línea %d: %w", count, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("leer jsonl: %w", err)
	}
	if count == 0 {
		return errors.New("no se encontraron registros en el JSONL")
	}
	if err := pw.WriteStop(); err != nil {
		return fmt.Errorf("cerrar parquet: %w", err)
	}
	fmt.Printf("✅ Exportación Parquet completada: %d registros → %s\n", count, outputPath)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
