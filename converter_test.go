// converter_test.go – Tests unitarios para converter.go
// ------------------------------------------------------
package main

import (
	"reflect"
	"testing"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

func TestParseTags(t *testing.T) {
	raw := "[[tag1]] [[tag 2]] [[tag3]]"
	want := []string{"tag1", "tag 2", "tag3"}
	if got := parseTags(raw); !reflect.DeepEqual(got, want) {
		t.Errorf("parseTags(%q) = %v, want %v", raw, got, want)
	}
}

func TestConvertTiddlers(t *testing.T) {
	tiddlers := []models.Tiddler{
		{
			Title:    "Foo",
			Text:     "plain text",
			Tags:     "[[a]] [[b]]",
			Created:  "20250101",
			Modified: "20250102",
			Type:     "text/plain",
		},
		{
			Title:    "Bar",
			Text:     "{\"key\":\"value\"}",
			Tags:     "[[x]]",
			Created:  "20250103",
			Modified: "20250104",
			Type:     "application/json",
		},
	}

	got := ConvertTiddlers(tiddlers)
	want := []models.Record{
		{
			ID:           "Foo",
			Tags:         []string{"a", "b"},
			ContentType:  "text/plain",
			TextMarkdown: "plain text",
			TextPlain:    "plain text",
			CreatedAt:    "20250101",
			ModifiedAt:   "20250102",
		},
		{
			ID:           "Bar",
			Tags:         []string{"x"},
			ContentType:  "application/json",
			TextMarkdown: "{\n  \"key\": \"value\"\n}",
			TextPlain:    "{\n  \"key\": \"value\"\n}",
			CreatedAt:    "20250103",
			ModifiedAt:   "20250104",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ConvertTiddlers() = %+v, want %+v", got, want)
	}
}
