# OpenPages Exporter

Convierte una exportación **JSON** de *TiddlyWiki* en formatos **JSONL** o **JSON** listos para ingestión por:

- Modelos de lenguaje (LLMs)
- Pipelines de IA o ETL
- Dashboards de conocimiento
- Buscadores semánticos, clasificadores automáticos o motores RAG

La herramienta preserva la semántica de cada *tiddler* (título, tags, fechas, color, relaciones) y ofrece **tres modos** de salida, priorizando compatibilidad con sistemas de gran escala y pipelines de IA:

| Modo   | Esquema                                              | Pensado para…                                   | Formato         |
|--------|------------------------------------------------------|-------------------------------------------------|-----------------|
| `v2`   | **Meta + Content**<br>`models.RecordV2`              | LLMs, RAG, BigQuery, Spark, ingestión masiva    | JSONL / JSON    |
| `hybrid` | **10 claves ricas por entrada**<br>texto claro     | LLMs, RAG, análisis semántico                   | JSONL           |
| `v1`   | **Compacto heredado**<br>TextPlain / TextMarkdown    | Scripts antiguos, pruebas semánticas, embedding | JSONL           |
| `v2_pretty` | **Meta + Content (multilínea, inspección)**     | Revisión visual, limpieza previa a ingestión    | JSON (pretty)   |
| `v1_pretty` | **Compacto heredado (multilínea, escapado)**    | Inspección humana, requiere preprocesamiento    | JSON (pretty)   |

> **Nota:** Los modos `pretty` no son JSONL estricto, pero pueden limpiarse fácilmente para ingestión automática. El modo `v2` es el recomendado para sistemas de IA y pipelines modernos.

## 🧱 Componentes

### Entrada

- Un archivo `.json` exportado desde TiddlyWiki (preferiblemente vía [TiddlyWiki Export](https://tiddlywiki.com)).

### Salida

- Archivos `.jsonl` válido (una línea por entrada).
- O un archivo `.jsonl` indentado (solo para inspección humana)(formato pretty).

---

## ⚙️ Cómo usarlo

```powershell
# Necesitas Go ≥ 1.20
go install github.com/diegoabeltran16/OpenPages-Source/cmd/exporter@latest

# o clona el repo y compila:
git clone https://github.com/diegoabeltran16/OpenPages-Source.git

cd OpenPages-Source
go build -o openpages_exporter.exe ./cmd/exporter
```

---

## ⚙️ Uso básico

```powershell
# JSONL estricto (una línea por objeto)
.\openpages_exporter.exe `
  -input  data\in\tiddlers.json `
  -output data\out\tiddlers.jsonl `
  -mode   v3              # v1 | v2 | v3

# Para inspección humana (indentado multilínea)
.\openpages_exporter.exe -input … -output pretty.json -mode v2 -pretty
```

### Script interactivo

- **Linux / macOS**

  ```bash
  ./scripts-frontend/run.sh
  ```

- **Windows**

  ```cmd
  scripts-frontend\run.bat
  ```

Los scripts detectan el primer `.json` en `data/in/`, te preguntan por el modo y crean la salida en `data/out/`.


## 🧠 Por qué importa

* **Interoperabilidad total** – exporta a un formato abierto, fácil de indexar o cargar en cualquier DB/no-SQL.
* **Optimizado para IA** – cada registro es autocontenido: ideal para _few-shot_ o *RAG*.
* **Sin vendor lock-in** – tu conocimiento sale de la wiki sin depender de su visor HTML.
* **Pipeline amigable** – integra con Airflow, Spark, n8n, etc. simplemente leyendo líneas JSONL.

---
## 📦 Requisitos

- Go 1.20 o superior
- Archivo `.json` exportado desde TiddlyWiki (estructura válida)
