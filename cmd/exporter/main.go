// main.go – Orquestador principal del pipeline
// --------------------------------------------------------------------------------
// Contexto pedagógico
// -------------------
// Este archivo conecta los tres componentes del pipeline: `reader.go`,
// `converter.go` y `writer.go`, permitiendo ejecutar la transformación
// desde línea de comandos con flexibilidad.
//
// --------------------------------------------------------------------------------
// RESPONSABILIDADES
// --------------------------------------------------------------------------------
// 1. Parsear flags: `-input`, `-output`, `-mode`, `-pretty`
// 2. Validar y resolver rutas (archivo o directorio)
// 3. Leer tiddlers → Convertir → Exportar como JSONL
// 4. Manejar errores de forma amigable
//
// --------------------------------------------------------------------------------
// CÓMO EJECUTAR (ejemplos)
// --------------------------------------------------------------------------------
// go run ./cmd/exporter \
//   -input ./data/in \
//   -output ./data/out \
//   -mode v2
//
// go run ./cmd/exporter \
//   -input ./data/in/tiddlers.json \
//   -output ./data/out/tiddlers.jsonl \
//   -mode v1 -pretty
// --------------------------------------------------------------------------------

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/diegoabeltran16/OpenPages-Source/internal/exporter"
	"github.com/diegoabeltran16/OpenPages-Source/internal/importer"
	"github.com/diegoabeltran16/OpenPages-Source/internal/transform"
)

func main() {
	ctx := context.Background()

	// ------------------------------------------------------------ Flags CLI
	in := flag.String("input", "", "Archivo o carpeta con JSON exportado de TiddlyWiki")
	out := flag.String("output", "", "Ruta de salida: archivo .jsonl o carpeta")
	mode := flag.String("mode", "v1", "Modo de conversión: v1 (plano) | v2 (meta/content)")
	pretty := flag.Bool("pretty", false, "Usar indentación en lugar de JSONL compacto")
	flag.Parse()

	// ------------------------------ Validar argumentos obligatorios
	if *in == "" || *out == "" {
		fmt.Println("Uso: exporter -input origen.json|carpeta -output destino.jsonl|carpeta [-mode v2]")
		os.Exit(1)
	}

	// ------------------------------ Resolver input (archivo o directorio)
	fi, err := os.Stat(*in)
	if err != nil {
		log.Fatalf("❌ no se pudo acceder a '%s': %v", *in, err)
	}
	if fi.IsDir() {
		files, err := os.ReadDir(*in)
		if err != nil {
			log.Fatalf("❌ no se pudo listar archivos en '%s': %v", *in, err)
		}
		found := false
		for _, f := range files {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
				*in = filepath.Join(*in, f.Name())
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("❌ no se encontró ningún archivo .json en la carpeta '%s'", *in)
		}
	}

	// ------------------------------ Resolver output (archivo o carpeta)
	fo, err := os.Stat(*out)
	if err == nil && fo.IsDir() {
		// Si existe y es carpeta: usar archivo por defecto dentro
		*out = filepath.Join(*out, "out.jsonl")
	} else if os.IsNotExist(err) && filepath.Ext(*out) == "" {
		// Si no existe y no tiene extensión: crear carpeta
		if err := os.MkdirAll(*out, 0755); err != nil {
			log.Fatalf("❌ no se pudo crear carpeta de salida: %v", err)
		}
		*out = filepath.Join(*out, "out.jsonl")
	}

	// ------------------------------ Leer tiddlers
	tiddlers, err := importer.Read(ctx, *in)
	if err != nil {
		log.Fatalf("❌ error leyendo tiddlers: %v", err)
	}
	fmt.Printf("📦 %d tiddlers cargados\n", len(tiddlers))

	// ------------------------------ Convertir y exportar según modo
	switch *mode {
	case "v2":
		recs := transform.ConvertTiddlersV2(tiddlers)
		if err := exporter.WriteJSONL(ctx, *out, recs, *pretty); err != nil {
			log.Fatalf("❌ escribir JSONL v2: %v", err)
		}
	case "v1":
		recs := transform.ConvertTiddlers(tiddlers)
		if err := exporter.WriteJSONL(ctx, *out, recs, *pretty); err != nil {
			log.Fatalf("❌ escribir JSONL v1: %v", err)
		}
	default:
		log.Fatalf("❌ modo desconocido: %s (usa 'v1' o 'v2')", *mode)
	}

	fmt.Printf("✅ Exportación completada (%s)\n", *out)
}
