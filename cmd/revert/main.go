package main

import (
	"log"
	"os"

	"github.com/diegoabeltran16/OpenPages-Source/internal/exporter"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("Uso: revert <plantilla.json> <textos.jsonl> <salida.json>")
	}
	textos := os.Args[2]
	salida := os.Args[3]

	err := exporter.ExportAllFromJSONL(textos, salida)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("✅ Proceso completado: %s actualizado con textos desde %s", salida, textos)
}
