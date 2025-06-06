# OpenPages Exporter

Convierte una exportación **JSON** de *TiddlyWiki* en formatos **JSONL** o **JSON** listos para ingestión por:

-➡️ Modelos de lenguaje (LLMs)
-➡️ Pipelines de IA o ETL
-➡️ Dashboards de conocimiento
-➡️ Buscadores semánticos o clasificadores automáticos

La herramienta preserva la semántica de cada *tiddler* (título, tags, fechas, color, relaciones) y ofrece **tres modos** de salida:

| Modo | Esquema | Pensado para… | Formato       |
|------|---------|---------------|---------------|
| `v1` | **Compacto heredado**<br>TextPlain / TextMarkdown | Back-compat con scripts antiguos | JSONL |
| `v2` | **Meta + Content**<br>`models.RecordV2` | LLMs que requieran contexto rico | JSONL / JSON |
| `v3` | **Minimal JSONL**<br>una línea = un objeto ligero | Spark, BigQuery, Elasticsearch | JSONL |


## 🧱 Componentes

### Entrada

- Un archivo `.json` exportado desde TiddlyWiki (preferiblemente vía [TiddlyWiki Export](https://tiddlywiki.com)).

### Salida

- Un archivo `.jsonl` válido (una línea por entrada).
- O un archivo `.json` indentado (solo para inspección humana).

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
