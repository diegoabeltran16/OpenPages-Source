// Package models define las estructuras de datos centrales para la Vuelta 1 del pipeline OpenPages-Source.
// Aquí modelamos los Tiddlers que provienen del JSON exportado de TiddlyWiki, incluyendo su tipo de contenido MIME.
package models

// Tiddler representa un elemento básico extraído de un JSON de TiddlyWiki.
// Cada campo refleja la información original que encontramos en el tiddler:
//   - Title:    El título legible para humanos del tiddler.
//   - Text:     El contenido bruto (puede estar en diversos formatos: Markdown, HTML, texto plano, etc.).
//   - Type:     El tipo de contenido MIME del tiddler (p.ej. text/markdown, text/html, image/png).
//   - Tags:     Etiquetas asociadas para clasificación.
//   - Created:  Fecha de creación en formato TiddlyWiki.
//   - Modified: Fecha de última modificación en formato TiddlyWiki.
type Tiddler struct {
    // Title es el nombre único del tiddler.
    Title string `json:"title"`

    // Text contiene el cuerpo del tiddler, tal como fue exportado.
    Text string `json:"text"`

    // Type indica el tipo de contenido MIME del tiddler.
    Type string `json:"type"`

    // Tags agrupa palabras clave para filtrado o agrupación.
    Tags []string `json:"tags"`

    // Created registra la fecha de creación tal cual la exporta TiddlyWiki.
    Created string `json:"created"`

    // Modified registra la última vez que se editó este tiddler.
    Modified string `json:"modified"`
}
