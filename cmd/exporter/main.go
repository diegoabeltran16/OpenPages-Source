// cmd/exporter/main.go – Orquestador principal del pipeline (v1, v2 y v3)
// --------------------------------------------------------------------------------
// Contexto pedagógico
// -------------------
//   1. Parsear flags: -input, -output, -mode (v1|v2|v3), -pretty
//   2. Si input es carpeta, buscar el primer .json adentro.
//   3. Si output es carpeta o no existe sin extensión, crear carpeta y usar out.jsonl dentro.
//   4. Llamar a importer.Read → transform.ConvertTiddlers{V1,V2,V3} → exporter.WriteJSONL
//   5. Mostrar mensajes en consola y manejar errores.
//
// Ejemplos de uso:
//   go run ./cmd/exporter \
//     -input ./data/in \
//     -output ./data/out \
//     -mode v3
//
//   go run ./cmd/exporter \
//     -input ./data/in/tiddlers.json \
//     -output ./data/out/tiddlers.jsonl \
//     -mode v1 -pretty
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

	// 1) Flags CLI
	in := flag.String("input", "", "Archivo o carpeta con JSON exportado de TiddlyWiki (requerido)")
	out := flag.String("output", "", "Ruta de salida: archivo .jsonl o carpeta (requerido)")
	mode := flag.String("mode", "v1", "Modo de conversión: v1 (plano) | v2 (meta/content) | v3 (JSONL mínimo)")
	pretty := flag.Bool("pretty", false, "Usar indentación en lugar de JSONL compacto")
	reverse := flag.Bool("reverse", false, "Revertir JSONL enriquecido a JSON TiddlyWiki")
	reverseSingle := flag.Bool("reverse-single", false, "Revertir solo el tiddler raíz a objeto único")
	rootTitle := flag.String("root-title", "_____Nombre del Proyecto", "Título del tiddler raíz para reversión")
	updateTexts := flag.Bool("update-texts", false, "Actualizar solo los campos 'text' y 'modified' en la plantilla usando otro archivo")
	flag.Parse()

	// **NUEVO: Actualizar texts en plantilla**
	if *updateTexts {
		if *in == "" || *out == "" {
			fmt.Println("Uso: exporter -update-texts -input plantilla.json -output destino.json -updates actualizaciones.json")
			os.Exit(1)
		}
		updatesPath := flag.Lookup("updates")
		if updatesPath == nil || updatesPath.Value.String() == "" {
			fmt.Println("Falta el flag -updates con el archivo de actualizaciones")
			os.Exit(1)
		}
		template, err := importer.Read(ctx, *in)
		if err != nil {
			log.Fatalf("❌ error leyendo plantilla: %v", err)
		}
		updates, err := importer.Read(ctx, updatesPath.Value.String())
		if err != nil {
			log.Fatalf("❌ error leyendo actualizaciones: %v", err)
		}
		result := transform.UpdateTexts(template, updates)
		if err := exporter.WriteJSON(*out, result, *pretty); err != nil {
			log.Fatalf("❌ error escribiendo resultado: %v", err)
		}
		fmt.Printf("✅ Actualización de texts completada (destino: %s)\n", *out)
		return
	}

	// **NUEVO: Modo reversa objeto único**
	if *reverseSingle {
		if *in == "" || *out == "" {
			fmt.Println("Uso: exporter -reverse-single -input archivo.json -output destino.json -root-title \"_____Nombre del Proyecto\"")
			os.Exit(1)
		}
		if err := exporter.RevertToSingleTiddler(ctx, *in, *out, *rootTitle); err != nil {
			log.Fatalf("❌ error en reversa objeto único: %v", err)
		}
		fmt.Printf("✅ Reversión objeto único completada (destino: %s)\n", *out)
		return
	}

	// **NUEVO: Modo reversa array**
	if *reverse {
		if *in == "" || *out == "" {
			fmt.Println("Uso: exporter -reverse -input archivo.jsonl -output destino.json")
			os.Exit(1)
		}
		if err := transform.ReverseJSONLToTiddlyJSON(*in, *out); err != nil {
			log.Fatalf("❌ error en reversa: %v", err)
		}
		fmt.Printf("✅ Reversión completada (destino: %s)\n", *out)
		return
	}

	// 2) Validar obligatorio para modo normal
	if *in == "" || *out == "" {
		fmt.Println("Uso: exporter -input origen.json|carpeta -output destino.jsonl|carpeta [-mode v2|v3] [-pretty]")
		os.Exit(1)
	}

	// 3) Resolver input (archivo o directorio)
	fi, err := os.Stat(*in)
	if err != nil {
		log.Fatalf("❌ no se pudo acceder a '%s': %v", *in, err)
	}
	if fi.IsDir() {
		files, err := os.ReadDir(*in)
		if err != nil {
			log.Fatalf("❌ no se pudo listar '%s': %v", *in, err)
		}
		found := false
		for _, f := range files {
			if !f.IsDir() {
				*in = filepath.Join(*in, f.Name())
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("❌ no se encontró ningún archivo en '%s'", *in)
		}
	}

	// 4) Resolver output (archivo o carpeta)
	fo, err := os.Stat(*out)
	if err == nil && fo.IsDir() {
		// Si existe y es carpeta: usar archivo por defecto dentro
		*out = filepath.Join(*out, "out.jsonl")
	} else if os.IsNotExist(err) && filepath.Ext(*out) == "" {
		// Si no existe y no tiene extensión: crear carpeta
		if mkdirErr := os.MkdirAll(*out, 0755); mkdirErr != nil {
			log.Fatalf("❌ no se pudo crear carpeta '%s': %v", *out, mkdirErr)
		}
		*out = filepath.Join(*out, "out.jsonl")
	}

	// 5) Leer tiddlers
	tiddlers, err := importer.Read(ctx, *in)
	if err != nil {
		log.Fatalf("❌ error leyendo tiddlers: %v", err)
	}
	fmt.Printf("📦 %d tiddlers cargados\n", len(tiddlers))

	// 6) Convertir y exportar según modo
	switch *mode {
	case "v3":
		recs := transform.ConvertTiddlersV3(tiddlers)
		if err := exporter.WriteJSONL(ctx, *out, recs, *pretty); err != nil {
			log.Fatalf("❌ escribir JSONL v3: %v", err)
		}

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
		log.Fatalf("❌ modo desconocido: %s (usa 'v1', 'v2' o 'v3')", *mode)
	}

	fmt.Printf("✅ Exportación completada (destino: %s)\n", *out)
}
