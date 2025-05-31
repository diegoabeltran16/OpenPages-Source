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

	// paquetes internos – sólo visibles dentro del módulo
	"github.com/diegoabeltran16/OpenPages-Source/internal/exporter"
	"github.com/diegoabeltran16/OpenPages-Source/internal/importer"
	"github.com/diegoabeltran16/OpenPages-Source/internal/transform"
)

func main() {
	ctx := context.Background()

	// -----------------------------------------------------------------
	// 1) Flags CLI
	// -----------------------------------------------------------------
	in := flag.String("input", "", "JSON exportado de TiddlyWiki")
	out := flag.String("output", "", "Archivo JSONL de salida")
	flag.Parse()

	if *in == "" || *out == "" {
		fmt.Println("❌ Uso: exporter -input export.json -output salida.jsonl")
		os.Exit(1)
	}

	// -----------------------------------------------------------------
	// 2) Leer tiddlers
	// -----------------------------------------------------------------
	tiddlers, err := importer.Read(ctx, *in)
	if err != nil {
		log.Fatalf("❌ error leyendo tiddlers: %v", err)
	}
	fmt.Printf("📦 %d tiddlers cargados\n", len(tiddlers))

	// -----------------------------------------------------------------
	// 3) Convertir a records
	// -----------------------------------------------------------------
	records := transform.ConvertTiddlers(tiddlers)

	// -----------------------------------------------------------------
	// 4) Escribir JSONL
	// -----------------------------------------------------------------
	if err := exporter.WriteJSONL(ctx, *out, records); err != nil {
		log.Fatalf("❌ error escribiendo JSONL: %v", err)
	}

	fmt.Printf("✅ Exportación completada: %s (%d registros)\n", *out, len(records))
}
