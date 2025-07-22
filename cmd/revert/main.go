package main

import (
	"log"

	"github.com/diegoabeltran16/OpenPages-Source/internal/exporter"
)

func main() {
	err := exporter.CloneAndUpdateTexts(
		"data/in/Plantilla (Estudiar OpenPages).json", // plantilla original
		"data/out/tiddlers.jsonl",                     // textos nuevos
		"data/in/tiddlers_revert.json",                // salida clonada y actualizada
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("✅ Proceso completado: tiddlers_revert.json actualizado con textos desde JSONL")
}
