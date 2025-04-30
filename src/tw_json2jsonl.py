# src/tw_json2jsonl.py
import json
from datetime import datetime
import re

def parse_tags(raw_tags: str) -> list:
    return [tag.strip("[] ") for tag in raw_tags.split("]] [[")] if tag]

def convert_date(tw_date: str) -> str:
    try:
        return datetime.strptime(tw_date[:14], "%Y%m%d%H%M%S").isoformat() + "Z"
    except:
        return None

def clean_markdown(text: str) -> str:
    return re.sub(r'[`*>#|]', '', text).strip()

def convert_tiddlywiki_to_jsonl(input_path: str, output_path: str):
    with open(input_path, "r", encoding="utf-8") as infile:
        tiddlers = json.load(infile)

    with open(output_path, "w", encoding="utf-8") as outfile:
        for t in tiddlers:
            data = {
                "id": re.sub(r'\W+', '-', t.get("title", "").lower()).strip("-"),
                "title": t.get("title", "").strip(),
                "tags": parse_tags(t.get("tags", "")),
                "text_markdown": t.get("text", "").strip(),
                "text_plain": clean_markdown(t.get("text", "")),
                "created_at": convert_date(t.get("created", "")),
                "modified_at": convert_date(t.get("modified", ""))
            }
            json.dump(data, outfile, ensure_ascii=False)
            outfile.write("\n")
