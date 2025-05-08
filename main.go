package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Definición de flags para entrada y salida
	inPath := flag.String("input", "", "Ruta al JSON de TiddlyWiki")
	outPath := flag.String("output", "", "Ruta al JSONL de salida")
	flag.Parse()

	// Validación de parámetros
	if *inPath == "" || *outPath == "" {
		fmt.Println("Uso: openpages-source -input <tiddlers.json> -output <salida.jsonl>")
		os.Exit(1)
	}

	// TODO: Leer tiddlers desde *inPath usando ReadTiddlers
	// TODO: Convertir cada Tiddler a Record con ConvertToRecord
	// TODO: Escribir registros en formato JSONL a *outPath usando WriteJSONL

	// Mensaje final
	fmt.Printf("✅ Exportación completada: %s\n", *outPath)
}
