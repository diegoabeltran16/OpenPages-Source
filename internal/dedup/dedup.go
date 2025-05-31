package dedup

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/diegoabeltran16/OpenPages-Source/models"
)

// Store representa un backend de deduplicación.
//   - Seen(hash)  → true  si el hash ya existía.
//   - Mark(hash)  → persiste el hash (cuando es nuevo).
//   - Close()     → libera recursos (flush, close, etc.).
type Store interface {
	Seen(hash string) bool
	Mark(hash string) error
	Close() error
}

// HashTiddler genera un SHA-256 estable a partir de campos que
// identifican de manera única la versión de un tiddler.
// Cambios en Title, Modified o Text ⇒ nuevo hash.
func HashTiddler(t models.Tiddler) string {
	h := sha256.New()
	h.Write([]byte(t.Title))
	h.Write([]byte{0})
	h.Write([]byte(t.Modified))
	h.Write([]byte{0})
	h.Write([]byte(t.Text))
	return hex.EncodeToString(h.Sum(nil))
}
