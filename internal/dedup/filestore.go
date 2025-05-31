package dedup

import (
	"bufio"
	"os"
	"sync"
)

// FileStore mantiene los hashes en un archivo append-only.
// Formato: un hash (hex) por línea.
type FileStore struct {
	mu     sync.RWMutex
	set    map[string]struct{}
	file   *os.File
	writer *bufio.Writer
}

// NewFileStore abre (o crea) el archivo y carga los hashes existentes.
func NewFileStore(path string) (*FileStore, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	fs := &FileStore{
		set:    make(map[string]struct{}),
		file:   f,
		writer: bufio.NewWriter(f),
	}

	// Cargar hashes previos
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fs.set[scanner.Text()] = struct{}{}
	}
	// Reposicionar al final para append
	if _, err := f.Seek(0, os.SEEK_END); err != nil {
		f.Close()
		return nil, err
	}
	return fs, scanner.Err()
}

func (fs *FileStore) Seen(h string) bool {
	fs.mu.RLock()
	_, ok := fs.set[h]
	fs.mu.RUnlock()
	return ok
}

func (fs *FileStore) Mark(h string) error {
	fs.mu.Lock()
	if _, exists := fs.set[h]; !exists {
		if _, err := fs.writer.WriteString(h + "\n"); err != nil {
			fs.mu.Unlock()
			return err
		}
		fs.set[h] = struct{}{}
	}
	fs.mu.Unlock()
	return nil
}

func (fs *FileStore) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if err := fs.writer.Flush(); err != nil {
		fs.file.Close()
		return err
	}
	return fs.file.Close()
}
