// main.go – Orquestador principal del pipeline
// --------------------------------------------------------------------------------
// Contexto pedagógico
// -------------------
// Este archivo *amarra* los tres componentes del pipeline: *reader.go*,
// *converter.go* y *writer.go*.
//
// --------------------------------------------------------------------------------
// RESPONSABILIDAD PRINCIPAL
// --------------------------------------------------------------------------------
// 1. **Parsear flags**:  `-input` para el export JSON de TiddlyWiki y `-output`
//    para el archivo destino JSONL.
// 2. Validar que ambos argumentos existan; si no, mostrar *usage* y abortar.
// 3. Orquestar:
//      • Leer tiddlers       → `ReadTiddlers`.
//      • Convertir a records → `ConvertTiddlers`.
//      • Escribir JSONL      → `WriteJSONL`.
// 4. Reportar progreso y errores de forma amigable.
//
// --------------------------------------------------------------------------------
// CÓMO COMPILAR Y EJECUTAR
// --------------------------------------------------------------------------------
//   go run ./cmd/exporter \
//     -input /home/naveen/Documents/OpenPages-Source/data/in/tiddlers.json \
//     -output /home/naveen/Documents/OpenPages-Source/data/out/tiddlers.jsonl
//
// --------------------------------------------------------------------------------

// cmd/exporter/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/diegoabeltran16/OpenPages-Source/internal/exporter"
	"github.com/diegoabeltran16/OpenPages-Source/internal/importer"
	"github.com/diegoabeltran16/OpenPages-Source/internal/transform"
)

func main() {
	ctx := context.Background()

	// ------------------------------------------------------------ Flags
	in := flag.String("input", "", "JSON exportado de TiddlyWiki")
	out := flag.String("output", "", "Archivo JSONL de salida")
	mode := flag.String("mode", "v1", "v1 | v2  (estructura del JSONL)")
	pretty := flag.Bool("pretty", false, "MarshalIndent en lugar de compacto")
	flag.Parse()

	if *in == "" || *out == "" {
		fmt.Println("Uso: exporter -input tiddlers.json -output sal.jsonl [-mode v2]")
		os.Exit(1)
	}

	// ------------------------------------------------------ Leer tiddlers
	tiddlers, err := importer.Read(ctx, *in)
	if err != nil {
		log.Fatalf("❌ error leyendo tiddlers: %v", err)
	}
	fmt.Printf("📦 %d tiddlers cargados\n", len(tiddlers))

	// -------------------------------------------------- Convertir según modo
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
		log.Fatalf("modo desconocido: %s (use v1 o v2)", *mode)
	}

	fmt.Printf("✅ Exportación completada (%s)\n", *out)
}
