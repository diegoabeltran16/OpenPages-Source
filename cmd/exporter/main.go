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
	mode := flag.String("mode", "v1", "Modo de conversión: v1 (plano) | v2 (meta/content) | v3 (JSONL mínimo) | hybrid (IA/RAG)")
	pretty := flag.Bool("pretty", false, "Usar indentación en lugar de JSONL compacto")
	reverse := flag.Bool("reverse", false, "Revertir JSONL enriquecido a JSON TiddlyWiki")
	reverseSingle := flag.Bool("reverse-single", false, "Revertir solo el tiddler raíz a objeto único")
	rootTitle := flag.String("root-title", "_____Nombre del Proyecto", "Título del tiddler raíz para reversión")
	updateTexts := flag.Bool("update-texts", false, "Actualizar solo los campos 'text' y 'modified' en la plantilla usando otro archivo")
	updates := flag.String("updates", "", "Archivo JSONL con actualizaciones de textos (para -update-texts)")
	flag.Parse()

	// 2) Modos especiales
	if *updateTexts {
		if *in == "" || *out == "" || *updates == "" {
			fmt.Println("Uso: exporter -update-texts -input plantilla.json -output destino.json -updates actualizaciones.json")
			os.Exit(1)
		}
		template, err := importer.Read(ctx, *in)
		if err != nil {
			log.Fatalf("❌ error leyendo plantilla: %v", err)
		}
		updatesData, err := importer.Read(ctx, *updates)
		if err != nil {
			log.Fatalf("❌ error leyendo actualizaciones: %v", err)
		}
		result := transform.UpdateTexts(template, updatesData)
		if err := exporter.WriteJSON(*out, result, *pretty); err != nil {
			log.Fatalf("❌ error escribiendo resultado: %v", err)
		}
		fmt.Printf("✅ Actualización de texts completada (destino: %s)\n", *out)
		return
	}

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

	// 3) Validar flags obligatorios
	if *in == "" || *out == "" {
		fmt.Println("Uso: exporter -input origen.json|carpeta -output destino.jsonl|carpeta [-mode v1|v2|v3|hybrid] [-pretty]")
		os.Exit(1)
	}

	// 4) Resolver input (archivo o directorio)
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

	// 5) Resolver output (archivo o carpeta)
	fo, err := os.Stat(*out)
	base := filepath.Base(*in)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	prettySuffix := ""
	if *pretty {
		prettySuffix = "_pretty"
	}
	if (err == nil && fo.IsDir()) || (os.IsNotExist(err) && filepath.Ext(*out) == "") {
		if os.IsNotExist(err) {
			if mkdirErr := os.MkdirAll(*out, 0755); mkdirErr != nil {
				log.Fatalf("❌ no se pudo crear carpeta '%s': %v", *out, mkdirErr)
			}
		}
		*out = filepath.Join(*out, fmt.Sprintf("%s_%s%s.jsonl", name, *mode, prettySuffix))
	} else if filepath.Ext(*out) == ".jsonl" {
		*out = filepath.Join(filepath.Dir(*out), fmt.Sprintf("%s_%s%s.jsonl", name, *mode, prettySuffix))
	}

	// 6) Leer tiddlers
	tiddlers, err := importer.Read(ctx, *in)
	if err != nil {
		log.Fatalf("❌ error leyendo tiddlers: %v", err)
	}
	fmt.Printf("📦 %d tiddlers cargados\n", len(tiddlers))

	// 7) Convertir y exportar según modo
	fmt.Println("--------------------------------------------------")
	fmt.Printf("🧠 Modo de exportación seleccionado: %s\n", *mode)
	switch *mode {
	case "v1":
		fmt.Println("  - Compacto heredado (TextPlain/TextMarkdown)")
	case "v2":
		fmt.Println("  - Meta + Content (AI-friendly, contexto rico)")
	case "v3":
		fmt.Println("  - Minimal JSONL (una línea por objeto, ideal para IA)")
	case "hybrid":
		fmt.Println("  - Híbrido (estructura extendida para IA/RAG)")
	default:
		fmt.Println("  - Modo desconocido")
	}
	fmt.Printf("📦 Formato de salida: %s\n", func() string {
		if *pretty {
			return "JSON indentado (multilínea, inspección humana)"
		}
		return "JSONL plano (una línea por objeto, ingestión IA)"
	}())
	fmt.Printf("📥 Archivo de entrada: %s\n", *in)
	fmt.Printf("📤 Archivo de salida:  %s\n", *out)
	fmt.Println("--------------------------------------------------")

	switch *mode {
	case "hybrid":
		recs := transform.ConvertTiddlersHybrid(tiddlers)
		if err := exporter.WriteJSONL(ctx, *out, recs, *pretty); err != nil {
			log.Fatalf("❌ escribir JSONL hybrid: %v", err)
		}

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
		log.Fatalf("❌ modo desconocido: %s (usa 'v1', 'v2', 'v3' o 'hybrid')", *mode)
	}

	fmt.Printf("✅ Exportación completada (destino: %s)\n", *out)
}
