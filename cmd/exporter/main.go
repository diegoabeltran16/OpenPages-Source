// cmd/exporter/main.go – Exportador con deduplicación por hash (versión corregida)
// ----------------------------------------------------------------------------
// Lee un archivo de tiddlers (.json), aplica deduplicación por hash
// (Título + Modified + Texto), utiliza ConvertTiddlersV3 para transformar
// cada tiddler individual al modelo RecordV2 y finalmente guarda todo en
// JSONL (o JSON indented si se solicita).
// ----------------------------------------------------------------------------

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/diegoabeltran16/OpenPages-Source/internal/dedup"
	"github.com/diegoabeltran16/OpenPages-Source/internal/exporter"
	"github.com/diegoabeltran16/OpenPages-Source/internal/importer"
	"github.com/diegoabeltran16/OpenPages-Source/internal/transform"
	"github.com/diegoabeltran16/OpenPages-Source/models"
)

func main() {
	// ------------------ 🔧 Flags CLI ------------------
	inPath := flag.String("input", "", "Ruta al archivo JSON exportado de TiddlyWiki (requerido)")
	outPath := flag.String("output", "", "Ruta al archivo JSONL de salida (requerido)")
	pretty := flag.Bool("pretty", false, "Si se establece, formatea cada JSON con indentación")
	flag.Parse()

	if *inPath == "" || *outPath == "" {
		fmt.Fprintf(os.Stderr, "Uso: %s -input <tiddlers.json> -output <salida.jsonl> [-pretty]\n",
			filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	// ------------------ 📥 Lectura del archivo de tiddlers ------------------
	ctx := context.Background()
	tiddlers, err := importer.Read(ctx, *inPath)
	if err != nil {
		log.Fatalf("❌ Error leyendo tiddlers desde '%s': %v", *inPath, err)
	}
	log.Printf("📦 %d tiddlers cargados", len(tiddlers))

	// ------------------ 🛠 Crear carpeta de cache para hashes ------------------
	cacheDir := "data/cache"
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		log.Fatalf("❌ No se pudo crear directorio '%s': %v", cacheDir, err)
	}

	// ------------------ 🧠 Deduplicación usando FileStore ------------------
	// Ahora que hemos creado data/cache, OpenFile puede crear hashes.txt allí.
	hashFile := filepath.Join(cacheDir, "hashes.txt")
	store, err := dedup.NewFileStore(hashFile)
	if err != nil {
		log.Fatalf("❌ No se pudo inicializar deduplicador: %v", err)
	}
	defer store.Close()

	var filteredRecords []models.RecordV2
	dedupedCount := 0

	for _, t := range tiddlers {
		// 1) Calcular hash usando título + modified + texto
		hash := dedup.HashTiddler(t)
		if store.Seen(hash) {
			dedupedCount++
			continue // Saltar tiddler ya visto
		}
		if err := store.Mark(hash); err != nil {
			log.Printf("⚠️  No se pudo registrar hash '%s': %v", hash, err)
			continue
		}

		// 2) Convertir este único tiddler a RecordV2 vía ConvertTiddlersV3:
		//    Pasamos un slice de longitud 1 y luego tomamos el [0].
		singleSlice := []models.Tiddler{t}
		recs := transform.ConvertTiddlersV3(singleSlice)
		// ConvertTiddlersV3 siempre retorna un slice de la misma longitud:
		// en este caso, len(recs) == 1.
		filteredRecords = append(filteredRecords, recs[0])
	}
	log.Printf("🧹 Deduplicación aplicada: %d descartados, %d a exportar",
		dedupedCount, len(filteredRecords))

	// ------------------ 📤 Escritura en JSONL / JSON indentado ------------------
	if err := exporter.WriteJSONL(ctx, *outPath, filteredRecords, *pretty); err != nil {
		log.Fatalf("❌ Error al escribir salida: %v", err)
	}
	log.Printf("✅ Exportación completada en '%s'", *outPath)
}
