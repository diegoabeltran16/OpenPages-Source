// Package models define las estructuras de datos resultantes para la Vuelta 1 del pipeline OpenPages-Source.
// Aquí convertimos cada Tiddler en un Record listo para exportar en formato JSONL,
// preservando el tipo de contenido MIME para su procesamiento downstream.
package models

// Record representa la forma final de cada entrada en el archivo JSONL.
// Incluye transformaciones útiles y metadatos necesarios:
//   - ID:           Identificador único y legible (slug) basado en el título.
//   - Title:        Título original del tiddler.
//   - ContentType:  Tipo MIME original (p.ej. text/markdown, image/png).
//   - Tags:         Etiquetas asociadas para filtrado o agrupación.
//   - TextMarkdown: Contenido original en Markdown (si aplica).
//   - TextPlain:    Versión en texto plano sin sintaxis Markdown.
//   - CreatedAt:    Fecha de creación en ISO 8601 (UTC).
//   - ModifiedAt:   Fecha de última modificación en ISO 8601 (UTC).
type Record struct {
    // ID es un "slug" único, generado a partir del título.
    ID string `json:"id"`

    // Title conserva el título original del tiddler.
    Title string `json:"title"`

    // ContentType indica el tipo de contenido MIME del tiddler,
    // asegurando que quien consuma el JSONL conozca el formato original.
    ContentType string `json:"content_type"`

    // Tags agrupa las etiquetas asociadas para clasificación.
    Tags []string `json:"tags"`

    // TextMarkdown mantiene el contenido original en Markdown,
    // útil para reprocesar o renderizar.
    TextMarkdown string `json:"text_markdown"`

    // TextPlain contiene una versión limpia sin sintaxis Markdown,
    // ideal para indexación de texto completo o análisis de sentimiento.
    TextPlain string `json:"text_plain"`

    // CreatedAt es la fecha de creación convertida a ISO 8601 (UTC).
    CreatedAt string `json:"created_at"`

    // ModifiedAt es la fecha de última modificación en ISO 8601 (UTC).
    ModifiedAt string `json:"modified_at"`
}
