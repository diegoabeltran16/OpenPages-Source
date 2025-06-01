# OpenPages Exporter

Exportador estructurado para transformar archivos de TiddlyWiki en formatos legibles por máquinas e integrables con herramientas modernas de análisis, indexación o inteligencia artificial.

## 🎯 Propósito

TiddlyWiki es una herramienta de conocimiento local, modular y sin dependencias externas. Sin embargo, sus exportaciones no están estructuradas para ser fácilmente consumidas por sistemas externos como:

- Modelos de lenguaje (LLMs)
- Pipelines de IA o ETL
- Dashboards de conocimiento
- Buscadores semánticos o clasificadores automáticos

Este proyecto permite convertir el contenido de TiddlyWiki en archivos `.jsonl` o `.json` estructurados, normalizando etiquetas, formateando metadatos y separando contenido de control semántico (meta) y contenido principal (content).

---

## 🧱 Componentes

### Entrada

- Un archivo `.json` exportado desde TiddlyWiki (preferiblemente vía plugin [TiddlyWiki Export](https://tiddlywiki.com/#Saving%20a%20wiki%20as%20JSON)).

### Salida

- Un archivo `.jsonl` válido (una línea por entrada).
- O un archivo `.json` indentado (solo para inspección humana).

---

## ⚙️ Cómo usarlo

### Opción 1: desde el binario o `go run`

bash
`go run ./cmd/exporter \ -input ./data/in/tiddlers.json \ -output ./data/out/tiddlers.jsonl \ -mode v2 \ -pretty`

### Opción 2: con script interactivo

Linux/macOS:

bash
`./scripts-frontend/run.sh`

Windows:

cmd
`scripts-frontend\run.bat`

Estos scripts guían al usuario para elegir:

- Modo de exportación (`v1` básico o `v2` con meta-content)
- Formato de salida (`.jsonl` plano o `.json` indentado)
- Rutas de entrada y salida personalizadas
  

---

## 🧠 Por qué importa

1. **Evita bloqueo de proveedor (vendor lock-in)**: los datos permanecen en formatos abiertos.
2. **Optimiza para IA**: cada entrada puede ser procesada línea por línea por modelos de lenguaje.
3. **Aumenta la interoperabilidad**: el contenido de TiddlyWiki puede integrarse con otros sistemas, sin depender de su visualizador HTML.

---
## 📦 Requisitos

- Go 1.20 o superior
- Archivo `.json` exportado desde TiddlyWiki (estructura válida)