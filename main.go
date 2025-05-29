// main.go – Orquestador principal del pipeline
// ----------------------------------------------
// Ubicación: raíz del proyecto.
// Responsabilidad: coordinar lectura de Tiddlers, conversión y escritura en JSONL.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// Definir flags de entrada y salida
	inputPath := flag.String("input", "", "Ruta al archivo JSON exportado de TiddlyWiki")
	outputPath := flag.String("output", "", "Ruta al archivo JSONL de salida")
	flag.Parse()

	// Validar argumentos
	if *inputPath == "" || *outputPath == "" {
		fmt.Println("❌ Uso: go run main.go -input archivo.json -output salida.jsonl")
		os.Exit(1)
	}

	// 1) Leer tiddlers desde JSON
	tiddlers, err := ReadTiddlers(*inputPath)
	if err != nil {
		log.Fatalf("❌ Error al leer tiddlers: %v", err)
	}
	fmt.Printf("📦 Se cargaron %d tiddlers desde '%s'\n", len(tiddlers), *inputPath)

	// 2) Convertir a records
	records := ConvertTiddlers(tiddlers)

	// 3) Escribir registro en JSONL
	if err := WriteJSONL(*outputPath, records); err != nil {
		log.Fatalf("❌ Error al escribir salida JSONL: %v", err)
	}

	// 4) Mensaje final
	fmt.Printf("✅ Exportación completada: %s (%d tiddlers exportados)\n", *outputPath, len(records))
}
