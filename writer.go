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

	"openpages-source/models"
)

// WriteJSONL crea (o sobrescribe) el archivo en 'path' y escribe
// cada Record como una línea válida de JSONL.
// Parámetros:
//   - path:   ruta del archivo de salida (.jsonl).
//   - records: slice de modelos.Record a serializar.
// Retorna:
//   - error: si ocurre un fallo al crear el archivo, serializar o escribir.
func WriteJSONL(path string, records []models.Record) error {
	// 1) Crear (o truncar) el archivo de salida
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("no se pudo crear el archivo '%s': %w", path, err)
	}
	defer file.Close()

	// 2) Preparar un escritor buffered para eficiencia
	writer := bufio.NewWriter(file)

	// 3) Iterar sobre cada registro, serializar y escribir
	for i, rec := range records {
		// 3.a) Convertir el Record a JSON
		lineBytes, err := json.Marshal(rec)
		if err != nil {
			return fmt.Errorf("error al serializar Record en posición %d: %w", i, err)
		}

		// 3.b) Escribir la línea JSON y un salto de línea
		if _, err := writer.Write(lineBytes); err != nil {
			return fmt.Errorf("error al escribir Record JSON en posición %d: %w", i, err)
		}
		if err := writer.WriteByte('\n'); err != nil {
			return fmt.Errorf("error al escribir salto de línea en posición %d: %w", i, err)
		}
	}

	// 4) Asegurar que todos los datos buffered se hayan volcado al disco
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error al vaciar el buffer de escritura: %w", err)
	}

	return nil
}
