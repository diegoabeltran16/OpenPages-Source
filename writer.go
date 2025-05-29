// writer.go – Escritura de registros en formato JSONL
// -----------------------------------------------------------
// Ubicación: raíz del proyecto.
// Responsabilidad: tomar un slice de models.Record y escribirlo
// en un archivo de salida en formato JSONL (una línea JSON por registro),
// con manejo eficiente de buffer y errores claros.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// WriteJSONL crea (o sobrescribe) el archivo en 'path' y escribe
// cada Record como una línea válida de JSONL.
// Parámetros:
//   - path:   ruta del archivo de salida (.jsonl).
//   - records: slice de models.Record a serializar.
//
// Retorna:
//   - error: si ocurre un fallo al crear el archivo, serializar o escribir.
func WriteJSONL(path string, records []models.Record) error {
	// Crear o truncar el archivo de salida
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("no se pudo crear el archivo '%s': %w", path, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("error al cerrar el archivo '%s': %w", path, cerr)
		}
	}()

	writer := bufio.NewWriter(file)
	// Iterar sobre cada registro y escribir en JSONL
	for i, rec := range records {
		line, err := json.Marshal(rec)
		if err != nil {
			return fmt.Errorf("error al serializar Record en posición %d: %w", i, err)
		}

		if _, err := writer.Write(line); err != nil {
			return fmt.Errorf("error al escribir JSON en posición %d: %w", i, err)
		}
		if err := writer.WriteByte('\n'); err != nil {
			return fmt.Errorf("error al escribir salto de línea en posición %d: %w", i, err)
		}
	}

	// Asegurar que todo se escribe en disco
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error al vaciar el buffer de escritura: %w", err)
	}

	return nil
}
