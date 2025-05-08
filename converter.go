// converter.go – Convertidores entre modelos Tiddler y Record
// -----------------------------------------------------------
// Ubicación: raíz del proyecto, junto a main.go, reader.go y writer.go.
// Responsabilidad: mapear cada Tiddler a un Record listo para JSONL,
// siguiendo Principios de Programación: separación de responsabilidades,
// manejo claro de errores y funciones con una única responsabilidad.
package main

import (
	"openpages-source/models"
)

// ConvertToRecord transforma un Tiddler en un Record estructurado.
//  - Preserva metadatos (ID, ContentType, Tags).
//  - Convierte Text a Markdown y Plain Text.
//  - Formatea fechas al estándar ISO 8601.
func ConvertToRecord(t models.Tiddler) models.Record {
	return models.Record{
		ID:           Slugify(t.Title),          // slug URL-friendly
		Title:        t.Title,                   // título original
		ContentType:  t.Type,                    // tipo MIME original
		Tags:         t.Tags,                    // etiquetas asociadas
		TextMarkdown: t.Text,                    // texto con Markdown intacto
		TextPlain:    StripMarkdown(t.Text),     // texto limpio para búsqueda
		CreatedAt:    FormatDateISO(t.Created),  // fecha en ISO 8601
		ModifiedAt:   FormatDateISO(t.Modified), // fecha en ISO 8601
	}
}

// Slugify genera un identificador legible y seguro para URLs a partir de un texto.
// Ejemplo: "Título Ejemplo" → "titulo-ejemplo"
func Slugify(input string) string {
	// TODO: implementar normalización Unicode, reemplazo de espacios por guiones,
	// eliminación de caracteres no alfanuméricos, minúsculas.
	return input // placeholder
}

// StripMarkdown elimina la sintaxis Markdown dejando solo texto plano.
// Esto mejora la indexación y el análisis de contenido.
func StripMarkdown(input string) string {
	// TODO: usar expresiones regulares o una librería ligera para remover
	// enlaces, énfasis, encabezados y otros tokens Markdown.
	return input // placeholder
}

// FormatDateISO convierte la fecha de TiddlyWiki (string) a ISO 8601 UTC.
// Asume una entrada tipo "20080130090807000" y retorna "2008-01-30T09:08:07Z".
func FormatDateISO(input string) string {
	// TODO: parsear según el formato TiddlyWiki y formatear con time.Format(time.RFC3339)
	return input // placeholder
}
