# main.py
import os
from src.tw_json2jsonl import convert_tiddlywiki_to_jsonl

def main():
    input_path = "input/tiddlers.json"
    output_path = "output/tiddlers_output.jsonl"

    if not os.path.exists(input_path):
        print(f"❌ No se encontró el archivo: {input_path}")
        return

    print("🔁 Procesando tiddlers...")
    convert_tiddlywiki_to_jsonl(input_path, output_path)
    print(f"✅ Conversión completa. Archivo generado: {output_path}")

if __name__ == "__main__":
    main()
