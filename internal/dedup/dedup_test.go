package dedup

import (
	"os"
	"testing"
)

func TestMemStore(t *testing.T) {
	s := NewMemStore()
	if s.Seen("x") {
		t.Fatalf("hash inesperado")
	}
	if err := s.Mark("x"); err != nil {
		t.Fatal(err)
	}
	if !s.Seen("x") {
		t.Fatalf("hash debería existir")
	}
}

func TestFileStore(t *testing.T) {
	tmp, err := os.CreateTemp("", "hashes-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	path := tmp.Name()
	tmp.Close()
	defer os.Remove(path)

	s, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Seen("a") {
		t.Fatalf("hash inesperado")
	}
	if err := s.Mark("a"); err != nil {
		t.Fatal(err)
	}
	if !s.Seen("a") {
		t.Fatalf("hash debería existir tras Mark")
	}
	s.Close()

	// Reabrir y asegurar que persiste
	s2, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if !s2.Seen("a") {
		t.Fatalf("hash debería persistir en disco")
	}
	s2.Close()
}
